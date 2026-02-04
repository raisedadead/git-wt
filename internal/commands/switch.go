package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/ui"
	"github.com/spf13/cobra"
)

// SwitchData represents the JSON output for the switch command
type SwitchData struct {
	Branch string `json:"branch"`
	Path   string `json:"path"`
}

var switchCmd = &cobra.Command{
	Use:   "switch [branch]",
	Short: "Switch to a worktree (outputs path for cd)",
	Long: `Switch to a worktree by outputting its path.

Use with cd for seamless directory switching:
  cd $(wt switch branch)

Or create a shell alias for convenience:
  # Add to .bashrc/.zshrc:
  wts() { cd "$(wt switch "$@")" || return; }

In interactive mode (no arguments), presents a list to choose from.

The branch can be specified by:
  - Exact branch name: wt switch feature/auth
  - Flattened directory name: wt switch feature-auth`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSwitch,
}

func init() {
	rootCmd.AddCommand(switchCmd)
}

func runSwitch(cmd *cobra.Command, args []string) error {
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "switch", nil, ui.NewCLIError(ui.ErrCodeNotInProject, "not in a wt project"))
		}
		return fmt.Errorf("not in a wt project: %w", err)
	}

	worktrees, err := git.ListWorktrees(projectRoot)
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "switch", nil, ui.NewCLIError(ui.ErrCodeGit, err.Error()))
		}
		return err
	}

	// Filter out .bare directory
	var validWorktrees []git.Worktree
	for _, wt := range worktrees {
		if strings.HasSuffix(wt.Path, "/.bare") || wt.Branch == "" {
			continue
		}
		validWorktrees = append(validWorktrees, wt)
	}

	if len(validWorktrees) == 0 {
		errMsg := "no worktrees found"
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "switch", nil, ui.NewCLIError(ui.ErrCodeNotFound, errMsg))
		}
		return fmt.Errorf("%s", errMsg)
	}

	var selectedBranch string
	var selectedPath string

	if len(args) > 0 {
		// Find worktree by branch name or flattened directory name
		branchArg := args[0]
		for _, wt := range validWorktrees {
			// Match exact branch name
			if wt.Branch == branchArg {
				selectedBranch = wt.Branch
				selectedPath = wt.Path
				break
			}
			// Match flattened directory name (e.g., "feature-auth" matches "feature/auth")
			if git.FlattenBranchName(wt.Branch) == branchArg {
				selectedBranch = wt.Branch
				selectedPath = wt.Path
				break
			}
		}

		if selectedPath == "" {
			errMsg := fmt.Sprintf("worktree not found: %s", branchArg)
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "switch", nil, ui.NewCLIError(ui.ErrCodeNotFound, errMsg))
			}
			return fmt.Errorf("%s", errMsg)
		}
	} else {
		// Interactive mode
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "switch", nil,
				ui.NewCLIError(ui.ErrCodeValidation, "branch name is required"))
		}

		// Build options for selection
		var options []huh.Option[string]
		for _, wt := range validWorktrees {
			status, _ := git.GetWorktreeStatus(wt.Path)
			label := wt.Branch
			if status != "clean" {
				label = fmt.Sprintf("%s (%s)", wt.Branch, status)
			}
			options = append(options, huh.NewOption(label, wt.Branch))
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select worktree to switch to").
					Options(options...).
					Value(&selectedBranch),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		// Find the path for selected branch
		for _, wt := range validWorktrees {
			if wt.Branch == selectedBranch {
				selectedPath = wt.Path
				break
			}
		}
	}

	// JSON output
	if IsJSONOutput() {
		data := SwitchData{
			Branch: selectedBranch,
			Path:   selectedPath,
		}
		return ui.OutputJSON(os.Stdout, "switch", data, nil)
	}

	// Default: print just the path for cd substitution
	fmt.Println(selectedPath)
	return nil
}
