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
	"syscall"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/models"
)

// TestEdgeCase_ServerCrash tests edge case 2: Server crash during operation
// Requirement: FR-023 (crash detection), FR-024 (status change events)
func TestEdgeCase_ServerCrash(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the test server
	testServerPath := buildTestServer(t)

	// Create temporary test environment
	tempDir, err := os.MkdirTemp("", "mcp-edge-crash-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create mock client config with test server
	configDir := createMockConfigDir(t, tempDir)
	createMockClientConfig(t, configDir, []map[string]interface{}{
		{
			"name":    "crash-test-server",
			"command": testServerPath,
			"args":    []interface{}{"18950"},
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

	// Test Step 1: Wait for server discovery
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
			if server.Name == "crash-test-server" {
				serverID = server.ID
				t.Logf("Found test server: ID=%s", serverID)
				return true
			}
		}
		return false
	}) {
		t.Fatal("Test server not discovered within timeout")
	}

	// Test Step 2: Start the server
	t.Logf("Starting server %s...", serverID)
	startResp, err := http.Post(
		fmt.Sprintf("%s/api/servers/%s/start", apiURL, serverID),
		"application/json",
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	startResp.Body.Close()

	// Wait for server to be running
	if !waitForCondition(10*time.Second, 500*time.Millisecond, func() bool {
		server := getServerStatus(t, apiURL, serverID)
		return server.Status.State == models.StatusRunning
	}) {
		t.Fatal("Server did not start within timeout")
	}

	// Test Step 3: Verify status=running
	server := getServerStatus(t, apiURL, serverID)
	if server.Status.State != models.StatusRunning {
		t.Errorf("Expected state 'running', got '%s'", server.Status.State)
	}
	t.Log("✓ Server running")

	// Capture the PID before killing
	if server.PID == nil {
		t.Fatal("Server PID is nil, cannot proceed with crash test")
	}
	serverPID := *server.PID
	t.Logf("Server PID: %d", serverPID)

	// Test Step 4: Kill process externally (simulate crash)
	t.Logf("Simulating crash by killing PID %d...", serverPID)
	if err := killProcess(serverPID); err != nil {
		t.Fatalf("Failed to kill process: %v", err)
	}
	t.Log("✓ Process killed externally")

	// Test Step 5: Poll GET /servers/{id}/status - verify status transitions to error within 5 seconds
	t.Log("Polling for crash detection (expecting transition to error state)...")
	crashDetected := false
	startTime := time.Now()

	if waitForCondition(10*time.Second, 500*time.Millisecond, func() bool {
		server := getServerStatus(t, apiURL, serverID)
		t.Logf("  Current state: %s", server.Status.State)

		// Check if transitioned to error state
		if server.Status.State == models.StatusError {
			crashDetected = true
			detectionTime := time.Since(startTime)
			t.Logf("✓ Crash detected in %v", detectionTime)

			// Test Step 6: Verify error message contains "crashed" or "exited unexpectedly"
			if server.Status.ErrorMessage == "" {
				t.Error("Expected error message to be set")
			} else {
				t.Logf("Error message: %s", server.Status.ErrorMessage)

				// Check message indicates crash/unexpected exit
				msg := server.Status.ErrorMessage
				if containsAny(msg, []string{"crash", "exited", "unexpected", "terminated", "killed"}) {
					t.Log("✓ Error message indicates crash/unexpected termination")
				} else {
					t.Logf("Warning: Error message may not clearly indicate crash: %s", msg)
				}
			}

			// Verify detection time (should be within 5 seconds per requirement)
			if detectionTime > 5*time.Second {
				t.Errorf("FAIL: Crash detection took %v, exceeds 5 second requirement", detectionTime)
			} else {
				t.Logf("✓ PASS: Crash detected within 5 second requirement")
			}

			return true
		}

		return false
	}) {
		// Crash detected successfully
		t.Log("✓ Crash detection test passed")
	} else {
		t.Error("FAIL: Crash not detected within timeout")
	}

	if !crashDetected {
		t.Error("FAIL: Server crash was not detected")
	}

	t.Log("✓ Edge Case Test: Server Crash completed successfully")
}

// killProcess kills a process by PID in a platform-specific way
func killProcess(pid int) error {
	// Find the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %w", pid, err)
	}

	// Send kill signal (SIGKILL on Unix, TerminateProcess on Windows)
	if err := process.Signal(syscall.SIGKILL); err != nil {
		// On Windows, Signal might not work, try Kill directly
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process %d: %w", pid, err)
		}
	}

	return nil
}

// containsAny checks if a string contains any of the given substrings (case-insensitive)
func containsAny(s string, substrs []string) bool {
	sLower := toLower(s)
	for _, substr := range substrs {
		if contains(sLower, toLower(substr)) {
			return true
		}
	}
	return false
}

// toLower converts string to lowercase
func toLower(s string) string {
	// Simple lowercase conversion for ASCII
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

// findSubstring searches for substr in s
func findSubstring(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(substr) > len(s) {
		return false
	}

	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}
