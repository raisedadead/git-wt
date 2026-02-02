package integration

import (
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
