package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/ui"
	"github.com/spf13/cobra"
)

var (
	listJSONOutput bool
	pathOutput     bool
)

// ListData represents the JSON output for the list command
type ListData struct {
	Worktrees []worktreeInfo `json:"worktrees"`
	Count     int            `json:"count"`
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all worktrees",
	RunE:    runList,
}

func init() {
	listCmd.Flags().BoolVar(&listJSONOutput, "json", false, "Output as JSON (legacy, use global --json)")
	listCmd.Flags().BoolVar(&pathOutput, "path", false, "Output paths only")
	rootCmd.AddCommand(listCmd)
}

type worktreeInfo struct {
	Branch     string `json:"branch"`
	Path       string `json:"path"`
	Status     string `json:"status"`
	Merged     bool   `json:"merged"`
	RemoteGone bool   `json:"remote_gone"`
}

func runList(cmd *cobra.Command, args []string) error {
	// Find project root
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		if IsJSONOutput() || listJSONOutput {
			return ui.OutputJSON(os.Stdout, "list", nil, ui.NewCLIError(ui.ErrCodeNotInProject, "not in a wt project"))
		}
		return fmt.Errorf("not in a wt project: %w", err)
	}

	worktrees, err := git.ListWorktrees(projectRoot)
	if err != nil {
		return err
	}

	// Get default branch for merge checking
	defaultBranch := git.GetDefaultBranchName(projectRoot)

	// Build info with status (skip .bare directory)
	var infos []worktreeInfo
	for _, wt := range worktrees {
		// Skip the bare repository itself
		if strings.HasSuffix(wt.Path, "/.bare") || wt.Branch == "" {
			continue
		}
		status, _ := git.GetWorktreeStatus(wt.Path)

		// Check if branch is merged or remote is gone (skip for default branch itself)
		merged := false
		remoteGone := false
		if wt.Branch != defaultBranch && wt.Branch != git.FallbackBranch {
			// Use IsTrulyMerged to avoid false positives on fresh branches
			merged = git.IsTrulyMerged(projectRoot, wt.Branch, defaultBranch)

			// Only check "gone" status if branch has an upstream configured
			// This avoids showing "gone" for branches that were never pushed
			if git.HasBranchUpstream(projectRoot, wt.Branch) {
				_, err := git.RunInDir(projectRoot, "rev-parse", "--verify", "refs/remotes/origin/"+wt.Branch)
				remoteGone = err != nil
			}
		}

		infos = append(infos, worktreeInfo{
			Branch:     wt.Branch,
			Path:       wt.Path,
			Status:     status,
			Merged:     merged,
			RemoteGone: remoteGone,
		})
	}

	// Check for JSON output - global --json takes precedence, legacy list --json for backward compatibility
	if IsJSONOutput() || listJSONOutput {
		data := ListData{
			Worktrees: infos,
			Count:     len(infos),
		}
		return ui.OutputJSON(os.Stdout, "list", data, nil)
	}

	if pathOutput {
		for _, info := range infos {
			fmt.Println(info.Path)
		}
		return nil
	}

	// Build rows with plain string content
	rows := make([][]string, len(infos))
	for i, info := range infos {
		statusParts := []string{info.Status}
		if info.Merged {
			statusParts = append(statusParts, "merged")
		}
		if info.RemoteGone {
			statusParts = append(statusParts, "gone")
		}
		rows[i] = []string{info.Branch, strings.Join(statusParts, ","), shortenPath(info.Path)}
	}

	// Table output with per-cell styling
	t := ui.NewStyledTable(func(row, col int) lipgloss.Style {
		cellStyle := lipgloss.NewStyle().Padding(0, 1)
		if row == table.HeaderRow {
			return cellStyle.Bold(true)
		}
		switch col {
		case 1: // STATUS
			info := infos[row]
			if info.Merged || info.RemoteGone {
				return cellStyle.Foreground(ui.Warning)
			} else if info.Status != "clean" {
				return cellStyle.Foreground(ui.Subtle)
			}
			return cellStyle.Foreground(ui.Success)
		case 2: // PATH
			return cellStyle.Foreground(ui.Subtle)
		}
		return cellStyle
	}).Headers("BRANCH", "STATUS", "PATH").Rows(rows...)

	fmt.Println(t.String())
	return nil
}

func shortenPath(path string) string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
