package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/raisedadead/git-wt/internal/config"
	"github.com/raisedadead/git-wt/internal/git"
	"github.com/raisedadead/git-wt/internal/ui"
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

var (
	forceDelete       bool
	dryRunDelete      bool
	yesDelete         bool
	deleteTimeoutFlag int
)

var deleteCmd = &cobra.Command{
	Use:     "delete [branch]",
	Aliases: []string{"rm"},
	Short:   "Remove a worktree and its branch",
	Args:    cobra.MaximumNArgs(1),
	RunE:    runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force delete even with uncommitted changes")
	deleteCmd.Flags().BoolVar(&dryRunDelete, "dry-run", false, "Show what would be deleted without deleting")
	deleteCmd.Flags().BoolVarP(&yesDelete, "yes", "y", false, "Skip confirmation prompt")
	deleteCmd.Flags().IntVar(&deleteTimeoutFlag, "timeout", 0, "Override git operation timeout (seconds)")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	var branchName string

	// Find project root first (needed for interactive mode)
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "delete", nil, ui.NewCLIError(ui.ErrCodeNotInProject, "not in a git-wt project"))
		}
		return fmt.Errorf("not in a git-wt project: %w", err)
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
		branchName = args[0]
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
			options = append(options, huh.NewOption(wt.Branch, wt.Branch))
		}

		if len(options) == 0 {
			fmt.Println("No worktrees to delete (only default branch exists)")
			return nil
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select worktree to delete").
					Options(options...).
					Value(&branchName),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}
	}

	// Use flattened branch name for directory path
	worktreeDir := git.FlattenBranchName(branchName)
	worktreePath := filepath.Join(projectRoot, worktreeDir)

	// Check if worktree exists
	if _, err := os.Stat(worktreePath); os.IsNotExist(err) {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "delete", nil, ui.NewCLIError(ui.ErrCodeNotFound, fmt.Sprintf("worktree not found: %s", branchName)))
		}
		return fmt.Errorf("worktree not found: %s", branchName)
	}

	// Dry run mode
	if dryRunDelete {
		status, _ := git.GetWorktreeStatus(worktreePath)
		if IsJSONOutput() {
			data := DeleteData{
				Branch: branchName,
				Path:   worktreePath,
				DryRun: true,
				Status: status,
			}
			return ui.OutputJSON(os.Stdout, "delete", data, nil)
		}
		fmt.Println(ui.InfoMsg("Dry run - would delete:"))
		fmt.Printf("  Worktree: %s\n", worktreePath)
		fmt.Printf("  Branch: %s\n", branchName)
		if status != "clean" {
			fmt.Println(ui.WarningMsg(fmt.Sprintf("  Status: %s", status)))
		}
		return nil
	}

	// Check for uncommitted changes
	status, _ := git.GetWorktreeStatus(worktreePath)
	if status != "clean" && !forceDelete {
		// Dirty worktrees require --force flag
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "delete", nil, ui.NewCLIError(ui.ErrCodeValidation, fmt.Sprintf("worktree has uncommitted changes, use --force to delete (status: %s)", status)))
		}

		fmt.Println(ui.WarningMsg(fmt.Sprintf("%s has uncommitted changes:", branchName)))

		// Show changed files
		output, _ := git.RunInDirWithTimeout(worktreePath, cfg.GitTimeout, "status", "--porcelain")
		for _, line := range splitByNewline(output) {
			fmt.Println("  " + line)
		}
		fmt.Println()
		fmt.Println("Use --force to delete worktrees with uncommitted changes.")
		return nil
	}

	// Confirmation prompt (skip with --yes or --json)
	if !yesDelete && !IsJSONOutput() {
		title := fmt.Sprintf("Delete worktree '%s'?", branchName)
		affirmative := "Yes, delete"
		if status != "clean" {
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
			return err
		}

		if !confirm {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	if !IsJSONOutput() {
		fmt.Println(ui.SubtleStyle.Render("Deleting worktree..."))
	}

	// Remove worktree
	var removeErr error
	if forceDelete {
		removeErr = git.RemoveWorktreeForce(projectRoot, worktreePath)
	} else {
		removeErr = git.RemoveWorktree(projectRoot, worktreePath)
	}

	if removeErr != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "delete", nil, ui.NewCLIError(ui.ErrCodeGit, removeErr.Error()))
		}
		return removeErr
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
		data := DeleteData{
			Branch:        branchName,
			Path:          worktreePath,
			BranchDeleted: branchDeleted,
		}
		return ui.OutputJSON(os.Stdout, "delete", data, nil)
	}

	return nil
}

func splitByNewline(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSpace(s), "\n")
}
