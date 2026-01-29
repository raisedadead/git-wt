package git

import (
	"testing"
)

func TestRun_GitVersion(t *testing.T) {
	output, err := Run("version")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(output) == 0 {
		t.Fatal("expected output, got empty string")
	}
}

func TestRun_InvalidCommand(t *testing.T) {
	_, err := Run("not-a-real-command")
	if err == nil {
		t.Fatal("expected error for invalid command")
	}
}

func TestRunInDir_GitVersion(t *testing.T) {
	output, err := RunInDir(".", "version")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(output) == 0 {
		t.Fatal("expected output, got empty string")
	}
}
