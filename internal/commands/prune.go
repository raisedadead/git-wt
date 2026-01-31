package commands

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/raisedadead/git-wt/internal/config"
	"github.com/raisedadead/git-wt/internal/git"
	"github.com/raisedadead/git-wt/internal/ui"
	"github.com/spf13/cobra"
)

// PruneData represents the JSON output for the prune command
type PruneData struct {
	StaleWorktrees []StaleWorktreeInfo `json:"stale_worktrees"`
	Removed        int                 `json:"removed"`
	DryRun         bool                `json:"dry_run,omitempty"`
}

// StaleWorktreeInfo represents info about a stale worktree
type StaleWorktreeInfo struct {
	Branch  string `json:"branch"`
	Path    string `json:"path"`
	Reason  string `json:"reason"`
	Removed bool   `json:"removed,omitempty"`
}

var (
	dryRunPrune      bool
	yesPrune         bool
	pruneRemoteFlag  string
	pruneTimeoutFlag int
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove stale worktrees",
	Long: `Remove worktrees whose branches have been deleted on remote or whose
directories no longer exist.`,
	RunE: runPrune,
}

func init() {
	pruneCmd.Flags().BoolVar(&dryRunPrune, "dry-run", false, "Show what would be pruned without pruning")
	pruneCmd.Flags().BoolVarP(&yesPrune, "yes", "y", false, "Skip confirmation prompt")
	pruneCmd.Flags().StringVar(&pruneRemoteFlag, "remote", "", "Override default remote")
	pruneCmd.Flags().IntVar(&pruneTimeoutFlag, "timeout", 0, "Override git operation timeout (seconds)")
	rootCmd.AddCommand(pruneCmd)
}

func runPrune(cmd *cobra.Command, args []string) error {
	// Find project root
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "prune", nil, ui.NewCLIError(ui.ErrCodeNotInProject, "not in a git-wt project"))
		}
		return fmt.Errorf("not in a git-wt project: %w", err)
	}

	// Load config with repo-level overrides
	cfg, err := config.LoadWithRepo(config.GetConfigPath(), projectRoot)
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "prune", nil, ui.NewCLIError(ui.ErrCodeGit, err.Error()))
		}
		return err
	}

	// Apply flag overrides
	if pruneRemoteFlag != "" {
		cfg.DefaultRemote = pruneRemoteFlag
	}
	if pruneTimeoutFlag > 0 {
		cfg.GitTimeout = pruneTimeoutFlag
	}

	// Fetch to get latest remote state
	if !IsJSONOutput() {
		fmt.Println(ui.SubtleStyle.Render("Fetching remote..."))
	}
	if _, err := git.RunInDirWithTimeout(projectRoot, cfg.GitTimeout, "fetch", "--prune"); err != nil {
		if !IsJSONOutput() {
			fmt.Println(ui.WarningMsg(fmt.Sprintf("Failed to fetch remote: %v (continuing with local state)", err)))
		}
	}

	// List worktrees
	worktrees, err := git.ListWorktrees(projectRoot)
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "prune", nil, ui.NewCLIError(ui.ErrCodeGit, err.Error()))
		}
		return err
	}

	// Find stale worktrees (branch deleted on remote)
	var stale []git.Worktree
	var staleInfos []StaleWorktreeInfo
	for _, wt := range worktrees {
		// Skip main/master
		if wt.Branch == git.DefaultBranch || wt.Branch == git.FallbackBranch {
			continue
		}

		// Check if branch exists on remote
		_, err := git.RunInDirWithTimeout(projectRoot, cfg.GitTimeout, "rev-parse", "--verify", fmt.Sprintf("refs/remotes/%s/%s", cfg.DefaultRemote, wt.Branch))
		if err != nil {
			stale = append(stale, wt)
			staleInfos = append(staleInfos, StaleWorktreeInfo{
				Branch: wt.Branch,
				Path:   wt.Path,
				Reason: "branch deleted on remote",
			})
		}
	}

	if len(stale) == 0 {
		if IsJSONOutput() {
			data := PruneData{
				StaleWorktrees: []StaleWorktreeInfo{},
				Removed:        0,
			}
			return ui.OutputJSON(os.Stdout, "prune", data, nil)
		}
		fmt.Println(ui.SuccessMsg("No stale worktrees found"))
		return nil
	}

	// Dry run mode - exit after showing what would be pruned
	if dryRunPrune {
		if IsJSONOutput() {
			data := PruneData{
				StaleWorktrees: staleInfos,
				Removed:        0,
				DryRun:         true,
			}
			return ui.OutputJSON(os.Stdout, "prune", data, nil)
		}
		fmt.Printf("Found %d stale worktrees:\n", len(stale))
		for _, wt := range stale {
			fmt.Println("  • " + wt.Branch + ui.SubtleStyle.Render(" (branch deleted on remote)"))
		}
		fmt.Println()
		fmt.Println(ui.InfoMsg("Dry run - no changes made"))
		return nil
	}

	// Show stale worktrees (always show in non-JSON mode)
	if !IsJSONOutput() {
		fmt.Printf("Found %d stale worktrees:\n", len(stale))
		for _, wt := range stale {
			fmt.Println("  • " + wt.Branch + ui.SubtleStyle.Render(" (branch deleted on remote)"))
		}
		fmt.Println()
	}

	// Confirmation prompt (skip with --yes or --json)
	if !yesPrune && !IsJSONOutput() {
		var action string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Remove these?").
					Options(
						huh.NewOption("Yes, remove all", "all"),
						huh.NewOption("Cancel", "cancel"),
					).
					Value(&action),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		if action == "cancel" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Remove stale worktrees
	removed := 0
	for i, wt := range stale {
		if err := git.RemoveWorktreeForce(projectRoot, wt.Path); err != nil {
			if !IsJSONOutput() {
				fmt.Println(ui.WarningMsg(fmt.Sprintf("Failed to remove %s: %v", wt.Branch, err)))
			}
			continue
		}

		if err := git.DeleteBranch(projectRoot, wt.Branch); err != nil {
			if !IsJSONOutput() {
				fmt.Println(ui.WarningMsg(fmt.Sprintf("Failed to delete branch %s: %v", wt.Branch, err)))
			}
		}

		staleInfos[i].Removed = true
		removed++
	}

	// Also run git worktree prune
	if err := git.PruneWorktrees(projectRoot); err != nil {
		if !IsJSONOutput() {
			fmt.Println(ui.WarningMsg(fmt.Sprintf("Failed to prune worktrees: %v", err)))
		}
	}

	if IsJSONOutput() {
		data := PruneData{
			StaleWorktrees: staleInfos,
			Removed:        removed,
		}
		return ui.OutputJSON(os.Stdout, "prune", data, nil)
	}

	fmt.Println(ui.SuccessMsg(fmt.Sprintf("Removed %d stale worktrees", removed)))

	return nil
}
