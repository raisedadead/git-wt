package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors - Teal and Coral theme
	Primary = lipgloss.Color("#14b8a6") // Teal
	Success = lipgloss.Color("#2dd4bf") // Teal (lighter)
	Warning = lipgloss.Color("#ff7f50") // Coral
	Error   = lipgloss.Color("#f87171") // Red (coral-tinted)
	Subtle  = lipgloss.Color("#6b7280") // Gray

	// Adaptive colors for TUI components (light/dark terminal support)
	AdaptivePrimary = lipgloss.AdaptiveColor{Light: "#0d9488", Dark: "#14b8a6"}
	AdaptiveSuccess = lipgloss.AdaptiveColor{Light: "#0d9488", Dark: "#2dd4bf"}
	AdaptiveWarning = lipgloss.AdaptiveColor{Light: "#ea580c", Dark: "#ff7f50"}
	AdaptiveError   = lipgloss.AdaptiveColor{Light: "#dc2626", Dark: "#f87171"}
	AdaptiveSubtle  = lipgloss.AdaptiveColor{Light: "#9ca3af", Dark: "#6b7280"}
	AdaptiveText    = lipgloss.AdaptiveColor{Light: "#1f2937", Dark: "#e5e7eb"}
	AdaptiveDim     = lipgloss.AdaptiveColor{Light: "#d1d5db", Dark: "#374151"}

	// Diff colors
	AdaptiveDiffAdd    = lipgloss.AdaptiveColor{Light: "#16a34a", Dark: "#4ade80"}
	AdaptiveDiffRemove = lipgloss.AdaptiveColor{Light: "#dc2626", Dark: "#f87171"}
	AdaptiveDiffHunk   = lipgloss.AdaptiveColor{Light: "#0891b2", Dark: "#22d3ee"}
	AdaptiveDiffMeta   = AdaptiveText

	// Modal colors
	AdaptiveModalDim = lipgloss.AdaptiveColor{Light: "#b0b0b0", Dark: "#444444"}

	// Tab bar colors
	AdaptiveTabActive   = lipgloss.AdaptiveColor{Light: "#0d9488", Dark: "#14b8a6"}
	AdaptiveTabInactive = lipgloss.AdaptiveColor{Light: "#9ca3af", Dark: "#6b7280"}
	AdaptiveTabBar      = lipgloss.AdaptiveColor{Light: "#e5e7eb", Dark: "#1f2937"}

	// Status indicator colors
	AdaptiveStatusClean  = AdaptiveSuccess
	AdaptiveStatusDirty  = AdaptiveWarning
	AdaptiveStatusGone   = AdaptiveError
	AdaptiveStatusMerged = AdaptiveSubtle

	// List item colors
	AdaptiveListSelected   = lipgloss.AdaptiveColor{Light: "#ccfbf1", Dark: "#042f2e"}
	AdaptiveListPath       = AdaptiveSubtle
	AdaptiveCursorBar      = AdaptivePrimary
	AdaptiveHeaderBg       = lipgloss.AdaptiveColor{Light: "#f0fdfa", Dark: "#042f2e"}
	AdaptiveHeaderFg       = lipgloss.AdaptiveColor{Light: "#0d9488", Dark: "#14b8a6"}
	AdaptiveBorderActive   = lipgloss.AdaptiveColor{Light: "#0d9488", Dark: "#5eead4"}
	AdaptiveBorderInactive = lipgloss.AdaptiveColor{Light: "#d1d5db", Dark: "#374151"}

	// Spinner color
	AdaptiveSpinner = AdaptivePrimary

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
