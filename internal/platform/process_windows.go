//go:build windows

package platform

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"
)

// Windows-specific API declarations (additional to those in procinfo_windows.go)
var (
	procGetExitCodeProcess = kernel32.NewProc("GetExitCodeProcess")
)

const (
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	STILL_ACTIVE                      = 259
)

// setProcAttributes sets Windows-specific process attributes
func setProcAttributes(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// isRunningWindows checks if a process is running on Windows using native Win32 API
func isRunningWindows(pid int) bool {
	fmt.Printf("[ProcessManager] IsRunning check for PID %d\n", pid)

	// Open the process with query permission
	handle, _, err := procOpenProcess.Call(
		uintptr(PROCESS_QUERY_LIMITED_INFORMATION),
		uintptr(0),
		uintptr(pid),
	)

	if handle == 0 {
		fmt.Printf("[ProcessManager] OpenProcess failed for PID %d: %v\n", pid, err)
		return false
	}
	defer procCloseHandle.Call(handle)

	// Get the exit code
	var exitCode uint32
	ret, _, err := procGetExitCodeProcess.Call(handle, uintptr(unsafe.Pointer(&exitCode)))
	if ret == 0 {
		fmt.Printf("[ProcessManager] GetExitCodeProcess failed for PID %d: %v\n", pid, err)
		return false
	}

	// If exit code is STILL_ACTIVE, the process is still running
	isRunning := exitCode == STILL_ACTIVE
	fmt.Printf("[ProcessManager] PID %d is running: %v (exitCode=%d)\n", pid, isRunning, exitCode)
	return isRunning
}
