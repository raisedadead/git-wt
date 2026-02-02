package hooks

import (
	"os"
	"path/filepath"
	"strings"
)

// ResolvedHook represents a resolved hook
type ResolvedHook struct {
	Name     string
	Path     string // Script path or inline command
	IsScript bool   // true if Path is a script file
}

// isValidHookName checks if a hook name is safe (no path traversal)
func isValidHookName(name string) bool {
	// Reject path separators and traversal sequences
	if strings.Contains(name, "/") || strings.Contains(name, "\\") || strings.Contains(name, "..") {
		return false
	}
	// Reject empty names
	if name == "" {
		return false
	}
	return true
}

// ResolveHook resolves a hook name to a script path or inline command.
// Resolution order: customDir -> communityDir -> inline command
// Returns (path_or_command, isScript)
func ResolveHook(name, customDir, communityDir string) (string, bool) {
	// Validate hook name to prevent path traversal attacks
	// If invalid, treat as inline command (won't match any script file)
	if !isValidHookName(name) {
		return name, false
	}

	// Try custom directory first
	if customDir != "" {
		customPath := filepath.Join(customDir, name+".sh")
		if _, err := os.Stat(customPath); err == nil {
			return customPath, true
		}
	}

	// Try community directory
	if communityDir != "" {
		communityPath := filepath.Join(communityDir, name+".sh")
		if _, err := os.Stat(communityPath); err == nil {
			return communityPath, true
		}
	}

	// Fall back to inline command
	return name, false
}

// ResolveHooks resolves a list of hook names
func ResolveHooks(names []string, customDir, communityDir string) []ResolvedHook {
	var resolved []ResolvedHook
	for _, name := range names {
		path, isScript := ResolveHook(name, customDir, communityDir)
		resolved = append(resolved, ResolvedHook{
			Name:     name,
			Path:     path,
			IsScript: isScript,
		})
	}
	return resolved
}
