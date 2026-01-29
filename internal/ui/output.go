package ui

import (
	"encoding/json"
	"io"
)

// Error code constants for structured error responses
const (
	ErrCodeValidation    = "validation_error"
	ErrCodeGit           = "git_error"
	ErrCodeGitHub        = "github_error"
	ErrCodeTimeout       = "timeout_error"
	ErrCodeNotInProject  = "not_in_project"
	ErrCodeAlreadyExists = "already_exists"
	ErrCodeNotFound      = "not_found"
)

// Exit code constants for CLI exit status
const (
	ExitSuccess    = 0
	ExitError      = 1
	ExitValidation = 2
	ExitGit        = 3
	ExitGitHub     = 4
	ExitTimeout    = 124
)

// CLIError represents a structured error with code, message, and exit status
type CLIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Exit    int    `json:"-"`
}

// Error implements the error interface
func (e *CLIError) Error() string {
	return e.Message
}

// NewCLIError creates a new CLIError with the appropriate exit code
func NewCLIError(code, message string) *CLIError {
	exit := ExitError
	switch code {
	case ErrCodeValidation:
		exit = ExitValidation
	case ErrCodeGit:
		exit = ExitGit
	case ErrCodeGitHub:
		exit = ExitGitHub
	case ErrCodeTimeout:
		exit = ExitTimeout
	}
	return &CLIError{
		Code:    code,
		Message: message,
		Exit:    exit,
	}
}

// Response is the envelope for all JSON output
type Response struct {
	Success bool        `json:"success"`
	Command string      `json:"command"`
	Data    interface{} `json:"data,omitempty"`
	Error   *CLIError   `json:"error,omitempty"`
}

// OutputJSON writes a JSON response to the writer
func OutputJSON(w io.Writer, command string, data interface{}, err error) error {
	resp := Response{
		Command: command,
	}

	if err != nil {
		resp.Success = false
		if cliErr, ok := err.(*CLIError); ok {
			resp.Error = cliErr
		} else {
			resp.Error = &CLIError{
				Code:    ErrCodeGit,
				Message: err.Error(),
			}
		}
	} else {
		resp.Success = true
		resp.Data = data
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(resp)
}

// GetExitCode returns the appropriate exit code for an error
func GetExitCode(err error) int {
	if err == nil {
		return ExitSuccess
	}
	if cliErr, ok := err.(*CLIError); ok {
		return cliErr.Exit
	}
	return ExitError
}
