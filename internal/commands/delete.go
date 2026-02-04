package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/raisedadead/wt/internal/config"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/ui"
	"github.com/spf13/cobra"
)

// DeleteData represents the JSON output for the delete command
type DeleteData struct {
	Branch        string `json:"branch"`
	Path          string `json:"path"`
	BranchDeleted bool   `json:"branch_deleted"`
	DryRun        bool   `json:"dry_run,omitempty"`
	Status        string `json:"status,omitempty"`
}

// DeleteMultiData represents the JSON output for multi-delete
type DeleteMultiData struct {
	Deleted []DeleteData `json:"deleted"`
	Failed  []string     `json:"failed,omitempty"`
}

var (
	forceDelete       bool
	dryRunDelete      bool
	yesDelete         bool
	deleteTimeoutFlag int
)

var deleteCmd = &cobra.Command{
	Use:     "delete [branch...]",
	Aliases: []string{"rm"},
	Short:   "Remove worktrees and their branches",
	Long: `Remove one or more worktrees and their associated branches.

In interactive mode (no arguments), presents a multi-select list.
Use space to select/deselect, enter to confirm.`,
	Args: cobra.ArbitraryArgs,
	RunE: runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force delete even with uncommitted changes")
	deleteCmd.Flags().BoolVar(&dryRunDelete, "dry-run", false, "Show what would be deleted without deleting")
	deleteCmd.Flags().BoolVarP(&yesDelete, "yes", "y", false, "Skip confirmation prompt")
	deleteCmd.Flags().IntVar(&deleteTimeoutFlag, "timeout", 0, "Override git operation timeout (seconds)")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	var branchNames []string

	// Find project root first (needed for interactive mode)
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "delete", nil, ui.NewCLIError(ui.ErrCodeNotInProject, "not in a wt project"))
		}
		return fmt.Errorf("not in a wt project: %w", err)
	}

	// Load config
	cfg, err := config.LoadWithRepo(config.GetConfigPath(), projectRoot)
	if err != nil {
		return err
	}
	if deleteTimeoutFlag > 0 {
		cfg.GitTimeout = deleteTimeoutFlag
	}

	if len(args) > 0 {
		branchNames = args
	} else {
		// Interactive mode - skip if JSON output
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "delete", nil,
				ui.NewCLIError(ui.ErrCodeValidation, "branch name is required"))
		}

		// Get worktrees, exclude default branch
		worktrees, err := git.ListWorktrees(projectRoot)
		if err != nil {
			return err
		}

		defaultBranch, _ := git.GetDefaultBranch(projectRoot)
		if defaultBranch == "" {
			defaultBranch = git.DefaultBranch
		}

		// Build options excluding default branch and .bare
		var options []huh.Option[string]
		for _, wt := range worktrees {
			if wt.Branch == "" || wt.Branch == defaultBranch ||
				strings.HasSuffix(wt.Path, "/.bare") {
				continue
			}
			// Show status hint in option label
			status, _ := git.GetWorktreeStatus(wt.Path)
			label := wt.Branch
			if status == "unknown" {
				label = fmt.Sprintf("%s (missing)", wt.Branch)
			} else if status != "clean" {
				label = fmt.Sprintf("%s (%s)", wt.Branch, status)
			}
			options = append(options, huh.NewOption(label, wt.Branch))
		}

		if len(options) == 0 {
			fmt.Println("No worktrees to delete (only default branch exists)")
			return nil
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewMultiSelect[string]().
					Title("Select worktrees to delete").
					Description("Space to select, Enter to confirm").
					Options(options...).
					Value(&branchNames),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		if len(branchNames) == 0 {
			fmt.Println("No worktrees selected.")
			return nil
		}
	}

	// Process each branch
	var results []DeleteData
	var failed []string

	for _, branchName := range branchNames {
		result, err := deleteSingleWorktree(projectRoot, branchName, cfg)
		if err != nil {
			if IsJSONOutput() {
				failed = append(failed, fmt.Sprintf("%s: %v", branchName, err))
			} else {
				fmt.Println(ui.ErrorMsg(fmt.Sprintf("Failed to delete %s: %v", branchName, err)))
			}
			continue
		}
		if result != nil {
			results = append(results, *result)
		}
	}

	// JSON output
	if IsJSONOutput() {
		if len(branchNames) == 1 && len(results) == 1 {
			// Single deletion - return single object for backward compatibility
			return ui.OutputJSON(os.Stdout, "delete", results[0], nil)
		}
		// Multiple deletions
		data := DeleteMultiData{
			Deleted: results,
			Failed:  failed,
		}
		return ui.OutputJSON(os.Stdout, "delete", data, nil)
	}

	return nil
}

func deleteSingleWorktree(projectRoot, branchName string, cfg *config.Config) (*DeleteData, error) {
	// Use flattened branch name for directory path
	worktreeDir := git.FlattenBranchName(branchName)
	worktreePath := filepath.Join(projectRoot, worktreeDir)

	// Check if worktree exists in git's list (not just if directory exists)
	worktrees, err := git.ListWorktrees(projectRoot)
	if err != nil {
		return nil, err
	}

	var worktreeExists bool
	var directoryMissing bool
	for _, wt := range worktrees {
		if wt.Branch == branchName || wt.Path == worktreePath {
			worktreeExists = true
			// Check if directory actually exists
			if _, err := os.Stat(wt.Path); os.IsNotExist(err) {
				directoryMissing = true
			}
			break
		}
	}

	if !worktreeExists {
		if IsJSONOutput() {
			return nil, ui.NewCLIError(ui.ErrCodeNotFound, fmt.Sprintf("worktree not found: %s", branchName))
		}
		return nil, fmt.Errorf("worktree not found: %s", branchName)
	}

	// If directory is missing, inform user
	if directoryMissing && !IsJSONOutput() {
		fmt.Println(ui.WarningMsg(fmt.Sprintf("Directory missing for worktree '%s', will clean up git reference", branchName)))
	}

	// Get status (skip if directory is missing)
	var status string
	if directoryMissing {
		status = "missing"
	} else {
		status, _ = git.GetWorktreeStatus(worktreePath)
	}

	// Dry run mode
	if dryRunDelete {
		if IsJSONOutput() {
			return &DeleteData{
				Branch: branchName,
				Path:   worktreePath,
				DryRun: true,
				Status: status,
			}, nil
		}
		fmt.Println(ui.InfoMsg("Dry run - would delete:"))
		fmt.Printf("  Worktree: %s\n", worktreePath)
		fmt.Printf("  Branch: %s\n", branchName)
		if status == "missing" {
			fmt.Println(ui.WarningMsg("  Status: directory missing (will clean up git reference)"))
		} else if status != "clean" {
			fmt.Println(ui.WarningMsg(fmt.Sprintf("  Status: %s", status)))
		}
		return nil, nil
	}

	// Check for uncommitted changes (skip if directory missing - nothing to lose)
	if status != "clean" && status != "missing" && !forceDelete {
		// Dirty worktrees require --force flag
		if IsJSONOutput() {
			return nil, ui.NewCLIError(ui.ErrCodeValidation, fmt.Sprintf("worktree has uncommitted changes, use --force to delete (status: %s)", status))
		}

		fmt.Println(ui.WarningMsg(fmt.Sprintf("%s has uncommitted changes:", branchName)))

		// Show changed files
		output, _ := git.RunInDirWithTimeout(worktreePath, cfg.GitTimeout, "status", "--porcelain")
		for _, line := range splitByNewline(output) {
			fmt.Println("  " + line)
		}
		fmt.Println()
		fmt.Println("Use --force to delete worktrees with uncommitted changes.")
		return nil, nil
	}

	// Confirmation prompt (skip with --yes or --json)
	if !yesDelete && !IsJSONOutput() {
		title := fmt.Sprintf("Delete worktree '%s'?", branchName)
		affirmative := "Yes, delete"
		if status == "missing" {
			title = fmt.Sprintf("Clean up missing worktree '%s'?", branchName)
			affirmative = "Yes, clean up"
		} else if status != "clean" {
			title = fmt.Sprintf("Delete worktree '%s' with uncommitted changes?", branchName)
			affirmative = "Yes, discard changes"
		}

		var confirm bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title(title).
					Affirmative(affirmative).
					Negative("Cancel").
					Value(&confirm),
			),
		)

		if err := form.Run(); err != nil {
			return nil, err
		}

		if !confirm {
			fmt.Println("Skipped", branchName)
			return nil, nil
		}
	}

	if !IsJSONOutput() {
		fmt.Println(ui.SubtleStyle.Render(fmt.Sprintf("Deleting %s...", branchName)))
	}

	// Remove worktree
	var removeErr error
	if forceDelete || directoryMissing {
		removeErr = git.RemoveWorktreeForce(projectRoot, worktreePath)
	} else {
		removeErr = git.RemoveWorktree(projectRoot, worktreePath)
	}

	if removeErr != nil {
		return nil, removeErr
	}
	if !IsJSONOutput() {
		fmt.Println(ui.SuccessMsg(fmt.Sprintf("Removed worktree %s/", branchName)))
	}

	// Delete branch
	branchDeleted := false
	if err := git.DeleteBranch(projectRoot, branchName); err != nil {
		if !IsJSONOutput() {
			fmt.Println(ui.WarningMsg(fmt.Sprintf("Could not delete branch: %v", err)))
		}
	} else {
		branchDeleted = true
		if !IsJSONOutput() {
			fmt.Println(ui.SuccessMsg(fmt.Sprintf("Deleted branch %s", branchName)))
		}
	}

	// JSON output
	if IsJSONOutput() {
		return &DeleteData{
			Branch:        branchName,
			Path:          worktreePath,
			BranchDeleted: branchDeleted,
		}, nil
	}

	return &DeleteData{
		Branch:        branchName,
		Path:          worktreePath,
		BranchDeleted: branchDeleted,
	}, nil
}

func splitByNewline(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSpace(s), "\n")
}
