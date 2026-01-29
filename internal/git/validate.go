package git

import (
	"fmt"
	"strings"
)

// ValidateProjectName validates a project name for safety
func ValidateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	// Check for path traversal
	if strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return fmt.Errorf("project name cannot contain path separators: %s", name)
	}

	if strings.Contains(name, "..") {
		return fmt.Errorf("project name cannot contain '..': %s", name)
	}

	// Check for current/parent directory references
	if name == "." || name == ".." {
		return fmt.Errorf("invalid project name: %s", name)
	}

	// Check for hidden files that might conflict
	if name == ".git" || name == ".bare" {
		return fmt.Errorf("reserved project name: %s", name)
	}

	return nil
}

// ValidateBranchName validates a git branch name
func ValidateBranchName(name string) error {
	if name == "" {
		return fmt.Errorf("branch name cannot be empty")
	}

	// Git branch name restrictions
	invalidPatterns := []string{
		"..",    // No double dots
		"~",     // No tilde
		"^",     // No caret
		":",     // No colon
		"\\",    // No backslash
		" ",     // No spaces
		"?",     // No question mark
		"*",     // No asterisk
		"[",     // No open bracket
		"@{",    // No @{ sequence
		".lock", // Cannot end with .lock
	}

	for _, pattern := range invalidPatterns {
		if strings.Contains(name, pattern) {
			return fmt.Errorf("branch name contains invalid pattern '%s': %s", pattern, name)
		}
	}

	// Cannot start or end with a dot
	if strings.HasPrefix(name, ".") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("branch name cannot start or end with a dot: %s", name)
	}

	// Cannot start or end with a slash
	if strings.HasPrefix(name, "/") || strings.HasSuffix(name, "/") {
		return fmt.Errorf("branch name cannot start or end with a slash: %s", name)
	}

	// Cannot have consecutive slashes
	if strings.Contains(name, "//") {
		return fmt.Errorf("branch name cannot contain consecutive slashes: %s", name)
	}

	return nil
}

// FlattenBranchName converts branch names with slashes to directory-safe names
// e.g., "feature/auth" becomes "feature-auth"
func FlattenBranchName(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}
