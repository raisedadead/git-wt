package commands

import (
	"fmt"
	"os"

	"github.com/raisedadead/git-wt/internal/config"
	"github.com/raisedadead/git-wt/internal/ui"
	"github.com/spf13/cobra"
)

var version = "dev"

// Global flags
var jsonOutputFlag bool

// IsJSONOutput returns true if JSON output is enabled
func IsJSONOutput() bool {
	return jsonOutputFlag
}

var rootCmd = &cobra.Command{
	Use:   "git-wt",
	Short: "Git worktree manager with bare repo support",
	Long: `git-wt streamlines the bare repository + worktree workflow.

Create isolated worktrees for features, issues, and PRs with
customizable post-create hooks.`,
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutputFlag, "json", false, "Output in JSON format")
	rootCmd.SetVersionTemplate(fmt.Sprintf("%s\n", ui.TitleStyle.Render("git-wt version {{.Version}}")))
}

func Execute() {
	err := rootCmd.Execute()

	// Show first-run hint (only once, only on success, only if not JSON)
	if err == nil && !jsonOutputFlag && !config.IsInitialized() {
		fmt.Println()
		fmt.Println(ui.SubtleStyle.Render("Tip: Customize git-wt at " + config.GetConfigPath()))
		config.MarkInitialized()
	}

	if err != nil {
		os.Exit(ui.GetExitCode(err))
	}
}
