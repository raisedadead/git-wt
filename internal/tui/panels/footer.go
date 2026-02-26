package panels

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

type Footer struct {
	Width    int
	Flash    string
	IsError  bool
	Bindings []key.Binding
	Spinner  string
	Spinning bool

	styles struct {
		bar     lipgloss.Style
		key     lipgloss.Style
		desc    lipgloss.Style
		sep     lipgloss.Style
		success lipgloss.Style
		error   lipgloss.Style
		spinner lipgloss.Style
	}
}

func NewFooter() Footer {
	f := Footer{}
	f.styles.bar = lipgloss.NewStyle().Padding(0, 1)
	f.styles.key = lipgloss.NewStyle().Foreground(ui.AdaptivePrimary).Bold(true)
	f.styles.desc = lipgloss.NewStyle().Foreground(ui.AdaptiveSubtle)
	f.styles.sep = lipgloss.NewStyle().Foreground(ui.AdaptiveDim)
	f.styles.success = lipgloss.NewStyle().Foreground(ui.AdaptiveSuccess)
	f.styles.error = lipgloss.NewStyle().Foreground(ui.AdaptiveError)
	f.styles.spinner = lipgloss.NewStyle().Foreground(ui.AdaptiveSpinner)
	return f
}

func (f *Footer) SetWidth(width int) {
	f.Width = width
}

func (f *Footer) SetBindings(bindings []key.Binding) {
	f.Bindings = bindings
}

func (f *Footer) SetGroups(groups [][]key.Binding) {
	var flat []key.Binding
	for _, g := range groups {
		flat = append(flat, g...)
	}
	f.Bindings = flat
}

func (f *Footer) SetFlash(text string, isError bool) {
	f.Flash = text
	f.IsError = isError
}

func (f *Footer) ClearFlash() {
	f.Flash = ""
	f.IsError = false
}

func (f *Footer) SetSpinner(text string) {
	f.Spinner = text
	f.Spinning = true
}

func (f *Footer) ClearSpinner() {
	f.Spinner = ""
	f.Spinning = false
}

func (f *Footer) renderHints() string {
	sep := f.styles.sep.Render(" \u00b7 ")
	ellipsis := f.styles.sep.Render("\u2026")
	maxWidth := f.Width - 2 // padding

	var parts []string
	totalWidth := 0

	for i, b := range f.Bindings {
		if !b.Enabled() {
			continue
		}
		k := b.Help().Key
		d := b.Help().Desc
		if k == "" || d == "" {
			continue
		}

		hint := f.styles.key.Render(k) + " " + f.styles.desc.Render(d)
		hintPlain := lipgloss.Width(hint)

		sepWidth := 0
		if len(parts) > 0 {
			sepWidth = lipgloss.Width(sep)
		}

		if maxWidth > 0 && i > 0 && totalWidth+sepWidth+hintPlain > maxWidth {
			parts = append(parts, ellipsis)
			break
		}

		totalWidth += sepWidth + hintPlain
		parts = append(parts, hint)
	}

	return strings.Join(parts, sep)
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

	if f.Spinning {
		spinnerText := f.styles.spinner.Render(f.Spinner)
		hints := f.renderHints()
		gap := f.Width - lipgloss.Width(spinnerText) - lipgloss.Width(hints) - 2
		if gap < 1 {
			gap = 1
		}
		content := spinnerText + strings.Repeat(" ", gap) + hints
		return f.styles.bar.Width(f.Width).Render(content)
	}

	return f.styles.bar.Width(f.Width).Render(f.renderHints())
}
