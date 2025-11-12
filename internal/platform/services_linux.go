//go:build linux
// +build linux

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
	// Run systemctl list-units --type=service --all to get all services
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--no-legend")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run systemctl: %w", err)
	}

	return parseServicesLinux(string(output))
}

// parseServicesLinux parses Linux systemctl output
func parseServicesLinux(output string) ([]Service, error) {
	var services []Service
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse line format: "UNIT                          LOAD   ACTIVE SUB     DESCRIPTION"
		// Example: "ssh.service                   loaded active running OpenSSH server daemon"
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		serviceName := fields[0]
		// load := fields[1]
		active := fields[2]
		sub := fields[3]
		description := ""
		if len(fields) > 4 {
			description = strings.Join(fields[4:], " ")
		}

		// Determine status from active and sub fields
		status := "UNKNOWN"
		if active == "active" {
			if sub == "running" {
				status = "RUNNING"
			} else if sub == "exited" {
				status = "EXITED"
			} else {
				status = strings.ToUpper(sub)
			}
		} else if active == "inactive" {
			status = "STOPPED"
		} else if active == "failed" {
			status = "FAILED"
		} else {
			status = strings.ToUpper(active)
		}

		// Try to get PID for running services
		var pid *int
		if status == "RUNNING" {
			pid = getServicePIDLinux(serviceName)
		}

		services = append(services, Service{
			Name:        serviceName,
			Status:      status,
			Description: description,
			PID:         pid,
		})
	}

	return services, nil
}

// getServicePIDLinux attempts to get the main PID for a service
func getServicePIDLinux(serviceName string) *int {
	cmd := exec.Command("systemctl", "show", serviceName, "--property=MainPID", "--value")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" || pidStr == "0" {
		return nil
	}

	if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
		return &pid
	}

	return nil
}
