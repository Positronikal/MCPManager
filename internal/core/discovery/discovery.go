package discovery

import (
	"sync"
	"time"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
	"github.com/hoytech/mcpmanager/internal/platform"
)

// DiscoveryService orchestrates all discovery sources
type DiscoveryService struct {
	clientConfigDiscovery *ClientConfigDiscovery
	filesystemDiscovery   *FilesystemDiscovery
	processDiscovery      *ProcessDiscovery
	eventBus              *events.EventBus
	mu                    sync.RWMutex
	cachedServers         map[string]*models.MCPServer // serverID -> server
	lastDiscovery         time.Time
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(pathResolver platform.PathResolver, eventBus *events.EventBus) *DiscoveryService {
	return &DiscoveryService{
		clientConfigDiscovery: NewClientConfigDiscovery(pathResolver, eventBus),
		filesystemDiscovery:   NewFilesystemDiscovery(pathResolver, eventBus),
		processDiscovery:      NewProcessDiscovery(eventBus),
		eventBus:              eventBus,
		cachedServers:         make(map[string]*models.MCPServer),
		lastDiscovery:         time.Time{},
	}
}

// Discover runs all discovery sources and returns a deduplicated list of servers
// Priority order: client_config > filesystem > process
func (ds *DiscoveryService) Discover() ([]models.MCPServer, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	var allServers []models.MCPServer

	// 1. Discover from client configs (highest priority)
	clientServers, err := ds.clientConfigDiscovery.DiscoverFromClientConfigs()
	if err == nil {
		allServers = append(allServers, clientServers...)
	}

	// 2. Discover from filesystem (secondary)
	filesystemServers, err := ds.filesystemDiscovery.DiscoverFromFilesystem()
	if err == nil {
		allServers = append(allServers, filesystemServers...)
	}

	// 3. Discover from running processes
	processServers, err := ds.processDiscovery.DiscoverFromProcesses()
	if err == nil {
		allServers = append(allServers, processServers...)
	}

	// Deduplicate and merge servers
	mergedServers := ds.deduplicateServers(allServers)

	// Update cache
	ds.cachedServers = make(map[string]*models.MCPServer)
	for i := range mergedServers {
		ds.cachedServers[mergedServers[i].ID] = &mergedServers[i]
	}
	ds.lastDiscovery = time.Now()

	return mergedServers, nil
}

// deduplicateServers merges servers found from multiple sources
// Priority: client_config > filesystem > process
func (ds *DiscoveryService) deduplicateServers(servers []models.MCPServer) []models.MCPServer {
	// Create a map to track unique servers by name and result index
	seen := make(map[string]int) // serverName -> index in result slice

	var result []models.MCPServer

	for i := range servers {
		server := servers[i]

		// Create a key for deduplication (use name as primary key)
		key := server.Name

		existingIdx, exists := seen[key]
		if !exists {
			// New server, add it
			seen[key] = len(result)
			result = append(result, server)
		} else {
			// Server already exists, merge based on priority
			existing := &result[existingIdx]

			if ds.shouldReplace(existing, &server) {
				// Replace with higher priority server
				*existing = server
			} else if server.Source == models.DiscoveryProcess {
				// If the new discovery is from a process, update PID and status
				if server.PID != nil {
					existing.PID = server.PID
					existing.Status.State = models.StatusRunning
				}
			}
		}
	}

	return result
}

// shouldReplace determines if the new server should replace the existing one
// based on discovery source priority
func (ds *DiscoveryService) shouldReplace(existing, new *models.MCPServer) bool {
	// Priority ranking
	priority := map[models.DiscoverySource]int{
		models.DiscoveryClientConfig: 3, // Highest
		models.DiscoveryFilesystem:   2,
		models.DiscoveryProcess:      1, // Lowest
	}

	existingPriority := priority[existing.Source]
	newPriority := priority[new.Source]

	return newPriority > existingPriority
}

// GetCachedServers returns the cached list of discovered servers
func (ds *DiscoveryService) GetCachedServers() []models.MCPServer {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var servers []models.MCPServer
	for _, server := range ds.cachedServers {
		servers = append(servers, *server)
	}

	return servers
}

// GetServerByID returns a specific server from the cache
func (ds *DiscoveryService) GetServerByID(serverID string) (*models.MCPServer, bool) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	server, exists := ds.cachedServers[serverID]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent external modification
	serverCopy := *server
	return &serverCopy, true
}

// GetLastDiscoveryTime returns when the last discovery was performed
func (ds *DiscoveryService) GetLastDiscoveryTime() time.Time {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return ds.lastDiscovery
}

// UpdateServer updates a server in the cache
func (ds *DiscoveryService) UpdateServer(server *models.MCPServer) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if server != nil {
		ds.cachedServers[server.ID] = server
	}
}

// RemoveServer removes a server from the cache
func (ds *DiscoveryService) RemoveServer(serverID string) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	delete(ds.cachedServers, serverID)
}
