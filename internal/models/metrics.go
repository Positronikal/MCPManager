package models

import (
	"time"
)

// ServerMetrics represents runtime metrics for an MCP server
type ServerMetrics struct {
	ServerID     string        `json:"serverId"`
	Uptime       time.Duration `json:"uptime"`       // Time since server started
	MemoryBytes  *uint64       `json:"memoryBytes"`  // Current memory usage in bytes (nil if unavailable)
	RequestCount *int64        `json:"requestCount"` // Total requests handled (nil if unavailable)
	Timestamp    time.Time     `json:"timestamp"`    // When these metrics were collected
}

// NewServerMetrics creates a new ServerMetrics instance
func NewServerMetrics(serverID string) *ServerMetrics {
	return &ServerMetrics{
		ServerID:  serverID,
		Timestamp: time.Now(),
	}
}

// IsAvailable returns true if any metrics are available
func (m *ServerMetrics) IsAvailable() bool {
	return m.MemoryBytes != nil || m.RequestCount != nil || m.Uptime > 0
}
