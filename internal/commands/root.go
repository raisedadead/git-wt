package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
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

// DefaultFormKeyMap returns a huh KeyMap with ESC added to quit binding
// so users can cancel forms with either Ctrl+C or ESC
func DefaultFormKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("ctrl+c/esc", "quit"),
	)
	return km
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
	// Silence Cobra's default error/usage printing - we handle it in Execute()
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
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
		// Handle silent exit (user abort with Ctrl+C/ESC)
		var silent SilentExit
		if errors.As(err, &silent) {
			os.Exit(silent.Code)
		}
		// Print error for non-silent errors (since we disabled Cobra's error printing)
		fmt.Fprintln(os.Stderr, ui.ErrorStyle.Render("Error: "+err.Error()))
		os.Exit(ui.GetExitCode(err))
	}
}
