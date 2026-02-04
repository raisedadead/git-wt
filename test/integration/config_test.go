package integration

import (
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	t.Run("show displays settings", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "config-test", "--timeout", "300")
		projectDir := filepath.Join(workspace, "config-test")
		mainDir := filepath.Join(projectDir, "main")

		stdout := runGitWTSuccess(t, mainDir, "config", "show")
		assertContains(t, stdout, "default_remote")
		assertContains(t, stdout, "git_timeout")
	})

	t.Run("repo config is detected", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "config-repo", "--timeout", "300")
		projectDir := filepath.Join(workspace, "config-repo")
		mainDir := filepath.Join(projectDir, "main")

		// Config show should work from within a worktree
		stdout := runGitWTSuccess(t, mainDir, "config", "show")
		// Should show default values since no project-level .wt.toml exists
		// (The .wt.toml in the worktree is a committed file, not project config)
		assertContains(t, stdout, "default_remote")
		assertContains(t, stdout, "origin")
	})
}

func TestConfigJSON(t *testing.T) {
	t.Parallel()

	workspace := setupTestWorkspace(t)

	runGitWTSuccess(t, workspace, "clone", localRemote, "config-json", "--timeout", "300")
	projectDir := filepath.Join(workspace, "config-json")
	mainDir := filepath.Join(projectDir, "main")

	result := runGitWTJSON(t, mainDir, "config", "show")

	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success: true")
	}

	data, ok := result["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data object")
	}

	// Should have config
	if _, ok := data["config"]; !ok {
		t.Errorf("Expected config field")
	}
}
