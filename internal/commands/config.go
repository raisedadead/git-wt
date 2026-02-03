package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/raisedadead/git-wt/internal/config"
	"github.com/raisedadead/git-wt/internal/git"
	"github.com/raisedadead/git-wt/internal/ui"
	"github.com/spf13/cobra"
)

var (
	configGlobal bool
	configLocal  bool
	configForce  bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage git-wt configuration",
	Long:  `View and manage git-wt configuration files.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a configuration file with documented defaults",
	Long: `Create a configuration file with all options commented out.

By default creates .git-wt.toml in the current project root (--local).
Use --global to create ~/.config/git-wt/config.toml instead.`,
	RunE: runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show effective configuration with sources",
	Long:  `Display the current effective configuration, showing where each value comes from.`,
	RunE:  runConfigShow,
}

func init() {
	configInitCmd.Flags().BoolVar(&configGlobal, "global", false, "Create global config (~/.config/git-wt/config.toml)")
	configInitCmd.Flags().BoolVar(&configLocal, "local", false, "Create repo config (.git-wt.toml) [default]")
	configInitCmd.Flags().BoolVar(&configForce, "force", false, "Overwrite existing config file")

	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	var configPath string

	if configGlobal {
		configPath = config.GetConfigPath()

		// Install hooks directories and bundled hooks
		if err := config.InstallBundledHooks(); err != nil {
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "config init", nil,
					ui.NewCLIError(ui.ErrCodeGit, fmt.Sprintf("failed to install hooks: %v", err)))
			}
			return fmt.Errorf("failed to install hooks: %w", err)
		}
	} else {
		// Default to local (repo) config
		projectRoot, err := git.GetProjectRoot(".")
		if err != nil {
			// Not in a project, use current directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			configPath = filepath.Join(cwd, ".git-wt.toml")
		} else {
			configPath = config.GetRepoConfigPath(projectRoot)
		}
	}

	// Check if file exists
	if _, err := os.Stat(configPath); err == nil && !configForce {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "config init", nil,
				ui.NewCLIError(ui.ErrCodeAlreadyExists, fmt.Sprintf("config file already exists: %s (use --force to overwrite)", configPath)))
		}
		return fmt.Errorf("config file already exists: %s (use --force to overwrite)", configPath)
	}

	// Create parent directory if needed
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write config template
	template := config.GenerateConfigTemplate()
	if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if IsJSONOutput() {
		data := map[string]string{"path": configPath}
		return ui.OutputJSON(os.Stdout, "config init", data, nil)
	}

	fmt.Println(ui.SuccessMsg(fmt.Sprintf("Created %s", configPath)))

	// For global config, offer to enable zoxide if detected
	if configGlobal {
		if _, err := exec.LookPath("zoxide"); err == nil {
			var enable bool
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title("zoxide detected. Enable auto-navigation for worktrees?").
						Description("Adds worktrees to zoxide for quick 'z' navigation").
						Value(&enable),
				),
			)

			if err := form.Run(); err == nil && enable {
				// Enable zoxide hook for its declared events (post_clone, post_add)
				for _, event := range []string{"post_clone", "post_add"} {
					if err := config.AddHookToConfig(configPath, "zoxide", event); err != nil {
						fmt.Println(ui.WarningMsg(fmt.Sprintf("Failed to enable zoxide hook: %v", err)))
						break
					}
				}
				fmt.Println(ui.SuccessMsg("Enabled zoxide hook for: post_clone, post_add"))
			}
		}
	}

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	// Try to find project root for repo config
	projectRoot, _ := git.GetProjectRoot(".")

	cfg, sources, err := config.LoadEffective(config.GetConfigPath(), projectRoot)
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "config show", nil, ui.NewCLIError(ui.ErrCodeGit, err.Error()))
		}
		return err
	}

	if IsJSONOutput() {
		data := map[string]interface{}{
			"config":  cfg,
			"sources": sources,
		}
		return ui.OutputJSON(os.Stdout, "config show", data, nil)
	}

	// Pretty print with sources
	printConfigValue("worktree_root", cfg.WorktreeRoot, sources["worktree_root"])
	printConfigValue("default_remote", cfg.DefaultRemote, sources["default_remote"])
	printConfigValue("default_base_branch", cfg.DefaultBaseBranch, sources["default_base_branch"])
	printConfigValue("branch_template", cfg.BranchTemplate, sources["branch_template"])
	printConfigValue("git_timeout", fmt.Sprintf("%d", cfg.GitTimeout), sources["git_timeout"])
	printConfigValue("git_long_timeout", fmt.Sprintf("%d", cfg.GitLongTimeout), sources["git_long_timeout"])
	printConfigValue("hook_timeout", fmt.Sprintf("%d", cfg.HookTimeout), sources["hook_timeout"])

	return nil
}

func printConfigValue(key, value, source string) {
	if value == "" {
		value = `""`
	} else if key != "git_timeout" && key != "git_long_timeout" && key != "hook_timeout" {
		value = fmt.Sprintf("%q", value)
	}

	var sourceDisplay string
	if source == "default" {
		sourceDisplay = ui.SubtleStyle.Render("default")
	} else {
		sourceDisplay = ui.SubtleStyle.Render(shortenConfigPath(source))
	}

	fmt.Printf("%s = %-40s # %s\n", key, value, sourceDisplay)
}

func shortenConfigPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
