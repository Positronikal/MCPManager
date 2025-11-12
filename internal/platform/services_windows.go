//go:build windows
// +build windows

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
	// Run sc query to get all services
	cmd := exec.Command("sc", "query", "type=", "service", "state=", "all")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run sc query: %w", err)
	}

	return parseServicesWindows(string(output))
}

// parseServicesWindows parses Windows sc query output
func parseServicesWindows(output string) ([]Service, error) {
	var services []Service
	lines := strings.Split(output, "\n")

	var currentService *Service

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			// Empty line indicates end of service entry
			if currentService != nil {
				services = append(services, *currentService)
				currentService = nil
			}
			continue
		}

		// Parse service fields
		if strings.HasPrefix(line, "SERVICE_NAME:") {
			// Start new service entry
			name := strings.TrimSpace(strings.TrimPrefix(line, "SERVICE_NAME:"))
			currentService = &Service{
				Name:        name,
				Status:      "UNKNOWN",
				Description: "",
				PID:         nil,
			}
		} else if currentService != nil {
			if strings.HasPrefix(line, "DISPLAY_NAME:") {
				description := strings.TrimSpace(strings.TrimPrefix(line, "DISPLAY_NAME:"))
				currentService.Description = description
			} else if strings.HasPrefix(line, "STATE") {
				// Parse state line: "STATE              : 4  RUNNING"
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					stateParts := strings.Fields(parts[1])
					if len(stateParts) >= 2 {
						currentService.Status = stateParts[1]
					}
				}
			} else if strings.Contains(line, "PID") {
				// Parse PID line: "        (PID: 1234)"
				if strings.Contains(line, "PID") {
					pidStart := strings.Index(line, "PID")
					if pidStart != -1 {
						pidPart := line[pidStart:]
						// Extract number from "PID: 1234)" or similar
						pidPart = strings.TrimPrefix(pidPart, "PID")
						pidPart = strings.Trim(pidPart, ": ()")
						pidPart = strings.TrimSpace(pidPart)
						if pidStr := strings.Fields(pidPart); len(pidStr) > 0 {
							if pid, err := strconv.Atoi(pidStr[0]); err == nil {
								currentService.PID = &pid
							}
						}
					}
				}
			}
		}
	}

	// Add last service if exists
	if currentService != nil {
		services = append(services, *currentService)
	}

	return services, nil
}
