package contract

import (
	"testing"

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

// setupFullTestServices creates a complete services setup for integration-style contract tests
func setupFullTestServices(t *testing.T) (*api.Services, func()) {
	pathResolver := platform.NewPathResolver()
	processManager := platform.NewProcessManager()
	processInfo := platform.NewProcessInfo()
	eventBus := events.NewEventBus()

	discoveryService := discovery.NewDiscoveryService(pathResolver, eventBus)
	discoveryService.Discover()

	lifecycleService := lifecycle.NewLifecycleService(processManager, discoveryService, eventBus)

	configService, err := config.NewConfigService(eventBus)
	if err != nil {
		t.Fatalf("Failed to create config service: %v", err)
	}

	monitoringService := monitoring.NewMonitoringService(eventBus)
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
