package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for git-wt.

To load completions:

Bash:
  $ source <(git-wt completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ git-wt completion bash > /etc/bash_completion.d/git-wt
  # macOS:
  $ git-wt completion bash > $(brew --prefix)/etc/bash_completion.d/git-wt

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ git-wt completion zsh > "${fpath[1]}/_git-wt"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ git-wt completion fish | source

  # To load completions for each session, execute once:
  $ git-wt completion fish > ~/.config/fish/completions/git-wt.fish

PowerShell:
  PS> git-wt completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> git-wt completion powershell > git-wt.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
