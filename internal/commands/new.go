package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/raisedadead/wt/internal/config"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/hooks"
	"github.com/raisedadead/wt/internal/ui"
	"github.com/spf13/cobra"
)

// NewData represents the JSON output for the new command
type NewData struct {
	Branch        string            `json:"branch"`
	Path          string            `json:"path"`
	Workflow      string            `json:"workflow,omitempty"`
	BaseBranch    string            `json:"base_branch,omitempty"`
	TrackedRemote string            `json:"tracked_remote,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	HookWarnings  []string          `json:"hook_warnings,omitempty"`
}

var (
	// Workflow flags
	featureFlag  bool
	bugfixFlag   bool
	prReviewFlag bool
	workflowFlag string

	// Hook input flags
	issueNum int
	prNum    int

	// Other flags
	baseFlag           string
	remoteFlag         string
	newTimeoutFlag     int
	newHookTimeoutFlag int
	trackFlag          bool
	newFlag            bool
	fetchFlag          bool
	noHooksFlag        bool
)

var newCmd = &cobra.Command{
	Use:     "add [branch]",
	Aliases: []string{"new"},
	Short:   "Create a new worktree",
	Long: `Create a new worktree with optional workflow support.

Workflows:
  --feature, -f    New feature development (branch: feat/{slug})
  --bugfix, -b     Bug fix (branch: fix/{slug})
  --pr-review      Review a pull request (uses PR's actual branch)
  --workflow, -w   Use a custom workflow from config

Hook Inputs:
  --issue <n>      Pass issue number to workflow hooks
  --pr <n>         Pass PR number to workflow hooks

Examples:
  git wt add my-branch              # Plain branch
  git wt add --feature auth         # Feature workflow
  git wt add -b --issue 42          # Bugfix linked to issue
  git wt add --pr-review 123        # Review PR #123`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNew,
}

func init() {
	// Workflow flags
	newCmd.Flags().BoolVarP(&featureFlag, "feature", "f", false, "Use feature workflow")
	newCmd.Flags().BoolVarP(&bugfixFlag, "bugfix", "b", false, "Use bugfix workflow")
	newCmd.Flags().BoolVar(&prReviewFlag, "pr-review", false, "Use pr-review workflow")
	newCmd.Flags().StringVarP(&workflowFlag, "workflow", "w", "", "Use named workflow from config")

	// Hook input flags
	newCmd.Flags().IntVar(&issueNum, "issue", 0, "Pass issue number to hooks")
	newCmd.Flags().IntVar(&prNum, "pr", 0, "Pass PR number to hooks")

	// Other flags
	newCmd.Flags().StringVar(&baseFlag, "base", "", "Base branch to create worktree from (default: HEAD)")
	newCmd.Flags().StringVar(&remoteFlag, "remote", "", "Override default remote")
	newCmd.Flags().IntVar(&newTimeoutFlag, "timeout", 0, "Override git operation timeout (seconds)")
	newCmd.Flags().IntVar(&newHookTimeoutFlag, "hook-timeout", 0, "Override hook timeout (seconds)")
	newCmd.Flags().BoolVar(&trackFlag, "track", false, "Track existing remote branch")
	newCmd.Flags().BoolVar(&newFlag, "new", false, "Force create new local branch even if remote exists")
	newCmd.Flags().BoolVar(&fetchFlag, "fetch", false, "Fetch all remotes before checking for branches")
	newCmd.Flags().BoolVar(&noHooksFlag, "no-hooks", false, "Skip all hooks")

	rootCmd.AddCommand(newCmd)
}

// determineWorkflow returns the workflow name based on flags
func determineWorkflow() string {
	if featureFlag {
		return "feature"
	}
	if bugfixFlag {
		return "bugfix"
	}
	if prReviewFlag {
		return "pr-review"
	}
	if workflowFlag != "" {
		return workflowFlag
	}
	return ""
}

// getWorkflowPrefix extracts the static prefix from a branch template
// Returns empty string if the prefix contains placeholders
func getWorkflowPrefix(template string) string {
	// Template format: "feat/{slug}" -> prefix is "feat"
	var prefix string
	if idx := strings.Index(template, "/"); idx > 0 {
		prefix = template[:idx]
	} else if idx := strings.Index(template, "-"); idx > 0 {
		prefix = template[:idx]
	}

	// If the prefix contains a placeholder, it's not a static prefix
	if strings.Contains(prefix, "{") {
		return ""
	}
	return prefix
}

// applyBranchTemplate applies the workflow branch template
func applyBranchTemplate(template, name string, metadata map[string]string) string {
	result := template

	// Strip workflow prefix from input if already present to avoid duplication
	// e.g., if template is "fix/{slug}" and user enters "fix/my-bug", use "my-bug"
	prefix := getWorkflowPrefix(template)
	if prefix != "" {
		prefixWithSlash := prefix + "/"
		prefixWithDash := prefix + "-"
		if strings.HasPrefix(strings.ToLower(name), prefixWithSlash) {
			name = name[len(prefixWithSlash):]
		} else if strings.HasPrefix(strings.ToLower(name), prefixWithDash) {
			name = name[len(prefixWithDash):]
		}
	}

	// Replace template variables
	// {name} and {branch} pass through the input as-is
	// {slug} converts to a URL-friendly slug
	result = strings.ReplaceAll(result, "{name}", name)
	result = strings.ReplaceAll(result, "{branch}", name)
	result = strings.ReplaceAll(result, "{slug}", slugify(name))

	// Replace metadata variables
	for k, v := range metadata {
		result = strings.ReplaceAll(result, "{"+k+"}", v)
	}

	return result
}

// slugify converts a string to a URL-friendly slug
func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")

	// Keep only alphanumeric and hyphens
	var result strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	s = result.String()

	// Remove multiple hyphens
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")

	// Limit length
	if len(s) > 50 {
		s = s[:50]
		s = strings.TrimSuffix(s, "-")
	}

	return s
}

func runNew(cmd *cobra.Command, args []string) error {
	// Find project root
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeNotInProject, "not in a wt project"))
		}
		return fmt.Errorf("not in a wt project: %w", err)
	}

	// Load config with repo-level overrides
	cfg, err := config.LoadWithRepo(config.GetConfigPath(), projectRoot)
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeGit, err.Error()))
		}
		return err
	}

	// Apply flag overrides
	if remoteFlag != "" {
		cfg.DefaultRemote = remoteFlag
	}
	if newTimeoutFlag > 0 {
		cfg.GitTimeout = newTimeoutFlag
	}
	if newHookTimeoutFlag > 0 {
		cfg.HookTimeout = newHookTimeoutFlag
	}

	// Determine workflow
	workflowName := determineWorkflow()
	var workflow *config.Workflow

	// Interactive mode if no workflow and no branch specified
	if workflowName == "" && len(args) == 0 && !IsJSONOutput() {
		var workType string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What are you working on?").
					Options(
						huh.NewOption("New feature", "feature"),
						huh.NewOption("Bug fix", "bugfix"),
						huh.NewOption("Review a PR", "pr-review"),
						huh.NewOption("Just a branch", "branch"),
					).
					Value(&workType),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}
		workflowName = workType
	}

	// Get workflow config if specified
	if workflowName != "" {
		workflow = cfg.GetWorkflow(workflowName)
		if workflow == nil {
			errMsg := fmt.Sprintf("unknown workflow: %s", workflowName)
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeValidation, errMsg))
			}
			return errors.New(errMsg)
		}
	}

	// Get default branch for hooks context
	defaultBranchName, err := git.GetDefaultBranch(projectRoot)
	if err != nil {
		defaultBranchName = git.DefaultBranch
	}

	// Build hook context
	hookCtx := hooks.Context{
		ProjectRoot:   projectRoot,
		DefaultBranch: defaultBranchName,
		Workflow:      workflowName,
		IssueNumber:   issueNum,
		PRNumber:      prNum,
		Metadata:      make(map[string]string),
	}

	if workflow != nil {
		hookCtx.WorkflowPrefix = getWorkflowPrefix(workflow.BranchTemplate)
	}

	// Run pre_create hooks if workflow has them
	var hookOutput *hooks.HookOutput
	if workflow != nil && len(workflow.Hooks.PreCreate) > 0 && !noHooksFlag {
		hookOutput, err = hooks.RunWorkflowHooks(
			workflow.Hooks.PreCreate,
			hookCtx,
			cfg.HookTimeout,
			config.GetCustomHooksDir(),
			config.GetCommunityHooksDir(),
		)
		if err != nil {
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeGitHub, err.Error()))
			}
			return err
		}

		// Check for hook error
		if hookOutput.Error != "" {
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeGitHub, hookOutput.Error))
			}
			return errors.New(hookOutput.Error)
		}

		// Merge hook metadata into context
		for k, v := range hookOutput.Metadata {
			hookCtx.Metadata[k] = v
		}
	}

	// Determine branch name
	var branchName string

	if hookOutput != nil && hookOutput.Branch != "" {
		// Use branch from hook
		branchName = hookOutput.Branch

		// Prompt to confirm/edit in interactive mode
		if !IsJSONOutput() {
			var confirmedBranch string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Branch name").
						Value(&confirmedBranch).
						Placeholder(branchName),
				),
			)

			if err := form.Run(); err != nil {
				return err
			}

			if confirmedBranch != "" {
				branchName = confirmedBranch
			}
		}
	} else if len(args) > 0 {
		// Use positional argument
		branchName = args[0]

		// Apply workflow template if available
		if workflow != nil && workflow.BranchTemplate != "" {
			branchName = applyBranchTemplate(workflow.BranchTemplate, args[0], hookCtx.Metadata)
		}
	} else if workflow != nil && !IsJSONOutput() {
		// Prompt for branch name in interactive mode
		var inputName string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Branch name (or description)").
					Value(&inputName),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		if inputName == "" {
			return errors.New("branch name is required")
		}

		branchName = applyBranchTemplate(workflow.BranchTemplate, inputName, hookCtx.Metadata)
	} else {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeValidation, "branch name is required"))
		}
		return errors.New("branch name is required")
	}

	// Validate branch name
	if err := git.ValidateBranchName(branchName); err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeValidation, fmt.Sprintf("invalid branch name: %v", err)))
		}
		return fmt.Errorf("invalid branch name: %w", err)
	}

	// Update hook context with final branch name
	hookCtx.Branch = branchName

	// Check if we should track remote (from hook metadata or flags)
	shouldTrack := trackFlag
	if hookCtx.Metadata["track_remote"] == "true" {
		shouldTrack = true
	}

	// Check for remote branches (unless --new or --base is specified)
	var trackedRemote string
	if !newFlag && baseFlag == "" {
		// Optionally fetch all remotes first
		if fetchFlag {
			if !IsJSONOutput() {
				fmt.Println(ui.SubtleStyle.Render("Fetching from all remotes..."))
			}
			if err := git.FetchAllRemotes(projectRoot); err != nil {
				if !IsJSONOutput() {
					fmt.Println(ui.WarningMsg(fmt.Sprintf("Failed to fetch remotes: %v (continuing anyway)", err)))
				}
			}
		}

		// Check for matching remote branches
		remoteBranches, err := git.FindRemoteBranches(projectRoot, branchName)
		if err != nil {
			if !IsJSONOutput() {
				fmt.Println(ui.WarningMsg(fmt.Sprintf("Could not check remote branches: %v", err)))
			}
		} else if len(remoteBranches) == 0 && shouldTrack && !prReviewFlag {
			// --track was specified but no remote branch found (skip for pr-review as we may need to fetch PR)
			errMsg := fmt.Sprintf("branch %q not found on any remote. Use --fetch to update remote refs, or omit --track to create a new local branch", branchName)
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeValidation, errMsg))
			}
			return errors.New(errMsg)
		} else if len(remoteBranches) == 1 {
			remote := remoteBranches[0].Remote

			// If --remote flag specified and doesn't match, error
			if remoteFlag != "" && remoteFlag != remote {
				errMsg := fmt.Sprintf("branch %q found on remote %q, not %q", branchName, remote, remoteFlag)
				if IsJSONOutput() {
					return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeValidation, errMsg))
				}
				return errors.New(errMsg)
			}

			// Decide whether to track
			if shouldTrack || (cfg.AutoTrack != nil && *cfg.AutoTrack) {
				trackedRemote = remote
			} else if IsJSONOutput() {
				// In JSON mode, require explicit --track or --new
				errMsg := fmt.Sprintf("branch %q exists on remote %q. Use --track to track it or --new to create a new local branch", branchName, remote)
				return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeValidation, errMsg))
			} else {
				// Interactive mode - prompt user
				fmt.Println(ui.WarningMsg(fmt.Sprintf("Branch %q found on remote %q", branchName, remote)))
				var choice string
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewSelect[string]().
							Title("What would you like to do?").
							Options(
								huh.NewOption(fmt.Sprintf("Track remote branch (%s/%s)", remote, branchName), "track"),
								huh.NewOption("Create new local branch (ignore remote)", "new"),
							).
							Value(&choice),
					),
				)

				if err := form.Run(); err != nil {
					return err
				}

				if choice == "track" {
					trackedRemote = remote
				}
			}
		} else if len(remoteBranches) > 1 {
			// Multiple remotes have this branch
			if remoteFlag != "" && shouldTrack {
				// User specified both --remote and --track, check if remote exists in list
				found := false
				for _, rb := range remoteBranches {
					if rb.Remote == remoteFlag {
						found = true
						trackedRemote = remoteFlag
						break
					}
				}
				if !found {
					var remoteNames []string
					for _, rb := range remoteBranches {
						remoteNames = append(remoteNames, rb.Remote)
					}
					errMsg := fmt.Sprintf("branch %q not found on remote %q (found on: %s)", branchName, remoteFlag, strings.Join(remoteNames, ", "))
					if IsJSONOutput() {
						return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeValidation, errMsg))
					}
					return errors.New(errMsg)
				}
			} else {
				// Ambiguous - error out
				var remoteNames []string
				for _, rb := range remoteBranches {
					remoteNames = append(remoteNames, rb.Remote)
				}
				errMsg := fmt.Sprintf("branch %q exists on multiple remotes: %s. Use --remote <name> --track to specify which to track, or --new to create a new local branch", branchName, strings.Join(remoteNames, ", "))
				if IsJSONOutput() {
					return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeValidation, errMsg))
				}
				return errors.New(errMsg)
			}
		}
	}

	if !IsJSONOutput() {
		fmt.Println(ui.SubtleStyle.Render("Creating worktree..."))
	}

	// Create the worktree (with optional base branch or remote tracking)
	var worktreePath string
	if trackedRemote != "" {
		worktreePath, err = git.CreateWorktreeFromRemote(projectRoot, branchName, trackedRemote)
	} else if baseFlag != "" {
		worktreePath, err = git.CreateWorktreeWithBase(projectRoot, branchName, baseFlag)
	} else {
		worktreePath, err = git.CreateWorktree(projectRoot, branchName)
	}
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeGit, err.Error()))
		}
		return err
	}

	// Update hook context with worktree path
	hookCtx.Path = worktreePath

	// Get flattened directory name for display
	worktreeDir := git.FlattenBranchName(branchName)
	if !IsJSONOutput() {
		if trackedRemote != "" {
			fmt.Println(ui.SuccessMsg(fmt.Sprintf("Created %s/ worktree (tracking %s/%s)", worktreeDir, trackedRemote, branchName)))
		} else if baseFlag != "" {
			fmt.Println(ui.SuccessMsg(fmt.Sprintf("Created %s/ worktree (from %s)", worktreeDir, baseFlag)))
		} else {
			fmt.Println(ui.SuccessMsg(fmt.Sprintf("Created %s/ worktree", worktreeDir)))
		}
	}

	// Collect all hook warnings
	var allWarnings []string
	if hookOutput != nil {
		allWarnings = append(allWarnings, hookOutput.Warnings...)
	}

	// Run post_add hooks
	if !noHooksFlag {
		var postAddHooks []string

		// Use workflow-specific hooks if available, otherwise fall back to global
		if workflow != nil && len(workflow.Hooks.PostAdd) > 0 {
			postAddHooks = workflow.Hooks.PostAdd
		} else {
			postAddHooks = cfg.Hooks.PostAdd
		}

		if len(postAddHooks) > 0 {
			hookWarnings := hooks.RunResolved(
				postAddHooks,
				hookCtx,
				cfg.HookTimeout,
				config.GetCustomHooksDir(),
				config.GetCommunityHooksDir(),
			)
			allWarnings = append(allWarnings, hookWarnings...)
		}
	}

	// Show warnings
	if len(allWarnings) > 0 {
		for _, w := range allWarnings {
			if !IsJSONOutput() {
				fmt.Println(ui.WarningMsg("Hook: " + w))
			}
		}
	}

	// JSON output
	if IsJSONOutput() {
		data := NewData{
			Branch:        branchName,
			Path:          worktreePath,
			Workflow:      workflowName,
			BaseBranch:    baseFlag,
			TrackedRemote: trackedRemote,
			Metadata:      hookCtx.Metadata,
			HookWarnings:  allWarnings,
		}
		return ui.OutputJSON(os.Stdout, "new", data, nil)
	}

	fmt.Println()
	fmt.Println(ui.BoldStyle.Render(fmt.Sprintf("cd %s", worktreePath)))

	return nil
}
