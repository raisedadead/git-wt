package overlays

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

// HelpOverlay shows all keybindings in a centered overlay.
type HelpOverlay struct {
	Active bool
	Help   help.Model
	Keys   help.KeyMap

	styles struct {
		overlay lipgloss.Style
		title   lipgloss.Style
		hint    lipgloss.Style
	}
}

func NewHelpOverlay(keys help.KeyMap) HelpOverlay {
	h := help.New()
	h.ShowAll = true
	h.Width = 80

	ho := HelpOverlay{
		Help: h,
		Keys: keys,
	}
	ho.styles.overlay = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.AdaptivePrimary).
		Padding(1, 2)
	ho.styles.title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.AdaptivePrimary).
		MarginBottom(1)
	ho.styles.hint = lipgloss.NewStyle().
		Foreground(ui.AdaptiveSubtle)
	return ho
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

	hint := ho.styles.hint.Render("Press ? or esc to close")
	content := title + "\n" + helpContent + "\n\n" + hint

	return ho.styles.overlay.Render(content)
}

// DismissKeys returns the keys that dismiss the help overlay.
func DismissKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("?", "esc", "q")),
	}
}
