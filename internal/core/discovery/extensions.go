package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/Positronikal/MCPManager/internal/models"
	"github.com/Positronikal/MCPManager/internal/platform"
)

// ClaudeExtensionsDiscovery discovers MCP servers from Claude Extensions
type ClaudeExtensionsDiscovery struct {
	pathResolver platform.PathResolver
	eventBus     *events.EventBus
}

// NewClaudeExtensionsDiscovery creates a new Claude extensions discovery instance
func NewClaudeExtensionsDiscovery(pathResolver platform.PathResolver, eventBus *events.EventBus) *ClaudeExtensionsDiscovery {
	return &ClaudeExtensionsDiscovery{
		pathResolver: pathResolver,
		eventBus:     eventBus,
	}
}

// ExtensionManifest represents the structure of a Claude extension manifest.json
type ExtensionManifest struct {
	DXTVersion      string                 `json:"dxt_version"`
	Name            string                 `json:"name"`
	DisplayName     string                 `json:"display_name"`
	Version         string                 `json:"version"`
	Description     string                 `json:"description"`
	LongDescription string                 `json:"long_description"`
	Author          map[string]string      `json:"author,omitempty"`
	Homepage        string                 `json:"homepage,omitempty"`
	Documentation   string                 `json:"documentation,omitempty"`
	Support         string                 `json:"support,omitempty"`
	Icon            string                 `json:"icon,omitempty"`
	Tools           []ExtensionTool        `json:"tools,omitempty"`
	Server          ExtensionServer        `json:"server"`
	Keywords        []string               `json:"keywords,omitempty"`
	License         string                 `json:"license,omitempty"`
	Compatibility   map[string]interface{} `json:"compatibility,omitempty"`
	UserConfig      map[string]interface{} `json:"user_config,omitempty"`
}

// ExtensionTool represents a tool provided by the extension
type ExtensionTool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ExtensionServer represents the server configuration in the manifest
type ExtensionServer struct {
	Type       string          `json:"type"`
	EntryPoint string          `json:"entry_point"`
	MCPConfig  ExtensionMCPCfg `json:"mcp_config"`
}

// ExtensionMCPCfg represents the MCP configuration
type ExtensionMCPCfg struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// ExtensionSettings represents the user settings for an extension
type ExtensionSettings struct {
	IsEnabled  bool                   `json:"isEnabled"`
	UserConfig map[string]interface{} `json:"userConfig,omitempty"`
}

// DiscoverFromExtensions discovers servers from Claude Extensions
func (ced *ClaudeExtensionsDiscovery) DiscoverFromExtensions() ([]models.MCPServer, error) {
	var allServers []models.MCPServer

	// Get config directory
	configDir := ced.pathResolver.GetConfigDir()
	fmt.Printf("  Config directory: %s\n", configDir)
	if configDir == "" {
		return allServers, fmt.Errorf("could not determine config directory")
	}

	// Define paths
	extensionsDir := filepath.Join(configDir, "Claude", "Claude Extensions")
	settingsDir := filepath.Join(configDir, "Claude", "Claude Extensions Settings")

	fmt.Printf("  Extensions directory: %s\n", extensionsDir)
	fmt.Printf("  Settings directory: %s\n", settingsDir)

	// Check if extensions directory exists
	if _, err := os.Stat(extensionsDir); os.IsNotExist(err) {
		fmt.Printf("  Extensions directory does not exist\n")
		return allServers, nil
	}

	// Scan extensions directory
	entries, err := os.ReadDir(extensionsDir)
	if err != nil {
		return allServers, fmt.Errorf("failed to read extensions directory: %w", err)
	}

	fmt.Printf("  Found %d extension entries\n", len(entries))

	// Process each extension
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extensionID := entry.Name()
		fmt.Printf("    Scanning extension: %s\n", extensionID)

		// Read manifest
		manifestPath := filepath.Join(extensionsDir, extensionID, "manifest.json")
		manifest, err := ced.readManifest(manifestPath)
		if err != nil {
			fmt.Printf("      ERROR reading manifest: %v\n", err)
			continue
		}

		fmt.Printf("      Found: %s v%s\n", manifest.DisplayName, manifest.Version)

		// Check if extension has an MCP server
		if manifest.Server.MCPConfig.Command == "" {
			fmt.Printf("      No MCP server configuration found\n")
			continue
		}

		// Read settings
		settingsPath := filepath.Join(settingsDir, extensionID+".json")
		settings, err := ced.readSettings(settingsPath)
		if err != nil {
			fmt.Printf("      WARNING: Could not read settings: %v (treating as enabled)\n", err)
			// Default to enabled if settings don't exist
			settings = &ExtensionSettings{IsEnabled: true}
		}

		fmt.Printf("      Extension enabled: %v\n", settings.IsEnabled)

		// Skip disabled extensions
		if !settings.IsEnabled {
			fmt.Printf("      Skipped (disabled)\n")
			continue
		}

		// Create server model
		extensionPath := filepath.Join(extensionsDir, extensionID)
		server := ced.createServerFromExtension(manifest, settings, extensionPath, extensionID)

		fmt.Printf("      Added to server list (ID: %s)\n", server.ID)
		allServers = append(allServers, *server)
	}

	// Publish discovery events
	for i := range allServers {
		if ced.eventBus != nil {
			ced.eventBus.Publish(events.ServerDiscoveredEvent(&allServers[i]))
		}
	}

	return allServers, nil
}

// readManifest reads and parses an extension manifest.json file
func (ced *ClaudeExtensionsDiscovery) readManifest(manifestPath string) (*ExtensionManifest, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest ExtensionManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	return &manifest, nil
}

// readSettings reads and parses an extension settings file
func (ced *ClaudeExtensionsDiscovery) readSettings(settingsPath string) (*ExtensionSettings, error) {
	// Check if file exists
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		// Settings don't exist - treat as enabled by default
		return &ExtensionSettings{IsEnabled: true}, nil
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read settings: %w", err)
	}

	var settings ExtensionSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, fmt.Errorf("failed to parse settings: %w", err)
	}

	return &settings, nil
}

// createServerFromExtension creates an MCPServer model from an extension
func (ced *ClaudeExtensionsDiscovery) createServerFromExtension(
	manifest *ExtensionManifest,
	settings *ExtensionSettings,
	extensionPath string,
	extensionID string,
) *models.MCPServer {
	// Use display name as server name
	serverName := manifest.DisplayName
	if serverName == "" {
		serverName = manifest.Name
	}
	if serverName == "" {
		serverName = extensionID
	}

	// Resolve command and args
	command := manifest.Server.MCPConfig.Command
	args := ced.resolveArgs(manifest.Server.MCPConfig.Args, extensionPath, settings)

	// Create server with command as InstallationPath (for lifecycle to execute)
	// Note: For extensions, the actual extension path is stored in __EXTENSION_PATH__ env var
	server := models.NewMCPServer(serverName, command, models.DiscoveryExtension)

	// Set version
	server.Version = manifest.Version

	// Set configuration
	server.Configuration.CommandLineArguments = args
	server.Configuration.EnvironmentVariables = manifest.Server.MCPConfig.Env

	// Store extension metadata in configuration
	if server.Configuration.EnvironmentVariables == nil {
		server.Configuration.EnvironmentVariables = make(map[string]string)
	}
	server.Configuration.EnvironmentVariables["__EXTENSION_ID__"] = extensionID
	server.Configuration.EnvironmentVariables["__EXTENSION_PATH__"] = extensionPath
	server.Configuration.EnvironmentVariables["__EXTENSION_TYPE__"] = manifest.Server.Type

	// Extract tools/capabilities
	server.Capabilities = make([]string, 0, len(manifest.Tools))
	server.Tools = make([]string, 0, len(manifest.Tools))
	for _, tool := range manifest.Tools {
		server.Tools = append(server.Tools, tool.Name)
		server.Capabilities = append(server.Capabilities, tool.Name)
	}

	// Detect transport type
	server.Transport = ced.detectTransport(manifest, command)

	return server
}

// detectTransport determines the transport type for a server
func (ced *ClaudeExtensionsDiscovery) detectTransport(manifest *ExtensionManifest, command string) models.TransportType {
	// Check if transport is explicitly defined in manifest
	// (Future: MCP spec may add transport field to manifest)

	// For now, use heuristics based on command and server type
	cmdLower := strings.ToLower(command)

	// Node.js servers typically use stdio
	if cmdLower == "node" || strings.Contains(cmdLower, "node") {
		return models.TransportStdio
	}

	// Python/UV servers typically use stdio
	if cmdLower == "python" || cmdLower == "python3" || cmdLower == "uv" {
		return models.TransportStdio
	}

	// If server type hints at HTTP/SSE
	serverType := strings.ToLower(manifest.Server.Type)
	if strings.Contains(serverType, "http") {
		return models.TransportHTTP
	}
	if strings.Contains(serverType, "sse") {
		return models.TransportSSE
	}

	// Default to stdio for most MCP servers
	return models.TransportStdio
}

// resolveArgs resolves template variables in args like ${__dirname} and ${user_config.*}
func (ced *ClaudeExtensionsDiscovery) resolveArgs(args []string, extensionPath string, settings *ExtensionSettings) []string {
	resolved := make([]string, 0, len(args))

	for _, arg := range args {
		// Replace ${__dirname} with extension path
		resolvedArg := strings.ReplaceAll(arg, "${__dirname}", extensionPath)

		// Replace user_config variables
		// Example: ${user_config.allowed_directories}
		if strings.Contains(resolvedArg, "${user_config.") {
			// Extract config key
			start := strings.Index(resolvedArg, "${user_config.") + 14
			end := strings.Index(resolvedArg[start:], "}")
			if end > 0 {
				configKey := resolvedArg[start : start+end]

				// Look up in settings.UserConfig
				if settings != nil && settings.UserConfig != nil {
					if value, ok := settings.UserConfig[configKey]; ok {
						// Check if value is an array - expand into multiple args
						if arrayValue, isArray := value.([]interface{}); isArray {
							// Array values expand into multiple arguments
							for _, item := range arrayValue {
								if str, ok := item.(string); ok {
									resolved = append(resolved, str)
								}
							}
							// Skip appending the template arg itself
							continue
						} else {
							// Non-array value - convert to string
							resolvedArg = ced.userConfigValueToString(value)
						}
					} else {
						// Config key not found, skip this arg
						continue
					}
				} else {
					// No user config, skip this arg
					continue
				}
			}
		}

		resolved = append(resolved, resolvedArg)
	}

	return resolved
}

// userConfigValueToString converts a user config value to a string
// Note: Arrays are handled separately in resolveArgs() by expanding to multiple args
func (ced *ClaudeExtensionsDiscovery) userConfigValueToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		// Arrays should not reach here - they're expanded in resolveArgs()
		// But if they do, join with commas as fallback
		strValues := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				strValues = append(strValues, str)
			}
		}
		return strings.Join(strValues, ",")
	case map[string]interface{}:
		// Object - serialize as JSON (best effort)
		if data, err := json.Marshal(v); err == nil {
			return string(data)
		}
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}
