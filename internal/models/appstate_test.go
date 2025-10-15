package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewApplicationState(t *testing.T) {
	state := NewApplicationState()

	// Check default values
	if state.Preferences.Theme != "dark" {
		t.Errorf("Expected default theme 'dark', got %s", state.Preferences.Theme)
	}
	if state.Preferences.LogRetentionPerServer != 1000 {
		t.Errorf("Expected default log retention 1000, got %d", state.Preferences.LogRetentionPerServer)
	}
	if state.WindowLayout.Width != 1024 {
		t.Errorf("Expected default width 1024, got %d", state.WindowLayout.Width)
	}
	if state.WindowLayout.Height != 768 {
		t.Errorf("Expected default height 768, got %d", state.WindowLayout.Height)
	}
	if state.WindowLayout.LogPanelHeight != 300 {
		t.Errorf("Expected default log panel height 300, got %d", state.WindowLayout.LogPanelHeight)
	}
	if len(state.DiscoveredServers) != 0 {
		t.Error("Expected empty discovered servers list")
	}
	if len(state.MonitoredConfigPaths) != 0 {
		t.Error("Expected empty monitored config paths list")
	}
}

func TestApplicationStateValidation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *ApplicationState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid state with all fields",
			setup: func() *ApplicationState {
				return NewApplicationState()
			},
			wantErr: false,
		},
		{
			name: "window width too small",
			setup: func() *ApplicationState {
				state := NewApplicationState()
				state.WindowLayout.Width = 600
				return state
			},
			wantErr: true,
			errMsg:  "window width must be at least 640",
		},
		{
			name: "window height too small",
			setup: func() *ApplicationState {
				state := NewApplicationState()
				state.WindowLayout.Height = 400
				return state
			},
			wantErr: true,
			errMsg:  "window height must be at least 480",
		},
		{
			name: "negative log panel height",
			setup: func() *ApplicationState {
				state := NewApplicationState()
				state.WindowLayout.LogPanelHeight = -100
				return state
			},
			wantErr: true,
			errMsg:  "log panel height cannot be negative",
		},
		{
			name: "log panel height exceeds window height",
			setup: func() *ApplicationState {
				state := NewApplicationState()
				state.WindowLayout.LogPanelHeight = 1000
				return state
			},
			wantErr: true,
			errMsg:  "log panel height",
		},
		{
			name: "invalid theme",
			setup: func() *ApplicationState {
				state := NewApplicationState()
				state.Preferences.Theme = "blue"
				return state
			},
			wantErr: true,
			errMsg:  "theme must be 'dark' or 'light'",
		},
		{
			name: "log retention too low",
			setup: func() *ApplicationState {
				state := NewApplicationState()
				state.Preferences.LogRetentionPerServer = 50
				return state
			},
			wantErr: true,
			errMsg:  "logRetentionPerServer must be between 100 and 10000",
		},
		{
			name: "log retention too high",
			setup: func() *ApplicationState {
				state := NewApplicationState()
				state.Preferences.LogRetentionPerServer = 20000
				return state
			},
			wantErr: true,
			errMsg:  "logRetentionPerServer must be between 100 and 10000",
		},
		{
			name: "discovered server not a UUID",
			setup: func() *ApplicationState {
				state := NewApplicationState()
				state.DiscoveredServers = []string{"not-a-uuid"}
				return state
			},
			wantErr: true,
			errMsg:  "not a valid UUID",
		},
		{
			name: "monitored path not absolute",
			setup: func() *ApplicationState {
				state := NewApplicationState()
				state.MonitoredConfigPaths = []string{"relative/path"}
				return state
			},
			wantErr: true,
			errMsg:  "not an absolute path",
		},
		{
			name: "valid with discovered servers",
			setup: func() *ApplicationState {
				state := NewApplicationState()
				state.DiscoveredServers = []string{
					uuid.New().String(),
					uuid.New().String(),
				}
				return state
			},
			wantErr: false,
		},
		{
			name: "invalid selected severity",
			setup: func() *ApplicationState {
				state := NewApplicationState()
				state.Filters.SelectedSeverity = LogSeverity("invalid")
				return state
			},
			wantErr: true,
			errMsg:  "invalid selected severity",
		},
		{
			name: "selected server not a UUID",
			setup: func() *ApplicationState {
				state := NewApplicationState()
				state.Filters.SelectedServer = "not-a-uuid"
				return state
			},
			wantErr: true,
			errMsg:  "selected server is not a valid UUID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := tt.setup()
			err := state.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errMsg != "" && !containsSubstring(err.Error(), tt.errMsg) {
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

func TestApplicationState_AddDiscoveredServer(t *testing.T) {
	state := NewApplicationState()

	// Add a valid server
	serverID := uuid.New().String()
	if err := state.AddDiscoveredServer(serverID); err != nil {
		t.Errorf("Failed to add server: %v", err)
	}

	if len(state.DiscoveredServers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(state.DiscoveredServers))
	}
	if state.DiscoveredServers[0] != serverID {
		t.Error("Server ID not added correctly")
	}

	// Add the same server again (should be idempotent)
	if err := state.AddDiscoveredServer(serverID); err != nil {
		t.Errorf("Failed to add duplicate server: %v", err)
	}
	if len(state.DiscoveredServers) != 1 {
		t.Error("Duplicate server was added")
	}

	// Add an invalid UUID
	if err := state.AddDiscoveredServer("not-a-uuid"); err == nil {
		t.Error("Should have failed to add invalid UUID")
	}
}

func TestApplicationState_RemoveDiscoveredServer(t *testing.T) {
	state := NewApplicationState()

	// Add some servers
	server1 := uuid.New().String()
	server2 := uuid.New().String()
	state.AddDiscoveredServer(server1)
	state.AddDiscoveredServer(server2)

	if len(state.DiscoveredServers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(state.DiscoveredServers))
	}

	// Remove first server
	state.RemoveDiscoveredServer(server1)
	if len(state.DiscoveredServers) != 1 {
		t.Errorf("Expected 1 server after removal, got %d", len(state.DiscoveredServers))
	}
	if state.DiscoveredServers[0] != server2 {
		t.Error("Wrong server was removed")
	}

	// Remove non-existent server (should be no-op)
	state.RemoveDiscoveredServer("nonexistent")
	if len(state.DiscoveredServers) != 1 {
		t.Error("List size changed when removing nonexistent server")
	}
}

func TestApplicationState_AddMonitoredPath(t *testing.T) {
	state := NewApplicationState()

	// Add a valid absolute path (use temp dir which is guaranteed to be absolute)
	tmpDir := t.TempDir()
	if err := state.AddMonitoredPath(tmpDir); err != nil {
		t.Errorf("Failed to add path: %v", err)
	}

	if len(state.MonitoredConfigPaths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(state.MonitoredConfigPaths))
	}
	if state.MonitoredConfigPaths[0] != tmpDir {
		t.Error("Path not added correctly")
	}

	// Add the same path again (should be idempotent)
	if err := state.AddMonitoredPath(tmpDir); err != nil {
		t.Errorf("Failed to add duplicate path: %v", err)
	}
	if len(state.MonitoredConfigPaths) != 1 {
		t.Error("Duplicate path was added")
	}

	// Try to add a relative path
	if err := state.AddMonitoredPath("relative/path"); err == nil {
		t.Error("Should have failed to add relative path")
	}
}

func TestUserPreferences(t *testing.T) {
	state := NewApplicationState()

	// Test theme changes
	state.Preferences.Theme = "light"
	if err := state.Validate(); err != nil {
		t.Errorf("Light theme should be valid: %v", err)
	}

	state.Preferences.Theme = "dark"
	if err := state.Validate(); err != nil {
		t.Errorf("Dark theme should be valid: %v", err)
	}

	// Test boolean preferences
	state.Preferences.AutoStartServers = true
	state.Preferences.MinimizeToTray = false
	state.Preferences.ShowNotifications = false
	if err := state.Validate(); err != nil {
		t.Errorf("Boolean preferences should be valid: %v", err)
	}
}

func TestWindowLayout(t *testing.T) {
	state := NewApplicationState()

	// Test minimum dimensions
	state.WindowLayout.Width = 640
	state.WindowLayout.Height = 480
	if err := state.Validate(); err != nil {
		t.Errorf("Minimum dimensions should be valid: %v", err)
	}

	// Test maximized state
	state.WindowLayout.Maximized = true
	if err := state.Validate(); err != nil {
		t.Error("Maximized state should be valid")
	}

	// Test position
	state.WindowLayout.X = -100
	state.WindowLayout.Y = -100
	if err := state.Validate(); err != nil {
		t.Error("Negative position should be valid (multi-monitor)")
	}
}

func TestFilters(t *testing.T) {
	state := NewApplicationState()

	// Test valid filters
	serverID := uuid.New().String()
	state.Filters.SelectedServer = serverID
	state.Filters.SelectedSeverity = LogInfo
	state.Filters.SearchQuery = "test query"

	if err := state.Validate(); err != nil {
		t.Errorf("Valid filters should pass validation: %v", err)
	}

	// Test empty filters (should be valid)
	state.Filters.SelectedServer = ""
	state.Filters.SelectedSeverity = ""
	state.Filters.SearchQuery = ""
	if err := state.Validate(); err != nil {
		t.Error("Empty filters should be valid")
	}
}
