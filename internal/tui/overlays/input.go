package overlays

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

// InputOverlay displays a text input for branch name entry.
type InputOverlay struct {
	Active   bool
	Input    textinput.Model
	Label    string
	Width    int
	Height   int
	Err      string
	OnSubmit func(string) tea.Cmd
	OnCancel func() tea.Cmd

	styles struct {
		overlay lipgloss.Style
		label   lipgloss.Style
		err     lipgloss.Style
	}
}

func NewInputOverlay() InputOverlay {
	ti := textinput.New()
	ti.Placeholder = "branch-name"
	ti.CharLimit = 128
	ti.Width = 40

	io := InputOverlay{
		Input: ti,
	}
	io.styles.overlay = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.AdaptivePrimary).
		Padding(1, 2)
	io.styles.label = lipgloss.NewStyle().
		Bold(true).
		Foreground(ui.AdaptivePrimary)
	io.styles.err = lipgloss.NewStyle().
		Foreground(ui.AdaptiveError)
	return io
}

func (io *InputOverlay) Show(label, placeholder string, onSubmit func(string) tea.Cmd, onCancel func() tea.Cmd) {
	io.Active = true
	io.Label = label
	io.Err = ""
	io.Input.SetValue("")
	io.Input.Placeholder = placeholder
	io.Input.Focus()
	io.OnSubmit = onSubmit
	io.OnCancel = onCancel
}

func (io *InputOverlay) Hide() {
	io.Active = false
	io.Input.Blur()
	io.OnSubmit = nil
	io.OnCancel = nil
}

func (io *InputOverlay) SetSize(width, height int) {
	io.Width = width
	io.Height = height
	inputWidth := width/2 - 8
	if inputWidth < 20 {
		inputWidth = 20
	}
	if inputWidth > 60 {
		inputWidth = 60
	}
	io.Input.Width = inputWidth
}

func (io *InputOverlay) SetError(err string) {
	io.Err = err
}

func (io *InputOverlay) Update(msg tea.Msg) (InputOverlay, tea.Cmd) {
	if !io.Active {
		return *io, nil
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter":
			value := io.Input.Value()
			if value != "" && io.OnSubmit != nil {
				io.Active = false
				io.Input.Blur()
				return *io, io.OnSubmit(value)
			}
		case "esc":
			io.Active = false
			io.Input.Blur()
			if io.OnCancel != nil {
				return *io, io.OnCancel()
			}
			return *io, nil
		}
	}

	var cmd tea.Cmd
	io.Input, cmd = io.Input.Update(msg)
	return *io, cmd
}

func (io *InputOverlay) View() string {
	if !io.Active {
		return ""
	}

	label := io.styles.label.Render(io.Label)
	input := io.Input.View()

	content := label + "\n\n" + input

	if io.Err != "" {
		content += "\n" + io.styles.err.Render(io.Err)
	}

	content += "\n\n" + lipgloss.NewStyle().
		Foreground(ui.AdaptiveSubtle).
		Render("enter:confirm  esc:cancel")

	overlay := io.styles.overlay.Render(content)

	return lipgloss.Place(io.Width, io.Height,
		lipgloss.Center, lipgloss.Center,
		overlay)
}
