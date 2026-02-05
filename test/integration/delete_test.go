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

		// Delete main while in main - wt allows this
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

func TestDeleteWorktreeWithCustomDirectoryName(t *testing.T) {
	t.Parallel()

	t.Run("deletes worktree with custom directory name", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "delete-custom", "--timeout", "300")
		projectDir := filepath.Join(workspace, "delete-custom")
		mainDir := filepath.Join(projectDir, "main")

		// Create a worktree using raw git with a CUSTOM directory name
		// Branch: feature/long-migration-name
		// Directory: short-name (doesn't match flattened branch name feature-long-migration-name)
		customDir := filepath.Join(projectDir, "short-name")
		branchName := "feature/long-migration-name"
		runGit(t, mainDir, "worktree", "add", customDir, "-b", branchName)

		// Verify worktree was created
		assertDirExists(t, customDir)

		// Verify list shows it (wt list finds it by branch name)
		output := runGitWTSuccess(t, mainDir, "list")
		if !containsString(output, branchName) {
			t.Errorf("Expected branch %s in list output, got: %s", branchName, output)
		}

		// Delete by branch name - this is the bug: it should find the worktree
		// even though the directory name doesn't match the flattened branch name
		runGitWTSuccess(t, mainDir, "delete", branchName, "-y")

		// Verify worktree directory is gone
		assertDirNotExists(t, customDir)

		// Verify branch was deleted
		branches := runGit(t, mainDir, "branch", "--list", branchName)
		if branches != "" {
			t.Errorf("Expected branch %s to be deleted, but found: %s", branchName, branches)
		}
	})
}
