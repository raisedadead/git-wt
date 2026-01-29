package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the git-wt configuration
type Config struct {
	WorktreeRoot string `toml:"worktree_root"`
	Hooks        Hooks  `toml:"hooks"`
}

// Hooks defines user-configurable hook commands
type Hooks struct {
	PostClone []string `toml:"post_clone"`
	PostAdd   []string `toml:"post_add"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		WorktreeRoot: filepath.Join(home, "DEV", "worktrees"),
		Hooks:        Hooks{},
	}
}

// GetConfigDir returns the config directory path following XDG spec
func GetConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "git-wt")
	}
	home, _ := os.UserHomeDir()
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
