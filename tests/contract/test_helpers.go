package contract

import (
	"github.com/hoytech/mcpmanager/internal/api"
	"github.com/hoytech/mcpmanager/internal/core/discovery"
	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/platform"
)

// createTestRouter creates a router with test services for contract testing
func createTestRouter() *api.Services {
	pathResolver := platform.NewPathResolver()
	eventBus := events.NewEventBus()
	discoveryService := discovery.NewDiscoveryService(pathResolver, eventBus)

	// Run initial discovery to have test data
	discoveryService.Discover()

	return &api.Services{
		DiscoveryService: discoveryService,
		EventBus:         eventBus,
	}
}
