package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

var (
	// Adaptive colors from shared ui package
	primaryColor = ui.AdaptivePrimary
	successColor = ui.AdaptiveSuccess
	warningColor = ui.AdaptiveWarning
	errorColor   = ui.AdaptiveError
	subtleColor  = ui.AdaptiveSubtle
	textColor    = ui.AdaptiveText
	dimColor     = ui.AdaptiveDim

	// Header styles
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	headerProjectStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor)

	headerStatsStyle = lipgloss.NewStyle().
				Foreground(subtleColor)

	// Panel border styles
	focusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor)

	unfocusedBorderStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(dimColor)

	// Panel title styles
	panelTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	panelTitleInactiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(subtleColor).
				Padding(0, 1)

	// Worktree list item styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(textColor)

	// Status indicator styles
	cleanStyle = lipgloss.NewStyle().
			Foreground(successColor)

	dirtyStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	// Detail tab styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Underline(true).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(subtleColor).
				Padding(0, 1)

	// Detail content styles
	detailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(subtleColor)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(textColor)

	detailSectionStyle = lipgloss.NewStyle().
				Foreground(subtleColor).
				Bold(true)

	// Overlay styles
	overlayStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 2)

	overlayTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(primaryColor)

	// Footer styles
	footerStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			Padding(0, 1)

	footerKeyStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	footerDescStyle = lipgloss.NewStyle().
			Foreground(subtleColor)

	// Flash message styles
	flashSuccessStyle = lipgloss.NewStyle().
				Foreground(successColor)

	flashErrorStyle = lipgloss.NewStyle().
			Foreground(errorColor)

	// Selection styles
	selectedMarkerStyle = lipgloss.NewStyle().
				Foreground(warningColor).
				Bold(true)
)
