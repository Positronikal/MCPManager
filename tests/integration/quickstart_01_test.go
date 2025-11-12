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
	"runtime"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/models"
)

// TestQuickstart01_InitialLaunchAndDiscovery tests the first quickstart scenario:
// Initial launch & discovery
// Requirement: FR-037 (startup time < 2s), FR-002 (discovery on launch)
func TestQuickstart01_InitialLaunchAndDiscovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create temporary test environment
	tempDir, err := os.MkdirTemp("", "mcp-quickstart-01-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create mock client config directory structure
	configDir := createMockConfigDir(t, tempDir)

	// Create a mock client config with test servers
	createMockClientConfig(t, configDir, []map[string]interface{}{
		{
			"name":    "quickstart-server-1",
			"command": "node",
			"args":    []interface{}{"server1.js"},
			"env":     map[string]interface{}{},
		},
		{
			"name":    "quickstart-server-2",
			"command": "python3",
			"args":    []interface{}{"-m", "http.server"},
			"env":     map[string]interface{}{"PORT": "8080"},
		},
	})

	// Build MCP Manager if not already built
	appPath := buildMCPManager(t)

	// Set up application state directory
	appStateDir := filepath.Join(tempDir, ".mcpmanager")
	logsDir := filepath.Join(appStateDir, "logs")

	// Start MCP Manager in background
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, appPath)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("MCP_CONFIG_DIR=%s", configDir),
		fmt.Sprintf("MCP_STATE_DIR=%s", appStateDir),
		"HEADLESS=true",
	)

	t.Log("Starting MCP Manager...")
	startTime := time.Now()

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start MCP Manager: %v", err)
	}
	defer func() {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}()

	// Wait for application to start (HTTP server to be ready)
	apiURL := "http://localhost:8080"
	if !waitForHTTPReady(apiURL+"/health", 10*time.Second) {
		t.Fatal("MCP Manager did not start within timeout")
	}

	startupDuration := time.Since(startTime)
	t.Logf("Startup time: %v", startupDuration)

	// Verify FR-037: Startup time < 2 seconds
	if startupDuration > 2*time.Second {
		t.Errorf("FAIL FR-037: Startup time %v exceeds 2 seconds", startupDuration)
	} else {
		t.Logf("PASS FR-037: Startup time within requirement")
	}

	// Poll GET /servers until discovery completes
	t.Log("Waiting for server discovery...")
	var servers []models.MCPServer
	discoveryTimeout := time.After(10 * time.Second)
	discoveryTicker := time.NewTicker(500 * time.Millisecond)
	defer discoveryTicker.Stop()

DiscoveryLoop:
	for {
		select {
		case <-discoveryTimeout:
			t.Fatal("Discovery did not complete within timeout")
		case <-discoveryTicker.C:
			resp, err := http.Get(apiURL + "/api/servers")
			if err != nil {
				continue
			}

			body, err := io.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				continue
			}

			if err := json.Unmarshal(body, &servers); err != nil {
				continue
			}

			if len(servers) >= 2 {
				t.Logf("Discovery complete: Found %d servers", len(servers))
				break DiscoveryLoop
			}
		}
	}

	// Verify: At least 2 servers discovered
	if len(servers) < 2 {
		t.Errorf("Expected at least 2 servers discovered, got %d", len(servers))
	}

	// Verify server details
	serverNames := make(map[string]bool)
	for _, server := range servers {
		serverNames[server.Name] = true
		t.Logf("  - Server: %s (source: %s, transport: %s)",
			server.Name, server.Source, server.Transport)
	}

	if !serverNames["quickstart-server-1"] {
		t.Error("Expected to find 'quickstart-server-1'")
	}
	if !serverNames["quickstart-server-2"] {
		t.Error("Expected to find 'quickstart-server-2'")
	}

	// Give application a moment to create state files
	time.Sleep(1 * time.Second)

	// Verify: Application state file created
	stateFile := filepath.Join(appStateDir, "state.json")
	if _, err := os.Stat(stateFile); err != nil {
		if os.IsNotExist(err) {
			t.Logf("Note: state.json not created (may be created on shutdown)")
		} else {
			t.Errorf("Error checking state file: %v", err)
		}
	} else {
		t.Log("PASS: Application state file created")
	}

	// Verify: Logs directory created
	if _, err := os.Stat(logsDir); err != nil {
		if os.IsNotExist(err) {
			t.Error("FAIL: Logs directory not created")
		} else {
			t.Errorf("Error checking logs directory: %v", err)
		}
	} else {
		t.Log("PASS: Logs directory created")
	}

	t.Log("Quickstart Test 1 completed successfully")
}

// buildMCPManager builds the MCP Manager application for testing
func buildMCPManager(t *testing.T) string {
	t.Helper()

	// Determine binary name
	binaryName := "mcpmanager"
	if runtime.GOOS == "windows" {
		binaryName = "mcpmanager.exe"
	}

	// Check if already built
	buildPath := filepath.Join("..", "..", "build", "bin", binaryName)
	if info, err := os.Stat(buildPath); err == nil {
		// If less than 1 hour old, reuse it
		if time.Since(info.ModTime()) < time.Hour {
			absPath, _ := filepath.Abs(buildPath)
			t.Logf("Using existing build: %s", absPath)
			return absPath
		}
	}

	// Build using wails
	t.Log("Building MCP Manager (this may take a minute)...")
	cmd := exec.Command("wails", "build", "-clean")
	cmd.Dir = filepath.Join("..", "..")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build MCP Manager: %v\nOutput: %s", err, output)
	}

	absPath, err := filepath.Abs(buildPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	t.Logf("Build complete: %s", absPath)
	return absPath
}

// waitForHTTPReady waits for an HTTP endpoint to respond successfully
func waitForHTTPReady(url string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
		time.Sleep(200 * time.Millisecond)
	}

	return false
}

// createMockConfigDir creates a platform-specific config directory structure
func createMockConfigDir(t *testing.T, tempDir string) string {
	t.Helper()

	var configDir string
	switch runtime.GOOS {
	case "windows":
		configDir = filepath.Join(tempDir, "AppData", "Roaming")
	case "darwin":
		configDir = filepath.Join(tempDir, "Library", "Application Support")
	default:
		configDir = filepath.Join(tempDir, ".config")
	}

	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	return configDir
}
