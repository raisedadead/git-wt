package github

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Fix login redirect after OAuth", "fix-login-redirect-after-oauth"},
		{"Add new feature!", "add-new-feature"},
		{"  Multiple   Spaces  ", "multiple-spaces"},
		{"UPPERCASE", "uppercase"},
	}

	for _, tt := range tests {
		result := Slugify(tt.input)
		if result != tt.expected {
			t.Errorf("Slugify(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestGenerateBranchName(t *testing.T) {
	tests := []struct {
		prefix   string
		number   int
		title    string
		expected string
	}{
		{"issue", 42, "Fix login redirect", "issue-42-fix-login-redirect"},
		{"pr", 123, "Add new feature!", "pr-123-add-new-feature"},
		{"issue", 1, "UPPERCASE Title", "issue-1-uppercase-title"},
	}

	for _, tt := range tests {
		branch := GenerateBranchName(tt.prefix, tt.number, tt.title)
		if branch != tt.expected {
			t.Errorf("GenerateBranchName(%q, %d, %q) = %q, want %q",
				tt.prefix, tt.number, tt.title, branch, tt.expected)
		}
	}
}
