package overlays

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/raisedadead/wt/internal/ui"
)

func RenderModal(baseView, modalView string, width, height int) string {
	dimStyle := lipgloss.NewStyle().Foreground(ui.AdaptiveModalDim)

	stripped := ansi.Strip(baseView)

	var baseLines []string
	if stripped == "" {
		baseLines = []string{}
	} else {
		baseLines = strings.Split(stripped, "\n")
	}

	dimmed := make([]string, height)
	for i := range height {
		if i < len(baseLines) {
			line := baseLines[i]
			lineWidth := ansi.StringWidth(line)
			if lineWidth > width {
				line = ansi.Truncate(line, width, "")
			} else if lineWidth < width {
				line = line + strings.Repeat(" ", width-lineWidth)
			}
			dimmed[i] = dimStyle.Render(line)
		} else {
			dimmed[i] = dimStyle.Render(strings.Repeat(" ", width))
		}
	}

	centered := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modalView)
	modalLines := strings.Split(centered, "\n")

	result := make([]string, height)
	for i := range height {
		var mLine string
		if i < len(modalLines) {
			mLine = modalLines[i]
		}
		if strings.TrimSpace(ansi.Strip(mLine)) != "" {
			result[i] = mLine
		} else {
			result[i] = dimmed[i]
		}
	}

	return strings.Join(result, "\n")
}
