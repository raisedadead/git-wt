package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
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
	forceDelete  bool
	dryRunDelete bool
)

var deleteCmd = &cobra.Command{
	Use:     "delete <branch>",
	Aliases: []string{"rm"},
	Short:   "Remove a worktree and its branch",
	Args:    cobra.ExactArgs(1),
	RunE:    runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force delete even with uncommitted changes")
	deleteCmd.Flags().BoolVar(&dryRunDelete, "dry-run", false, "Show what would be deleted without deleting")
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
	branchName := args[0]

	// Find project root
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "delete", nil, ui.NewCLIError(ui.ErrCodeNotInProject, "not in a git-wt project"))
		}
		return fmt.Errorf("not in a git-wt project: %w", err)
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
		// In JSON mode, require --force for dirty worktrees
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "delete", nil, ui.NewCLIError(ui.ErrCodeValidation, fmt.Sprintf("worktree has uncommitted changes, use --force to delete (status: %s)", status)))
		}

		fmt.Println(ui.WarningMsg(fmt.Sprintf("%s has uncommitted changes:", branchName)))

		// Show changed files
		output, _ := git.RunInDir(worktreePath, "status", "--porcelain")
		for _, line := range splitByNewline(output) {
			fmt.Println("  " + line)
		}
		fmt.Println()

		var confirm bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Delete anyway?").
					Affirmative("Yes, discard changes").
					Negative("No, cancel").
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

		forceDelete = true
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
