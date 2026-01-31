//go:build windows

package hooks

import "os/exec"

// setPlatformAttrs is a no-op on Windows
// Windows doesn't support Unix-style process groups
func setPlatformAttrs(cmd *exec.Cmd) {
	// No-op on Windows
}
