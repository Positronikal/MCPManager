package discovery

import (
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/Positronikal/MCPManager/internal/models"
)

func TestNewDiscoveryService(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewDiscoveryService(resolver, eventBus)
	if service == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestDiscoveryService_Discover(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewDiscoveryService(resolver, eventBus)

	servers, err := service.Discover()
	if err != nil {
		t.Errorf("Discover should not error: %v", err)
	}

	if servers == nil {
		t.Error("Should return non-nil slice")
	}

	// Verify last discovery time was updated
	lastTime := service.GetLastDiscoveryTime()
	if lastTime.IsZero() {
		t.Error("Last discovery time should be set")
	}
}

func TestDiscoveryService_Cache(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewDiscoveryService(resolver, eventBus)

	// Initially cache should be empty
	cached := service.GetCachedServers()
	if len(cached) != 0 {
		t.Error("Cache should be empty initially")
	}

	// Run discovery
	_, err := service.Discover()
	if err != nil {
		t.Fatal(err)
	}

	// Cache should be updated (may or may not have servers depending on system)
	// Just verify the cache works
	cached = service.GetCachedServers()
	if cached == nil {
		t.Error("Cached servers should not be nil")
	}
}

func TestDiscoveryService_GetServerByID(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewDiscoveryService(resolver, eventBus)

	// Add a server to cache
	server := models.NewMCPServer("test", "/path", models.DiscoveryClientConfig)
	service.UpdateServer(server)

	// Retrieve by ID
	retrieved, exists := service.GetServerByID(server.ID)
	if !exists {
		t.Fatal("Server should exist in cache")
	}

	if retrieved.ID != server.ID {
		t.Error("Retrieved server ID mismatch")
	}

	// Try non-existent ID
	_, exists = service.GetServerByID("nonexistent")
	if exists {
		t.Error("Should not find non-existent server")
	}
}

func TestDiscoveryService_UpdateServer(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewDiscoveryService(resolver, eventBus)

	// Create and add server
	server := models.NewMCPServer("test", "/path", models.DiscoveryClientConfig)
	originalID := server.ID

	service.UpdateServer(server)

	// Modify server
	server.Version = "1.0.0"
	service.UpdateServer(server)

	// Retrieve and verify update
	retrieved, exists := service.GetServerByID(originalID)
	if !exists {
		t.Fatal("Server should exist")
	}

	if retrieved.Version != "1.0.0" {
		t.Error("Server version should be updated")
	}
}

func TestDiscoveryService_RemoveServer(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewDiscoveryService(resolver, eventBus)

	// Add server
	server := models.NewMCPServer("test", "/path", models.DiscoveryClientConfig)
	service.UpdateServer(server)

	// Verify it exists
	_, exists := service.GetServerByID(server.ID)
	if !exists {
		t.Fatal("Server should exist")
	}

	// Remove it
	service.RemoveServer(server.ID)

	// Verify it's gone
	_, exists = service.GetServerByID(server.ID)
	if exists {
		t.Error("Server should be removed")
	}
}

func TestDiscoveryService_ConcurrentAccess(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewDiscoveryService(resolver, eventBus)

	// Test concurrent reads and writes
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 10; i++ {
			server := models.NewMCPServer("test", "/path", models.DiscoveryClientConfig)
			service.UpdateServer(server)
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 10; i++ {
			_ = service.GetCachedServers()
			time.Sleep(time.Millisecond)
		}
		done <- true
	}()

	// Wait for both
	<-done
	<-done

	// If we get here without deadlock, test passes
}

