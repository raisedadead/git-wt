package commands

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

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

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, ui.BoldStyle.Render("BRANCH\tSTATUS\tPATH"))

	for _, info := range infos {
		// Build status string
		statusParts := []string{info.Status}
		if info.Merged {
			statusParts = append(statusParts, "merged")
		}
		if info.RemoteGone {
			statusParts = append(statusParts, "gone")
		}
		statusStr := strings.Join(statusParts, ",")

		// Style based on status - gone/merged get warning style
		var statusStyle = ui.SuccessStyle
		if info.Merged || info.RemoteGone {
			statusStyle = ui.WarningStyle
		} else if info.Status != "clean" {
			statusStyle = ui.SubtleStyle
		}

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n",
			info.Branch,
			statusStyle.Render(statusStr),
			ui.SubtleStyle.Render(shortenPath(info.Path)),
		)
	}

	return w.Flush()
}

func shortenPath(path string) string {
	home, _ := os.UserHomeDir()
	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
