package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

var (
	// Panel border styles
	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ui.AdaptivePrimary)

	unfocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(ui.AdaptiveDim)

	// Panel title styles
	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ui.AdaptivePrimary).
			Padding(0, 1)

	panelTitleInactiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ui.AdaptiveSubtle).
				Padding(0, 1)
)
