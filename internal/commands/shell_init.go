package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var shellInitCmd = &cobra.Command{
	Use:   "shell-init [bash|zsh|fish]",
	Short: "Print shell integration script for directory switching",
	Long: `Print a shell wrapper function that enables wt switch, wt add --switch,
and bare wt (TUI) to change the shell's working directory.

Without arguments, prints setup instructions for all supported shells.
With a shell name, outputs the wrapper function for that shell.`,
	Args:              cobra.MaximumNArgs(1),
	ValidArgsFunction: completeShellInitNames,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), `# Bash
eval "$(wt shell-init bash)"

# Zsh
eval "$(wt shell-init zsh)"

# Fish
wt shell-init fish | source`)
			return err
		}

		switch args[0] {
		case "bash":
			return genBashWrapper(os.Stdout)
		case "zsh":
			return genZshWrapper(os.Stdout)
		case "fish":
			return genFishWrapper(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s (use bash, zsh, or fish)", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(shellInitCmd)
}

func completeShellInitNames(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	shells := []string{"bash", "zsh", "fish"}
	var matches []string
	for _, s := range shells {
		if strings.HasPrefix(s, toComplete) {
			matches = append(matches, s)
		}
	}

	return matches, cobra.ShellCompDirectiveNoFileComp
}

const bashWrapper = `# wt wrapper function for directory switching (switch, add --switch, TUI)
wt() {
    if [[ "$1" == "switch" || "$1" == "cd" ]]; then
        local target
        target="$(command wt switch "${@:2}")"
        local ret=$?
        if [[ $ret -eq 0 && -n "$target" && -d "$target" ]]; then
            cd "$target" || return 1
        else
            return $ret
        fi
    elif [[ "$1" == "add" || "$1" == "new" || "$1" == "create" ]]; then
        local has_switch=false
        for arg in "${@:2}"; do
            if [[ "$arg" == "--switch" || "$arg" == "-s" ]]; then
                has_switch=true
                break
            fi
        done
        if [[ "$has_switch" == true ]]; then
            local target
            target="$(command wt "$@")"
            local ret=$?
            if [[ $ret -eq 0 && -n "$target" && -d "$target" ]]; then
                cd "$target" || return 1
            else
                return $ret
            fi
        else
            command wt "$@"
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

const fishWrapper = `# wt wrapper function for directory switching (switch, add --switch, TUI)
function wt --wraps='command wt' --description 'git worktree manager'
    if test (count $argv) -gt 0 && begin; test "$argv[1]" = "switch" || test "$argv[1]" = "cd"; end
        set -l target (command wt switch $argv[2..])
        set -l ret $status
        if test $ret -eq 0 && test -n "$target" && test -d "$target"
            cd "$target"
        else
            return $ret
        end
    else if test (count $argv) -gt 0 && begin; test "$argv[1]" = "add" || test "$argv[1]" = "new" || test "$argv[1]" = "create"; end
        set -l has_switch false
        for arg in $argv[2..]
            if test "$arg" = "--switch" || test "$arg" = "-s"
                set has_switch true
                break
            end
        end
        if test "$has_switch" = true
            set -l target (command wt $argv)
            set -l ret $status
            if test $ret -eq 0 && test -n "$target" && test -d "$target"
                cd "$target"
            else
                return $ret
            end
        else
            command wt $argv
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

func genBashWrapper(w io.Writer) error {
	_, err := fmt.Fprint(w, bashWrapper)
	return err
}

func genZshWrapper(w io.Writer) error {
	_, err := fmt.Fprint(w, bashWrapper)
	return err
}

func genFishWrapper(w io.Writer) error {
	_, err := fmt.Fprint(w, fishWrapper)
	return err
}
