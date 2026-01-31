package git

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Default timeout for git operations
const (
	DefaultTimeout = 2 * time.Minute
	LongTimeout    = 10 * time.Minute // For clone/fetch operations
)

// Run executes a git command and returns the output
func Run(args ...string) (string, error) {
	return RunInDir("", args...)
}

// RunInDir executes a git command in a specific directory
func RunInDir(dir string, args ...string) (string, error) {
	return RunInDirWithTimeout(dir, int(DefaultTimeout.Seconds()), args...)
}

// RunInDirWithTimeout executes a git command with specified timeout in seconds
func RunInDirWithTimeout(dir string, timeoutSec int, args ...string) (string, error) {
	timeout := time.Duration(timeoutSec) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return RunInDirWithContext(ctx, dir, args...)
}

// RunInDirWithContext executes a git command with context for cancellation/timeout
func RunInDirWithContext(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}

	// WaitDelay ensures pipes are drained even if context is cancelled
	cmd.WaitDelay = 5 * time.Second

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			// Handle both DeadlineExceeded and Canceled
			return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), ctx.Err())
		}
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// RunWithLongTimeout executes a git command with extended timeout (for clone/fetch)
func RunWithLongTimeout(dir string, args ...string) (string, error) {
	return RunInDirWithTimeout(dir, int(LongTimeout.Seconds()), args...)
}

// RunWithProgress executes a git command showing progress to the user (for clone/fetch)
func RunWithProgress(dir string, args ...string) error {
	return RunWithProgressAndTimeout(dir, int(LongTimeout.Seconds()), args...)
}

// RunWithProgressAndTimeout executes a git command with progress output and specified timeout
func RunWithProgressAndTimeout(dir string, timeoutSec int, args ...string) error {
	timeout := time.Duration(timeoutSec) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}

	// WaitDelay ensures process cleanup even if context is cancelled
	cmd.WaitDelay = 5 * time.Second

	// Connect to terminal for progress display
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			// Handle both DeadlineExceeded and Canceled
			return fmt.Errorf("git %s: %w", strings.Join(args, " "), ctx.Err())
		}
		return fmt.Errorf("git %s failed", strings.Join(args, " "))
	}

	return nil
}

// RunSilent executes a git command without capturing output
func RunSilent(args ...string) error {
	_, err := Run(args...)
	return err
}
