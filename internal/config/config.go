package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/raisedadead/wt/internal/hooks/bundled"
)

// Config holds the wt configuration
type Config struct {
	WorktreeRoot      string              `toml:"worktree_root"`
	DefaultOwner      string              `toml:"default_owner"`
	DefaultRemote     string              `toml:"default_remote"`
	DefaultBaseBranch string              `toml:"default_base_branch"`
	BranchTemplate    string              `toml:"branch_template"`
	GitTimeout        int                 `toml:"git_timeout"`
	GitLongTimeout    int                 `toml:"git_long_timeout"`
	HookTimeout       int                 `toml:"hook_timeout"`
	AutoTrack         *bool               `toml:"auto_track"`
	Hooks             Hooks               `toml:"hooks"`
	Workflows         map[string]Workflow `toml:"workflows"`
}

// Hooks defines user-configurable hook commands
type Hooks struct {
	PostClone []string `toml:"post_clone"`
	PostAdd   []string `toml:"post_add"`
}

// Workflow defines a workflow configuration
type Workflow struct {
	Description    string        `toml:"description"`
	BranchTemplate string        `toml:"branch_template"`
	Hooks          WorkflowHooks `toml:"hooks"`
}

// WorkflowHooks defines hooks for different workflow stages
type WorkflowHooks struct {
	PreCreate []string `toml:"pre_create"`
	PostAdd   []string `toml:"post_add"`
}

// ptrBool returns a pointer to a bool value
func ptrBool(b bool) *bool {
	return &b
}

// IntegrationSelections holds the user's selected integrations for config generation
type IntegrationSelections struct {
	Zoxide bool
	GitHub bool
	Direnv bool
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		WorktreeRoot:      "",
		DefaultOwner:      "",
		DefaultRemote:     "origin",
		DefaultBaseBranch: "",
		BranchTemplate:    "{{type}}-{{number}}-{{slug}}",
		GitTimeout:        120,
		GitLongTimeout:    600,
		HookTimeout:       30,
		AutoTrack:         ptrBool(false),
		Hooks:             Hooks{},
		Workflows:         DefaultWorkflows(),
	}
}

// DefaultWorkflows returns the default workflow configurations
func DefaultWorkflows() map[string]Workflow {
	return map[string]Workflow{
		"feature": {
			Description:    "New feature development",
			BranchTemplate: "feat/{slug}",
			Hooks: WorkflowHooks{
				PreCreate: []string{},
				PostAdd:   []string{"direnv", "zoxide"},
			},
		},
		"bugfix": {
			Description:    "Bug fix",
			BranchTemplate: "fix/{slug}",
			Hooks: WorkflowHooks{
				PreCreate: []string{},
				PostAdd:   []string{"direnv", "zoxide"},
			},
		},
		"pr-review": {
			Description:    "Review a pull request",
			BranchTemplate: "{branch}",
			Hooks: WorkflowHooks{
				PreCreate: []string{"github-pr"},
				PostAdd:   []string{"direnv", "zoxide"},
			},
		},
		"branch": {
			Description:    "Plain branch",
			BranchTemplate: "{name}",
			Hooks: WorkflowHooks{
				PreCreate: []string{},
				PostAdd:   []string{"direnv", "zoxide"},
			},
		},
	}
}

// GetWorkflow returns the workflow configuration for the given name
func (c *Config) GetWorkflow(name string) *Workflow {
	if c.Workflows == nil {
		return nil
	}
	if w, ok := c.Workflows[name]; ok {
		return &w
	}
	return nil
}

// GetConfigDir returns the config directory path following XDG spec
func GetConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "wt")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home cannot be determined
		return ".wt"
	}
	return filepath.Join(home, ".config", "wt")
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.toml")
}

// GetInitMarkerPath returns the path to the initialization marker
func GetInitMarkerPath() string {
	return filepath.Join(GetConfigDir(), ".initialized")
}

// GetHooksDir returns the hooks directory path
func GetHooksDir() string {
	return filepath.Join(GetConfigDir(), "hooks")
}

// GetCommunityHooksDir returns the community hooks directory path
func GetCommunityHooksDir() string {
	return filepath.Join(GetHooksDir(), "community")
}

// GetCustomHooksDir returns the custom hooks directory path
func GetCustomHooksDir() string {
	return filepath.Join(GetHooksDir(), "custom")
}

// InstallBundledHooks copies bundled hooks to community directory
func InstallBundledHooks() error {
	communityDir := GetCommunityHooksDir()
	customDir := GetCustomHooksDir()

	// Create directories
	if err := os.MkdirAll(communityDir, 0755); err != nil {
		return fmt.Errorf("failed to create community hooks dir: %w", err)
	}
	if err := os.MkdirAll(customDir, 0755); err != nil {
		return fmt.Errorf("failed to create custom hooks dir: %w", err)
	}

	// Copy bundled hooks
	for _, name := range bundled.List() {
		content, err := bundled.Scripts.ReadFile(name + ".sh")
		if err != nil {
			return fmt.Errorf("failed to read bundled hook %s: %w", name, err)
		}

		destPath := filepath.Join(communityDir, name+".sh")
		if err := os.WriteFile(destPath, content, 0755); err != nil {
			return fmt.Errorf("failed to write hook %s: %w", name, err)
		}
	}

	return nil
}

// Load loads configuration from the given path
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Restore defaults for zero timeout values (prevents instant timeouts when
	// users have git_long_timeout = 0 in their config, which was a common issue
	// from copying the config template without editing)
	defaults := DefaultConfig()
	if cfg.GitTimeout == 0 {
		cfg.GitTimeout = defaults.GitTimeout
	}
	if cfg.GitLongTimeout == 0 {
		cfg.GitLongTimeout = defaults.GitLongTimeout
	}
	if cfg.HookTimeout == 0 {
		cfg.HookTimeout = defaults.HookTimeout
	}

	return cfg, nil
}

// loadRaw loads configuration from the given path without applying defaults
// Returns an empty config if file doesn't exist (for merging purposes)
func loadRaw(path string) (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadGlobal loads the global configuration
func LoadGlobal() (*Config, error) {
	return Load(GetConfigPath())
}

// IsInitialized checks if the first-run hint has been shown
func IsInitialized() bool {
	_, err := os.Stat(GetInitMarkerPath())
	return err == nil
}

// MarkInitialized creates the initialization marker
func MarkInitialized() error {
	dir := GetConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(GetInitMarkerPath(), []byte{}, 0644)
}

// GetRepoConfigPath returns the path to repo-level config
func GetRepoConfigPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".wt.toml")
}

// MergeConfig merges override config into base, returning a new config
// Non-zero values in override take precedence
func MergeConfig(base, override *Config) *Config {
	merged := *base // Copy base

	if override.WorktreeRoot != "" {
		merged.WorktreeRoot = override.WorktreeRoot
	}
	if override.DefaultOwner != "" {
		merged.DefaultOwner = override.DefaultOwner
	}
	if override.DefaultRemote != "" {
		merged.DefaultRemote = override.DefaultRemote
	}
	if override.DefaultBaseBranch != "" {
		merged.DefaultBaseBranch = override.DefaultBaseBranch
	}
	if override.BranchTemplate != "" {
		merged.BranchTemplate = override.BranchTemplate
	}
	if override.GitTimeout != 0 {
		merged.GitTimeout = override.GitTimeout
	}
	if override.GitLongTimeout != 0 {
		merged.GitLongTimeout = override.GitLongTimeout
	}
	if override.HookTimeout != 0 {
		merged.HookTimeout = override.HookTimeout
	}
	if override.AutoTrack != nil {
		merged.AutoTrack = override.AutoTrack
	}
	if len(override.Hooks.PostClone) > 0 {
		merged.Hooks.PostClone = override.Hooks.PostClone
	}
	if len(override.Hooks.PostAdd) > 0 {
		merged.Hooks.PostAdd = override.Hooks.PostAdd
	}

	// Merge workflows - override replaces base for each workflow key
	if len(override.Workflows) > 0 {
		if merged.Workflows == nil {
			merged.Workflows = make(map[string]Workflow)
		}
		for k, v := range override.Workflows {
			merged.Workflows[k] = v
		}
	}

	return &merged
}

// LoadWithRepo loads config with hierarchy: repo > global > defaults
func LoadWithRepo(globalPath, projectRoot string) (*Config, error) {
	// Start with defaults
	cfg := DefaultConfig()

	// Load global config (raw, without defaults, for proper merging)
	globalCfg, err := loadRaw(globalPath)
	if err != nil {
		return nil, err
	}
	cfg = MergeConfig(cfg, globalCfg)

	// Load repo config if exists (raw, without defaults, for proper merging)
	if projectRoot != "" {
		repoPath := GetRepoConfigPath(projectRoot)
		repoCfg, err := loadRaw(repoPath)
		if err != nil {
			return nil, err
		}
		cfg = MergeConfig(cfg, repoCfg)
	}

	return cfg, nil
}

// LoadEffective loads config and tracks source of each value
// Returns config, map of field->source path, and error
func LoadEffective(globalPath, projectRoot string) (*Config, map[string]string, error) {
	sources := make(map[string]string)
	cfg := DefaultConfig()

	// Mark all as default initially
	for _, field := range []string{"worktree_root", "default_owner", "default_remote", "default_base_branch",
		"branch_template", "git_timeout", "git_long_timeout", "hook_timeout", "auto_track"} {
		sources[field] = "default"
	}

	// Load and track global config
	if data, err := os.ReadFile(globalPath); err == nil {
		var globalCfg Config
		if err := toml.Unmarshal(data, &globalCfg); err != nil {
			return nil, nil, fmt.Errorf("invalid config %s: %w", globalPath, err)
		}
		if globalCfg.WorktreeRoot != "" {
			cfg.WorktreeRoot = globalCfg.WorktreeRoot
			sources["worktree_root"] = globalPath
		}
		if globalCfg.DefaultOwner != "" {
			cfg.DefaultOwner = globalCfg.DefaultOwner
			sources["default_owner"] = globalPath
		}
		if globalCfg.DefaultRemote != "" {
			cfg.DefaultRemote = globalCfg.DefaultRemote
			sources["default_remote"] = globalPath
		}
		if globalCfg.DefaultBaseBranch != "" {
			cfg.DefaultBaseBranch = globalCfg.DefaultBaseBranch
			sources["default_base_branch"] = globalPath
		}
		if globalCfg.BranchTemplate != "" {
			cfg.BranchTemplate = globalCfg.BranchTemplate
			sources["branch_template"] = globalPath
		}
		if globalCfg.GitTimeout != 0 {
			cfg.GitTimeout = globalCfg.GitTimeout
			sources["git_timeout"] = globalPath
		}
		if globalCfg.GitLongTimeout != 0 {
			cfg.GitLongTimeout = globalCfg.GitLongTimeout
			sources["git_long_timeout"] = globalPath
		}
		if globalCfg.HookTimeout != 0 {
			cfg.HookTimeout = globalCfg.HookTimeout
			sources["hook_timeout"] = globalPath
		}
		if globalCfg.AutoTrack != nil {
			cfg.AutoTrack = globalCfg.AutoTrack
			sources["auto_track"] = globalPath
		}
		if len(globalCfg.Hooks.PostClone) > 0 {
			cfg.Hooks.PostClone = globalCfg.Hooks.PostClone
		}
		if len(globalCfg.Hooks.PostAdd) > 0 {
			cfg.Hooks.PostAdd = globalCfg.Hooks.PostAdd
		}
	}

	// Load and track repo config
	if projectRoot != "" {
		repoPath := GetRepoConfigPath(projectRoot)
		if data, err := os.ReadFile(repoPath); err == nil {
			var repoCfg Config
			if err := toml.Unmarshal(data, &repoCfg); err != nil {
				return nil, nil, fmt.Errorf("invalid config %s: %w", repoPath, err)
			}
			if repoCfg.WorktreeRoot != "" {
				cfg.WorktreeRoot = repoCfg.WorktreeRoot
				sources["worktree_root"] = repoPath
			}
			if repoCfg.DefaultOwner != "" {
				cfg.DefaultOwner = repoCfg.DefaultOwner
				sources["default_owner"] = repoPath
			}
			if repoCfg.DefaultRemote != "" {
				cfg.DefaultRemote = repoCfg.DefaultRemote
				sources["default_remote"] = repoPath
			}
			if repoCfg.DefaultBaseBranch != "" {
				cfg.DefaultBaseBranch = repoCfg.DefaultBaseBranch
				sources["default_base_branch"] = repoPath
			}
			if repoCfg.BranchTemplate != "" {
				cfg.BranchTemplate = repoCfg.BranchTemplate
				sources["branch_template"] = repoPath
			}
			if repoCfg.GitTimeout != 0 {
				cfg.GitTimeout = repoCfg.GitTimeout
				sources["git_timeout"] = repoPath
			}
			if repoCfg.GitLongTimeout != 0 {
				cfg.GitLongTimeout = repoCfg.GitLongTimeout
				sources["git_long_timeout"] = repoPath
			}
			if repoCfg.HookTimeout != 0 {
				cfg.HookTimeout = repoCfg.HookTimeout
				sources["hook_timeout"] = repoPath
			}
			if repoCfg.AutoTrack != nil {
				cfg.AutoTrack = repoCfg.AutoTrack
				sources["auto_track"] = repoPath
			}
			if len(repoCfg.Hooks.PostClone) > 0 {
				cfg.Hooks.PostClone = repoCfg.Hooks.PostClone
			}
			if len(repoCfg.Hooks.PostAdd) > 0 {
				cfg.Hooks.PostAdd = repoCfg.Hooks.PostAdd
			}
		}
	}

	return cfg, sources, nil
}

// SaveConfig writes config to the specified path
func SaveConfig(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}

	encoder := toml.NewEncoder(f)
	encodeErr := encoder.Encode(cfg)
	closeErr := f.Close()
	if encodeErr != nil {
		return encodeErr
	}
	return closeErr
}

// AddHookToConfig adds a hook to the specified event
func AddHookToConfig(path, hookName, event string) error {
	cfg, err := loadRaw(path)
	if err != nil {
		// Only create empty config if file doesn't exist
		// For other errors (permission, parse), propagate the error
		if os.IsNotExist(err) {
			cfg = &Config{}
		} else {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	switch event {
	case "post_clone":
		if !contains(cfg.Hooks.PostClone, hookName) {
			cfg.Hooks.PostClone = append(cfg.Hooks.PostClone, hookName)
		}
	case "post_add":
		if !contains(cfg.Hooks.PostAdd, hookName) {
			cfg.Hooks.PostAdd = append(cfg.Hooks.PostAdd, hookName)
		}
	default:
		return fmt.Errorf("unknown event: %s", event)
	}

	return SaveConfig(cfg, path)
}

// AddWorkflowHook adds a hook to a workflow's pre_create or post_add hooks
func AddWorkflowHook(path, workflowName, hookName, stage string) error {
	cfg, err := loadRaw(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = &Config{}
		} else {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	// Initialize workflows map if needed
	if cfg.Workflows == nil {
		cfg.Workflows = make(map[string]Workflow)
	}

	// Get existing workflow or create from defaults
	workflow, exists := cfg.Workflows[workflowName]
	if !exists {
		defaults := DefaultWorkflows()
		if defaultWf, ok := defaults[workflowName]; ok {
			workflow = defaultWf
		} else {
			workflow = Workflow{
				Description:    workflowName,
				BranchTemplate: "{name}",
			}
		}
	}

	// Add hook to appropriate stage
	switch stage {
	case "pre_create":
		if !contains(workflow.Hooks.PreCreate, hookName) {
			workflow.Hooks.PreCreate = append(workflow.Hooks.PreCreate, hookName)
		}
	case "post_add":
		if !contains(workflow.Hooks.PostAdd, hookName) {
			workflow.Hooks.PostAdd = append(workflow.Hooks.PostAdd, hookName)
		}
	default:
		return fmt.Errorf("unknown stage: %s", stage)
	}

	cfg.Workflows[workflowName] = workflow
	return SaveConfig(cfg, path)
}

// RemoveHookFromConfig removes a hook from all events
func RemoveHookFromConfig(path, hookName string) error {
	cfg, err := loadRaw(path)
	if err != nil {
		return err
	}

	cfg.Hooks.PostClone = removeFromSlice(cfg.Hooks.PostClone, hookName)
	cfg.Hooks.PostAdd = removeFromSlice(cfg.Hooks.PostAdd, hookName)

	return SaveConfig(cfg, path)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeFromSlice(slice []string, item string) []string {
	var result []string
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}

// GenerateConfigTemplate returns a config file template with all options commented
func GenerateConfigTemplate() string {
	return GenerateConfigWithIntegrations(IntegrationSelections{})
}

// GenerateConfigWithIntegrations returns a config template with selected integrations pre-configured
func GenerateConfigWithIntegrations(selections IntegrationSelections) string {
	// Build the base template (settings section)
	template := `# ============================================================
# wt configuration
# Generated by: wt config init
# Uncomment and modify options as needed
# ============================================================

# --- Directory Settings ---

# Where to clone repos (empty = current directory)
# Applies to: clone
# Flag: --root
# worktree_root = ""

# --- Owner/Organization Settings ---

# Default owner/org for shorthand clone (e.g., "wt clone repo" becomes "owner/repo")
# Applies to: clone
# default_owner = ""

# --- Remote Settings ---

# Git remote name for operations
# Applies to: prune, new
# Flag: --remote
# default_remote = "origin"

# --- Branch Settings ---

# Base branch for new worktrees (empty = HEAD)
# Applies to: new
# Flag: --base
# default_base_branch = ""

# Branch name template for GitHub issues/PRs
# Variables: {{type}}, {{number}}, {{slug}}
# Applies to: new --issue, new --pr
# Flag: --branch-template
# branch_template = "{{type}}-{{number}}-{{slug}}"

# Auto-track remote branches without prompting (like git checkout)
# Applies to: new
# auto_track = false

# --- Timeout Settings (seconds) ---

# Standard git operations (status, branch, etc.)
# Flag: --timeout
# git_timeout = 120

# Long git operations (clone, fetch)
# git_long_timeout = 600

# Hook execution timeout
# Flag: --hook-timeout
# hook_timeout = 30

`

	// Build hooks section based on selections
	template += generateHooksSection(selections)

	// Build workflows section if GitHub is enabled
	if selections.GitHub {
		template += generateWorkflowsSection(selections)
	}

	return template
}

// generateHooksSection creates the hooks section of the config template
func generateHooksSection(selections IntegrationSelections) string {
	hasAnyHook := selections.Zoxide || selections.GitHub || selections.Direnv

	if !hasAnyHook {
		// No integrations selected - show commented template
		return `# --- Hooks ---
# Shell commands or bundled hook names to run after operations
# Environment variables: WT_PATH, WT_BRANCH, WT_PROJECT_ROOT, WT_DEFAULT_BRANCH
# Available hooks: zoxide, gh-default, direnv, github-issue, github-pr
# Run 'wt hooks list' to see all available hooks

# [hooks]
# post_clone = []
# post_add = []
`
	}

	// Build hook arrays
	var postClone, postAdd []string

	if selections.Zoxide {
		postClone = append(postClone, "zoxide")
		postAdd = append(postAdd, "zoxide")
	}
	if selections.GitHub {
		postClone = append(postClone, "gh-default")
	}
	if selections.Direnv {
		postAdd = append(postAdd, "direnv")
	}

	section := `# --- Hooks ---
# Shell commands or bundled hook names to run after operations
# Environment variables: WT_PATH, WT_BRANCH, WT_PROJECT_ROOT, WT_DEFAULT_BRANCH
# Available hooks: zoxide, gh-default, direnv, github-issue, github-pr
# Run 'wt hooks list' to see all available hooks

[hooks]
`

	// Format post_clone
	section += "# Runs after cloning a repository (wt clone)\n"
	section += fmt.Sprintf("post_clone = %s\n\n", formatStringSlice(postClone))

	// Format post_add
	section += "# Runs after creating a new worktree (wt add/new)\n"
	section += fmt.Sprintf("post_add = %s\n", formatStringSlice(postAdd))

	return section
}

// generateWorkflowsSection creates the workflows section for GitHub integration
func generateWorkflowsSection(selections IntegrationSelections) string {
	// Build post_add hooks for workflows
	var workflowPostAdd []string
	if selections.Direnv {
		workflowPostAdd = append(workflowPostAdd, "direnv")
	}
	if selections.Zoxide {
		workflowPostAdd = append(workflowPostAdd, "zoxide")
	}

	section := `
# --- Workflows ---
# Predefined branch naming and hook configurations
# Use with: wt new --workflow=feature "my feature"
# Or use shortcuts: wt new --issue 123, wt new --pr 456

[workflows.feature]
# For new feature development
description = "New feature development"
branch_template = "feat/{slug}"
[workflows.feature.hooks]
# Prompts for GitHub issue to link (optional)
pre_create = ["github-issue"]
`
	section += fmt.Sprintf("post_add = %s\n", formatStringSlice(workflowPostAdd))

	section += `
[workflows.bugfix]
# For bug fixes
description = "Bug fix"
branch_template = "fix/{slug}"
[workflows.bugfix.hooks]
# Prompts for GitHub issue to link (optional)
pre_create = ["github-issue"]
`
	section += fmt.Sprintf("post_add = %s\n", formatStringSlice(workflowPostAdd))

	return section
}

// formatStringSlice formats a string slice as a TOML array
func formatStringSlice(s []string) string {
	if len(s) == 0 {
		return "[]"
	}
	quoted := make([]string, len(s))
	for i, v := range s {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return "[" + joinStrings(quoted, ", ") + "]"
}

// joinStrings joins strings with a separator (avoiding strings import for this simple case)
func joinStrings(s []string, sep string) string {
	if len(s) == 0 {
		return ""
	}
	result := s[0]
	for i := 1; i < len(s); i++ {
		result += sep + s[i]
	}
	return result
}
