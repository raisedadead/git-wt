package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/raisedadead/wt/internal/config"
	"github.com/raisedadead/wt/internal/ui"
	"github.com/spf13/cobra"
)

var version = "dev"

// Global flags
var jsonOutputFlag bool

// IsJSONOutput returns true if JSON output is enabled
func IsJSONOutput() bool {
	return jsonOutputFlag
}

// SilentExit is returned when we want to exit without printing an error
type SilentExit struct {
	Code int
}

func (e SilentExit) Error() string {
	return ""
}

// IsUserAbort checks if the error is a user abort (Ctrl+C) and returns
// a SilentExit error if so. Otherwise returns the original error.
func IsUserAbort(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, huh.ErrUserAborted) {
		return SilentExit{Code: 130} // 128 + SIGINT(2)
	}
	return err
}

var rootCmd = &cobra.Command{
	Use:   "wt",
	Short: "Git worktree manager with bare repo support",
	Long: `wt streamlines the bare repository + worktree workflow.

Create isolated worktrees for features, issues, and PRs with
customizable post-create hooks.`,
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutputFlag, "json", false, "Output in JSON format")
	rootCmd.SetVersionTemplate(fmt.Sprintf("%s\n", ui.TitleStyle.Render("wt version {{.Version}}")))
}

func Execute() {
	err := rootCmd.Execute()

	// Show first-run hint (only once, only on success, only if not JSON)
	if err == nil && !jsonOutputFlag && !config.IsInitialized() {
		fmt.Println()
		fmt.Println(ui.SubtleStyle.Render("Tip: Customize wt at " + config.GetConfigPath()))
		_ = config.MarkInitialized()
	}

	if err != nil {
		// Handle silent exit (user abort with Ctrl+C)
		var silent SilentExit
		if errors.As(err, &silent) {
			os.Exit(silent.Code)
		}
		os.Exit(ui.GetExitCode(err))
	}
}
