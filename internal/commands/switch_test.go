package commands

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/raisedadead/wt/internal/git"
)

func TestSwitchBranchMatching(t *testing.T) {
	// Test that flattened branch names match correctly
	tests := []struct {
		name       string
		branchArg  string
		wtBranch   string
		shouldFind bool
	}{
		{"exact match", "main", "main", true},
		{"exact match with slash", "feature/auth", "feature/auth", true},
		{"flattened match", "feature-auth", "feature/auth", true},
		{"flattened match deep", "fix-security-issue", "fix/security/issue", true},
		{"no match", "nonexistent", "main", false},
		{"partial no match", "feat", "feature/auth", false},
		{"case sensitive", "Main", "main", false},
		{"empty branch arg", "", "main", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate matching logic from switch command
			found := false
			if tt.wtBranch == tt.branchArg {
				found = true
			} else if git.FlattenBranchName(tt.wtBranch) == tt.branchArg {
				found = true
			}

			if found != tt.shouldFind {
				t.Errorf("branch matching for %q against %q: got %v, want %v",
					tt.branchArg, tt.wtBranch, found, tt.shouldFind)
			}
		})
	}
}

func TestSwitchCommandExists(t *testing.T) {
	// Verify the switch command is registered
	cmd, _, err := rootCmd.Find([]string{"switch"})
	if err != nil {
		t.Fatalf("switch command not found: %v", err)
	}
	if cmd.Use != "switch [branch]" {
		t.Errorf("unexpected command use: %s", cmd.Use)
	}
}

func TestSwitchCommandHelp(t *testing.T) {
	cmd, _, _ := rootCmd.Find([]string{"switch"})

	// Check short description
	if !strings.Contains(cmd.Short, "Switch to a worktree") {
		t.Errorf("unexpected short description: %s", cmd.Short)
	}

	// Check long description mentions shell completions
	if !strings.Contains(cmd.Long, "shell completions") {
		t.Errorf("long description should mention shell completions: %s", cmd.Long)
	}
}

func TestBashCompletionContainsWrapper(t *testing.T) {
	// Create a temp file to write to
	tmpFile, err := os.CreateTemp("", "bash-completion-*.bash")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	})

	// Generate bash completion
	err = genBashCompletionWithWrapper(tmpFile)
	if err != nil {
		t.Fatalf("failed to generate bash completion: %v", err)
	}

	// Read back and verify wrapper function is present
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	contentStr := string(content)

	// Check for wrapper function
	if !strings.Contains(contentStr, "wt()") {
		t.Error("bash completion should contain wt() wrapper function")
	}
	if !strings.Contains(contentStr, `"$1" == "switch"`) {
		t.Error("bash completion wrapper should handle switch command")
	}
	if !strings.Contains(contentStr, "command wt") {
		t.Error("bash completion wrapper should call 'command wt'")
	}
	if !strings.Contains(contentStr, "cd \"$target\"") {
		t.Error("bash completion wrapper should change directory")
	}
}

func TestZshCompletionContainsWrapper(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "zsh-completion-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	})

	err = genZshCompletionWithWrapper(tmpFile)
	if err != nil {
		t.Fatalf("failed to generate zsh completion: %v", err)
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "wt()") {
		t.Error("zsh completion should contain wt() wrapper function")
	}
	if !strings.Contains(contentStr, `"$1" == "switch"`) {
		t.Error("zsh completion wrapper should handle switch command")
	}
}

func TestFishCompletionContainsWrapper(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "fish-completion-*.fish")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	})

	err = genFishCompletionWithWrapper(tmpFile)
	if err != nil {
		t.Fatalf("failed to generate fish completion: %v", err)
	}

	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	contentStr := string(content)

	if !strings.Contains(contentStr, "function wt") {
		t.Error("fish completion should contain wt function")
	}
	if !strings.Contains(contentStr, `"switch"`) {
		t.Error("fish completion wrapper should handle switch command")
	}
	if !strings.Contains(contentStr, "cd \"$target\"") {
		t.Error("fish completion wrapper should change directory")
	}
}

func TestCompletionWrapperPreservesOriginal(t *testing.T) {
	// Ensure standard completion is still included
	var buf bytes.Buffer
	tmpFile, err := os.CreateTemp("", "test-completion-*")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	})

	// Generate standard completion for comparison
	_ = rootCmd.GenBashCompletion(&buf)
	standardCompletion := buf.String()

	// Generate with wrapper
	err = genBashCompletionWithWrapper(tmpFile)
	if err != nil {
		t.Fatalf("failed to generate completion: %v", err)
	}

	content, _ := os.ReadFile(tmpFile.Name())
	withWrapper := string(content)

	// Check that standard completion content is preserved
	// Look for characteristic bash completion patterns
	if !strings.Contains(withWrapper, "_wt_root_command") || !strings.Contains(standardCompletion, "_wt_root_command") {
		// Check for any completion function pattern
		if strings.Contains(standardCompletion, "__wt") && !strings.Contains(withWrapper, "__wt") {
			t.Error("wrapper completion should preserve standard completion functions")
		}
	}
}
