package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
	"github.com/hoytech/mcpmanager/internal/platform"
)

// ConfigService manages server configurations
type ConfigService struct {
	baseDir  string
	eventBus *events.EventBus
	mu       sync.RWMutex
}

// NewConfigService creates a new configuration service
func NewConfigService(eventBus *events.EventBus) (*ConfigService, error) {
	baseDir := platform.GetMCPManagerDir()
	if baseDir == "" {
		return nil, fmt.Errorf("could not determine MCP Manager directory")
	}

	return &ConfigService{
		baseDir:  baseDir,
		eventBus: eventBus,
	}, nil
}

// NewConfigServiceWithPath creates a new configuration service with a custom base directory
// Useful for testing
func NewConfigServiceWithPath(baseDir string, eventBus *events.EventBus) *ConfigService {
	return &ConfigService{
		baseDir:  baseDir,
		eventBus: eventBus,
	}
}

// GetConfiguration loads the configuration for a specific server from disk
func (cs *ConfigService) GetConfiguration(serverID string) (*models.ServerConfiguration, error) {
	if serverID == "" {
		return nil, fmt.Errorf("serverID cannot be empty")
	}

	cs.mu.RLock()
	defer cs.mu.RUnlock()

	configFile := cs.getConfigFilePath(serverID)

	// If file doesn't exist, return default configuration
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return models.NewServerConfiguration(), nil
	}

	// Read the file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Unmarshal
	var config models.ServerConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse configuration: %w", err)
	}

	return &config, nil
}

// UpdateConfiguration updates the configuration for a specific server
// This validates the configuration before saving it to disk
func (cs *ConfigService) UpdateConfiguration(serverID string, config *models.ServerConfiguration) error {
	if serverID == "" {
		return fmt.Errorf("serverID cannot be empty")
	}

	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	// Validate configuration first
	if err := cs.ValidateConfiguration(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Ensure server directory exists
	serverDir := cs.getServerDir(serverID)
	if err := cs.ensureDir(serverDir); err != nil {
		return err
	}

	configFile := cs.getConfigFilePath(serverID)
	tmpFile := configFile + ".tmp"
	backupFile := configFile + ".backup"

	// Create backup of existing file if it exists
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err == nil {
			_ = os.WriteFile(backupFile, data, 0644) // Backup is best effort
		}
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write to temporary file
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomically rename temporary file to final file
	if err := os.Rename(tmpFile, configFile); err != nil {
		// Clean up temporary file on error
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	// Publish configuration changed event
	if cs.eventBus != nil {
		cs.eventBus.Publish(events.ConfigFileChangedEvent(configFile))
	}

	return nil
}

// ValidateConfiguration validates a server configuration
// This delegates to the configuration's own Validate method
func (cs *ConfigService) ValidateConfiguration(config *models.ServerConfiguration) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	return config.Validate()
}

// DeleteConfiguration removes the configuration file for a specific server
func (cs *ConfigService) DeleteConfiguration(serverID string) error {
	if serverID == "" {
		return fmt.Errorf("serverID cannot be empty")
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	configFile := cs.getConfigFilePath(serverID)

	// If file doesn't exist, consider it already deleted
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil
	}

	// Remove the file
	if err := os.Remove(configFile); err != nil {
		return fmt.Errorf("failed to delete configuration file: %w", err)
	}

	return nil
}

// getServerDir returns the directory for a specific server
func (cs *ConfigService) getServerDir(serverID string) string {
	return filepath.Join(cs.baseDir, "servers", serverID)
}

// getConfigFilePath returns the path to the configuration file for a specific server
func (cs *ConfigService) getConfigFilePath(serverID string) string {
	return filepath.Join(cs.getServerDir(serverID), "config.json")
}

// ensureDir creates the directory if it doesn't exist
func (cs *ConfigService) ensureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}
