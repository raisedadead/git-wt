package commands

import (
	"strings"
	"testing"
)

// TestCloneCommandExists verifies the clone command is registered
func TestCloneCommandExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"clone"})
	if err != nil {
		t.Fatalf("clone command not found: %v", err)
	}
	if !strings.HasPrefix(cmd.Use, "clone") {
		t.Errorf("unexpected command use: %s", cmd.Use)
	}
}

// TestCloneCommandHelp checks the command documentation
func TestCloneCommandHelp(t *testing.T) {
	cmd, _, _ := rootCmd.Find([]string{"clone"})

	// Check short description
	if !strings.Contains(cmd.Short, "Clone") {
		t.Errorf("unexpected short description: %s", cmd.Short)
	}

	// Check long description mentions supported formats
	if !strings.Contains(cmd.Long, "owner/repo") {
		t.Errorf("long description should mention owner/repo shorthand: %s", cmd.Long)
	}
	if !strings.Contains(cmd.Long, "git@github.com") {
		t.Errorf("long description should mention SSH URL format: %s", cmd.Long)
	}
}

// TestCloneFormURLDescription verifies the URL input has a description explaining supported formats
func TestCloneFormURLDescription(t *testing.T) {
	// The description should explain the supported URL formats
	expected := "Supports: owner/repo, git@github.com:..., https://..."
	if CloneURLDescription != expected {
		t.Errorf("CloneURLDescription = %q, want %q", CloneURLDescription, expected)
	}
}

// TestCloneFormURLPlaceholder verifies the URL placeholder shows a simple example
func TestCloneFormURLPlaceholder(t *testing.T) {
	// Placeholder should show the simplest format: owner/repo
	expected := "owner/repo"
	if CloneURLPlaceholder != expected {
		t.Errorf("CloneURLPlaceholder = %q, want %q", CloneURLPlaceholder, expected)
	}
}
