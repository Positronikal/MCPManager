package discovery

import (
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
)

// ProcessDiscovery discovers already-running MCP servers
type ProcessDiscovery struct {
	eventBus *events.EventBus
}

// NewProcessDiscovery creates a new process discovery instance
func NewProcessDiscovery(eventBus *events.EventBus) *ProcessDiscovery {
	return &ProcessDiscovery{
		eventBus: eventBus,
	}
}

// DiscoverFromProcesses is DEPRECATED and should not be called directly.
// Use DiscoveryService.Discover() instead, which properly matches processes
// against discovered servers per the spec's three-tier strategy.
//
// This method is kept for backward compatibility but returns empty list.
func (pd *ProcessDiscovery) DiscoverFromProcesses() ([]models.MCPServer, error) {
	// Per spec research.md §16: Process discovery should NOT create new server entries
	// It should only match PIDs to servers discovered from client configs and filesystem
	return []models.MCPServer{}, nil
}

// ProcessInfo represents a running process
type ProcessInfo struct {
	PID         int
	Name        string
	CommandLine string
	ParentPID   int
}

// listProcesses returns a list of running processes (platform-specific)
func (pd *ProcessDiscovery) listProcesses() ([]ProcessInfo, error) {
	switch runtime.GOOS {
	case "windows":
		return pd.listProcessesWindows()
	case "darwin", "linux":
		return pd.listProcessesUnix()
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// listProcessesWindows is implemented in process_windows.go using native Win32 API
// This avoids dependency on WMIC which may be disabled in enterprise environments

// listProcessesUnix lists processes on Unix systems using ps
func (pd *ProcessDiscovery) listProcessesUnix() ([]ProcessInfo, error) {
	cmd := exec.Command("ps", "aux")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run ps: %w", err)
	}

	return pd.parsePsOutput(string(output))
}

// parsePsOutput parses Unix ps aux output
func (pd *ProcessDiscovery) parsePsOutput(output string) ([]ProcessInfo, error) {
	var processes []ProcessInfo

	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if i == 0 || line == "" {
			continue // Skip header
		}

		fields := strings.Fields(line)
		if len(fields) < 11 {
			continue
		}

		pidStr := fields[1]
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		// Command line starts at field 10
		commandLine := strings.Join(fields[10:], " ")
		name := fields[10]

		processes = append(processes, ProcessInfo{
			PID:         pid,
			Name:        name,
			CommandLine: commandLine,
		})
	}

	return processes, nil
}

// parseCSVLine parses a CSV line respecting quoted fields
func parseCSVLine(line string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		char := line[i]

		if char == '"' {
			inQuotes = !inQuotes
			current.WriteByte(char)
		} else if char == ',' && !inQuotes {
			fields = append(fields, current.String())
			current.Reset()
		} else {
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		fields = append(fields, current.String())
	}

	return fields
}
