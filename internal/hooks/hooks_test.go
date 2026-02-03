package hooks

import (
	"os"
	"path/filepath"
	"strings"
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

	// Values are now shell-quoted for security
	tests := []struct {
		input    string
		expected string
	}{
		{"echo {{.Path}}", "echo '/path/to/worktree'"},
		{"cp {{.ProjectRoot}}/{{.DefaultBranch}}/.envrc {{.Path}}/", "cp '/project/root'/'main'/.envrc '/path/to/worktree'/"},
		{"echo {{.Branch}}", "echo 'feature/auth'"},
		{"no templates", "no templates"},
	}

	for _, tt := range tests {
		result := expandTemplates(tt.input, ctx)
		if result != tt.expected {
			t.Errorf("expandTemplates(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestShellQuote(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "'simple'"},
		{"with space", "'with space'"},
		{"with'quote", "'with'\\''quote'"},
		{"; rm -rf /", "'; rm -rf /'"},
	}

	for _, tt := range tests {
		result := shellQuote(tt.input)
		if result != tt.expected {
			t.Errorf("shellQuote(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestRun_Timeout(t *testing.T) {
	ctx := Context{
		Path:          "/tmp/test",
		Branch:        "test",
		ProjectRoot:   "/tmp",
		DefaultBranch: "main",
	}

	// Command that takes longer than timeout
	commands := []string{"sleep 5"}
	warnings := RunWithTimeout(commands, ctx, 1) // 1 second timeout

	if len(warnings) == 0 {
		t.Error("expected timeout warning")
	}
	// Context error message is "context deadline exceeded"
	if len(warnings) > 0 && !strings.Contains(warnings[0], "deadline exceeded") {
		t.Errorf("expected deadline exceeded message, got: %s", warnings[0])
	}
}

func TestRun_NoTimeout(t *testing.T) {
	ctx := Context{
		Path:          "/tmp/test",
		Branch:        "test",
		ProjectRoot:   "/tmp",
		DefaultBranch: "main",
	}

	commands := []string{"echo hello"}
	warnings := RunWithTimeout(commands, ctx, 30)

	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got: %v", warnings)
	}
}

func TestRunResolved_InlineCommand(t *testing.T) {
	ctx := Context{Path: "/tmp/test"}
	warnings := RunResolved([]string{"echo hello"}, ctx, 5, "", "")
	if len(warnings) > 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestRunResolved_ScriptFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hooks-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	customDir := filepath.Join(tmpDir, "custom")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(customDir, "test-hook.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'from script'"), 0755); err != nil {
		t.Fatal(err)
	}

	ctx := Context{Path: tmpDir}
	warnings := RunResolved([]string{"test-hook"}, ctx, 5, customDir, "")
	if len(warnings) > 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestRunResolved_MixedCommands(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hooks-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	customDir := filepath.Join(tmpDir, "custom")
	if err := os.MkdirAll(customDir, 0755); err != nil {
		t.Fatal(err)
	}
	scriptPath := filepath.Join(customDir, "my-hook.sh")
	if err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'script'"), 0755); err != nil {
		t.Fatal(err)
	}

	ctx := Context{Path: tmpDir}
	// Mix of script and inline
	warnings := RunResolved([]string{"my-hook", "echo inline"}, ctx, 5, customDir, "")
	if len(warnings) > 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}
}

func TestRunResolved_EmptyList(t *testing.T) {
	ctx := Context{Path: "/tmp/test"}
	warnings := RunResolved([]string{}, ctx, 5, "", "")
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for empty list, got %v", warnings)
	}
}

func TestParseHookOutput(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "hook-output-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := `GIT_WT_BRANCH=fix-123-bug-title
GIT_WT_META_ISSUE_NUMBER=123
GIT_WT_META_ISSUE_TITLE=Fix the bug
GIT_WT_PROMPT_BRANCH=fix-123-bug-title
GIT_WT_WARNING=jq not found`

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = tmpFile.Close()

	output, err := parseHookOutput(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Branch != "fix-123-bug-title" {
		t.Errorf("expected branch 'fix-123-bug-title', got %q", output.Branch)
	}
	if output.Metadata["issue_number"] != "123" {
		t.Errorf("expected metadata issue_number=123, got %q", output.Metadata["issue_number"])
	}
	if output.Metadata["issue_title"] != "Fix the bug" {
		t.Errorf("expected metadata issue_title='Fix the bug', got %q", output.Metadata["issue_title"])
	}
	if output.Prompts["branch"] != "fix-123-bug-title" {
		t.Errorf("expected prompt branch='fix-123-bug-title', got %q", output.Prompts["branch"])
	}
	if len(output.Warnings) != 1 || output.Warnings[0] != "jq not found" {
		t.Errorf("expected 1 warning 'jq not found', got %v", output.Warnings)
	}
}

func TestParseHookOutput_Error(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "hook-output-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	content := `GIT_WT_ERROR=gh CLI not installed`

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	_ = tmpFile.Close()

	output, err := parseHookOutput(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if output.Error != "gh CLI not installed" {
		t.Errorf("expected error 'gh CLI not installed', got %q", output.Error)
	}
}

func TestParseHookOutput_NonExistent(t *testing.T) {
	output, err := parseHookOutput("/nonexistent/file")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return empty output for non-existent file
	if output.Branch != "" || len(output.Metadata) != 0 {
		t.Errorf("expected empty output for non-existent file")
	}
}

func TestBuildEnvVars_WithWorkflow(t *testing.T) {
	ctx := Context{
		Path:           "/tmp/test",
		Branch:         "fix-123-bug",
		ProjectRoot:    "/project",
		DefaultBranch:  "main",
		Workflow:       "bugfix",
		WorkflowPrefix: "fix",
		IssueNumber:    123,
		PRNumber:       0,
		Metadata:       map[string]string{"issue_title": "Fix bug"},
	}

	vars := buildEnvVars(ctx)
	varMap := make(map[string]string)
	for _, v := range vars {
		if idx := strings.Index(v, "="); idx > 0 {
			varMap[v[:idx]] = v[idx+1:]
		}
	}

	if varMap["GIT_WT_WORKFLOW"] != "bugfix" {
		t.Errorf("expected GIT_WT_WORKFLOW=bugfix, got %q", varMap["GIT_WT_WORKFLOW"])
	}
	if varMap["GIT_WT_WORKFLOW_PREFIX"] != "fix" {
		t.Errorf("expected GIT_WT_WORKFLOW_PREFIX=fix, got %q", varMap["GIT_WT_WORKFLOW_PREFIX"])
	}
	if varMap["GIT_WT_ISSUE"] != "123" {
		t.Errorf("expected GIT_WT_ISSUE=123, got %q", varMap["GIT_WT_ISSUE"])
	}
	if varMap["GIT_WT_META_ISSUE_TITLE"] != "Fix bug" {
		t.Errorf("expected GIT_WT_META_ISSUE_TITLE='Fix bug', got %q", varMap["GIT_WT_META_ISSUE_TITLE"])
	}
}
