package discovery

import (
	"testing"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
)

func TestNewProcessDiscovery(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	discovery := NewProcessDiscovery(eventBus)
	if discovery == nil {
		t.Fatal("Expected discovery instance to be created")
	}
}

func TestProcessDiscovery_DiscoverFromProcesses(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	discovery := NewProcessDiscovery(eventBus)

	// Should work without error (may find 0 or more processes)
	servers, err := discovery.DiscoverFromProcesses()
	if err != nil {
		t.Errorf("Should not error: %v", err)
	}

	if servers == nil {
		t.Error("Should return non-nil slice")
	}

	// If any servers found, verify they have correct source
	for _, server := range servers {
		if server.Source != models.DiscoveryProcess {
			t.Errorf("Expected source to be process, got %s", server.Source)
		}
		if server.Status.State != models.StatusRunning {
			t.Error("Process-discovered servers should be marked as running")
		}
		if server.PID == nil {
			t.Error("Process-discovered servers should have PID set")
		}
	}
}

func TestProcessDiscovery_ParseCSVLine(t *testing.T) {
	testCases := []struct {
		input    string
		expected int // number of fields
	}{
		{`"process.exe","1234","Console","1","1,234 K","Running","user","0:00:01","Title"`, 9},
		{`"simple","5678","Session","2","2,345 K","Running","admin","0:00:02","Another"`, 9},
	}

	for _, tc := range testCases {
		fields := parseCSVLine(tc.input)
		if len(fields) != tc.expected {
			t.Errorf("Expected %d fields, got %d for input: %s", tc.expected, len(fields), tc.input)
		}
	}
}

func TestProcessDiscovery_MCPPatternMatching(t *testing.T) {
	testCases := []struct {
		commandLine string
		shouldMatch bool
	}{
		{"node mcp-server.js", true},
		{"python -m mcp_toolkit", true},
		{"/usr/bin/mcp-server", true},
		{"MCP-Server.exe", true},
		{"regular-process", false},
		{"node server.js", false},
	}

	for _, tc := range testCases {
		// Simulate the matching logic from DiscoverFromProcesses
		matches := containsSubstring(tc.commandLine, "mcp")
		if matches != tc.shouldMatch {
			t.Errorf("Pattern match for '%s': expected %v, got %v", tc.commandLine, tc.shouldMatch, matches)
		}
	}
}
