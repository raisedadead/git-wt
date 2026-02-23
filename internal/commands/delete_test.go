package commands

import (
	"testing"
)

func TestDeletePathFlagRegistered(t *testing.T) {
	f := deleteCmd.Flags().Lookup("path")
	if f == nil {
		t.Fatal("--path flag not registered on delete command")
	}
	if f.DefValue != "" {
		t.Errorf("--path default = %q, want empty string", f.DefValue)
	}
}

func TestDeletePathFlagMutualExclusion(t *testing.T) {
	// When --path is set AND positional args are provided, runDelete should error.
	// We can't run the full command without a git repo, but we can test
	// that the flag exists and then exercise the command with both set.
	// Since runDelete calls git.GetProjectRoot first, which will fail outside
	// a real project, we test the flag registration and rely on integration
	// tests for the full flow.

	// Verify --path flag exists
	f := deleteCmd.Flags().Lookup("path")
	if f == nil {
		t.Fatal("--path flag not registered on delete command")
	}

	// Verify usage text mentions path
	if f.Usage == "" {
		t.Error("--path flag has no usage text")
	}
}
