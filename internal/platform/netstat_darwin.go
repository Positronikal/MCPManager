//go:build darwin
// +build darwin

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
	var allEntries []NetstatEntry

	// Get TCP connections: netstat -anvp tcp
	tcpCmd := exec.Command("netstat", "-anvp", "tcp")
	tcpOutput, err := tcpCmd.Output()
	if err == nil {
		tcpEntries, _ := parseNetstatDarwin(string(tcpOutput), "TCP", pids)
		allEntries = append(allEntries, tcpEntries...)
	}

	// Get UDP connections: netstat -anvp udp
	udpCmd := exec.Command("netstat", "-anvp", "udp")
	udpOutput, err := udpCmd.Output()
	if err == nil {
		udpEntries, _ := parseNetstatDarwin(string(udpOutput), "UDP", pids)
		allEntries = append(allEntries, udpEntries...)
	}

	return allEntries, nil
}

// parseNetstatDarwin parses macOS netstat output
func parseNetstatDarwin(output string, protocol string, pids []int) ([]NetstatEntry, error) {
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

		// macOS netstat format (with -v flag for PIDs):
		// tcp4       0      0  192.168.1.100.52000  93.184.216.34.443   ESTABLISHED 12345
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

		// PID is in the last field if available
		pidStr := ""
		if len(fields) >= 7 {
			pidStr = fields[6]
		} else if len(fields) == 6 && isNumeric(state) {
			// Sometimes state is actually the PID for UDP
			pidStr = state
			state = "N/A"
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
			Protocol:      strings.ToUpper(protocol),
			LocalAddress:  localAddr,
			RemoteAddress: remoteAddr,
			State:         state,
			PID:           pid,
		})
	}

	return entries, nil
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}
