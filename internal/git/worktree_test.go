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

func TestCreateWorktreeFromRemote_InvalidRepo(t *testing.T) {
	_, err := CreateWorktreeFromRemote("/nonexistent", "feature/auth", "origin")
	if err == nil {
		t.Error("expected error for non-repo directory")
	}
}

func TestHasBranchUpstream_InvalidRepo(t *testing.T) {
	// Should return false for non-existent repo
	result := HasBranchUpstream("/nonexistent", "main")
	if result {
		t.Error("expected false for non-repo directory")
	}
}

func TestGetCommitsAhead_InvalidRepo(t *testing.T) {
	// Should return 0 for non-existent repo
	count, err := GetCommitsAhead("/nonexistent", "feature", "main")
	if err == nil {
		t.Error("expected error for non-repo directory")
	}
	if count != 0 {
		t.Errorf("expected 0 commits ahead for non-repo, got %d", count)
	}
}

func TestGetCommitsBehind_InvalidRepo(t *testing.T) {
	// Should return 0 for non-existent repo
	count, err := GetCommitsBehind("/nonexistent", "feature", "main")
	if err == nil {
		t.Error("expected error for non-repo directory")
	}
	if count != 0 {
		t.Errorf("expected 0 commits behind for non-repo, got %d", count)
	}
}

func TestIsTrulyMerged_InvalidRepo(t *testing.T) {
	// Should return false for non-existent repo
	result := IsTrulyMerged("/nonexistent", "feature", "main")
	if result {
		t.Error("expected false for non-repo directory")
	}
}

func TestParseWorktreeList_DetachedHead(t *testing.T) {
	output := `worktree /path/to/main
HEAD abc123def456789
branch refs/heads/main

worktree /path/to/detached
HEAD 1234567890abcdef
detached

worktree /path/to/feature
HEAD def456
branch refs/heads/feature
`

	worktrees := parseWorktreeList(output)

	if len(worktrees) != 3 {
		t.Fatalf("expected 3 worktrees, got %d", len(worktrees))
	}

	// First worktree should have normal branch
	if worktrees[0].Branch != "main" {
		t.Errorf("expected main, got %s", worktrees[0].Branch)
	}

	// Second worktree should be detached with short commit hash
	expectedDetached := "HEAD detached at 1234567"
	if worktrees[1].Branch != expectedDetached {
		t.Errorf("expected %q, got %q", expectedDetached, worktrees[1].Branch)
	}

	// Third worktree should have normal branch
	if worktrees[2].Branch != "feature" {
		t.Errorf("expected feature, got %s", worktrees[2].Branch)
	}
}

func TestParseWorktreeList_BareRepoFiltered(t *testing.T) {
	output := `worktree /path/to/project/.bare
HEAD abc123def456789
bare

worktree /path/to/project/main
HEAD def456ghi789
branch refs/heads/main

worktree /path/to/project/feature-auth
HEAD 111222333444
branch refs/heads/feature/auth
`

	worktrees := parseWorktreeList(output)

	// Bare entry should be filtered out, only 2 real worktrees remain
	if len(worktrees) != 2 {
		t.Fatalf("expected 2 worktrees (bare filtered out), got %d", len(worktrees))
	}

	if worktrees[0].Path != "/path/to/project/main" {
		t.Errorf("expected /path/to/project/main, got %s", worktrees[0].Path)
	}
	if worktrees[0].Branch != "main" {
		t.Errorf("expected main, got %s", worktrees[0].Branch)
	}

	if worktrees[1].Path != "/path/to/project/feature-auth" {
		t.Errorf("expected /path/to/project/feature-auth, got %s", worktrees[1].Path)
	}
	if worktrees[1].Branch != "feature/auth" {
		t.Errorf("expected feature/auth, got %s", worktrees[1].Branch)
	}
}

func TestParseWorktreeList_BareRepoIsBareField(t *testing.T) {
	output := `worktree /path/to/project/.bare
HEAD abc123def456789
bare

worktree /path/to/project/main
HEAD def456ghi789
branch refs/heads/main
`

	// Use parseWorktreeListAll to get unfiltered results including bare entries
	worktrees := parseWorktreeListAll(output)

	if len(worktrees) != 2 {
		t.Fatalf("expected 2 worktrees from parseWorktreeListAll, got %d", len(worktrees))
	}

	if !worktrees[0].IsBare {
		t.Errorf("expected first worktree to have IsBare=true")
	}
	if worktrees[0].Branch != "" {
		t.Errorf("expected bare worktree to have empty branch, got %s", worktrees[0].Branch)
	}

	if worktrees[1].IsBare {
		t.Errorf("expected second worktree to have IsBare=false")
	}
	if worktrees[1].Branch != "main" {
		t.Errorf("expected main, got %s", worktrees[1].Branch)
	}
}

func TestParseWorktreeList_DetachedHeadShortCommit(t *testing.T) {
	// Test when commit is already short (less than 7 chars)
	output := `worktree /path/to/detached
HEAD abc
detached
`

	worktrees := parseWorktreeList(output)

	if len(worktrees) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(worktrees))
	}

	// Should use the full short commit since it's less than 7 chars
	expectedDetached := "HEAD detached at abc"
	if worktrees[0].Branch != expectedDetached {
		t.Errorf("expected %q, got %q", expectedDetached, worktrees[0].Branch)
	}
}
