package contract

import (
	"testing"

	"github.com/Positronikal/MCPManager/internal/api"
	"github.com/Positronikal/MCPManager/internal/core/config"
	"github.com/Positronikal/MCPManager/internal/core/dependencies"
	"github.com/Positronikal/MCPManager/internal/core/discovery"
	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/Positronikal/MCPManager/internal/core/lifecycle"
	"github.com/Positronikal/MCPManager/internal/core/monitoring"
	"github.com/Positronikal/MCPManager/internal/platform"
	"github.com/Positronikal/MCPManager/internal/storage"
)

// setupFullTestServices creates a complete services setup for integration-style contract tests
func setupFullTestServices(t *testing.T) (*api.Services, func()) {
	pathResolver := platform.NewPathResolver()
	processManager := platform.NewProcessManager()
	processInfo := platform.NewProcessInfo()
	eventBus := events.NewEventBus()

	discoveryService := discovery.NewDiscoveryService(pathResolver, eventBus)
	discoveryService.Discover()

	monitoringService := monitoring.NewMonitoringService(eventBus)
	lifecycleService := lifecycle.NewLifecycleService(processManager, discoveryService, monitoringService, eventBus)

	configService, err := config.NewConfigService(eventBus)
	if err != nil {
		t.Fatalf("Failed to create config service: %v", err)
	}

	metricsCollector := monitoring.NewMetricsCollector(processInfo, eventBus)
	dependencyService := dependencies.NewDependencyService()
	updateChecker := dependencies.NewUpdateChecker()

	storageService, err := storage.NewFileStorage()
	if err != nil {
		t.Fatalf("Failed to create storage service: %v", err)
	}

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

	cleanup := func() {
		lifecycleService.StopAll()
		discoveryService.Close()
		eventBus.Close()
	}

	return services, cleanup
}
