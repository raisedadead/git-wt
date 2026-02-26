package tui

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/config"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/tui/overlays"
	"github.com/raisedadead/wt/internal/tui/panels"
	"github.com/raisedadead/wt/internal/ui"
)

type panel int

const (
	panelList panel = iota
	panelDetail
)

type model struct {
	width, height int
	focused       panel
	projectRoot   string
	defaultBranch string
	currentPath   string
	keys          keyMap
	loading       bool
	err           error

	// Panels
	header    panels.Header
	footer    panels.Footer
	worktrees panels.WorktreeList
	detail    panels.DetailPanel

	// Overlays
	helpOverlay    overlays.HelpOverlay
	confirmOverlay overlays.ConfirmOverlay
	inputOverlay   overlays.InputOverlay
	menuOverlay    overlays.MenuOverlay

	// Layout
	splitRatio float64
	leftWidth  int
	rightWidth int

	// Spinner
	spinning    bool
	spinnerText string

	// State
	switchPath string
	cloneHint  string
}

func newModel() model {
	keys := newKeyMap()
	m := model{
		keys:        keys,
		loading:     true,
		currentPath: currentWorktreePath(),
		focused:     panelList,
		splitRatio:  0.33,

		header:    panels.NewHeader(),
		footer:    panels.NewFooter(),
		worktrees: panels.NewWorktreeList(),
		detail:    panels.NewDetailPanel(),

		helpOverlay:    overlays.NewHelpOverlay(keys),
		confirmOverlay: overlays.NewConfirmOverlay(),
		inputOverlay:   overlays.NewInputOverlay(),
		menuOverlay:    overlays.NewMenuOverlay(),
	}
	m.setFocus(panelList)
	return m
}

func (m model) Init() tea.Cmd {
	return loadProjectCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case projectLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		m.projectRoot = msg.projectRoot
		m.defaultBranch = msg.defaultBranch
		m.header.ProjectRoot = msg.projectRoot
		m.header.DefaultBranch = msg.defaultBranch
		return m, loadWorktreesCmd(msg.projectRoot)

	case worktreesLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.setWorktrees(msg.worktrees)
		m.header.WorktreeCount = len(msg.worktrees)
		// Load status for selected item
		if item := m.worktrees.SelectedItem(); item != nil {
			return m, tea.Batch(
				loadStatusCmd(item.Wt.Path),
				loadDetailCmd(item.Wt.Path),
			)
		}
		return m, nil

	case statusLoadedMsg:
		if msg.err == nil {
			m.worktrees.UpdateItemStatus(msg.path, msg.status, msg.ahead, msg.behind)
		}
		return m, nil

	case detailLoadedMsg:
		if msg.err == nil {
			if item := m.worktrees.SelectedItem(); item != nil && item.Wt.Path == msg.path {
				m.detail.SetInfo(item.Wt.Branch, item.Wt.Path, msg.status, msg.commits, msg.files)
			}
		}
		return m, nil

	case diffLoadedMsg:
		if item := m.worktrees.SelectedItem(); item != nil && item.Wt.Path == msg.path {
			m.detail.SetDiff(msg.diff)
		}
		return m, nil

	case logLoadedMsg:
		if item := m.worktrees.SelectedItem(); item != nil && item.Wt.Path == msg.path {
			m.detail.SetLog(msg.log)
		}
		return m, nil

	case worktreeCreatedMsg:
		m.spinning = false
		m.footer.ClearSpinner()
		if msg.err != nil {
			m.setFlash(fmt.Sprintf("Error: %s", msg.err), true)
			return m, clearFlashCmd()
		}
		m.setFlash(fmt.Sprintf("Created %s", msg.branch), false)
		return m, tea.Batch(loadWorktreesCmd(m.projectRoot), clearFlashCmd())

	case worktreeDeletedMsg:
		if msg.err != nil {
			m.setFlash(fmt.Sprintf("Error: %s", msg.err), true)
			return m, clearFlashCmd()
		}
		m.setFlash(fmt.Sprintf("Deleted %s", msg.branch), false)
		m.worktrees.ClearSelection()
		return m, tea.Batch(loadWorktreesCmd(m.projectRoot), clearFlashCmd())

	case pruneMsg:
		m.spinning = false
		m.footer.ClearSpinner()
		if msg.err != nil {
			m.setFlash(fmt.Sprintf("Prune error: %s", msg.err), true)
		} else {
			m.setFlash("Pruned stale worktrees", false)
		}
		return m, tea.Batch(loadWorktreesCmd(m.projectRoot), clearFlashCmd())

	case fetchMsg:
		m.spinning = false
		m.footer.ClearSpinner()
		if msg.err != nil {
			m.setFlash(fmt.Sprintf("Fetch error: %s", msg.err), true)
		} else {
			m.setFlash("Fetched all remotes", false)
		}
		return m, tea.Batch(loadWorktreesCmd(m.projectRoot), clearFlashCmd())

	case flashMsg:
		m.setFlash(msg.text, msg.isError)
		return m, clearFlashCmd()

	case clearFlashMsg:
		m.footer.ClearFlash()
		return m, nil

	case spinnerTickMsg:
		if m.spinning {
			return m, tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return spinnerTickMsg{} })
		}
		return m, nil

	case editorFinishedMsg:
		if msg.err != nil {
			m.setFlash(fmt.Sprintf("Editor error: %s", msg.err), true)
		} else {
			m.setFlash("Editor closed", false)
		}
		return m, clearFlashCmd()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Input overlay takes priority
	if m.inputOverlay.Active {
		var cmd tea.Cmd
		m.inputOverlay, cmd = m.inputOverlay.Update(msg)
		return m, cmd
	}

	// Confirm overlay
	if m.confirmOverlay.Active {
		var cmd tea.Cmd
		m.confirmOverlay, cmd = m.confirmOverlay.Update(msg)
		return m, cmd
	}

	// Menu overlay
	if m.menuOverlay.Active {
		var cmd tea.Cmd
		m.menuOverlay, cmd = m.menuOverlay.Update(msg)
		return m, cmd
	}

	// Help overlay
	if m.helpOverlay.Active {
		var cmd tea.Cmd
		m.helpOverlay, cmd = m.helpOverlay.Update(msg)
		return m, cmd
	}

	// If the list is filtering, pass keys to it
	if m.worktrees.IsFiltering() {
		var cmd tea.Cmd
		m.worktrees, cmd = m.worktrees.Update(msg)
		return m, cmd
	}

	// Global keys
	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		m.helpOverlay.Toggle()
		return m, nil
	case key.Matches(msg, m.keys.Refresh):
		m.setFlash("Refreshing...", false)
		return m, tea.Batch(loadWorktreesCmd(m.projectRoot), clearFlashCmd())
	case key.Matches(msg, m.keys.Tab):
		m.cycleFocus()
		return m, nil
	case key.Matches(msg, m.keys.Filter):
		if m.focused != panelList {
			m.setFocus(panelList)
		}
		var cmd tea.Cmd
		m.worktrees, cmd = m.worktrees.Update(msg)
		return m, cmd
	case msg.String() == "l" && m.focused == panelList:
		m.setFocus(panelDetail)
		return m, nil
	case msg.String() == "h" && m.focused == panelDetail:
		m.setFocus(panelList)
		return m, nil
	}

	// Focus-specific keys
	switch m.focused {
	case panelList:
		return m.handleListKey(msg)
	case panelDetail:
		return m.handleDetailKey(msg)
	}

	return m, nil
}

func (m model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	prevIdx := m.worktrees.SelectedIndex()

	switch {
	case key.Matches(msg, m.keys.Enter):
		return m.handleSwitch()
	case key.Matches(msg, m.keys.New):
		return m.handleNew()
	case key.Matches(msg, m.keys.Workflow):
		return m.handleWorkflow()
	case key.Matches(msg, m.keys.Delete):
		return m.handleDelete(false)
	case key.Matches(msg, m.keys.Force):
		return m.handleDelete(true)
	case key.Matches(msg, m.keys.Prune):
		return m.handlePrune()
	case key.Matches(msg, m.keys.Fetch):
		m.spinning = true
		m.spinnerText = "Fetching..."
		m.footer.SetSpinner("Fetching...")
		return m, tea.Batch(fetchCmd(m.projectRoot), tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return spinnerTickMsg{} }))
	case key.Matches(msg, m.keys.Clone):
		m.switchPath = ""
		m.cloneHint = "Clone via CLI: wt clone <url> [name]"
		return m, tea.Quit
	case key.Matches(msg, m.keys.Editor):
		item := m.worktrees.SelectedItem()
		if item == nil {
			return m, nil
		}
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			editor = "vim"
		}
		c := exec.Command(editor, item.Wt.Path)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return editorFinishedMsg{err: err}
		})
	case key.Matches(msg, m.keys.ResizeLeft):
		m.splitRatio = max(0.2, m.splitRatio-0.05)
		m.updateLayout()
		return m, nil
	case key.Matches(msg, m.keys.ResizeRight):
		m.splitRatio = min(0.6, m.splitRatio+0.05)
		m.updateLayout()
		return m, nil
	case key.Matches(msg, m.keys.Select):
		m.worktrees.ToggleSelection()
		// Move down after selection
		var cmd tea.Cmd
		downMsg := tea.KeyMsg{Type: tea.KeyDown}
		m.worktrees, cmd = m.worktrees.Update(downMsg)
		return m, cmd
	}

	// Pass navigation keys to list
	var cmd tea.Cmd
	m.worktrees, cmd = m.worktrees.Update(msg)

	// If cursor moved, load new item details
	newIdx := m.worktrees.SelectedIndex()
	if newIdx != prevIdx {
		cmds := []tea.Cmd{cmd}
		if item := m.worktrees.SelectedItem(); item != nil {
			m.detail.Clear()
			cmds = append(cmds,
				loadStatusCmd(item.Wt.Path),
				loadDetailCmd(item.Wt.Path),
			)
			if m.detail.Tab == panels.TabDiff {
				cmds = append(cmds, loadDiffCmd(item.Wt.Path))
			}
			if m.detail.Tab == panels.TabLog {
				cmds = append(cmds, loadLogCmd(item.Wt.Path))
			}
		}
		return m, tea.Batch(cmds...)
	}

	return m, cmd
}

func (m model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.TabInfo):
		m.detail.SetTab(panels.TabInfo)
		return m, nil
	case key.Matches(msg, m.keys.TabDiff):
		m.detail.SetTab(panels.TabDiff)
		if item := m.worktrees.SelectedItem(); item != nil {
			return m, loadDiffCmd(item.Wt.Path)
		}
		return m, nil
	case key.Matches(msg, m.keys.TabLog):
		m.detail.SetTab(panels.TabLog)
		if item := m.worktrees.SelectedItem(); item != nil {
			return m, loadLogCmd(item.Wt.Path)
		}
		return m, nil
	case key.Matches(msg, m.keys.TabPrev):
		return m.cycleDetailTab(-1)
	case key.Matches(msg, m.keys.TabNext):
		return m.cycleDetailTab(1)
	case key.Matches(msg, m.keys.ResizeLeft):
		m.splitRatio = max(0.2, m.splitRatio-0.05)
		m.updateLayout()
		return m, nil
	case key.Matches(msg, m.keys.ResizeRight):
		m.splitRatio = min(0.6, m.splitRatio+0.05)
		m.updateLayout()
		return m, nil
	}

	// Scroll detail viewport
	var cmd tea.Cmd
	m.detail, cmd = m.detail.Update(msg)
	return m, cmd
}

func (m *model) handleSwitch() (model, tea.Cmd) {
	item := m.worktrees.SelectedItem()
	if item == nil {
		return *m, nil
	}
	m.switchPath = item.Wt.Path
	return *m, tea.Quit
}

func (m *model) handleNew() (model, tea.Cmd) {
	m.inputOverlay.Show("Branch name:", "feature/my-feature",
		func(name string) tea.Cmd {
			if err := git.ValidateBranchName(name); err != nil {
				return func() tea.Msg {
					return flashMsg{text: err.Error(), isError: true}
				}
			}
			m.spinning = true
			m.spinnerText = "Creating..."
			m.footer.SetSpinner("Creating...")
			return tea.Batch(createWorktreeCmd(m.projectRoot, name), tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return spinnerTickMsg{} }))
		},
		func() tea.Cmd { return nil },
	)
	return *m, nil
}

func (m *model) handleWorkflow() (model, tea.Cmd) {
	workflows := config.DefaultWorkflows()

	// Sort workflow keys for stable order
	var keys []string
	for k := range workflows {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var items []overlays.MenuItem
	for _, k := range keys {
		wf := workflows[k]
		items = append(items, overlays.MenuItem{
			Key:         k,
			Label:       k,
			Description: wf.Description,
		})
	}

	m.menuOverlay.Show("What are you working on?", items,
		func(item overlays.MenuItem) tea.Cmd {
			wf := workflows[item.Key]
			template := wf.BranchTemplate
			label := fmt.Sprintf("Branch name (%s):", template)
			placeholder := strings.Replace(template, "{slug}", "description", 1)
			placeholder = strings.Replace(placeholder, "{name}", "branch-name", 1)
			placeholder = strings.Replace(placeholder, "{branch}", "branch-name", 1)

			m.inputOverlay.Show(label, placeholder,
				func(name string) tea.Cmd {
					branchName := strings.Replace(template, "{slug}", git.FlattenBranchName(name), 1)
					branchName = strings.Replace(branchName, "{name}", name, 1)
					branchName = strings.Replace(branchName, "{branch}", name, 1)

					if err := git.ValidateBranchName(branchName); err != nil {
						return func() tea.Msg {
							return flashMsg{text: err.Error(), isError: true}
						}
					}
					m.spinning = true
					m.spinnerText = "Creating..."
					m.footer.SetSpinner("Creating...")
					return tea.Batch(createWorktreeCmd(m.projectRoot, branchName), tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return spinnerTickMsg{} }))
				},
				func() tea.Cmd { return nil },
			)
			return nil
		},
		func() tea.Cmd { return nil },
	)
	return *m, nil
}

func (m *model) handleDelete(force bool) (model, tea.Cmd) {
	selected := m.worktrees.GetSelected()
	if len(selected) == 0 {
		item := m.worktrees.SelectedItem()
		if item == nil {
			return *m, nil
		}
		selected = []panels.WorktreeItem{*item}
	}

	// Validate: cannot delete current or default worktree
	for _, item := range selected {
		if item.Current {
			m.setFlash("Cannot delete current worktree", true)
			return *m, clearFlashCmd()
		}
		if item.Wt.Branch == m.defaultBranch {
			m.setFlash("Cannot delete default branch worktree", true)
			return *m, clearFlashCmd()
		}
	}

	if force {
		var cmds []tea.Cmd
		for _, item := range selected {
			cmds = append(cmds, deleteWorktreeCmd(m.projectRoot, item.Wt.Path, item.Wt.Branch, true))
		}
		return *m, tea.Batch(cmds...)
	}

	// Check if any are dirty
	hasDirty := false
	for _, item := range selected {
		if item.Dirty {
			hasDirty = true
			break
		}
	}

	if hasDirty {
		var names []string
		for _, item := range selected {
			names = append(names, item.Wt.Branch)
		}
		title := fmt.Sprintf("Delete %s?", strings.Join(names, ", "))
		message := "Has uncommitted changes!"
		if len(selected) > 1 {
			message = "Some worktrees have uncommitted changes!"
		}

		deleteCopy := make([]panels.WorktreeItem, len(selected))
		copy(deleteCopy, selected)
		projectRoot := m.projectRoot

		m.confirmOverlay.Show(title, message, "Yes, delete", "Cancel",
			func() tea.Cmd {
				var cmds []tea.Cmd
				for _, item := range deleteCopy {
					cmds = append(cmds, deleteWorktreeCmd(projectRoot, item.Wt.Path, item.Wt.Branch, true))
				}
				return tea.Batch(cmds...)
			},
			func() tea.Cmd { return nil },
		)
		return *m, nil
	}

	// Clean worktrees: delete directly
	var cmds []tea.Cmd
	for _, item := range selected {
		cmds = append(cmds, deleteWorktreeCmd(m.projectRoot, item.Wt.Path, item.Wt.Branch, false))
	}
	return *m, tea.Batch(cmds...)
}

func (m *model) handlePrune() (model, tea.Cmd) {
	m.spinning = true
	m.spinnerText = "Pruning..."
	m.footer.SetSpinner("Pruning...")
	return *m, tea.Batch(pruneWorktreesCmd(m.projectRoot), tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg { return spinnerTickMsg{} }))
}

func (m *model) setFlash(text string, isError bool) {
	m.footer.SetFlash(text, isError)
}

func (m model) cycleDetailTab(dir int) (tea.Model, tea.Cmd) {
	tabs := []panels.DetailTab{panels.TabInfo, panels.TabDiff, panels.TabLog}
	cur := int(m.detail.Tab)
	next := (cur + dir + len(tabs)) % len(tabs)
	m.detail.SetTab(tabs[next])
	if item := m.worktrees.SelectedItem(); item != nil {
		switch tabs[next] {
		case panels.TabDiff:
			return m, loadDiffCmd(item.Wt.Path)
		case panels.TabLog:
			return m, loadLogCmd(item.Wt.Path)
		}
	}
	return m, nil
}

func (m *model) cycleFocus() {
	switch m.focused {
	case panelList:
		m.setFocus(panelDetail)
	case panelDetail:
		m.setFocus(panelList)
	}
}

func (m *model) setFocus(p panel) {
	m.focused = p
	m.worktrees.SetFocused(p == panelList)
	m.detail.SetFocused(p == panelDetail)

	switch p {
	case panelList:
		m.footer.SetBindings([]key.Binding{
			m.keys.Enter, m.keys.New, m.keys.Delete,
			m.keys.Editor, m.keys.Filter, m.keys.Help, m.keys.Quit,
		})
	case panelDetail:
		m.footer.SetBindings([]key.Binding{
			m.keys.TabPrev, m.keys.TabNext,
			m.keys.ResizeLeft, m.keys.ResizeRight,
			m.keys.Tab, m.keys.Help, m.keys.Quit,
		})
	}
}

func (m *model) setWorktrees(wts []git.Worktree) {
	cwd, _ := os.Getwd()
	items := make([]panels.WorktreeItem, len(wts))
	for i, wt := range wts {
		items[i] = panels.WorktreeItem{
			Wt:        wt,
			Current:   wt.Path == cwd || wt.Path == m.currentPath,
			IsDefault: wt.Branch == m.defaultBranch,
		}
	}
	m.worktrees.SetItems(items)
}

// buildBorderTitle renders a top border line with an embedded title.
// Produces: ╭─ Title ────────────╮
func buildBorderTitle(title string, width int, borderColor lipgloss.TerminalColor) string {
	bc := lipgloss.NewStyle().Foreground(borderColor)
	corner := bc.Render("\u256d")
	endCorner := bc.Render("\u256e")
	dash := "\u2500"

	titleWidth := lipgloss.Width(title)
	// Available space for dashes: width - 2 corners - title
	remaining := width - 2 - titleWidth - 1 // -1 for the dash before title
	if remaining < 0 {
		remaining = 0
	}
	return corner + bc.Render(dash) + title + bc.Render(strings.Repeat(dash, remaining)) + endCorner
}

func (m *model) updateLayout() {
	m.header.SetWidth(m.width)
	m.footer.SetWidth(m.width)

	// Main area height: total - header(1) - footer(1) - border title(1)
	mainHeight := m.height - 3
	if mainHeight < 0 {
		mainHeight = 0
	}

	// Left panel width from splitRatio, clamped
	if m.width < 40 {
		m.leftWidth = m.width / 2
	} else {
		m.leftWidth = int(float64(m.width) * m.splitRatio)
		if m.leftWidth < 25 {
			m.leftWidth = 25
		}
		if m.leftWidth > m.width-30 {
			m.leftWidth = m.width - 30
		}
	}
	m.rightWidth = m.width - m.leftWidth

	// Subtract 2 for left+right borders; 1 for bottom border (top is rendered separately)
	leftInner := m.leftWidth - 2
	rightInner := m.rightWidth - 2
	innerHeight := mainHeight - 1

	if leftInner < 0 {
		leftInner = 0
	}
	if rightInner < 0 {
		rightInner = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	m.worktrees.SetSize(leftInner, innerHeight)
	m.detail.SetSize(rightInner, innerHeight)
}

func (m model) View() string {
	if m.err != nil && m.projectRoot == "" {
		return ""
	}

	if m.loading {
		return "\n  Loading..."
	}

	// Header bar with background
	projectName := headerProjectStyle.Render("wt")
	branchInfo := headerStatsStyle.Render(
		fmt.Sprintf("%s \u00b7 %d worktrees", m.defaultBranch, m.worktrees.ItemCount()))
	headerContent := projectName + headerStatsStyle.Render(" \u2022 ") + branchInfo
	headerGap := m.width - lipgloss.Width(headerContent) - 2
	if headerGap < 0 {
		headerGap = 0
	}
	headerView := headerBarStyle.Width(m.width).Render(
		headerContent + strings.Repeat(" ", headerGap))

	// Panels: no top border, we add it manually with embedded title
	mainHeight := m.height - 2

	var leftBorderStyle, rightBorderStyle lipgloss.Style
	var leftTitleStr, rightTitleStr string
	var leftBorderColor, rightBorderColor lipgloss.TerminalColor

	if m.focused == panelList {
		leftBorderStyle = focusedBorderStyle.Width(m.leftWidth - 2).Height(mainHeight - 2)
		rightBorderStyle = unfocusedBorderStyle.Width(m.rightWidth - 2).Height(mainHeight - 2)
		leftTitleStr = panelTitleStyle.Render(" Worktrees ")
		rightTitleStr = panelTitleInactiveStyle.Render(" Detail ")
		leftBorderColor = ui.AdaptiveBorderActive
		rightBorderColor = ui.AdaptiveBorderInactive
	} else {
		leftBorderStyle = unfocusedBorderStyle.Width(m.leftWidth - 2).Height(mainHeight - 2)
		rightBorderStyle = focusedBorderStyle.Width(m.rightWidth - 2).Height(mainHeight - 2)
		leftTitleStr = panelTitleInactiveStyle.Render(" Worktrees ")
		rightTitleStr = panelTitleStyle.Render(" Detail ")
		leftBorderColor = ui.AdaptiveBorderInactive
		rightBorderColor = ui.AdaptiveBorderActive
	}

	leftContent := m.worktrees.View()
	rightContent := m.detail.View()

	leftPanel := leftBorderStyle.Render(leftContent)
	rightPanel := rightBorderStyle.Render(rightContent)

	// Build top border lines with embedded titles
	leftTopBorder := buildBorderTitle(leftTitleStr, m.leftWidth, leftBorderColor)
	rightTopBorder := buildBorderTitle(rightTitleStr, m.rightWidth, rightBorderColor)

	leftPanel = leftTopBorder + "\n" + leftPanel
	rightPanel = rightTopBorder + "\n" + rightPanel

	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	footerView := m.footer.View()

	fullView := headerView + "\n" + mainArea + "\n" + footerView

	// Overlay on top of main view
	if m.helpOverlay.Active {
		return overlays.RenderModal(fullView, m.helpOverlay.View(), m.width, m.height)
	}
	if m.confirmOverlay.Active {
		return overlays.RenderModal(fullView, m.confirmOverlay.View(), m.width, m.height)
	}
	if m.inputOverlay.Active {
		return overlays.RenderModal(fullView, m.inputOverlay.View(), m.width, m.height)
	}
	if m.menuOverlay.Active {
		return overlays.RenderModal(fullView, m.menuOverlay.View(), m.width, m.height)
	}

	return fullView
}
