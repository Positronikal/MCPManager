package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
)

func TestNewConfigService(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	cs, err := NewConfigService(eventBus)
	if err != nil {
		t.Fatalf("NewConfigService failed: %v", err)
	}

	if cs == nil {
		t.Fatal("Expected non-nil ConfigService")
	}

	if cs.baseDir == "" {
		t.Error("Expected baseDir to be set")
	}

	if cs.eventBus == nil {
		t.Error("Expected eventBus to be set")
	}
}

func TestNewConfigServiceWithPath(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	if cs == nil {
		t.Fatal("Expected non-nil ConfigService")
	}

	if cs.baseDir != testDir {
		t.Errorf("Expected baseDir to be %s, got %s", testDir, cs.baseDir)
	}
}

func TestGetConfiguration_NotExists(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	config, err := cs.GetConfiguration("test-server-1")
	if err != nil {
		t.Fatalf("GetConfiguration failed: %v", err)
	}

	if config == nil {
		t.Fatal("Expected non-nil default configuration")
	}

	// Check default values
	if config.AutoStart != false {
		t.Error("Expected AutoStart to be false by default")
	}
	if config.MaxRestartAttempts != 3 {
		t.Errorf("Expected MaxRestartAttempts to be 3, got %d", config.MaxRestartAttempts)
	}
	if config.StartupTimeout != 30 {
		t.Errorf("Expected StartupTimeout to be 30, got %d", config.StartupTimeout)
	}
}

func TestGetConfiguration_EmptyServerID(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	_, err := cs.GetConfiguration("")
	if err == nil {
		t.Error("Expected error for empty serverID")
	}
}

func TestUpdateConfiguration_Success(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	// Subscribe to config change events
	eventChan := eventBus.Subscribe(events.EventConfigFileChanged)

	serverID := "test-server-1"
	config := models.NewServerConfiguration()
	config.AutoStart = true
	config.MaxRestartAttempts = 5
	config.EnvironmentVariables = map[string]string{
		"TEST_VAR": "test_value",
	}

	// Update configuration
	err := cs.UpdateConfiguration(serverID, config)
	if err != nil {
		t.Fatalf("UpdateConfiguration failed: %v", err)
	}

	// Verify file was created
	configFile := cs.getConfigFilePath(serverID)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Error("Expected configuration file to be created")
	}

	// Verify event was published
	select {
	case event := <-eventChan:
		if event.Type != events.EventConfigFileChanged {
			t.Errorf("Expected EventConfigFileChanged, got %s", event.Type)
		}
		filePath, ok := event.Data["filePath"].(string)
		if !ok || filePath != configFile {
			t.Errorf("Expected filePath %s in event, got %v", configFile, event.Data["filePath"])
		}
	default:
		t.Error("Expected config changed event to be published")
	}

	// Read back and verify
	loadedConfig, err := cs.GetConfiguration(serverID)
	if err != nil {
		t.Fatalf("GetConfiguration failed: %v", err)
	}

	if !loadedConfig.AutoStart {
		t.Error("Expected AutoStart to be true")
	}
	if loadedConfig.MaxRestartAttempts != 5 {
		t.Errorf("Expected MaxRestartAttempts to be 5, got %d", loadedConfig.MaxRestartAttempts)
	}
	if loadedConfig.EnvironmentVariables["TEST_VAR"] != "test_value" {
		t.Errorf("Expected TEST_VAR to be test_value, got %s", loadedConfig.EnvironmentVariables["TEST_VAR"])
	}
}

func TestUpdateConfiguration_EmptyServerID(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	config := models.NewServerConfiguration()
	err := cs.UpdateConfiguration("", config)
	if err == nil {
		t.Error("Expected error for empty serverID")
	}
}

func TestUpdateConfiguration_NilConfig(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	err := cs.UpdateConfiguration("test-server-1", nil)
	if err == nil {
		t.Error("Expected error for nil configuration")
	}
}

func TestUpdateConfiguration_InvalidConfig(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	config := models.NewServerConfiguration()
	config.MaxRestartAttempts = -1 // Invalid value

	err := cs.UpdateConfiguration("test-server-1", config)
	if err == nil {
		t.Error("Expected error for invalid configuration")
	}
}

func TestUpdateConfiguration_InvalidEnvVar(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	config := models.NewServerConfiguration()
	config.EnvironmentVariables = map[string]string{
		"invalid-var": "value", // Invalid env var name (contains hyphen)
	}

	err := cs.UpdateConfiguration("test-server-1", config)
	if err == nil {
		t.Error("Expected error for invalid environment variable name")
	}
}

func TestUpdateConfiguration_Backup(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	serverID := "test-server-1"

	// Create initial configuration
	config1 := models.NewServerConfiguration()
	config1.AutoStart = true
	err := cs.UpdateConfiguration(serverID, config1)
	if err != nil {
		t.Fatalf("First UpdateConfiguration failed: %v", err)
	}

	// Update with new configuration
	config2 := models.NewServerConfiguration()
	config2.AutoStart = false
	config2.MaxRestartAttempts = 7
	err = cs.UpdateConfiguration(serverID, config2)
	if err != nil {
		t.Fatalf("Second UpdateConfiguration failed: %v", err)
	}

	// Verify backup file exists
	backupFile := cs.getConfigFilePath(serverID) + ".backup"
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		t.Error("Expected backup file to be created")
	}

	// Current config should have new values
	currentConfig, err := cs.GetConfiguration(serverID)
	if err != nil {
		t.Fatalf("GetConfiguration failed: %v", err)
	}
	if currentConfig.AutoStart {
		t.Error("Expected AutoStart to be false in current config")
	}
	if currentConfig.MaxRestartAttempts != 7 {
		t.Errorf("Expected MaxRestartAttempts to be 7, got %d", currentConfig.MaxRestartAttempts)
	}
}

func TestValidateConfiguration_Valid(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	config := models.NewServerConfiguration()
	config.MaxRestartAttempts = 5
	config.StartupTimeout = 60
	config.ShutdownTimeout = 20

	err := cs.ValidateConfiguration(config)
	if err != nil {
		t.Errorf("ValidateConfiguration failed for valid config: %v", err)
	}
}

func TestValidateConfiguration_Nil(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	err := cs.ValidateConfiguration(nil)
	if err == nil {
		t.Error("Expected error for nil configuration")
	}
}

func TestValidateConfiguration_InvalidMaxRestartAttempts(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	testCases := []int{-1, 11, 100}
	for _, maxRestarts := range testCases {
		config := models.NewServerConfiguration()
		config.MaxRestartAttempts = maxRestarts

		err := cs.ValidateConfiguration(config)
		if err == nil {
			t.Errorf("Expected error for MaxRestartAttempts=%d", maxRestarts)
		}
	}
}

func TestValidateConfiguration_InvalidTimeouts(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	// Invalid startup timeout
	config1 := models.NewServerConfiguration()
	config1.StartupTimeout = 0
	err := cs.ValidateConfiguration(config1)
	if err == nil {
		t.Error("Expected error for zero StartupTimeout")
	}

	// Invalid shutdown timeout
	config2 := models.NewServerConfiguration()
	config2.ShutdownTimeout = -5
	err = cs.ValidateConfiguration(config2)
	if err == nil {
		t.Error("Expected error for negative ShutdownTimeout")
	}
}

func TestDeleteConfiguration_Success(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	serverID := "test-server-1"

	// Create configuration
	config := models.NewServerConfiguration()
	err := cs.UpdateConfiguration(serverID, config)
	if err != nil {
		t.Fatalf("UpdateConfiguration failed: %v", err)
	}

	// Verify file exists
	configFile := cs.getConfigFilePath(serverID)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Configuration file should exist")
	}

	// Delete configuration
	err = cs.DeleteConfiguration(serverID)
	if err != nil {
		t.Fatalf("DeleteConfiguration failed: %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		t.Error("Configuration file should be deleted")
	}
}

func TestDeleteConfiguration_NotExists(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	// Delete non-existent configuration should succeed
	err := cs.DeleteConfiguration("non-existent-server")
	if err != nil {
		t.Errorf("DeleteConfiguration should succeed for non-existent file: %v", err)
	}
}

func TestDeleteConfiguration_EmptyServerID(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	err := cs.DeleteConfiguration("")
	if err == nil {
		t.Error("Expected error for empty serverID")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	serverID := "test-server-roundtrip"

	// Create working directory for the test
	workDir := t.TempDir()

	// Create a complex configuration
	config := models.NewServerConfiguration()
	config.AutoStart = true
	config.RestartOnCrash = true
	config.MaxRestartAttempts = 8
	config.StartupTimeout = 45
	config.ShutdownTimeout = 15
	config.HealthCheckInterval = 30
	config.HealthCheckEndpoint = "http://localhost:8080/health"
	config.EnvironmentVariables = map[string]string{
		"NODE_ENV":   "production",
		"LOG_LEVEL":  "debug",
		"API_KEY":    "secret123",
	}
	config.CommandLineArguments = []string{"--verbose", "--port", "8080"}
	config.WorkingDirectory = workDir // Use temp dir as working directory

	// Save configuration
	err := cs.UpdateConfiguration(serverID, config)
	if err != nil {
		t.Fatalf("UpdateConfiguration failed: %v", err)
	}

	// Load configuration
	loadedConfig, err := cs.GetConfiguration(serverID)
	if err != nil {
		t.Fatalf("GetConfiguration failed: %v", err)
	}

	// Verify all fields
	if loadedConfig.AutoStart != config.AutoStart {
		t.Error("AutoStart mismatch")
	}
	if loadedConfig.RestartOnCrash != config.RestartOnCrash {
		t.Error("RestartOnCrash mismatch")
	}
	if loadedConfig.MaxRestartAttempts != config.MaxRestartAttempts {
		t.Error("MaxRestartAttempts mismatch")
	}
	if loadedConfig.StartupTimeout != config.StartupTimeout {
		t.Error("StartupTimeout mismatch")
	}
	if loadedConfig.ShutdownTimeout != config.ShutdownTimeout {
		t.Error("ShutdownTimeout mismatch")
	}
	if loadedConfig.HealthCheckInterval != config.HealthCheckInterval {
		t.Error("HealthCheckInterval mismatch")
	}
	if loadedConfig.HealthCheckEndpoint != config.HealthCheckEndpoint {
		t.Error("HealthCheckEndpoint mismatch")
	}
	if len(loadedConfig.EnvironmentVariables) != len(config.EnvironmentVariables) {
		t.Error("EnvironmentVariables count mismatch")
	}
	for key, value := range config.EnvironmentVariables {
		if loadedConfig.EnvironmentVariables[key] != value {
			t.Errorf("EnvironmentVariable %s mismatch", key)
		}
	}
	if len(loadedConfig.CommandLineArguments) != len(config.CommandLineArguments) {
		t.Error("CommandLineArguments count mismatch")
	}
	for i, arg := range config.CommandLineArguments {
		if loadedConfig.CommandLineArguments[i] != arg {
			t.Errorf("CommandLineArgument[%d] mismatch", i)
		}
	}
	if loadedConfig.WorkingDirectory != config.WorkingDirectory {
		t.Error("WorkingDirectory mismatch")
	}
}

func TestConcurrentAccess(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	testDir := filepath.Join(t.TempDir(), "mcpmanager")
	cs := NewConfigServiceWithPath(testDir, eventBus)

	serverID := "test-server-concurrent"

	// Initial config
	initialConfig := models.NewServerConfiguration()
	err := cs.UpdateConfiguration(serverID, initialConfig)
	if err != nil {
		t.Fatalf("Initial UpdateConfiguration failed: %v", err)
	}

	// Test concurrent reads and writes
	done := make(chan bool)
	errors := make(chan error, 20)

	// Start 10 concurrent readers
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			_, err := cs.GetConfiguration(serverID)
			if err != nil {
				errors <- err
			}
		}()
	}

	// Start 10 concurrent writers
	for i := 0; i < 10; i++ {
		go func(attempt int) {
			defer func() { done <- true }()
			config := models.NewServerConfiguration()
			config.MaxRestartAttempts = attempt % 10
			err := cs.UpdateConfiguration(serverID, config)
			if err != nil {
				errors <- err
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}
