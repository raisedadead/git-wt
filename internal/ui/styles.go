package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors - Teal and Coral theme
	Primary = lipgloss.Color("#14b8a6") // Teal
	Success = lipgloss.Color("#2dd4bf") // Teal (lighter)
	Warning = lipgloss.Color("#ff7f50") // Coral
	Error   = lipgloss.Color("#f87171") // Red (coral-tinted)
	Subtle  = lipgloss.Color("#6b7280") // Gray

	// Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success)

	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error)

	SubtleStyle = lipgloss.NewStyle().
			Foreground(Subtle)

	BoldStyle = lipgloss.NewStyle().
			Bold(true)
)

// SuccessMsg returns a formatted success message with a checkmark prefix
func SuccessMsg(msg string) string {
	return SuccessStyle.Render("✓ ") + msg
}

// ErrorMsg returns a formatted error message with an X prefix
func ErrorMsg(msg string) string {
	return ErrorStyle.Render("✗ ") + msg
}

// WarningMsg returns a formatted warning message with a warning symbol prefix
func WarningMsg(msg string) string {
	return WarningStyle.Render("⚠ ") + msg
}

// InfoMsg returns a formatted info message with an arrow prefix
func InfoMsg(msg string) string {
	return TitleStyle.Render("→ ") + msg
}
