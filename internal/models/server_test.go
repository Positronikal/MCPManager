package models

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewMCPServer(t *testing.T) {
	server := NewMCPServer("test-server", "/path/to/server", DiscoveryClientConfig)

	if server.ID == "" {
		t.Error("ID should be generated")
	}
	if server.Name != "test-server" {
		t.Errorf("Expected name 'test-server', got '%s'", server.Name)
	}
	if server.Status.State != StatusStopped {
		t.Errorf("Expected initial state to be stopped, got %s", server.Status.State)
	}
	if server.PID != nil {
		t.Error("PID should be nil initially")
	}
}

func TestMCPServerValidation(t *testing.T) {
	// Create a temp directory for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.exe")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		setup   func() *MCPServer
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid server with all required fields",
			setup: func() *MCPServer {
				server := NewMCPServer("valid-server", testFile, DiscoveryClientConfig)
				return server
			},
			wantErr: false,
		},
		{
			name: "empty name",
			setup: func() *MCPServer {
				server := NewMCPServer("", testFile, DiscoveryClientConfig)
				return server
			},
			wantErr: true,
			errMsg:  "name cannot be empty",
		},
		{
			name: "invalid UUID",
			setup: func() *MCPServer {
				server := NewMCPServer("test", testFile, DiscoveryClientConfig)
				server.ID = "not-a-uuid"
				return server
			},
			wantErr: true,
			errMsg:  "invalid server ID",
		},
		{
			name: "PID set when status is stopped",
			setup: func() *MCPServer {
				server := NewMCPServer("test", testFile, DiscoveryClientConfig)
				pid := 1234
				server.PID = &pid
				server.Status.State = StatusStopped
				return server
			},
			wantErr: true,
			errMsg:  "PID must be nil when status is stopped",
		},
		{
			name: "PID nil when status is running",
			setup: func() *MCPServer {
				server := NewMCPServer("test", testFile, DiscoveryClientConfig)
				server.Status.State = StatusRunning
				server.PID = nil
				return server
			},
			wantErr: true,
			errMsg:  "PID must be set when status is",
		},
		{
			name: "lastSeenAt before discoveredAt",
			setup: func() *MCPServer {
				server := NewMCPServer("test", testFile, DiscoveryClientConfig)
				server.LastSeenAt = server.DiscoveredAt.Add(-1 * time.Hour)
				return server
			},
			wantErr: true,
			errMsg:  "lastSeenAt",
		},
		{
			name: "installation path doesn't exist",
			setup: func() *MCPServer {
				server := NewMCPServer("test", "/nonexistent/path", DiscoveryClientConfig)
				return server
			},
			wantErr: true,
			errMsg:  "installation path does not exist",
		},
		{
			name: "valid server with PID in running state",
			setup: func() *MCPServer {
				server := NewMCPServer("test", testFile, DiscoveryClientConfig)
				server.Status.State = StatusRunning
				pid := 1234
				server.PID = &pid
				return server
			},
			wantErr: false,
		},
		{
			name: "invalid discovery source",
			setup: func() *MCPServer {
				server := NewMCPServer("test", testFile, DiscoveryClientConfig)
				server.Source = DiscoverySource("invalid")
				return server
			},
			wantErr: true,
			errMsg:  "invalid discovery source",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setup()
			err := server.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestMCPServerJSONMarshaling(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.exe")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	server := NewMCPServer("test-server", testFile, DiscoveryClientConfig)
	pid := 1234
	server.PID = &pid
	server.Version = "1.0.0"
	server.Capabilities = []string{"tools", "prompts"}
	server.Tools = []string{"tool1", "tool2"}

	// Marshal to JSON
	data, err := json.Marshal(server)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal back
	var unmarshaled MCPServer
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Verify key fields
	if unmarshaled.ID != server.ID {
		t.Errorf("ID mismatch: got %s, want %s", unmarshaled.ID, server.ID)
	}
	if unmarshaled.Name != server.Name {
		t.Errorf("Name mismatch: got %s, want %s", unmarshaled.Name, server.Name)
	}
	if *unmarshaled.PID != *server.PID {
		t.Errorf("PID mismatch: got %d, want %d", *unmarshaled.PID, *server.PID)
	}

	// Verify timestamp format (ISO 8601)
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		t.Fatalf("Failed to unmarshal to map: %v", err)
	}

	discoveredAt, ok := jsonMap["discoveredAt"].(string)
	if !ok {
		t.Error("discoveredAt should be a string in JSON")
	}
	// Verify it's RFC3339 format (ISO 8601)
	if _, err := time.Parse(time.RFC3339, discoveredAt); err != nil {
		t.Errorf("discoveredAt is not in RFC3339 format: %s", discoveredAt)
	}
}

func TestMCPServerHelperMethods(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.exe")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	server := NewMCPServer("test", testFile, DiscoveryClientConfig)
	initialLastSeen := server.LastSeenAt

	// Test UpdateLastSeen
	time.Sleep(10 * time.Millisecond)
	server.UpdateLastSeen()
	if !server.LastSeenAt.After(initialLastSeen) {
		t.Error("UpdateLastSeen should update the timestamp")
	}

	// Test SetPID
	server.SetPID(1234)
	if server.PID == nil || *server.PID != 1234 {
		t.Error("SetPID should set the PID")
	}

	// Test ClearPID
	server.ClearPID()
	if server.PID != nil {
		t.Error("ClearPID should clear the PID")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
