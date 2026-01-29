package ui

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestOutputJSON_Success(t *testing.T) {
	var buf bytes.Buffer
	data := map[string]string{"branch": "main", "path": "/tmp/worktree"}

	err := OutputJSON(&buf, "new", data, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !resp.Success {
		t.Error("expected Success to be true")
	}
	if resp.Command != "new" {
		t.Errorf("expected Command to be 'new', got %q", resp.Command)
	}
	if resp.Error != nil {
		t.Error("expected Error to be nil for success response")
	}
	if resp.Data == nil {
		t.Error("expected Data to be non-nil")
	}
}

func TestOutputJSON_Error(t *testing.T) {
	var buf bytes.Buffer
	cliErr := NewCLIError(ErrCodeValidation, "branch name is required")

	err := OutputJSON(&buf, "new", nil, cliErr)
	if err != nil {
		t.Fatalf("expected no error writing JSON, got %v", err)
	}

	var resp Response
	if err := json.Unmarshal(buf.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Success {
		t.Error("expected Success to be false for error response")
	}
	if resp.Command != "new" {
		t.Errorf("expected Command to be 'new', got %q", resp.Command)
	}
	if resp.Error == nil {
		t.Fatal("expected Error to be non-nil")
	}
	if resp.Error.Code != ErrCodeValidation {
		t.Errorf("expected error code %q, got %q", ErrCodeValidation, resp.Error.Code)
	}
	if resp.Error.Message != "branch name is required" {
		t.Errorf("expected error message %q, got %q", "branch name is required", resp.Error.Message)
	}
}

func TestGetExitCode_CLIError(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		expected int
	}{
		{"validation error", ErrCodeValidation, ExitValidation},
		{"git error", ErrCodeGit, ExitGit},
		{"github error", ErrCodeGitHub, ExitGitHub},
		{"timeout error", ErrCodeTimeout, ExitTimeout},
		{"not in project", ErrCodeNotInProject, ExitError},
		{"already exists", ErrCodeAlreadyExists, ExitError},
		{"not found", ErrCodeNotFound, ExitError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCLIError(tt.code, "test message")
			exitCode := GetExitCode(err)
			if exitCode != tt.expected {
				t.Errorf("expected exit code %d for %s, got %d", tt.expected, tt.code, exitCode)
			}
		})
	}
}

func TestGetExitCode_NilError(t *testing.T) {
	exitCode := GetExitCode(nil)
	if exitCode != ExitSuccess {
		t.Errorf("expected exit code %d for nil error, got %d", ExitSuccess, exitCode)
	}
}

func TestGetExitCode_GenericError(t *testing.T) {
	err := &genericError{msg: "some error"}
	exitCode := GetExitCode(err)
	if exitCode != ExitError {
		t.Errorf("expected exit code %d for generic error, got %d", ExitError, exitCode)
	}
}

// genericError is a simple error for testing non-CLIError cases
type genericError struct {
	msg string
}

func (e *genericError) Error() string {
	return e.msg
}
