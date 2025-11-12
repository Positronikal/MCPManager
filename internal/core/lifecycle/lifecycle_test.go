package lifecycle

import (
	"context"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/Positronikal/MCPManager/internal/models"
)

// MockProcessManager is a mock implementation of ProcessManager for testing
type MockProcessManager struct {
	StartFunc           func(cmd string, args []string, env map[string]string) (int, error)
	StartWithOutputFunc func(cmd string, args []string, env map[string]string) (int, io.ReadCloser, io.ReadCloser, error)
	StopFunc            func(pid int, graceful bool, timeout int) error
	IsRunningFunc       func(pid int) bool
}

func (m *MockProcessManager) Start(cmd string, args []string, env map[string]string) (int, error) {
	if m.StartFunc != nil {
		return m.StartFunc(cmd, args, env)
	}
	return 1234, nil
}

func (m *MockProcessManager) StartWithOutput(cmd string, args []string, env map[string]string) (int, io.ReadCloser, io.ReadCloser, error) {
	if m.StartWithOutputFunc != nil {
		return m.StartWithOutputFunc(cmd, args, env)
	}
	// Return mock readers (empty)
	return 1234, io.NopCloser(strings.NewReader("")), io.NopCloser(strings.NewReader("")), nil
}

func (m *MockProcessManager) Stop(pid int, graceful bool, timeout int) error {
	if m.StopFunc != nil {
		return m.StopFunc(pid, graceful, timeout)
	}
	return nil
}

func (m *MockProcessManager) IsRunning(pid int) bool {
	if m.IsRunningFunc != nil {
		return m.IsRunningFunc(pid)
	}
	return true
}

// MockDiscoveryService is a mock implementation of DiscoveryService for testing (BUG-001 fix)
type MockDiscoveryService struct {
	UpdateServerFunc     func(server *models.MCPServer)
	GetCachedServersFunc func() []models.MCPServer
}

func (m *MockDiscoveryService) UpdateServer(server *models.MCPServer) {
	if m.UpdateServerFunc != nil {
		m.UpdateServerFunc(server)
	}
}

func (m *MockDiscoveryService) GetCachedServers() []models.MCPServer {
	if m.GetCachedServersFunc != nil {
		return m.GetCachedServersFunc()
	}
	return []models.MCPServer{}
}

// MockMonitoringService is a mock implementation of MonitoringService for testing
type MockMonitoringService struct {
	CaptureOutputFunc func(ctx context.Context, serverID string, reader io.Reader)
}

func (m *MockMonitoringService) CaptureOutput(ctx context.Context, serverID string, reader io.Reader) {
	if m.CaptureOutputFunc != nil {
		m.CaptureOutputFunc(ctx, serverID, reader)
	}
}

func TestNewLifecycleService(t *testing.T) {
	pm := &MockProcessManager{}
	ds := &MockDiscoveryService{}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewLifecycleService(pm, ds, &MockMonitoringService{}, eventBus)
	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.processManager != pm {
		t.Error("ProcessManager should be set")
	}

	if service.discoveryService != ds {
		t.Error("DiscoveryService should be set")
	}

	if service.eventBus != eventBus {
		t.Error("EventBus should be set")
	}

	if service.monitors == nil {
		t.Error("Monitors map should be initialized")
	}
}

func TestLifecycleService_StartServer(t *testing.T) {
	pm := &MockProcessManager{
		StartFunc: func(cmd string, args []string, env map[string]string) (int, error) {
			return 1234, nil
		},
	}

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewLifecycleService(pm, &MockDiscoveryService{}, &MockMonitoringService{}, eventBus)

	// Create a server
	server := models.NewMCPServer("test-server", "test-cmd", models.DiscoveryClientConfig)
	server.Configuration.CommandLineArguments = []string{"arg1", "arg2"}
	server.Configuration.EnvironmentVariables = map[string]string{"KEY": "value"}

	// Start the server
	err := service.StartServer(server)
	if err != nil {
		t.Fatalf("StartServer should not error: %v", err)
	}

	// Verify PID was set
	if server.PID == nil {
		t.Fatal("PID should be set")
	}

	if *server.PID != 1234 {
		t.Errorf("Expected PID 1234, got %d", *server.PID)
	}

	// Verify state is starting (monitoring will transition to running)
	if server.Status.State != models.StatusStarting {
		t.Errorf("Expected starting state, got %s", server.Status.State)
	}

	// Clean up
	service.StopAll()
}

func TestLifecycleService_StartServer_InvalidState(t *testing.T) {
	pm := &MockProcessManager{}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewLifecycleService(pm, &MockDiscoveryService{}, &MockMonitoringService{}, eventBus)

	// Create a server already running
	server := models.NewMCPServer("test-server", "test-cmd", models.DiscoveryClientConfig)
	server.Status.State = models.StatusRunning

	// Try to start - should fail
	err := service.StartServer(server)
	if err == nil {
		t.Error("StartServer should error when server is already running")
	}
}

func TestLifecycleService_StartServer_MissingConfiguration(t *testing.T) {
	pm := &MockProcessManager{}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewLifecycleService(pm, &MockDiscoveryService{}, &MockMonitoringService{}, eventBus)

	// Create a server without installation path
	server := models.NewMCPServer("test-server", "", models.DiscoveryClientConfig)

	// Try to start - should fail
	err := service.StartServer(server)
	if err == nil {
		t.Error("StartServer should error when installation path is missing")
	}
}

func TestLifecycleService_StopServer(t *testing.T) {
	stopCalled := false
	pm := &MockProcessManager{
		StopFunc: func(pid int, graceful bool, timeout int) error {
			stopCalled = true
			if pid != 1234 {
				t.Errorf("Expected PID 1234, got %d", pid)
			}
			if !graceful {
				t.Error("Expected graceful stop")
			}
			if timeout != 10 {
				t.Errorf("Expected timeout 10, got %d", timeout)
			}
			return nil
		},
	}

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewLifecycleService(pm, &MockDiscoveryService{}, &MockMonitoringService{}, eventBus)

	// Create a running server
	server := models.NewMCPServer("test-server", "/path/to/server", models.DiscoveryClientConfig)
	pid := 1234
	server.PID = &pid
	server.Status.State = models.StatusRunning

	// Stop the server
	err := service.StopServer(server, false, 10)
	if err != nil {
		t.Fatalf("StopServer should not error: %v", err)
	}

	// Verify stop was called
	if !stopCalled {
		t.Error("ProcessManager.Stop should have been called")
	}

	// Verify PID was cleared
	if server.PID != nil {
		t.Error("PID should be cleared")
	}

	// Verify state is stopped
	if server.Status.State != models.StatusStopped {
		t.Errorf("Expected stopped state, got %s", server.Status.State)
	}
}

func TestLifecycleService_StopServer_Force(t *testing.T) {
	gracefulUsed := false
	pm := &MockProcessManager{
		StopFunc: func(pid int, graceful bool, timeout int) error {
			gracefulUsed = graceful
			return nil
		},
	}

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewLifecycleService(pm, &MockDiscoveryService{}, &MockMonitoringService{}, eventBus)

	// Create a running server
	server := models.NewMCPServer("test-server", "/path/to/server", models.DiscoveryClientConfig)
	pid := 1234
	server.PID = &pid
	server.Status.State = models.StatusRunning

	// Stop the server with force
	err := service.StopServer(server, true, 5)
	if err != nil {
		t.Fatalf("StopServer should not error: %v", err)
	}

	// Verify graceful was false
	if gracefulUsed {
		t.Error("Should not use graceful stop when force=true")
	}
}

func TestLifecycleService_StopServer_InvalidState(t *testing.T) {
	pm := &MockProcessManager{}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewLifecycleService(pm, &MockDiscoveryService{}, &MockMonitoringService{}, eventBus)

	// Create a stopped server
	server := models.NewMCPServer("test-server", "/path/to/server", models.DiscoveryClientConfig)
	server.Status.State = models.StatusStopped

	// Try to stop - should fail
	err := service.StopServer(server, false, 10)
	if err == nil {
		t.Error("StopServer should error when server is already stopped")
	}
}

func TestLifecycleService_RestartServer(t *testing.T) {
	pm := &MockProcessManager{
		StartFunc: func(cmd string, args []string, env map[string]string) (int, error) {
			return 5678, nil
		},
		StopFunc: func(pid int, graceful bool, timeout int) error {
			return nil
		},
	}

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewLifecycleService(pm, &MockDiscoveryService{}, &MockMonitoringService{}, eventBus)

	// Create a running server
	server := models.NewMCPServer("test-server", "/path/to/server", models.DiscoveryClientConfig)
	pid := 1234
	server.PID = &pid
	server.Status.State = models.StatusRunning

	// Restart the server
	err := service.RestartServer(server)
	if err != nil {
		t.Fatalf("RestartServer should not error: %v", err)
	}

	// Verify new PID was set
	if server.PID == nil {
		t.Fatal("PID should be set after restart")
	}

	if *server.PID != 5678 {
		t.Errorf("Expected new PID 5678, got %d", *server.PID)
	}

	// Clean up
	service.StopAll()
}

func TestLifecycleService_MonitorProcess_TransitionToRunning(t *testing.T) {
	pm := &MockProcessManager{
		StartFunc: func(cmd string, args []string, env map[string]string) (int, error) {
			return 1234, nil
		},
		IsRunningFunc: func(pid int) bool {
			return true // Process is running
		},
	}

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewLifecycleService(pm, &MockDiscoveryService{}, &MockMonitoringService{}, eventBus)

	// Subscribe to status change events
	eventChan := eventBus.Subscribe(events.EventServerStatusChanged)

	// Create and start server
	server := models.NewMCPServer("test-server", "/path/to/server", models.DiscoveryClientConfig)

	err := service.StartServer(server)
	if err != nil {
		t.Fatalf("StartServer should not error: %v", err)
	}

	// Wait for transition to running
	timeout := time.After(2 * time.Second)
	foundRunning := false

	for !foundRunning {
		select {
		case event := <-eventChan:
			if event == nil {
				continue
			}
			if newState, ok := event.Data["newState"].(models.StatusState); ok {
				if newState == models.StatusRunning {
					foundRunning = true
				}
			}
		case <-timeout:
			t.Fatal("Timeout waiting for running state transition")
		}
	}

	// Verify server is in running state
	if server.Status.State != models.StatusRunning {
		t.Errorf("Expected running state, got %s", server.Status.State)
	}

	// Clean up
	service.StopAll()
}

func TestLifecycleService_MonitorProcess_EarlyExit(t *testing.T) {
	var mu sync.Mutex
	processRunning := true

	pm := &MockProcessManager{
		StartFunc: func(cmd string, args []string, env map[string]string) (int, error) {
			return 1234, nil
		},
		IsRunningFunc: func(pid int) bool {
			mu.Lock()
			defer mu.Unlock()
			return processRunning
		},
	}

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewLifecycleService(pm, &MockDiscoveryService{}, &MockMonitoringService{}, eventBus)

	// Subscribe to status change events
	eventChan := eventBus.Subscribe(events.EventServerStatusChanged)

	// Create and start server
	server := models.NewMCPServer("test-server", "/path/to/server", models.DiscoveryClientConfig)

	err := service.StartServer(server)
	if err != nil {
		t.Fatalf("StartServer should not error: %v", err)
	}

	// Simulate early process exit (within 5 seconds)
	time.Sleep(200 * time.Millisecond)
	mu.Lock()
	processRunning = false
	mu.Unlock()

	// Wait for transition to error
	timeout := time.After(2 * time.Second)
	foundError := false

	for !foundError {
		select {
		case event := <-eventChan:
			if event == nil {
				continue
			}
			if newState, ok := event.Data["newState"].(models.StatusState); ok {
				if newState == models.StatusError {
					foundError = true
				}
			}
		case <-timeout:
			t.Fatal("Timeout waiting for error state transition")
		}
	}

	// Verify server is in error state
	if server.Status.State != models.StatusError {
		t.Errorf("Expected error state, got %s", server.Status.State)
	}

	// Verify PID was cleared
	if server.PID != nil {
		t.Error("PID should be cleared on error")
	}
}

func TestLifecycleService_StopAll(t *testing.T) {
	pm := &MockProcessManager{
		StartFunc: func(cmd string, args []string, env map[string]string) (int, error) {
			return 1234, nil
		},
	}

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewLifecycleService(pm, &MockDiscoveryService{}, &MockMonitoringService{}, eventBus)

	// Start multiple servers
	server1 := models.NewMCPServer("server1", "/path1", models.DiscoveryClientConfig)
	server2 := models.NewMCPServer("server2", "/path2", models.DiscoveryClientConfig)

	service.StartServer(server1)
	service.StartServer(server2)

	// Verify monitors are running
	service.mu.RLock()
	monitorCount := len(service.monitors)
	service.mu.RUnlock()

	if monitorCount != 2 {
		t.Errorf("Expected 2 monitors, got %d", monitorCount)
	}

	// Stop all
	service.StopAll()

	// Verify monitors are cleared
	service.mu.RLock()
	monitorCount = len(service.monitors)
	service.mu.RUnlock()

	if monitorCount != 0 {
		t.Errorf("Expected 0 monitors after StopAll, got %d", monitorCount)
	}
}

func TestLifecycleService_EventsPublished(t *testing.T) {
	pm := &MockProcessManager{
		StartFunc: func(cmd string, args []string, env map[string]string) (int, error) {
			return 1234, nil
		},
		StopFunc: func(pid int, graceful bool, timeout int) error {
			return nil
		},
		IsRunningFunc: func(pid int) bool {
			return true
		},
	}

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewLifecycleService(pm, &MockDiscoveryService{}, &MockMonitoringService{}, eventBus)

	// Subscribe to status change events
	eventChan := eventBus.Subscribe(events.EventServerStatusChanged)

	// Create and start server
	server := models.NewMCPServer("test-server", "/path/to/server", models.DiscoveryClientConfig)

	// Start server
	service.StartServer(server)

	// Should receive starting event
	timeout := time.After(1 * time.Second)
	select {
	case event := <-eventChan:
		if event.Type != events.EventServerStatusChanged {
			t.Error("Should receive status changed event")
		}
		newState := event.Data["newState"].(models.StatusState)
		if newState != models.StatusStarting {
			t.Errorf("Expected starting state, got %s", newState)
		}
	case <-timeout:
		t.Error("Timeout waiting for starting event")
	}

	// Wait for running transition
	foundRunning := false
	for !foundRunning {
		select {
		case event := <-eventChan:
			if event == nil {
				continue
			}
			if newState, ok := event.Data["newState"].(models.StatusState); ok {
				if newState == models.StatusRunning {
					foundRunning = true
				}
			}
		case <-timeout:
			t.Fatal("Timeout waiting for running state transition")
		}
	}

	// Stop server
	service.StopServer(server, false, 10)

	// Should receive stopped event
	select {
	case event := <-eventChan:
		newState := event.Data["newState"].(models.StatusState)
		if newState != models.StatusStopped {
			t.Errorf("Expected stopped state event, got %s", newState)
		}
	case <-timeout:
		t.Error("Timeout waiting for stopped event")
	}
}
