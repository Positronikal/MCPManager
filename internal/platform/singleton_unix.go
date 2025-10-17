//go:build darwin || linux

package platform

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

// UnixSingleInstance implements SingleInstance for Unix-like systems using Unix domain socket
type UnixSingleInstance struct {
	socketPath string
	listener   net.Listener
	appName    string
}

// NewSingleInstance creates a new Unix single instance enforcer
func NewSingleInstance(appName string, windowTitle string) SingleInstance {
	// Use /tmp for the lock file (standard location for temporary files)
	// windowTitle is not used on Unix (X11/Wayland focus management is more complex)
	socketPath := filepath.Join("/tmp", fmt.Sprintf("%s.lock", appName))

	return &UnixSingleInstance{
		socketPath: socketPath,
		appName:    appName,
	}
}

// Acquire attempts to acquire the single instance lock
func (u *UnixSingleInstance) Acquire() (bool, error) {
	// Try to connect to existing socket
	conn, err := net.Dial("unix", u.socketPath)
	if err == nil {
		// Successfully connected - another instance is running
		conn.Close()

		// Try to signal the existing instance to show its window
		u.signalExistingInstance()

		return false, nil
	}

	// Socket doesn't exist or connection failed - try to create it
	// First, clean up any stale socket file
	if err := u.cleanupStaleSocket(); err != nil {
		return false, fmt.Errorf("failed to cleanup stale socket: %w", err)
	}

	// Create the Unix domain socket listener
	listener, err := net.Listen("unix", u.socketPath)
	if err != nil {
		return false, fmt.Errorf("failed to create socket listener: %w", err)
	}

	u.listener = listener

	// Start a goroutine to accept connections (for future instance detection)
	go u.acceptConnections()

	return true, nil
}

// Release releases the single instance lock
func (u *UnixSingleInstance) Release() error {
	if u.listener != nil {
		u.listener.Close()
		u.listener = nil
	}

	// Remove the socket file
	if err := os.Remove(u.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove socket file: %w", err)
	}

	return nil
}

// cleanupStaleSocket removes a socket file if the process that created it is no longer running
func (u *UnixSingleInstance) cleanupStaleSocket() error {
	// Check if socket file exists
	if _, err := os.Stat(u.socketPath); os.IsNotExist(err) {
		return nil // No socket file, nothing to clean up
	}

	// Try to connect to the socket
	conn, err := net.Dial("unix", u.socketPath)
	if err == nil {
		// Connection successful - another instance is actually running
		conn.Close()
		return fmt.Errorf("another instance is running")
	}

	// Connection failed - socket is stale, remove it
	if err := os.Remove(u.socketPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove stale socket: %w", err)
	}

	return nil
}

// acceptConnections accepts incoming connections from new instances trying to start
func (u *UnixSingleInstance) acceptConnections() {
	for {
		conn, err := u.listener.Accept()
		if err != nil {
			// Listener closed or error occurred
			return
		}

		// Close the connection immediately - we just needed to detect the attempt
		conn.Close()

		// Optionally, trigger window focus here
		// This is complex on Unix (requires X11/Wayland APIs)
		// For now, we just detect the duplicate instance
	}
}

// signalExistingInstance attempts to signal the running instance to show its window
func (u *UnixSingleInstance) signalExistingInstance() {
	// Try to read PID from a lock file if we create one in the future
	// For now, we just attempt the connection which the existing instance will detect

	// On X11 systems, you could use wmctrl or xdotool to bring window to front
	// On Wayland, this is even more restricted
	// Since this requires external dependencies, we keep it simple for now

	// The existing instance will detect the connection attempt in acceptConnections()
	// and could implement window focus logic there if needed
}

// getPIDFromSocket attempts to get the PID of the process holding the socket
func (u *UnixSingleInstance) getPIDFromSocket() (int, error) {
	// This is a simplified version - in a real implementation,
	// we might write the PID to a separate lock file

	// Try to read /proc to find the process using the socket
	// This is Linux-specific and complex, so we skip it for now

	return 0, fmt.Errorf("PID detection not implemented")
}

// sendSignalToPID sends a signal to the process to bring its window to front
func (u *UnixSingleInstance) sendSignalToPID(pid int) error {
	// Send SIGUSR1 to the existing instance
	// The application would need to handle this signal to bring its window to front
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// Note: syscall.SIGUSR1 is platform-specific
	// On Unix systems, SIGUSR1 = 10 (Linux/macOS)
	if err := process.Signal(syscall.Signal(10)); err != nil {
		return fmt.Errorf("failed to send signal: %w", err)
	}

	return nil
}

// Helper function to write PID to a lock file (optional enhancement)
func (u *UnixSingleInstance) writePIDFile() error {
	pidFile := u.socketPath + ".pid"
	pid := os.Getpid()

	return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}

// Helper function to read PID from a lock file (optional enhancement)
func (u *UnixSingleInstance) readPIDFile() (int, error) {
	pidFile := u.socketPath + ".pid"

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("invalid PID in lock file: %w", err)
	}

	return pid, nil
}
