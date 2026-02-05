package bundled

import (
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	hooks := List()

	// Should have bundled hooks
	if len(hooks) == 0 {
		t.Error("List() returned no hooks")
	}

	// Check for known hooks
	knownHooks := []string{"github-issue", "github-pr", "direnv", "zoxide"}
	for _, name := range knownHooks {
		found := false
		for _, h := range hooks {
			if h == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("List() missing expected hook %q", name)
		}
	}
}

func TestGetHelpers(t *testing.T) {
	content, err := GetHelpers()
	if err != nil {
		t.Fatalf("GetHelpers() error: %v", err)
	}

	// Check for key helper functions
	expectedFuncs := []string{
		"wt_info",
		"wt_error",
		"wt_warn",
		"wt_set_branch",
		"wt_set_meta",
		"wt_requires",
		"wt_slugify",
	}

	for _, fn := range expectedFuncs {
		if !strings.Contains(string(content), fn) {
			t.Errorf("GetHelpers() missing function %q", fn)
		}
	}
}

func TestScriptsEmbed(t *testing.T) {
	// Verify Scripts filesystem works
	entries, err := Scripts.ReadDir(".")
	if err != nil {
		t.Fatalf("Scripts.ReadDir() error: %v", err)
	}

	// Should have .sh files
	hasShFile := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".sh") {
			hasShFile = true
			break
		}
	}
	if !hasShFile {
		t.Error("Scripts embed has no .sh files")
	}
}

func TestHelpersScript(t *testing.T) {
	// Test that helpers.sh is readable directly
	content, err := Scripts.ReadFile("helpers.sh")
	if err != nil {
		t.Fatalf("Scripts.ReadFile(helpers.sh) error: %v", err)
	}

	if len(content) == 0 {
		t.Error("helpers.sh is empty")
	}

	// Should be a bash script
	if !strings.HasPrefix(string(content), "#!/bin/bash") {
		t.Error("helpers.sh doesn't start with shebang")
	}
}

func TestGithubIssueHook(t *testing.T) {
	content, err := Scripts.ReadFile("github-issue.sh")
	if err != nil {
		t.Fatalf("Scripts.ReadFile(github-issue.sh) error: %v", err)
	}

	// Should have TTY check after our fix
	if !strings.Contains(string(content), "[ -t 0 ]") {
		t.Error("github-issue.sh missing TTY check")
	}

	// Should source helpers
	if !strings.Contains(string(content), "source \"$WT_LIB/helpers.sh\"") {
		t.Error("github-issue.sh doesn't source helpers.sh")
	}

	// Should have metadata comments
	if !strings.Contains(string(content), "@name: github-issue") {
		t.Error("github-issue.sh missing @name metadata")
	}
}

func TestGithubPRHook(t *testing.T) {
	content, err := Scripts.ReadFile("github-pr.sh")
	if err != nil {
		t.Fatalf("Scripts.ReadFile(github-pr.sh) error: %v", err)
	}

	// Should have TTY check after our fix
	if !strings.Contains(string(content), "[ -t 0 ]") {
		t.Error("github-pr.sh missing TTY check")
	}

	// Should source helpers
	if !strings.Contains(string(content), "source \"$WT_LIB/helpers.sh\"") {
		t.Error("github-pr.sh doesn't source helpers.sh")
	}
}

func TestGithubPRHookPRListFeature(t *testing.T) {
	content, err := Scripts.ReadFile("github-pr.sh")
	if err != nil {
		t.Fatalf("Scripts.ReadFile(github-pr.sh) error: %v", err)
	}

	scriptStr := string(content)

	// Should have get_pr_repo function for detecting correct repo
	if !strings.Contains(scriptStr, "get_pr_repo") {
		t.Error("github-pr.sh missing get_pr_repo function for remote detection")
	}

	// Should have gh pr list command to fetch open PRs
	if !strings.Contains(scriptStr, "gh pr list") {
		t.Error("github-pr.sh missing gh pr list command for fetching open PRs")
	}

	// Should support manual entry fallback when no PRs available
	if !strings.Contains(scriptStr, "No open PRs") || !strings.Contains(scriptStr, "Enter PR number manually") {
		t.Error("github-pr.sh missing manual entry fallback message for empty PR list")
	}

	// Should have select menu for PR selection
	if !strings.Contains(scriptStr, "select") || !strings.Contains(scriptStr, "Select a PR") {
		t.Error("github-pr.sh missing interactive PR selection menu")
	}
}

func TestAllBundledHooksExist(t *testing.T) {
	for _, hook := range List() {
		filename := hook + ".sh"
		content, err := Scripts.ReadFile(filename)
		if err != nil {
			t.Errorf("bundled hook %q not found: %v", hook, err)
			continue
		}
		if len(content) == 0 {
			t.Errorf("bundled hook %q is empty", hook)
		}
		// Should be executable scripts
		if !strings.HasPrefix(string(content), "#!/bin/bash") {
			t.Errorf("bundled hook %q doesn't start with bash shebang", hook)
		}
	}
}
