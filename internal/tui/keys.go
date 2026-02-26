package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Quit    key.Binding
	Help    key.Binding
	Filter  key.Binding
	Refresh key.Binding
	Tab     key.Binding

	// List keys
	Up       key.Binding
	Down     key.Binding
	Enter    key.Binding
	New      key.Binding
	Workflow key.Binding
	Delete   key.Binding
	Force    key.Binding
	Prune    key.Binding
	Clone    key.Binding
	Fetch    key.Binding
	Select   key.Binding

	// Detail keys
	TabInfo     key.Binding
	TabDiff     key.Binding
	TabLog      key.Binding
	TabPrev     key.Binding
	TabNext     key.Binding
	PgUp        key.Binding
	PgDown      key.Binding
	ResizeLeft  key.Binding
	ResizeRight key.Binding

	// Action keys
	Editor key.Binding

	// Input/overlay keys
	Confirm key.Binding
	Cancel  key.Binding
}

func newKeyMap() keyMap {
	return keyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "refresh"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch panel"),
		),
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "down"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "switch"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new"),
		),
		Workflow: key.NewBinding(
			key.WithKeys("N"),
			key.WithHelp("N", "workflow"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Force: key.NewBinding(
			key.WithKeys("D"),
			key.WithHelp("D", "force delete"),
		),
		Prune: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "prune"),
		),
		Clone: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "clone"),
		),
		Fetch: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "fetch"),
		),
		Select: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "select"),
		),
		TabInfo: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "info"),
		),
		TabDiff: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "diff"),
		),
		TabLog: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "log"),
		),
		TabPrev: key.NewBinding(
			key.WithKeys("["),
			key.WithHelp("[", "prev tab"),
		),
		TabNext: key.NewBinding(
			key.WithKeys("]"),
			key.WithHelp("]", "next tab"),
		),
		PgUp: key.NewBinding(
			key.WithKeys("pgup"),
			key.WithHelp("pgup", "page up"),
		),
		PgDown: key.NewBinding(
			key.WithKeys("pgdown"),
			key.WithHelp("pgdn", "page down"),
		),
		ResizeLeft: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "shrink left"),
		),
		ResizeRight: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "grow left"),
		),
		Editor: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "editor"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// ShortHelp returns the short help keybindings shown in the footer.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.New, k.Delete, k.Tab, k.Help, k.Quit}
}

// FullHelp returns the full keybinding help grouped by context.
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Quit, k.Help, k.Filter, k.Refresh, k.Tab},
		{k.Up, k.Down, k.Enter, k.New, k.Workflow, k.Editor},
		{k.Delete, k.Force, k.Prune, k.Fetch, k.Clone, k.Select},
		{k.TabPrev, k.TabNext, k.TabInfo, k.TabDiff, k.TabLog, k.PgUp, k.PgDown, k.ResizeLeft, k.ResizeRight},
	}
}
