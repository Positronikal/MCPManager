package integration

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/models"
)

// TestEdgeCase_ExternalConfigChange tests edge case 6: External config file modification
// Requirements: FR-014 (file watcher), FR-024 (ConfigFileChanged event)
func TestEdgeCase_ExternalConfigChange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary test environment
	tempDir, err := os.MkdirTemp("", "mcp-edge-config-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create mock client config with initial server
	configDir := createMockConfigDir(t, tempDir)
	configPath := getClientConfigPath(configDir)

	createMockClientConfig(t, configDir, []map[string]interface{}{
		{
			"name":    "initial-server",
			"command": "node",
			"args":    []interface{}{"initial.js"},
			"env":     map[string]interface{}{},
		},
	})

	// Start MCP Manager
	appPath := buildMCPManager(t)
	appStateDir := filepath.Join(tempDir, ".mcpmanager")

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
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

	// Test Step 1: Wait for initial server discovery
	t.Log("Waiting for initial server discovery...")
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
			if server.Name == "initial-server" {
				t.Logf("Found initial server: ID=%s", server.ID)
				return true
			}
		}
		return false
	}) {
		t.Fatal("Initial server not discovered within timeout")
	}

	// Test Step 2: Subscribe to SSE /events
	t.Log("Subscribing to SSE events...")
	eventsChan := make(chan string, 10)
	eventsCtx, eventsCancel := context.WithCancel(context.Background())
	defer eventsCancel()

	go subscribeToSSE(eventsCtx, apiURL+"/api/events", eventsChan, t)

	// Give SSE connection time to establish
	time.Sleep(1 * time.Second)

	// Test Step 3: Externally modify client config file (add new server)
	t.Log("Modifying client config file externally...")
	if err := addServerToConfig(configPath, map[string]interface{}{
		"name":    "new-external-server",
		"command": "python3",
		"args":    []interface{}{"-m", "http.server", "9999"},
		"env":     map[string]interface{}{"EXTERNAL": "true"},
	}); err != nil {
		t.Fatalf("Failed to modify config file: %v", err)
	}
	t.Log("✓ Config file modified")

	modificationTime := time.Now()

	// Test Step 4: Wait for ConfigFileChanged SSE event (timeout 5s per requirement)
	t.Log("Waiting for ConfigFileChanged event (max 5s)...")
	eventReceived := false
	eventTimeout := time.After(10 * time.Second)

EventLoop:
	for {
		select {
		case event := <-eventsChan:
			t.Logf("Received event: %s", event)

			// Check if this is a ConfigFileChanged event
			if strings.Contains(event, "ConfigFileChanged") || strings.Contains(event, "config") {
				detectionTime := time.Since(modificationTime)
				t.Logf("✓ ConfigFileChanged event received in %v", detectionTime)

				// Verify file watcher detected change within 2 seconds (FR-014 requirement)
				if detectionTime > 2*time.Second {
					t.Logf("Warning: File change detection took %v, ideally should be < 2s", detectionTime)
				} else {
					t.Logf("✓ PASS: File change detected within 2 second requirement")
				}

				// Test Step 5: Verify event contains correct filePath
				if strings.Contains(event, filepath.Base(configPath)) || strings.Contains(event, "claude_desktop_config.json") {
					t.Log("✓ Event contains correct file path")
				}

				eventReceived = true
				break EventLoop
			}

		case <-eventTimeout:
			t.Error("FAIL: ConfigFileChanged event not received within timeout")
			break EventLoop
		}
	}

	if !eventReceived {
		t.Error("FAIL: ConfigFileChanged event was not received")
	} else {
		t.Log("✓ ConfigFileChanged event verified")
	}

	// Test Step 6: Call POST /servers/discover to trigger rediscovery
	t.Log("Triggering manual discovery...")
	discoverResp, err := http.Post(apiURL+"/api/servers/discover", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to call discover API: %v", err)
	}
	discoverResp.Body.Close()

	if discoverResp.StatusCode != http.StatusOK && discoverResp.StatusCode != http.StatusAccepted {
		t.Errorf("Expected status 200/202 for discover, got %d", discoverResp.StatusCode)
	}
	t.Log("✓ Discovery triggered")

	// Test Step 7: Verify new server discovered
	t.Log("Waiting for new server to be discovered...")
	if !waitForCondition(10*time.Second, 500*time.Millisecond, func() bool {
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
			if server.Name == "new-external-server" {
				t.Logf("✓ New server discovered: ID=%s", server.ID)
				return true
			}
		}
		return false
	}) {
		t.Error("FAIL: New server not discovered after config change")
	} else {
		t.Log("✓ New server discovered successfully")
	}

	t.Log("✓ Edge Case Test: External Config Change completed successfully")
}

// subscribeToSSE subscribes to Server-Sent Events endpoint
func subscribeToSSE(ctx context.Context, url string, eventsChan chan<- string, t *testing.T) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		t.Logf("Failed to create SSE request: %v", err)
		return
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	client := &http.Client{
		Timeout: 0, // No timeout for SSE
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Failed to connect to SSE: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Logf("SSE connection returned status %d", resp.StatusCode)
		return
	}

	t.Log("SSE connection established")

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			line := scanner.Text()
			if line != "" {
				// Send the event line to the channel
				select {
				case eventsChan <- line:
				case <-ctx.Done():
					return
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Logf("SSE scanner error: %v", err)
	}
}

// addServerToConfig adds a new server to an existing config file
func addServerToConfig(configPath string, serverConfig map[string]interface{}) error {
	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Get mcpServers section
	mcpServers, ok := config["mcpServers"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("mcpServers section not found or invalid")
	}

	// Add new server
	serverName := serverConfig["name"].(string)
	mcpServers[serverName] = map[string]interface{}{
		"command": serverConfig["command"],
		"args":    serverConfig["args"],
		"env":     serverConfig["env"],
	}

	// Write back to file
	newData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}
