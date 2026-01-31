package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the git-wt configuration
type Config struct {
	WorktreeRoot      string `toml:"worktree_root"`
	DefaultRemote     string `toml:"default_remote"`
	DefaultBaseBranch string `toml:"default_base_branch"`
	BranchTemplate    string `toml:"branch_template"`
	GitTimeout        int    `toml:"git_timeout"`
	GitLongTimeout    int    `toml:"git_long_timeout"`
	HookTimeout       int    `toml:"hook_timeout"`
	Hooks             Hooks  `toml:"hooks"`
}

// Hooks defines user-configurable hook commands
type Hooks struct {
	PostClone []string `toml:"post_clone"`
	PostAdd   []string `toml:"post_add"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		WorktreeRoot:      "",
		DefaultRemote:     "origin",
		DefaultBaseBranch: "",
		BranchTemplate:    "{{type}}-{{number}}-{{slug}}",
		GitTimeout:        120,
		GitLongTimeout:    600,
		HookTimeout:       30,
		Hooks:             Hooks{},
	}
}

// GetConfigDir returns the config directory path following XDG spec
func GetConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "git-wt")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home cannot be determined
		return ".git-wt"
	}
	return filepath.Join(home, ".config", "git-wt")
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() string {
	return filepath.Join(GetConfigDir(), "config.toml")
}

// GetInitMarkerPath returns the path to the initialization marker
func GetInitMarkerPath() string {
	return filepath.Join(GetConfigDir(), ".initialized")
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
	return filepath.Join(projectRoot, ".git-wt.toml")
}

// MergeConfig merges override config into base, returning a new config
// Non-zero values in override take precedence
func MergeConfig(base, override *Config) *Config {
	merged := *base // Copy base

	if override.WorktreeRoot != "" {
		merged.WorktreeRoot = override.WorktreeRoot
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
	if len(override.Hooks.PostClone) > 0 {
		merged.Hooks.PostClone = override.Hooks.PostClone
	}
	if len(override.Hooks.PostAdd) > 0 {
		merged.Hooks.PostAdd = override.Hooks.PostAdd
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
	for _, field := range []string{"worktree_root", "default_remote", "default_base_branch",
		"branch_template", "git_timeout", "git_long_timeout", "hook_timeout"} {
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

// GenerateConfigTemplate returns a config file template with all options commented
func GenerateConfigTemplate() string {
	return `# ============================================================
# git-wt configuration
# Generated by: git wt config init
# Uncomment and modify options as needed
# ============================================================

# --- Directory Settings ---

# Where to clone repos (empty = current directory)
# Applies to: clone
# Flag: --root
# worktree_root = ""

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

# --- Timeout Settings (seconds) ---

# Standard git operations (status, branch, etc.)
# Flag: --timeout
# git_timeout = 120

# Long git operations (clone, fetch)
# git_long_timeout = 600

# Hook execution timeout
# Flag: --hook-timeout
# hook_timeout = 30

# --- Hooks ---
# Shell commands to run after operations
# Environment variables: GIT_WT_PATH, GIT_WT_BRANCH, GIT_WT_PROJECT_ROOT, GIT_WT_DEFAULT_BRANCH
# Template variables: {{.Path}}, {{.Branch}}, {{.ProjectRoot}}, {{.DefaultBranch}}

# [hooks]
# post_clone = []
# post_add = []
`
}
