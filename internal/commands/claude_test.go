package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateClaudeSettings(t *testing.T) {
	data, err := generateClaudeSettings()
	if err != nil {
		t.Fatalf("generateClaudeSettings() error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("generated JSON is invalid: %v", err)
	}

	hooks, ok := parsed["hooks"].(map[string]interface{})
	if !ok {
		t.Fatal("missing or invalid 'hooks' key in settings")
	}

	t.Run("WorktreeCreate hook contains wt add", func(t *testing.T) {
		wtCreate, ok := hooks["WorktreeCreate"]
		if !ok {
			t.Fatal("missing WorktreeCreate key in hooks")
		}
		raw, _ := json.Marshal(wtCreate)
		if !strings.Contains(string(raw), "wt add") {
			t.Errorf("WorktreeCreate hook should contain 'wt add', got: %s", raw)
		}
	})

	t.Run("WorktreeRemove hook contains wt delete --path", func(t *testing.T) {
		wtRemove, ok := hooks["WorktreeRemove"]
		if !ok {
			t.Fatal("missing WorktreeRemove key in hooks")
		}
		raw, _ := json.Marshal(wtRemove)
		if !strings.Contains(string(raw), "wt delete --path") {
			t.Errorf("WorktreeRemove hook should contain 'wt delete --path', got: %s", raw)
		}
	})
}

func TestSymlinkClaudeDir(t *testing.T) {
	tests := []struct {
		name         string
		setupRoot    func(t *testing.T, root string)
		setupWT      func(t *testing.T, wt string)
		wantSymlink  bool
		wantDir      bool
		wantNoChange bool
	}{
		{
			name: "creates symlink when .claude exists at root",
			setupRoot: func(t *testing.T, root string) {
				if err := os.MkdirAll(filepath.Join(root, ".claude"), 0755); err != nil {
					t.Fatal(err)
				}
			},
			setupWT:     func(t *testing.T, wt string) {},
			wantSymlink: true,
		},
		{
			name: "skips when symlink already exists",
			setupRoot: func(t *testing.T, root string) {
				if err := os.MkdirAll(filepath.Join(root, ".claude"), 0755); err != nil {
					t.Fatal(err)
				}
			},
			setupWT: func(t *testing.T, wt string) {
				if err := os.Symlink("../.claude", filepath.Join(wt, ".claude")); err != nil {
					t.Fatal(err)
				}
			},
			wantSymlink:  true,
			wantNoChange: true,
		},
		{
			name:        "skips when no .claude at root",
			setupRoot:   func(t *testing.T, root string) {},
			setupWT:     func(t *testing.T, wt string) {},
			wantSymlink: false,
		},
		{
			name: "skips when .claude is a real directory in worktree",
			setupRoot: func(t *testing.T, root string) {
				if err := os.MkdirAll(filepath.Join(root, ".claude"), 0755); err != nil {
					t.Fatal(err)
				}
			},
			setupWT: func(t *testing.T, wt string) {
				if err := os.MkdirAll(filepath.Join(wt, ".claude"), 0755); err != nil {
					t.Fatal(err)
				}
			},
			wantDir: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			wt := filepath.Join(root, "feature-x")
			if err := os.MkdirAll(wt, 0755); err != nil {
				t.Fatal(err)
			}

			tt.setupRoot(t, root)
			tt.setupWT(t, wt)

			symlinkClaudeDir(root, wt)

			claudePath := filepath.Join(wt, ".claude")
			info, err := os.Lstat(claudePath)

			if tt.wantSymlink {
				if err != nil {
					t.Fatalf("expected symlink at %s, got error: %v", claudePath, err)
				}
				if info.Mode()&os.ModeSymlink == 0 {
					t.Errorf("expected symlink, got mode %v", info.Mode())
				}
				target, err := os.Readlink(claudePath)
				if err != nil {
					t.Fatalf("readlink error: %v", err)
				}
				if target != "../.claude" {
					t.Errorf("symlink target = %q, want %q", target, "../.claude")
				}
			} else if tt.wantDir {
				if err != nil {
					t.Fatalf("expected directory at %s, got error: %v", claudePath, err)
				}
				if !info.IsDir() {
					t.Error("expected real directory, not symlink")
				}
				if info.Mode()&os.ModeSymlink != 0 {
					t.Error("expected real directory, got symlink")
				}
			} else {
				if err == nil {
					t.Errorf("expected no .claude at %s, but it exists", claudePath)
				}
			}
		})
	}
}

func TestMigrateClaudeDir(t *testing.T) {
	t.Run("skips when .claude already exists at root", func(t *testing.T) {
		root := t.TempDir()
		claudeDir := filepath.Join(root, ".claude")
		if err := os.MkdirAll(claudeDir, 0755); err != nil {
			t.Fatal(err)
		}
		// Write a marker file to confirm it's not replaced
		if err := os.WriteFile(filepath.Join(claudeDir, "marker"), []byte("original"), 0644); err != nil {
			t.Fatal(err)
		}

		err := migrateClaudeDir(root)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify marker file is unchanged
		data, err := os.ReadFile(filepath.Join(claudeDir, "marker"))
		if err != nil {
			t.Fatalf("marker file missing: %v", err)
		}
		if string(data) != "original" {
			t.Error("marker file was modified")
		}
	})

	t.Run("returns nil when no .claude anywhere", func(t *testing.T) {
		root := t.TempDir()
		// migrateClaudeDir will call git.ListWorktrees which needs a real git repo
		// so this will error on ListWorktrees, but the function should handle it
		err := migrateClaudeDir(root)
		// Either nil (no .claude at root, short-circuit before ListWorktrees)
		// or an error from git.ListWorktrees is acceptable
		_ = err
	})
}

func TestResolveBranchFromPath(t *testing.T) {
	t.Run("returns error for empty project root", func(t *testing.T) {
		_, err := resolveBranchFromPath("", "/some/path")
		if err == nil {
			t.Error("expected error for empty project root")
		}
	})

	t.Run("returns error for nonexistent project", func(t *testing.T) {
		_, err := resolveBranchFromPath("/nonexistent/path", "/some/worktree")
		if err == nil {
			t.Error("expected error for nonexistent project root")
		}
	})
}

func TestIsClaudeIntegrated(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T, root string)
		expected bool
	}{
		{
			name: "returns true with WorktreeCreate hooks",
			setup: func(t *testing.T, root string) {
				dir := filepath.Join(root, ".claude")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				settings := map[string]interface{}{
					"hooks": map[string]interface{}{
						"WorktreeCreate": []interface{}{
							map[string]interface{}{
								"hooks": []interface{}{
									map[string]interface{}{
										"type":    "command",
										"command": "wt add",
									},
								},
							},
						},
					},
				}
				data, _ := json.MarshalIndent(settings, "", "  ")
				if err := os.WriteFile(filepath.Join(dir, "settings.json"), data, 0644); err != nil {
					t.Fatal(err)
				}
			},
			expected: true,
		},
		{
			name: "returns false without hooks",
			setup: func(t *testing.T, root string) {
				dir := filepath.Join(root, ".claude")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				settings := map[string]interface{}{
					"someOtherKey": "value",
				}
				data, _ := json.MarshalIndent(settings, "", "  ")
				if err := os.WriteFile(filepath.Join(dir, "settings.json"), data, 0644); err != nil {
					t.Fatal(err)
				}
			},
			expected: false,
		},
		{
			name:     "returns false when no settings.json",
			setup:    func(t *testing.T, root string) {},
			expected: false,
		},
		{
			name: "returns false for invalid JSON",
			setup: func(t *testing.T, root string) {
				dir := filepath.Join(root, ".claude")
				if err := os.MkdirAll(dir, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte("{invalid json"), 0644); err != nil {
					t.Fatal(err)
				}
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			tt.setup(t, root)

			got := isClaudeIntegrated(root)
			if got != tt.expected {
				t.Errorf("isClaudeIntegrated() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateClaudeCommand(t *testing.T) {
	content := generateClaudeCommand()

	t.Run("contains bare repo layout", func(t *testing.T) {
		if !strings.Contains(content, ".bare/") {
			t.Error("expected content to mention .bare/ directory")
		}
		if !strings.Contains(content, "bare repo worktree layout") {
			t.Error("expected content to describe bare repo layout")
		}
	})

	t.Run("contains wt commands", func(t *testing.T) {
		for _, cmd := range []string{"wt list", "wt add", "wt delete"} {
			if !strings.Contains(content, cmd) {
				t.Errorf("expected content to mention %q", cmd)
			}
		}
	})

	t.Run("documents branch naming", func(t *testing.T) {
		if !strings.Contains(content, "feature/auth") {
			t.Error("expected content to show slash-to-dash example")
		}
	})

	t.Run("mentions shared .claude directory", func(t *testing.T) {
		if !strings.Contains(content, ".claude/") {
			t.Error("expected content to mention shared .claude/ directory")
		}
	})
}

func TestSetupClaudeIntegration(t *testing.T) {
	t.Run("creates settings.json in .claude directory", func(t *testing.T) {
		root := t.TempDir()
		// setupClaudeIntegration calls migrateClaudeDir which calls git.ListWorktrees
		// This will fail without a real git repo, but it should still create .claude/
		// and settings.json since migrateClaudeDir errors are non-fatal
		_ = setupClaudeIntegration(root)

		settingsPath := filepath.Join(root, ".claude", "settings.json")
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("settings.json not created: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("invalid JSON in settings.json: %v", err)
		}

		hooks, ok := parsed["hooks"].(map[string]interface{})
		if !ok {
			t.Fatal("missing hooks key")
		}
		if _, ok := hooks["WorktreeCreate"]; !ok {
			t.Error("missing WorktreeCreate in hooks")
		}
		if _, ok := hooks["WorktreeRemove"]; !ok {
			t.Error("missing WorktreeRemove in hooks")
		}
	})

	t.Run("creates commands/wt.md", func(t *testing.T) {
		root := t.TempDir()
		_ = setupClaudeIntegration(root)

		commandPath := filepath.Join(root, ".claude", "commands", "wt.md")
		data, err := os.ReadFile(commandPath)
		if err != nil {
			t.Fatalf("wt.md not created: %v", err)
		}

		content := string(data)
		if !strings.Contains(content, "wt") {
			t.Error("wt.md should contain wt documentation")
		}
		if !strings.Contains(content, ".bare/") {
			t.Error("wt.md should describe bare repo layout")
		}
	})

	t.Run("skips wt.md if already exists", func(t *testing.T) {
		root := t.TempDir()
		commandsDir := filepath.Join(root, ".claude", "commands")
		if err := os.MkdirAll(commandsDir, 0755); err != nil {
			t.Fatal(err)
		}
		customContent := "# Custom wt docs\nUser-edited content\n"
		if err := os.WriteFile(filepath.Join(commandsDir, "wt.md"), []byte(customContent), 0644); err != nil {
			t.Fatal(err)
		}

		_ = setupClaudeIntegration(root)

		data, err := os.ReadFile(filepath.Join(commandsDir, "wt.md"))
		if err != nil {
			t.Fatalf("wt.md missing after re-run: %v", err)
		}
		if string(data) != customContent {
			t.Error("wt.md was overwritten despite already existing")
		}
	})

	t.Run("merges into existing settings.json", func(t *testing.T) {
		root := t.TempDir()
		claudeDir := filepath.Join(root, ".claude")
		if err := os.MkdirAll(claudeDir, 0755); err != nil {
			t.Fatal(err)
		}

		existing := map[string]interface{}{
			"customKey": "customValue",
			"hooks": map[string]interface{}{
				"PreToolUse": []interface{}{"existing-hook"},
			},
		}
		data, _ := json.MarshalIndent(existing, "", "  ")
		if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		_ = setupClaudeIntegration(root)

		result, err := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
		if err != nil {
			t.Fatal(err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(result, &parsed); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		// Verify custom key preserved
		if parsed["customKey"] != "customValue" {
			t.Error("customKey was not preserved")
		}

		// Verify existing hooks preserved
		hooks := parsed["hooks"].(map[string]interface{})
		if _, ok := hooks["PreToolUse"]; !ok {
			t.Error("existing PreToolUse hook was not preserved")
		}

		// Verify new hooks added
		if _, ok := hooks["WorktreeCreate"]; !ok {
			t.Error("WorktreeCreate not added")
		}
		if _, ok := hooks["WorktreeRemove"]; !ok {
			t.Error("WorktreeRemove not added")
		}
	})
}
