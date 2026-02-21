package commands

import (
	"os"
	"strings"

	"github.com/raisedadead/wt/internal/config"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/hooks"
	"github.com/spf13/cobra"
)

// completeWorktreeBranches returns branch names of existing worktrees for shell completion
func completeWorktreeBranches(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	worktrees, err := git.ListWorktrees(projectRoot)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var branches []string
	for _, wt := range worktrees {
		if wt.IsBare || wt.Branch == "" {
			continue
		}
		// Match against both the original branch name and the flattened form,
		// but only suggest the original branch name to avoid duplicates
		flat := git.FlattenBranchName(wt.Branch)
		if strings.HasPrefix(wt.Branch, toComplete) || strings.HasPrefix(flat, toComplete) {
			branches = append(branches, wt.Branch)
		}
	}

	return branches, cobra.ShellCompDirectiveNoFileComp
}

// completeGitBranches returns local git branch names for shell completion
func completeGitBranches(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectRoot, err := git.GetProjectRoot(".")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	output, err := git.RunInDir(projectRoot, "branch", "--format=%(refname:short)")
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" && strings.HasPrefix(line, toComplete) {
			branches = append(branches, line)
		}
	}

	return branches, cobra.ShellCompDirectiveNoFileComp
}

// completeWorkflowNames returns workflow names from config for shell completion
func completeWorkflowNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectRoot := ""
	if pr, err := git.GetProjectRoot("."); err == nil {
		projectRoot = pr
	}

	cfg, _, err := config.LoadEffective(config.GetConfigPath(), projectRoot)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	var names []string
	for name := range cfg.Workflows {
		if strings.HasPrefix(name, toComplete) {
			names = append(names, name)
		}
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeHookNames returns available hook script names for shell completion
func completeHookNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	communityDir := config.GetCommunityHooksDir()
	customDir := config.GetCustomHooksDir()

	var names []string
	seen := make(map[string]bool)

	for _, dir := range []string{customDir, communityDir} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sh") {
				name := strings.TrimSuffix(entry.Name(), ".sh")
				if !seen[name] && strings.HasPrefix(name, toComplete) {
					names = append(names, name)
					seen[name] = true
				}
			}
		}
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeEnabledHookNames returns hook names currently enabled in config
func completeEnabledHookNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	projectRoot := ""
	if pr, err := git.GetProjectRoot("."); err == nil {
		projectRoot = pr
	}

	cfg, _, err := config.LoadEffective(config.GetConfigPath(), projectRoot)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	seen := make(map[string]bool)
	var names []string

	for _, hookList := range [][]string{cfg.Hooks.PostClone, cfg.Hooks.PostAdd} {
		for _, name := range hookList {
			// Only include script-based hooks (not inline commands)
			_, isScript := hooks.ResolveHook(name, config.GetCustomHooksDir(), config.GetCommunityHooksDir())
			if isScript && !seen[name] && strings.HasPrefix(name, toComplete) {
				names = append(names, name)
				seen[name] = true
			}
		}
	}

	return names, cobra.ShellCompDirectiveNoFileComp
}

// completeShellNames returns supported shell names for completion command
func completeShellNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	shells := []string{"bash", "zsh", "fish", "powershell"}
	var matches []string
	for _, s := range shells {
		if strings.HasPrefix(s, toComplete) {
			matches = append(matches, s)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}
