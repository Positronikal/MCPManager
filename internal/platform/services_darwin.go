//go:build darwin
// +build darwin

package platform

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Service represents a system service
type Service struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
	PID         *int   `json:"pid,omitempty"`
}

// GetServices retrieves all system services
func GetServices() ([]Service, error) {
	// Run launchctl list to get all services
	cmd := exec.Command("launchctl", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run launchctl list: %w", err)
	}

	return parseServicesDarwin(string(output))
}

// parseServicesDarwin parses macOS launchctl list output
func parseServicesDarwin(output string) ([]Service, error) {
	var services []Service
	lines := strings.Split(output, "\n")

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip header line
		if i == 0 && strings.Contains(line, "PID") {
			continue
		}

		// Parse line format: "PID    Status    Label"
		// Example: "12345  0         com.apple.Spotlight"
		// Example: "-      0         com.example.service" (not running)
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		pidStr := fields[0]
		statusCode := fields[1]
		label := strings.Join(fields[2:], " ")

		var pid *int
		var status string

		if pidStr == "-" {
			status = "STOPPED"
		} else {
			if p, err := strconv.Atoi(pidStr); err == nil {
				pid = &p
				status = "RUNNING"
			} else {
				status = "UNKNOWN"
			}
		}

		// Parse status code if available
		if statusCode != "0" {
			status = fmt.Sprintf("ERROR_%s", statusCode)
		}

		services = append(services, Service{
			Name:        label,
			Status:      status,
			Description: label,
			PID:         pid,
		})
	}

	return services, nil
}
