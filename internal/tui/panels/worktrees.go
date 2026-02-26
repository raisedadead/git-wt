package panels

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/ui"
)

var (
	wtSelectedStyle = lipgloss.NewStyle().Foreground(ui.AdaptivePrimary).Bold(true)
	wtNormalStyle   = lipgloss.NewStyle().Foreground(ui.AdaptiveText)
	wtPathStyle     = lipgloss.NewStyle().Foreground(ui.AdaptiveSubtle)
	wtDimStyle      = lipgloss.NewStyle().Foreground(ui.AdaptiveSubtle)
	wtIconDefault   = lipgloss.NewStyle().Foreground(ui.AdaptivePrimary).Bold(true)
	wtIconCurrent   = lipgloss.NewStyle().Foreground(ui.AdaptivePrimary)
	wtStatusClean   = lipgloss.NewStyle().Foreground(ui.AdaptiveStatusClean)
	wtStatusDirty   = lipgloss.NewStyle().Foreground(ui.AdaptiveStatusDirty)
	wtStatusGone    = lipgloss.NewStyle().Foreground(ui.AdaptiveStatusGone)
	wtStatusMerged  = lipgloss.NewStyle().Foreground(ui.AdaptiveStatusMerged)
	wtMarkerStyle   = lipgloss.NewStyle().Foreground(ui.AdaptiveWarning).Bold(true)
	wtSuccessStyle  = lipgloss.NewStyle().Foreground(ui.AdaptiveSuccess)
	wtWarningStyle  = lipgloss.NewStyle().Foreground(ui.AdaptiveWarning)
	wtCursorBar     = lipgloss.NewStyle().Foreground(ui.AdaptiveCursorBar).Bold(true)
	wtSelectedBg    = lipgloss.NewStyle().Background(ui.AdaptiveListSelected)
)

// WorktreeItem represents a worktree in the list.
type WorktreeItem struct {
	Wt        git.Worktree
	Status    string
	Dirty     bool
	Current   bool
	Selected  bool
	Ahead     int
	Behind    int
	Gone      bool
	Merged    bool
	IsDefault bool
}

func (i WorktreeItem) Title() string       { return i.Wt.Branch }
func (i WorktreeItem) Description() string { return i.Status }
func (i WorktreeItem) FilterValue() string { return i.Wt.Branch + " " + i.Wt.Path }

// WorktreeDelegate renders worktree items in the list.
type WorktreeDelegate struct {
	Focused bool
}

func NewWorktreeDelegate() WorktreeDelegate {
	return WorktreeDelegate{}
}

func (d WorktreeDelegate) Height() int                               { return 2 }
func (d WorktreeDelegate) Spacing() int                              { return 1 }
func (d WorktreeDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d WorktreeDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(WorktreeItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	maxWidth := m.Width()

	// --- Line 1: {cursor}{icon} {branch}          {status} ---

	var icon string
	if isSelected && d.Focused {
		// Cursor bar replaces leading space
		cursor := wtCursorBar.Render("\u258e")
		if item.Selected {
			icon = cursor + wtMarkerStyle.Render("x") + " "
		} else if item.IsDefault {
			icon = cursor + wtIconDefault.Render("\u25cf")
		} else if item.Current {
			icon = cursor + wtIconCurrent.Render("\u25ba")
		} else {
			icon = cursor + " "
		}
	} else if item.Selected {
		icon = wtMarkerStyle.Render("[x]") + " "
	} else if item.IsDefault {
		icon = " " + wtIconDefault.Render("\u25cf") + " "
	} else if item.Current {
		icon = " " + wtIconCurrent.Render("\u25ba") + " "
	} else {
		icon = "   "
	}

	branch := item.Wt.Branch
	if isSelected && d.Focused {
		branch = wtSelectedStyle.Render(branch)
	} else if item.Selected {
		branch = wtSelectedStyle.UnsetBold().Render(branch)
	} else {
		branch = wtNormalStyle.Render(branch)
	}

	var statusStr string
	switch {
	case item.Gone:
		statusStr = wtStatusGone.Render("gone")
	case item.Merged:
		statusStr = wtStatusMerged.Render("merged")
	case item.Status == "":
		statusStr = wtDimStyle.Render("\u2026")
	case item.Status == "clean":
		statusStr = wtStatusClean.Render("\u2713")
	default:
		statusStr = wtStatusDirty.Render("\u2717 " + item.Status)
	}

	left1 := icon + branch
	right1 := statusStr

	line1 := padLine(left1, right1, maxWidth)
	if maxWidth > 0 && lipgloss.Width(line1) > maxWidth {
		line1 = ansi.Truncate(line1, maxWidth, "")
	}

	// --- Line 2: {metadata summary} ---

	var metaParts []string
	if item.Current {
		metaParts = append(metaParts, wtDimStyle.Render("(current)"))
	}
	if item.Ahead > 0 {
		metaParts = append(metaParts, wtSuccessStyle.Render(fmt.Sprintf("\u2191%d", item.Ahead)))
	}
	if item.Behind > 0 {
		metaParts = append(metaParts, wtWarningStyle.Render(fmt.Sprintf("\u2193%d", item.Behind)))
	}
	dirName := filepath.Base(item.Wt.Path)
	if dirName != item.Wt.Branch {
		metaParts = append(metaParts, wtPathStyle.Render(dirName))
	}

	left2 := "   "
	if len(metaParts) > 0 {
		left2 += wtDimStyle.Render(metaParts[0])
		for _, p := range metaParts[1:] {
			left2 += wtDimStyle.Render(" \u00b7 ") + p
		}
	} else {
		left2 += wtPathStyle.Render(filepath.Base(item.Wt.Path))
	}

	line2 := left2
	if maxWidth > 0 && lipgloss.Width(line2) > maxWidth {
		line2 = ansi.Truncate(line2, maxWidth, "")
	}

	// Apply full-width background highlight on selected item
	if isSelected && d.Focused {
		bgStyle := wtSelectedBg.Width(maxWidth)
		line1 = bgStyle.Render(line1)
		line2 = bgStyle.Render(line2)
	}

	_, _ = fmt.Fprint(w, line1+"\n"+line2)
}

func padLine(left, right string, maxWidth int) string {
	if maxWidth <= 0 {
		return left + "  " + right
	}
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)
	gap := maxWidth - leftW - rightW
	if gap < 1 {
		gap = 1
	}
	spaces := make([]byte, gap)
	for i := range spaces {
		spaces[i] = ' '
	}
	return left + string(spaces) + right
}

// WorktreeList wraps the bubbles list for worktree display.
type WorktreeList struct {
	List    list.Model
	Focused bool
}

func NewWorktreeList() WorktreeList {
	delegate := NewWorktreeDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()
	l.SetShowPagination(false)
	return WorktreeList{List: l, Focused: true}
}

func (wl *WorktreeList) SetSize(width, height int) {
	wl.List.SetSize(width, height)
}

func (wl *WorktreeList) SetItems(items []WorktreeItem) {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}
	wl.List.SetItems(listItems)
}

func (wl *WorktreeList) SetFocused(focused bool) {
	wl.Focused = focused
	d := NewWorktreeDelegate()
	d.Focused = focused
	wl.List.SetDelegate(d)
}

func (wl *WorktreeList) SelectedItem() *WorktreeItem {
	item := wl.List.SelectedItem()
	if item == nil {
		return nil
	}
	if wt, ok := item.(WorktreeItem); ok {
		return &wt
	}
	return nil
}

func (wl *WorktreeList) SelectedIndex() int {
	return wl.List.Index()
}

func (wl *WorktreeList) Update(msg tea.Msg) (WorktreeList, tea.Cmd) {
	var cmd tea.Cmd
	wl.List, cmd = wl.List.Update(msg)
	return *wl, cmd
}

func (wl *WorktreeList) View() string {
	return wl.List.View()
}

// UpdateItemStatus updates the status of a specific worktree item by path.
func (wl *WorktreeList) UpdateItemStatus(path, status string, ahead, behind int) {
	items := wl.List.Items()
	for i, item := range items {
		if wt, ok := item.(WorktreeItem); ok && wt.Wt.Path == path {
			wt.Status = status
			wt.Dirty = status != "clean" && status != "" && status != "unknown"
			wt.Ahead = ahead
			wt.Behind = behind
			items[i] = wt
			break
		}
	}
	wl.List.SetItems(items)
}

// ToggleSelection toggles the selected state of the current item.
func (wl *WorktreeList) ToggleSelection() {
	idx := wl.List.Index()
	items := wl.List.Items()
	if idx >= 0 && idx < len(items) {
		if wt, ok := items[idx].(WorktreeItem); ok {
			wt.Selected = !wt.Selected
			items[idx] = wt
			wl.List.SetItems(items)
		}
	}
}

// GetSelected returns all items with Selected=true.
func (wl *WorktreeList) GetSelected() []WorktreeItem {
	var selected []WorktreeItem
	for _, item := range wl.List.Items() {
		if wt, ok := item.(WorktreeItem); ok && wt.Selected {
			selected = append(selected, wt)
		}
	}
	return selected
}

// ClearSelection unselects all items.
func (wl *WorktreeList) ClearSelection() {
	items := wl.List.Items()
	for i, item := range items {
		if wt, ok := item.(WorktreeItem); ok && wt.Selected {
			wt.Selected = false
			items[i] = wt
		}
	}
	wl.List.SetItems(items)
}

// IsFiltering returns true if the list filter is active.
func (wl *WorktreeList) IsFiltering() bool {
	return wl.List.FilterState() == list.Filtering
}

// ItemCount returns the number of items in the list.
func (wl *WorktreeList) ItemCount() int {
	return len(wl.List.Items())
}
