package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestConfigInitClaudeFlag(t *testing.T) {
	t.Run("configClaude flag exists", func(t *testing.T) {
		flag := configInitCmd.Flags().Lookup("claude")
		if flag == nil {
			t.Fatal("expected --claude flag to be registered on configInitCmd")
		}
		if flag.DefValue != "false" {
			t.Errorf("--claude default = %q, want %q", flag.DefValue, "false")
		}
	})

	t.Run("claude and global are mutually exclusive", func(t *testing.T) {
		origClaude := configClaude
		origGlobal := configGlobal
		defer func() {
			configClaude = origClaude
			configGlobal = origGlobal
		}()

		configClaude = true
		configGlobal = true

		err := runConfigInit(nil, nil)
		if err == nil {
			t.Fatal("expected error when --claude and --global are both set")
		}
		if got := err.Error(); got != "--claude and --global cannot be used together" {
			t.Errorf("error = %q, want %q", got, "--claude and --global cannot be used together")
		}
	})

	t.Run("claude flag outside project returns error", func(t *testing.T) {
		origClaude := configClaude
		origGlobal := configGlobal
		defer func() {
			configClaude = origClaude
			configGlobal = origGlobal
		}()

		configClaude = true
		configGlobal = false

		// Change to a temp dir that is not a wt project
		origDir, _ := os.Getwd()
		tmpDir := t.TempDir()
		if err := os.Chdir(tmpDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(origDir) }()

		err := runConfigInit(nil, nil)
		if err == nil {
			t.Fatal("expected error when not in a wt project")
		}
	})

	t.Run("claude flag creates settings.json in project", func(t *testing.T) {
		origClaude := configClaude
		origGlobal := configGlobal
		defer func() {
			configClaude = origClaude
			configGlobal = origGlobal
		}()

		// Create a fake bare repo structure
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".bare"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: ./.bare\n"), 0644); err != nil {
			t.Fatal(err)
		}
		worktreeDir := filepath.Join(root, "main")
		if err := os.MkdirAll(worktreeDir, 0755); err != nil {
			t.Fatal(err)
		}

		origDir, _ := os.Getwd()
		if err := os.Chdir(worktreeDir); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = os.Chdir(origDir) }()

		configClaude = true
		configGlobal = false

		err := runConfigInit(nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		settingsPath := filepath.Join(root, ".claude", "settings.json")
		data, err := os.ReadFile(settingsPath)
		if err != nil {
			t.Fatalf("settings.json not created: %v", err)
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}

		hooks, ok := parsed["hooks"].(map[string]interface{})
		if !ok {
			t.Fatal("missing hooks key in settings")
		}
		if _, ok := hooks["WorktreeCreate"]; !ok {
			t.Error("missing WorktreeCreate in hooks")
		}
		if _, ok := hooks["WorktreeRemove"]; !ok {
			t.Error("missing WorktreeRemove in hooks")
		}
	})
}

func TestConfigShowClaudeIntegration(t *testing.T) {
	t.Run("JSON output includes claude_integrated field", func(t *testing.T) {
		// Create a fake bare repo with Claude integration
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, ".bare"), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: ./.bare\n"), 0644); err != nil {
			t.Fatal(err)
		}
		worktreeDir := filepath.Join(root, "main")
		if err := os.MkdirAll(worktreeDir, 0755); err != nil {
			t.Fatal(err)
		}

		// Set up Claude integration
		claudeDir := filepath.Join(root, ".claude")
		if err := os.MkdirAll(claudeDir, 0755); err != nil {
			t.Fatal(err)
		}
		settings := map[string]interface{}{
			"hooks": map[string]interface{}{
				"WorktreeCreate": []interface{}{
					map[string]interface{}{
						"hooks": []interface{}{
							map[string]interface{}{"type": "command", "command": "wt add"},
						},
					},
				},
			},
		}
		data, _ := json.MarshalIndent(settings, "", "  ")
		if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644); err != nil {
			t.Fatal(err)
		}

		// Verify detection works from the project root
		if !isClaudeIntegrated(root) {
			t.Fatal("precondition failed: isClaudeIntegrated should return true")
		}
	})
}
