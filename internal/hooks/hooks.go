package hooks

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Context provides variables for hook commands
type Context struct {
	Path          string // Full path to the worktree
	Branch        string // Branch name (e.g., feature/auth)
	ProjectRoot   string // Project root (contains .bare/)
	DefaultBranch string // Default branch name (e.g., main)
}

// Run executes hook commands with default timeout (30 seconds)
// Returns a list of warning messages for failed commands
func Run(commands []string, ctx Context) []string {
	return RunWithTimeout(commands, ctx, 30)
}

// RunWithTimeout executes hook commands with specified timeout in seconds
// Returns a list of warning messages for failed commands
func RunWithTimeout(commands []string, ctx Context, timeoutSec int) []string {
	var warnings []string

	for _, cmdStr := range commands {
		cmdStr = expandTemplates(cmdStr, ctx)

		// Create context with timeout
		timeout := time.Duration(timeoutSec) * time.Second
		execCtx, cancel := context.WithTimeout(context.Background(), timeout)

		cmd := exec.CommandContext(execCtx, "sh", "-c", cmdStr)
		cmd.Env = append(os.Environ(), buildEnvVars(ctx)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Set platform-specific process attributes (process group on Unix)
		setPlatformAttrs(cmd)

		// WaitDelay ensures process cleanup even if context is cancelled
		cmd.WaitDelay = 3 * time.Second

		err := cmd.Run()
		cancel()

		if err != nil {
			if execCtx.Err() != nil {
				// Handle both DeadlineExceeded and Canceled
				warnings = append(warnings, fmt.Sprintf("%s: %v", cmdStr, execCtx.Err()))
			} else {
				warnings = append(warnings, cmdStr+": "+err.Error())
			}
		}
	}

	return warnings
}

// buildEnvVars creates environment variables from context
func buildEnvVars(ctx Context) []string {
	return []string{
		"GIT_WT_PATH=" + ctx.Path,
		"GIT_WT_BRANCH=" + ctx.Branch,
		"GIT_WT_PROJECT_ROOT=" + ctx.ProjectRoot,
		"GIT_WT_DEFAULT_BRANCH=" + ctx.DefaultBranch,
	}
}

// expandTemplates replaces {{.Field}} with shell-quoted values from context
func expandTemplates(s string, ctx Context) string {
	replacements := map[string]string{
		"{{.Path}}":          shellQuote(ctx.Path),
		"{{.Branch}}":        shellQuote(ctx.Branch),
		"{{.ProjectRoot}}":   shellQuote(ctx.ProjectRoot),
		"{{.DefaultBranch}}": shellQuote(ctx.DefaultBranch),
	}

	for placeholder, value := range replacements {
		s = strings.ReplaceAll(s, placeholder, value)
	}

	return s
}

// shellQuote escapes a string for safe use in shell commands
// Uses single quotes with escaped single quotes for safety
func shellQuote(s string) string {
	// Replace single quotes with '\'' (end quote, escaped quote, start quote)
	escaped := strings.ReplaceAll(s, "'", "'\\''")
	return "'" + escaped + "'"
}
