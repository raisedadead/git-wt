package panels

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/ui"
)

// WorktreeItem represents a worktree in the list.
type WorktreeItem struct {
	Wt       git.Worktree
	Status   string
	Dirty    bool
	Current  bool
	Selected bool
}

func (i WorktreeItem) Title() string       { return i.Wt.Branch }
func (i WorktreeItem) Description() string { return i.Status }
func (i WorktreeItem) FilterValue() string { return i.Wt.Branch + " " + i.Wt.Path }

// WorktreeDelegate renders worktree items in the list.
type WorktreeDelegate struct {
	Focused       bool
	SelectedStyle lipgloss.Style
	NormalStyle   lipgloss.Style
	CleanStyle    lipgloss.Style
	DirtyStyle    lipgloss.Style
	MarkerStyle   lipgloss.Style
	CurrentStyle  lipgloss.Style
	DimStyle      lipgloss.Style
}

func NewWorktreeDelegate() WorktreeDelegate {
	return WorktreeDelegate{
		SelectedStyle: lipgloss.NewStyle().Foreground(ui.AdaptivePrimary).Bold(true),
		NormalStyle:   lipgloss.NewStyle().Foreground(ui.AdaptiveText),
		CleanStyle:    lipgloss.NewStyle().Foreground(ui.AdaptiveSuccess),
		DirtyStyle:    lipgloss.NewStyle().Foreground(ui.AdaptiveWarning),
		MarkerStyle:   lipgloss.NewStyle().Foreground(ui.AdaptiveWarning).Bold(true),
		CurrentStyle:  lipgloss.NewStyle().Foreground(ui.AdaptiveSubtle),
		DimStyle:      lipgloss.NewStyle().Foreground(ui.AdaptiveSubtle),
	}
}

func (d WorktreeDelegate) Height() int                               { return 1 }
func (d WorktreeDelegate) Spacing() int                              { return 0 }
func (d WorktreeDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d WorktreeDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(WorktreeItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	// Build prefix
	prefix := "  "
	if item.Selected {
		prefix = d.MarkerStyle.Render("[x]") + " "
	} else if isSelected {
		prefix = d.SelectedStyle.Render("> ")
	}

	// Branch name
	branch := item.Wt.Branch
	if isSelected && d.Focused {
		branch = d.SelectedStyle.Render(branch)
	} else {
		branch = d.NormalStyle.Render(branch)
	}

	// Status indicator
	var statusStr string
	switch {
	case item.Status == "":
		statusStr = d.DimStyle.Render("...")
	case item.Status == "clean":
		statusStr = d.CleanStyle.Render("✓")
	default:
		statusStr = d.DirtyStyle.Render("✗ " + item.Status)
	}

	// Current worktree marker
	currentMark := ""
	if item.Current {
		currentMark = d.CurrentStyle.Render(" (here)")
	}

	line := fmt.Sprintf("%s%s %s%s", prefix, branch, statusStr, currentMark)

	// Truncate to fit width (ANSI-aware)
	maxWidth := m.Width()
	if maxWidth > 0 && lipgloss.Width(line) > maxWidth {
		line = ansi.Truncate(line, maxWidth, "")
	}

	fmt.Fprint(w, line)
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
func (wl *WorktreeList) UpdateItemStatus(path, status string) {
	items := wl.List.Items()
	for i, item := range items {
		if wt, ok := item.(WorktreeItem); ok && wt.Wt.Path == path {
			wt.Status = status
			wt.Dirty = status != "clean" && status != "" && status != "unknown"
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
