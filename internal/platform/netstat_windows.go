//go:build windows
// +build windows

package platform

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// NetstatEntry represents a network connection
type NetstatEntry struct {
	Protocol      string `json:"protocol"`
	LocalAddress  string `json:"localAddress"`
	RemoteAddress string `json:"remoteAddress"`
	State         string `json:"state"`
	PID           int    `json:"pid"`
}

// GetNetstat retrieves network connections for the specified PIDs
// If pids is empty, returns all connections
func GetNetstat(pids []int) ([]NetstatEntry, error) {
	// Run netstat -ano to get all connections with PIDs
	cmd := exec.Command("netstat", "-ano")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run netstat: %w", err)
	}

	return parseNetstatWindows(string(output), pids)
}

// parseNetstatWindows parses Windows netstat output
func parseNetstatWindows(output string, pids []int) ([]NetstatEntry, error) {
	var entries []NetstatEntry
	lines := strings.Split(output, "\n")

	// Create a map for faster PID lookup
	pidMap := make(map[int]bool)
	for _, pid := range pids {
		pidMap[pid] = true
	}
	filterByPID := len(pids) > 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip headers
		if strings.Contains(line, "Proto") || strings.Contains(line, "Active") {
			continue
		}

		// Parse line: TCP    0.0.0.0:135            0.0.0.0:0              LISTENING       1234
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		protocol := strings.ToUpper(fields[0])
		if protocol != "TCP" && protocol != "UDP" {
			continue
		}

		localAddr := fields[1]
		remoteAddr := fields[2]
		state := ""
		pidStr := ""

		// TCP has state field, UDP doesn't
		if protocol == "TCP" {
			if len(fields) >= 5 {
				state = fields[3]
				pidStr = fields[4]
			}
		} else {
			// UDP: Proto LocalAddr RemoteAddr PID
			if len(fields) >= 4 {
				state = "N/A"
				pidStr = fields[3]
			}
		}

		if pidStr == "" {
			continue
		}

		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		// Filter by PID if requested
		if filterByPID && !pidMap[pid] {
			continue
		}

		entries = append(entries, NetstatEntry{
			Protocol:      protocol,
			LocalAddress:  localAddr,
			RemoteAddress: remoteAddr,
			State:         state,
			PID:           pid,
		})
	}

	return entries, nil
}
