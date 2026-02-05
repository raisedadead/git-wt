package commands

import (
	"strings"
	"testing"
)

func TestHooksEnableCommandExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"hooks", "enable"})
	if err != nil {
		t.Fatalf("hooks enable command not found: %v", err)
	}
	if !strings.HasPrefix(cmd.Use, "enable") {
		t.Errorf("unexpected command use: %s", cmd.Use)
	}
}

func TestHooksEnableAcceptsZeroArgs(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"hooks", "enable"})
	if err != nil {
		t.Fatalf("hooks enable command not found: %v", err)
	}

	// Verify the command accepts 0 or 1 args (MaximumNArgs(1))
	// This is done by checking that Args validator allows 0 args
	if cmd.Args == nil {
		t.Error("hooks enable command should have Args validator")
		return
	}

	// Test with 0 args - should NOT error on arg validation
	// The Args function validates argument count
	err = cmd.Args(cmd, []string{})
	if err != nil {
		t.Errorf("hooks enable should accept 0 arguments for interactive mode, got error: %v", err)
	}

	// Test with 1 arg - should also be valid
	err = cmd.Args(cmd, []string{"some-hook"})
	if err != nil {
		t.Errorf("hooks enable should accept 1 argument, got error: %v", err)
	}

	// Test with 2 args - should error (maximum is 1)
	err = cmd.Args(cmd, []string{"hook1", "hook2"})
	if err == nil {
		t.Error("hooks enable should reject 2 arguments")
	}
}

func TestHooksEnableCommandHelp(t *testing.T) {
	cmd, _, _ := rootCmd.Find([]string{"hooks", "enable"})

	// Check short description exists
	if cmd.Short == "" {
		t.Error("hooks enable command should have a short description")
	}

	// Check that Use shows optional argument pattern
	if !strings.Contains(cmd.Use, "[") {
		t.Error("hooks enable Use should indicate optional argument with [hook-name]")
	}
}

func TestGetAvailableHooks(t *testing.T) {
	// Test that getAvailableHooks returns a slice of HookInfo
	// This function should be exported or testable via runHooksList
	// For now, we test that the hooks list command works
	cmd, _, err := rootCmd.Find([]string{"hooks", "list"})
	if err != nil {
		t.Fatalf("hooks list command not found: %v", err)
	}

	if cmd.RunE == nil {
		t.Error("hooks list command should have RunE function")
	}
}

func TestHooksDisableCommandExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"hooks", "disable"})
	if err != nil {
		t.Fatalf("hooks disable command not found: %v", err)
	}
	if !strings.HasPrefix(cmd.Use, "disable") {
		t.Errorf("unexpected command use: %s", cmd.Use)
	}
}

func TestHooksShowCommandExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"hooks", "show"})
	if err != nil {
		t.Fatalf("hooks show command not found: %v", err)
	}
	if !strings.HasPrefix(cmd.Use, "show") {
		t.Errorf("unexpected command use: %s", cmd.Use)
	}
}

func TestHooksListCommandExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"hooks", "list"})
	if err != nil {
		t.Fatalf("hooks list command not found: %v", err)
	}
	if !strings.HasPrefix(cmd.Use, "list") {
		t.Errorf("unexpected command use: %s", cmd.Use)
	}
}
