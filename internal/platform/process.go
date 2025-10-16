package platform

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// DefaultProcessManager implements ProcessManager for the current platform
type DefaultProcessManager struct{}

// NewProcessManager creates a new ProcessManager for the current platform
func NewProcessManager() ProcessManager {
	return &DefaultProcessManager{}
}

// Start launches a new process with the given command, arguments, and environment
func (pm *DefaultProcessManager) Start(cmd string, args []string, env map[string]string) (int, error) {
	// Create the command
	command := exec.Command(cmd, args...)

	// Set environment variables
	if env != nil {
		// Start with current environment
		command.Env = os.Environ()

		// Add/override with provided env vars
		for key, value := range env {
			command.Env = append(command.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Platform-specific process group settings
	setProcAttributes(command)

	// Start the process
	if err := command.Start(); err != nil {
		return 0, fmt.Errorf("failed to start process: %w", err)
	}

	// Return the PID
	return command.Process.Pid, nil
}

// Stop terminates a process by its ID
func (pm *DefaultProcessManager) Stop(pid int, graceful bool, timeout int) error {
	// Find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	if graceful {
		// Try graceful shutdown first
		if err := pm.sendTermSignal(process); err != nil {
			// Process might already be dead, which is fine
			return nil
		}

		// Wait for process to exit
		done := make(chan bool)
		go func() {
			// Poll for process exit
			for i := 0; i < timeout*10; i++ {
				if !pm.IsRunning(pid) {
					done <- true
					return
				}
				time.Sleep(100 * time.Millisecond)
			}
			done <- false
		}()

		if <-done {
			// Process exited gracefully
			return nil
		}

		// Graceful shutdown timed out, fall through to force kill
	}

	// Force kill
	if err := process.Kill(); err != nil {
		// Process might already be dead
		return nil
	}

	return nil
}

// sendTermSignal sends a termination signal to the process
func (pm *DefaultProcessManager) sendTermSignal(process *os.Process) error {
	if runtime.GOOS == "windows" {
		// Windows doesn't support SIGTERM, use Kill
		return process.Kill()
	}
	// Unix systems: send SIGTERM
	return process.Signal(syscall.SIGTERM)
}

// IsRunning checks if a process with the given ID is currently running
func (pm *DefaultProcessManager) IsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Try to send signal 0 (no-op) to check if process exists
	if runtime.GOOS == "windows" {
		// On Windows, os.FindProcess always succeeds, so we need a different check
		// Signal(0) doesn't work on Windows, but we can check by trying to kill
		// and seeing if we get "process already finished" error
		err = process.Signal(os.Kill)
		if err == nil {
			// Successfully sent signal, process exists
			// Note: We're not actually killing it because this is just a check
			return true
		}
		// Check if error is "process already finished"
		return !strings.Contains(err.Error(), "finished")
	}

	// On Unix, sending signal 0 checks if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
