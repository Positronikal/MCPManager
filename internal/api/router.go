package api

import (
	"github.com/Positronikal/MCPManager/internal/core/config"
	"github.com/Positronikal/MCPManager/internal/core/dependencies"
	"github.com/Positronikal/MCPManager/internal/core/discovery"
	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/Positronikal/MCPManager/internal/core/lifecycle"
	"github.com/Positronikal/MCPManager/internal/core/monitoring"
	"github.com/Positronikal/MCPManager/internal/storage"
	"github.com/go-chi/chi/v5"
)

// Services contains all application services
type Services struct {
	DiscoveryService  *discovery.DiscoveryService
	LifecycleService  *lifecycle.LifecycleService
	ConfigService     *config.ConfigService
	MonitoringService *monitoring.MonitoringService
	MetricsCollector  *monitoring.MetricsCollector
	DependencyService *dependencies.DependencyService
	UpdateChecker     *dependencies.UpdateChecker
	StorageService    storage.StorageService
	EventBus          *events.EventBus
}

// NewRouter creates and configures the Chi router with all endpoints
func NewRouter(services *Services) *chi.Mux {
	r := chi.NewRouter()

	// Apply middleware
	r.Use(LoggingMiddleware)
	r.Use(ErrorHandlingMiddleware)
	r.Use(CORSMiddleware)

	// Create handler instances
	discoveryHandlers := NewDiscoveryHandlers(services.DiscoveryService)
	lifecycleHandlers := NewLifecycleHandlers(services.LifecycleService, services.DiscoveryService)
	configHandlers := NewConfigHandlers(services.ConfigService, services.DiscoveryService)
	monitoringHandlers := NewMonitoringHandlers(services.MonitoringService, services.MetricsCollector, services.DiscoveryService)
	dependencyHandlers := NewDependencyHandlers(services.DependencyService, services.UpdateChecker, services.DiscoveryService)
	appStateHandlers := NewAppStateHandlers(services.StorageService)
	sseHandlers := NewSSEHandlers(services.EventBus)

	// Define API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Discovery endpoints
		r.Get("/servers", discoveryHandlers.ListServers)
		r.Post("/servers/discover", discoveryHandlers.DiscoverServers)
		r.Get("/servers/{serverId}", discoveryHandlers.GetServerByID)

		// Lifecycle endpoints
		r.Post("/servers/{serverId}/start", lifecycleHandlers.StartServer)
		r.Post("/servers/{serverId}/stop", lifecycleHandlers.StopServer)
		r.Post("/servers/{serverId}/restart", lifecycleHandlers.RestartServer)
		r.Get("/servers/{serverId}/status", lifecycleHandlers.GetServerStatus)

		// Configuration endpoints
		r.Get("/servers/{serverId}/configuration", configHandlers.GetConfiguration)
		r.Put("/servers/{serverId}/configuration", configHandlers.UpdateConfiguration)

		// Monitoring endpoints
		r.Get("/servers/{serverId}/logs", monitoringHandlers.GetServerLogs)
		r.Get("/logs", monitoringHandlers.GetAllLogs)
		r.Get("/servers/{serverId}/metrics", monitoringHandlers.GetServerMetrics)
		r.Get("/netstat", monitoringHandlers.GetNetstat)
		r.Get("/services", monitoringHandlers.GetServices)

		// Dependency endpoints
		r.Get("/servers/{serverId}/dependencies", dependencyHandlers.GetServerDependencies)
		r.Get("/servers/{serverId}/updates", dependencyHandlers.GetServerUpdates)

		// Application state endpoints
		r.Get("/application/state", appStateHandlers.GetApplicationState)
		r.Put("/application/state", appStateHandlers.UpdateApplicationState)

		// SSE events endpoint
		r.Get("/events", sseHandlers.SSEStream)
	})

	return r
}
