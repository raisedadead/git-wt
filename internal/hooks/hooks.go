package hooks

import (
	"os"
	"os/exec"
	"strings"
)

// Context provides variables for hook commands
type Context struct {
	Path          string // Full path to the worktree
	Branch        string // Branch name (e.g., feature/auth)
	ProjectRoot   string // Project root (contains .bare/)
	DefaultBranch string // Default branch name (e.g., main)
}

// Run executes hook commands with the given context
// Returns a list of warning messages for failed commands
func Run(commands []string, ctx Context) []string {
	var warnings []string

	for _, cmdStr := range commands {
		// Expand template variables
		cmdStr = expandTemplates(cmdStr, ctx)

		// Execute via shell
		cmd := exec.Command("sh", "-c", cmdStr)
		cmd.Env = append(os.Environ(), buildEnvVars(ctx)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			warnings = append(warnings, cmdStr+": "+err.Error())
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

// expandTemplates replaces {{.Field}} with values from context
func expandTemplates(s string, ctx Context) string {
	replacements := map[string]string{
		"{{.Path}}":          ctx.Path,
		"{{.Branch}}":        ctx.Branch,
		"{{.ProjectRoot}}":   ctx.ProjectRoot,
		"{{.DefaultBranch}}": ctx.DefaultBranch,
	}

	for placeholder, value := range replacements {
		s = strings.ReplaceAll(s, placeholder, value)
	}

	return s
}
