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
	YesLabel string
	NoLabel  string
	Selected int // 0 = yes, 1 = no

	OnConfirm func() tea.Cmd
	OnCancel  func() tea.Cmd

	styles struct {
		overlay        lipgloss.Style
		title          lipgloss.Style
		message        lipgloss.Style
		buttonActive   lipgloss.Style
		buttonInactive lipgloss.Style
		hint           lipgloss.Style
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
	co.styles.hint = lipgloss.NewStyle().
		Foreground(ui.AdaptiveSubtle)
	return co
}

func (co *ConfirmOverlay) Show(title, message, yesLabel, noLabel string, onConfirm, onCancel func() tea.Cmd) {
	co.Active = true
	co.Title = title
	co.Message = message
	co.Selected = 1
	co.YesLabel = yesLabel
	co.NoLabel = noLabel
	if co.YesLabel == "" {
		co.YesLabel = "Confirm"
	}
	if co.NoLabel == "" {
		co.NoLabel = "Cancel"
	}
	co.OnConfirm = onConfirm
	co.OnCancel = onCancel
}

func (co *ConfirmOverlay) Hide() {
	co.Active = false
	co.OnConfirm = nil
	co.OnCancel = nil
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

	title := co.styles.title.Render("\u26a0 " + co.Title)
	message := co.styles.message.Render(co.Message)

	yesLabel := co.YesLabel
	if yesLabel == "" {
		yesLabel = "Confirm"
	}
	noLabel := co.NoLabel
	if noLabel == "" {
		noLabel = "Cancel"
	}

	var yesBtn, noBtn string
	if co.Selected == 0 {
		yesBtn = co.styles.buttonActive.Render("[" + yesLabel + "]")
		noBtn = co.styles.buttonInactive.Render("[" + noLabel + "]")
	} else {
		yesBtn = co.styles.buttonInactive.Render("[" + yesLabel + "]")
		noBtn = co.styles.buttonActive.Render("[" + noLabel + "]")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, yesBtn, "  ", noBtn)
	hint := co.styles.hint.Render("y/n  enter/esc")
	content := title + "\n" + message + "\n" + buttons + "\n\n" + hint

	return co.styles.overlay.Render(content)
}
