package hooks

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/raisedadead/git-wt/internal/hooks/bundled"
)

// Context provides variables for hook commands
type Context struct {
	Path          string // Full path to the worktree
	Branch        string // Branch name (e.g., feature/auth)
	ProjectRoot   string // Project root (contains .bare/)
	DefaultBranch string // Default branch name (e.g., main)

	// Workflow-related fields
	Workflow       string            // Workflow name (feature, bugfix, pr-review, branch)
	WorkflowPrefix string            // Branch prefix from workflow template
	IssueNumber    int               // GitHub issue number (if provided)
	PRNumber       int               // GitHub PR number (if provided)
	Metadata       map[string]string // Arbitrary key-value metadata from hooks
}

// HookOutput represents the parsed output from a hook
type HookOutput struct {
	Branch   string            // Suggested branch name
	Metadata map[string]string // Key-value metadata
	Prompts  map[string]string // Prompts with default values
	Error    string            // Error message if hook failed
	Warnings []string          // Warning messages
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
	vars := []string{
		"GIT_WT_PATH=" + ctx.Path,
		"GIT_WT_BRANCH=" + ctx.Branch,
		"GIT_WT_PROJECT_ROOT=" + ctx.ProjectRoot,
		"GIT_WT_DEFAULT_BRANCH=" + ctx.DefaultBranch,
	}

	// Add workflow-related vars
	if ctx.Workflow != "" {
		vars = append(vars, "GIT_WT_WORKFLOW="+ctx.Workflow)
	}
	if ctx.WorkflowPrefix != "" {
		vars = append(vars, "GIT_WT_WORKFLOW_PREFIX="+ctx.WorkflowPrefix)
	}
	if ctx.IssueNumber > 0 {
		vars = append(vars, fmt.Sprintf("GIT_WT_ISSUE=%d", ctx.IssueNumber))
	}
	if ctx.PRNumber > 0 {
		vars = append(vars, fmt.Sprintf("GIT_WT_PR=%d", ctx.PRNumber))
	}

	// Add any metadata
	for k, v := range ctx.Metadata {
		vars = append(vars, fmt.Sprintf("GIT_WT_META_%s=%s", strings.ToUpper(k), v))
	}

	return vars
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

// RunResolved executes hooks with name resolution and helper library
// Resolves hook names from custom -> community -> inline
// Provides GIT_WT_LIB so hooks can use the helper library
func RunResolved(hookNames []string, ctx Context, timeoutSec int, customDir, communityDir string) []string {
	if len(hookNames) == 0 {
		return nil
	}

	// Create temp directory for helpers library
	libDir, err := os.MkdirTemp("", "git-wt-lib-*")
	if err != nil {
		return []string{fmt.Sprintf("failed to create lib directory: %v", err)}
	}
	defer func() { _ = os.RemoveAll(libDir) }()

	// Write helpers.sh to lib directory
	helpersContent, err := bundled.GetHelpers()
	if err != nil {
		return []string{fmt.Sprintf("failed to get helpers: %v", err)}
	}
	helpersPath := filepath.Join(libDir, "helpers.sh")
	if err := os.WriteFile(helpersPath, helpersContent, 0644); err != nil {
		return []string{fmt.Sprintf("failed to write helpers: %v", err)}
	}

	var warnings []string

	for _, name := range hookNames {
		path, _ := ResolveHook(name, customDir, communityDir)
		cmdStr := expandTemplates(path, ctx)

		// Create context with timeout
		timeout := time.Duration(timeoutSec) * time.Second
		execCtx, cancel := context.WithTimeout(context.Background(), timeout)

		cmd := exec.CommandContext(execCtx, "sh", "-c", cmdStr)
		env := append(os.Environ(), buildEnvVars(ctx)...)
		env = append(env, "GIT_WT_LIB="+libDir)
		cmd.Env = env
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		// Set platform-specific process attributes (process group on Unix)
		setPlatformAttrs(cmd)
		cmd.WaitDelay = 3 * time.Second

		err := cmd.Run()
		cancel()

		if err != nil {
			if execCtx.Err() != nil {
				warnings = append(warnings, fmt.Sprintf("%s: %v", name, execCtx.Err()))
			} else {
				warnings = append(warnings, name+": "+err.Error())
			}
		}
	}

	return warnings
}

// RunWorkflowHook executes a single hook and captures its output
// Returns the parsed output including branch suggestion, metadata, and errors
func RunWorkflowHook(hookName string, ctx Context, timeoutSec int, customDir, communityDir string) (*HookOutput, error) {
	// Create temp file for hook output
	outputFile, err := os.CreateTemp("", "git-wt-hook-output-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	outputPath := outputFile.Name()
	_ = outputFile.Close()
	defer func() { _ = os.Remove(outputPath) }()

	// Create temp directory for helpers library
	libDir, err := os.MkdirTemp("", "git-wt-lib-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create lib directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(libDir) }()

	// Write helpers.sh to lib directory
	helpersContent, err := bundled.GetHelpers()
	if err != nil {
		return nil, fmt.Errorf("failed to get helpers: %w", err)
	}
	helpersPath := filepath.Join(libDir, "helpers.sh")
	if err := os.WriteFile(helpersPath, helpersContent, 0644); err != nil {
		return nil, fmt.Errorf("failed to write helpers: %w", err)
	}

	// Resolve hook path
	hookPath, isScript := ResolveHook(hookName, customDir, communityDir)
	if !isScript {
		// For inline commands, we don't support output parsing
		warnings := RunWithTimeout([]string{hookPath}, ctx, timeoutSec)
		return &HookOutput{Warnings: warnings}, nil
	}

	// Build environment with output file and lib paths
	env := append(os.Environ(), buildEnvVars(ctx)...)
	env = append(env, "GIT_WT_OUTPUT="+outputPath)
	env = append(env, "GIT_WT_LIB="+libDir)

	// Create context with timeout
	timeout := time.Duration(timeoutSec) * time.Second
	execCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, "sh", "-c", hookPath)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = ctx.ProjectRoot

	// Set platform-specific process attributes
	setPlatformAttrs(cmd)
	cmd.WaitDelay = 3 * time.Second

	runErr := cmd.Run()

	// Parse output file
	output, parseErr := parseHookOutput(outputPath)
	if parseErr != nil {
		return nil, fmt.Errorf("failed to parse hook output: %w", parseErr)
	}

	// Handle execution errors
	if runErr != nil {
		if execCtx.Err() != nil {
			output.Warnings = append(output.Warnings, fmt.Sprintf("%s: %v", hookName, execCtx.Err()))
		} else if output.Error == "" {
			// Only set error if hook didn't set one
			output.Error = runErr.Error()
		}
	}

	return output, nil
}

// parseHookOutput reads and parses the hook output file
func parseHookOutput(path string) (*HookOutput, error) {
	output := &HookOutput{
		Metadata: make(map[string]string),
		Prompts:  make(map[string]string),
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No output file means hook didn't write anything
			return output, nil
		}
		return nil, err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse key=value pairs
		if idx := strings.Index(line, "="); idx > 0 {
			key := line[:idx]
			value := line[idx+1:]

			switch {
			case key == "GIT_WT_BRANCH":
				output.Branch = value
			case key == "GIT_WT_ERROR":
				output.Error = value
			case key == "GIT_WT_WARNING":
				output.Warnings = append(output.Warnings, value)
			case strings.HasPrefix(key, "GIT_WT_META_"):
				metaKey := strings.TrimPrefix(key, "GIT_WT_META_")
				output.Metadata[strings.ToLower(metaKey)] = value
			case strings.HasPrefix(key, "GIT_WT_PROMPT_"):
				promptKey := strings.TrimPrefix(key, "GIT_WT_PROMPT_")
				output.Prompts[strings.ToLower(promptKey)] = value
			}
		}
	}

	return output, scanner.Err()
}

// RunWorkflowHooks executes multiple hooks in sequence, merging their outputs
func RunWorkflowHooks(hookNames []string, ctx Context, timeoutSec int, customDir, communityDir string) (*HookOutput, error) {
	combined := &HookOutput{
		Metadata: make(map[string]string),
		Prompts:  make(map[string]string),
	}

	for _, hookName := range hookNames {
		output, err := RunWorkflowHook(hookName, ctx, timeoutSec, customDir, communityDir)
		if err != nil {
			return nil, fmt.Errorf("hook %s failed: %w", hookName, err)
		}

		// Check for hook error
		if output.Error != "" {
			combined.Error = output.Error
			return combined, nil
		}

		// Merge outputs (later hooks override earlier ones)
		if output.Branch != "" {
			combined.Branch = output.Branch
		}
		for k, v := range output.Metadata {
			combined.Metadata[k] = v
		}
		for k, v := range output.Prompts {
			combined.Prompts[k] = v
		}
		combined.Warnings = append(combined.Warnings, output.Warnings...)

		// Update context metadata for next hook
		if ctx.Metadata == nil {
			ctx.Metadata = make(map[string]string)
		}
		for k, v := range output.Metadata {
			ctx.Metadata[k] = v
		}
	}

	return combined, nil
}
