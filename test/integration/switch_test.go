package integration

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSwitch(t *testing.T) {
	t.Parallel()

	t.Run("switch by branch name", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "switch-test", "--timeout", "300")
		projectDir := filepath.Join(workspace, "switch-test")
		mainDir := filepath.Join(projectDir, "main")

		// Add another worktree
		runGitWTSuccess(t, mainDir, "add", "feature-test", "--new")

		// Switch by branch name should output the path
		stdout := runGitWTSuccess(t, mainDir, "switch", "feature-test")
		expectedPath := filepath.Join(projectDir, "feature-test")
		if !strings.Contains(strings.TrimSpace(stdout), expectedPath) {
			t.Errorf("Expected path to contain %q, got %q", expectedPath, stdout)
		}
	})

	t.Run("switch with slash branch name", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "switch-slash", "--timeout", "300")
		projectDir := filepath.Join(workspace, "switch-slash")
		mainDir := filepath.Join(projectDir, "main")

		// Add worktree with slash in branch name (use unique name to avoid conflicts)
		runGitWTSuccess(t, mainDir, "add", "test/switch-slash", "--new")

		// Switch by exact branch name
		stdout := runGitWTSuccess(t, mainDir, "switch", "test/switch-slash")
		expectedPath := filepath.Join(projectDir, "test-switch-slash")
		if !strings.Contains(strings.TrimSpace(stdout), expectedPath) {
			t.Errorf("Expected path to contain %q, got %q", expectedPath, stdout)
		}
	})

	t.Run("switch by flattened name", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "switch-flat", "--timeout", "300")
		projectDir := filepath.Join(workspace, "switch-flat")
		mainDir := filepath.Join(projectDir, "main")

		// Add worktree with slash in branch name
		runGitWTSuccess(t, mainDir, "add", "feature/deep/nested", "--new")

		// Switch by flattened directory name
		stdout := runGitWTSuccess(t, mainDir, "switch", "feature-deep-nested")
		expectedPath := filepath.Join(projectDir, "feature-deep-nested")
		if !strings.Contains(strings.TrimSpace(stdout), expectedPath) {
			t.Errorf("Expected path to contain %q, got %q", expectedPath, stdout)
		}
	})

	t.Run("switch to main", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "switch-main", "--timeout", "300")
		projectDir := filepath.Join(workspace, "switch-main")
		mainDir := filepath.Join(projectDir, "main")

		// Add another worktree and switch back to main
		runGitWTSuccess(t, mainDir, "add", "feature-x", "--new")
		featureDir := filepath.Join(projectDir, "feature-x")

		stdout := runGitWTSuccess(t, featureDir, "switch", "main")
		expectedPath := filepath.Join(projectDir, "main")
		if !strings.Contains(strings.TrimSpace(stdout), expectedPath) {
			t.Errorf("Expected path to contain %q, got %q", expectedPath, stdout)
		}
	})

	t.Run("switch nonexistent fails", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "switch-fail", "--timeout", "300")
		projectDir := filepath.Join(workspace, "switch-fail")
		mainDir := filepath.Join(projectDir, "main")

		_, stderr := runGitWTFail(t, mainDir, "switch", "nonexistent-branch")
		assertContains(t, stderr, "worktree not found")
	})

	t.Run("switch outside project fails", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		_, stderr := runGitWTFail(t, workspace, "switch", "main")
		assertContains(t, stderr, "not in a wt project")
	})
}

func TestSwitchJSON(t *testing.T) {
	t.Parallel()

	t.Run("switch outputs JSON with branch and path", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "switch-json", "--timeout", "300")
		projectDir := filepath.Join(workspace, "switch-json")
		mainDir := filepath.Join(projectDir, "main")

		// Add a worktree
		runGitWTSuccess(t, mainDir, "add", "feature-json", "--new")

		result := runGitWTJSON(t, mainDir, "switch", "feature-json")

		// Check envelope
		if success, ok := result["success"].(bool); !ok || !success {
			t.Errorf("Expected success: true")
		}

		data, ok := result["data"].(map[string]any)
		if !ok {
			t.Fatalf("Expected data object")
		}

		branch, ok := data["branch"].(string)
		if !ok || branch != "feature-json" {
			t.Errorf("Expected branch 'feature-json', got %v", data["branch"])
		}

		path, ok := data["path"].(string)
		if !ok || !strings.HasSuffix(path, "feature-json") {
			t.Errorf("Expected path ending with 'feature-json', got %v", data["path"])
		}
	})

	t.Run("switch JSON error for nonexistent", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "switch-json-err", "--timeout", "300")
		projectDir := filepath.Join(workspace, "switch-json-err")
		mainDir := filepath.Join(projectDir, "main")

		stdout, _, code := runGitWT(t, mainDir, "switch", "nonexistent", "--json")

		// Should fail with non-zero exit code
		if code == 0 {
			t.Errorf("Expected non-zero exit code for nonexistent worktree")
		}

		// Should still be valid JSON with success: false
		if !strings.Contains(stdout, `"success": false`) {
			t.Errorf("Expected JSON with success: false, got %s", stdout)
		}
		if !strings.Contains(stdout, "worktree not found") {
			t.Errorf("Expected error message about worktree not found, got %s", stdout)
		}
	})

	t.Run("switch JSON requires branch argument", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "switch-json-noarg", "--timeout", "300")
		projectDir := filepath.Join(workspace, "switch-json-noarg")
		mainDir := filepath.Join(projectDir, "main")

		stdout, _, code := runGitWT(t, mainDir, "switch", "--json")
		if code == 0 {
			t.Fatalf("Expected failure when no branch specified with --json")
		}

		if !strings.Contains(stdout, `"success": false`) {
			t.Errorf("Expected JSON with success: false, got %s", stdout)
		}
		if !strings.Contains(stdout, "worktree name required") {
			t.Errorf("Expected error about worktree name required, got %s", stdout)
		}
	})
}
