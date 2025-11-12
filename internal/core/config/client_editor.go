package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ClientType represents the type of MCP client
type ClientType string

const (
	ClientClaudeDesktop ClientType = "claude_desktop"
	ClientCursor        ClientType = "cursor"
	ClientUnknown       ClientType = "unknown"
)

// ClientInfo represents information about an installed MCP client
type ClientInfo struct {
	Type       ClientType `json:"type"`
	Name       string     `json:"name"`
	ConfigPath string     `json:"configPath"`
	Installed  bool       `json:"installed"`
}

// ServerEntry represents a server entry in the client config
type ServerEntry struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// ClientConfig represents the structure of an MCP client configuration file
type ClientConfig struct {
	MCPServers map[string]ServerEntry `json:"mcpServers"`
}

// ClientEditor provides functionality for editing MCP client configuration files
type ClientEditor struct{}

// NewClientEditor creates a new ClientEditor instance
func NewClientEditor() *ClientEditor {
	return &ClientEditor{}
}

// DetectClients detects which MCP clients are installed on the system
// Returns a list of detected clients with their config file paths
func (ce *ClientEditor) DetectClients() ([]ClientInfo, error) {
	var clients []ClientInfo

	// Detect Claude Desktop
	claudeConfig := ce.getClaudeDesktopConfigPath()
	claudeInstalled := ce.fileExists(claudeConfig)
	clients = append(clients, ClientInfo{
		Type:       ClientClaudeDesktop,
		Name:       "Claude Desktop",
		ConfigPath: claudeConfig,
		Installed:  claudeInstalled,
	})

	// Detect Cursor (common paths)
	cursorConfig := ce.getCursorConfigPath()
	cursorInstalled := ce.fileExists(cursorConfig)
	clients = append(clients, ClientInfo{
		Type:       ClientCursor,
		Name:       "Cursor",
		ConfigPath: cursorConfig,
		Installed:  cursorInstalled,
	})

	return clients, nil
}

// ReadConfig reads and parses an MCP client configuration file
func (ce *ClientEditor) ReadConfig(configPath string) (*ClientConfig, error) {
	// Check if file exists
	if !ce.fileExists(configPath) {
		// Return empty config if file doesn't exist
		return &ClientConfig{
			MCPServers: make(map[string]ServerEntry),
		}, nil
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var config ClientConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure MCPServers map is initialized
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]ServerEntry)
	}

	return &config, nil
}

// WriteConfig writes an updated configuration to the client config file
// Creates a backup before writing and validates the JSON structure
func (ce *ClientEditor) WriteConfig(configPath string, config *ClientConfig) error {
	// Validate config structure
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.MCPServers == nil {
		config.MCPServers = make(map[string]ServerEntry)
	}

	// Create backup if file exists
	if ce.fileExists(configPath) {
		if err := ce.createBackup(configPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal config to JSON with indentation for readability
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// AddServer adds a new server entry to the configuration
func (ce *ClientEditor) AddServer(config *ClientConfig, serverName string, command string, args []string, env map[string]string) error {
	// Validate inputs
	if serverName == "" {
		return fmt.Errorf("server name cannot be empty")
	}
	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Check if server already exists
	if _, exists := config.MCPServers[serverName]; exists {
		return fmt.Errorf("server '%s' already exists in config", serverName)
	}

	// Add server entry
	config.MCPServers[serverName] = ServerEntry{
		Command: command,
		Args:    args,
		Env:     env,
	}

	return nil
}

// UpdateServer updates an existing server entry in the configuration
func (ce *ClientEditor) UpdateServer(config *ClientConfig, serverName string, command string, args []string, env map[string]string) error {
	// Validate inputs
	if serverName == "" {
		return fmt.Errorf("server name cannot be empty")
	}
	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Check if server exists
	if _, exists := config.MCPServers[serverName]; !exists {
		return fmt.Errorf("server '%s' not found in config", serverName)
	}

	// Update server entry
	config.MCPServers[serverName] = ServerEntry{
		Command: command,
		Args:    args,
		Env:     env,
	}

	return nil
}

// RemoveServer removes a server entry from the configuration
func (ce *ClientEditor) RemoveServer(config *ClientConfig, serverName string) error {
	// Validate input
	if serverName == "" {
		return fmt.Errorf("server name cannot be empty")
	}

	// Check if server exists
	if _, exists := config.MCPServers[serverName]; !exists {
		return fmt.Errorf("server '%s' not found in config", serverName)
	}

	// Remove server entry
	delete(config.MCPServers, serverName)

	return nil
}

// getClaudeDesktopConfigPath returns the path to the Claude Desktop config file
func (ce *ClientEditor) getClaudeDesktopConfigPath() string {
	// Windows: %APPDATA%\Claude\claude_desktop_config.json
	// macOS: ~/Library/Application Support/Claude/claude_desktop_config.json
	// Linux: ~/.config/Claude/claude_desktop_config.json

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Detect OS
	switch {
	case fileExists(filepath.Join(os.Getenv("APPDATA"), "Claude")):
		// Windows
		return filepath.Join(os.Getenv("APPDATA"), "Claude", "claude_desktop_config.json")
	case fileExists(filepath.Join(homeDir, "Library", "Application Support", "Claude")):
		// macOS
		return filepath.Join(homeDir, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	default:
		// Linux/Unix
		return filepath.Join(homeDir, ".config", "Claude", "claude_desktop_config.json")
	}
}

// getCursorConfigPath returns the path to the Cursor config file
func (ce *ClientEditor) getCursorConfigPath() string {
	// Cursor typically uses similar paths to VSCode
	// Windows: %APPDATA%\Cursor\User\mcp_config.json
	// macOS: ~/Library/Application Support/Cursor/User/mcp_config.json
	// Linux: ~/.config/Cursor/User/mcp_config.json

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	// Detect OS
	switch {
	case fileExists(filepath.Join(os.Getenv("APPDATA"), "Cursor")):
		// Windows
		return filepath.Join(os.Getenv("APPDATA"), "Cursor", "User", "mcp_config.json")
	case fileExists(filepath.Join(homeDir, "Library", "Application Support", "Cursor")):
		// macOS
		return filepath.Join(homeDir, "Library", "Application Support", "Cursor", "User", "mcp_config.json")
	default:
		// Linux/Unix
		return filepath.Join(homeDir, ".config", "Cursor", "User", "mcp_config.json")
	}
}

// createBackup creates a timestamped backup of the config file
func (ce *ClientEditor) createBackup(configPath string) error {
	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102-150405")
	backupPath := fmt.Sprintf("%s.backup.%s", configPath, timestamp)

	// Read original file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read original file: %w", err)
	}

	// Write backup
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// fileExists checks if a file exists at the given path
func (ce *ClientEditor) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Helper function for OS detection
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
