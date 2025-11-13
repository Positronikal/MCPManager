package platform

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
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

// StartWithOutput launches a new process and returns stdout/stderr readers
func (pm *DefaultProcessManager) StartWithOutput(cmd string, args []string, env map[string]string) (int, io.ReadCloser, io.ReadCloser, error) {
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

	// Create pipes for stdout and stderr
	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrPipe, err := command.StderrPipe()
	if err != nil {
		return 0, nil, nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Platform-specific process group settings
	setProcAttributes(command)

	// Start the process
	if err := command.Start(); err != nil {
		return 0, nil, nil, fmt.Errorf("failed to start process: %w", err)
	}

	// Return PID and pipes (StdoutPipe and StderrPipe already return io.ReadCloser)
	return command.Process.Pid, stdoutPipe, stderrPipe, nil
}

// Stop terminates a process by its ID
func (pm *DefaultProcessManager) Stop(pid int, graceful bool, timeout int) error {
	fmt.Printf("[ProcessManager] Stop called: pid=%d, graceful=%v, timeout=%d\n", pid, graceful, timeout)

	// Find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		fmt.Printf("[ProcessManager] os.FindProcess failed: pid=%d, error=%v\n", pid, err)
		return fmt.Errorf("process not found: %w", err)
	}
	fmt.Printf("[ProcessManager] Process found: pid=%d\n", pid)

	if graceful {
		fmt.Printf("[ProcessManager] Attempting graceful shutdown: pid=%d\n", pid)
		// Try graceful shutdown first
		if err := pm.sendTermSignal(process); err != nil {
			fmt.Printf("[ProcessManager] Graceful termination signal failed: pid=%d, error=%v\n", pid, err)
			// If signal failed, process might already be dead
			if !pm.IsRunning(pid) {
				fmt.Printf("[ProcessManager] Process already dead after signal failure: pid=%d\n", pid)
				return nil
			}
			// Signal failed but process still running, fall through to force kill
		} else {
			fmt.Printf("[ProcessManager] Graceful termination signal sent: pid=%d, waiting for exit...\n", pid)

			// Wait for process to exit
			done := make(chan bool)
			go func() {
				// Poll for process exit
				for i := 0; i < timeout*10; i++ {
					if !pm.IsRunning(pid) {
						fmt.Printf("[ProcessManager] Process exited gracefully: pid=%d\n", pid)
						done <- true
						return
					}
					time.Sleep(100 * time.Millisecond)
				}
				fmt.Printf("[ProcessManager] Graceful shutdown timed out: pid=%d\n", pid)
				done <- false
			}()

			if <-done {
				// Process exited gracefully
				return nil
			}

			// Graceful shutdown timed out, fall through to force kill
			fmt.Printf("[ProcessManager] Proceeding to force kill: pid=%d\n", pid)
		}
	}

	// Force kill
	fmt.Printf("[ProcessManager] Attempting force kill: pid=%d\n", pid)
	if err := process.Kill(); err != nil {
		fmt.Printf("[ProcessManager] Force kill failed: pid=%d, error=%v\n", pid, err)
		// Check if process is actually dead
		if !pm.IsRunning(pid) {
			fmt.Printf("[ProcessManager] Process is dead despite kill error: pid=%d\n", pid)
			return nil
		}
		return fmt.Errorf("failed to kill process %d: %w", pid, err)
	}

	fmt.Printf("[ProcessManager] Force kill succeeded: pid=%d\n", pid)
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

	// On Windows, we need to use platform-specific process checking
	if runtime.GOOS == "windows" {
		// Use Windows-specific process check
		return isRunningWindows(pid)
	}

	// On Unix, sending signal 0 checks if process exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
