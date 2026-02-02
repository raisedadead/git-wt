package integration

import (
	"path/filepath"
	"testing"
)

func TestPrune(t *testing.T) {
	t.Parallel()

	t.Run("dry run shows no stale by default", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "prune-test", "--timeout", "300")
		projectDir := filepath.Join(workspace, "prune-test")
		mainDir := filepath.Join(projectDir, "main")

		// Should run successfully even with no stale worktrees
		runGitWTSuccess(t, mainDir, "prune", "--dry-run")
	})

	t.Run("does not flag .bare as stale", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "prune-bare", "--timeout", "300")
		projectDir := filepath.Join(workspace, "prune-bare")
		mainDir := filepath.Join(projectDir, "main")

		stdout := runGitWTSuccess(t, mainDir, "prune", "--dry-run")
		assertNotContains(t, stdout, ".bare")
	})
}

func TestPruneJSON(t *testing.T) {
	t.Parallel()

	workspace := setupTestWorkspace(t)

	runGitWTSuccess(t, workspace, "clone", localRemote, "prune-json", "--timeout", "300")
	projectDir := filepath.Join(workspace, "prune-json")
	mainDir := filepath.Join(projectDir, "main")

	result := runGitWTJSON(t, mainDir, "prune", "--dry-run")

	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success: true")
	}

	if _, ok := result["data"]; !ok {
		t.Errorf("Expected data field")
	}
}
