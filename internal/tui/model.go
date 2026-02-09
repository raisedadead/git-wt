package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/config"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/tui/overlays"
	"github.com/raisedadead/wt/internal/tui/panels"
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

		header:    panels.NewHeader(),
		footer:    panels.NewFooter(),
		worktrees: panels.NewWorktreeList(),
		detail:    panels.NewDetailPanel(),

		helpOverlay:    overlays.NewHelpOverlay(keys),
		confirmOverlay: overlays.NewConfirmOverlay(),
		inputOverlay:   overlays.NewInputOverlay(),
		menuOverlay:    overlays.NewMenuOverlay(),
	}
	m.footer.SetBindings(keys.listShortHelp())
	return m
}

func (m model) Init() tea.Cmd {
	return loadProjectCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

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
			m.worktrees.UpdateItemStatus(msg.path, msg.status)
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
		if msg.err != nil {
			m.setFlash(fmt.Sprintf("Prune error: %s", msg.err), true)
		} else {
			m.setFlash("Pruned stale worktrees", false)
		}
		return m, tea.Batch(loadWorktreesCmd(m.projectRoot), clearFlashCmd())

	case fetchMsg:
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

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, tea.Batch(cmds...)
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
	case key.Matches(msg, m.keys.FocusList):
		// Already on list, do nothing
		return m, nil
	case key.Matches(msg, m.keys.FocusDetail):
		m.setFocus(panelDetail)
		return m, nil
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
		m.setFlash("Fetching...", false)
		return m, fetchCmd(m.projectRoot)
	case key.Matches(msg, m.keys.Clone):
		m.switchPath = ""
		m.cloneHint = "Clone via CLI: wt clone <url> [name]"
		return m, tea.Quit
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
	case key.Matches(msg, m.keys.FocusList):
		m.setFocus(panelList)
		return m, nil
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
			return createWorktreeCmd(m.projectRoot, name)
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
					return createWorktreeCmd(m.projectRoot, branchName)
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

		m.confirmOverlay.Show(title, message,
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
	m.setFlash("Pruning stale worktrees...", false)
	return *m, pruneWorktreesCmd(m.projectRoot)
}

func (m *model) setFlash(text string, isError bool) {
	m.footer.SetFlash(text, isError)
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
		m.footer.SetBindings(m.keys.listShortHelp())
	case panelDetail:
		m.footer.SetBindings(m.keys.detailShortHelp())
	}
}

func (m *model) setWorktrees(wts []git.Worktree) {
	cwd, _ := os.Getwd()
	items := make([]panels.WorktreeItem, len(wts))
	for i, wt := range wts {
		items[i] = panels.WorktreeItem{
			Wt:      wt,
			Current: wt.Path == cwd || wt.Path == m.currentPath,
		}
	}
	m.worktrees.SetItems(items)
}

func (m *model) updateLayout() {
	m.header.SetWidth(m.width)
	m.footer.SetWidth(m.width)
	m.helpOverlay.SetSize(m.width, m.height)
	m.confirmOverlay.SetSize(m.width, m.height)
	m.inputOverlay.SetSize(m.width, m.height)
	m.menuOverlay.SetSize(m.width, m.height)

	// Main area height: total - header(1) - footer(1)
	mainHeight := m.height - 2
	if mainHeight < 0 {
		mainHeight = 0
	}

	// Left panel: ~33% width minus borders
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth

	// Subtract 2 for panel borders on each side
	leftInner := leftWidth - 2
	rightInner := rightWidth - 2
	innerHeight := mainHeight - 2

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

	// Render overlays on top if active
	if m.helpOverlay.Active {
		return m.helpOverlay.View()
	}

	// Build main layout
	headerView := m.header.View()

	// Left panel
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth
	mainHeight := m.height - 2

	var leftBorder, rightBorder lipgloss.Style
	if m.focused == panelList {
		leftBorder = focusedBorderStyle.Width(leftWidth - 2).Height(mainHeight - 2)
		rightBorder = unfocusedBorderStyle.Width(rightWidth - 2).Height(mainHeight - 2)
	} else {
		leftBorder = unfocusedBorderStyle.Width(leftWidth - 2).Height(mainHeight - 2)
		rightBorder = focusedBorderStyle.Width(rightWidth - 2).Height(mainHeight - 2)
	}

	leftTitle := panelTitleStyle.Render("Worktrees")
	rightTitle := panelTitleStyle.Render("Detail")
	if m.focused != panelList {
		leftTitle = panelTitleInactiveStyle.Render("Worktrees")
	}
	if m.focused != panelDetail {
		rightTitle = panelTitleInactiveStyle.Render("Detail")
	}

	leftContent := leftTitle + "\n" + m.worktrees.View()
	rightContent := rightTitle + "\n" + m.detail.View()

	leftPanel := leftBorder.Render(leftContent)
	rightPanel := rightBorder.Render(rightContent)

	mainArea := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	footerView := m.footer.View()

	fullView := headerView + "\n" + mainArea + "\n" + footerView

	// Overlay on top of main view
	if m.confirmOverlay.Active {
		return m.renderWithOverlay(fullView, m.confirmOverlay.View())
	}
	if m.inputOverlay.Active {
		return m.renderWithOverlay(fullView, m.inputOverlay.View())
	}
	if m.menuOverlay.Active {
		return m.renderWithOverlay(fullView, m.menuOverlay.View())
	}

	return fullView
}

func (m model) renderWithOverlay(base, overlay string) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	result := make([]string, max(len(baseLines), len(overlayLines)))
	for i := range result {
		var oLine, bLine string
		if i < len(overlayLines) {
			oLine = overlayLines[i]
		}
		if i < len(baseLines) {
			bLine = baseLines[i]
		}
		if strings.TrimSpace(oLine) == "" {
			result[i] = bLine
		} else {
			result[i] = oLine
		}
	}
	return strings.Join(result, "\n")
}
