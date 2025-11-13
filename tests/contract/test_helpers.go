package contract

import (
	"github.com/Positronikal/MCPManager/internal/api"
	"github.com/Positronikal/MCPManager/internal/core/discovery"
	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/Positronikal/MCPManager/internal/platform"
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
