package integration

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig(t *testing.T) {
	t.Parallel()

	t.Run("show displays settings in table", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "config-test", "--timeout", "300")
		projectDir := filepath.Join(workspace, "config-test")
		mainDir := filepath.Join(projectDir, "main")

		stdout := runGitWTSuccess(t, mainDir, "config", "show")
		assertContains(t, stdout, "default_remote")
		assertContains(t, stdout, "git_timeout")
		// Table should have rounded borders
		assertContains(t, stdout, "╭")
		assertContains(t, stdout, "╰")
		// Should have settings table headers
		assertContains(t, stdout, "KEY")
		assertContains(t, stdout, "VALUE")
		assertContains(t, stdout, "SOURCE")
		// Should have hooks table headers
		assertContains(t, stdout, "EVENT")
		assertContains(t, stdout, "COMMANDS")
	})

	t.Run("repo config is detected", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "config-repo", "--timeout", "300")
		projectDir := filepath.Join(workspace, "config-repo")
		mainDir := filepath.Join(projectDir, "main")

		stdout := runGitWTSuccess(t, mainDir, "config", "show")
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

func TestConfigInit(t *testing.T) {
	t.Parallel()

	t.Run("creates_local_config", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "config-init-local", "--timeout", "300")
		projectDir := filepath.Join(workspace, "config-init-local")
		mainDir := filepath.Join(projectDir, "main")

		stdout := runGitWTSuccess(t, mainDir, "config", "init")
		assertContains(t, stdout, "Created")

		// Verify .wt.toml was created in the project root
		configPath := filepath.Join(projectDir, ".wt.toml")
		assertFileExists(t, configPath)
	})

	t.Run("fails_if_exists", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "config-init-exists", "--timeout", "300")
		projectDir := filepath.Join(workspace, "config-init-exists")
		mainDir := filepath.Join(projectDir, "main")

		// Create config first time
		runGitWTSuccess(t, mainDir, "config", "init")

		// Second attempt should fail
		_, stderr := runGitWTFail(t, mainDir, "config", "init")
		assertContains(t, stderr, "already exists")
	})

	t.Run("force_overwrites", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "config-init-force", "--timeout", "300")
		projectDir := filepath.Join(workspace, "config-init-force")
		mainDir := filepath.Join(projectDir, "main")

		// Create initial config
		runGitWTSuccess(t, mainDir, "config", "init")

		configPath := filepath.Join(projectDir, ".wt.toml")

		// Write some content to verify it gets overwritten
		if err := os.WriteFile(configPath, []byte("# old content\n"), 0644); err != nil {
			t.Fatal(err)
		}

		// Force overwrite
		stdout := runGitWTSuccess(t, mainDir, "config", "init", "--force")
		assertContains(t, stdout, "Created")

		// Verify file was overwritten (should have default template content)
		content, err := os.ReadFile(configPath)
		if err != nil {
			t.Fatal(err)
		}
		assertNotContains(t, string(content), "# old content")
	})
}

func TestConfigInitJSON(t *testing.T) {
	t.Parallel()

	workspace := setupTestWorkspace(t)

	runGitWTSuccess(t, workspace, "clone", localRemote, "config-init-json", "--timeout", "300")
	projectDir := filepath.Join(workspace, "config-init-json")
	mainDir := filepath.Join(projectDir, "main")

	result := runGitWTJSON(t, mainDir, "config", "init")

	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success: true")
	}

	data, ok := result["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data object")
	}

	// Check required fields
	if _, ok := data["path"]; !ok {
		t.Errorf("Missing required field: path")
	}

	// Verify the file was actually created
	configPath := filepath.Join(projectDir, ".wt.toml")
	assertFileExists(t, configPath)
}
