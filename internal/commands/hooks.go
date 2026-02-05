package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/raisedadead/wt/internal/config"
	"github.com/raisedadead/wt/internal/git"
	"github.com/raisedadead/wt/internal/hooks"
	"github.com/raisedadead/wt/internal/ui"
	"github.com/spf13/cobra"
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage wt hooks",
	Long:  `List, enable, disable, and view wt hooks.`,
}

var hooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available and enabled hooks",
	RunE:  runHooksList,
}

var hooksEnableCmd = &cobra.Command{
	Use:   "enable [hook-name]",
	Short: "Enable a hook",
	Long: `Enable a hook for automatic execution during wt operations.

When run without arguments, shows an interactive picker to select from available hooks.
Each hook declares which events it runs on (post_clone, post_add).`,
	Args: cobra.MaximumNArgs(1),
	RunE: runHooksEnable,
}

var hooksDisableCmd = &cobra.Command{
	Use:   "disable <hook-name>",
	Short: "Disable a hook",
	Args:  cobra.ExactArgs(1),
	RunE:  runHooksDisable,
}

var hooksShowCmd = &cobra.Command{
	Use:   "show <hook-name>",
	Short: "Show hook details and content",
	Args:  cobra.ExactArgs(1),
	RunE:  runHooksShow,
}

func init() {
	hooksEnableCmd.Flags().String("event", "", "Override event (post_clone, post_add)")

	hooksCmd.AddCommand(hooksListCmd)
	hooksCmd.AddCommand(hooksEnableCmd)
	hooksCmd.AddCommand(hooksDisableCmd)
	hooksCmd.AddCommand(hooksShowCmd)
	rootCmd.AddCommand(hooksCmd)
}

// HookInfo represents hook information for display
type HookInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Events      []string `json:"events"`
	Requires    string   `json:"requires,omitempty"`
	Source      string   `json:"source"`
	Enabled     []string `json:"enabled,omitempty"`
}

func runHooksList(cmd *cobra.Command, args []string) error {
	communityDir := config.GetCommunityHooksDir()
	customDir := config.GetCustomHooksDir()

	projectRoot := ""
	if pr, err := git.GetProjectRoot("."); err == nil {
		projectRoot = pr
	}
	cfg, _, err := config.LoadEffective(config.GetConfigPath(), projectRoot)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	enabledMap := make(map[string][]string)
	for _, name := range cfg.Hooks.PostClone {
		enabledMap[name] = append(enabledMap[name], "post_clone")
	}
	for _, name := range cfg.Hooks.PostAdd {
		enabledMap[name] = append(enabledMap[name], "post_add")
	}

	var hookInfos []HookInfo

	if entries, err := os.ReadDir(communityDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sh") {
				name := strings.TrimSuffix(entry.Name(), ".sh")
				info := getHookInfo(filepath.Join(communityDir, entry.Name()), name, "community", enabledMap)
				hookInfos = append(hookInfos, info)
			}
		}
	}

	if entries, err := os.ReadDir(customDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sh") {
				name := strings.TrimSuffix(entry.Name(), ".sh")
				info := getHookInfo(filepath.Join(customDir, entry.Name()), name, "custom", enabledMap)
				hookInfos = append(hookInfos, info)
			}
		}
	}

	if IsJSONOutput() {
		return ui.OutputJSON(os.Stdout, "hooks list", map[string]interface{}{
			"hooks": hookInfos,
		}, nil)
	}

	if len(hookInfos) == 0 {
		fmt.Println("No hooks found. Run 'git wt config init --global' to install bundled hooks.")
		return nil
	}

	fmt.Println(ui.BoldStyle.Render("Available hooks:"))
	fmt.Println()

	for _, info := range hookInfos {
		status := ui.SubtleStyle.Render("[disabled]")
		if len(info.Enabled) > 0 {
			status = ui.SuccessStyle.Render("[enabled: " + strings.Join(info.Enabled, ", ") + "]")
		}

		fmt.Printf("  %-16s %-40s %s\n", info.Name, info.Description, status)
		if info.Source == "custom" {
			fmt.Printf("                   %s\n", ui.SubtleStyle.Render("(custom)"))
		}
	}

	return nil
}

func getHookInfo(path, name, source string, enabledMap map[string][]string) HookInfo {
	info := HookInfo{
		Name:    name,
		Source:  source,
		Enabled: enabledMap[name],
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return info
	}

	meta, err := hooks.ParseMetadata(string(content))
	if err != nil {
		return info
	}

	info.Description = meta.Description
	info.Events = meta.Events
	info.Requires = meta.Requires

	return info
}

// getAvailableHookInfos returns all available hooks from community and custom directories
// with their enabled status from the current config
func getAvailableHookInfos(communityDir, customDir string) []HookInfo {
	projectRoot := ""
	if pr, err := git.GetProjectRoot("."); err == nil {
		projectRoot = pr
	}
	cfg, _, err := config.LoadEffective(config.GetConfigPath(), projectRoot)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	enabledMap := make(map[string][]string)
	for _, name := range cfg.Hooks.PostClone {
		enabledMap[name] = append(enabledMap[name], "post_clone")
	}
	for _, name := range cfg.Hooks.PostAdd {
		enabledMap[name] = append(enabledMap[name], "post_add")
	}

	var hookInfos []HookInfo

	if entries, err := os.ReadDir(communityDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sh") {
				name := strings.TrimSuffix(entry.Name(), ".sh")
				info := getHookInfo(filepath.Join(communityDir, entry.Name()), name, "community", enabledMap)
				hookInfos = append(hookInfos, info)
			}
		}
	}

	if entries, err := os.ReadDir(customDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sh") {
				name := strings.TrimSuffix(entry.Name(), ".sh")
				info := getHookInfo(filepath.Join(customDir, entry.Name()), name, "custom", enabledMap)
				hookInfos = append(hookInfos, info)
			}
		}
	}

	return hookInfos
}

func runHooksEnable(cmd *cobra.Command, args []string) error {
	communityDir := config.GetCommunityHooksDir()
	customDir := config.GetCustomHooksDir()

	var hookName string

	if len(args) == 0 {
		// Interactive mode: show picker with available hooks
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "hooks enable", nil,
				ui.NewCLIError(ui.ErrCodeValidation, "hook name is required"))
		}

		// Get available hooks (reuse logic from runHooksList)
		availableHooks := getAvailableHookInfos(communityDir, customDir)
		if len(availableHooks) == 0 {
			return fmt.Errorf("no hooks found. Run 'git wt config init --global' to install bundled hooks")
		}

		// Filter to only disabled hooks (hooks that aren't already enabled)
		var disabledHooks []HookInfo
		for _, h := range availableHooks {
			if len(h.Enabled) == 0 {
				disabledHooks = append(disabledHooks, h)
			}
		}

		if len(disabledHooks) == 0 {
			fmt.Println(ui.SuccessMsg("All available hooks are already enabled"))
			return nil
		}

		// Build picker options
		var options []huh.Option[string]
		for _, h := range disabledHooks {
			label := h.Name
			if h.Description != "" {
				label = fmt.Sprintf("%s - %s", h.Name, h.Description)
			}
			options = append(options, huh.NewOption(label, h.Name))
		}

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select hook to enable").
					Description("Enabling adds the hook to your global config for automatic execution").
					Options(options...).
					Value(&hookName),
			),
		).WithKeyMap(DefaultFormKeyMap())

		if err := form.Run(); err != nil {
			return IsUserAbort(err)
		}
	} else {
		hookName = args[0]
	}

	path, isScript := hooks.ResolveHook(hookName, customDir, communityDir)
	if !isScript {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "hooks enable", nil,
				ui.NewCLIError(ui.ErrCodeNotFound, fmt.Sprintf("hook not found: %s", hookName)))
		}
		return fmt.Errorf("hook not found: %s", hookName)
	}

	// Read flag inside function to avoid package-level state (concurrent test safety)
	eventFlag, _ := cmd.Flags().GetString("event")

	var events []string
	if eventFlag != "" {
		events = []string{eventFlag}
	} else {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		meta, _ := hooks.ParseMetadata(string(content))
		events = meta.Events
		if len(events) == 0 {
			if IsJSONOutput() {
				return ui.OutputJSON(os.Stdout, "hooks enable", nil,
					ui.NewCLIError(ui.ErrCodeValidation, "hook has no declared events, use --event to specify"))
			}
			return fmt.Errorf("hook has no declared events, use --event to specify")
		}
	}

	configPath := config.GetConfigPath()
	for _, event := range events {
		if err := config.AddHookToConfig(configPath, hookName, event); err != nil {
			return err
		}
	}

	if IsJSONOutput() {
		return ui.OutputJSON(os.Stdout, "hooks enable", map[string]interface{}{
			"hook":   hookName,
			"events": events,
		}, nil)
	}

	fmt.Println(ui.SuccessMsg(fmt.Sprintf("Enabled %s for: %s", hookName, strings.Join(events, ", "))))
	return nil
}

func runHooksDisable(cmd *cobra.Command, args []string) error {
	hookName := args[0]
	configPath := config.GetConfigPath()

	if err := config.RemoveHookFromConfig(configPath, hookName); err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "hooks disable", nil, ui.NewCLIError(ui.ErrCodeGit, err.Error()))
		}
		return err
	}

	if IsJSONOutput() {
		return ui.OutputJSON(os.Stdout, "hooks disable", map[string]interface{}{
			"hook": hookName,
		}, nil)
	}

	fmt.Println(ui.SuccessMsg(fmt.Sprintf("Disabled %s", hookName)))
	return nil
}

func runHooksShow(cmd *cobra.Command, args []string) error {
	hookName := args[0]

	communityDir := config.GetCommunityHooksDir()
	customDir := config.GetCustomHooksDir()

	path, isScript := hooks.ResolveHook(hookName, customDir, communityDir)
	if !isScript {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "hooks show", nil,
				ui.NewCLIError(ui.ErrCodeNotFound, fmt.Sprintf("hook not found: %s", hookName)))
		}
		return fmt.Errorf("hook not found: %s", hookName)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if IsJSONOutput() {
			return ui.OutputJSON(os.Stdout, "hooks show", nil,
				ui.NewCLIError(ui.ErrCodeNotFound, fmt.Sprintf("failed to read hook: %s", err)))
		}
		return fmt.Errorf("failed to read hook: %w", err)
	}

	meta, _ := hooks.ParseMetadata(string(content))

	if IsJSONOutput() {
		return ui.OutputJSON(os.Stdout, "hooks show", map[string]interface{}{
			"name":        hookName,
			"path":        path,
			"description": meta.Description,
			"events":      meta.Events,
			"requires":    meta.Requires,
			"content":     string(content),
		}, nil)
	}

	fmt.Println(ui.BoldStyle.Render(hookName))
	fmt.Println()

	if meta.Description != "" {
		fmt.Printf("Description: %s\n", meta.Description)
	}
	if len(meta.Events) > 0 {
		fmt.Printf("Events:      %s\n", strings.Join(meta.Events, ", "))
	}
	if meta.Requires != "" {
		fmt.Printf("Requires:    %s\n", meta.Requires)
	}
	fmt.Printf("Path:        %s\n", path)

	fmt.Println()
	fmt.Println(ui.SubtleStyle.Render(strings.Repeat("-", 60)))
	fmt.Println()
	fmt.Print(string(content))

	return nil
}
