package integration

import (
	"path/filepath"
	"testing"
)

func TestAdd(t *testing.T) {
	t.Parallel()

	t.Run("new local branch", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "add-new", "--timeout", "300")
		projectDir := filepath.Join(workspace, "add-new")
		mainDir := filepath.Join(projectDir, "main")

		runGitWTSuccess(t, mainDir, "add", "feature-new", "--new")

		assertDirExists(t, filepath.Join(projectDir, "feature-new"))
	})

	t.Run("track existing remote branch", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "add-track", "--timeout", "300")
		projectDir := filepath.Join(workspace, "add-track")
		mainDir := filepath.Join(projectDir, "main")

		runGitWTSuccess(t, mainDir, "add", "feature/auth", "--track")

		// Branch with slash becomes dash in directory name
		assertDirExists(t, filepath.Join(projectDir, "feature-auth"))
	})

	t.Run("branch with slash flattens to dash", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "add-slash", "--timeout", "300")
		projectDir := filepath.Join(workspace, "add-slash")
		mainDir := filepath.Join(projectDir, "main")

		runGitWTSuccess(t, mainDir, "add", "feature/nested/deep", "--track")

		assertDirExists(t, filepath.Join(projectDir, "feature-nested-deep"))
	})

	t.Run("track nonexistent branch fails", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "add-noexist", "--timeout", "300")
		projectDir := filepath.Join(workspace, "add-noexist")
		mainDir := filepath.Join(projectDir, "main")

		runGitWTFail(t, mainDir, "add", "nonexistent-branch-xyz", "--track")
	})

	t.Run("duplicate worktree fails", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "add-dup", "--timeout", "300")
		projectDir := filepath.Join(workspace, "add-dup")
		mainDir := filepath.Join(projectDir, "main")

		// Create first
		runGitWTSuccess(t, mainDir, "add", "duplicate-test", "--new")

		// Try to create again - should fail
		runGitWTFail(t, mainDir, "add", "duplicate-test", "--new")
	})
}

func TestAddJSON(t *testing.T) {
	t.Parallel()

	workspace := setupTestWorkspace(t)

	runGitWTSuccess(t, workspace, "clone", localRemote, "add-json", "--timeout", "300")
	projectDir := filepath.Join(workspace, "add-json")
	mainDir := filepath.Join(projectDir, "main")

	result := runGitWTJSON(t, mainDir, "add", "json-feature", "--new")

	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success: true")
	}

	data, ok := result["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data object")
	}

	// Check required fields
	requiredFields := []string{"branch", "path"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			t.Errorf("Missing required field: %s", field)
		}
	}
}
