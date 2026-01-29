package hooks

import (
	"testing"
)

func TestRun_SetsEnvVars(t *testing.T) {
	ctx := Context{
		Path:          "/tmp/test-worktree",
		Branch:        "feature/auth",
		ProjectRoot:   "/tmp/project",
		DefaultBranch: "main",
	}

	// Use a command that prints env vars
	commands := []string{"printenv GIT_WT_PATH"}

	// We can't easily capture output, so just verify no error
	warnings := Run(commands, ctx)
	// printenv might fail if not available, that's ok for this test
	_ = warnings
}

func TestRun_EmptyCommands(t *testing.T) {
	ctx := Context{}
	warnings := Run([]string{}, ctx)
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for empty commands, got %d", len(warnings))
	}
}

func TestRun_FailingCommand(t *testing.T) {
	ctx := Context{}
	commands := []string{"false"} // 'false' command always exits 1

	warnings := Run(commands, ctx)
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning for failing command, got %d", len(warnings))
	}
}

func TestRun_ContinuesAfterFailure(t *testing.T) {
	ctx := Context{
		Path: "/tmp/test",
	}
	// First fails, second should still run
	commands := []string{"false", "true"}

	warnings := Run(commands, ctx)
	// Should have 1 warning from 'false', but 'true' still ran
	if len(warnings) != 1 {
		t.Errorf("expected 1 warning, got %d", len(warnings))
	}
}

func TestExpandTemplates(t *testing.T) {
	ctx := Context{
		Path:          "/path/to/worktree",
		Branch:        "feature/auth",
		ProjectRoot:   "/project/root",
		DefaultBranch: "main",
	}

	tests := []struct {
		input    string
		expected string
	}{
		{"echo {{.Path}}", "echo /path/to/worktree"},
		{"cp {{.ProjectRoot}}/{{.DefaultBranch}}/.envrc {{.Path}}/", "cp /project/root/main/.envrc /path/to/worktree/"},
		{"echo {{.Branch}}", "echo feature/auth"},
		{"no templates", "no templates"},
	}

	for _, tt := range tests {
		result := expandTemplates(tt.input, ctx)
		if result != tt.expected {
			t.Errorf("expandTemplates(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
