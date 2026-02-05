package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/raisedadead/wt/internal/config"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/ui"
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
	pruneFetchFlag   bool
	pruneMergedFlag  bool
)

var pruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove stale worktrees",
	Long: `Remove worktrees whose branches have been deleted on remote or whose
directories no longer exist.

Use --merged to also remove worktrees for branches that have been merged
into the default branch (main/master).`,
	RunE: runPrune,
}

func init() {
	pruneCmd.Flags().BoolVar(&dryRunPrune, "dry-run", false, "Show what would be pruned without pruning")
	pruneCmd.Flags().BoolVarP(&yesPrune, "yes", "y", false, "Skip confirmation prompt")
	pruneCmd.Flags().StringVar(&pruneRemoteFlag, "remote", "", "Override default remote")
	pruneCmd.Flags().IntVar(&pruneTimeoutFlag, "timeout", 0, "Override git operation timeout (seconds)")
	pruneCmd.Flags().BoolVar(&pruneFetchFlag, "fetch", false, "Fetch remote before checking for stale branches")
	pruneCmd.Flags().BoolVar(&pruneMergedFlag, "merged", false, "Also remove worktrees for merged branches")
	rootCmd.AddCommand(pruneCmd)
}

func runPrune(cmd *cobra.Command, args []string) error {
	// Find project root
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "prune", nil, ui.NewCLIError(ui.ErrCodeNotInProject, "not in a wt project"))
		}
		return fmt.Errorf("not in a wt project: %w", err)
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

	// First, clean up stale worktree entries (missing directories)
	if !IsJSONOutput() {
		fmt.Println(ui.SubtleStyle.Render("Cleaning up stale worktree entries..."))
	}
	if err := git.PruneWorktrees(projectRoot); err != nil {
		if !IsJSONOutput() {
			fmt.Println(ui.WarningMsg(fmt.Sprintf("Failed to prune stale entries: %v", err)))
		}
	}

	// Optionally fetch to get latest remote state
	if pruneFetchFlag {
		if !IsJSONOutput() {
			fmt.Println(ui.SubtleStyle.Render("Fetching remote..."))
		}
		if _, err := git.RunInDirWithTimeout(projectRoot, cfg.GitTimeout, "fetch", "--prune"); err != nil {
			if !IsJSONOutput() {
				fmt.Println(ui.WarningMsg(fmt.Sprintf("Failed to fetch remote: %v (continuing with local state)", err)))
			}
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

	// Get default branch for merge checking
	defaultBranch := git.GetDefaultBranchName(projectRoot)

	// Find stale worktrees (branch deleted on remote or merged)
	var stale []git.Worktree
	var staleInfos []StaleWorktreeInfo
	for _, wt := range worktrees {
		// Skip entries without a branch (bare repo, detached HEAD)
		if wt.Branch == "" {
			continue
		}

		// Skip main/master
		if wt.Branch == git.DefaultBranch || wt.Branch == git.FallbackBranch {
			continue
		}

		// Skip the bare repo directory
		if strings.HasSuffix(wt.Path, git.BareDir) {
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
			continue
		}

		// If --merged flag, also check for merged branches
		if pruneMergedFlag {
			if git.IsBranchMerged(projectRoot, wt.Branch, defaultBranch) {
				stale = append(stale, wt)
				staleInfos = append(staleInfos, StaleWorktreeInfo{
					Branch: wt.Branch,
					Path:   wt.Path,
					Reason: fmt.Sprintf("merged into %s", defaultBranch),
				})
			}
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
		fmt.Printf("Found %d worktrees to prune:\n", len(stale))
		for _, info := range staleInfos {
			fmt.Println("  • " + info.Branch + ui.SubtleStyle.Render(fmt.Sprintf(" (%s)", info.Reason)))
		}
		fmt.Println()
		fmt.Println(ui.InfoMsg("Dry run - no changes made"))
		return nil
	}

	// Show stale worktrees (always show in non-JSON mode)
	if !IsJSONOutput() {
		fmt.Printf("Found %d worktrees to prune:\n", len(stale))
		for _, info := range staleInfos {
			fmt.Println("  • " + info.Branch + ui.SubtleStyle.Render(fmt.Sprintf(" (%s)", info.Reason)))
		}
		fmt.Println()
	}

	// Confirmation prompt (skip with --yes or --json)
	if !yesPrune && !IsJSONOutput() {
		var action string
		// Extract branch names for description
		branchNames := make([]string, len(staleInfos))
		for i, info := range staleInfos {
			branchNames[i] = info.Branch
		}
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title(pruneConfirmationTitle(len(stale))).
					Description(pruneConfirmationDescription(branchNames)).
					Options(
						huh.NewOption("Yes, remove all", "all"),
						huh.NewOption("Cancel", "cancel"),
					).
					Value(&action),
			),
		).WithKeyMap(DefaultFormKeyMap())

		if err := form.Run(); err != nil {
			return IsUserAbort(err)
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

func pruneConfirmationTitle(count int) string {
	if count == 1 {
		return "Remove 1 stale worktree?"
	}
	return fmt.Sprintf("Remove %d stale worktrees?", count)
}

func pruneConfirmationDescription(branches []string) string {
	if len(branches) == 0 {
		return ""
	}
	const maxDisplay = 5
	if len(branches) <= maxDisplay {
		return "Branches: " + strings.Join(branches, ", ")
	}
	displayed := branches[:maxDisplay]
	remaining := len(branches) - maxDisplay
	return fmt.Sprintf("Branches: %s and %d more...", strings.Join(displayed, ", "), remaining)
}
