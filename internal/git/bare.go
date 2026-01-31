package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Constants for bare repo structure
const (
	BareDir        = ".bare"
	GitPointerFile = ".git"
	DefaultBranch  = "main"
	FallbackBranch = "master"
)

// BareClone clones a repository as a bare repo into the specified directory
// Extra arguments are passed directly to git clone
func BareClone(url, targetDir string, extraArgs ...string) error {
	return BareCloneWithTimeout(url, targetDir, int(LongTimeout.Seconds()), extraArgs...)
}

// BareCloneWithTimeout clones a repository with a specified timeout in seconds
func BareCloneWithTimeout(url, targetDir string, timeoutSec int, extraArgs ...string) error {
	bareDir := filepath.Join(targetDir, BareDir)

	// Build clone args: clone --bare --progress [extraArgs...] url bareDir
	args := []string{"clone", "--bare", "--progress"}
	args = append(args, extraArgs...)
	args = append(args, url, bareDir)

	// Clone as bare with progress shown to user
	if err := RunWithProgressAndTimeout(targetDir, timeoutSec, args...); err != nil {
		return fmt.Errorf("failed to clone: %w", err)
	}

	// Create .git file pointing to .bare
	gitFile := filepath.Join(targetDir, GitPointerFile)
	if err := os.WriteFile(gitFile, []byte(fmt.Sprintf("gitdir: ./%s\n", BareDir)), 0644); err != nil {
		return fmt.Errorf("failed to create .git file: %w", err)
	}

	// Configure fetch to get all remote branches
	if _, err := RunInDirWithTimeout(bareDir, timeoutSec, "config", "remote.origin.fetch", "+refs/heads/*:refs/remotes/origin/*"); err != nil {
		return fmt.Errorf("failed to configure fetch: %w", err)
	}

	// Fetch to get remote tracking branches
	if _, err := RunInDirWithTimeout(targetDir, timeoutSec, "fetch", "origin"); err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	return nil
}

// IsBareRepo checks if the directory contains a bare repo structure
func IsBareRepo(dir string) bool {
	bareDir := filepath.Join(dir, BareDir)
	gitFile := filepath.Join(dir, GitPointerFile)

	// Check for .bare directory
	if info, err := os.Stat(bareDir); err != nil || !info.IsDir() {
		return false
	}

	// Check for .git file (not directory)
	info, err := os.Stat(gitFile)
	if err != nil {
		return false
	}

	return !info.IsDir()
}

// GetProjectRoot finds the project root (containing .bare) from a worktree
func GetProjectRoot(worktreePath string) (string, error) {
	// Resolve to absolute path first
	absPath, err := filepath.Abs(worktreePath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}
	dir := absPath

	for {
		if IsBareRepo(dir) {
			return dir, nil
		}

		// Move up to parent
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}

		dir = parent
	}

	return "", fmt.Errorf("not in a git-wt project")
}

// GetDefaultBranch returns the default branch name (main or master)
func GetDefaultBranch(dir string) (string, error) {
	// Try to get from remote HEAD
	output, err := RunInDir(dir, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		// Extract branch name from refs/remotes/origin/main
		parts := strings.Split(strings.TrimSpace(output), "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}

	// Fallback: check if main or master exists
	if _, err := RunInDir(dir, "rev-parse", "--verify", "refs/remotes/origin/"+DefaultBranch); err == nil {
		return DefaultBranch, nil
	}

	if _, err := RunInDir(dir, "rev-parse", "--verify", "refs/remotes/origin/"+FallbackBranch); err == nil {
		return FallbackBranch, nil
	}

	// Return error if neither exists
	return "", fmt.Errorf("could not determine default branch: neither %s nor %s found", DefaultBranch, FallbackBranch)
}
