package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	testRepoURL      = "https://github.com/experiments-by-mrugesh/test-repo.git"
	fixtureStaleDays = 3 // Cache fixture in temp dir for 3 days
)

var (
	gitWTBinary string
	fixtureRepo string
	localRemote string // Local bare repo to use as "remote" for clone tests
	projectRoot string
)

func TestMain(m *testing.M) {
	// Find project root (where go.mod is)
	var err error
	projectRoot, err = findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to find project root: %v\n", err)
		os.Exit(1)
	}

	// Build binary
	gitWTBinary, err = buildBinary()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build binary: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Using binary: %s\n", gitWTBinary)

	// Setup fixture
	fixtureRepo, err = setupFixture()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup fixture: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Using fixture: %s\n", fixtureRepo)

	// Setup local remote (bare repo for clone tests)
	localRemote, err = setupLocalRemote()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup local remote: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Using local remote: %s\n", localRemote)

	// Run tests
	code := m.Run()

	os.Exit(code)
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find go.mod")
		}
		dir = parent
	}
}

func buildBinary() (string, error) {
	binaryPath := filepath.Join(projectRoot, "bin", "wt")

	cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/wt")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("build failed: %w", err)
	}

	return binaryPath, nil
}

func setupFixture() (string, error) {
	// Use system temp directory to avoid polluting the project
	tempBase := filepath.Join(os.TempDir(), "wt-test-fixture")
	fixturePath := filepath.Join(tempBase, "fixture")

	// Check if fixture exists and is fresh
	if info, err := os.Stat(fixturePath); err == nil {
		age := time.Since(info.ModTime())
		if age < time.Duration(fixtureStaleDays)*24*time.Hour {
			fmt.Printf("Reusing cached fixture (age: %v)\n", age.Round(time.Minute))
			// Ensure local branches exist even for cached fixture
			setupLocalBranches(fixturePath)
			return fixturePath, nil
		}
		fmt.Printf("Fixture stale (age: %v), re-cloning\n", age.Round(time.Minute))
		_ = os.RemoveAll(fixturePath)
	}

	// Clone fresh
	fmt.Printf("Cloning fixture from %s\n", testRepoURL)
	if err := os.MkdirAll(tempBase, 0755); err != nil {
		return "", err
	}

	cmd := exec.Command("git", "clone", testRepoURL, fixturePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("clone failed: %w", err)
	}

	// Create local branches from remote tracking branches
	setupLocalBranches(fixturePath)

	return fixturePath, nil
}

// setupLocalBranches creates local tracking branches from remote branches
func setupLocalBranches(repoPath string) {
	branches := []string{"feature/auth", "feature/nested/deep", "develop", "bugfix-123"}
	for _, branch := range branches {
		cmd := exec.Command("git", "branch", "--track", branch, "origin/"+branch)
		cmd.Dir = repoPath
		cmd.Env = filterGitEnv(os.Environ())
		_ = cmd.Run() // Ignore errors - branch may already exist
	}
}

func setupLocalRemote() (string, error) {
	// Use system temp directory alongside the fixture
	tempBase := filepath.Join(os.TempDir(), "wt-test-fixture")
	remotePath := filepath.Join(tempBase, "remote.git")

	// Always recreate the local remote from fixture
	_ = os.RemoveAll(remotePath)

	// Filter git env vars to avoid interference when running inside git hooks
	cleanEnv := filterGitEnv(os.Environ())

	// Prune any stale worktree references from fixture first
	pruneCmd := exec.Command("git", "worktree", "prune")
	pruneCmd.Dir = fixtureRepo
	pruneCmd.Env = cleanEnv
	_ = pruneCmd.Run()

	// Create bare clone from fixture
	fmt.Printf("Creating local bare remote from fixture\n")
	cmd := exec.Command("git", "clone", "--bare", fixtureRepo, remotePath)
	cmd.Env = cleanEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create local remote: %w", err)
	}

	// Push only the specific branches needed for tests to the bare remote
	// git clone --bare only copies default branch, not all local branches
	// Local branches were set up in setupFixture() via setupLocalBranches()
	fmt.Printf("Pushing test branches to local remote\n")
	testBranches := []string{"main", "feature/auth", "feature/nested/deep", "develop", "bugfix-123"}
	pushArgs := append([]string{"push", "--force", remotePath}, testBranches...)
	pushCmd := exec.Command("git", pushArgs...)
	pushCmd.Dir = fixtureRepo
	pushCmd.Env = cleanEnv
	pushCmd.Stdout = os.Stdout
	pushCmd.Stderr = os.Stderr
	if err := pushCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to push branches to local remote: %w", err)
	}

	// Prune worktree refs from the bare remote (inherited from fixture)
	pruneRemote := exec.Command("git", "worktree", "prune")
	pruneRemote.Dir = remotePath
	pruneRemote.Env = cleanEnv
	_ = pruneRemote.Run()

	return remotePath, nil
}

// setupTestWorkspace creates an isolated workspace for a test.
// Returns the workspace directory (use t.TempDir() for auto-cleanup).
func setupTestWorkspace(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// filterGitEnv removes git-related environment variables to ensure test isolation.
// This is necessary when tests run inside git hooks (e.g., pre-push) which set
// GIT_DIR, GIT_WORK_TREE, etc. that would affect the test's git operations.
func filterGitEnv(env []string) []string {
	gitVars := map[string]bool{
		"GIT_DIR":                          true,
		"GIT_WORK_TREE":                    true,
		"GIT_INDEX_FILE":                   true,
		"GIT_OBJECT_DIRECTORY":             true,
		"GIT_ALTERNATE_OBJECT_DIRECTORIES": true,
		"GIT_QUARANTINE_PATH":              true,
	}

	var filtered []string
	for _, e := range env {
		key := strings.SplitN(e, "=", 2)[0]
		if !gitVars[key] {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// runGitWT executes wt with the given arguments and returns stdout, stderr, and exit code
func runGitWT(t *testing.T, dir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(gitWTBinary, args...)
	cmd.Dir = dir

	// Clear git environment variables that might be inherited from parent processes
	// (e.g., when running in a git hook). This ensures tests run in isolation.
	cmd.Env = filterGitEnv(os.Environ())

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	stdout = stdoutBuf.String()
	stderr = stderrBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return stdout, stderr, exitCode
}

// runGitWTSuccess runs wt and fails the test if it doesn't succeed
func runGitWTSuccess(t *testing.T, dir string, args ...string) string {
	t.Helper()
	stdout, stderr, code := runGitWT(t, dir, args...)
	if code != 0 {
		t.Fatalf("wt %v failed (exit %d):\nstdout: %s\nstderr: %s", args, code, stdout, stderr)
	}
	return stdout
}

// runGitWTFail runs wt and fails the test if it succeeds
func runGitWTFail(t *testing.T, dir string, args ...string) (stdout, stderr string) {
	t.Helper()
	stdout, stderr, code := runGitWT(t, dir, args...)
	if code == 0 {
		t.Fatalf("wt %v should have failed but succeeded:\nstdout: %s", args, stdout)
	}
	return stdout, stderr
}

// runGitWTJSON runs wt with --json and parses the result
func runGitWTJSON(t *testing.T, dir string, args ...string) map[string]any {
	t.Helper()

	args = append(args, "--json")
	stdout, stderr, code := runGitWT(t, dir, args...)
	if code != 0 {
		t.Fatalf("wt %v --json failed (exit %d):\nstdout: %s\nstderr: %s", args, code, stdout, stderr)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, stdout)
	}

	return result
}

// runGit runs a git command in the given directory
func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	// Set git identity for commit operations in CI environments
	cmd.Env = append(filterGitEnv(os.Environ()),
		"GIT_AUTHOR_NAME=Test",
		"GIT_AUTHOR_EMAIL=test@test.com",
		"GIT_COMMITTER_NAME=Test",
		"GIT_COMMITTER_EMAIL=test@test.com",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, output)
	}

	return strings.TrimSpace(string(output))
}

// assertFileExists fails the test if the file doesn't exist
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Expected file to exist: %s", path)
	}
}

// assertDirExists fails the test if the directory doesn't exist
func assertDirExists(t *testing.T, path string) {
	t.Helper()
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		t.Errorf("Expected directory to exist: %s", path)
	} else if err == nil && !info.IsDir() {
		t.Errorf("Expected directory but found file: %s", path)
	}
}

// assertDirNotExists fails the test if the directory exists
func assertDirNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("Expected directory to not exist: %s", path)
	}
}

// assertContains fails if substr is not in s
func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("Expected %q to contain %q", s, substr)
	}
}

// assertNotContains fails if substr is in s
func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("Expected %q to not contain %q", s, substr)
	}
}
