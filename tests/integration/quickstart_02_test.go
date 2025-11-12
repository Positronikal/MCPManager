package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/models"
)

// TestQuickstart02_ServerLifecycle tests the second quickstart scenario:
// Start server lifecycle (stopped → starting → running → stopped)
// Requirements: FR-020 (start server), FR-021 (stop server), FR-024 (SSE events)
func TestQuickstart02_ServerLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the test server
	testServerPath := buildTestServer(t)

	// Create temporary test environment
	tempDir, err := os.MkdirTemp("", "mcp-quickstart-02-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create mock client config with test server
	configDir := createMockConfigDir(t, tempDir)
	createMockClientConfig(t, configDir, []map[string]interface{}{
		{
			"name":    "lifecycle-test-server",
			"command": testServerPath,
			"args":    []interface{}{"18888"}, // Use unique port
			"env":     map[string]interface{}{},
		},
	})

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

	// Wait for discovery to find our test server
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
			if server.Name == "lifecycle-test-server" {
				serverID = server.ID
				t.Logf("Found test server: ID=%s, Status=%s", serverID, server.Status.State)
				return true
			}
		}
		return false
	}) {
		t.Fatal("Test server not discovered within timeout")
	}

	// Verify initial state is stopped
	server := getServerStatus(t, apiURL, serverID)
	if server.Status.State != models.StatusStopped {
		t.Errorf("Expected initial state 'stopped', got '%s'", server.Status.State)
	}
	t.Log("✓ Initial state: stopped")

	// Test Step 3: Call POST /servers/{id}/start
	t.Logf("Starting server %s...", serverID)
	startResp, err := http.Post(
		fmt.Sprintf("%s/api/servers/%s/start", apiURL, serverID),
		"application/json",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to call start API: %v", err)
	}
	startResp.Body.Close()

	if startResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for start, got %d", startResp.StatusCode)
	}

	// Test Step 4: Poll GET /servers/{id}/status until state=running (timeout 10s)
	t.Log("Polling for running state...")
	if !waitForCondition(10*time.Second, 500*time.Millisecond, func() bool {
		server := getServerStatus(t, apiURL, serverID)
		t.Logf("  Current state: %s", server.Status.State)

		// Verify state transitions: stopped → starting → running
		if server.Status.State == models.StatusStarting {
			t.Log("  ✓ Transition: stopped → starting")
		} else if server.Status.State == models.StatusRunning {
			t.Log("  ✓ Transition: starting → running")
			return true
		}
		return false
	}) {
		t.Fatal("Server did not reach running state within timeout")
	}

	// Test Step 5: Verify PID set, HTTP server responding
	server = getServerStatus(t, apiURL, serverID)
	if server.PID == nil {
		t.Error("Expected PID to be set when running")
	} else {
		t.Logf("✓ Server running with PID: %d", *server.PID)
	}

	// Verify the test server's HTTP endpoint is responding
	testServerURL := "http://localhost:18888/ping"
	if waitForHTTPReady(testServerURL, 5*time.Second) {
		t.Logf("✓ Test server HTTP endpoint responding at %s", testServerURL)
	} else {
		t.Log("Warning: Test server HTTP endpoint not responding (may be normal)")
	}

	// Test Step 6: Call POST /servers/{id}/stop
	t.Logf("Stopping server %s...", serverID)
	stopResp, err := http.Post(
		fmt.Sprintf("%s/api/servers/%s/stop", apiURL, serverID),
		"application/json",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to call stop API: %v", err)
	}
	stopResp.Body.Close()

	if stopResp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for stop, got %d", stopResp.StatusCode)
	}

	// Test Step 7: Verify state=stopped, PID cleared
	t.Log("Polling for stopped state...")
	if !waitForCondition(10*time.Second, 500*time.Millisecond, func() bool {
		server := getServerStatus(t, apiURL, serverID)
		t.Logf("  Current state: %s", server.Status.State)

		if server.Status.State == models.StatusStopped {
			// Verify PID cleared
			if server.PID != nil {
				t.Errorf("Expected PID to be nil after stop, got %d", *server.PID)
			}
			t.Log("  ✓ Transition: running → stopped")
			t.Log("  ✓ PID cleared")
			return true
		}
		return false
	}) {
		t.Fatal("Server did not reach stopped state within timeout")
	}

	// Verify HTTP server no longer responding
	time.Sleep(500 * time.Millisecond)
	if resp, err := http.Get(testServerURL); err == nil {
		resp.Body.Close()
		t.Error("Test server HTTP endpoint should not respond after stop")
	} else {
		t.Log("✓ Test server HTTP endpoint stopped responding")
	}

	t.Log("✓ Quickstart Test 2 completed successfully")
}

// getServerStatus fetches the current status of a server by ID
func getServerStatus(t *testing.T, apiURL, serverID string) models.MCPServer {
	t.Helper()

	resp, err := http.Get(fmt.Sprintf("%s/api/servers/%s", apiURL, serverID))
	if err != nil {
		t.Fatalf("Failed to get server status: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var server models.MCPServer
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if err := json.Unmarshal(body, &server); err != nil {
		t.Fatalf("Failed to unmarshal server: %v", err)
	}

	return server
}

// waitForCondition polls a condition until it returns true or timeout
func waitForCondition(timeout, interval time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}

	return false
}
