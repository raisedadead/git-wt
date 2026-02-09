package overlays

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpOverlay shows all keybindings in a centered overlay.
type HelpOverlay struct {
	Active bool
	Help   help.Model
	Keys   help.KeyMap
	Width  int
	Height int

	styles struct {
		overlay lipgloss.Style
		title   lipgloss.Style
	}
}

func NewHelpOverlay(keys help.KeyMap) HelpOverlay {
	h := help.New()
	h.ShowAll = true

	primary := lipgloss.AdaptiveColor{Light: "#0d9488", Dark: "#14b8a6"}

	ho := HelpOverlay{
		Help: h,
		Keys: keys,
	}
	ho.styles.overlay = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(primary).
		Padding(1, 2)
	ho.styles.title = lipgloss.NewStyle().
		Bold(true).
		Foreground(primary).
		MarginBottom(1)
	return ho
}

func (ho *HelpOverlay) SetSize(width, height int) {
	ho.Width = width
	ho.Height = height
	ho.Help.Width = width - 8
}

func (ho *HelpOverlay) Toggle() {
	ho.Active = !ho.Active
}

func (ho *HelpOverlay) Update(msg tea.Msg) (HelpOverlay, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "?", "esc", "q":
			ho.Active = false
		}
	}
	return *ho, nil
}

func (ho *HelpOverlay) View() string {
	if !ho.Active {
		return ""
	}

	title := ho.styles.title.Render("Keybindings")
	helpContent := ho.Help.View(ho.Keys)

	content := title + "\n" + helpContent + "\n\n" +
		lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#9ca3af", Dark: "#6b7280"}).
			Render("Press ? or esc to close")

	overlay := ho.styles.overlay.Render(content)

	return lipgloss.Place(ho.Width, ho.Height,
		lipgloss.Center, lipgloss.Center,
		overlay)
}

// DismissKeys returns the keys that dismiss the help overlay.
func DismissKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("?", "esc", "q")),
	}
}
