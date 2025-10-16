//go:build windows

package platform

import (
	"os/exec"
	"syscall"
)

// setProcAttributes sets Windows-specific process attributes
func setProcAttributes(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
