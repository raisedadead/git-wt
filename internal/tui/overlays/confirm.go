package overlays

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

// ConfirmOverlay displays a yes/no confirmation dialog.
type ConfirmOverlay struct {
	Active   bool
	Title    string
	Message  string
	Selected int // 0 = yes, 1 = no
	Width    int
	Height   int

	OnConfirm func() tea.Cmd
	OnCancel  func() tea.Cmd

	styles struct {
		overlay        lipgloss.Style
		title          lipgloss.Style
		message        lipgloss.Style
		buttonActive   lipgloss.Style
		buttonInactive lipgloss.Style
	}
}

func NewConfirmOverlay() ConfirmOverlay {
	co := ConfirmOverlay{Selected: 1} // default to Cancel
	co.styles.overlay = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.AdaptiveError).
		Padding(1, 2).
		Width(40)
	co.styles.title = lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.AdaptiveError)
	co.styles.message = lipgloss.NewStyle().
		Foreground(ui.AdaptiveText).
		MarginTop(1).
		MarginBottom(1)
	co.styles.buttonActive = lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.AdaptivePrimary).
		Padding(0, 2)
	co.styles.buttonInactive = lipgloss.NewStyle().
		Foreground(ui.AdaptiveSubtle).
		Padding(0, 2)
	return co
}

func (co *ConfirmOverlay) Show(title, message string, onConfirm, onCancel func() tea.Cmd) {
	co.Active = true
	co.Title = title
	co.Message = message
	co.Selected = 1
	co.OnConfirm = onConfirm
	co.OnCancel = onCancel
}

func (co *ConfirmOverlay) Hide() {
	co.Active = false
	co.OnConfirm = nil
	co.OnCancel = nil
}

func (co *ConfirmOverlay) SetSize(width, height int) {
	co.Width = width
	co.Height = height
}

func (co *ConfirmOverlay) Update(msg tea.Msg) (ConfirmOverlay, tea.Cmd) {
	if !co.Active {
		return *co, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "left", "h":
			co.Selected = 0
		case "right", "l":
			co.Selected = 1
		case "tab":
			co.Selected = (co.Selected + 1) % 2
		case "y":
			co.Active = false
			if co.OnConfirm != nil {
				return *co, co.OnConfirm()
			}
		case "n", "esc":
			co.Active = false
			if co.OnCancel != nil {
				return *co, co.OnCancel()
			}
		case "enter":
			co.Active = false
			if co.Selected == 0 && co.OnConfirm != nil {
				return *co, co.OnConfirm()
			}
			if co.OnCancel != nil {
				return *co, co.OnCancel()
			}
		}
	}

	return *co, nil
}

func (co *ConfirmOverlay) View() string {
	if !co.Active {
		return ""
	}

	title := co.styles.title.Render(co.Title)
	message := co.styles.message.Render(co.Message)

	var yesBtn, noBtn string
	if co.Selected == 0 {
		yesBtn = co.styles.buttonActive.Render("[Yes, delete]")
		noBtn = co.styles.buttonInactive.Render("[Cancel]")
	} else {
		yesBtn = co.styles.buttonInactive.Render("[Yes, delete]")
		noBtn = co.styles.buttonActive.Render("[Cancel]")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, yesBtn, "  ", noBtn)
	content := title + "\n" + message + "\n" + buttons

	overlay := co.styles.overlay.Render(content)

	return lipgloss.Place(co.Width, co.Height,
		lipgloss.Center, lipgloss.Center,
		overlay)
}
