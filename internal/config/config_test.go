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
	expected := "/tmp/test-xdg/wt/config.toml"
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
	expected := filepath.Join(tmpDir, ".wt.toml")
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
	repoConfig := filepath.Join(repoDir, ".wt.toml")

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
	repoConfig := filepath.Join(repoDir, ".wt.toml")

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
	if !strings.Contains(template, "wt configuration") {
		t.Error("should have header comment")
	}
}

func TestAutoTrackConfig(t *testing.T) {
	// Test that AutoTrack defaults to false
	cfg := DefaultConfig()
	if cfg.AutoTrack == nil || *cfg.AutoTrack != false {
		t.Errorf("expected AutoTrack default false, got %v", cfg.AutoTrack)
	}

	// Test parsing from TOML with true
	tomlContent := `auto_track = true`
	tmpFile, err := os.CreateTemp("", "config-*.toml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.WriteString(tomlContent); err != nil {
		t.Fatal(err)
	}
	_ = tmpFile.Close()

	loaded, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if loaded.AutoTrack == nil || *loaded.AutoTrack != true {
		t.Errorf("expected AutoTrack true after loading, got %v", loaded.AutoTrack)
	}
}

func TestAutoTrackOverride(t *testing.T) {
	// Test that repo config can override global auto_track=true with auto_track=false
	globalCfg := &Config{AutoTrack: ptrBool(true)}
	repoCfg := &Config{AutoTrack: ptrBool(false)}

	merged := MergeConfig(globalCfg, repoCfg)
	if merged.AutoTrack == nil || *merged.AutoTrack != false {
		t.Errorf("expected repo auto_track=false to override global auto_track=true, got %v", merged.AutoTrack)
	}

	// Test that unset repo config doesn't override global
	globalCfg2 := &Config{AutoTrack: ptrBool(true)}
	repoCfg2 := &Config{AutoTrack: nil} // not set

	merged2 := MergeConfig(globalCfg2, repoCfg2)
	if merged2.AutoTrack == nil || *merged2.AutoTrack != true {
		t.Errorf("expected global auto_track=true to persist when repo unset, got %v", merged2.AutoTrack)
	}
}

func TestGetHooksDir(t *testing.T) {
	dir := GetHooksDir()
	if !strings.HasSuffix(dir, "wt/hooks") {
		t.Errorf("expected hooks dir to end with wt/hooks, got %s", dir)
	}
}

func TestGetCommunityHooksDir(t *testing.T) {
	dir := GetCommunityHooksDir()
	if !strings.HasSuffix(dir, "hooks/community") {
		t.Errorf("expected community hooks dir, got %s", dir)
	}
}

func TestGetCustomHooksDir(t *testing.T) {
	dir := GetCustomHooksDir()
	if !strings.HasSuffix(dir, "hooks/custom") {
		t.Errorf("expected custom hooks dir, got %s", dir)
	}
}

func TestInstallBundledHooks(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	err := InstallBundledHooks()
	if err != nil {
		t.Fatalf("InstallBundledHooks failed: %v", err)
	}

	communityDir := filepath.Join(tmpDir, "wt", "hooks", "community")
	customDir := filepath.Join(tmpDir, "wt", "hooks", "custom")

	if _, err := os.Stat(communityDir); os.IsNotExist(err) {
		t.Errorf("community hooks dir was not created")
	}
	if _, err := os.Stat(customDir); os.IsNotExist(err) {
		t.Errorf("custom hooks dir was not created")
	}

	ghDefaultPath := filepath.Join(communityDir, "gh-default.sh")
	if _, err := os.Stat(ghDefaultPath); os.IsNotExist(err) {
		t.Errorf("gh-default.sh was not copied")
	}

	direnvPath := filepath.Join(communityDir, "direnv.sh")
	if _, err := os.Stat(direnvPath); os.IsNotExist(err) {
		t.Errorf("direnv.sh was not copied")
	}

	info, err := os.Stat(ghDefaultPath)
	if err != nil {
		t.Fatalf("failed to stat gh-default.sh: %v", err)
	}
	if info.Mode().Perm()&0100 == 0 {
		t.Errorf("gh-default.sh should be executable")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	cfg := &Config{
		DefaultRemote: "upstream",
		GitTimeout:    180,
	}

	err := SaveConfig(cfg, configPath)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists and can be loaded
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.DefaultRemote != "upstream" {
		t.Errorf("expected 'upstream', got %s", loaded.DefaultRemote)
	}
}

func TestAddHookToConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Start with empty config
	cfg := DefaultConfig()
	if err := SaveConfig(cfg, configPath); err != nil {
		t.Fatal(err)
	}

	// Add a hook (path, hookName, event)
	if err := AddHookToConfig(configPath, "echo hello", "post_clone"); err != nil {
		t.Fatalf("AddHookToConfig failed: %v", err)
	}

	// Verify hook was added
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Hooks.PostClone) != 1 || loaded.Hooks.PostClone[0] != "echo hello" {
		t.Errorf("expected ['echo hello'], got %v", loaded.Hooks.PostClone)
	}
}

func TestRemoveHookFromConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	content := `[hooks]
post_clone = ["echo hello", "echo world"]
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Remove a hook (removes from all events)
	if err := RemoveHookFromConfig(configPath, "echo hello"); err != nil {
		t.Fatalf("RemoveHookFromConfig failed: %v", err)
	}

	// Verify hook was removed
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded.Hooks.PostClone) != 1 || loaded.Hooks.PostClone[0] != "echo world" {
		t.Errorf("expected ['echo world'], got %v", loaded.Hooks.PostClone)
	}
}

func TestDefaultWorkflows(t *testing.T) {
	workflows := DefaultWorkflows()

	// Should have default workflows
	if len(workflows) == 0 {
		t.Error("expected default workflows")
	}

	// Check feature workflow exists
	feature, ok := workflows["feature"]
	if !ok {
		t.Error("expected feature workflow")
	}
	if feature.BranchTemplate == "" {
		t.Error("expected feature branch template")
	}
}

func TestGetWorkflow(t *testing.T) {
	cfg := DefaultConfig()

	// Get existing workflow
	wf := cfg.GetWorkflow("feature")
	if wf == nil {
		t.Error("expected feature workflow")
	}

	// Get non-existent workflow returns nil
	wf = cfg.GetWorkflow("nonexistent")
	if wf != nil {
		t.Error("expected nil for nonexistent workflow")
	}
}

func TestAddWorkflowHookFunc(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")

	// Start with empty file
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Add hook to workflow (path, workflowName, hookName, stage)
	if err := AddWorkflowHook(configPath, "feature", "my-hook", "pre_create"); err != nil {
		t.Fatalf("AddWorkflowHook failed: %v", err)
	}

	// Verify hook was added
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatal(err)
	}
	wf := loaded.GetWorkflow("feature")
	if wf == nil {
		t.Fatal("expected feature workflow")
	}
	if len(wf.Hooks.PreCreate) != 1 || wf.Hooks.PreCreate[0] != "my-hook" {
		t.Errorf("expected ['my-hook'], got %v", wf.Hooks.PreCreate)
	}
}

func TestMergeConfigWorkflows(t *testing.T) {
	base := DefaultConfig()
	override := &Config{
		Workflows: map[string]Workflow{
			"feature": {
				Description:    "Custom feature",
				BranchTemplate: "custom/{{.Slug}}",
			},
		},
	}

	merged := MergeConfig(base, override)

	// Override workflow should win
	wf := merged.GetWorkflow("feature")
	if wf == nil {
		t.Fatal("expected feature workflow")
	}
	if wf.BranchTemplate != "custom/{{.Slug}}" {
		t.Errorf("expected custom template, got %s", wf.BranchTemplate)
	}

	// Other workflows from base should still exist
	bugfix := merged.GetWorkflow("bugfix")
	if bugfix == nil {
		t.Error("expected bugfix workflow from base")
	}
}

func TestIsInitializedFunc(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Not initialized (no marker file)
	if IsInitialized() {
		t.Error("expected false when not initialized")
	}

	// Mark as initialized
	if err := MarkInitialized(); err != nil {
		t.Fatal(err)
	}

	if !IsInitialized() {
		t.Error("expected true after marking initialized")
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if !contains(slice, "a") {
		t.Error("expected true for 'a'")
	}
	if !contains(slice, "c") {
		t.Error("expected true for 'c'")
	}
	if contains(slice, "d") {
		t.Error("expected false for 'd'")
	}
	if contains(nil, "a") {
		t.Error("expected false for nil slice")
	}
}

func TestRemoveFromSlice(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected []string
	}{
		{"remove middle", []string{"a", "b", "c"}, "b", []string{"a", "c"}},
		{"remove first", []string{"a", "b", "c"}, "a", []string{"b", "c"}},
		{"remove last", []string{"a", "b", "c"}, "c", []string{"a", "b"}},
		{"remove nonexistent", []string{"a", "b"}, "x", []string{"a", "b"}},
		{"remove from empty", []string{}, "a", []string{}},
		{"remove all duplicates", []string{"a", "a", "a"}, "a", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeFromSlice(tt.slice, tt.item)
			if len(result) != len(tt.expected) {
				t.Errorf("len mismatch: got %d, want %d", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("result[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}
