package lifecycle

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
	"github.com/hoytech/mcpmanager/internal/platform"
)

// LifecycleService manages server lifecycle operations (start, stop, restart)
type LifecycleService struct {
	processManager   platform.ProcessManager
	discoveryService DiscoveryService // Interface for cache synchronization
	eventBus         *events.EventBus
	mu               sync.RWMutex
	monitors         map[string]chan struct{} // serverID -> stop channel for monitor
}

// DiscoveryService interface for cache updates (avoid circular dependency)
type DiscoveryService interface {
	UpdateServer(server *models.MCPServer)
}

// NewLifecycleService creates a new lifecycle service
func NewLifecycleService(
	processManager platform.ProcessManager,
	discoveryService DiscoveryService,
	eventBus *events.EventBus,
) *LifecycleService {
	return &LifecycleService{
		processManager:   processManager,
		discoveryService: discoveryService,
		eventBus:         eventBus,
		monitors:         make(map[string]chan struct{}),
	}
}

// StartServer starts an MCP server
// Validates state, transitions to starting, launches process, and begins monitoring
func (ls *LifecycleService) StartServer(server *models.MCPServer) error {
	if server == nil {
		return fmt.Errorf("server cannot be nil")
	}

	// Validate current state
	if server.Status.State != models.StatusStopped && server.Status.State != models.StatusError {
		return fmt.Errorf("server must be in stopped or error state to start, current state: %s", server.Status.State)
	}

	// Validate configuration
	if server.InstallationPath == "" {
		return fmt.Errorf("server installation path is missing")
	}

	// Transition to starting state
	oldState := server.Status.State
	if err := server.Status.TransitionTo(models.StatusStarting, "Starting server"); err != nil {
		return fmt.Errorf("failed to transition to starting state: %w", err)
	}

	// Publish status changed event
	if ls.eventBus != nil {
		slog.Info("[EVENT] Publishing server.status.changed", "serverId", server.ID, "oldState", oldState, "newState", models.StatusStarting)
		ls.eventBus.Publish(events.ServerStatusChangedEvent(server.ID, oldState, models.StatusStarting))
	}

	// Extract command and arguments
	cmd := server.InstallationPath
	args := server.Configuration.CommandLineArguments

	// Log command and args for debugging
	slog.Info("[PROCESS] Starting process", "serverId", server.ID, "command", cmd, "args", args, "argsCount", len(args))
	for i, arg := range args {
		slog.Info("[PROCESS] Argument", "index", i, "value", arg)
	}

	// Use environment variables from configuration
	env := server.Configuration.EnvironmentVariables

	// Start the process
	pid, err := ls.processManager.Start(cmd, args, env)
	if err != nil {
		// Transition to error state
		server.Status.TransitionTo(models.StatusError, fmt.Sprintf("Failed to start: %v", err))
		if ls.eventBus != nil {
			ls.eventBus.Publish(events.ServerStatusChangedEvent(server.ID, models.StatusStarting, models.StatusError))
		}
		return fmt.Errorf("failed to start process: %w", err)
	}

	// Update server with PID
	server.SetPID(pid)

	// Synchronously update discovery cache (BUG-001 fix)
	if ls.discoveryService != nil {
		ls.discoveryService.UpdateServer(server)
	}

	// Start monitoring the process
	ls.startMonitoring(server)

	return nil
}

// StopServer stops an MCP server
// If graceful is true, attempts graceful shutdown before forcing termination
func (ls *LifecycleService) StopServer(server *models.MCPServer, force bool, timeout int) error {
	if server == nil {
		return fmt.Errorf("server cannot be nil")
	}

	slog := slog.With("serverId", server.ID, "serverName", server.Name)
	slog.Info("StopServer: Starting stop operation")

	// Validate current state
	if server.Status.State != models.StatusRunning && server.Status.State != models.StatusStarting {
		slog.Warn("StopServer: Invalid state for stop operation", "currentState", server.Status.State)
		return fmt.Errorf("server must be running or starting to stop, current state: %s", server.Status.State)
	}

	// Check if we have a PID
	if server.PID == nil {
		slog.Error("StopServer: Server has no PID")
		return fmt.Errorf("server has no PID")
	}

	pid := *server.PID
	slog.Info("StopServer: Stopping process", "pid", pid, "force", force, "timeout", timeout)

	// Stop monitoring first (prevents race conditions)
	ls.stopMonitoring(server.ID)

	// Verify process is still running before attempting to stop
	if !ls.processManager.IsRunning(pid) {
		slog.Warn("StopServer: Process is not running", "pid", pid)
		// Process already dead, just update state
		oldState := server.Status.State
		server.Status.TransitionTo(models.StatusStopped, "Process not running")
		server.PID = nil

		// Synchronously update discovery cache (BUG-001 fix)
		if ls.discoveryService != nil {
			ls.discoveryService.UpdateServer(server)
			slog.Info("StopServer: Cache synchronized (process was not running)")
		}

		if ls.eventBus != nil {
			ls.eventBus.Publish(events.ServerStatusChangedEvent(server.ID, oldState, models.StatusStopped))
		}
		return nil
	}

	// Stop the process
	graceful := !force
	slog.Info("StopServer: Calling process manager Stop", "pid", pid, "graceful", graceful)
	if err := ls.processManager.Stop(pid, graceful, timeout); err != nil {
		slog.Error("StopServer: Process manager Stop failed", "pid", pid, "error", err)
		return fmt.Errorf("failed to stop process %d: %w", pid, err)
	}

	slog.Info("StopServer: Process stopped successfully", "pid", pid)

	// Transition to stopped state
	oldState := server.Status.State
	if err := server.Status.TransitionTo(models.StatusStopped, "Server stopped"); err != nil {
		slog.Error("StopServer: Failed to transition to stopped state", "error", err)
		return fmt.Errorf("failed to transition to stopped state: %w", err)
	}

	// Clear PID
	server.PID = nil

	// Synchronously update discovery cache (BUG-001 fix)
	if ls.discoveryService != nil {
		ls.discoveryService.UpdateServer(server)
		slog.Info("StopServer: Cache synchronized with stopped state")
	}

	// Publish status changed event
	if ls.eventBus != nil {
		ls.eventBus.Publish(events.ServerStatusChangedEvent(server.ID, oldState, models.StatusStopped))
	}

	slog.Info("StopServer: Stop operation completed successfully")
	return nil
}

// RestartServer restarts an MCP server
// Implements restart as stop + start
func (ls *LifecycleService) RestartServer(server *models.MCPServer) error {
	if server == nil {
		return fmt.Errorf("server cannot be nil")
	}

	// Stop the server (graceful with 10s timeout)
	if err := ls.StopServer(server, false, 10); err != nil {
		return fmt.Errorf("failed to stop server during restart: %w", err)
	}

	// Wait for stopped state
	// (StopServer should have already transitioned to stopped)
	if server.Status.State != models.StatusStopped {
		return fmt.Errorf("server not in stopped state after stop: %s", server.Status.State)
	}

	// Start the server
	if err := ls.StartServer(server); err != nil {
		return fmt.Errorf("failed to start server during restart: %w", err)
	}

	return nil
}

// startMonitoring begins monitoring a server process
// Monitors for process exit and transitions to error if process dies within 5 seconds
func (ls *LifecycleService) startMonitoring(server *models.MCPServer) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	// Stop any existing monitor
	if stopChan, exists := ls.monitors[server.ID]; exists {
		close(stopChan)
	}

	// Create new stop channel
	stopChan := make(chan struct{})
	ls.monitors[server.ID] = stopChan

	// Start monitoring goroutine
	go ls.monitorProcess(server, stopChan)
}

// stopMonitoring stops monitoring a server process
func (ls *LifecycleService) stopMonitoring(serverID string) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if stopChan, exists := ls.monitors[serverID]; exists {
		close(stopChan)
		delete(ls.monitors, serverID)
	}
}

// monitorProcess monitors a server process for unexpected exits
func (ls *LifecycleService) monitorProcess(server *models.MCPServer, stopChan chan struct{}) {
	if server.PID == nil {
		return
	}

	pid := *server.PID
	startTime := time.Now()

	// Check process every 100ms
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Track if we've transitioned to running
	transitionedToRunning := false

	for {
		select {
		case <-stopChan:
			// Monitoring stopped
			return

		case <-ticker.C:
			// Check if process is still running
			if !ls.processManager.IsRunning(pid) {
				// Process has exited
				elapsed := time.Since(startTime)

				if elapsed < 5*time.Second {
					// Process exited too quickly - transition to error
					oldState := server.Status.State
					slog.Error("[MONITOR] Process crashed (exited too quickly)", "serverId", server.ID, "serverName", server.Name, "elapsed", elapsed, "pid", pid)
					server.Status.TransitionTo(models.StatusError, "Process exited unexpectedly")
					server.PID = nil

					// Synchronously update discovery cache (BUG-001 fix)
					if ls.discoveryService != nil {
						ls.discoveryService.UpdateServer(server)
					}

					if ls.eventBus != nil {
						slog.Info("[EVENT] Publishing server.status.changed (crashed)", "serverId", server.ID, "oldState", oldState, "newState", models.StatusError)
						ls.eventBus.Publish(events.ServerStatusChangedEvent(server.ID, oldState, models.StatusError))
					}
				} else {
					// Process exited after running for a while - transition to stopped
					oldState := server.Status.State
					server.Status.TransitionTo(models.StatusStopped, "Process exited")
					server.PID = nil

					// Synchronously update discovery cache (BUG-001 fix)
					if ls.discoveryService != nil {
						ls.discoveryService.UpdateServer(server)
					}

					if ls.eventBus != nil {
						ls.eventBus.Publish(events.ServerStatusChangedEvent(server.ID, oldState, models.StatusStopped))
					}
				}

				// Stop monitoring
				ls.stopMonitoring(server.ID)
				return
			}

			// If still in starting state and process is alive, transition to running
			if !transitionedToRunning && server.Status.State == models.StatusStarting {
				// Wait at least 500ms before transitioning to running
				if time.Since(startTime) >= 500*time.Millisecond {
					oldState := server.Status.State
					server.Status.TransitionTo(models.StatusRunning, "Server started successfully")

					// Synchronously update discovery cache (BUG-001 fix)
					if ls.discoveryService != nil {
						ls.discoveryService.UpdateServer(server)
					}

					if ls.eventBus != nil {
						slog.Info("[EVENT] Publishing server.status.changed (monitor)", "serverId", server.ID, "oldState", oldState, "newState", models.StatusRunning)
						ls.eventBus.Publish(events.ServerStatusChangedEvent(server.ID, oldState, models.StatusRunning))
					}

					transitionedToRunning = true
				}
			}
		}
	}
}

// StopAll stops all monitored servers
func (ls *LifecycleService) StopAll() {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	for _, stopChan := range ls.monitors {
		close(stopChan)
	}

	ls.monitors = make(map[string]chan struct{})
}
