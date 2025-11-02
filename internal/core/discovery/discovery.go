package discovery

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
	"github.com/hoytech/mcpmanager/internal/platform"
)

// DiscoveryService orchestrates all discovery sources
type DiscoveryService struct {
	clientConfigDiscovery *ClientConfigDiscovery
	extensionsDiscovery   *ClaudeExtensionsDiscovery
	filesystemDiscovery   *FilesystemDiscovery
	processDiscovery      *ProcessDiscovery
	configFileWatcher     *ConfigFileWatcher // FR-050: Monitor config files for external changes
	eventBus              *events.EventBus
	mu                    sync.RWMutex
	cachedServers         map[string]*models.MCPServer // serverID -> server
	lastDiscovery         time.Time
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(pathResolver platform.PathResolver, eventBus *events.EventBus) *DiscoveryService {
	// FR-050: Get config file paths to watch
	configDir := pathResolver.GetConfigDir()
	configPaths := []string{
		fmt.Sprintf("%s/Claude/claude_desktop_config.json", configDir),
		fmt.Sprintf("%s/Cursor/mcp_config.json", configDir),
	}

	// FR-050: Initialize file watcher for client config files
	watcher, err := NewConfigFileWatcher(eventBus, configPaths)
	if err != nil {
		// Log error but continue - file watching is non-critical
		fmt.Printf("Warning: Failed to create config file watcher: %v\n", err)
	} else {
		// Start watching
		if err := watcher.Start(); err != nil {
			fmt.Printf("Warning: Failed to start config file watcher: %v\n", err)
		}
	}

	return &DiscoveryService{
		clientConfigDiscovery: NewClientConfigDiscovery(pathResolver, eventBus),
		extensionsDiscovery:   NewClaudeExtensionsDiscovery(pathResolver, eventBus),
		filesystemDiscovery:   NewFilesystemDiscovery(pathResolver, eventBus),
		processDiscovery:      NewProcessDiscovery(eventBus),
		configFileWatcher:     watcher,
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
	fmt.Println("\n=== MCP SERVER DISCOVERY START ===")

	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Phase 1: Discover from client configs (PRIMARY - highest priority)
	// FR-002: Read MCP client configuration files without modifying them
	fmt.Println("\n[PHASE 1] Discovering from client configs...")
	clientServers, err := ds.clientConfigDiscovery.DiscoverFromClientConfigs()
	if err != nil {
		fmt.Printf("[PHASE 1] ERROR: %v\n", err)
		// Log but continue - client configs may not exist yet
	} else {
		fmt.Printf("[PHASE 1] Found %d servers from client configs\n", len(clientServers))
		for i, srv := range clientServers {
			fmt.Printf("  [%d] %s (cmd: %s, source: %s)\n", i+1, srv.Name, srv.InstallationPath, srv.Source)
		}
	}

	// Phase 1.5: Discover from Claude Extensions (HIGH priority - after client configs)
	fmt.Println("\n[PHASE 1.5] Discovering from Claude Extensions...")
	extensionServers, err := ds.extensionsDiscovery.DiscoverFromExtensions()
	if err != nil {
		fmt.Printf("[PHASE 1.5] ERROR: %v\n", err)
		// Log but continue - extensions may not exist
	} else {
		fmt.Printf("[PHASE 1.5] Found %d servers from Claude Extensions\n", len(extensionServers))
		for i, srv := range extensionServers {
			fmt.Printf("  [%d] %s (cmd: %s, source: %s, version: %s)\n", i+1, srv.Name, srv.InstallationPath, srv.Source, srv.Version)
		}
	}

	// Phase 2: Discover from filesystem (SECONDARY)
	// FR-001: Scan common installation locations
	fmt.Println("\n[PHASE 2] Discovering from filesystem...")
	filesystemServers, err := ds.filesystemDiscovery.DiscoverFromFilesystem()
	if err != nil {
		fmt.Printf("[PHASE 2] ERROR: %v\n", err)
		// Log but continue - packages may not be installed
	} else {
		fmt.Printf("[PHASE 2] Found %d servers from filesystem\n", len(filesystemServers))
		for i, srv := range filesystemServers {
			fmt.Printf("  [%d] %s (path: %s, source: %s)\n", i+1, srv.Name, srv.InstallationPath, srv.Source)
		}
	}

	// Merge Phase 1, 1.5, & 2: Create authoritative server list
	// Priority: client_config > extensions > filesystem
	fmt.Println("\n[MERGE] Merging servers from all sources...")
	allServers := ds.mergeServersByName(clientServers, extensionServers, filesystemServers)
	fmt.Printf("[MERGE] Total unique servers after merge: %d\n", len(allServers))

	// Phase 3: Match running processes against discovered servers (TERTIARY)
	// FR-010: Track server process IDs for lifecycle management
	// Per spec: "Match PIDs to discovered servers" NOT "discover new servers from processes"
	fmt.Println("\n[PHASE 3] Matching running processes to discovered servers...")
	allServers = ds.matchProcessesToServers(allServers)

	runningCount := 0
	for _, srv := range allServers {
		if srv.Status.State == models.StatusRunning {
			runningCount++
		}
	}
	fmt.Printf("[PHASE 3] Matched %d running processes\n", runningCount)

	// Update cache - preserve existing servers and merge new discoveries
	fmt.Println("\n[CACHE UPDATE] Merging discovered servers into cache...")
	newCache := make(map[string]*models.MCPServer)
	for i := range allServers {
		serverID := allServers[i].ID
		discoveredServer := &allServers[i]

		// Check if this server already exists in cache
		if existingServer, exists := ds.cachedServers[serverID]; exists {
			fmt.Printf("  Server %s already in cache (state: %s), updating with discovery results\n",
				existingServer.Name, existingServer.Status.State)

			// Preserve PID and status if process is still running but not matched
			// (e.g., server was stopped manually but process hasn't died yet)
			if existingServer.Status.State == models.StatusStopped && discoveredServer.Status.State == models.StatusStopped {
				// Both stopped - use discovered server
				newCache[serverID] = discoveredServer
			} else if discoveredServer.Status.State == models.StatusRunning {
				// Process found during discovery - use discovered state
				newCache[serverID] = discoveredServer
			} else {
				// No process found - preserve existing state if it's stopped
				if existingServer.Status.State == models.StatusStopped {
					newCache[serverID] = existingServer
				} else {
					// Server was running, now stopped
					newCache[serverID] = discoveredServer
				}
			}
		} else {
			// New server discovery
			fmt.Printf("  New server discovered: %s (state: %s)\n",
				discoveredServer.Name, discoveredServer.Status.State)
			newCache[serverID] = discoveredServer
		}
	}

	ds.cachedServers = newCache
	ds.lastDiscovery = time.Now()

	fmt.Printf("\n=== DISCOVERY COMPLETE: %d total servers ===\n\n", len(allServers))
	return allServers, nil
}

// mergeServersByName combines servers from multiple sources with priority handling
// Priority: client_config > extensions > filesystem
func (ds *DiscoveryService) mergeServersByName(clientServers, extensionServers, filesystemServers []models.MCPServer) []models.MCPServer {
	serverMap := make(map[string]*models.MCPServer) // name -> server

	// Add filesystem servers first (lowest priority)
	for i := range filesystemServers {
		server := &filesystemServers[i]
		serverMap[server.Name] = server
	}

	// Add extension servers (medium priority - will override filesystem)
	for i := range extensionServers {
		server := &extensionServers[i]
		serverMap[server.Name] = server
	}

	// Add client config servers (highest priority - will override extensions and filesystem)
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
	fmt.Println("  Getting running processes...")
	processes, err := ds.processDiscovery.listProcesses()
	if err != nil {
		fmt.Printf("  ERROR getting processes: %v\n", err)
		// Log but continue - process matching is best-effort
		return servers
	}

	fmt.Printf("  Found %d processes to match against\n", len(processes))

	// Debug: Show relevant processes
	for _, proc := range processes {
		if proc.PID == currentPID {
			continue
		}
		fmt.Printf("    PID %d: %s\n", proc.PID, proc.Name)
		fmt.Printf("      CMD: %s\n", proc.CommandLine)
	}

	// For each discovered server, try to find a matching process
	for i := range servers {
		server := &servers[i]

		fmt.Printf("  Matching server: %s (cmd: %s)\n", server.Name, server.InstallationPath)

		// Try to match process by command/path
		matchedProcess := ds.findMatchingProcess(server, processes, currentPID)

		if matchedProcess != nil {
			// Found a running process for this server
			fmt.Printf("    ✓ MATCHED PID %d\n", matchedProcess.PID)
			server.SetPID(matchedProcess.PID)
			server.Status.State = models.StatusRunning
			server.UpdateLastSeen()
		} else {
			// No matching process found - server is stopped
			fmt.Printf("    ✗ No match found\n")
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

		// Try to match process to server
		if ds.processMatchesServer(proc, server) {
			return proc
		}
	}

	return nil
}

// processMatchesServer determines if a process matches a server definition
// Uses sophisticated matching based on command + arguments
func (ds *DiscoveryService) processMatchesServer(proc *ProcessInfo, server *models.MCPServer) bool {
	cmdLine := strings.ToLower(proc.CommandLine)
	serverCmd := strings.ToLower(server.InstallationPath)

	fmt.Printf("      Checking process PID %d:\n", proc.PID)
	fmt.Printf("        Process: %s\n", proc.Name)
	fmt.Printf("        CmdLine: %s\n", cmdLine)
	fmt.Printf("        Server cmd: %s, args: %v\n", serverCmd, server.Configuration.CommandLineArguments)

	// Method 1: Match by command + args pattern
	// For extension servers, match the command and key arguments
	if server.Source == models.DiscoveryExtension {
		// Extension servers have specific patterns

		// Check if process name matches the server command
		procName := strings.ToLower(proc.Name)
		commandMatch := strings.Contains(procName, serverCmd) || strings.HasPrefix(serverCmd, procName)

		if !commandMatch {
			fmt.Printf("        ✗ Process name doesn't match server command\n")
			return false
		}

		fmt.Printf("        ✓ Command name matches\n")

		// For node servers, match the entry point script exactly
		if serverCmd == "node" || strings.Contains(procName, "node") {
			// Get the entry point (first argument should be the server's main script)
			if len(server.Configuration.CommandLineArguments) == 0 {
				fmt.Printf("        ✗ No arguments configured for node server\n")
				return false
			}

			entryPoint := server.Configuration.CommandLineArguments[0]
			entryPointLower := strings.ToLower(strings.ReplaceAll(entryPoint, "\\", "/"))
			cmdLinePath := strings.ReplaceAll(cmdLine, "\\", "/")

			// The entry point script must appear in the command line
			if !strings.Contains(cmdLinePath, entryPointLower) {
				fmt.Printf("        ✗ Entry point script not found: %s\n", entryPoint)
				return false
			}

			fmt.Printf("        ✓ Entry point script match: %s\n", entryPoint)
			return true
		}

		// For uv/python servers, check for the extension directory path
		if serverCmd == "uv" || serverCmd == "python" || serverCmd == "python3" {
			// Get extension path from environment variables
			extPath, ok := server.Configuration.EnvironmentVariables["__EXTENSION_PATH__"]
			if !ok {
				fmt.Printf("        ✗ No extension path found in environment\n")
				return false
			}

			extPathLower := strings.ToLower(strings.ReplaceAll(extPath, "\\", "/"))
			cmdLinePath := strings.ReplaceAll(cmdLine, "\\", "/")

			// Check if command line contains the extension path
			if !strings.Contains(cmdLinePath, extPathLower) {
				fmt.Printf("        ✗ Extension path not found: %s\n", extPath)
				return false
			}

			fmt.Printf("        ✓ Extension path match: %s\n", extPath)
			return true
		}

		// Unknown command type for extension
		fmt.Printf("        ✗ Unknown extension command type: %s\n", serverCmd)
		return false
	}

	// Method 2: Direct command path match
	if cmdLine == serverCmd {
		fmt.Printf("        ✓ Direct command match\n")
		return true
	}

	// Method 3: Match by server name in command line
	serverNameLower := strings.ToLower(server.Name)
	if strings.Contains(cmdLine, serverNameLower) {
		fmt.Printf("        ✓ Server name match\n")
		return true
	}

	fmt.Printf("        ✗ No match\n")
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

// Close stops the file watcher and cleans up resources (FR-050)
func (ds *DiscoveryService) Close() error {
	if ds.configFileWatcher != nil {
		return ds.configFileWatcher.Stop()
	}
	return nil
}
