package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/raisedadead/git-wt/internal/git"
	"github.com/raisedadead/git-wt/internal/ui"
	"github.com/spf13/cobra"
)

// RepairData represents the JSON output for the repair command
type RepairData struct {
	ProjectRoot string `json:"project_root"`
	Repaired    bool   `json:"repaired"`
	Output      string `json:"output,omitempty"`
}

var repairCmd = &cobra.Command{
	Use:   "repair",
	Short: "Repair worktree paths after moving a repository",
	Long: `Repair worktree administrative files after a repository has been moved.

This fixes broken worktree paths by updating the gitdir links between
the main repository and its worktrees. Run this command from within
any worktree after moving a git-wt managed repository.`,
	RunE: runRepair,
}

func init() {
	rootCmd.AddCommand(repairCmd)
}

func runRepair(cmd *cobra.Command, args []string) error {
	// Find project root
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "repair", nil, ui.NewCLIError(ui.ErrCodeNotInProject, "not in a git-wt project"))
		}
		return fmt.Errorf("not in a git-wt project: %w", err)
	}

	if !IsJSONOutput() {
		fmt.Println(ui.SubtleStyle.Render("Repairing worktree paths..."))
	}

	output, err := git.RepairWorktrees(projectRoot)
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "repair", nil, ui.NewCLIError(ui.ErrCodeGit, err.Error()))
		}
		return err
	}

	// Check if any repairs were made
	// git worktree repair only outputs text when repairs are made
	repaired := strings.TrimSpace(output) != ""

	if IsJSONOutput() {
		data := RepairData{
			ProjectRoot: projectRoot,
			Repaired:    repaired,
			Output:      strings.TrimSpace(output),
		}
		return ui.OutputJSON(os.Stdout, "repair", data, nil)
	}

	if repaired {
		fmt.Println(ui.SuccessMsg("Worktree paths repaired"))
		fmt.Println(ui.SubtleStyle.Render(output))
	} else {
		fmt.Println(ui.SuccessMsg("All worktree paths are correct"))
	}

	return nil
}
