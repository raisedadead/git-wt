package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.WorktreeRoot == "" {
		t.Error("expected default worktree_root")
	}
}

func TestGetConfigPath(t *testing.T) {
	// Test XDG_CONFIG_HOME takes precedence
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg")
	path := GetConfigPath()
	expected := "/tmp/test-xdg/git-wt/config.toml"
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestLoadConfig_NoFile(t *testing.T) {
	cfg, err := Load("/nonexistent/config.toml")
	if err != nil {
		t.Fatalf("expected no error for missing config, got %v", err)
	}
	if cfg.WorktreeRoot == "" {
		t.Error("expected default worktree_root")
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `worktree_root = "/custom/path"

[hooks]
post_clone = ["echo hello"]
post_add = ["echo world", "echo again"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.WorktreeRoot != "/custom/path" {
		t.Errorf("expected /custom/path, got %s", cfg.WorktreeRoot)
	}

	if len(cfg.Hooks.PostClone) != 1 {
		t.Errorf("expected 1 post_clone hook, got %d", len(cfg.Hooks.PostClone))
	}

	if len(cfg.Hooks.PostAdd) != 2 {
		t.Errorf("expected 2 post_add hooks, got %d", len(cfg.Hooks.PostAdd))
	}
}
