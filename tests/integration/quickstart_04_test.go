package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/models"
)

// TestQuickstart04_ConfigurationEditing tests the fourth quickstart scenario:
// Configuration editing via API
// Requirements: FR-018 (edit config), FR-019 (CRITICAL: never modify client config files)
func TestQuickstart04_ConfigurationEditing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary test environment
	tempDir, err := os.MkdirTemp("", "mcp-quickstart-04-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create mock client config with test server
	configDir := createMockConfigDir(t, tempDir)
	createMockClientConfig(t, configDir, []map[string]interface{}{
		{
			"name":    "config-test-server",
			"command": "node",
			"args":    []interface{}{"server.js"},
			"env": map[string]interface{}{
				"ORIGINAL_VAR": "original_value",
			},
		},
	})

	// Calculate checksum of original client config file (FR-019)
	clientConfigPath := getClientConfigPath(configDir)
	originalChecksum, err := calculateFileChecksum(clientConfigPath)
	if err != nil {
		t.Fatalf("Failed to calculate config file checksum: %v", err)
	}
	t.Logf("Original client config checksum: %x", originalChecksum)

	// Start MCP Manager
	appPath := buildMCPManager(t)
	appStateDir := filepath.Join(tempDir, ".mcpmanager")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, appPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("MCP_CONFIG_DIR=%s", configDir),
		fmt.Sprintf("MCP_STATE_DIR=%s", appStateDir),
		"HEADLESS=true",
	)

	t.Log("Starting MCP Manager...")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start MCP Manager: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Wait for application to start
	apiURL := "http://localhost:8080"
	if !waitForHTTPReady(apiURL+"/health", 10*time.Second) {
		t.Fatal("MCP Manager did not start within timeout")
	}

	// Test Step 1: Discover test server
	var serverID string
	t.Log("Waiting for server discovery...")
	if !waitForCondition(15*time.Second, 500*time.Millisecond, func() bool {
		resp, err := http.Get(apiURL + "/api/servers")
		if err != nil {
			return false
		}
		defer resp.Body.Close()

		var servers []models.MCPServer
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &servers); err != nil {
			return false
		}

		for _, server := range servers {
			if server.Name == "config-test-server" {
				serverID = server.ID
				t.Logf("Found test server: ID=%s", serverID)
				return true
			}
		}
		return false
	}) {
		t.Fatal("Test server not discovered within timeout")
	}

	// Test Step 2: Call GET /servers/{id}/configuration
	t.Log("Fetching original configuration...")
	configURL := fmt.Sprintf("%s/api/servers/%s/configuration", apiURL, serverID)

	resp, err := http.Get(configURL)
	if err != nil {
		t.Fatalf("Failed to get configuration: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var originalConfig models.ServerConfiguration
	if err := json.Unmarshal(body, &originalConfig); err != nil {
		t.Fatalf("Failed to unmarshal configuration: %v", err)
	}

	t.Logf("Original config - Env vars: %v", originalConfig.EnvironmentVariables)

	// Verify original environment variable exists
	if val, exists := originalConfig.EnvironmentVariables["ORIGINAL_VAR"]; !exists || val != "original_value" {
		t.Errorf("Expected ORIGINAL_VAR=original_value, got %v", val)
	}

	// Test Step 3: Modify config - Add env var TEST_VAR=test_value
	t.Log("Modifying configuration: Adding TEST_VAR=test_value")
	modifiedConfig := originalConfig
	if modifiedConfig.EnvironmentVariables == nil {
		modifiedConfig.EnvironmentVariables = make(map[string]string)
	}
	modifiedConfig.EnvironmentVariables["TEST_VAR"] = "test_value"

	configJSON, err := json.Marshal(modifiedConfig)
	if err != nil {
		t.Fatalf("Failed to marshal modified config: %v", err)
	}

	// Test Step 4: Call PUT /servers/{id}/configuration with updated config
	t.Log("Sending updated configuration...")
	req, err := http.NewRequest(http.MethodPut, configURL, bytes.NewBuffer(configJSON))
	if err != nil {
		t.Fatalf("Failed to create PUT request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	putResp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to PUT configuration: %v", err)
	}
	defer putResp.Body.Close()

	// Test Step 5: Verify 200 response
	if putResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(putResp.Body)
		t.Fatalf("Expected status 200 for PUT, got %d. Body: %s", putResp.StatusCode, body)
	}
	t.Log("✓ Configuration updated successfully (status 200)")

	// Test Step 6: Call GET /servers/{id}/configuration again
	t.Log("Fetching configuration to verify persistence...")
	time.Sleep(500 * time.Millisecond) // Brief delay to ensure write completes

	getResp, err := http.Get(configURL)
	if err != nil {
		t.Fatalf("Failed to get configuration: %v", err)
	}
	defer getResp.Body.Close()

	body, err = io.ReadAll(getResp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	var updatedConfig models.ServerConfiguration
	if err := json.Unmarshal(body, &updatedConfig); err != nil {
		t.Fatalf("Failed to unmarshal configuration: %v", err)
	}

	// Test Step 7: Verify TEST_VAR persisted
	if val, exists := updatedConfig.EnvironmentVariables["TEST_VAR"]; !exists {
		t.Error("Expected TEST_VAR to be persisted, but it's not present")
	} else if val != "test_value" {
		t.Errorf("Expected TEST_VAR=test_value, got TEST_VAR=%s", val)
	} else {
		t.Log("✓ TEST_VAR persisted correctly")
	}

	// Verify original variable still exists
	if val, exists := updatedConfig.EnvironmentVariables["ORIGINAL_VAR"]; !exists {
		t.Error("Original ORIGINAL_VAR was lost")
	} else if val != "original_value" {
		t.Errorf("ORIGINAL_VAR changed: expected 'original_value', got '%s'", val)
	} else {
		t.Log("✓ Original ORIGINAL_VAR unchanged")
	}

	// Test Step 8: Verify client config file UNCHANGED (FR-019 CRITICAL CHECK)
	t.Log("Verifying client config file unchanged (FR-019)...")
	currentChecksum, err := calculateFileChecksum(clientConfigPath)
	if err != nil {
		t.Fatalf("Failed to calculate current config file checksum: %v", err)
	}
	t.Logf("Current client config checksum:  %x", currentChecksum)

	if !bytes.Equal(originalChecksum, currentChecksum) {
		t.Error("FAIL FR-019: Client config file was modified! This is a CRITICAL requirement violation.")
		t.Error("MCP Manager MUST store modified configurations separately, not modify client config files.")

		// Show what changed
		originalContent, _ := os.ReadFile(clientConfigPath)
		t.Logf("Client config content:\n%s", originalContent)
	} else {
		t.Log("✓ PASS FR-019: Client config file unchanged (critical requirement met)")
	}

	// Additional verification: Check if modifications are stored elsewhere
	// Modified configs should be stored in application state directory
	modifiedConfigPath := filepath.Join(appStateDir, "modified_configs", serverID+".json")
	if _, err := os.Stat(modifiedConfigPath); err == nil {
		t.Logf("✓ Modified configuration stored separately at: %s", modifiedConfigPath)
	} else {
		t.Logf("Note: Modified config not found at expected path: %s (may be stored differently)", modifiedConfigPath)
	}

	t.Log("✓ Quickstart Test 4 completed successfully")
}

// calculateFileChecksum calculates SHA-256 checksum of a file
func calculateFileChecksum(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	hash := sha256.Sum256(data)
	return hash[:], nil
}

// getClientConfigPath returns the platform-specific client config file path
func getClientConfigPath(configDir string) string {
	var configPath string
	switch runtime.GOOS {
	case "windows":
		configPath = filepath.Join(configDir, "Claude", "claude_desktop_config.json")
	case "darwin":
		configPath = filepath.Join(configDir, "Claude", "claude_desktop_config.json")
	default:
		configPath = filepath.Join(configDir, "Claude", "claude_desktop_config.json")
	}
	return configPath
}
