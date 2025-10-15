package models

import (
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

// MCPServer represents an MCP server instance
type MCPServer struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Version          string              `json:"version,omitempty"`
	InstallationPath string              `json:"installationPath"`
	Status           ServerStatus        `json:"status"`
	PID              *int                `json:"pid,omitempty"`
	Capabilities     []string            `json:"capabilities,omitempty"`
	Tools            []string            `json:"tools,omitempty"`
	Configuration    ServerConfiguration `json:"configuration"`
	Dependencies     []Dependency        `json:"dependencies,omitempty"`
	DiscoveredAt     time.Time           `json:"discoveredAt"`
	LastSeenAt       time.Time           `json:"lastSeenAt"`
	Source           DiscoverySource     `json:"source"`
}

// NewMCPServer creates a new MCPServer with default values
func NewMCPServer(name, installationPath string, source DiscoverySource) *MCPServer {
	now := time.Now()
	return &MCPServer{
		ID:               uuid.New().String(),
		Name:             name,
		InstallationPath: installationPath,
		Status:           *NewServerStatus(),
		Configuration:    *NewServerConfiguration(),
		Capabilities:     []string{},
		Tools:            []string{},
		Dependencies:     []Dependency{},
		DiscoveredAt:     now,
		LastSeenAt:       now,
		Source:           source,
	}
}

// Validate checks if the MCPServer is in a valid state
func (s *MCPServer) Validate() error {
	// Validate ID is a valid UUID
	if _, err := uuid.Parse(s.ID); err != nil {
		return fmt.Errorf("invalid server ID (must be UUID): %s", s.ID)
	}

	// Validate name is not empty
	if s.Name == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	// Validate installation path is not empty
	if s.InstallationPath == "" {
		return fmt.Errorf("installation path cannot be empty")
	}

	// Check if installation path exists
	if _, err := os.Stat(s.InstallationPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("installation path does not exist: %s", s.InstallationPath)
		}
		return fmt.Errorf("cannot access installation path: %w", err)
	}

	// Validate PID consistency with status
	if s.Status.State == StatusRunning || s.Status.State == StatusStarting {
		if s.PID == nil {
			return fmt.Errorf("PID must be set when status is %s", s.Status.State)
		}
		if *s.PID <= 0 {
			return fmt.Errorf("PID must be positive, got: %d", *s.PID)
		}
	} else if s.Status.State == StatusStopped || s.Status.State == StatusError {
		// Note: PID can be set in error state (crashed process)
		// but must be nil in stopped state
		if s.Status.State == StatusStopped && s.PID != nil {
			return fmt.Errorf("PID must be nil when status is stopped")
		}
	}

	// Validate timestamp consistency
	if s.LastSeenAt.Before(s.DiscoveredAt) {
		return fmt.Errorf("lastSeenAt (%s) cannot be before discoveredAt (%s)",
			s.LastSeenAt.Format(time.RFC3339), s.DiscoveredAt.Format(time.RFC3339))
	}

	// Validate discovery source
	if !s.Source.IsValid() {
		return fmt.Errorf("invalid discovery source: %s", s.Source)
	}

	// Validate nested structures
	if err := s.Status.Validate(); err != nil {
		return fmt.Errorf("invalid status: %w", err)
	}

	if err := s.Configuration.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	for i, dep := range s.Dependencies {
		if err := dep.Validate(); err != nil {
			return fmt.Errorf("invalid dependency at index %d: %w", i, err)
		}
	}

	return nil
}

// UpdateLastSeen updates the LastSeenAt timestamp to now
func (s *MCPServer) UpdateLastSeen() {
	s.LastSeenAt = time.Now()
}

// SetPID sets the process ID for the server
func (s *MCPServer) SetPID(pid int) {
	s.PID = &pid
}

// ClearPID clears the process ID
func (s *MCPServer) ClearPID() {
	s.PID = nil
}
