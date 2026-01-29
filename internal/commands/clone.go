package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/raisedadead/git-wt/internal/config"
	"github.com/raisedadead/git-wt/internal/git"
	"github.com/raisedadead/git-wt/internal/hooks"
	"github.com/raisedadead/git-wt/internal/ui"
	"github.com/spf13/cobra"
)

var forceClone bool

// CloneData represents the JSON output for the clone command
type CloneData struct {
	Project       string `json:"project"`
	Path          string `json:"path"`
	BarePath      string `json:"bare_path"`
	DefaultBranch string `json:"default_branch"`
	WorktreePath  string `json:"worktree_path"`
}

var cloneCmd = &cobra.Command{
	Use:   "clone <repo> [name] [-- <git-args>]",
	Short: "Clone a repository as a bare repo with worktree structure",
	Long: `Clone a repository as a bare repo and set up the worktree structure.

Supports GitHub shorthand (like gh CLI):
  git wt clone owner/repo
  git wt clone freeCodeCamp/freeCodeCamp

Or full URLs:
  git wt clone git@github.com:owner/repo.git
  git wt clone https://github.com/owner/repo.git

Passthrough git flags after --:
  git wt clone owner/repo -- --depth=1
  git wt clone owner/repo -- --single-branch

This creates:
  <name>/
  ├── .bare/     (bare git repository)
  ├── .git       (pointer to .bare)
  └── main/      (default worktree)`,
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: false,
	RunE:               runClone,
}

func init() {
	cloneCmd.Flags().BoolVarP(&forceClone, "force", "f", false, "Remove existing directory and re-clone")
	rootCmd.AddCommand(cloneCmd)
}

func runClone(cmd *cobra.Command, args []string) error {
	var url, name string
	var gitArgs []string

	// Parse args: split at "--" separator using Cobra's ArgsLenAtDash
	// Before "--": positional args (url, name)
	// After "--": passthrough git args
	dashIdx := cmd.ArgsLenAtDash()
	if dashIdx >= 0 {
		gitArgs = args[dashIdx:]
		args = args[:dashIdx]
	}

	// Get URL
	if len(args) >= 1 {
		url = args[0]
	} else {
		// Interactive mode - skip if JSON output
		if IsJSONOutput() {
			err := ui.NewCLIError(ui.ErrCodeValidation, "repository URL is required")
			return ui.OutputJSON(os.Stdout, "clone", nil, err)
		}
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Repository URL").
					Placeholder("git@github.com:user/repo.git").
					Value(&url),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}
	}

	if url == "" {
		if IsJSONOutput() {
			err := ui.NewCLIError(ui.ErrCodeValidation, "repository URL is required")
			return ui.OutputJSON(os.Stdout, "clone", nil, err)
		}
		return fmt.Errorf("repository URL is required")
	}

	// Expand shorthand (owner/repo) to full URL like gh CLI
	url = expandRepoShorthand(url)

	// Get name (extract from URL if not provided)
	if len(args) >= 2 {
		name = args[1]
	} else {
		// Extract default name from URL
		defaultName := extractRepoName(url)

		if len(args) == 0 {
			// Interactive mode - skip if JSON output, use default
			if IsJSONOutput() {
				name = defaultName
			} else {
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("Project name").
							Placeholder(defaultName).
							Value(&name),
					),
				)

				if err := form.Run(); err != nil {
					return err
				}

				if name == "" {
					name = defaultName
				}
			}
		} else {
			name = defaultName
		}
	}

	// Validate project name for safety
	if err := git.ValidateProjectName(name); err != nil {
		return fmt.Errorf("invalid project name: %w", err)
	}

	// Load config
	cfg, err := config.LoadGlobal()
	if err != nil {
		return err
	}

	// Determine target directory
	targetDir := filepath.Join(cfg.WorktreeRoot, name)

	// Handle existing directory
	if _, err := os.Stat(targetDir); err == nil {
		if forceClone {
			if !IsJSONOutput() {
				fmt.Println(ui.WarningMsg(fmt.Sprintf("Removing existing directory: %s", targetDir)))
			}
			if err := os.RemoveAll(targetDir); err != nil {
				if IsJSONOutput() {
					return ui.OutputJSON(os.Stdout, "clone", nil, ui.NewCLIError(ui.ErrCodeGit, fmt.Sprintf("failed to remove existing directory: %v", err)))
				}
				return fmt.Errorf("failed to remove existing directory: %w", err)
			}
		} else {
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "clone", nil, ui.NewCLIError(ui.ErrCodeAlreadyExists, fmt.Sprintf("directory already exists: %s (use --force to overwrite)", targetDir)))
			}
			return fmt.Errorf("directory already exists: %s (use --force to overwrite)", targetDir)
		}
	}

	// Create target directory atomically (avoids TOCTOU race)
	// os.Mkdir fails if directory already exists
	if err := os.Mkdir(targetDir, 0755); err != nil {
		// Parent directory might not exist, try to create it
		if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory: %w", err)
		}
		// Try again
		if err := os.Mkdir(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	if !IsJSONOutput() {
		fmt.Println(ui.SubtleStyle.Render("Cloning repository..."))
	}

	// Clone as bare (pass through any extra git args)
	if err := git.BareClone(url, targetDir, gitArgs...); err != nil {
		os.RemoveAll(targetDir) // Clean up on failure
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "clone", nil, ui.NewCLIError(ui.ErrCodeGit, err.Error()))
		}
		return err
	}
	if !IsJSONOutput() {
		fmt.Println(ui.SuccessMsg("Bare clone complete"))
	}

	// Get default branch
	defaultBranch, err := git.GetDefaultBranch(targetDir)
	if err != nil {
		defaultBranch = git.DefaultBranch
	}

	// Create main worktree
	mainPath, err := git.CreateWorktreeFromBranch(targetDir, defaultBranch)
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "clone", nil, ui.NewCLIError(ui.ErrCodeGit, fmt.Sprintf("failed to create main worktree: %v", err)))
		}
		return fmt.Errorf("failed to create main worktree: %w", err)
	}
	if !IsJSONOutput() {
		fmt.Println(ui.SuccessMsg(fmt.Sprintf("Created %s/ worktree", defaultBranch)))
	}

	// Run post_clone hooks
	hookCtx := hooks.Context{
		Path:          mainPath,
		Branch:        defaultBranch,
		ProjectRoot:   targetDir,
		DefaultBranch: defaultBranch,
	}
	if warnings := hooks.Run(cfg.Hooks.PostClone, hookCtx); len(warnings) > 0 {
		for _, w := range warnings {
			if !IsJSONOutput() {
				fmt.Println(ui.WarningMsg("Hook: " + w))
			}
		}
	}

	// JSON output
	if IsJSONOutput() {
		data := CloneData{
			Project:       name,
			Path:          targetDir,
			BarePath:      filepath.Join(targetDir, ".bare"),
			DefaultBranch: defaultBranch,
			WorktreePath:  mainPath,
		}
		return ui.OutputJSON(os.Stdout, "clone", data, nil)
	}

	fmt.Println()
	fmt.Println(ui.BoldStyle.Render("cd " + mainPath))

	return nil
}

// expandRepoShorthand expands owner/repo shorthand to full GitHub URL
// Supports: owner/repo -> git@github.com:owner/repo.git
// Passes through full URLs unchanged
func expandRepoShorthand(input string) string {
	// Already a full URL (HTTPS or other protocol)
	if strings.Contains(input, "://") {
		return input
	}

	// Already an SSH URL (git@...)
	if strings.HasPrefix(input, "git@") {
		return input
	}

	// Check if it looks like owner/repo (exactly one slash, no special chars)
	parts := strings.Split(input, "/")
	if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
		// Looks like owner/repo shorthand
		owner := parts[0]
		repo := strings.TrimSuffix(parts[1], ".git")
		return fmt.Sprintf("git@github.com:%s/%s.git", owner, repo)
	}

	// Return as-is (might be a local path or other format)
	return input
}

// extractRepoName extracts the repository name from a URL
func extractRepoName(url string) string {
	// Handle SSH URLs: git@github.com:user/repo.git
	if strings.Contains(url, ":") && !strings.Contains(url, "://") {
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			name := parts[len(parts)-1]
			return strings.TrimSuffix(name, ".git")
		}
	}

	// Handle HTTPS URLs: https://github.com/user/repo.git
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		return strings.TrimSuffix(name, ".git")
	}

	return "repo"
}
