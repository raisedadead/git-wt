package panels

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

// Footer renders the bottom bar with contextual keybinding hints.
type Footer struct {
	Width    int
	Flash    string
	IsError  bool
	Bindings []key.Binding

	styles struct {
		bar     lipgloss.Style
		key     lipgloss.Style
		desc    lipgloss.Style
		success lipgloss.Style
		error   lipgloss.Style
	}
}

func NewFooter() Footer {
	f := Footer{}
	f.styles.bar = lipgloss.NewStyle().Padding(0, 1)
	f.styles.key = lipgloss.NewStyle().Foreground(ui.AdaptivePrimary).Bold(true)
	f.styles.desc = lipgloss.NewStyle().Foreground(ui.AdaptiveSubtle)
	f.styles.success = lipgloss.NewStyle().Foreground(ui.AdaptiveSuccess)
	f.styles.error = lipgloss.NewStyle().Foreground(ui.AdaptiveError)
	return f
}

func (f *Footer) SetWidth(width int) {
	f.Width = width
}

func (f *Footer) SetBindings(bindings []key.Binding) {
	f.Bindings = bindings
}

func (f *Footer) SetFlash(text string, isError bool) {
	f.Flash = text
	f.IsError = isError
}

func (f *Footer) ClearFlash() {
	f.Flash = ""
	f.IsError = false
}

func (f *Footer) View() string {
	if f.Flash != "" {
		var flashStyle lipgloss.Style
		if f.IsError {
			flashStyle = f.styles.error
		} else {
			flashStyle = f.styles.success
		}
		return f.styles.bar.Width(f.Width).Render(flashStyle.Render(f.Flash))
	}

	var parts []string
	for _, b := range f.Bindings {
		if !b.Enabled() {
			continue
		}
		keys := b.Help().Key
		desc := b.Help().Desc
		if keys == "" || desc == "" {
			continue
		}
		parts = append(parts, f.styles.key.Render(keys)+":"+f.styles.desc.Render(desc))
	}

	content := strings.Join(parts, "  ")
	return f.styles.bar.Width(f.Width).Render(content)
}
