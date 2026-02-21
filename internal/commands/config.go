package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/raisedadead/wt/internal/config"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/ui"
	"github.com/spf13/cobra"
)

var (
	configGlobal bool
	configLocal  bool
	configForce  bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage wt configuration",
	Long:  `View and manage wt configuration files.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a configuration file with documented defaults",
	Long: `Create a configuration file with all options commented out.

By default creates .wt.toml in the current project root (--local).
Use --global to create ~/.config/wt/config.toml instead.`,
	RunE: runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show effective configuration with sources",
	Long:  `Display the current effective configuration, showing where each value comes from.`,
	RunE:  runConfigShow,
}

func init() {
	configInitCmd.Flags().BoolVar(&configGlobal, "global", false, "Create global config (~/.config/wt/config.toml)")
	configInitCmd.Flags().BoolVar(&configLocal, "local", false, "Create repo config (.wt.toml) [default]")
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
			configPath = filepath.Join(cwd, ".wt.toml")
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

	// For global config, detect integrations and enable all by default
	var selections config.IntegrationSelections
	if configGlobal {
		selections = detectIntegrations()
	}

	// Generate config template with selected integrations
	template := config.GenerateConfigWithIntegrations(selections)
	if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	if IsJSONOutput() {
		data := map[string]string{"path": configPath}
		return ui.OutputJSON(os.Stdout, "config init", data, nil)
	}

	fmt.Println(ui.SuccessMsg(fmt.Sprintf("Created %s", configPath)))

	// Show what was enabled
	if configGlobal {
		printEnabledIntegrations(selections)
	}

	return nil
}

// detectIntegrations detects available tools and enables all detected by default
func detectIntegrations() config.IntegrationSelections {
	return config.IntegrationSelections{
		Zoxide: isCommandAvailable("zoxide"),
		GitHub: isCommandAvailable("gh"),
		Direnv: isCommandAvailable("direnv"),
	}
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// printEnabledIntegrations shows the user what integrations were configured
func printEnabledIntegrations(selections config.IntegrationSelections) {
	if selections.Zoxide {
		fmt.Println(ui.SuccessMsg("Enabled zoxide for quick worktree navigation"))
	}
	if selections.GitHub {
		fmt.Println(ui.SuccessMsg("Enabled GitHub CLI integration (gh-default, github-issue)"))
	}
	if selections.Direnv {
		fmt.Println(ui.SuccessMsg("Enabled direnv for auto-loading .envrc files"))
	}
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

	// Build settings table
	autoTrackValue := "false"
	if cfg.AutoTrack != nil && *cfg.AutoTrack {
		autoTrackValue = "true"
	}

	settingsRows := [][]string{
		{"worktree_root", formatConfigVal(cfg.WorktreeRoot), formatSource(sources["worktree_root"])},
		{"default_owner", formatConfigVal(cfg.DefaultOwner), formatSource(sources["default_owner"])},
		{"default_remote", formatConfigVal(cfg.DefaultRemote), formatSource(sources["default_remote"])},
		{"default_base_branch", formatConfigVal(cfg.DefaultBaseBranch), formatSource(sources["default_base_branch"])},
		{"branch_template", formatConfigVal(cfg.BranchTemplate), formatSource(sources["branch_template"])},
		{"git_timeout", fmt.Sprintf("%d", cfg.GitTimeout), formatSource(sources["git_timeout"])},
		{"git_long_timeout", fmt.Sprintf("%d", cfg.GitLongTimeout), formatSource(sources["git_long_timeout"])},
		{"hook_timeout", fmt.Sprintf("%d", cfg.HookTimeout), formatSource(sources["hook_timeout"])},
		{"auto_track", autoTrackValue, formatSource(sources["auto_track"])},
	}

	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	settingsTable := ui.NewStyledTable(func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return cellStyle.Bold(true)
		}
		if col == 2 {
			return cellStyle.Foreground(ui.Subtle)
		}
		return cellStyle
	}).Headers("KEY", "VALUE", "SOURCE").Rows(settingsRows...)

	fmt.Println(settingsTable.String())

	// Build hooks table
	hooksRows := [][]string{
		{"post_clone", formatSlice(cfg.Hooks.PostClone)},
		{"post_add", formatSlice(cfg.Hooks.PostAdd)},
	}

	hooksTable := ui.NewTable().Headers("EVENT", "COMMANDS").Rows(hooksRows...)
	fmt.Println(hooksTable.String())

	// Build workflows table
	if len(cfg.Workflows) > 0 {
		var wfRows [][]string
		names := make([]string, 0, len(cfg.Workflows))
		for name := range cfg.Workflows {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			wf := cfg.Workflows[name]
			hooks := formatSlice(append(wf.Hooks.PreCreate, wf.Hooks.PostAdd...))
			wfRows = append(wfRows, []string{name, wf.Description, wf.BranchTemplate, hooks})
		}

		wfTable := ui.NewTable().Headers("WORKFLOW", "DESCRIPTION", "TEMPLATE", "HOOKS").Rows(wfRows...)
		fmt.Println(wfTable.String())
	}

	return nil
}

func formatConfigVal(v string) string {
	if v == "" {
		return `""`
	}
	return v
}

func formatSource(s string) string {
	if s == "" || s == "default" {
		return "default"
	}
	return shortenConfigPath(s)
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

func formatSlice(s []string) string {
	if len(s) == 0 {
		return "[]"
	}
	quoted := make([]string, len(s))
	for i, v := range s {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return "[" + strings.Join(quoted, ", ") + "]"
}
