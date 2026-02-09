package panels

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

// Header renders the top bar with project info.
type Header struct {
	ProjectRoot   string
	DefaultBranch string
	WorktreeCount int
	Width         int

	styles struct {
		project lipgloss.Style
		stats   lipgloss.Style
		bar     lipgloss.Style
	}
}

func NewHeader() Header {
	h := Header{}
	h.styles.project = lipgloss.NewStyle().Bold(true).Foreground(ui.AdaptivePrimary)
	h.styles.stats = lipgloss.NewStyle().Foreground(ui.AdaptiveSubtle)
	h.styles.bar = lipgloss.NewStyle().Padding(0, 1)
	return h
}

func (h *Header) SetWidth(width int) {
	h.Width = width
}

func (h *Header) View() string {
	projectName := filepath.Base(h.ProjectRoot)
	if projectName == "" || projectName == "." {
		projectName = "wt"
	}

	left := h.styles.project.Render("wt: " + projectName)
	right := h.styles.stats.Render(fmt.Sprintf("%s  %d worktrees",
		h.DefaultBranch, h.WorktreeCount))

	// Pad middle with spaces
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	padding := h.Width - leftWidth - rightWidth - 2 // -2 for outer padding
	if padding < 1 {
		padding = 1
	}

	content := left + strings.Repeat(" ", padding) + right
	return h.styles.bar.Width(h.Width).Render(content)
}
