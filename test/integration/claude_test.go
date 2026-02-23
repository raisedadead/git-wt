package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigInitClaude(t *testing.T) {
	t.Parallel()

	t.Run("creates settings.json with hooks", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "claude-init", "--timeout", "300")
		projectDir := filepath.Join(workspace, "claude-init")
		mainDir := filepath.Join(projectDir, "main")

		runGitWTSuccess(t, mainDir, "config", "init", "--claude")

		settingsPath := filepath.Join(projectDir, ".claude", "settings.json")
		assertFileExists(t, settingsPath)

		data, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read settings.json: %v", err)
		}

		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("Failed to parse settings.json: %v", err)
		}

		hooks, ok := settings["hooks"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected hooks object in settings.json")
		}

		if _, ok := hooks["WorktreeCreate"]; !ok {
			t.Error("Expected WorktreeCreate hook in settings.json")
		}
		if _, ok := hooks["WorktreeRemove"]; !ok {
			t.Error("Expected WorktreeRemove hook in settings.json")
		}

		commandPath := filepath.Join(projectDir, ".claude", "commands", "wt.md")
		assertFileExists(t, commandPath)
	})

	t.Run("idempotent on second run", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "claude-idem", "--timeout", "300")
		projectDir := filepath.Join(workspace, "claude-idem")
		mainDir := filepath.Join(projectDir, "main")

		runGitWTSuccess(t, mainDir, "config", "init", "--claude")
		runGitWTSuccess(t, mainDir, "config", "init", "--claude")

		settingsPath := filepath.Join(projectDir, ".claude", "settings.json")
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("Failed to read settings.json: %v", err)
		}

		var settings map[string]interface{}
		if err := json.Unmarshal(data, &settings); err != nil {
			t.Fatalf("Failed to parse settings.json: %v", err)
		}

		hooks, ok := settings["hooks"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected hooks object in settings.json")
		}

		wtCreate, ok := hooks["WorktreeCreate"].([]interface{})
		if !ok {
			t.Fatal("Expected WorktreeCreate to be an array")
		}
		if len(wtCreate) != 1 {
			t.Errorf("Expected exactly 1 WorktreeCreate hook entry, got %d", len(wtCreate))
		}

		wtRemove, ok := hooks["WorktreeRemove"].([]interface{})
		if !ok {
			t.Fatal("Expected WorktreeRemove to be an array")
		}
		if len(wtRemove) != 1 {
			t.Errorf("Expected exactly 1 WorktreeRemove hook entry, got %d", len(wtRemove))
		}
	})
}

func TestConfigInitClaudeMigrate(t *testing.T) {
	t.Parallel()

	workspace := setupTestWorkspace(t)

	runGitWTSuccess(t, workspace, "clone", localRemote, "claude-migrate", "--timeout", "300")
	projectDir := filepath.Join(workspace, "claude-migrate")
	mainDir := filepath.Join(projectDir, "main")

	// Create .claude/plans/test.md inside the main worktree directory
	plansDir := filepath.Join(mainDir, ".claude", "plans")
	if err := os.MkdirAll(plansDir, 0755); err != nil {
		t.Fatalf("Failed to create plans dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(plansDir, "test.md"), []byte("# Test Plan\n"), 0644); err != nil {
		t.Fatalf("Failed to write test.md: %v", err)
	}

	runGitWTSuccess(t, mainDir, "config", "init", "--claude")

	// .claude/ directory should exist at project root
	assertDirExists(t, filepath.Join(projectDir, ".claude"))

	// main/.claude should be a symlink
	info, err := os.Lstat(filepath.Join(mainDir, ".claude"))
	if err != nil {
		t.Fatalf("Failed to lstat main/.claude: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Expected main/.claude to be a symlink")
	}

	// plans/test.md should be accessible via the symlink
	content, err := os.ReadFile(filepath.Join(mainDir, ".claude", "plans", "test.md"))
	if err != nil {
		t.Fatalf("Failed to read test.md via symlink: %v", err)
	}
	assertContains(t, string(content), "# Test Plan")
}

func TestAddCreatesClaudeSymlink(t *testing.T) {
	t.Parallel()

	workspace := setupTestWorkspace(t)

	runGitWTSuccess(t, workspace, "clone", localRemote, "claude-symlink", "--timeout", "300")
	projectDir := filepath.Join(workspace, "claude-symlink")
	mainDir := filepath.Join(projectDir, "main")

	// Create .claude/ directory at project root with a marker file
	claudeDir := filepath.Join(projectDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("Failed to create .claude dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "marker.txt"), []byte("marker"), 0644); err != nil {
		t.Fatalf("Failed to write marker file: %v", err)
	}

	runGitWTSuccess(t, mainDir, "add", "test-branch", "--new")

	// test-branch/.claude should be a symlink
	symlinkPath := filepath.Join(projectDir, "test-branch", ".claude")
	info, err := os.Lstat(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to lstat test-branch/.claude: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("Expected test-branch/.claude to be a symlink")
	}

	// Verify symlink target
	target, err := os.Readlink(symlinkPath)
	if err != nil {
		t.Fatalf("Failed to readlink: %v", err)
	}
	if target != "../.claude" {
		t.Errorf("Expected symlink target ../.claude, got %s", target)
	}

	// Marker file should be readable through the symlink
	content, err := os.ReadFile(filepath.Join(symlinkPath, "marker.txt"))
	if err != nil {
		t.Fatalf("Failed to read marker file through symlink: %v", err)
	}
	if string(content) != "marker" {
		t.Errorf("Expected marker content, got %q", string(content))
	}
}

func TestDeleteByPath(t *testing.T) {
	t.Parallel()

	t.Run("delete worktree using path flag", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "delete-path", "--timeout", "300")
		projectDir := filepath.Join(workspace, "delete-path")
		mainDir := filepath.Join(projectDir, "main")

		runGitWTSuccess(t, mainDir, "add", "path-delete-test", "--new")
		worktreePath := filepath.Join(projectDir, "path-delete-test")
		assertDirExists(t, worktreePath)

		// Resolve symlinks so the path matches what git worktree stores internally
		// (on macOS /var -> /private/var)
		realPath, err := filepath.EvalSymlinks(worktreePath)
		if err != nil {
			t.Fatalf("Failed to resolve symlinks: %v", err)
		}

		runGitWTSuccess(t, mainDir, "delete", "--path", realPath, "-y")

		assertDirNotExists(t, worktreePath)
	})

	t.Run("delete by path with json output", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "delete-path-json", "--timeout", "300")
		projectDir := filepath.Join(workspace, "delete-path-json")
		mainDir := filepath.Join(projectDir, "main")

		runGitWTSuccess(t, mainDir, "add", "json-path-test", "--new")
		worktreePath := filepath.Join(projectDir, "json-path-test")
		assertDirExists(t, worktreePath)

		// Resolve symlinks so the path matches what git worktree stores internally
		realPath, err := filepath.EvalSymlinks(worktreePath)
		if err != nil {
			t.Fatalf("Failed to resolve symlinks: %v", err)
		}

		result := runGitWTJSON(t, mainDir, "delete", "--path", realPath, "-y")

		if success, ok := result["success"].(bool); !ok || !success {
			t.Errorf("Expected success: true")
		}

		data, ok := result["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("Expected data object in JSON output")
		}

		if _, ok := data["branch"]; !ok {
			t.Error("Expected branch field in JSON output")
		}
		if _, ok := data["path"]; !ok {
			t.Error("Expected path field in JSON output")
		}

		assertDirNotExists(t, worktreePath)
	})
}

func TestConfigShowClaudeStatus(t *testing.T) {
	t.Parallel()

	workspace := setupTestWorkspace(t)

	runGitWTSuccess(t, workspace, "clone", localRemote, "claude-status", "--timeout", "300")
	projectDir := filepath.Join(workspace, "claude-status")
	mainDir := filepath.Join(projectDir, "main")

	runGitWTSuccess(t, mainDir, "config", "init", "--claude")

	result := runGitWTJSON(t, mainDir, "config", "show")

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data object in JSON output")
	}

	claudeIntegrated, ok := data["claude_integrated"].(bool)
	if !ok {
		t.Fatal("Expected claude_integrated field to be a boolean")
	}
	if !claudeIntegrated {
		t.Error("Expected claude_integrated to be true after running config init --claude")
	}
}
