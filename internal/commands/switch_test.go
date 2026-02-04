package commands

import (
	"testing"

	"github.com/raisedadead/wt/internal/git"
)

func TestSwitchBranchMatching(t *testing.T) {
	// Test that flattened branch names match correctly
	tests := []struct {
		name       string
		branchArg  string
		wtBranch   string
		shouldFind bool
	}{
		{"exact match", "main", "main", true},
		{"exact match with slash", "feature/auth", "feature/auth", true},
		{"flattened match", "feature-auth", "feature/auth", true},
		{"flattened match deep", "fix-security-issue", "fix/security/issue", true},
		{"no match", "nonexistent", "main", false},
		{"partial no match", "feat", "feature/auth", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate matching logic from switch command
			found := false
			if tt.wtBranch == tt.branchArg {
				found = true
			} else if git.FlattenBranchName(tt.wtBranch) == tt.branchArg {
				found = true
			}

			if found != tt.shouldFind {
				t.Errorf("branch matching for %q against %q: got %v, want %v",
					tt.branchArg, tt.wtBranch, found, tt.shouldFind)
			}
		})
	}
}

func TestSwitchCommandExists(t *testing.T) {
	// Verify the switch command is registered
	cmd, _, err := rootCmd.Find([]string{"switch"})
	if err != nil {
		t.Fatalf("switch command not found: %v", err)
	}
	if cmd.Use != "switch [branch]" {
		t.Errorf("unexpected command use: %s", cmd.Use)
	}
}
