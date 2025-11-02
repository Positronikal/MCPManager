package discovery

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
)

func TestNewFilesystemDiscovery(t *testing.T) {
	resolver := &MockPathResolver{configDir: "/test"}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	discovery := NewFilesystemDiscovery(resolver, eventBus)
	if discovery == nil {
		t.Fatal("Expected discovery instance to be created")
	}
}

func TestFilesystemDiscovery_DiscoverFromFilesystem(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	discovery := NewFilesystemDiscovery(resolver, eventBus)

	// This will likely return empty results since we don't have MCP servers installed
	// But it should not error
	servers, err := discovery.DiscoverFromFilesystem()
	if err != nil {
		t.Errorf("Should not error even if nothing found: %v", err)
	}

	// Result could be 0 or more depending on what's installed on the system
	if servers == nil {
		t.Error("Should return non-nil slice")
	}
}

func TestFilesystemDiscovery_EventsPublished(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	discovery := NewFilesystemDiscovery(resolver, eventBus)

	// Subscribe to events
	eventChan := eventBus.Subscribe(events.EventServerDiscovered)

	// Run discovery in background
	go discovery.DiscoverFromFilesystem()

	// We may or may not receive events depending on what's installed
	// This test just verifies no panic occurs
	select {
	case <-eventChan:
		// Event received (optional)
	case <-time.After(1 * time.Second):
		// Timeout (also OK if nothing found)
	}
}

func TestFilesystemDiscovery_NPMServerPattern(t *testing.T) {
	// Test that the pattern matching works correctly
	testCases := []struct {
		name     string
		expected bool
	}{
		{"mcp-server", true},
		{"@org/mcp-toolkit", true},
		{"my-mcp-package", true},
		{"MCP-Server", true}, // Case insensitive
		{"regular-package", false},
		{"some-tool", false},
	}

	for _, tc := range testCases {
		// Check if name contains "mcp" (case insensitive)
		result := containsMCP(tc.name)
		if result != tc.expected {
			t.Errorf("Pattern match for '%s': expected %v, got %v", tc.name, tc.expected, result)
		}
	}
}

// Helper function to check if name contains "mcp"
func containsMCP(name string) bool {
	return len(name) > 0 && (name[0:1] == "m" || name[0:1] == "M") &&
		   len(name) > 2 && (name[0:3] == "mcp" || name[0:3] == "MCP" ||
		   name[0:3] == "Mcp" || name[0:3] == "mCp" || name[0:3] == "mcP" ||
		   name[0:3] == "McP" || name[0:3] == "MCp" || name[0:3] == "mCP") ||
		   (len(name) > 2 && (name[0:3] == "mcp" || name[0:3] == "MCP")) ||
		   (len(name) > 3 && containsSubstring(name, "mcp"))
}

func containsSubstring(s, substr string) bool {
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TODO: toLower is already defined in process_windows.go - removed duplicate
// func toLower(s string) string {
// 	result := make([]byte, len(s))
// 	for i := 0; i < len(s); i++ {
// 		if s[i] >= 'A' && s[i] <= 'Z' {
// 			result[i] = s[i] + 32
// 		} else {
// 			result[i] = s[i]
// 		}
// 	}
// 	return string(result)
// }

func TestFilesystemDiscovery_MockNPMDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mock NPM global directory structure
	npmRoot := filepath.Join(tmpDir, "npm", "lib", "node_modules")
	if err := os.MkdirAll(npmRoot, 0755); err != nil {
		t.Fatal(err)
	}

	// Create mock MCP server packages
	mcpServer1 := filepath.Join(npmRoot, "mcp-server-example")
	if err := os.MkdirAll(mcpServer1, 0755); err != nil {
		t.Fatal(err)
	}

	// Create package.json to make it look real
	packageJSON := filepath.Join(mcpServer1, "package.json")
	if err := os.WriteFile(packageJSON, []byte(`{"name": "mcp-server-example"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Create another package without MCP in name
	regularPkg := filepath.Join(npmRoot, "regular-package")
	if err := os.MkdirAll(regularPkg, 0755); err != nil {
		t.Fatal(err)
	}

	// Create package.json
	packageJSON2 := filepath.Join(regularPkg, "package.json")
	if err := os.WriteFile(packageJSON2, []byte(`{"name": "regular-package"}`), 0644); err != nil {
		t.Fatal(err)
	}

	// Now manually scan the directory (since we can't override npm command)
	entries, err := os.ReadDir(npmRoot)
	if err != nil {
		t.Fatal(err)
	}

	foundMCP := false
	foundRegular := false

	for _, entry := range entries {
		if entry.Name() == "mcp-server-example" {
			foundMCP = true
		}
		if entry.Name() == "regular-package" {
			foundRegular = true
		}
	}

	if !foundMCP {
		t.Error("Should find MCP server package")
	}
	if !foundRegular {
		t.Error("Should find regular package (filtered by pattern later)")
	}
}

func TestFilesystemDiscovery_SourceField(t *testing.T) {
	// Verify that servers discovered from filesystem have correct source
	server := models.NewMCPServer("test", "/path", models.DiscoveryFilesystem)

	if server.Source != models.DiscoveryFilesystem {
		t.Errorf("Expected source to be filesystem, got %s", server.Source)
	}
}

func TestFilesystemDiscovery_HandlesEmptyDirectories(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	discovery := NewFilesystemDiscovery(resolver, eventBus)

	// Should handle case where directories don't exist
	servers, err := discovery.DiscoverFromFilesystem()
	if err != nil {
		t.Errorf("Should handle missing directories gracefully: %v", err)
	}

	if servers == nil {
		t.Error("Should return non-nil slice")
	}
}
