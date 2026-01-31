//go:build unix

package hooks

import (
	"os/exec"
	"syscall"
)

// setPlatformAttrs sets Unix-specific process attributes
// Creates a new process group so child processes can be killed on timeout
func setPlatformAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
