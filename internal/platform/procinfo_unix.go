//go:build !windows

package platform

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// DefaultProcessInfo implements ProcessInfo for Unix systems
type DefaultProcessInfo struct{}

// NewProcessInfo creates a new ProcessInfo for the current platform
func NewProcessInfo() ProcessInfo {
	return &DefaultProcessInfo{}
}

// GetMemoryUsage returns the memory usage of a process in bytes
// On Linux, reads from /proc/[pid]/status
// On macOS, uses the same approach as Linux if /proc is available
func (pi *DefaultProcessInfo) GetMemoryUsage(pid int) (uint64, error) {
	// Try to read from /proc/[pid]/status (Linux)
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	file, err := os.Open(statusPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open process status: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Look for VmRSS (Resident Set Size - actual physical memory used)
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}

			// Parse the memory value (in kB)
			memKB, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, fmt.Errorf("failed to parse memory value: %w", err)
			}

			// Convert kB to bytes
			return memKB * 1024, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading process status: %w", err)
	}

	return 0, fmt.Errorf("VmRSS not found in process status")
}
