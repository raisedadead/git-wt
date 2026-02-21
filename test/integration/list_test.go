package integration

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestList(t *testing.T) {
	t.Parallel()

	t.Run("lists main worktree", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "list-test", "--timeout", "300")
		projectDir := filepath.Join(workspace, "list-test")
		mainDir := filepath.Join(projectDir, "main")

		stdout := runGitWTSuccess(t, mainDir, "list")
		assertContains(t, stdout, "main")
		// Table should have rounded borders
		assertContains(t, stdout, "╭")
		assertContains(t, stdout, "╰")
		// Table should have column headers
		assertContains(t, stdout, "BRANCH")
		assertContains(t, stdout, "STATUS")
		assertContains(t, stdout, "PATH")
	})

	t.Run("lists multiple worktrees", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "list-multi", "--timeout", "300")
		projectDir := filepath.Join(workspace, "list-multi")
		mainDir := filepath.Join(projectDir, "main")

		// Add another worktree
		runGitWTSuccess(t, mainDir, "add", "feature-test", "--new")

		stdout := runGitWTSuccess(t, mainDir, "list")
		assertContains(t, stdout, "main")
		assertContains(t, stdout, "feature-test")
		// Both should be in the same table
		assertContains(t, stdout, "│")
	})
}

func TestListJSON(t *testing.T) {
	t.Parallel()

	workspace := setupTestWorkspace(t)

	runGitWTSuccess(t, workspace, "clone", localRemote, "list-json", "--timeout", "300")
	projectDir := filepath.Join(workspace, "list-json")
	mainDir := filepath.Join(projectDir, "main")

	result := runGitWTJSON(t, mainDir, "list")

	// Check envelope
	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success: true")
	}

	data, ok := result["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data object")
	}

	worktrees, ok := data["worktrees"].([]any)
	if !ok {
		t.Fatalf("Expected worktrees array")
	}

	if len(worktrees) < 1 {
		t.Errorf("Expected at least 1 worktree")
	}
}

func TestListStatus(t *testing.T) {
	t.Parallel()

	t.Run("freshly created branch shows no merged or gone", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "list-fresh", "--timeout", "300")
		projectDir := filepath.Join(workspace, "list-fresh")
		mainDir := filepath.Join(projectDir, "main")

		// Create a new worktree with a fresh branch (never pushed)
		runGitWTSuccess(t, mainDir, "add", "fresh-branch", "--new")

		// Get JSON output to check status
		result := runGitWTJSON(t, mainDir, "list")
		worktrees := getWorktreesFromResult(t, result)

		// Find the fresh-branch worktree
		var freshWT map[string]any
		for _, wt := range worktrees {
			wtMap := wt.(map[string]any)
			if wtMap["branch"] == "fresh-branch" {
				freshWT = wtMap
				break
			}
		}

		if freshWT == nil {
			t.Fatal("Expected to find fresh-branch worktree")
		}

		// Fresh branch should NOT be marked as merged
		if merged, ok := freshWT["merged"].(bool); ok && merged {
			t.Error("Fresh branch should NOT be marked as merged")
		}

		// Fresh branch should NOT be marked as remote_gone (never had remote)
		if remoteGone, ok := freshWT["remote_gone"].(bool); ok && remoteGone {
			t.Error("Fresh branch should NOT be marked as remote_gone (never pushed)")
		}
	})

	t.Run("branch with commits ahead shows not merged", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "list-ahead", "--timeout", "300")
		projectDir := filepath.Join(workspace, "list-ahead")
		mainDir := filepath.Join(projectDir, "main")

		// Create a new worktree
		runGitWTSuccess(t, mainDir, "add", "ahead-branch", "--new")
		aheadDir := filepath.Join(projectDir, "ahead-branch")

		// Make a commit in the new branch
		runGit(t, aheadDir, "commit", "--allow-empty", "-m", "test commit")

		// Get JSON output
		result := runGitWTJSON(t, mainDir, "list")
		worktrees := getWorktreesFromResult(t, result)

		// Find the ahead-branch worktree
		var aheadWT map[string]any
		for _, wt := range worktrees {
			wtMap := wt.(map[string]any)
			if wtMap["branch"] == "ahead-branch" {
				aheadWT = wtMap
				break
			}
		}

		if aheadWT == nil {
			t.Fatal("Expected to find ahead-branch worktree")
		}

		// Branch with commits ahead should NOT be marked as merged
		if merged, ok := aheadWT["merged"].(bool); ok && merged {
			t.Error("Branch with commits ahead should NOT be marked as merged")
		}
	})

	t.Run("truly merged branch shows merged", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "list-merged", "--timeout", "300")
		projectDir := filepath.Join(workspace, "list-merged")
		mainDir := filepath.Join(projectDir, "main")

		// Create a new worktree
		runGitWTSuccess(t, mainDir, "add", "merged-branch", "--new")
		mergedDir := filepath.Join(projectDir, "merged-branch")

		// Make a commit in the new branch
		runGit(t, mergedDir, "commit", "--allow-empty", "-m", "feature commit")

		// Make a commit in main first so we can do a real merge (not fast-forward)
		runGit(t, mainDir, "commit", "--allow-empty", "-m", "main commit")

		// Merge the branch into main with --no-ff to ensure a merge commit
		runGit(t, mainDir, "merge", "merged-branch", "--no-ff", "--no-edit")

		// Get JSON output
		result := runGitWTJSON(t, mainDir, "list")
		worktrees := getWorktreesFromResult(t, result)

		// Find the merged-branch worktree
		var mergedWT map[string]any
		for _, wt := range worktrees {
			wtMap := wt.(map[string]any)
			if wtMap["branch"] == "merged-branch" {
				mergedWT = wtMap
				break
			}
		}

		if mergedWT == nil {
			t.Fatal("Expected to find merged-branch worktree")
		}

		// Branch that was merged SHOULD be marked as merged
		// (main has the merge commit + main's commit, branch only has feature commit)
		if merged, ok := mergedWT["merged"].(bool); !ok || !merged {
			t.Errorf("Actually merged branch should be marked as merged (got merged=%v)", mergedWT["merged"])
		}
	})

	t.Run("branch pushed then remote deleted shows gone", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "list-gone", "--timeout", "300")
		projectDir := filepath.Join(workspace, "list-gone")
		mainDir := filepath.Join(projectDir, "main")

		// Create a new worktree
		runGitWTSuccess(t, mainDir, "add", "gone-branch", "--new")
		goneDir := filepath.Join(projectDir, "gone-branch")

		// Push the branch to set up tracking
		runGit(t, goneDir, "push", "-u", "origin", "gone-branch")

		// Delete the remote branch
		runGit(t, goneDir, "push", "origin", "--delete", "gone-branch")

		// Fetch to update remote refs
		runGit(t, mainDir, "fetch", "--prune")

		// Get JSON output
		result := runGitWTJSON(t, mainDir, "list")
		worktrees := getWorktreesFromResult(t, result)

		// Find the gone-branch worktree
		var goneWT map[string]any
		for _, wt := range worktrees {
			wtMap := wt.(map[string]any)
			if wtMap["branch"] == "gone-branch" {
				goneWT = wtMap
				break
			}
		}

		if goneWT == nil {
			t.Fatal("Expected to find gone-branch worktree")
		}

		// Branch that was pushed then deleted SHOULD be marked as remote_gone
		if remoteGone, ok := goneWT["remote_gone"].(bool); !ok || !remoteGone {
			t.Error("Branch with deleted remote should be marked as remote_gone")
		}
	})
}

// getWorktreesFromResult extracts the worktrees array from a JSON result
func getWorktreesFromResult(t *testing.T, result map[string]any) []any {
	t.Helper()

	data, ok := result["data"].(map[string]any)
	if !ok {
		// Try to get raw JSON for debugging
		jsonBytes, _ := json.MarshalIndent(result, "", "  ")
		t.Fatalf("Expected data object, got: %s", string(jsonBytes))
	}

	worktrees, ok := data["worktrees"].([]any)
	if !ok {
		t.Fatalf("Expected worktrees array")
	}

	return worktrees
}
