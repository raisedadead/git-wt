package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Print shell completion instructions or generate scripts",
	Long:  `Print instructions for setting up shell completions, or generate completion scripts for a specific shell.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Print setup instructions for all shells
			fmt.Println(`# Bash
source <(git-wt completion bash)

# Zsh
echo "autoload -U compinit; compinit" >> ~/.zshrc
git-wt completion zsh > "${fpath[1]}/_git-wt"

# Fish
git-wt completion fish > ~/.config/fish/completions/git-wt.fish

# PowerShell
git-wt completion powershell >> $PROFILE`)
			return nil
		}

		// Generate completion script for specified shell
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s (use bash, zsh, fish, or powershell)", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
