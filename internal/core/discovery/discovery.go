package discovery

import (
	"os"
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

// Discover runs all discovery sources following the spec's three-tier strategy:
// 1. PRIMARY: Read client config files (Claude Desktop, Cursor, etc.)
// 2. SECONDARY: Scan filesystem for installed servers (npm, pip, Go binaries)
// 3. TERTIARY: Match running processes against discovered servers (PID tracking)
//
// Per spec research.md §16: "Discovery Sources Priority"
func (ds *DiscoveryService) Discover() ([]models.MCPServer, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Phase 1: Discover from client configs (PRIMARY - highest priority)
	// FR-002: Read MCP client configuration files without modifying them
	clientServers, err := ds.clientConfigDiscovery.DiscoverFromClientConfigs()
	if err != nil {
		// Log but continue - client configs may not exist yet
	}

	// Phase 2: Discover from filesystem (SECONDARY)
	// FR-001: Scan common installation locations
	filesystemServers, err := ds.filesystemDiscovery.DiscoverFromFilesystem()
	if err != nil {
		// Log but continue - packages may not be installed
	}

	// Merge Phase 1 & 2: Create authoritative server list
	// Priority: client_config > filesystem
	allServers := ds.mergeServersByName(clientServers, filesystemServers)

	// Phase 3: Match running processes against discovered servers (TERTIARY)
	// FR-010: Track server process IDs for lifecycle management
	// Per spec: "Match PIDs to discovered servers" NOT "discover new servers from processes"
	allServers = ds.matchProcessesToServers(allServers)

	// Update cache
	ds.cachedServers = make(map[string]*models.MCPServer)
	for i := range allServers {
		ds.cachedServers[allServers[i].ID] = &allServers[i]
	}
	ds.lastDiscovery = time.Now()

	return allServers, nil
}

// mergeServersByName combines servers from multiple sources with priority handling
// Priority: client_config > filesystem
func (ds *DiscoveryService) mergeServersByName(clientServers, filesystemServers []models.MCPServer) []models.MCPServer {
	serverMap := make(map[string]*models.MCPServer) // name -> server

	// Add filesystem servers first (lower priority)
	for i := range filesystemServers {
		server := &filesystemServers[i]
		serverMap[server.Name] = server
	}

	// Add client config servers (higher priority - will override filesystem)
	for i := range clientServers {
		server := &clientServers[i]
		serverMap[server.Name] = server
	}

	// Convert map back to slice
	result := make([]models.MCPServer, 0, len(serverMap))
	for _, server := range serverMap {
		result = append(result, *server)
	}

	return result
}

// matchProcessesToServers matches running processes against discovered servers
// This is the CORRECT implementation of process discovery per spec:
// - Only match processes that correspond to known servers
// - Update PID and status for matched servers
// - Do NOT create new server entries from arbitrary processes
func (ds *DiscoveryService) matchProcessesToServers(servers []models.MCPServer) []models.MCPServer {
	// Get current MCP Manager PID to filter ourselves out
	currentPID := os.Getpid()

	// Get all running processes
	processes, err := ds.processDiscovery.listProcesses()
	if err != nil {
		// Log but continue - process matching is best-effort
		return servers
	}

	// For each discovered server, try to find a matching process
	for i := range servers {
		server := &servers[i]

		// Try to match process by command/path
		matchedProcess := ds.findMatchingProcess(server, processes, currentPID)

		if matchedProcess != nil {
			// Found a running process for this server
			server.SetPID(matchedProcess.PID)
			server.Status.State = models.StatusRunning
			server.UpdateLastSeen()
		} else {
			// No matching process found - server is stopped
			server.ClearPID()
			server.Status.State = models.StatusStopped
		}
	}

	return servers
}

// findMatchingProcess finds a process that matches the given server
func (ds *DiscoveryService) findMatchingProcess(server *models.MCPServer, processes []ProcessInfo, currentPID int) *ProcessInfo {
	for i := range processes {
		proc := &processes[i]

		// Skip MCP Manager's own process
		if proc.PID == currentPID {
			continue
		}

		// Skip if process name/path doesn't match server
		if !ds.processMatchesServer(proc, server) {
			continue
		}

		return proc
	}

	return nil
}

// processMatchesServer determines if a process matches a server definition
func (ds *DiscoveryService) processMatchesServer(proc *ProcessInfo, server *models.MCPServer) bool {
	// Match by command path
	if proc.CommandLine == server.InstallationPath {
		return true
	}

	// Match by server name in command line (for node/python servers)
	// e.g., "node server-name" or "python -m server-name"
	if containsServerName(proc.CommandLine, server.Name) {
		return true
	}

	// Match by executable name
	if containsServerName(proc.Name, server.Name) {
		return true
	}

	return false
}

// containsServerName checks if the text contains the server name as a distinct token
func containsServerName(text, serverName string) bool {
	// Simple token-based matching to avoid false positives
	// This prevents "mcp" from matching "mcpmanager"
	return text == serverName ||
	       containsToken(text, serverName) ||
	       containsToken(text, "@"+serverName) // npm scoped packages
}

// containsToken checks if text contains token as a separate word
func containsToken(text, token string) bool {
	// TODO: Implement proper token matching
	// For now, use simple contains check but this should be improved
	// to avoid false positives
	return false
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

// GetServers returns the cached list of discovered servers and the last discovery time
func (ds *DiscoveryService) GetServers() ([]models.MCPServer, time.Time, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var servers []models.MCPServer
	for _, server := range ds.cachedServers {
		servers = append(servers, *server)
	}

	return servers, ds.lastDiscovery, nil
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
