package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hoytech/mcpmanager/internal/api"
	"github.com/hoytech/mcpmanager/internal/core/config"
	"github.com/hoytech/mcpmanager/internal/core/dependencies"
	"github.com/hoytech/mcpmanager/internal/core/discovery"
	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/core/lifecycle"
	"github.com/hoytech/mcpmanager/internal/core/monitoring"
	"github.com/hoytech/mcpmanager/internal/platform"
	"github.com/hoytech/mcpmanager/internal/storage"
)

const (
	serverAddr = "localhost:8080"
	appVersion = "0.1.0"
)

func main() {
	// Configure structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting MCP Manager", "version", appVersion)

	// Initialize platform-specific components
	pathResolver := platform.NewPathResolver()
	processManager := platform.NewProcessManager()
	processInfo := platform.NewProcessInfo()

	// Initialize EventBus (central event dispatcher)
	eventBus := events.NewEventBus()
	slog.Info("EventBus initialized")

	// Initialize storage service
	storageService, err := storage.NewFileStorage()
	if err != nil {
		slog.Error("Failed to initialize storage service", "error", err)
		os.Exit(1)
	}
	slog.Info("Storage service initialized")

	// Initialize core services
	discoveryService := discovery.NewDiscoveryService(pathResolver, eventBus)
	slog.Info("Discovery service initialized")

	lifecycleService := lifecycle.NewLifecycleService(processManager, eventBus)
	slog.Info("Lifecycle service initialized")

	configService, err := config.NewConfigService(eventBus)
	if err != nil {
		slog.Error("Failed to initialize config service", "error", err)
		os.Exit(1)
	}
	slog.Info("Config service initialized")

	monitoringService := monitoring.NewMonitoringService(eventBus)
	slog.Info("Monitoring service initialized")

	metricsCollector := monitoring.NewMetricsCollector(processInfo, eventBus)
	slog.Info("Metrics collector initialized")

	dependencyService := dependencies.NewDependencyService()
	slog.Info("Dependency service initialized")

	updateChecker := dependencies.NewUpdateChecker()
	slog.Info("Update checker initialized")

	// Perform initial discovery
	slog.Info("Running initial server discovery...")
	servers, err := discoveryService.Discover()
	if err != nil {
		slog.Warn("Initial discovery returned error", "error", err)
	} else {
		slog.Info("Initial discovery complete", "servers_found", len(servers))
	}

	// Create services struct for API handlers
	services := &api.Services{
		DiscoveryService:  discoveryService,
		LifecycleService:  lifecycleService,
		ConfigService:     configService,
		MonitoringService: monitoringService,
		MetricsCollector:  metricsCollector,
		DependencyService: dependencyService,
		UpdateChecker:     updateChecker,
		StorageService:    storageService,
		EventBus:          eventBus,
	}

	// Create router with all API endpoints
	router := api.NewRouter(services)
	slog.Info("API router configured with all endpoints")

	// Create HTTP server
	server := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("HTTP server starting", "address", serverAddr)
		slog.Info("API documentation available at http://localhost:8080/api/v1")
		serverErrors <- server.ListenAndServe()
	}()

	// Setup graceful shutdown on signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Wait for either server error or shutdown signal
	select {
	case err := <-serverErrors:
		slog.Error("Server error", "error", err)
		os.Exit(1)

	case sig := <-shutdown:
		slog.Info("Shutdown signal received", "signal", sig.String())

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		slog.Info("Shutting down HTTP server gracefully...")
		if err := server.Shutdown(ctx); err != nil {
			slog.Error("Error during server shutdown", "error", err)
			// Force close
			if closeErr := server.Close(); closeErr != nil {
				slog.Error("Error forcing server close", "error", closeErr)
			}
		}

		// Stop lifecycle service (gracefully stop managed servers)
		slog.Info("Stopping managed servers...")
		lifecycleService.StopAll()

		// Close EventBus
		slog.Info("Closing EventBus...")
		eventBus.Close()

		slog.Info("Shutdown complete")
	}
}
