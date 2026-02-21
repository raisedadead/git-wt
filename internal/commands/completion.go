package commands

import (
	"bytes"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var completionCmd = &cobra.Command{
	Use:               "completion [bash|zsh|fish|powershell]",
	Short:             "Print shell completion instructions or generate scripts",
	Long:              `Print instructions for setting up shell completions, or generate completion scripts for a specific shell.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeShellNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// Print setup instructions for all shells
			fmt.Println(`# Bash
source <(wt completion bash)

# Zsh
echo "autoload -U compinit; compinit" >> ~/.zshrc
wt completion zsh > "${fpath[1]}/_wt"

# Fish
wt completion fish > ~/.config/fish/completions/wt.fish

# PowerShell
wt completion powershell >> $PROFILE`)
			return nil
		}

		// Generate completion script for specified shell
		switch args[0] {
		case "bash":
			return genBashCompletionWithWrapper(os.Stdout)
		case "zsh":
			return genZshCompletionWithWrapper(os.Stdout)
		case "fish":
			return genFishCompletionWithWrapper(os.Stdout)
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

// genBashCompletionWithWrapper generates bash completions with a wt wrapper function
func genBashCompletionWithWrapper(w *os.File) error {
	var buf bytes.Buffer
	if err := rootCmd.GenBashCompletion(&buf); err != nil {
		return err
	}

	// Write the standard completions
	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}

	// Append the wrapper function
	wrapper := `
# wt wrapper function to handle 'switch' and TUI directory change
wt() {
    if [[ "$1" == "switch" ]]; then
        local target
        target="$(command wt switch "${@:2}")"
        local ret=$?
        if [[ $ret -eq 0 && -n "$target" && -d "$target" ]]; then
            cd "$target" || return 1
        else
            return $ret
        fi
    elif [[ $# -eq 0 ]]; then
        local target
        target="$(command wt)"
        local ret=$?
        if [[ $ret -eq 0 && -n "$target" && -d "$target" ]]; then
            cd "$target" || return 1
        elif [[ -n "$target" ]]; then
            echo "$target"
        fi
        return $ret
    else
        command wt "$@"
    fi
}
`
	_, err := w.WriteString(wrapper)
	return err
}

// genZshCompletionWithWrapper generates zsh completions with a wt wrapper function
func genZshCompletionWithWrapper(w *os.File) error {
	var buf bytes.Buffer
	if err := rootCmd.GenZshCompletion(&buf); err != nil {
		return err
	}

	// Write the standard completions
	if _, err := w.Write(buf.Bytes()); err != nil {
		return err
	}

	// Append the wrapper function
	wrapper := `
# wt wrapper function to handle 'switch' and TUI directory change
wt() {
    if [[ "$1" == "switch" ]]; then
        local target
        target="$(command wt switch "${@:2}")"
        local ret=$?
        if [[ $ret -eq 0 && -n "$target" && -d "$target" ]]; then
            cd "$target" || return 1
        else
            return $ret
        fi
    elif [[ $# -eq 0 ]]; then
        local target
        target="$(command wt)"
        local ret=$?
        if [[ $ret -eq 0 && -n "$target" && -d "$target" ]]; then
            cd "$target" || return 1
        elif [[ -n "$target" ]]; then
            echo "$target"
        fi
        return $ret
    else
        command wt "$@"
    fi
}
`
	_, err := w.WriteString(wrapper)
	return err
}

// genFishCompletionWithWrapper generates fish completions with a wt wrapper function
func genFishCompletionWithWrapper(w *os.File) error {
	if err := rootCmd.GenFishCompletion(w, true); err != nil {
		return err
	}

	// Append the wrapper function
	wrapper := `
# wt wrapper function to handle 'switch' and TUI directory change
function wt --wraps='command wt' --description 'git worktree manager'
    if test (count $argv) -gt 0 && test "$argv[1]" = "switch"
        set -l target (command wt switch $argv[2..])
        set -l ret $status
        if test $ret -eq 0 && test -n "$target" && test -d "$target"
            cd "$target"
        else
            return $ret
        end
    else if test (count $argv) -eq 0
        set -l target (command wt)
        set -l ret $status
        if test $ret -eq 0 && test -n "$target" && test -d "$target"
            cd "$target"
        else if test -n "$target"
            echo "$target"
        end
        return $ret
    else
        command wt $argv
    end
end
`
	_, err := w.WriteString(wrapper)
	return err
}
