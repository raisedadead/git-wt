package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Empty WorktreeRoot means use current directory
	if cfg.WorktreeRoot != "" {
		t.Errorf("expected empty worktree_root (use current dir), got %s", cfg.WorktreeRoot)
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
	// Empty WorktreeRoot means use current directory
	if cfg.WorktreeRoot != "" {
		t.Errorf("expected empty worktree_root (use current dir), got %s", cfg.WorktreeRoot)
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

func TestDefaultConfig_AllFields(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.WorktreeRoot != "" {
		t.Errorf("expected empty worktree_root, got %s", cfg.WorktreeRoot)
	}
	if cfg.DefaultRemote != "origin" {
		t.Errorf("expected default_remote 'origin', got %s", cfg.DefaultRemote)
	}
	if cfg.DefaultBaseBranch != "" {
		t.Errorf("expected empty default_base_branch, got %s", cfg.DefaultBaseBranch)
	}
	if cfg.BranchTemplate != "{{type}}-{{number}}-{{slug}}" {
		t.Errorf("expected default branch_template, got %s", cfg.BranchTemplate)
	}
	if cfg.GitTimeout != 120 {
		t.Errorf("expected git_timeout 120, got %d", cfg.GitTimeout)
	}
	if cfg.GitLongTimeout != 600 {
		t.Errorf("expected git_long_timeout 600, got %d", cfg.GitLongTimeout)
	}
	if cfg.HookTimeout != 30 {
		t.Errorf("expected hook_timeout 30, got %d", cfg.HookTimeout)
	}
}

func TestLoadConfig_NewFields(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `default_remote = "upstream"
default_base_branch = "develop"
branch_template = "feat/{{type}}-{{number}}"
git_timeout = 180
git_long_timeout = 900
hook_timeout = 60
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.DefaultRemote != "upstream" {
		t.Errorf("expected 'upstream', got %s", cfg.DefaultRemote)
	}
	if cfg.DefaultBaseBranch != "develop" {
		t.Errorf("expected 'develop', got %s", cfg.DefaultBaseBranch)
	}
	if cfg.BranchTemplate != "feat/{{type}}-{{number}}" {
		t.Errorf("expected custom template, got %s", cfg.BranchTemplate)
	}
	if cfg.GitTimeout != 180 {
		t.Errorf("expected 180, got %d", cfg.GitTimeout)
	}
	if cfg.GitLongTimeout != 900 {
		t.Errorf("expected 900, got %d", cfg.GitLongTimeout)
	}
	if cfg.HookTimeout != 60 {
		t.Errorf("expected 60, got %d", cfg.HookTimeout)
	}
}

func TestGetRepoConfigPath(t *testing.T) {
	tmpDir := t.TempDir()
	path := GetRepoConfigPath(tmpDir)
	expected := filepath.Join(tmpDir, ".git-wt.toml")
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}

func TestMergeConfig(t *testing.T) {
	base := &Config{
		DefaultRemote:  "origin",
		GitTimeout:     120,
		HookTimeout:    30,
		BranchTemplate: "{{type}}-{{number}}-{{slug}}",
	}
	override := &Config{
		DefaultRemote: "upstream",
		GitTimeout:    180,
	}

	merged := MergeConfig(base, override)

	if merged.DefaultRemote != "upstream" {
		t.Errorf("expected 'upstream', got %s", merged.DefaultRemote)
	}
	if merged.GitTimeout != 180 {
		t.Errorf("expected 180, got %d", merged.GitTimeout)
	}
	// Non-overridden values should come from base
	if merged.HookTimeout != 30 {
		t.Errorf("expected 30, got %d", merged.HookTimeout)
	}
	if merged.BranchTemplate != "{{type}}-{{number}}-{{slug}}" {
		t.Errorf("expected default template, got %s", merged.BranchTemplate)
	}
}

func TestLoadWithRepo(t *testing.T) {
	globalDir := t.TempDir()
	repoDir := t.TempDir()

	globalConfig := filepath.Join(globalDir, "config.toml")
	repoConfig := filepath.Join(repoDir, ".git-wt.toml")

	globalContent := `git_timeout = 180`
	if err := os.WriteFile(globalConfig, []byte(globalContent), 0644); err != nil {
		t.Fatal(err)
	}

	repoContent := `default_remote = "upstream"`
	if err := os.WriteFile(repoConfig, []byte(repoContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadWithRepo(globalConfig, repoDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.DefaultRemote != "upstream" {
		t.Errorf("expected 'upstream', got %s", cfg.DefaultRemote)
	}
	if cfg.GitTimeout != 180 {
		t.Errorf("expected 180, got %d", cfg.GitTimeout)
	}
	if cfg.HookTimeout != 30 {
		t.Errorf("expected default 30, got %d", cfg.HookTimeout)
	}
}

func TestLoadEffective_Sources(t *testing.T) {
	globalDir := t.TempDir()
	repoDir := t.TempDir()

	globalConfig := filepath.Join(globalDir, "config.toml")
	repoConfig := filepath.Join(repoDir, ".git-wt.toml")

	globalContent := `git_timeout = 180`
	if err := os.WriteFile(globalConfig, []byte(globalContent), 0644); err != nil {
		t.Fatal(err)
	}

	repoContent := `default_remote = "upstream"`
	if err := os.WriteFile(repoConfig, []byte(repoContent), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, sources, err := LoadEffective(globalConfig, repoDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.DefaultRemote != "upstream" {
		t.Errorf("expected 'upstream', got %s", cfg.DefaultRemote)
	}

	if sources["default_remote"] != repoConfig {
		t.Errorf("expected source %s, got %s", repoConfig, sources["default_remote"])
	}
	if sources["git_timeout"] != globalConfig {
		t.Errorf("expected source %s, got %s", globalConfig, sources["git_timeout"])
	}
	if sources["hook_timeout"] != "default" {
		t.Errorf("expected source 'default', got %s", sources["hook_timeout"])
	}
}

func TestGenerateConfigTemplate(t *testing.T) {
	template := GenerateConfigTemplate()

	// Should contain key sections
	if !strings.Contains(template, "default_remote") {
		t.Error("template should contain default_remote")
	}
	if !strings.Contains(template, "hook_timeout") {
		t.Error("template should contain hook_timeout")
	}
	if !strings.Contains(template, "# default_remote") {
		t.Error("options should be commented out")
	}
	if !strings.Contains(template, "git-wt configuration") {
		t.Error("should have header comment")
	}
}
