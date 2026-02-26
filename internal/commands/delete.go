package commands

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
	pathFlag          string
)

var deleteCmd = &cobra.Command{
	Use:     "remove [branch...]",
	Aliases: []string{"delete", "rm"},
	Short:   "Remove worktrees and their branches",
	Long: `Remove one or more worktrees and their associated branches.

In interactive mode (no arguments), presents a multi-select list.
Use space to select/deselect, enter to confirm.`,
	Args:              cobra.ArbitraryArgs,
	RunE:              runDelete,
	ValidArgsFunction: completeWorktreeBranches,
}

func init() {
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Force delete even with uncommitted changes")
	deleteCmd.Flags().BoolVarP(&dryRunDelete, "dry-run", "d", false, "Show what would be deleted without deleting")
	deleteCmd.Flags().BoolVarP(&yesDelete, "yes", "y", false, "Skip confirmation prompt")
	deleteCmd.Flags().IntVar(&deleteTimeoutFlag, "timeout", 0, "Override git operation timeout (seconds)")
	deleteCmd.Flags().StringVarP(&pathFlag, "path", "p", "", "Delete worktree by path instead of branch name")
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

	if pathFlag != "" {
		if len(args) > 0 {
			errMsg := "--path and branch arguments cannot be used together"
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "delete", nil, ui.NewCLIError(ui.ErrCodeValidation, errMsg))
			}
			return errors.New(errMsg)
		}

		absPath, err := filepath.Abs(pathFlag)
		if err != nil {
			errMsg := fmt.Sprintf("invalid path: %v", err)
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "delete", nil, ui.NewCLIError(ui.ErrCodeValidation, errMsg))
			}
			return errors.New(errMsg)
		}

		branch, err := resolveBranchFromPath(projectRoot, absPath)
		if err != nil {
			errMsg := fmt.Sprintf("worktree not found at path: %s", absPath)
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "delete", nil, ui.NewCLIError(ui.ErrCodeNotFound, errMsg))
			}
			return errors.New(errMsg)
		}

		branchNames = []string{branch}
	} else if len(args) > 0 {
		branchNames = args
	} else {
		errMsg := "branch name(s) required. Usage: wt delete <branch> [<branch>...]"
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "delete", nil,
				ui.NewCLIError(ui.ErrCodeValidation, errMsg))
		}
		return errors.New(errMsg)
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
	// Check if worktree exists in git's list (not just if directory exists)
	worktrees, err := git.ListWorktrees(projectRoot)
	if err != nil {
		return nil, err
	}

	// Find the worktree by branch name and get its actual path
	// (don't assume path from flattened branch name - user may have custom directory)
	var worktreePath string
	var worktreeExists bool
	var directoryMissing bool
	for _, wt := range worktrees {
		if wt.Branch == branchName {
			worktreeExists = true
			worktreePath = wt.Path // Use actual path from git worktree list
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
		errMsg := fmt.Sprintf("worktree '%s' has uncommitted changes. Use --force to delete", branchName)
		if IsJSONOutput() {
			return nil, ui.NewCLIError(ui.ErrCodeValidation, fmt.Sprintf("worktree has uncommitted changes, use --force to delete (status: %s)", status))
		}
		return nil, errors.New(errMsg)
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
