package overlays

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

// MenuItem represents an item in the menu overlay.
type MenuItem struct {
	Key         string
	Label       string
	Description string
}

// MenuOverlay displays a selection menu for things like workflow picker.
type MenuOverlay struct {
	Active   bool
	Title    string
	Items    []MenuItem
	Cursor   int
	Width    int
	Height   int
	OnSelect func(MenuItem) tea.Cmd
	OnCancel func() tea.Cmd

	styles struct {
		overlay  lipgloss.Style
		title    lipgloss.Style
		selected lipgloss.Style
		normal   lipgloss.Style
		desc     lipgloss.Style
	}
}

func NewMenuOverlay() MenuOverlay {
	mo := MenuOverlay{}
	mo.styles.overlay = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.AdaptivePrimary).
		Padding(1, 2)
	mo.styles.title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.AdaptivePrimary).
		MarginBottom(1)
	mo.styles.selected = lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.AdaptivePrimary)
	mo.styles.normal = lipgloss.NewStyle().
		Foreground(ui.AdaptiveText)
	mo.styles.desc = lipgloss.NewStyle().
		Foreground(ui.AdaptiveSubtle)
	return mo
}

func (mo *MenuOverlay) Show(title string, items []MenuItem, onSelect func(MenuItem) tea.Cmd, onCancel func() tea.Cmd) {
	mo.Active = true
	mo.Title = title
	mo.Items = items
	mo.Cursor = 0
	mo.OnSelect = onSelect
	mo.OnCancel = onCancel
}

func (mo *MenuOverlay) Hide() {
	mo.Active = false
	mo.OnSelect = nil
	mo.OnCancel = nil
}

func (mo *MenuOverlay) SetSize(width, height int) {
	mo.Width = width
	mo.Height = height
}

func (mo *MenuOverlay) Update(msg tea.Msg) (MenuOverlay, tea.Cmd) {
	if !mo.Active {
		return *mo, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "j", "down":
			if mo.Cursor < len(mo.Items)-1 {
				mo.Cursor++
			}
		case "k", "up":
			if mo.Cursor > 0 {
				mo.Cursor--
			}
		case "enter":
			if mo.Cursor < len(mo.Items) && mo.OnSelect != nil {
				mo.Active = false
				return *mo, mo.OnSelect(mo.Items[mo.Cursor])
			}
		case "esc", "q":
			mo.Active = false
			if mo.OnCancel != nil {
				return *mo, mo.OnCancel()
			}
		}
	}

	return *mo, nil
}

func (mo *MenuOverlay) View() string {
	if !mo.Active {
		return ""
	}

	title := mo.styles.title.Render(mo.Title)

	var items string
	for i, item := range mo.Items {
		prefix := "  "
		var line string
		if i == mo.Cursor {
			prefix = "> "
			line = mo.styles.selected.Render(prefix + item.Label)
		} else {
			line = mo.styles.normal.Render(prefix + item.Label)
		}
		if item.Description != "" {
			line += " " + mo.styles.desc.Render(fmt.Sprintf("(%s)", item.Description))
		}
		items += line + "\n"
	}

	content := title + "\n" + items + "\n" +
		lipgloss.NewStyle().
			Foreground(ui.AdaptiveSubtle).
			Render("j/k:navigate  enter:select  esc:cancel")

	overlay := mo.styles.overlay.Render(content)

	return lipgloss.Place(mo.Width, mo.Height,
		lipgloss.Center, lipgloss.Center,
		overlay)
}
