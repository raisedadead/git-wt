package integration

import (
	"path/filepath"
	"testing"
)

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
