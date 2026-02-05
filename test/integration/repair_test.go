package integration

import (
	"path/filepath"
	"testing"
)

func TestRepair(t *testing.T) {
	t.Parallel()

	t.Run("no_op_when_not_needed", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "repair-noop", "--timeout", "300")
		projectDir := filepath.Join(workspace, "repair-noop")
		mainDir := filepath.Join(projectDir, "main")

		// Running repair on a healthy repo should succeed with no changes
		stdout := runGitWTSuccess(t, mainDir, "repair")
		assertContains(t, stdout, "correct")
	})

	t.Run("works_from_any_worktree", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		runGitWTSuccess(t, workspace, "clone", localRemote, "repair-worktree", "--timeout", "300")
		projectDir := filepath.Join(workspace, "repair-worktree")
		mainDir := filepath.Join(projectDir, "main")

		// Add a feature worktree
		runGitWTSuccess(t, mainDir, "add", "feature-repair-test", "--new")
		featureDir := filepath.Join(projectDir, "feature-repair-test")

		// Run repair from the feature worktree
		stdout := runGitWTSuccess(t, featureDir, "repair")
		assertContains(t, stdout, "correct")
	})

	t.Run("fails_outside_project", func(t *testing.T) {
		t.Parallel()
		workspace := setupTestWorkspace(t)

		// Run repair from a directory that is not a wt project
		_, stderr := runGitWTFail(t, workspace, "repair")
		assertContains(t, stderr, "not in a wt project")
	})
}

func TestRepairJSON(t *testing.T) {
	t.Parallel()

	workspace := setupTestWorkspace(t)

	runGitWTSuccess(t, workspace, "clone", localRemote, "repair-json", "--timeout", "300")
	projectDir := filepath.Join(workspace, "repair-json")
	mainDir := filepath.Join(projectDir, "main")

	result := runGitWTJSON(t, mainDir, "repair")

	if success, ok := result["success"].(bool); !ok || !success {
		t.Errorf("Expected success: true")
	}

	data, ok := result["data"].(map[string]any)
	if !ok {
		t.Fatalf("Expected data object")
	}

	// Check required fields
	requiredFields := []string{"project_root", "repaired"}
	for _, field := range requiredFields {
		if _, ok := data[field]; !ok {
			t.Errorf("Missing required field: %s", field)
		}
	}

	// On a healthy repo, repaired should be false
	if repaired, ok := data["repaired"].(bool); !ok || repaired {
		t.Errorf("Expected repaired: false for healthy repo")
	}
}
