package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Positronikal/MCPManager/internal/core/config"
	"github.com/Positronikal/MCPManager/internal/core/dependencies"
	"github.com/Positronikal/MCPManager/internal/core/discovery"
	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/Positronikal/MCPManager/internal/core/lifecycle"
	"github.com/Positronikal/MCPManager/internal/core/monitoring"
	"github.com/Positronikal/MCPManager/internal/models"
	"github.com/Positronikal/MCPManager/internal/platform"
	"github.com/Positronikal/MCPManager/internal/storage"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds all application services
type App struct {
	ctx               context.Context
	discoveryService  *discovery.DiscoveryService
	lifecycleService  *lifecycle.LifecycleService
	configService     *config.ConfigService
	monitoringService *monitoring.MonitoringService
	metricsCollector  *monitoring.MetricsCollector
	dependencyService *dependencies.DependencyService
	updateChecker     *dependencies.UpdateChecker
	storageService    storage.StorageService
	eventBus          *events.EventBus
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	slog.Info("Starting MCP Manager Wails application")

	// Initialize platform-specific components
	pathResolver := platform.NewPathResolver()
	processManager := platform.NewProcessManager()
	processInfo := platform.NewProcessInfo()

	// Initialize EventBus (central event dispatcher)
	a.eventBus = events.NewEventBus()
	slog.Info("EventBus initialized")

	// Subscribe to EventBus and emit events to frontend
	a.subscribeToEvents()

	// Initialize storage service
	storageService, err := storage.NewFileStorage()
	if err != nil {
		slog.Error("Failed to initialize storage service", "error", err)
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Initialization Error",
			Message: fmt.Sprintf("Failed to initialize storage service: %v", err),
		})
		return
	}
	a.storageService = storageService
	slog.Info("Storage service initialized")

	// Initialize core services
	a.discoveryService = discovery.NewDiscoveryService(pathResolver, a.eventBus)
	slog.Info("Discovery service initialized")

	// Initialize monitoring service (needed by lifecycle for log capture)
	a.monitoringService = monitoring.NewMonitoringService(a.eventBus)
	slog.Info("Monitoring service initialized")

	// Initialize lifecycle service with discovery and monitoring dependencies
	a.lifecycleService = lifecycle.NewLifecycleService(processManager, a.discoveryService, a.monitoringService, a.eventBus)
	slog.Info("Lifecycle service initialized")

	configService, err := config.NewConfigService(a.eventBus)
	if err != nil {
		slog.Error("Failed to initialize config service", "error", err)
		runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
			Type:    runtime.ErrorDialog,
			Title:   "Initialization Error",
			Message: fmt.Sprintf("Failed to initialize config service: %v", err),
		})
		return
	}
	a.configService = configService
	slog.Info("Config service initialized")

	a.metricsCollector = monitoring.NewMetricsCollector(processInfo, a.eventBus)
	slog.Info("Metrics collector initialized")

	a.dependencyService = dependencies.NewDependencyService()
	slog.Info("Dependency service initialized")

	a.updateChecker = dependencies.NewUpdateChecker()
	slog.Info("Update checker initialized")

	// Perform initial discovery
	slog.Info("Running initial server discovery...")
	servers, err := a.discoveryService.Discover()
	if err != nil {
		slog.Warn("Initial discovery returned error", "error", err)
	} else {
		slog.Info("Initial discovery complete", "servers_found", len(servers))
		// Emit initial servers to frontend
		runtime.EventsEmit(ctx, "servers:initial", servers)
	}
}

// shutdown is called when the app is shutting down
func (a *App) shutdown(ctx context.Context) {
	slog.Info("Shutting down MCP Manager...")

	// Stop lifecycle service (gracefully stop managed servers)
	if a.lifecycleService != nil {
		slog.Info("Stopping managed servers...")
		a.lifecycleService.StopAll()
	}

	// Close discovery service (stops file watcher - FR-050)
	if a.discoveryService != nil {
		slog.Info("Closing discovery service...")
		if err := a.discoveryService.Close(); err != nil {
			slog.Warn("Failed to close discovery service", "error", err)
		}
	}

	// Close EventBus
	if a.eventBus != nil {
		slog.Info("Closing EventBus...")
		a.eventBus.Close()
	}

	slog.Info("Shutdown complete")
}

// subscribeToEvents sets up event listeners and forwards them to the frontend
func (a *App) subscribeToEvents() {
	// Server discovered event
	serverDiscoveredCh := a.eventBus.Subscribe(events.EventServerDiscovered)
	go func() {
		for event := range serverDiscoveredCh {
			runtime.EventsEmit(a.ctx, "server:discovered", event.Data)
		}
	}()

	// Server status changed event
	serverStatusCh := a.eventBus.Subscribe(events.EventServerStatusChanged)
	go func() {
		for event := range serverStatusCh {
			runtime.EventsEmit(a.ctx, "server:status:changed", event.Data)
		}
	}()

	// Server log entry event
	serverLogCh := a.eventBus.Subscribe(events.EventServerLogEntry)
	go func() {
		for event := range serverLogCh {
			runtime.EventsEmit(a.ctx, "server:log:entry", event.Data)
		}
	}()

	// Server metrics updated event
	serverMetricsCh := a.eventBus.Subscribe(events.EventServerMetricsUpdated)
	go func() {
		for event := range serverMetricsCh {
			runtime.EventsEmit(a.ctx, "server:metrics:updated", event.Data)
		}
	}()

	// Config file changed event
	configChangedCh := a.eventBus.Subscribe(events.EventConfigFileChanged)
	go func() {
		for event := range configChangedCh {
			runtime.EventsEmit(a.ctx, "server:config:updated", event.Data)
		}
	}()

	slog.Info("Event subscriptions configured")
}

// ========================================
// Discovery Methods
// ========================================

// ListServersResponse represents the response from ListServers
type ListServersResponse struct {
	Servers       []models.MCPServer `json:"servers"`
	Count         int                `json:"count"`
	LastDiscovery string             `json:"lastDiscovery"`
}

// ListServers returns all discovered servers
func (a *App) ListServers() (*ListServersResponse, error) {
	slog.Info("ListServers called")
	servers, lastDiscovery, err := a.discoveryService.GetServers()
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}
	return &ListServersResponse{
		Servers:       servers,
		Count:         len(servers),
		LastDiscovery: lastDiscovery.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// DiscoverServersResponse represents the response from DiscoverServers
type DiscoverServersResponse struct {
	Message string `json:"message"`
	ScanID  string `json:"scanId"`
}

// DiscoverServers triggers a new server discovery scan
func (a *App) DiscoverServers() (*DiscoverServersResponse, error) {
	slog.Info("DiscoverServers called")
	servers, err := a.discoveryService.Discover()
	if err != nil {
		return nil, fmt.Errorf("discovery failed: %w", err)
	}

	// Emit discovered servers to frontend
	runtime.EventsEmit(a.ctx, "servers:discovered", servers)

	_, lastDiscovery, _ := a.discoveryService.GetServers()
	return &DiscoverServersResponse{
		Message: fmt.Sprintf("Discovery complete. Found %d servers.", len(servers)),
		ScanID:  "scan-" + lastDiscovery.Format("20060102-150405"),
	}, nil
}

// GetServer returns a specific server by ID
func (a *App) GetServer(serverID string) (*models.MCPServer, error) {
	slog.Info("GetServer called", "serverId", serverID)
	server, exists := a.discoveryService.GetServerByID(serverID)
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}
	return server, nil
}

// ========================================
// Lifecycle Methods
// ========================================

// ServerOperationResponse represents a generic server operation response
type ServerOperationResponse struct {
	Message  string `json:"message"`
	ServerID string `json:"serverId"`
	Status   string `json:"status,omitempty"`
}

// StartServer starts a server by ID
func (a *App) StartServer(serverID string) (*ServerOperationResponse, error) {
	slog.Info("StartServer called", "serverId", serverID)

	// Get server info
	server, exists := a.discoveryService.GetServerByID(serverID)
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	// Check transport type (Option D: stdio servers require client configuration)
	if server.Transport == models.TransportStdio {
		return nil, fmt.Errorf("stdio_requires_client: This server uses stdio transport and must be started through an MCP client (e.g., Claude Desktop). Use the configuration editor to add it to your client's config")
	}

	// Start standalone servers directly (http/sse/unknown transports)
	if err := a.lifecycleService.StartServer(server); err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	// Update server in discovery cache
	a.discoveryService.UpdateServer(server)

	// Emit real-time status change event to frontend
	runtime.EventsEmit(a.ctx, "server:status:changed", map[string]interface{}{
		"serverId": serverID,
		"status":   server.Status,
	})

	return &ServerOperationResponse{
		Message:  "Server started successfully",
		ServerID: serverID,
		Status:   string(server.Status.State),
	}, nil
}

// StopServer stops a server by ID
func (a *App) StopServer(serverID string, force bool, timeout int) (*ServerOperationResponse, error) {
	slog.Info("[App.StopServer] Called", "serverId", serverID, "force", force, "timeout", timeout)

	// Get server info
	slog.Info("[App.StopServer] Looking up server", "serverId", serverID)
	server, exists := a.discoveryService.GetServerByID(serverID)
	if !exists {
		slog.Error("[App.StopServer] Server not found", "serverId", serverID)
		return nil, fmt.Errorf("server not found: %s", serverID)
	}
	slog.Info("[App.StopServer] Server found", "serverId", serverID, "name", server.Name, "currentState", server.Status.State, "pid", server.PID)

	// Log current state
	if server.PID != nil {
		slog.Info("[App.StopServer] Server has PID", "pid", *server.PID)
	} else {
		slog.Warn("[App.StopServer] Server has no PID!", "serverId", serverID)
	}

	// Check state
	slog.Info("[App.StopServer] Checking server state", "state", server.Status.State)
	if server.Status.State != models.StatusRunning && server.Status.State != models.StatusStarting {
		slog.Warn("[App.StopServer] Server not in stoppable state", "serverId", serverID, "state", server.Status.State)
		return nil, fmt.Errorf("server is not running (current state: %s)", server.Status.State)
	}

	// Stop the server
	slog.Info("[App.StopServer] Calling lifecycle service StopServer", "serverId", serverID)
	if err := a.lifecycleService.StopServer(server, force, timeout); err != nil {
		slog.Error("[App.StopServer] Lifecycle service returned error", "serverId", serverID, "error", err)
		return nil, fmt.Errorf("failed to stop server: %w", err)
	}

	slog.Info("[App.StopServer] Lifecycle service succeeded", "serverId", serverID)

	// Update server in discovery cache
	slog.Info("[App.StopServer] Updating server in cache", "serverId", serverID, "newState", server.Status.State, "pid", server.PID)
	a.discoveryService.UpdateServer(server)

	// Emit real-time status change event to frontend
	slog.Info("[App.StopServer] Emitting status change event to frontend", "serverId", serverID, "state", server.Status.State)
	runtime.EventsEmit(a.ctx, "server:status:changed", map[string]interface{}{
		"serverId": serverID,
		"status":   server.Status,
	})

	slog.Info("[App.StopServer] Completed successfully", "serverId", serverID)
	return &ServerOperationResponse{
		Message:  "Server stopped successfully",
		ServerID: serverID,
	}, nil
}

// RestartServer restarts a server by ID
func (a *App) RestartServer(serverID string) (*ServerOperationResponse, error) {
	slog.Info("RestartServer called", "serverId", serverID)

	// Get server info
	server, exists := a.discoveryService.GetServerByID(serverID)
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	// Restart the server
	if err := a.lifecycleService.RestartServer(server); err != nil {
		return nil, fmt.Errorf("failed to restart server: %w", err)
	}

	// Update server in discovery cache
	a.discoveryService.UpdateServer(server)

	// Emit real-time status change event to frontend
	runtime.EventsEmit(a.ctx, "server:status:changed", map[string]interface{}{
		"serverId": serverID,
		"status":   server.Status,
	})

	return &ServerOperationResponse{
		Message:  "Server restarted successfully",
		ServerID: serverID,
		Status:   string(server.Status.State),
	}, nil
}

// GetServerStatus returns the current status of a server
func (a *App) GetServerStatus(serverID string) (*models.ServerStatus, error) {
	slog.Info("GetServerStatus called", "serverId", serverID)

	server, exists := a.discoveryService.GetServerByID(serverID)
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	return &server.Status, nil
}

// ========================================
// Configuration Methods
// ========================================

// GetConfiguration returns the configuration for a server
func (a *App) GetConfiguration(serverID string) (*models.ServerConfiguration, error) {
	slog.Info("GetConfiguration called", "serverId", serverID)

	server, exists := a.discoveryService.GetServerByID(serverID)
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	config, err := a.configService.GetConfiguration(server.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get configuration: %w", err)
	}

	return config, nil
}

// UpdateConfiguration updates the configuration for a server
func (a *App) UpdateConfiguration(serverID string, newConfig *models.ServerConfiguration) (*models.ServerConfiguration, error) {
	slog.Info("UpdateConfiguration called", "serverId", serverID)

	server, exists := a.discoveryService.GetServerByID(serverID)
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	if err := a.configService.UpdateConfiguration(server.ID, newConfig); err != nil {
		return nil, fmt.Errorf("failed to update configuration: %w", err)
	}

	return newConfig, nil
}

// ========================================
// Monitoring Methods
// ========================================

// GetLogsResponse represents the response from GetLogs
type GetLogsResponse struct {
	Logs    []models.LogEntry `json:"logs"`
	Total   int               `json:"total"`
	HasMore bool              `json:"hasMore"`
}

// GetLogs returns logs for a specific server
func (a *App) GetLogs(serverID string, severity string, limit int, offset int) (*GetLogsResponse, error) {
	slog.Info("GetLogs called", "serverId", serverID)

	server, exists := a.discoveryService.GetServerByID(serverID)
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	logs := a.monitoringService.GetLogs(server.ID, offset, limit)

	// Apply severity filter if provided
	if severity != "" {
		filtered := []models.LogEntry{}
		for _, log := range logs {
			if strings.EqualFold(string(log.Severity), severity) {
				filtered = append(filtered, log)
			}
		}
		logs = filtered
	}

	return &GetLogsResponse{
		Logs:    logs,
		Total:   len(logs),
		HasMore: false, // TODO: implement proper pagination
	}, nil
}

// GetAllLogs returns all logs with optional filtering
func (a *App) GetAllLogs(serverID string, severity string, search string, limit int) (*GetLogsResponse, error) {
	slog.Info("GetAllLogs called")

	var allLogs []models.LogEntry

	if serverID != "" {
		// Get logs for specific server
		allLogs = a.monitoringService.GetLogs(serverID, 0, 1000)
	} else {
		// Get all logs (iterate through all servers)
		servers, _, _ := a.discoveryService.GetServers()
		for _, server := range servers {
			logs := a.monitoringService.GetLogs(server.ID, 0, 1000)
			allLogs = append(allLogs, logs...)
		}
	}

	// Apply filters
	filtered := []models.LogEntry{}
	for _, log := range allLogs {
		if severity != "" && !strings.EqualFold(string(log.Severity), severity) {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(log.Message), strings.ToLower(search)) {
			continue
		}
		filtered = append(filtered, log)
	}

	// Apply limit
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return &GetLogsResponse{
		Logs:  filtered,
		Total: len(filtered),
	}, nil
}

// GetMetrics returns metrics for a specific server
func (a *App) GetMetrics(serverID string) (*models.ServerMetrics, error) {
	slog.Info("GetMetrics called", "serverId", serverID)

	server, exists := a.discoveryService.GetServerByID(serverID)
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	// Get PID from server, default to 0 if not set
	pid := 0
	if server.PID != nil {
		pid = *server.PID
	}

	metrics, err := a.metricsCollector.GetMetrics(server.ID, &server.Status, pid)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	return metrics, nil
}

// ========================================
// Dependency Methods
// ========================================

// GetDependenciesResponse represents the response from GetDependencies
type GetDependenciesResponse struct {
	Dependencies []models.Dependency `json:"dependencies"`
	AllSatisfied bool                `json:"allSatisfied"`
}

// GetDependencies returns dependencies for a specific server
func (a *App) GetDependencies(serverID string) (*GetDependenciesResponse, error) {
	slog.Info("GetDependencies called", "serverId", serverID)

	server, exists := a.discoveryService.GetServerByID(serverID)
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	deps, err := a.dependencyService.CheckDependencies(server)
	if err != nil {
		return nil, fmt.Errorf("failed to check dependencies: %w", err)
	}

	// Check if all dependencies are satisfied
	allSatisfied := true
	for _, dep := range deps {
		if !dep.IsInstalled() {
			allSatisfied = false
			break
		}
	}

	return &GetDependenciesResponse{
		Dependencies: deps,
		AllSatisfied: allSatisfied,
	}, nil
}

// GetUpdates returns update information for a specific server
func (a *App) GetUpdates(serverID string) (*dependencies.UpdateInfo, error) {
	slog.Info("GetUpdates called", "serverId", serverID)

	server, exists := a.discoveryService.GetServerByID(serverID)
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}

	updateInfo, err := a.updateChecker.CheckForUpdates(server)
	if err != nil {
		return nil, fmt.Errorf("failed to check for updates: %w", err)
	}

	return updateInfo, nil
}

// ========================================
// Utilities Methods
// ========================================

// NetstatResponse represents the response from GetNetstat
type NetstatResponse struct {
	Connections []platform.NetstatEntry `json:"connections"`
}

// GetNetstat retrieves network connections for the specified PIDs
func (a *App) GetNetstat(pids []int) (*NetstatResponse, error) {
	slog.Info("GetNetstat called", "pids", pids)

	entries, err := platform.GetNetstat(pids)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve network connections: %w", err)
	}

	// Handle nil slice
	if entries == nil {
		entries = []platform.NetstatEntry{}
	}

	return &NetstatResponse{
		Connections: entries,
	}, nil
}

// ServicesResponse represents the response from GetServices
type ServicesResponse struct {
	Services []platform.Service `json:"services"`
}

// GetServices retrieves all system services
func (a *App) GetServices() (*ServicesResponse, error) {
	slog.Info("GetServices called")

	services, err := platform.GetServices()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve services: %w", err)
	}

	// Handle nil slice
	if services == nil {
		services = []platform.Service{}
	}

	return &ServicesResponse{
		Services: services,
	}, nil
}

// ========================================
// Application State Methods
// ========================================

// GetApplicationState returns the current application state
func (a *App) GetApplicationState() (*models.ApplicationState, error) {
	slog.Info("GetApplicationState called")

	state, err := a.storageService.LoadState()
	if err != nil {
		return nil, fmt.Errorf("failed to load application state: %w", err)
	}

	return state, nil
}

// UpdateApplicationStateResponse represents the response from UpdateApplicationState
type UpdateApplicationStateResponse struct {
	Message string `json:"message"`
}

// UpdateApplicationState updates the application state
func (a *App) UpdateApplicationState(state *models.ApplicationState) (*UpdateApplicationStateResponse, error) {
	slog.Info("UpdateApplicationState called")

	if err := a.storageService.SaveState(state); err != nil {
		return nil, fmt.Errorf("failed to save application state: %w", err)
	}

	return &UpdateApplicationStateResponse{
		Message: "Application state updated successfully",
	}, nil
}
