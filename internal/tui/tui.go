package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/raisedadead/wt/internal/ui"
)

// Run launches the TUI and returns the path to switch to (if any).
// Returns ("", nil) when the user quits with q.
// Returns (path, nil) when the user presses enter on a worktree.
// Returns ("", err) on error.
func Run() (string, error) {
	m := newModel()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithOutput(os.Stderr))

	result, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("TUI error: %w", err)
	}

	finalModel, ok := result.(model)
	if !ok {
		return "", nil
	}

	if finalModel.err != nil {
		return "", finalModel.err
	}

	if finalModel.cloneHint != "" {
		return ui.InfoMsg(finalModel.cloneHint), nil
	}

	return finalModel.switchPath, nil
}
