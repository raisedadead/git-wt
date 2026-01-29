package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/raisedadead/git-wt/internal/config"
	"github.com/raisedadead/git-wt/internal/git"
	"github.com/raisedadead/git-wt/internal/github"
	"github.com/raisedadead/git-wt/internal/hooks"
	"github.com/raisedadead/git-wt/internal/ui"
	"github.com/spf13/cobra"
)

// NewData represents the JSON output for the new command
type NewData struct {
	Branch     string     `json:"branch"`
	Path       string     `json:"path"`
	BaseBranch string     `json:"base_branch,omitempty"`
	Issue      *IssueData `json:"issue,omitempty"`
	PR         *PRData    `json:"pr,omitempty"`
}

// IssueData represents GitHub issue data for JSON output
type IssueData struct {
	Number int      `json:"number"`
	Title  string   `json:"title"`
	Labels []string `json:"labels,omitempty"`
}

// PRData represents GitHub PR data for JSON output
type PRData struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var (
	issueNum int
	prNum    int
	baseFlag string
)

var newCmd = &cobra.Command{
	Use:     "add [branch]",
	Aliases: []string{"new"},
	Short:   "Create a new worktree",
	Long: `Create a new worktree for a feature branch, GitHub issue, or pull request.

Examples:
  git wt add feature/auth
  git wt add --issue 42
  git wt add --pr 123`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNew,
}

func init() {
	newCmd.Flags().IntVar(&issueNum, "issue", 0, "Create worktree from GitHub issue number")
	newCmd.Flags().IntVar(&prNum, "pr", 0, "Create worktree from GitHub PR number")
	newCmd.Flags().StringVar(&baseFlag, "base", "", "Base branch to create worktree from (default: HEAD)")
	rootCmd.AddCommand(newCmd)
}

func runNew(cmd *cobra.Command, args []string) error {
	// Find project root
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeNotInProject, "not in a git-wt project"))
		}
		return fmt.Errorf("not in a git-wt project: %w", err)
	}

	// Load config
	cfg, err := config.LoadGlobal()
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeGit, err.Error()))
		}
		return err
	}

	var branchName string
	var issue *github.Issue
	var pr *github.PullRequest

	// Determine what we're creating
	if issueNum > 0 {
		// From issue
		issue, err = github.GetIssue(issueNum)
		if err != nil {
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeGitHub, err.Error()))
			}
			return err
		}

		branchName = github.GenerateBranchName("issue", issue.Number, issue.Title)
		if !IsJSONOutput() {
			fmt.Println(ui.SubtleStyle.Render(fmt.Sprintf("#%d - %s", issue.Number, issue.Title)))
			if len(issue.Labels) > 0 {
				fmt.Println(ui.SubtleStyle.Render("Labels: " + strings.Join(issue.GetLabelNames(), ", ")))
			}
			fmt.Println()
		}

	} else if prNum > 0 {
		// From PR
		pr, err = github.GetPullRequest(prNum)
		if err != nil {
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeGitHub, err.Error()))
			}
			return err
		}

		branchName = github.GenerateBranchName("pr", pr.Number, pr.Title)
		if !IsJSONOutput() {
			fmt.Println(ui.SubtleStyle.Render(fmt.Sprintf("#%d - %s", pr.Number, pr.Title)))
			fmt.Println(ui.SubtleStyle.Render(fmt.Sprintf("Author: @%s", pr.Author.Login)))
			fmt.Println()
		}

	} else if len(args) > 0 {
		// Direct branch name
		branchName = args[0]

	} else {
		// Interactive mode - skip if JSON output
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeValidation, "branch name is required (use positional arg, --issue, or --pr)"))
		}
		var workType string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What are you working on?").
					Options(
						huh.NewOption("New feature branch", "feature"),
						huh.NewOption("GitHub issue", "issue"),
						huh.NewOption("GitHub pull request", "pr"),
					).
					Value(&workType),
			),
		)

		if err := form.Run(); err != nil {
			return err
		}

		switch workType {
		case "issue":
			var issueInput string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Issue number").
						Value(&issueInput),
				),
			)

			if err := form.Run(); err != nil {
				return err
			}

			issueNum, err = strconv.Atoi(issueInput)
			if err != nil {
				return fmt.Errorf("invalid issue number: %s", issueInput)
			}

			issue, err = github.GetIssue(issueNum)
			if err != nil {
				return err
			}

			defaultBranch := github.GenerateBranchName("issue", issue.Number, issue.Title)
			fmt.Println(ui.SubtleStyle.Render(fmt.Sprintf("#%d - %s", issue.Number, issue.Title)))

			form = huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Branch name").
						Placeholder(defaultBranch).
						Value(&branchName),
				),
			)

			if err := form.Run(); err != nil {
				return err
			}

			if branchName == "" {
				branchName = defaultBranch
			}

		case "pr":
			var prInput string
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("PR number").
						Value(&prInput),
				),
			)

			if err := form.Run(); err != nil {
				return err
			}

			prNum, err = strconv.Atoi(prInput)
			if err != nil {
				return fmt.Errorf("invalid PR number: %s", prInput)
			}

			pr, err = github.GetPullRequest(prNum)
			if err != nil {
				return err
			}

			defaultBranch := github.GenerateBranchName("pr", pr.Number, pr.Title)
			fmt.Println(ui.SubtleStyle.Render(fmt.Sprintf("#%d - %s", pr.Number, pr.Title)))

			form = huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Branch name").
						Placeholder(defaultBranch).
						Value(&branchName),
				),
			)

			if err := form.Run(); err != nil {
				return err
			}

			if branchName == "" {
				branchName = defaultBranch
			}

		default:
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Branch name").
						Value(&branchName),
				),
			)

			if err := form.Run(); err != nil {
				return err
			}
		}
	}

	if branchName == "" {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeValidation, "branch name is required"))
		}
		return fmt.Errorf("branch name is required")
	}

	// Validate branch name
	if err := git.ValidateBranchName(branchName); err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "new", nil, ui.NewCLIError(ui.ErrCodeValidation, fmt.Sprintf("invalid branch name: %v", err)))
		}
		return fmt.Errorf("invalid branch name: %w", err)
	}

	if !IsJSONOutput() {
		fmt.Println(ui.SubtleStyle.Render("Creating worktree..."))
	}

	// Create the worktree (with optional base branch)
	var worktreePath string
	if baseFlag != "" {
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
	// Get flattened directory name for display
	worktreeDir := git.FlattenBranchName(branchName)
	if !IsJSONOutput() {
		if baseFlag != "" {
			fmt.Println(ui.SuccessMsg(fmt.Sprintf("Created %s/ worktree (from %s)", worktreeDir, baseFlag)))
		} else {
			fmt.Println(ui.SuccessMsg(fmt.Sprintf("Created %s/ worktree", worktreeDir)))
		}
	}

	// Get default branch name for hooks context
	defaultBranchName, err := git.GetDefaultBranch(projectRoot)
	if err != nil {
		defaultBranchName = git.DefaultBranch
	}

	// Run post_add hooks
	hookCtx := hooks.Context{
		Path:          worktreePath,
		Branch:        branchName,
		ProjectRoot:   projectRoot,
		DefaultBranch: defaultBranchName,
	}
	if warnings := hooks.Run(cfg.Hooks.PostAdd, hookCtx); len(warnings) > 0 {
		for _, w := range warnings {
			if !IsJSONOutput() {
				fmt.Println(ui.WarningMsg("Hook: " + w))
			}
		}
	}

	// JSON output
	if IsJSONOutput() {
		data := NewData{
			Branch:     branchName,
			Path:       worktreePath,
			BaseBranch: baseFlag,
		}
		if issue != nil {
			data.Issue = &IssueData{
				Number: issue.Number,
				Title:  issue.Title,
				Labels: issue.GetLabelNames(),
			}
		}
		if pr != nil {
			data.PR = &PRData{
				Number: pr.Number,
				Title:  pr.Title,
				Author: pr.Author.Login,
			}
		}
		return ui.OutputJSON(os.Stdout, "new", data, nil)
	}

	fmt.Println()
	fmt.Println(ui.BoldStyle.Render(fmt.Sprintf("cd %s", worktreePath)))

	return nil
}
