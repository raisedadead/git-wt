package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestClone(t *testing.T) {
	t.Parallel()

	t.Run("creates bare repo structure", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "test-project", "--timeout", "300")

		projectDir := filepath.Join(workspace, "test-project")
		assertDirExists(t, filepath.Join(projectDir, ".bare"))
		assertDirExists(t, filepath.Join(projectDir, "main"))
		assertFileExists(t, filepath.Join(projectDir, ".git"))
	})

	t.Run("clone with custom name", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "my-custom-name", "--timeout", "300")

		projectDir := filepath.Join(workspace, "my-custom-name")
		assertDirExists(t, filepath.Join(projectDir, ".bare"))
		assertDirExists(t, filepath.Join(projectDir, "main"))
	})

	t.Run("clone to existing directory fails", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		// Create existing directory
		existingDir := filepath.Join(workspace, "existing")
		if err := os.MkdirAll(existingDir, 0755); err != nil {
			t.Fatal(err)
		}

		runGitWTFail(t, workspace, "clone", localRemote, "existing", "--timeout", "300")
	})

	t.Run("clone with --force overwrites existing", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		// Create existing directory with a file
		existingDir := filepath.Join(workspace, "force-test")
		if err := os.MkdirAll(existingDir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(existingDir, "old-file.txt"), []byte("old"), 0644); err != nil {
			t.Fatal(err)
		}

		runGitWTSuccess(t, workspace, "clone", localRemote, "force-test", "--force", "--timeout", "300")

		// Should have bare repo structure now
		assertDirExists(t, filepath.Join(existingDir, ".bare"))
		assertDirExists(t, filepath.Join(existingDir, "main"))
	})
}

func TestCloneJSON(t *testing.T) {
	t.Parallel()

	workspace := setupTestWorkspace(t)

	result := runGitWTJSON(t, workspace, "clone", localRemote, "json-test", "--timeout", "300")

	// Check envelope structure
	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success: true, got %v", result["success"])
	}

	data, ok := result["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data object, got %T", result["data"])
	}

	// Check required fields
	requiredFields := []string{"project", "path", "bare_path", "default_branch", "worktree_path"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			t.Errorf("Missing required field: %s", field)
		}
	}
}
