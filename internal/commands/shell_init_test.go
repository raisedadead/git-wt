package commands

import (
	"bytes"
	"strings"
	"testing"
)

// TestShellInitCommandExists verifies the shell-init command is registered
func TestShellInitCommandExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"shell-init"})
	if err != nil {
		t.Fatalf("shell-init command not found: %v", err)
	}
	if !strings.HasPrefix(cmd.Use, "shell-init") {
		t.Errorf("unexpected command use: %s", cmd.Use)
	}
}

// TestShellInitCommandHelp checks the command documentation
func TestShellInitCommandHelp(t *testing.T) {
	cmd, _, _ := rootCmd.Find([]string{"shell-init"})

	if cmd.Short == "" {
		t.Error("short description should not be empty")
	}
	if !strings.Contains(cmd.Long, "switch") {
		t.Error("long description should mention switch")
	}
	if !strings.Contains(cmd.Long, "TUI") {
		t.Error("long description should mention TUI")
	}
}

// TestShellInitNoArgs prints setup instructions
func TestShellInitNoArgs(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetArgs([]string{"shell-init"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()
	rootCmd.SetOut(nil)

	if !strings.Contains(output, `eval "$(wt shell-init bash)"`) {
		t.Error("setup instructions should contain bash eval")
	}
	if !strings.Contains(output, `eval "$(wt shell-init zsh)"`) {
		t.Error("setup instructions should contain zsh eval")
	}
	if !strings.Contains(output, "wt shell-init fish | source") {
		t.Error("setup instructions should contain fish source")
	}
}

// TestShellInitBashWrapper tests bash wrapper output
func TestShellInitBashWrapper(t *testing.T) {
	var buf bytes.Buffer
	if err := genBashWrapper(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()

	checks := []string{
		"wt()",
		`"$1" == "switch"`,
		`command wt switch`,
		`cd "$target"`,
		`"$1" == "add" || "$1" == "new"`,
		`"--switch"`,
		`"-s"`,
		`$# -eq 0`,
		`command wt "$@"`,
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("bash wrapper missing: %s", check)
		}
	}
}

// TestShellInitZshWrapper tests zsh wrapper output (same as bash)
func TestShellInitZshWrapper(t *testing.T) {
	var bashBuf, zshBuf bytes.Buffer
	if err := genBashWrapper(&bashBuf); err != nil {
		t.Fatalf("unexpected error generating bash: %v", err)
	}
	if err := genZshWrapper(&zshBuf); err != nil {
		t.Fatalf("unexpected error generating zsh: %v", err)
	}
	if bashBuf.String() != zshBuf.String() {
		t.Error("zsh wrapper should be identical to bash wrapper")
	}
}

// TestShellInitFishWrapper tests fish wrapper output
func TestShellInitFishWrapper(t *testing.T) {
	var buf bytes.Buffer
	if err := genFishWrapper(&buf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	output := buf.String()

	checks := []string{
		"function wt",
		`--wraps='command wt'`,
		`"$argv[1]" = "switch"`,
		`command wt switch $argv[2..]`,
		`cd "$target"`,
		`"$argv[1]" = "add"`,
		`"$argv[1]" = "new"`,
		`"--switch"`,
		`"-s"`,
		`(count $argv) -eq 0`,
		`command wt $argv`,
		"end",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("fish wrapper missing: %s", check)
		}
	}
}

// TestShellInitUnsupportedShell tests error for unsupported shell
func TestShellInitUnsupportedShell(t *testing.T) {
	rootCmd.SetArgs([]string{"shell-init", "powershell"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported shell")
	}
	if !strings.Contains(err.Error(), "unsupported shell: powershell") {
		t.Errorf("unexpected error: %v", err)
	}
}

// TestShellInitValidArgs tests shell name completion
func TestShellInitValidArgs(t *testing.T) {
	cmd, _, _ := rootCmd.Find([]string{"shell-init"})
	if cmd.ValidArgsFunction == nil {
		t.Fatal("ValidArgsFunction should be set")
	}
}

// TestShellInitMaxOneArg tests that at most one arg is accepted
func TestShellInitMaxOneArg(t *testing.T) {
	rootCmd.SetArgs([]string{"shell-init", "bash", "zsh"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for too many args")
	}
}

// TestBashCompletionNoWrapper verifies completions don't contain the shell wrapper
func TestBashCompletionNoWrapper(t *testing.T) {
	var buf bytes.Buffer
	err := rootCmd.GenBashCompletion(&buf)
	if err != nil {
		t.Fatalf("failed to generate bash completion: %v", err)
	}
	content := buf.String()
	if strings.Contains(content, "wt() {") {
		t.Error("bash completion should not contain wt() wrapper function")
	}
}

// TestZshCompletionNoWrapper verifies completions don't contain the shell wrapper
func TestZshCompletionNoWrapper(t *testing.T) {
	var buf bytes.Buffer
	err := rootCmd.GenZshCompletion(&buf)
	if err != nil {
		t.Fatalf("failed to generate zsh completion: %v", err)
	}
	content := buf.String()
	if strings.Contains(content, "wt() {") {
		t.Error("zsh completion should not contain wt() wrapper function")
	}
}

// TestFishCompletionNoWrapper verifies completions don't contain the shell wrapper
func TestFishCompletionNoWrapper(t *testing.T) {
	var buf bytes.Buffer
	err := rootCmd.GenFishCompletion(&buf, true)
	if err != nil {
		t.Fatalf("failed to generate fish completion: %v", err)
	}
	content := buf.String()
	if strings.Contains(content, "function wt --wraps") {
		t.Error("fish completion should not contain 'function wt --wraps' wrapper")
	}
}

// TestCompleteShellInitNames tests the shell-init completion function
func TestCompleteShellInitNames(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		toComplete string
		expected   []string
	}{
		{
			name:       "no input returns all shells",
			args:       nil,
			toComplete: "",
			expected:   []string{"bash", "zsh", "fish"},
		},
		{
			name:       "prefix b matches bash",
			args:       nil,
			toComplete: "b",
			expected:   []string{"bash"},
		},
		{
			name:       "prefix f matches fish",
			args:       nil,
			toComplete: "f",
			expected:   []string{"fish"},
		},
		{
			name:       "already has arg returns nothing",
			args:       []string{"bash"},
			toComplete: "",
			expected:   nil,
		},
	}

	cmd, _, _ := rootCmd.Find([]string{"shell-init"})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := completeShellInitNames(cmd, tt.args, tt.toComplete)
			if len(result) != len(tt.expected) {
				t.Errorf("completeShellInitNames(%v, %q) = %v, want %v", tt.args, tt.toComplete, result, tt.expected)
				return
			}
			for i, r := range result {
				if r != tt.expected[i] {
					t.Errorf("completeShellInitNames(%v, %q)[%d] = %q, want %q", tt.args, tt.toComplete, i, r, tt.expected[i])
				}
			}
		})
	}
}
