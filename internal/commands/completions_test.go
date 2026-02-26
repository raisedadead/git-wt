package commands

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/raisedadead/wt/internal/git"
	"github.com/spf13/cobra"
)

// --- completeWorktreeBranches tests (unit, no real git repo) ---

// filterWorktreeBranches simulates the core logic of completeWorktreeBranches
// so we can unit-test matching without needing a real git repo.
func filterWorktreeBranches(worktrees []git.Worktree, toComplete string) []string {
	var branches []string
	for _, wt := range worktrees {
		if wt.IsBare || wt.Branch == "" {
			continue
		}
		flat := git.FlattenBranchName(wt.Branch)
		if strings.HasPrefix(wt.Branch, toComplete) || strings.HasPrefix(flat, toComplete) {
			branches = append(branches, wt.Branch)
		}
	}
	return branches
}

func TestFilterWorktreeBranches(t *testing.T) {
	t.Parallel()

	worktrees := []git.Worktree{
		{Path: "/p/main", Branch: "main"},
		{Path: "/p/feat-auth", Branch: "feat/auth"},
		{Path: "/p/feat-ui", Branch: "feat/ui"},
		{Path: "/p/fix-login", Branch: "fix/login"},
		{Path: "/p/hotfix", Branch: "hotfix"},
		{Path: "/p/.bare", Branch: "", IsBare: true},
		{Path: "/p/detached", Branch: ""},
	}

	tests := []struct {
		name       string
		toComplete string
		expected   []string
	}{
		{
			name:       "empty prefix returns all non-bare branches",
			toComplete: "",
			expected:   []string{"main", "feat/auth", "feat/ui", "fix/login", "hotfix"},
		},
		{
			name:       "exact branch prefix",
			toComplete: "feat/",
			expected:   []string{"feat/auth", "feat/ui"},
		},
		{
			name:       "flattened prefix matches original branches",
			toComplete: "feat-",
			expected:   []string{"feat/auth", "feat/ui"},
		},
		{
			name:       "partial prefix",
			toComplete: "fix",
			expected:   []string{"fix/login"},
		},
		{
			name:       "flattened partial prefix",
			toComplete: "fix-",
			expected:   []string{"fix/login"},
		},
		{
			name:       "exact match",
			toComplete: "main",
			expected:   []string{"main"},
		},
		{
			name:       "no match",
			toComplete: "nonexistent",
			expected:   nil,
		},
		{
			name:       "prefix h matches hotfix only",
			toComplete: "h",
			expected:   []string{"hotfix"},
		},
		{
			name:       "bare worktree excluded",
			toComplete: ".bare",
			expected:   nil,
		},
		{
			name:       "empty branch excluded",
			toComplete: "detached",
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := filterWorktreeBranches(worktrees, tt.toComplete)
			if !stringSliceEqual(result, tt.expected) {
				t.Errorf("filterWorktreeBranches(%q) = %v, want %v", tt.toComplete, result, tt.expected)
			}
		})
	}
}

func TestFilterWorktreeBranchesNoDuplicates(t *testing.T) {
	t.Parallel()

	// When branch name has no slashes, flat == original.
	// Should NOT produce duplicates.
	worktrees := []git.Worktree{
		{Path: "/p/main", Branch: "main"},
		{Path: "/p/hotfix", Branch: "hotfix"},
	}

	result := filterWorktreeBranches(worktrees, "")
	seen := make(map[string]int)
	for _, b := range result {
		seen[b]++
	}
	for b, count := range seen {
		if count > 1 {
			t.Errorf("duplicate branch in completion: %q appeared %d times", b, count)
		}
	}
}

func TestFilterWorktreeBranchesSlashNoDuplicates(t *testing.T) {
	t.Parallel()

	// With slashes: flat ("feat-auth") != original ("feat/auth").
	// When toComplete is empty, both prefixes match — but should appear only once.
	worktrees := []git.Worktree{
		{Path: "/p/feat-auth", Branch: "feat/auth"},
		{Path: "/p/feat-ui", Branch: "feat/ui"},
	}

	result := filterWorktreeBranches(worktrees, "")
	seen := make(map[string]int)
	for _, b := range result {
		seen[b]++
	}
	for b, count := range seen {
		if count > 1 {
			t.Errorf("duplicate branch in completion: %q appeared %d times", b, count)
		}
	}
}

func TestFilterWorktreeBranchesDeepSlashes(t *testing.T) {
	t.Parallel()

	worktrees := []git.Worktree{
		{Path: "/p/user-feat-thing", Branch: "user/feat/thing"},
	}

	tests := []struct {
		name       string
		toComplete string
		expected   []string
	}{
		{"original prefix", "user/", []string{"user/feat/thing"}},
		{"flat prefix", "user-", []string{"user/feat/thing"}},
		{"deep original prefix", "user/feat/", []string{"user/feat/thing"}},
		{"deep flat prefix", "user-feat-", []string{"user/feat/thing"}},
		{"partial", "us", []string{"user/feat/thing"}},
		{"no match", "admin", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := filterWorktreeBranches(worktrees, tt.toComplete)
			if !stringSliceEqual(result, tt.expected) {
				t.Errorf("filterWorktreeBranches(%q) = %v, want %v", tt.toComplete, result, tt.expected)
			}
		})
	}
}

// --- completeShellNames tests ---

func TestCompleteShellNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		args       []string
		toComplete string
		expected   []string
	}{
		{
			name:       "empty prefix returns all shells",
			args:       nil,
			toComplete: "",
			expected:   []string{"bash", "zsh", "fish", "powershell"},
		},
		{
			name:       "prefix b",
			args:       nil,
			toComplete: "b",
			expected:   []string{"bash"},
		},
		{
			name:       "prefix z",
			args:       nil,
			toComplete: "z",
			expected:   []string{"zsh"},
		},
		{
			name:       "prefix f",
			args:       nil,
			toComplete: "f",
			expected:   []string{"fish"},
		},
		{
			name:       "prefix p",
			args:       nil,
			toComplete: "p",
			expected:   []string{"powershell"},
		},
		{
			name:       "prefix power",
			args:       nil,
			toComplete: "power",
			expected:   []string{"powershell"},
		},
		{
			name:       "no match",
			args:       nil,
			toComplete: "xyz",
			expected:   nil,
		},
		{
			name:       "arg already provided returns nothing",
			args:       []string{"bash"},
			toComplete: "",
			expected:   nil,
		},
	}

	cmd := &cobra.Command{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, directive := completeShellNames(cmd, tt.args, tt.toComplete)
			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("expected ShellCompDirectiveNoFileComp, got %v", directive)
			}
			if !stringSliceEqual(result, tt.expected) {
				t.Errorf("completeShellNames(%v, %q) = %v, want %v", tt.args, tt.toComplete, result, tt.expected)
			}
		})
	}
}

// --- completeHookNames tests ---

func TestCompleteHookNames(t *testing.T) {
	t.Parallel()

	// Create temp dirs with hook scripts
	tmpDir := t.TempDir()
	communityDir := filepath.Join(tmpDir, "community")
	customDir := filepath.Join(tmpDir, "custom")
	if err := os.MkdirAll(communityDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create community hooks
	for _, name := range []string{"zoxide.sh", "direnv.sh", "gh-default.sh"} {
		if err := os.WriteFile(filepath.Join(communityDir, name), []byte("#!/bin/sh"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Create a custom hook
	if err := os.WriteFile(filepath.Join(customDir, "my-hook.sh"), []byte("#!/bin/sh"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a directory (should be excluded)
	if err := os.MkdirAll(filepath.Join(communityDir, "not-a-hook.sh"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a non-.sh file (should be excluded)
	if err := os.WriteFile(filepath.Join(communityDir, "README.md"), []byte("# hooks"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test the core logic directly (completeHookNames reads from config dirs)
	result := collectHookNames(customDir, communityDir, "")
	sort.Strings(result)
	expected := []string{"direnv", "gh-default", "my-hook", "zoxide"}
	sort.Strings(expected)
	if !stringSliceEqual(result, expected) {
		t.Errorf("collectHookNames('') = %v, want %v", result, expected)
	}

	// Test with prefix
	result = collectHookNames(customDir, communityDir, "z")
	if !stringSliceEqual(result, []string{"zoxide"}) {
		t.Errorf("collectHookNames('z') = %v, want [zoxide]", result)
	}

	// Test with prefix that matches multiple
	result = collectHookNames(customDir, communityDir, "d")
	if !stringSliceEqual(result, []string{"direnv"}) {
		t.Errorf("collectHookNames('d') = %v, want [direnv]", result)
	}

	// Test with no match
	result = collectHookNames(customDir, communityDir, "xyz")
	if len(result) != 0 {
		t.Errorf("collectHookNames('xyz') = %v, want empty", result)
	}
}

func TestCompleteHookNamesCustomOverridesCommunity(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	communityDir := filepath.Join(tmpDir, "community")
	customDir := filepath.Join(tmpDir, "custom")
	if err := os.MkdirAll(communityDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Same hook name in both dirs — should appear only once
	if err := os.WriteFile(filepath.Join(communityDir, "zoxide.sh"), []byte("#!/bin/sh\n# community"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(customDir, "zoxide.sh"), []byte("#!/bin/sh\n# custom"), 0644); err != nil {
		t.Fatal(err)
	}

	result := collectHookNames(customDir, communityDir, "")
	count := 0
	for _, name := range result {
		if name == "zoxide" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected zoxide to appear once, got %d times in %v", count, result)
	}
}

func TestCompleteHookNamesEmptyDirs(t *testing.T) {
	t.Parallel()

	// Non-existent dirs — should not panic, return empty
	result := collectHookNames("/nonexistent/custom", "/nonexistent/community", "")
	if len(result) != 0 {
		t.Errorf("expected empty result for nonexistent dirs, got %v", result)
	}
}

// collectHookNames extracts the hook name listing logic for testability
func collectHookNames(customDir, communityDir, toComplete string) []string {
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

	return names
}

// --- completeWorkflowNames tests ---

func TestCompleteWorkflowNamesLogic(t *testing.T) {
	t.Parallel()

	workflows := map[string]bool{
		"feature":   true,
		"bugfix":    true,
		"pr-review": true,
		"custom-wf": true,
	}

	tests := []struct {
		name       string
		toComplete string
		expected   []string
	}{
		{
			name:       "empty prefix returns all",
			toComplete: "",
			expected:   []string{"bugfix", "custom-wf", "feature", "pr-review"},
		},
		{
			name:       "prefix f",
			toComplete: "f",
			expected:   []string{"feature"},
		},
		{
			name:       "prefix pr",
			toComplete: "pr",
			expected:   []string{"pr-review"},
		},
		{
			name:       "prefix bug",
			toComplete: "bug",
			expected:   []string{"bugfix"},
		},
		{
			name:       "prefix custom",
			toComplete: "custom",
			expected:   []string{"custom-wf"},
		},
		{
			name:       "no match",
			toComplete: "xyz",
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var result []string
			for name := range workflows {
				if strings.HasPrefix(name, tt.toComplete) {
					result = append(result, name)
				}
			}
			sort.Strings(result)
			sort.Strings(tt.expected)
			if !stringSliceEqual(result, tt.expected) {
				t.Errorf("workflow completion(%q) = %v, want %v", tt.toComplete, result, tt.expected)
			}
		})
	}
}

// --- completeGitBranches parsing tests ---

func TestParseGitBranchOutput(t *testing.T) {
	t.Parallel()

	// Simulate `git branch --format=%(refname:short)` output
	output := "main\nfeat/auth\nfeat/ui\nfix/login\nhotfix\n"

	tests := []struct {
		name       string
		toComplete string
		expected   []string
	}{
		{
			name:       "empty prefix returns all",
			toComplete: "",
			expected:   []string{"main", "feat/auth", "feat/ui", "fix/login", "hotfix"},
		},
		{
			name:       "prefix feat",
			toComplete: "feat",
			expected:   []string{"feat/auth", "feat/ui"},
		},
		{
			name:       "prefix fix",
			toComplete: "fix",
			expected:   []string{"fix/login"},
		},
		{
			name:       "exact match",
			toComplete: "main",
			expected:   []string{"main"},
		},
		{
			name:       "no match",
			toComplete: "xyz",
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var branches []string
			for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
				line = strings.TrimSpace(line)
				if line != "" && strings.HasPrefix(line, tt.toComplete) {
					branches = append(branches, line)
				}
			}
			if !stringSliceEqual(branches, tt.expected) {
				t.Errorf("branch parsing(%q) = %v, want %v", tt.toComplete, branches, tt.expected)
			}
		})
	}
}

func TestParseGitBranchOutputEmpty(t *testing.T) {
	t.Parallel()

	// Empty repo has no branches
	output := ""
	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}

	if len(branches) != 0 {
		t.Errorf("expected empty branches for empty output, got %v", branches)
	}
}

func TestParseGitBranchOutputWhitespace(t *testing.T) {
	t.Parallel()

	// Handle trailing newlines and whitespace
	output := "  main  \n  feat/auth  \n\n"
	var branches []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}

	expected := []string{"main", "feat/auth"}
	if !stringSliceEqual(branches, expected) {
		t.Errorf("expected %v, got %v", expected, branches)
	}
}

// --- Directive tests: all completions return NoFileComp ---

func TestAllCompletionsReturnNoFileComp(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	// Every completion function should return ShellCompDirectiveNoFileComp
	// regardless of whether it succeeds or fails to find context.
	// We call each from the current directory — some may find a project,
	// some may not — but the directive should always be NoFileComp.
	completionFuncs := map[string]func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective){
		"completeShellNames":       completeShellNames,
		"completeWorktreeBranches": completeWorktreeBranches,
		"completeGitBranches":      completeGitBranches,
		"completeWorkflowNames":    completeWorkflowNames,
		"completeHookNames":        completeHookNames,
		"completeEnabledHookNames": completeEnabledHookNames,
	}

	for name, fn := range completionFuncs {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, directive := fn(cmd, nil, "")
			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("%s: expected ShellCompDirectiveNoFileComp, got %v", name, directive)
			}
		})
	}
}

// --- ValidArgsFunction registration tests ---

func TestValidArgsFunctionRegistered(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cmdPath  []string
		wantFunc bool
	}{
		{"switch", []string{"switch"}, true},
		{"delete", []string{"delete"}, true},
		{"completion", []string{"completion"}, true},
		{"list", []string{"list"}, true},
		{"prune", []string{"prune"}, true},
		{"repair", []string{"repair"}, true},
		{"hooks enable", []string{"hooks", "enable"}, true},
		{"hooks disable", []string{"hooks", "disable"}, true},
		{"hooks show", []string{"hooks", "show"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cmd, _, err := rootCmd.Find(tt.cmdPath)
			if err != nil {
				t.Fatalf("command %v not found: %v", tt.cmdPath, err)
			}
			hasFunc := cmd.ValidArgsFunction != nil
			if hasFunc != tt.wantFunc {
				t.Errorf("command %v ValidArgsFunction: got registered=%v, want %v", tt.cmdPath, hasFunc, tt.wantFunc)
			}
		})
	}
}

func TestNoArgsCommandsReturnNil(t *testing.T) {
	t.Parallel()

	// Commands that accept no positional args should return nil completions
	noArgsCmds := [][]string{
		{"list"},
		{"prune"},
		{"repair"},
	}

	cmd := &cobra.Command{}
	for _, cmdPath := range noArgsCmds {
		t.Run(strings.Join(cmdPath, " "), func(t *testing.T) {
			t.Parallel()
			found, _, err := rootCmd.Find(cmdPath)
			if err != nil {
				t.Fatalf("command %v not found: %v", cmdPath, err)
			}
			if found.ValidArgsFunction == nil {
				t.Fatalf("command %v has no ValidArgsFunction", cmdPath)
			}
			result, directive := found.ValidArgsFunction(cmd, nil, "")
			if result != nil {
				t.Errorf("command %v should return nil completions, got %v", cmdPath, result)
			}
			if directive != cobra.ShellCompDirectiveNoFileComp {
				t.Errorf("command %v should return NoFileComp directive", cmdPath)
			}
		})
	}
}

// --- Flag completion registration tests ---

func TestFlagCompletionRegistered(t *testing.T) {
	t.Parallel()

	// add/new command should have flag completions for --base and --workflow
	cmd, _, err := rootCmd.Find([]string{"add"})
	if err != nil {
		t.Fatalf("add command not found: %v", err)
	}

	// Check that the flags exist
	baseFlag := cmd.Flags().Lookup("base")
	if baseFlag == nil {
		t.Fatal("expected --base flag on add command")
	}

	workflowFlag := cmd.Flags().Lookup("workflow")
	if workflowFlag == nil {
		t.Fatal("expected --workflow flag on add command")
	}

	// Check hooks enable --event flag completion
	hooksEnableCmd, _, err := rootCmd.Find([]string{"hooks", "enable"})
	if err != nil {
		t.Fatalf("hooks enable command not found: %v", err)
	}

	eventFlag := hooksEnableCmd.Flags().Lookup("event")
	if eventFlag == nil {
		t.Fatal("expected --event flag on hooks enable command")
	}
}

// --- Delete alias test ---

func TestDeleteAliasHasCompletion(t *testing.T) {
	t.Parallel()

	// "delete", "rm", and "remove" should all resolve to the same command with completion
	deleteCmd, _, err := rootCmd.Find([]string{"delete"})
	if err != nil {
		t.Fatalf("delete command not found: %v", err)
	}
	if deleteCmd.ValidArgsFunction == nil {
		t.Error("delete command should have ValidArgsFunction")
	}

	for _, alias := range []string{"rm", "remove"} {
		cmdAlias, _, err := rootCmd.Find([]string{alias})
		if err != nil {
			t.Fatalf("%s alias not found: %v", alias, err)
		}
		if cmdAlias != deleteCmd {
			t.Errorf("%s should resolve to the same command as delete", alias)
		}
		if cmdAlias.ValidArgsFunction == nil {
			t.Errorf("%s alias should have ValidArgsFunction", alias)
		}
	}
}

// --- List alias test ---

func TestListAliasHasCompletion(t *testing.T) {
	t.Parallel()

	cmd, _, err := rootCmd.Find([]string{"ls"})
	if err != nil {
		t.Fatalf("ls alias not found: %v", err)
	}
	if cmd.ValidArgsFunction == nil {
		t.Error("ls alias should have ValidArgsFunction")
	}
}

// --- Add/new/create alias test ---

func TestAddNewCreateAliasesResolveSame(t *testing.T) {
	t.Parallel()

	addCmd, _, err := rootCmd.Find([]string{"add"})
	if err != nil {
		t.Fatalf("add command not found: %v", err)
	}

	for _, alias := range []string{"new", "create"} {
		aliasCmd, _, err := rootCmd.Find([]string{alias})
		if err != nil {
			t.Fatalf("%s alias not found: %v", alias, err)
		}
		if aliasCmd != addCmd {
			t.Errorf("%s should resolve to the same command as add", alias)
		}
	}
}

// --- Switch/cd alias test ---

func TestSwitchCdAliasResolveSame(t *testing.T) {
	t.Parallel()

	switchCmd, _, err := rootCmd.Find([]string{"switch"})
	if err != nil {
		t.Fatalf("switch command not found: %v", err)
	}

	cdCmd, _, err := rootCmd.Find([]string{"cd"})
	if err != nil {
		t.Fatalf("cd alias not found: %v", err)
	}

	if cdCmd != switchCmd {
		t.Error("cd should resolve to the same command as switch")
	}
	if cdCmd.ValidArgsFunction == nil {
		t.Error("cd alias should have ValidArgsFunction")
	}
}

// --- Prune/clean alias test ---

func TestPruneCleanAliasResolveSame(t *testing.T) {
	t.Parallel()

	pruneCmd, _, err := rootCmd.Find([]string{"prune"})
	if err != nil {
		t.Fatalf("prune command not found: %v", err)
	}

	cleanCmd, _, err := rootCmd.Find([]string{"clean"})
	if err != nil {
		t.Fatalf("clean alias not found: %v", err)
	}

	if cleanCmd != pruneCmd {
		t.Error("clean should resolve to the same command as prune")
	}
}

// --- Repair/fix alias test ---

func TestRepairFixAliasResolveSame(t *testing.T) {
	t.Parallel()

	repairCmd, _, err := rootCmd.Find([]string{"repair"})
	if err != nil {
		t.Fatalf("repair command not found: %v", err)
	}

	fixCmd, _, err := rootCmd.Find([]string{"fix"})
	if err != nil {
		t.Fatalf("fix alias not found: %v", err)
	}

	if fixCmd != repairCmd {
		t.Error("fix should resolve to the same command as repair")
	}
}

// --- Edge case: completeShellNames with already-selected arg ---

func TestCompleteShellNamesStopsAfterFirst(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	// After selecting one shell, no more completions
	result, directive := completeShellNames(cmd, []string{"bash"}, "")
	if result != nil {
		t.Errorf("should return nil after first arg, got %v", result)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected NoFileComp directive")
	}

	// Even with prefix after first arg
	result, directive = completeShellNames(cmd, []string{"zsh"}, "f")
	if result != nil {
		t.Errorf("should return nil after first arg, got %v", result)
	}
	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected NoFileComp directive")
	}
}

// --- Helper ---

func stringSliceEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
