package models

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

// TransportType represents the communication transport used by an MCP server
type TransportType string

const (
	TransportStdio   TransportType = "stdio"   // Standard input/output (requires client)
	TransportHTTP    TransportType = "http"    // HTTP-based transport (standalone)
	TransportSSE     TransportType = "sse"     // Server-Sent Events (standalone)
	TransportUnknown TransportType = "unknown" // Transport not yet determined
)

// MCPServer represents an MCP server instance
type MCPServer struct {
	ID               string              `json:"id"`
	Name             string              `json:"name"`
	Version          string              `json:"version,omitempty"`
	InstallationPath string              `json:"installationPath"`
	Transport        TransportType       `json:"transport"`
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

// GenerateDeterministicUUID creates a stable UUID based on server identity
// This ensures the same server always gets the same UUID across discoveries
func GenerateDeterministicUUID(name, installationPath string, source DiscoverySource) string {
	// Create a unique identifier string combining:
	// - Server name (e.g., "Filesystem")
	// - Installation path/command (e.g., "node")
	// - Source (e.g., "extension")
	// This ensures the same server always gets the same ID
	identityString := fmt.Sprintf("%s|%s|%s", name, installationPath, source)

	// Hash the identity string to create a deterministic but unique value
	hash := sha256.Sum256([]byte(identityString))

	// Convert first 16 bytes of hash to UUID format
	// This creates a valid UUID v5-style deterministic ID
	hashHex := hex.EncodeToString(hash[:16])

	// Format as UUID: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	uuidStr := fmt.Sprintf("%s-%s-%s-%s-%s",
		hashHex[0:8],
		hashHex[8:12],
		hashHex[12:16],
		hashHex[16:20],
		hashHex[20:32],
	)

	return uuidStr
}

// NewMCPServer creates a new MCPServer with default values
// Uses deterministic UUID generation to ensure stability across discoveries
func NewMCPServer(name, installationPath string, source DiscoverySource) *MCPServer {
	now := time.Now()
	return &MCPServer{
		ID:               GenerateDeterministicUUID(name, installationPath, source),
		Name:             name,
		InstallationPath: installationPath,
		Transport:        TransportUnknown, // Will be detected later
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
