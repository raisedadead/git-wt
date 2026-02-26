package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

var (
	// Panel border styles — no top border; we render the top line manually with title
	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), false, true, true, true).
				BorderForeground(ui.AdaptiveBorderActive)

	unfocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder(), false, true, true, true).
				BorderForeground(ui.AdaptiveBorderInactive)

	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ui.AdaptiveBorderActive)

	panelTitleInactiveStyle = lipgloss.NewStyle().
				Foreground(ui.AdaptiveBorderInactive)

	headerBarStyle = lipgloss.NewStyle().
			Background(ui.AdaptiveHeaderBg).
			Padding(0, 1)

	headerProjectStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ui.AdaptiveHeaderFg).
				Background(ui.AdaptiveHeaderBg)

	headerStatsStyle = lipgloss.NewStyle().
				Foreground(ui.AdaptiveSubtle).
				Background(ui.AdaptiveHeaderBg)
)
