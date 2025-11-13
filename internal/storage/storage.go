package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Positronikal/MCPManager/internal/models"
	"github.com/Positronikal/MCPManager/internal/platform"
)

// StorageService defines the interface for persisting application state
type StorageService interface {
	LoadState() (*models.ApplicationState, error)
	SaveState(state *models.ApplicationState) error
	LoadServerLogs(serverID string) ([]models.LogEntry, error)
	SaveServerLogs(serverID string, logs []models.LogEntry) error
}

// FileStorage implements StorageService using JSON files
type FileStorage struct {
	baseDir string
}

// NewFileStorage creates a new file storage instance
// Uses ~/.mcpmanager as the base directory
func NewFileStorage() (*FileStorage, error) {
	baseDir := platform.GetMCPManagerDir()
	if baseDir == "" {
		return nil, fmt.Errorf("could not determine MCP Manager directory")
	}

	return &FileStorage{
		baseDir: baseDir,
	}, nil
}

// NewFileStorageWithPath creates a new file storage instance with a custom base directory
func NewFileStorageWithPath(baseDir string) *FileStorage {
	return &FileStorage{
		baseDir: baseDir,
	}
}

// ensureDir creates the directory if it doesn't exist
func (fs *FileStorage) ensureDir(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// LoadState loads the application state from state.json
func (fs *FileStorage) LoadState() (*models.ApplicationState, error) {
	stateFile := filepath.Join(fs.baseDir, "state.json")

	// If file doesn't exist, return a new state
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		return models.NewApplicationState(), nil
	}

	// Read the file
	data, err := os.ReadFile(stateFile)
	if err != nil {
		// Try to load from backup
		backupFile := filepath.Join(fs.baseDir, "state.json.backup")
		if _, backupErr := os.Stat(backupFile); backupErr == nil {
			data, err = os.ReadFile(backupFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load state from backup: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load state: %w", err)
		}
	}

	// Unmarshal
	var state models.ApplicationState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}

	return &state, nil
}

// SaveState saves the application state to state.json atomically
func (fs *FileStorage) SaveState(state *models.ApplicationState) error {
	if state == nil {
		return fmt.Errorf("state cannot be nil")
	}

	// Ensure base directory exists
	if err := fs.ensureDir(fs.baseDir); err != nil {
		return err
	}

	stateFile := filepath.Join(fs.baseDir, "state.json")
	tmpFile := filepath.Join(fs.baseDir, "state.json.tmp")
	backupFile := filepath.Join(fs.baseDir, "state.json.backup")

	// Create backup of existing file
	if _, err := os.Stat(stateFile); err == nil {
		// Copy current state to backup
		data, err := os.ReadFile(stateFile)
		if err == nil {
			_ = os.WriteFile(backupFile, data, 0644) // Ignore error - backup is best effort
		}
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temporary file
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomically rename temporary file to final file
	if err := os.Rename(tmpFile, stateFile); err != nil {
		// Clean up temporary file on error
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// LoadServerLogs loads logs for a specific server
func (fs *FileStorage) LoadServerLogs(serverID string) ([]models.LogEntry, error) {
	if serverID == "" {
		return nil, fmt.Errorf("serverID cannot be empty")
	}

	logFile := filepath.Join(fs.baseDir, "servers", serverID, "logs.json")

	// If file doesn't exist, return empty slice
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return []models.LogEntry{}, nil
	}

	// Read the file
	data, err := os.ReadFile(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load logs: %w", err)
	}

	// Unmarshal
	var logs []models.LogEntry
	if err := json.Unmarshal(data, &logs); err != nil {
		return nil, fmt.Errorf("failed to parse logs: %w", err)
	}

	return logs, nil
}

// SaveServerLogs saves logs for a specific server atomically
func (fs *FileStorage) SaveServerLogs(serverID string, logs []models.LogEntry) error {
	if serverID == "" {
		return fmt.Errorf("serverID cannot be empty")
	}

	serverDir := filepath.Join(fs.baseDir, "servers", serverID)
	if err := fs.ensureDir(serverDir); err != nil {
		return err
	}

	logFile := filepath.Join(serverDir, "logs.json")
	tmpFile := filepath.Join(serverDir, "logs.json.tmp")

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(logs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal logs: %w", err)
	}

	// Write to temporary file
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomically rename temporary file to final file
	if err := os.Rename(tmpFile, logFile); err != nil {
		// Clean up temporary file on error
		_ = os.Remove(tmpFile)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}
