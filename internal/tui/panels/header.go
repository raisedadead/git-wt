package panels

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/lipgloss"
	"github.com/raisedadead/wt/internal/ui"
)

type Header struct {
	ProjectRoot   string
	DefaultBranch string
	WorktreeCount int
	Width         int

	styles struct {
		project lipgloss.Style
		stats   lipgloss.Style
	}
}

func NewHeader() Header {
	h := Header{}
	h.styles.project = lipgloss.NewStyle().Bold(true).Foreground(ui.AdaptivePrimary)
	h.styles.stats = lipgloss.NewStyle().Foreground(ui.AdaptiveSubtle)
	return h
}

func (h *Header) SetWidth(width int) {
	h.Width = width
}

func (h *Header) TitleLeft() string {
	projectName := filepath.Base(h.ProjectRoot)
	if projectName == "" || projectName == "." {
		projectName = "wt"
	}
	return h.styles.project.Render("wt: " + projectName)
}

func (h *Header) TitleRight() string {
	return h.styles.stats.Render(fmt.Sprintf("%s  %d worktrees",
		h.DefaultBranch, h.WorktreeCount))
}
