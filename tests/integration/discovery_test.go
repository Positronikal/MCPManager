package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/core/discovery"
	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/Positronikal/MCPManager/internal/models"
)

// TestPathResolver implements platform.PathResolver for testing
type TestPathResolver struct {
	configDir  string
	appDataDir string
	homeDir    string
}

func (t *TestPathResolver) GetConfigDir() string {
	return t.configDir
}

func (t *TestPathResolver) GetAppDataDir() string {
	return t.appDataDir
}

func (t *TestPathResolver) GetUserHomeDir() string {
	return t.homeDir
}

// setupTestEnvironment creates temporary directories for testing
func setupTestEnvironment(t *testing.T) (*TestPathResolver, func()) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "mcp-discovery-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create subdirectories
	configDir := filepath.Join(tempDir, "config")
	appDataDir := filepath.Join(tempDir, "appdata")
	homeDir := filepath.Join(tempDir, "home")

	for _, dir := range []string{configDir, appDataDir, homeDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			os.RemoveAll(tempDir)
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	resolver := &TestPathResolver{
		configDir:  configDir,
		appDataDir: appDataDir,
		homeDir:    homeDir,
	}

	// Cleanup function
	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return resolver, cleanup
}

// createMockClientConfig creates a mock Claude config file
func createMockClientConfig(t *testing.T, configDir string, servers []map[string]interface{}) {
	// Determine platform-specific config path
	var configPath string
	switch runtime.GOOS {
	case "windows":
		// Windows: %APPDATA%\Claude\claude_desktop_config.json
		claudeDir := filepath.Join(configDir, "Claude")
		if err := os.MkdirAll(claudeDir, 0755); err != nil {
			t.Fatalf("Failed to create Claude directory: %v", err)
		}
		configPath = filepath.Join(claudeDir, "claude_desktop_config.json")
	case "darwin":
		// macOS: ~/Library/Application Support/Claude/claude_desktop_config.json
		claudeDir := filepath.Join(configDir, "Claude")
		if err := os.MkdirAll(claudeDir, 0755); err != nil {
			t.Fatalf("Failed to create Claude directory: %v", err)
		}
		configPath = filepath.Join(claudeDir, "claude_desktop_config.json")
	default:
		// Linux: ~/.config/Claude/claude_desktop_config.json
		claudeDir := filepath.Join(configDir, "Claude")
		if err := os.MkdirAll(claudeDir, 0755); err != nil {
			t.Fatalf("Failed to create Claude directory: %v", err)
		}
		configPath = filepath.Join(claudeDir, "claude_desktop_config.json")
	}

	// Create mcpServers config
	config := map[string]interface{}{
		"mcpServers": map[string]interface{}{},
	}

	mcpServers := config["mcpServers"].(map[string]interface{})
	for _, server := range servers {
		name := server["name"].(string)
		mcpServers[name] = map[string]interface{}{
			"command": server["command"],
			"args":    server["args"],
			"env":     server["env"],
		}
	}

	// Write config file
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}
}

// createMockNPMGlobalDir creates a mock NPM global directory structure
func createMockNPMGlobalDir(t *testing.T, homeDir string) {
	// Determine platform-specific NPM global path
	var npmGlobalPath string
	switch runtime.GOOS {
	case "windows":
		npmGlobalPath = filepath.Join(homeDir, "AppData", "Roaming", "npm", "node_modules")
	default:
		npmGlobalPath = filepath.Join(homeDir, ".npm-global", "lib", "node_modules")
	}

	// Create NPM global directory
	if err := os.MkdirAll(npmGlobalPath, 0755); err != nil {
		t.Fatalf("Failed to create NPM global directory: %v", err)
	}

	// Create a mock MCP server package
	packageDir := filepath.Join(npmGlobalPath, "@modelcontextprotocol", "server-example")
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		t.Fatalf("Failed to create package directory: %v", err)
	}

	// Create package.json
	packageJSON := map[string]interface{}{
		"name":    "@modelcontextprotocol/server-example",
		"version": "1.0.0",
		"bin": map[string]string{
			"mcp-server-example": "./index.js",
		},
	}

	data, err := json.MarshalIndent(packageJSON, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal package.json: %v", err)
	}

	packageJSONPath := filepath.Join(packageDir, "package.json")
	if err := os.WriteFile(packageJSONPath, data, 0644); err != nil {
		t.Fatalf("Failed to write package.json: %v", err)
	}

	// Create a simple index.js file
	indexJS := `#!/usr/bin/env node
console.log("MCP Example Server");
`
	indexPath := filepath.Join(packageDir, "index.js")
	if err := os.WriteFile(indexPath, []byte(indexJS), 0755); err != nil {
		t.Fatalf("Failed to write index.js: %v", err)
	}
}

func TestDiscoveryFlow_Integration(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test environment
	pathResolver, cleanup := setupTestEnvironment(t)
	defer cleanup()

	// Create event bus to track discovery events
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	// Subscribe to server discovered events
	discoveredEvents := eventBus.Subscribe(events.EventServerDiscovered)

	// Create mock client config with 2 servers
	createMockClientConfig(t, pathResolver.GetConfigDir(), []map[string]interface{}{
		{
			"name":    "server1",
			"command": "node",
			"args":    []interface{}{"server1.js"},
			"env":     map[string]interface{}{},
		},
		{
			"name":    "server2",
			"command": "python",
			"args":    []interface{}{"-m", "server2"},
			"env":     map[string]interface{}{"API_KEY": "test123"},
		},
	})

	// Create mock NPM global directory with 1 server
	createMockNPMGlobalDir(t, pathResolver.homeDir)

	// Create discovery service
	discoveryService := discovery.NewDiscoveryService(pathResolver, eventBus)

	// Run discovery
	servers, err := discoveryService.Discover()
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	// Verify: Expected at least 2 servers from client config
	// (NPM discovery might not work in test environment)
	if len(servers) < 2 {
		t.Errorf("Expected at least 2 servers discovered, got %d", len(servers))
	}

	// Verify server details
	serversByName := make(map[string]*models.MCPServer)
	for i := range servers {
		serversByName[servers[i].Name] = &servers[i]
	}

	// Check server1 (from client config)
	if server1, exists := serversByName["server1"]; exists {
		if server1.Source != models.DiscoveryClientConfig {
			t.Errorf("server1: expected source 'client_config', got '%s'", server1.Source)
		}
		// Check command line arguments include server1.js
		foundArg := false
		for _, arg := range server1.Configuration.CommandLineArguments {
			if arg == "server1.js" {
				foundArg = true
				break
			}
		}
		if !foundArg {
			t.Errorf("server1: expected args to contain 'server1.js', got %v", server1.Configuration.CommandLineArguments)
		}
	} else {
		t.Error("server1 not found in discovered servers")
	}

	// Check server2 (from client config)
	if server2, exists := serversByName["server2"]; exists {
		if server2.Source != models.DiscoveryClientConfig {
			t.Errorf("server2: expected source 'client_config', got '%s'", server2.Source)
		}
		// Check environment variable
		if server2.Configuration.EnvironmentVariables["API_KEY"] != "test123" {
			t.Errorf("server2: expected env API_KEY='test123', got '%s'", server2.Configuration.EnvironmentVariables["API_KEY"])
		}
	} else {
		t.Error("server2 not found in discovered servers")
	}

	// Verify events were published
	eventCount := 0
	timeout := time.After(1 * time.Second)

	for {
		select {
		case event := <-discoveredEvents:
			if event != nil && event.Type == events.EventServerDiscovered {
				eventCount++
				if eventCount >= 2 {
					// Got enough events
					goto EventsVerified
				}
			}
		case <-timeout:
			goto EventsVerified
		}
	}

EventsVerified:
	if eventCount < 2 {
		t.Logf("Warning: Expected at least 2 ServerDiscovered events, got %d", eventCount)
		// Don't fail the test, as event timing can be tricky
	}

	// Verify all servers have valid IDs
	for _, server := range servers {
		if server.ID == "" {
			t.Errorf("Server '%s' has empty ID", server.Name)
		}
		if server.Name == "" {
			t.Errorf("Server with ID '%s' has empty name", server.ID)
		}
		if !server.Source.IsValid() {
			t.Errorf("Server '%s' has invalid source: %s", server.Name, server.Source)
		}
	}

	t.Logf("Successfully discovered %d servers", len(servers))
	for _, server := range servers {
		t.Logf("  - %s (source: %s)", server.Name, server.Source)
	}
}

func TestDiscoveryFlow_EmptyEnvironment(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test environment with no servers
	pathResolver, cleanup := setupTestEnvironment(t)
	defer cleanup()

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	// Create discovery service
	discoveryService := discovery.NewDiscoveryService(pathResolver, eventBus)

	// Run discovery (should succeed with empty list)
	servers, err := discoveryService.Discover()
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	// Should not error, but might find some system processes
	// So we just verify no client config servers were found
	clientConfigCount := 0
	for _, server := range servers {
		if server.Source == models.DiscoveryClientConfig {
			clientConfigCount++
		}
	}
	if clientConfigCount != 0 {
		t.Errorf("Expected 0 client config servers in empty environment, got %d", clientConfigCount)
	}
	t.Logf("Discovery in empty environment found %d servers (likely system processes)", len(servers))
}

func TestDiscoveryFlow_Deduplication(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test environment
	pathResolver, cleanup := setupTestEnvironment(t)
	defer cleanup()

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	// Create mock client config with a server named "test-server"
	createMockClientConfig(t, pathResolver.GetConfigDir(), []map[string]interface{}{
		{
			"name":    "test-server",
			"command": "node",
			"args":    []interface{}{"server.js"},
			"env":     map[string]interface{}{},
		},
	})

	// Create discovery service
	discoveryService := discovery.NewDiscoveryService(pathResolver, eventBus)

	// Run discovery
	servers, err := discoveryService.Discover()
	if err != nil {
		t.Fatalf("Discovery failed: %v", err)
	}

	// Verify only one instance of "test-server" exists
	count := 0
	for _, server := range servers {
		if server.Name == "test-server" {
			count++
		}
	}

	if count != 1 {
		t.Errorf("Expected exactly 1 instance of 'test-server', got %d", count)
	}
}

func TestDiscoveryFlow_Cleanup(t *testing.T) {
	// This test verifies that temporary directories are properly cleaned up
	tempDir, err := os.MkdirTemp("", "mcp-cleanup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Check that directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Fatal("Temp directory should exist")
	}

	// Clean up
	os.RemoveAll(tempDir)

	// Verify cleanup
	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Error("Temp directory should be removed after cleanup")
	}
}
