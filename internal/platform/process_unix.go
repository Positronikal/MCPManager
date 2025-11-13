//go:build !windows

package platform

import (
	"os/exec"
	"syscall"
)

// setProcAttributes sets Unix-specific process attributes
func setProcAttributes(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}
