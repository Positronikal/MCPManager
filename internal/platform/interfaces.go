package platform

import "io"

// PathResolver provides platform-specific path resolution
type PathResolver interface {
	// GetConfigDir returns the platform-specific configuration directory
	// Windows: %APPDATA%, macOS: ~/Library/Application Support, Linux: ~/.config
	GetConfigDir() string

	// GetAppDataDir returns the platform-specific application data directory
	// Windows: %LOCALAPPDATA%, macOS: ~/Library/Application Support, Linux: ~/.local/share
	GetAppDataDir() string

	// GetUserHomeDir returns the user's home directory
	GetUserHomeDir() string
}

// ProcessManager handles process lifecycle operations
type ProcessManager interface {
	// Start launches a new process with the given command, arguments, and environment
	// Returns the process ID on success
	Start(cmd string, args []string, env map[string]string) (pid int, err error)

	// StartWithOutput launches a new process and returns stdout/stderr readers
	// Returns the process ID and readers for stdout and stderr
	// The readers will be closed when the process exits
	StartWithOutput(cmd string, args []string, env map[string]string) (pid int, stdout, stderr io.ReadCloser, err error)

	// Stop terminates a process by its ID
	// If graceful is true, attempts graceful shutdown before forcing termination
	// timeout specifies how long to wait for graceful shutdown (in seconds)
	Stop(pid int, graceful bool, timeout int) error

	// IsRunning checks if a process with the given ID is currently running
	IsRunning(pid int) bool
}

// SingleInstance ensures only one instance of the application runs at a time
type SingleInstance interface {
	// Acquire attempts to acquire the single instance lock
	// Returns true if the lock was acquired, false if another instance is running
	Acquire() (bool, error)

	// Release releases the single instance lock
	Release() error
}

// ProcessInfo provides information about running processes
type ProcessInfo interface {
	// GetMemoryUsage returns the memory usage of a process in bytes
	// Returns 0 if the process is not found or memory cannot be determined
	GetMemoryUsage(pid int) (uint64, error)
}
