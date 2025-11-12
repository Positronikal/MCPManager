//go:build linux
// +build linux

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
	// Try netstat -antp (TCP) and -anup (UDP)
	// Note: Requires root/sudo for PID information, but will work without it
	var allEntries []NetstatEntry

	// Get TCP connections
	tcpCmd := exec.Command("netstat", "-antp")
	tcpOutput, err := tcpCmd.Output()
	if err == nil {
		tcpEntries, _ := parseNetstatLinux(string(tcpOutput), "TCP", pids)
		allEntries = append(allEntries, tcpEntries...)
	}

	// Get UDP connections
	udpCmd := exec.Command("netstat", "-anup")
	udpOutput, err := udpCmd.Output()
	if err == nil {
		udpEntries, _ := parseNetstatLinux(string(udpOutput), "UDP", pids)
		allEntries = append(allEntries, udpEntries...)
	}

	return allEntries, nil
}

// parseNetstatLinux parses Linux netstat output
func parseNetstatLinux(output string, protocol string, pids []int) ([]NetstatEntry, error) {
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

		// Linux netstat format:
		// tcp        0      0 0.0.0.0:22              0.0.0.0:*               LISTEN      1234/sshd
		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		proto := strings.ToLower(fields[0])
		if !strings.HasPrefix(proto, "tcp") && !strings.HasPrefix(proto, "udp") {
			continue
		}

		localAddr := fields[3]
		remoteAddr := fields[4]
		state := fields[5]

		// PID/Program is in the last field, format: "1234/program-name" or "-"
		pidProgramStr := ""
		if len(fields) >= 7 {
			pidProgramStr = fields[6]
		}

		if pidProgramStr == "" || pidProgramStr == "-" {
			continue
		}

		// Extract PID from "1234/program-name"
		pidStr := strings.Split(pidProgramStr, "/")[0]
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		// Filter by PID if requested
		if filterByPID && !pidMap[pid] {
			continue
		}

		entries = append(entries, NetstatEntry{
			Protocol:      strings.ToUpper(protocol),
			LocalAddress:  localAddr,
			RemoteAddress: remoteAddr,
			State:         state,
			PID:           pid,
		})
	}

	return entries, nil
}
