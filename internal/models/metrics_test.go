package models

import (
	"testing"
	"time"
)

func TestNewServerMetrics(t *testing.T) {
	serverID := "test-server"
	metrics := NewServerMetrics(serverID)

	if metrics.ServerID != serverID {
		t.Errorf("Expected serverID %s, got %s", serverID, metrics.ServerID)
	}

	if metrics.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if metrics.MemoryBytes != nil {
		t.Error("Expected MemoryBytes to be nil initially")
	}

	if metrics.RequestCount != nil {
		t.Error("Expected RequestCount to be nil initially")
	}

	if metrics.Uptime != 0 {
		t.Error("Expected Uptime to be 0 initially")
	}
}

func TestServerMetrics_IsAvailable(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*ServerMetrics)
		available bool
	}{
		{
			name:      "No metrics available",
			setup:     func(m *ServerMetrics) {},
			available: false,
		},
		{
			name: "Memory available",
			setup: func(m *ServerMetrics) {
				mem := uint64(1024)
				m.MemoryBytes = &mem
			},
			available: true,
		},
		{
			name: "Request count available",
			setup: func(m *ServerMetrics) {
				count := int64(100)
				m.RequestCount = &count
			},
			available: true,
		},
		{
			name: "Uptime available",
			setup: func(m *ServerMetrics) {
				m.Uptime = 5 * time.Minute
			},
			available: true,
		},
		{
			name: "All metrics available",
			setup: func(m *ServerMetrics) {
				mem := uint64(2048)
				count := int64(200)
				m.MemoryBytes = &mem
				m.RequestCount = &count
				m.Uptime = 10 * time.Minute
			},
			available: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metrics := NewServerMetrics("test")
			tt.setup(metrics)

			if metrics.IsAvailable() != tt.available {
				t.Errorf("Expected IsAvailable() to be %v, got %v", tt.available, metrics.IsAvailable())
			}
		})
	}
}
