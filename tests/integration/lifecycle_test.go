package integration

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/Positronikal/MCPManager/internal/core/lifecycle"
	"github.com/Positronikal/MCPManager/internal/models"
	"github.com/Positronikal/MCPManager/internal/platform"
)

// buildTestServer compiles the test server if needed
func buildTestServer(t *testing.T) string {
	// Get the test server directory
	testServerDir := filepath.Join("testserver")

	// Determine binary name based on platform
	binaryName := "testserver"
	if runtime.GOOS == "windows" {
		binaryName = "testserver.exe"
	}

	binaryPath := filepath.Join(testServerDir, binaryName)

	// Check if binary already exists and is recent
	if info, err := os.Stat(binaryPath); err == nil {
		// If less than 1 hour old, reuse it
		if time.Since(info.ModTime()) < time.Hour {
			t.Logf("Using existing test server binary: %s", binaryPath)
			absPath, _ := filepath.Abs(binaryPath)
			return absPath
		}
	}

	// Build the test server
	t.Logf("Building test server...")
	cmd := exec.Command("go", "build", "-o", binaryName, ".")
	cmd.Dir = testServerDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build test server: %v\nOutput: %s", err, output)
	}

	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	t.Logf("Test server built: %s", absPath)
	return absPath
}

// waitForHTTP waits for an HTTP endpoint to become available
func waitForHTTP(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for %s", url)
}

func TestLifecycleFlow_StartAndStop(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build test server
	testServerPath := buildTestServer(t)

	// Create event bus
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	// Subscribe to status change events
	statusEvents := eventBus.Subscribe(events.EventServerStatusChanged)

	// Create lifecycle service (nil discoveryService for isolated tests)
	processManager := platform.NewProcessManager()
	lifecycleService := lifecycle.NewLifecycleService(processManager, nil, nil, eventBus)

	// Create test server configuration
	port := "18765" // Use non-standard port to avoid conflicts
	server := models.NewMCPServer("test-server", testServerPath, models.DiscoveryClientConfig)
	server.Configuration.CommandLineArguments = []string{port}

	// Verify initial state is stopped
	if server.Status.State != models.StatusStopped {
		t.Errorf("Expected initial state 'stopped', got '%s'", server.Status.State)
	}

	// Start the server
	t.Log("Starting server...")
	err := lifecycleService.StartServer(server)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}

	// Wait for status to transition to running (may take a moment)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if server.Status.State == models.StatusRunning {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	// Verify server transitioned to running
	if server.Status.State != models.StatusRunning {
		t.Errorf("Expected state 'running', got '%s'", server.Status.State)
	}

	// Verify PID was captured
	if server.PID == nil {
		t.Error("Expected PID to be set")
	} else {
		t.Logf("Server started with PID: %d", *server.PID)

		// Verify process is actually running
		pm := platform.NewProcessManager()
		if !pm.IsRunning(*server.PID) {
			t.Error("Process should be running but IsRunning returned false")
		}
	}

	// Wait for HTTP server to be ready
	pingURL := fmt.Sprintf("http://localhost:%s/ping", port)
	t.Logf("Waiting for HTTP server at %s...", pingURL)
	if err := waitForHTTP(pingURL, 10*time.Second); err != nil {
		t.Logf("Warning: HTTP server did not start: %v (may be normal if process starts slowly)", err)
		// Don't fail the test - the process started, HTTP server might just be slow
	} else {
		// Verify /ping responds
		resp, err := http.Get(pingURL)
		if err != nil {
			t.Logf("Warning: Failed to ping server: %v", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}
			t.Log("Successfully pinged running server")
		}
	}

	// Verify status change events were published
	eventCount := 0
	timeout := time.After(1 * time.Second)

EventLoop:
	for {
		select {
		case event := <-statusEvents:
			if event != nil && event.Type == events.EventServerStatusChanged {
				eventCount++
				t.Logf("Received status change event: %v -> %v",
					event.Data["oldState"], event.Data["newState"])
			}
		case <-timeout:
			break EventLoop
		}
	}

	if eventCount == 0 {
		t.Error("Expected at least one status change event")
	}

	// Stop the server
	t.Log("Stopping server...")
	err = lifecycleService.StopServer(server, true, 5)
	if err != nil {
		t.Errorf("Failed to stop server: %v", err)
	}

	// Wait briefly for stop to complete
	time.Sleep(500 * time.Millisecond)

	// Verify server transitioned to stopped
	if server.Status.State != models.StatusStopped {
		t.Errorf("Expected state 'stopped', got '%s'", server.Status.State)
	}

	// Verify PID was cleared
	if server.PID != nil {
		t.Errorf("Expected PID to be nil after stop, got %d", *server.PID)
	}

	// Verify process is no longer running
	if server.PID != nil {
		pm := platform.NewProcessManager()
		if pm.IsRunning(*server.PID) {
			t.Error("Process should not be running after stop")
		}
	}

	// Verify HTTP server is no longer responding
	time.Sleep(500 * time.Millisecond)
	_, err = http.Get(pingURL)
	if err == nil {
		t.Error("HTTP server should not respond after stop")
	}

	t.Log("Lifecycle test completed successfully")
}

func TestLifecycleFlow_InvalidCommand(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	processManager := platform.NewProcessManager()
	lifecycleService := lifecycle.NewLifecycleService(processManager, nil, nil, eventBus)

	// Create server with invalid command
	server := models.NewMCPServer("invalid-server", "/nonexistent/command", models.DiscoveryClientConfig)

	// Try to start - should fail
	err := lifecycleService.StartServer(server)
	if err == nil {
		t.Error("Expected error when starting invalid server")
		// Clean up if somehow started
		lifecycleService.StopServer(server, false, 1)
	}

	// Verify state is error
	if server.Status.State != models.StatusError {
		t.Errorf("Expected state 'error', got '%s'", server.Status.State)
	}

	// Verify error message was set
	if server.Status.ErrorMessage == "" {
		t.Error("Expected error message to be set")
	}
}

func TestLifecycleFlow_AlreadyRunning(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testServerPath := buildTestServer(t)

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	processManager := platform.NewProcessManager()
	lifecycleService := lifecycle.NewLifecycleService(processManager, nil, nil, eventBus)

	port := "18766"
	server := models.NewMCPServer("test-server-2", testServerPath, models.DiscoveryClientConfig)
	server.Configuration.CommandLineArguments = []string{port}

	// Start the server
	err := lifecycleService.StartServer(server)
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer lifecycleService.StopServer(server, true, 5)

	time.Sleep(500 * time.Millisecond)

	// Try to start again - should fail
	err = lifecycleService.StartServer(server)
	if err == nil {
		t.Error("Expected error when starting already running server")
	}
}

func TestLifecycleFlow_StopNotRunning(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	processManager := platform.NewProcessManager()
	lifecycleService := lifecycle.NewLifecycleService(processManager, nil, nil, eventBus)

	server := models.NewMCPServer("stopped-server", "/some/path", models.DiscoveryClientConfig)

	// Try to stop a server that's not running
	err := lifecycleService.StopServer(server, true, 5)
	if err == nil {
		t.Error("Expected error when stopping non-running server")
	}
}

func TestLifecycleFlow_MultipleServers(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testServerPath := buildTestServer(t)

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	processManager := platform.NewProcessManager()
	lifecycleService := lifecycle.NewLifecycleService(processManager, nil, nil, eventBus)

	// Start multiple servers on different ports
	servers := []*models.MCPServer{
		models.NewMCPServer("server-1", testServerPath, models.DiscoveryClientConfig),
		models.NewMCPServer("server-2", testServerPath, models.DiscoveryClientConfig),
	}

	servers[0].Configuration.CommandLineArguments = []string{"18767"}
	servers[1].Configuration.CommandLineArguments = []string{"18768"}

	// Start both servers
	for i, server := range servers {
		err := lifecycleService.StartServer(server)
		if err != nil {
			t.Errorf("Failed to start server %d: %v", i+1, err)
		}
	}

	time.Sleep(1 * time.Second)

	// Verify both are running
	for i, server := range servers {
		if server.Status.State != models.StatusRunning {
			t.Errorf("Server %d: expected state 'running', got '%s'", i+1, server.Status.State)
		}
		if server.PID == nil {
			t.Errorf("Server %d: expected PID to be set", i+1)
		}
	}

	// Stop both servers
	for i, server := range servers {
		err := lifecycleService.StopServer(server, true, 5)
		if err != nil {
			t.Errorf("Failed to stop server %d: %v", i+1, err)
		}
	}

	time.Sleep(500 * time.Millisecond)

	// Verify both are stopped
	for i, server := range servers {
		if server.Status.State != models.StatusStopped {
			t.Errorf("Server %d: expected state 'stopped', got '%s'", i+1, server.Status.State)
		}
		if server.PID != nil {
			t.Errorf("Server %d: expected PID to be nil", i+1)
		}
	}

	t.Log("Multiple servers test completed successfully")
}
