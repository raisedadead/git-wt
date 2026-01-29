package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsBareRepo(t *testing.T) {
	// Test non-existent directory
	if IsBareRepo("/nonexistent") {
		t.Error("expected false for non-existent directory")
	}

	// Test empty directory
	tmpDir := t.TempDir()
	if IsBareRepo(tmpDir) {
		t.Error("expected false for empty directory")
	}
}

func TestGetProjectRoot(t *testing.T) {
	// Create a mock bare repo structure
	tmpDir := t.TempDir()
	bareDir := filepath.Join(tmpDir, ".bare")
	if err := os.MkdirAll(bareDir, 0755); err != nil {
		t.Fatal(err)
	}

	gitFile := filepath.Join(tmpDir, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: ./.bare"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a worktree directory
	worktreeDir := filepath.Join(tmpDir, "main")
	if err := os.MkdirAll(worktreeDir, 0755); err != nil {
		t.Fatal(err)
	}

	root, err := GetProjectRoot(worktreeDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if root != tmpDir {
		t.Errorf("expected %s, got %s", tmpDir, root)
	}
}
