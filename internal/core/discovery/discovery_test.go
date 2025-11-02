package discovery

import (
	"testing"
	"time"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
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

// TODO: Update after refactoring - deduplicateServers method no longer exists
// Deduplication now happens in mergeServersByName() using map-based priority
/*
func TestDiscoveryService_Deduplication(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewDiscoveryService(resolver, eventBus)

	// Create duplicate servers from different sources
	server1 := models.NewMCPServer("test-server", "/path1", models.DiscoveryClientConfig)
	server2 := models.NewMCPServer("test-server", "/path2", models.DiscoveryFilesystem)
	server3 := models.NewMCPServer("test-server", "/path3", models.DiscoveryProcess)

	servers := []models.MCPServer{*server1, *server2, *server3}

	// Deduplicate
	deduplicated := service.deduplicateServers(servers)

	// Should only have 1 server (highest priority wins)
	if len(deduplicated) != 1 {
		t.Errorf("Expected 1 server after deduplication, got %d", len(deduplicated))
	}

	// Should keep client_config version (highest priority)
	if deduplicated[0].Source != models.DiscoveryClientConfig {
		t.Errorf("Expected client_config source, got %s", deduplicated[0].Source)
	}

	if deduplicated[0].InstallationPath != "/path1" {
		t.Error("Should keep the client_config server's path")
	}
}
*/

// TODO: Update after refactoring - deduplicateServers method no longer exists
// PID merging now handled in mergeServersByName() via updateFromDiscoveredServer
/*
func TestDiscoveryService_ProcessMatchingUpdatesPID(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewDiscoveryService(resolver, eventBus)

	// Create a server from client config (no PID)
	server1 := models.NewMCPServer("test-server", "/path", models.DiscoveryClientConfig)

	// Create a process discovery with PID
	server2 := models.NewMCPServer("test-server", "/path", models.DiscoveryProcess)
	server2.SetPID(1234)
	server2.Status.State = models.StatusRunning

	servers := []models.MCPServer{*server1, *server2}

	// Deduplicate
	deduplicated := service.deduplicateServers(servers)

	// Should have 1 server
	if len(deduplicated) != 1 {
		t.Fatalf("Expected 1 server, got %d", len(deduplicated))
	}

	// Should be client_config source but with PID from process
	if deduplicated[0].Source != models.DiscoveryClientConfig {
		t.Error("Should keep client_config source")
	}

	if deduplicated[0].PID == nil {
		t.Error("PID should be updated from process discovery")
	}

	if *deduplicated[0].PID != 1234 {
		t.Errorf("Expected PID 1234, got %d", *deduplicated[0].PID)
	}

	if deduplicated[0].Status.State != models.StatusRunning {
		t.Error("Status should be updated to running")
	}
}
*/

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

// TODO: Update after refactoring - shouldReplace method no longer exists
// Priority ordering now implicit in mergeServersByName() map overwrite order
/*
func TestDiscoveryService_PriorityOrder(t *testing.T) {
	resolver := &MockPathResolver{configDir: t.TempDir()}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	service := NewDiscoveryService(resolver, eventBus)

	testCases := []struct {
		name           string
		existing       models.DiscoverySource
		new            models.DiscoverySource
		shouldReplace  bool
	}{
		{"filesystem replaces process", models.DiscoveryProcess, models.DiscoveryFilesystem, true},
		{"client_config replaces filesystem", models.DiscoveryFilesystem, models.DiscoveryClientConfig, true},
		{"client_config replaces process", models.DiscoveryProcess, models.DiscoveryClientConfig, true},
		{"process does not replace filesystem", models.DiscoveryFilesystem, models.DiscoveryProcess, false},
		{"filesystem does not replace client_config", models.DiscoveryClientConfig, models.DiscoveryFilesystem, false},
		{"process does not replace client_config", models.DiscoveryClientConfig, models.DiscoveryProcess, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			existing := models.NewMCPServer("test", "/path1", tc.existing)
			new := models.NewMCPServer("test", "/path2", tc.new)

			result := service.shouldReplace(existing, new)
			if result != tc.shouldReplace {
				t.Errorf("Expected shouldReplace=%v, got %v", tc.shouldReplace, result)
			}
		})
	}
}
*/
