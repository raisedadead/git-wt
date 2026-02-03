package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func removeAll(path string) error {
	return os.RemoveAll(path)
}

func containsString(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestDelete(t *testing.T) {
	t.Parallel()

	t.Run("deletes worktree and branch", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "delete-test", "--timeout", "300")
		projectDir := filepath.Join(workspace, "delete-test")
		mainDir := filepath.Join(projectDir, "main")

		// Create a worktree to delete
		runGitWTSuccess(t, mainDir, "add", "to-delete", "--new")
		assertDirExists(t, filepath.Join(projectDir, "to-delete"))

		// Delete it
		runGitWTSuccess(t, mainDir, "delete", "to-delete", "-y")

		// Verify worktree directory is gone
		assertDirNotExists(t, filepath.Join(projectDir, "to-delete"))

		// Verify branch was deleted using git directly
		branches := runGit(t, mainDir, "branch", "--list", "to-delete")
		if branches != "" {
			t.Errorf("Expected branch to-delete to be deleted, but found: %s", branches)
		}
	})

	t.Run("can delete current worktree", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "delete-current", "--timeout", "300")
		projectDir := filepath.Join(workspace, "delete-current")
		mainDir := filepath.Join(projectDir, "main")

		// Delete main while in main - git-wt allows this
		runGitWTSuccess(t, mainDir, "delete", "main", "-y")

		// Verify worktree is gone
		assertDirNotExists(t, mainDir)
	})

	t.Run("can delete worktree with missing directory", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "delete-missing", "--timeout", "300")
		projectDir := filepath.Join(workspace, "delete-missing")
		mainDir := filepath.Join(projectDir, "main")

		// Create a worktree
		runGitWTSuccess(t, mainDir, "add", "to-remove", "--new")
		worktreePath := filepath.Join(projectDir, "to-remove")
		assertDirExists(t, worktreePath)

		// Manually delete the directory (simulating accidental deletion)
		if err := removeAll(worktreePath); err != nil {
			t.Fatalf("failed to manually remove directory: %v", err)
		}
		assertDirNotExists(t, worktreePath)

		// List should still show it with "unknown" status
		output := runGitWTSuccess(t, mainDir, "list")
		if !containsString(output, "to-remove") {
			t.Errorf("Expected worktree to still appear in list, got: %s", output)
		}

		// Delete should succeed and clean up the git reference
		runGitWTSuccess(t, mainDir, "delete", "to-remove", "-y")

		// Verify branch was deleted
		branches := runGit(t, mainDir, "branch", "--list", "to-remove")
		if branches != "" {
			t.Errorf("Expected branch to-remove to be deleted, but found: %s", branches)
		}

		// List should no longer show it
		output = runGitWTSuccess(t, mainDir, "list")
		if containsString(output, "to-remove") {
			t.Errorf("Expected worktree to be gone from list, got: %s", output)
		}
	})
}

func TestDeleteJSON(t *testing.T) {
	t.Parallel()

	workspace := setupTestWorkspace(t)

	runGitWTSuccess(t, workspace, "clone", localRemote, "delete-json", "--timeout", "300")
	projectDir := filepath.Join(workspace, "delete-json")
	mainDir := filepath.Join(projectDir, "main")

	// Create and delete
	runGitWTSuccess(t, mainDir, "add", "json-delete", "--new")
	result := runGitWTJSON(t, mainDir, "delete", "json-delete", "-y")

	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success: true")
	}
}
