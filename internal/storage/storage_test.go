package storage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hoytech/mcpmanager/internal/models"
)

func TestNewFileStorage(t *testing.T) {
	storage, err := NewFileStorage()
	if err != nil {
		t.Fatalf("Failed to create file storage: %v", err)
	}
	if storage.baseDir == "" {
		t.Error("Base directory should not be empty")
	}
}

func TestFileStorage_SaveAndLoadState(t *testing.T) {
	// Use temporary directory for testing
	tmpDir := t.TempDir()
	storage := NewFileStorageWithPath(tmpDir)

	// Create a test state
	state := models.NewApplicationState()
	state.Preferences.Theme = "light"
	state.WindowLayout.Width = 1280
	state.WindowLayout.Height = 720
	server := models.NewMCPServer("test-server", tmpDir, models.DiscoveryClientConfig)
	err := state.AddDiscoveredServer(server.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Save the state
	if err := storage.SaveState(state); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Verify state.json exists
	stateFile := filepath.Join(tmpDir, "state.json")
	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		t.Error("state.json should exist after save")
	}

	// Load the state
	loadedState, err := storage.LoadState()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	// Verify data matches
	if loadedState.Preferences.Theme != "light" {
		t.Errorf("Expected theme 'light', got '%s'", loadedState.Preferences.Theme)
	}
	if loadedState.WindowLayout.Width != 1280 {
		t.Errorf("Expected width 1280, got %d", loadedState.WindowLayout.Width)
	}
	if loadedState.WindowLayout.Height != 720 {
		t.Errorf("Expected height 720, got %d", loadedState.WindowLayout.Height)
	}
	if len(loadedState.DiscoveredServers) != 1 {
		t.Errorf("Expected 1 discovered server, got %d", len(loadedState.DiscoveredServers))
	}
}

func TestFileStorage_LoadNonExistentState(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileStorageWithPath(tmpDir)

	// Load state from empty directory (should return new state)
	state, err := storage.LoadState()
	if err != nil {
		t.Fatalf("Failed to load state: %v", err)
	}

	if state == nil {
		t.Fatal("Expected new state to be returned")
	}

	// Should have default values
	if state.Preferences.Theme != "dark" {
		t.Errorf("Expected default theme 'dark', got '%s'", state.Preferences.Theme)
	}
}

func TestFileStorage_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileStorageWithPath(tmpDir)

	// Save initial state
	state1 := models.NewApplicationState()
	state1.Preferences.Theme = "dark"
	if err := storage.SaveState(state1); err != nil {
		t.Fatal(err)
	}

	// Save second state (should create backup)
	state2 := models.NewApplicationState()
	state2.Preferences.Theme = "light"
	if err := storage.SaveState(state2); err != nil {
		t.Fatal(err)
	}

	// Verify backup exists
	backupFile := filepath.Join(tmpDir, "state.json.backup")
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("Backup file should exist")
	}

	// Verify current state has new data
	loadedState, err := storage.LoadState()
	if err != nil {
		t.Fatal(err)
	}
	if loadedState.Preferences.Theme != "light" {
		t.Error("Current state should have new data")
	}
}

func TestFileStorage_SaveNilState(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileStorageWithPath(tmpDir)

	err := storage.SaveState(nil)
	if err == nil {
		t.Error("Should not allow saving nil state")
	}
}

func TestFileStorage_LoadCorruptedState(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileStorageWithPath(tmpDir)

	// Create directory
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write corrupted JSON
	stateFile := filepath.Join(tmpDir, "state.json")
	if err := os.WriteFile(stateFile, []byte("{invalid json}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Try to load
	_, err := storage.LoadState()
	if err == nil {
		t.Error("Should fail to load corrupted state")
	}
}

func TestFileStorage_LoadFromBackup(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileStorageWithPath(tmpDir)

	// Create directory
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a valid backup
	state := models.NewApplicationState()
	state.Preferences.Theme = "backup-theme"
	_ = storage.SaveState(state)

	// Copy to backup
	stateFile := filepath.Join(tmpDir, "state.json")
	backupFile := filepath.Join(tmpDir, "state.json.backup")
	content, _ := os.ReadFile(stateFile)
	_ = os.WriteFile(backupFile, content, 0644)

	// Make main file unreadable (permission error triggers backup fallback)
	_ = os.Chmod(stateFile, 0000)
	defer os.Chmod(stateFile, 0644) // Restore for cleanup

	// Should load from backup
	loadedState, err := storage.LoadState()
	if err != nil {
		t.Fatalf("Should load from backup: %v", err)
	}

	if loadedState.Preferences.Theme != "backup-theme" {
		t.Error("Should have loaded from backup")
	}
}

func TestFileStorage_SaveAndLoadServerLogs(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileStorageWithPath(tmpDir)

	serverID := "test-server-123"

	// Create test logs
	logs := []models.LogEntry{
		*models.NewLogEntry(models.LogInfo, serverID, "Test message 1"),
		*models.NewLogEntry(models.LogError, serverID, "Test message 2"),
		*models.NewLogEntry(models.LogWarning, serverID, "Test message 3"),
	}

	// Save logs
	if err := storage.SaveServerLogs(serverID, logs); err != nil {
		t.Fatalf("Failed to save logs: %v", err)
	}

	// Verify logs directory exists
	serverDir := filepath.Join(tmpDir, "servers", serverID)
	if _, err := os.Stat(serverDir); os.IsNotExist(err) {
		t.Error("Server directory should exist")
	}

	// Load logs
	loadedLogs, err := storage.LoadServerLogs(serverID)
	if err != nil {
		t.Fatalf("Failed to load logs: %v", err)
	}

	// Verify data matches
	if len(loadedLogs) != 3 {
		t.Errorf("Expected 3 logs, got %d", len(loadedLogs))
	}

	for i, log := range loadedLogs {
		if log.Message != logs[i].Message {
			t.Errorf("Log %d message mismatch", i)
		}
		if log.Severity != logs[i].Severity {
			t.Errorf("Log %d severity mismatch", i)
		}
	}
}

func TestFileStorage_LoadNonExistentServerLogs(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileStorageWithPath(tmpDir)

	// Load logs for non-existent server (should return empty slice)
	logs, err := storage.LoadServerLogs("nonexistent-server")
	if err != nil {
		t.Fatalf("Should not error on non-existent logs: %v", err)
	}

	if len(logs) != 0 {
		t.Error("Should return empty slice for non-existent logs")
	}
}

func TestFileStorage_SaveLogsEmptyServerID(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileStorageWithPath(tmpDir)

	err := storage.SaveServerLogs("", []models.LogEntry{})
	if err == nil {
		t.Error("Should not allow empty server ID")
	}
}

func TestFileStorage_LoadLogsEmptyServerID(t *testing.T) {
	tmpDir := t.TempDir()
	storage := NewFileStorageWithPath(tmpDir)

	_, err := storage.LoadServerLogs("")
	if err == nil {
		t.Error("Should not allow empty server ID")
	}
}
