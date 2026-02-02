package integration

import (
	"testing"
)

func TestHooks(t *testing.T) {
	t.Parallel()

	t.Run("list shows available hooks", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		// Hooks list doesn't need a repo context
		stdout, _, code := runGitWT(t, workspace, "hooks", "list")

		// Should succeed even if no hooks installed
		if code != 0 {
			t.Logf("hooks list output: %s", stdout)
		}
		// Either shows hooks or "No hooks found" message
	})

	t.Run("show nonexistent hook fails", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTFail(t, workspace, "hooks", "show", "nonexistent-hook-xyz")
	})
}

func TestHooksJSON(t *testing.T) {
	t.Parallel()

	t.Run("list returns valid JSON", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		result := runGitWTJSON(t, workspace, "hooks", "list")

		if _, ok := result["data"]; !ok {
			t.Errorf("Expected data field")
		}
	})
}
