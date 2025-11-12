## Phase C: Core Services (Tasks C001-C020)

**Objective**: Implement business logic services (discovery, lifecycle, config, monitoring, storage)

### T-C001 Implement event bus (pub/sub)
**File**: `internal/core/events/eventbus.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001, T-B007

**Description**:
Implement lightweight pub/sub event bus per research.md §6.

**Steps**:
1. Define event types:
   - `ServerDiscovered`, `ServerStatusChanged`, `ServerLogEntry`, `ConfigFileChanged`, `ServerMetricsUpdated`
2. Create `EventBus` struct:
   ```go
   type EventBus struct {
       subscribers map[string][]chan Event
       mu          sync.RWMutex
   }
   func (eb *EventBus) Subscribe(eventType string) <-chan Event
   func (eb *EventBus) Publish(event Event)
   func (eb *EventBus) Close()
   ```
3. Buffered channels (size 100) to prevent blocking
4. Implement subscriber cleanup on channel close

**Acceptance**:
- Subscribers receive published events
- Multiple subscribers to same event work
- No blocking on slow consumers (drop or buffer)
- Unit test: concurrent publish/subscribe

---

### T-C002 Implement storage service with atomic writes
**File**: `internal/storage/storage.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001, T-B006

**Description**:
Implement JSON file persistence with atomic writes per research.md §5.

**Steps**:
1. Create `StorageService` interface:
   ```go
   type StorageService interface {
       LoadState() (*models.ApplicationState, error)
       SaveState(state *models.ApplicationState) error
       LoadServerLogs(serverID string) ([]models.LogEntry, error)
       SaveServerLogs(serverID string, logs []models.LogEntry) error
   }
   ```
2. Implement `FileStorage`:
   - SaveState: Write to `~/.mcpmanager/state.json.tmp`, then `os.Rename` to `state.json`
   - Create backup: Copy `state.json` to `state.json.backup` before write
   - LoadState: Read `state.json`, unmarshal to ApplicationState
3. Auto-create `~/.mcpmanager/` directory on first write

**Acceptance**:
- State persists across restarts
- Atomic writes prevent corruption (interrupt during write)
- Backup file created on each save
- Unit test: Load after Save returns same data

---

### T-C003 Implement debounced state auto-save
**File**: `internal/storage/autosave.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-C002

**Description**:
Add auto-save with debouncing (max 1 write/sec) per research.md §5.

**Steps**:
1. Create `AutoSaver`:
   ```go
   type AutoSaver struct {
       storage    StorageService
       state      *models.ApplicationState
       mu         sync.RWMutex
       dirty      bool
       stopChan   chan struct{}
   }
   func (as *AutoSaver) MarkDirty()
   func (as *AutoSaver) Start() // goroutine that saves if dirty every 1s
   func (as *AutoSaver) Stop()
   ```
2. Start background goroutine that checks `dirty` flag every 1 second
3. If dirty, call `storage.SaveState`, reset dirty flag

**Acceptance**:
- Multiple rapid MarkDirty() calls result in single write after 1s
- Stop() flushes pending changes immediately
- No writes if state hasn't changed
- Unit test: 10 MarkDirty() calls → 1 SaveState call

---

### T-C004 Implement server discovery service - client config parsing
**File**: `internal/core/discovery/clientconfig.go`
**Effort**: L (2+ hours)
**Dependencies**: T-A006, T-B001, T-C001

**Description**:
Implement client config file discovery per research.md §3 (Claude Desktop, Cursor, etc.).

**Steps**:
1. Create `ClientConfigDiscovery`:
   ```go
   type ClientConfigDiscovery struct {
       pathResolver platform.PathResolver
       eventBus     *events.EventBus
   }
   func (ccd *ClientConfigDiscovery) DiscoverFromClientConfigs() ([]models.MCPServer, error)
   ```
2. Scan paths:
   - `{ConfigDir}/Claude/claude_desktop_config.json`
   - `{ConfigDir}/Cursor/mcp_config.json`
   - Parse JSON: `{"mcpServers": {"name": {"command": "...", "args": [...]}}}`
3. For each server in config:
   - Create `MCPServer` entity with Source=client_config
   - Set status=stopped initially
   - Publish `ServerDiscovered` event
4. Handle missing/malformed JSON gracefully (log warning, continue)

**Acceptance**:
- Parses valid Claude Desktop config
- Parses valid Cursor config
- Handles missing files (no error, empty result)
- Handles malformed JSON (log error, skip file)
- Unit test with mock config files

---

### T-C005 Implement server discovery service - filesystem scanning
**File**: `internal/core/discovery/filesystem.go`
**Effort**: L (2+ hours)
**Dependencies**: T-A006, T-B001, T-C001

**Description**:
Implement filesystem scanning for NPM, Python, Go packages per research.md §3.

**Steps**:
1. Create `FilesystemDiscovery`:
   ```go
   func (fd *FilesystemDiscovery) DiscoverFromFilesystem() ([]models.MCPServer, error)
   ```
2. Scan NPM global:
   - Run `npm root -g` to get path
   - Search for packages matching `*mcp-server*`, `*mcp*` patterns
3. Scan Python site-packages:
   - Run `python -m site --user-site` to get path
   - Search for MCP server patterns
4. Scan Go binaries:
   - Check `$GOPATH/bin` or `~/go/bin`
   - Match binary names containing "mcp"
5. For each found:
   - Create MCPServer with Source=filesystem
   - Set InstallationPath to discovered path
   - Publish ServerDiscovered event

**Acceptance**:
- Discovers installed NPM MCP servers
- Discovers Python MCP servers
- Discovers Go MCP binaries
- Handles missing package managers gracefully
- Unit test with mock filesystem

---

### T-C006 Implement server discovery service - process matching
**File**: `internal/core/discovery/processes.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A007, T-B001, T-C001

**Description**:
Discover already-running MCP servers by scanning processes per research.md §3.

**Steps**:
1. Create `ProcessDiscovery`:
   ```go
   func (pd *ProcessDiscovery) DiscoverFromProcesses() ([]models.MCPServer, error)
   ```
2. Platform-specific process listing:
   - Unix: Parse `ps aux` output
   - Windows: Use WMI or tasklist
3. Match patterns: Look for processes with "mcp" in command line
4. For each matched process:
   - Create MCPServer with Source=process
   - Set status=running, PID=<detected_pid>
   - Publish ServerDiscovered event

**Acceptance**:
- Detects running MCP server processes
- Captures PID correctly
- Works on Windows, macOS, Linux
- Unit test with mock process list

---

### T-C007 Implement discovery orchestrator
**File**: `internal/core/discovery/discovery.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-C004, T-C005, T-C006

**Description**:
Orchestrate all discovery sources and deduplicate results.

**Steps**:
1. Create `DiscoveryService`:
   ```go
   func (ds *DiscoveryService) Discover() ([]models.MCPServer, error)
   ```
2. Run discovery sources in order:
   - ClientConfigDiscovery (primary, highest priority)
   - FilesystemDiscovery (secondary)
   - ProcessDiscovery (runtime, for PID matching)
3. Deduplicate servers by name or installation path:
   - If same server found via multiple sources, prefer client_config source
   - If found via process, match to existing entry and update PID + status
4. Cache results in memory
5. Return merged server list

**Acceptance**:
- Deduplication works correctly
- Priority: client_config > filesystem > process
- Running processes matched to discovered servers (PID updated)
- Unit test with overlapping discoveries

---

### T-C008 Implement file watcher for client config changes
**File**: `internal/core/discovery/filewatcher.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A004, T-C001, T-C007

**Description**:
Monitor client config files for external changes per FR-050, publish ConfigFileChanged event.

**Steps**:
1. Create `ConfigFileWatcher`:
   ```go
   type ConfigFileWatcher struct {
       watcher  *fsnotify.Watcher
       eventBus *events.EventBus
       paths    []string
   }
   func (cfw *ConfigFileWatcher) Start()
   func (cfw *ConfigFileWatcher) Stop()
   ```
2. Watch paths:
   - Claude Desktop config
   - Cursor config
   - Other client configs
3. On fsnotify event (Write or Create):
   - Publish `ConfigFileChanged` event with filePath
   - Do NOT auto-trigger discovery (per clarification: notify user, let them decide)
4. Handle watch errors (file deleted, permissions changed)

**Acceptance**:
- Detects external file modifications
- ConfigFileChanged event published
- No auto-refresh (manual trigger only)
- Unit test: modify watched file → event received

---

### T-C009 Implement lifecycle service - start server
**File**: `internal/core/lifecycle/start.go`
**Effort**: L (2+ hours)
**Dependencies**: T-A007, T-B001, T-B002, T-C001

**Description**:
Implement server start logic per FR-007, FR-052 (stopped → starting → running/error).

**Steps**:
1. Create `LifecycleService`:
   ```go
   func (ls *LifecycleService) StartServer(server *models.MCPServer) error
   ```
2. Validate: server.Status.State must be stopped or error
3. Transition to "starting" state, publish ServerStatusChanged event
4. Use ProcessManager to start server:
   - Command from server.Configuration or client config
   - Pass environment variables
   - Capture PID
5. Monitor process in goroutine:
   - If process exits within 5s → transition to error
   - If process running after 5s → transition to running
6. Capture stdout/stderr to log buffer

**Acceptance**:
- Server transitions through starting → running
- PID captured and stored
- Failure detection works (process exits early → error state)
- Unit test with mock ProcessManager

---

### T-C010 Implement lifecycle service - stop server
**File**: `internal/core/lifecycle/stop.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A007, T-B001, T-B002, T-C001

**Description**:
Implement server stop logic per FR-008 (graceful shutdown with timeout).

**Steps**:
1. Add to `LifecycleService`:
   ```go
   func (ls *LifecycleService) StopServer(server *models.MCPServer, force bool, timeout int) error
   ```
2. Validate: server.Status.State must be running
3. Use ProcessManager.Stop(pid, graceful=!force, timeout)
4. Wait for process exit (monitor goroutine completes)
5. Transition to stopped state, clear PID
6. Publish ServerStatusChanged event

**Acceptance**:
- Graceful stop sends SIGTERM (Unix) or WM_CLOSE (Windows)
- Force stop sends SIGKILL or taskkill /F
- Timeout honored (force kill after timeout)
- Status transitions to stopped
- Unit test with mock ProcessManager

---

### T-C011 Implement lifecycle service - restart server
**File**: `internal/core/lifecycle/restart.go`
**Effort**: S (<1 hour)
**Dependencies**: T-C009, T-C010

**Description**:
Implement restart as stop + start per FR-009.

**Steps**:
1. Add to `LifecycleService`:
   ```go
   func (ls *LifecycleService) RestartServer(server *models.MCPServer) error
   ```
2. Call StopServer(server, force=false, timeout=10)
3. Wait for stopped state
4. Call StartServer(server)
5. Publish ServerStatusChanged events for each transition

**Acceptance**:
- Restart completes successfully
- Server goes through: running → stopped → starting → running
- Events published for each state change
- Unit test verifies stop then start called

---

### T-C012 Implement lifecycle service - process monitoring
**File**: `internal/core/lifecycle/monitor.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A007, T-B001, T-C001

**Description**:
Background goroutine to monitor running server processes for crashes (edge case #2).

**Steps**:
1. Create `ProcessMonitor`:
   ```go
   type ProcessMonitor struct {
       processManager platform.ProcessManager
       eventBus       *events.EventBus
       servers        map[string]*models.MCPServer // serverID -> server
       stopChan       chan struct{}
   }
   func (pm *ProcessMonitor) Start()
   func (pm *ProcessMonitor) Stop()
   ```
2. Every 5 seconds, check all running servers:
   - Call ProcessManager.IsRunning(pid)
   - If no longer running → transition to error state
   - Publish ServerStatusChanged event with crash details
3. Goroutine-safe access to servers map

**Acceptance**:
- Detects crashed servers within 5 seconds
- Status transitions to error automatically
- Crash logs appear in log buffer
- Unit test: kill process externally → detected

---

### T-C013 Implement configuration service - CRUD operations
**File**: `internal/core/config/config.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-B003, T-C002

**Description**:
Implement configuration management per FR-014 through FR-018.

**Steps**:
1. Create `ConfigService`:
   ```go
   func (cs *ConfigService) GetConfiguration(serverID string) (*models.ServerConfiguration, error)
   func (cs *ConfigService) UpdateConfiguration(serverID string, config *models.ServerConfiguration) error
   func (cs *ConfigService) ValidateConfiguration(config *models.ServerConfiguration) error
   ```
2. GetConfiguration: Load from `~/.mcpmanager/servers/{serverID}/config.json`
3. UpdateConfiguration:
   - Validate configuration (FR-018)
   - Save to server-specific config file
   - Do NOT modify client config files (FR-019)
4. ValidateConfiguration:
   - Env var name regex
   - Path existence checks
   - MaxRestartAttempts range

**Acceptance**:
- Config persists per server
- Validation catches invalid inputs
- Client configs never modified
- Unit test: save/load round-trip works

---

### T-C014 Implement monitoring service - log capture
**File**: `internal/core/monitoring/logs.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-B004, T-C001

**Description**:
Capture server stdout/stderr to CircularLogBuffer per FR-020, FR-053.

**Steps**:
1. Create `MonitoringService`:
   ```go
   type MonitoringService struct {
       logBuffers map[string]*models.CircularLogBuffer // serverID -> buffer
       eventBus   *events.EventBus
   }
   func (ms *MonitoringService) CaptureOutput(serverID string, reader io.Reader)
   func (ms *MonitoringService) GetLogs(serverID string, offset, limit int) []models.LogEntry
   ```
2. CaptureOutput:
   - Read lines from stdout/stderr pipe
   - Parse severity from keywords ("error", "warn", "success")
   - Create LogEntry, add to buffer
   - Publish ServerLogEntry event for real-time UI
3. Initialize buffer per server (1000 entry capacity)

**Acceptance**:
- Logs captured from running server
- Severity auto-detected from keywords
- Circular buffer enforces 1000 entry limit
- Events published for UI updates
- Unit test with mock io.Reader

---

### T-C015 Implement monitoring service - metrics collection
**File**: `internal/core/monitoring/metrics.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A007, T-C001

**Description**:
Collect server metrics (uptime, memory, request count) per FR-025, FR-026.

**Steps**:
1. Add to `MonitoringService`:
   ```go
   func (ms *MonitoringService) GetMetrics(serverID string) (*ServerMetrics, error)
   ```
2. For running servers:
   - Uptime: Calculate from status.LastStateChange
   - Memory: Use platform-specific process memory query (Unix: /proc/{pid}/status, Windows: WMI)
   - Request count: Parse from MCP server if exposed, else null
3. Rate-limit updates to 1Hz per research.md §6

**Acceptance**:
- Uptime calculated correctly
- Memory usage retrieved for running servers
- Null values for stopped servers
- Unit test with mock process info

---

### T-C016 Implement monitoring service - log filtering
**File**: `internal/core/monitoring/filter.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-B004, T-C014

**Description**:
Implement log filtering by server, severity, and search query per FR-022, FR-023.

**Steps**:
1. Add to `MonitoringService`:
   ```go
   func (ms *MonitoringService) FilterLogs(serverID *string, severity *models.LogSeverity, search string, limit int) []models.LogEntry
   ```
2. Filter logic:
   - If serverID provided, get logs from that buffer only; else aggregate all buffers
   - If severity provided, filter by severity level
   - If search non-empty, filter by message substring match (case-insensitive)
   - Apply limit (default 100, max 1000)
3. Optimize: Use index/cache for frequently accessed filters

**Acceptance**:
- Filter by server ID works
- Filter by severity works
- Full-text search works (case-insensitive)
- Combined filters work (server + severity + search)
- Performance: <50ms for 50k entries

---

### T-C017 Implement dependency checking service
**File**: `internal/core/dependencies/check.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-B005

**Description**:
Check server dependencies per FR-027, FR-028.

**Steps**:
1. Create `DependencyService`:
   ```go
   func (ds *DependencyService) CheckDependencies(server *models.MCPServer) ([]models.Dependency, error)
   ```
2. For each dependency type:
   - runtime: Run `node --version`, `python --version`, `go version`
   - library: Check for shared library existence (platform-specific)
   - tool: Run `which npm`, `which pip`, etc.
   - environment: Check env var presence
3. Parse version from output, compare to requiredVersion
4. Set detectedVersion and installed flag
5. Provide actionable installationInstructions per platform

**Acceptance**:
- Detects Node.js version
- Detects Python version
- Detects missing dependencies
- Instructions are platform-appropriate (Windows/macOS/Linux)
- Unit test with mock command execution

---

### T-C018 Implement update checking service
**File**: `internal/core/dependencies/updates.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-B001

**Description**:
Check for server updates per FR-029.

**Steps**:
1. Add to `DependencyService`:
   ```go
   func (ds *DependencyService) CheckForUpdates(server *models.MCPServer) (*UpdateInfo, error)
   ```
2. For NPM servers:
   - Run `npm view {package} version` to get latest
   - Compare to server.Version
3. For Python servers:
   - Query PyPI API: `https://pypi.org/pypi/{package}/json`
4. For Go servers:
   - Query Go proxy: `https://proxy.golang.org/{module}/@latest`
5. Return UpdateInfo{updateAvailable, currentVersion, latestVersion, releaseNotes}

**Acceptance**:
- Detects newer versions for NPM packages
- Detects newer versions for Python packages
- Detects newer versions for Go modules
- Handles network errors gracefully (return "unknown" status)
- Unit test with mock HTTP responses

---

### T-C019 Write integration test for discovery flow
**File**: `tests/integration/discovery_test.go`
**Effort**: L (2+ hours)
**Dependencies**: T-C004 through T-C008

**Description**:
End-to-end test for server discovery from multiple sources.

**Test Steps**:
1. Create mock client config file with 2 servers
2. Create mock NPM global directory with 1 server package
3. Create mock running process (test script)
4. Run DiscoveryService.Discover()
5. Verify:
   - 3 servers discovered (2 from config, 1 from NPM)
   - Running process matched to existing server (PID updated)
   - Source field set correctly for each
   - ServerDiscovered events published

**Acceptance**:
- Test passes with all assertions
- Uses temporary directories (no pollution)
- Cleans up after completion

---

### T-C020 Write integration test for lifecycle flow
**File**: `tests/integration/lifecycle_test.go`
**Effort**: L (2+ hours)
**Dependencies**: T-C009 through T-C012

**Description**:
End-to-end test for server start → monitor → stop flow.

**Test Steps**:
1. Create test server: Simple Go HTTP server that responds to /ping
2. Create MCPServer model entry with test server command
3. Call LifecycleService.StartServer()
4. Verify:
   - Status transitions: stopped → starting → running
   - PID captured
   - Process running (HTTP /ping responds)
5. Call LifecycleService.StopServer()
6. Verify:
   - Status transitions: running → stopped
   - PID cleared
   - Process no longer running

**Acceptance**:
- Test passes with all assertions
- State transitions correct
- Process actually starts/stops
- Events published for each state change

---

