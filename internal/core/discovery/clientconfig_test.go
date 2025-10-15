package discovery

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
)

// MockPathResolver for testing
type MockPathResolver struct {
	configDir string
}

func (m *MockPathResolver) GetConfigDir() string {
	return m.configDir
}

func (m *MockPathResolver) GetAppDataDir() string {
	return m.configDir
}

func (m *MockPathResolver) GetUserHomeDir() string {
	return m.configDir
}

func TestNewClientConfigDiscovery(t *testing.T) {
	resolver := &MockPathResolver{configDir: "/test"}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	discovery := NewClientConfigDiscovery(resolver, eventBus)
	if discovery == nil {
		t.Fatal("Expected discovery instance to be created")
	}
}

func TestClientConfigDiscovery_ParseValidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Claude Desktop config structure
	claudeDir := filepath.Join(tmpDir, "Claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a valid config file
	config := ClientConfig{
		MCPServers: map[string]ServerConfig{
			"test-server": {
				Command: "node",
				Args:    []string{"server.js"},
				Env: map[string]string{
					"NODE_ENV": "production",
				},
			},
			"python-server": {
				Command: "python",
				Args:    []string{"-m", "mcp_server"},
			},
		},
	}

	configData, _ := json.MarshalIndent(config, "", "  ")
	configPath := filepath.Join(claudeDir, "claude_desktop_config.json")
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatal(err)
	}

	// Create discovery instance
	resolver := &MockPathResolver{configDir: tmpDir}
	eventBus := events.NewEventBus()
	defer eventBus.Close()
	discovery := NewClientConfigDiscovery(resolver, eventBus)

	// Discover servers
	servers, err := discovery.DiscoverFromClientConfigs()
	if err != nil {
		t.Fatalf("Failed to discover servers: %v", err)
	}

	// Verify servers were found
	if len(servers) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(servers))
	}

	// Check server properties
	foundTestServer := false
	foundPythonServer := false

	for _, server := range servers {
		if server.Name == "test-server" {
			foundTestServer = true
			if server.InstallationPath != "node" {
				t.Errorf("Expected command 'node', got '%s'", server.InstallationPath)
			}
			if len(server.Configuration.CommandLineArguments) != 1 {
				t.Error("Expected 1 argument")
			}
			if server.Configuration.EnvironmentVariables["NODE_ENV"] != "production" {
				t.Error("Environment variable not set correctly")
			}
			if server.Source != models.DiscoveryClientConfig {
				t.Error("Source should be client_config")
			}
		}

		if server.Name == "python-server" {
			foundPythonServer = true
			if server.InstallationPath != "python" {
				t.Errorf("Expected command 'python', got '%s'", server.InstallationPath)
			}
			if len(server.Configuration.CommandLineArguments) != 2 {
				t.Error("Expected 2 arguments")
			}
		}
	}

	if !foundTestServer {
		t.Error("test-server not found")
	}
	if !foundPythonServer {
		t.Error("python-server not found")
	}
}

func TestClientConfigDiscovery_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	resolver := &MockPathResolver{configDir: tmpDir}
	eventBus := events.NewEventBus()
	defer eventBus.Close()
	discovery := NewClientConfigDiscovery(resolver, eventBus)

	// Should not error when files don't exist
	servers, err := discovery.DiscoverFromClientConfigs()
	if err != nil {
		t.Errorf("Should not error on missing files: %v", err)
	}

	// Should return empty list
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers, got %d", len(servers))
	}
}

func TestClientConfigDiscovery_MalformedJSON(t *testing.T) {
	tmpDir := t.TempDir()

	// Create Claude Desktop config structure
	claudeDir := filepath.Join(tmpDir, "Claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a malformed config file
	configPath := filepath.Join(claudeDir, "claude_desktop_config.json")
	if err := os.WriteFile(configPath, []byte("{invalid json}"), 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &MockPathResolver{configDir: tmpDir}
	eventBus := events.NewEventBus()
	defer eventBus.Close()
	discovery := NewClientConfigDiscovery(resolver, eventBus)

	// Should continue with other files despite error
	servers, err := discovery.DiscoverFromClientConfigs()
	if err != nil {
		t.Errorf("Should not return error for malformed file: %v", err)
	}

	// Should return empty list (skipped malformed file)
	if len(servers) != 0 {
		t.Errorf("Expected 0 servers from malformed config, got %d", len(servers))
	}
}

func TestClientConfigDiscovery_DisabledServer(t *testing.T) {
	tmpDir := t.TempDir()

	claudeDir := filepath.Join(tmpDir, "Claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create config with disabled server
	disabled := false
	config := ClientConfig{
		MCPServers: map[string]ServerConfig{
			"enabled-server": {
				Command: "node",
				Args:    []string{"server.js"},
			},
			"disabled-server": {
				Command: "python",
				Args:    []string{"server.py"},
				Enabled: &disabled,
			},
		},
	}

	configData, _ := json.MarshalIndent(config, "", "  ")
	configPath := filepath.Join(claudeDir, "claude_desktop_config.json")
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &MockPathResolver{configDir: tmpDir}
	eventBus := events.NewEventBus()
	defer eventBus.Close()
	discovery := NewClientConfigDiscovery(resolver, eventBus)

	servers, err := discovery.DiscoverFromClientConfigs()
	if err != nil {
		t.Fatal(err)
	}

	// Should only find enabled server
	if len(servers) != 1 {
		t.Errorf("Expected 1 server (disabled should be skipped), got %d", len(servers))
	}

	if len(servers) > 0 && servers[0].Name != "enabled-server" {
		t.Error("Should only find enabled-server")
	}
}

func TestClientConfigDiscovery_EventsPublished(t *testing.T) {
	tmpDir := t.TempDir()

	claudeDir := filepath.Join(tmpDir, "Claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	config := ClientConfig{
		MCPServers: map[string]ServerConfig{
			"test-server": {
				Command: "node",
				Args:    []string{"server.js"},
			},
		},
	}

	configData, _ := json.MarshalIndent(config, "", "  ")
	configPath := filepath.Join(claudeDir, "claude_desktop_config.json")
	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &MockPathResolver{configDir: tmpDir}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	// Subscribe to events
	eventChan := eventBus.Subscribe(events.EventServerDiscovered)

	discovery := NewClientConfigDiscovery(resolver, eventBus)

	// Discover servers (should publish events)
	go discovery.DiscoverFromClientConfigs()

	// Wait for event
	select {
	case event := <-eventChan:
		if event.Type != events.EventServerDiscovered {
			t.Error("Wrong event type")
		}
		if event.Data["name"] != "test-server" {
			t.Error("Event should contain server name")
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for event")
	}
}

func TestClientConfigDiscovery_GetConfigPaths(t *testing.T) {
	tmpDir := t.TempDir()

	resolver := &MockPathResolver{configDir: tmpDir}
	eventBus := events.NewEventBus()
	defer eventBus.Close()
	discovery := NewClientConfigDiscovery(resolver, eventBus)

	paths := discovery.GetConfigPaths()

	if len(paths) != 2 {
		t.Errorf("Expected 2 config paths, got %d", len(paths))
	}

	// Check paths contain expected components
	foundClaude := false
	foundCursor := false

	for _, path := range paths {
		if filepath.Base(path) == "claude_desktop_config.json" {
			foundClaude = true
		}
		if filepath.Base(path) == "mcp_config.json" {
			foundCursor = true
		}
	}

	if !foundClaude {
		t.Error("Should include Claude Desktop config path")
	}
	if !foundCursor {
		t.Error("Should include Cursor config path")
	}
}

func TestClientConfigDiscovery_DiscoverFromPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Create custom config file
	config := ClientConfig{
		MCPServers: map[string]ServerConfig{
			"custom-server": {
				Command: "custom",
				Args:    []string{"--custom"},
			},
		},
	}

	configData, _ := json.MarshalIndent(config, "", "  ")
	customPath := filepath.Join(tmpDir, "custom_config.json")
	if err := os.WriteFile(customPath, configData, 0644); err != nil {
		t.Fatal(err)
	}

	resolver := &MockPathResolver{configDir: tmpDir}
	eventBus := events.NewEventBus()
	defer eventBus.Close()
	discovery := NewClientConfigDiscovery(resolver, eventBus)

	// Discover from custom path
	servers, err := discovery.DiscoverFromPath(customPath)
	if err != nil {
		t.Fatalf("Failed to discover from custom path: %v", err)
	}

	if len(servers) != 1 {
		t.Errorf("Expected 1 server, got %d", len(servers))
	}

	if len(servers) > 0 && servers[0].Name != "custom-server" {
		t.Error("Should find custom-server")
	}
}
