package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
	"github.com/hoytech/mcpmanager/internal/platform"
)

// ClientConfigDiscovery discovers MCP servers from client configuration files
type ClientConfigDiscovery struct {
	pathResolver platform.PathResolver
	eventBus     *events.EventBus
}

// NewClientConfigDiscovery creates a new client config discovery instance
func NewClientConfigDiscovery(pathResolver platform.PathResolver, eventBus *events.EventBus) *ClientConfigDiscovery {
	return &ClientConfigDiscovery{
		pathResolver: pathResolver,
		eventBus:     eventBus,
	}
}

// ClientConfig represents the structure of a client configuration file
type ClientConfig struct {
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// ServerConfig represents a single server configuration
type ServerConfig struct {
	Command string                 `json:"command"`
	Args    []string               `json:"args,omitempty"`
	Env     map[string]string      `json:"env,omitempty"`
	Enabled *bool                  `json:"enabled,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// IsEnabled returns whether the server is enabled (default true)
func (sc *ServerConfig) IsEnabled() bool {
	if sc.Enabled == nil {
		return true
	}
	return *sc.Enabled
}

// DiscoverFromClientConfigs discovers servers from all known client config files
func (ccd *ClientConfigDiscovery) DiscoverFromClientConfigs() ([]models.MCPServer, error) {
	var allServers []models.MCPServer

	// Get config directory
	configDir := ccd.pathResolver.GetConfigDir()
	if configDir == "" {
		return allServers, fmt.Errorf("could not determine config directory")
	}

	// Define known client config paths
	configPaths := []struct {
		name string
		path string
	}{
		{
			name: "Claude Desktop",
			path: filepath.Join(configDir, "Claude", "claude_desktop_config.json"),
		},
		{
			name: "Cursor",
			path: filepath.Join(configDir, "Cursor", "mcp_config.json"),
		},
	}

	// Discover from each config file
	for _, cfg := range configPaths {
		servers, err := ccd.discoverFromFile(cfg.path, cfg.name)
		if err != nil {
			// Log warning but continue with other files
			// In production, this would use proper logging
			continue
		}
		allServers = append(allServers, servers...)
	}

	// Publish discovery events
	for i := range allServers {
		if ccd.eventBus != nil {
			ccd.eventBus.Publish(events.ServerDiscoveredEvent(&allServers[i]))
		}
	}

	return allServers, nil
}

// discoverFromFile discovers servers from a specific config file
func (ccd *ClientConfigDiscovery) discoverFromFile(configPath, clientName string) ([]models.MCPServer, error) {
	var servers []models.MCPServer

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// File doesn't exist - not an error, just no servers from this source
		return servers, nil
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return servers, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse JSON
	var config ClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return servers, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Extract servers
	for name, serverCfg := range config.MCPServers {
		// Skip disabled servers
		if !serverCfg.IsEnabled() {
			continue
		}

		// Create server model
		server := models.NewMCPServer(name, serverCfg.Command, models.DiscoveryClientConfig)

		// Set configuration from client config
		server.Configuration.CommandLineArguments = serverCfg.Args
		server.Configuration.EnvironmentVariables = serverCfg.Env

		// Store the source config path in metadata
		if server.Configuration.EnvironmentVariables == nil {
			server.Configuration.EnvironmentVariables = make(map[string]string)
		}

		servers = append(servers, *server)
	}

	return servers, nil
}

// GetConfigPaths returns all known client config paths
func (ccd *ClientConfigDiscovery) GetConfigPaths() []string {
	configDir := ccd.pathResolver.GetConfigDir()
	if configDir == "" {
		return []string{}
	}

	return []string{
		filepath.Join(configDir, "Claude", "claude_desktop_config.json"),
		filepath.Join(configDir, "Cursor", "mcp_config.json"),
	}
}

// DiscoverFromPath discovers servers from a specific config file path
func (ccd *ClientConfigDiscovery) DiscoverFromPath(configPath string) ([]models.MCPServer, error) {
	return ccd.discoverFromFile(configPath, "custom")
}
