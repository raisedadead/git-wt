package commands

import (
	"os"
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "Hello World", "hello-world"},
		{"special chars", "Fix: bug #42", "fix-bug-42"},
		{"multiple spaces", "Fix   multiple   spaces", "fix-multiple-spaces"},
		{"leading trailing spaces", "  trim me  ", "trim-me"},
		{"empty", "", ""},
		{"numbers", "Version 123", "version-123"},
		{"slashes become nothing", "feature/auth", "featureauth"},
		{"consecutive dashes", "foo---bar", "foo-bar"},
		{"leading dash", "-foo", "foo"},
		{"trailing dash", "foo-", "foo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slugify(tt.input)
			if result != tt.expected {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExpandRepoShorthand(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		defaultOwner string
		expected     string
	}{
		{"owner/repo", "owner/repo", "", "git@github.com:owner/repo.git"},
		{"with .git suffix", "owner/repo.git", "", "git@github.com:owner/repo.git"},
		{"https url", "https://github.com/owner/repo.git", "", "https://github.com/owner/repo.git"},
		{"ssh url", "git@github.com:owner/repo.git", "", "git@github.com:owner/repo.git"},
		{"local path absolute", "/path/to/repo", "", "/path/to/repo"},
		{"relative path dot", "./local/repo", "", "./local/repo"},
		{"relative path parent", "../sibling/repo", "", "../sibling/repo"},
		{"just name no default", "repo", "", "repo"},
		{"empty", "", "", ""},
		// Tests with default_owner
		{"just name with default", "repo", "myorg", "git@github.com:myorg/repo.git"},
		{"just name with .git suffix", "repo.git", "myorg", "git@github.com:myorg/repo.git"},
		{"owner/repo ignores default", "owner/repo", "myorg", "git@github.com:owner/repo.git"},
		{"https url ignores default", "https://github.com/owner/repo.git", "myorg", "https://github.com/owner/repo.git"},
		{"ssh url ignores default", "git@github.com:owner/repo.git", "myorg", "git@github.com:owner/repo.git"},
		{"local path ignores default", "/path/to/repo", "myorg", "/path/to/repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandRepoShorthand(tt.input, tt.defaultOwner)
			if result != tt.expected {
				t.Errorf("expandRepoShorthand(%q, %q) = %q, want %q", tt.input, tt.defaultOwner, result, tt.expected)
			}
		})
	}
}

func TestExtractRepoName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"ssh url", "git@github.com:owner/repo.git", "repo"},
		{"https url", "https://github.com/owner/repo.git", "repo"},
		{"no .git suffix", "https://github.com/owner/repo", "repo"},
		{"local path", "/path/to/myrepo", "myrepo"},
		{"just name", "repo", "repo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRepoName(tt.input)
			if result != tt.expected {
				t.Errorf("extractRepoName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetWorkflowPrefix(t *testing.T) {
	tests := []struct {
		name     string
		template string
		expected string
	}{
		{"slash separator", "feat/{slug}", "feat"},
		{"dash separator", "fix-{slug}", "fix"},
		{"no separator", "{slug}", ""},
		{"empty", "", ""},
		{"complex", "feature/issue-{number}/{slug}", "feature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getWorkflowPrefix(tt.template)
			if result != tt.expected {
				t.Errorf("getWorkflowPrefix(%q) = %q, want %q", tt.template, result, tt.expected)
			}
		})
	}
}

func TestApplyBranchTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		brName   string
		metadata map[string]string
		expected string
	}{
		{"simple name", "{name}", "my-branch", nil, "my-branch"},
		{"simple slug", "{slug}", "My Feature", nil, "my-feature"},
		{"with prefix", "feat/{slug}", "Add Login", nil, "feat/add-login"},
		{"with metadata", "fix/{number}/{slug}", "Bug Fix", map[string]string{"number": "123"}, "fix/123/bug-fix"},
		{"empty template", "", "test", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyBranchTemplate(tt.template, tt.brName, tt.metadata)
			if result != tt.expected {
				t.Errorf("applyBranchTemplate(%q, %q, %v) = %q, want %q",
					tt.template, tt.brName, tt.metadata, result, tt.expected)
			}
		})
	}
}

func TestSplitByNewline(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"simple", "line1\nline2\nline3", []string{"line1", "line2", "line3"}},
		{"empty lines preserved", "line1\n\nline2", []string{"line1", "", "line2"}},
		{"trailing newline trimmed", "line1\nline2\n", []string{"line1", "line2"}},
		{"single line", "only one", []string{"only one"}},
		{"empty string returns nil", "", nil},
		{"whitespace lines preserved", "line1\n   \nline2", []string{"line1", "   ", "line2"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitByNewline(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitByNewline(%q) len = %d, want %d\nGot: %v", tt.input, len(result), len(tt.expected), result)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("splitByNewline(%q)[%d] = %q, want %q", tt.input, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestShortenPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"home path", home + "/projects/foo", "~/projects/foo"},
		{"non-home path", "/usr/local/bin", "/usr/local/bin"},
		{"exact home", home, "~"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shortenPath(tt.input)
			if result != tt.expected {
				t.Errorf("shortenPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsJSONOutput(t *testing.T) {
	// Test that IsJSONOutput returns a boolean without panic
	// The default state should be false (no --json flag set)
	result := IsJSONOutput()
	if result {
		t.Log("IsJSONOutput() returned true (--json flag was set)")
	}
	// We just verify it returns without panic - actual value depends on test execution context
}
