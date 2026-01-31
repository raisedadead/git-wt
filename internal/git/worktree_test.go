package git

import (
	"testing"
)

func TestListWorktrees_NotInRepo(t *testing.T) {
	_, err := ListWorktrees("/nonexistent")
	if err == nil {
		t.Error("expected error for non-repo directory")
	}
}

func TestRepairWorktrees_NotInRepo(t *testing.T) {
	_, err := RepairWorktrees("/nonexistent")
	if err == nil {
		t.Error("expected error for non-repo directory")
	}
}

func TestParseWorktreeList(t *testing.T) {
	output := `worktree /path/to/main
HEAD abc123
branch refs/heads/main

worktree /path/to/feature
HEAD def456
branch refs/heads/feature
`

	worktrees := parseWorktreeList(output)

	if len(worktrees) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(worktrees))
	}

	if worktrees[0].Path != "/path/to/main" {
		t.Errorf("expected /path/to/main, got %s", worktrees[0].Path)
	}

	if worktrees[0].Branch != "main" {
		t.Errorf("expected main, got %s", worktrees[0].Branch)
	}

	if worktrees[1].Branch != "feature" {
		t.Errorf("expected feature, got %s", worktrees[1].Branch)
	}
}

func TestParseWorktreeList_BranchWithSlashes(t *testing.T) {
	output := `worktree /path/to/feature-auth
HEAD abc123
branch refs/heads/feature/auth

worktree /path/to/fix-bug
HEAD def456
branch refs/heads/fix/security/issue-42
`

	worktrees := parseWorktreeList(output)

	if len(worktrees) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(worktrees))
	}

	// Test that branches with slashes are parsed correctly
	if worktrees[0].Branch != "feature/auth" {
		t.Errorf("expected feature/auth, got %s", worktrees[0].Branch)
	}

	// Test deeply nested branch names
	if worktrees[1].Branch != "fix/security/issue-42" {
		t.Errorf("expected fix/security/issue-42, got %s", worktrees[1].Branch)
	}
}
