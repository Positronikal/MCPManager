# Implementation Tasks: MCP Manager

**Feature**: MCP Manager - Cross-Platform Server Management Application
**Branch**: `001-mcp-manager-specification`
**Date**: 2025-10-15
**Source Documents**: plan.md, research.md, data-model.md, contracts/api-spec.yaml, quickstart.md

---

## Task Execution Overview

This document contains **98 implementation tasks** organized in **6 phases** (A-F) following Test-Driven Development principles. Tasks marked with **[P]** can be executed in parallel as they operate on independent files or modules.

###

 Dependency Flow

```
Phase A (Foundation) → Phase B (Models) → Phase C (Services) → Phase D (API) → Phase E (Frontend) → Phase F (Testing)
     [P] Tasks              [P] Tasks         Sequential          [P] Tasks      Sequential         [P] Tasks
```

### Parallel Execution Example

Tasks within the same phase marked [P] can run concurrently:

```bash
# Example: Run multiple Phase A tasks in parallel
Task A001 & Task A002 & Task A003 & wait
```

---

## Phase A: Foundation & Setup (Tasks A001-A008)

**Objective**: Establish project structure, dependencies, and platform abstractions

### T-A001 [P] Initialize Go module and project structure
**File**: `go.mod`, project directories
**Effort**: M (1-2 hours)
**Dependencies**: None

**Description**:
Initialize Go module for MCP Manager and create the complete directory structure per plan.md.

**Steps**:
1. Run `go mod init github.com/your-org/mcpmanager` from repo root
2. Create directories:
   - `cmd/mcpmanager/`
   - `internal/api/`, `internal/core/{discovery,lifecycle,config,monitoring}/`
   - `internal/models/`, `internal/storage/`, `internal/platform/`
   - `frontend/src/{components,stores,services,routes}/`
   - `tests/{contract,integration,unit,frontend}/`
   - `pkg/mcpclient/`
3. Create `.gitignore` for Go + Node.js + Wails

**Acceptance**:
- `go.mod` exists with module path
- All directories created and committed
- `go mod tidy` runs without errors

---

### T-A002 [P] Setup local MCP SDK module reference
**File**: `go.mod`
**Effort**: S (<1 hour)
**Dependencies**: T-A001

**Description**:
Configure go.mod to use local MCP Go SDK per research.md decision.

**Steps**:
1. Add replace directive to `go.mod`:
   ```go
   replace github.com/modelcontextprotocol/go-sdk => ../../../_MCP-Tools-Dev/go-sdk
   ```
2. Add requirement: `require github.com/modelcontextprotocol/go-sdk v0.0.0`
3. Run `go mod tidy`

**Acceptance**:
- go.mod contains replace directive
- SDK imports resolve without errors
- No network fetch attempts for MCP SDK

---

### T-A003 [P] Initialize Wails project with Svelte frontend
**File**: `wails.json`, `frontend/` structure
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Initialize Wails v2 project with Svelte 4.x frontend per research.md.

**Steps**:
1. Run: `wails init -n mcpmanager -t svelte` (use existing structure, merge if needed)
2. Configure `wails.json`:
   - Set `outputfilename: "mcpmanager"`
   - Enable `devServer` for frontend hot reload
   - Configure `frontend:build` command: `npm run build`
3. Initialize frontend dependencies: `cd frontend && npm install`
4. Verify Wails CLI: `wails doctor`

**Acceptance**:
- `wails.json` configured correctly
- `frontend/package.json` contains Svelte 4.x dependencies
- `wails dev` starts without errors (Ctrl+C to stop)
- Frontend displays default Wails + Svelte template

---

### T-A004 [P] Add fsnotify dependency for file watching
**File**: `go.mod`
**Effort**: S (<1 hour)
**Dependencies**: T-A001

**Description**:
Add fsnotify library for monitoring client config file changes (FR-050).

**Steps**:
1. Run: `go get github.com/fsnotify/fsnotify`
2. Verify installation: `go mod tidy`
3. Create placeholder: `internal/platform/filewatcher.go` with interface:
   ```go
   type FileWatcher interface {
       Watch(path string) error
       Events() <-chan Event
       Close() error
   }
   ```

**Acceptance**:
- fsnotify appears in `go.mod`
- FileWatcher interface defined
- No build errors

---

### T-A005 [P] Implement platform abstraction interfaces
**File**: `internal/platform/interfaces.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Define platform-specific abstraction interfaces per research.md §8.

**Steps**:
1. Create `internal/platform/interfaces.go`:
   ```go
   type PathResolver interface {
       GetConfigDir() string
       GetAppDataDir() string
       GetUserHomeDir() string
   }

   type ProcessManager interface {
       Start(cmd string, args []string, env map[string]string) (pid int, err error)
       Stop(pid int, graceful bool, timeout int) error
       IsRunning(pid int) bool
   }

   type SingleInstance interface {
       Acquire() (bool, error)
       Release() error
   }
   ```

**Acceptance**:
- All interfaces defined in `internal/platform/interfaces.go`
- Interfaces exported (capitalized)
- Go build succeeds

---

### T-A006 [P] Implement platform-specific path resolvers
**Files**: `internal/platform/paths_windows.go`, `paths_darwin.go`, `paths_linux.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A005

**Description**:
Implement PathResolver for Windows, macOS, and Linux using build tags.

**Steps**:
1. Create `paths_windows.go` with `//go:build windows`:
   - GetConfigDir → `%APPDATA%`
   - GetAppDataDir → `%LOCALAPPDATA%`
2. Create `paths_darwin.go` with `//go:build darwin`:
   - GetConfigDir → `~/Library/Application Support`
   - GetAppDataDir → `~/Library/Application Support`
3. Create `paths_linux.go` with `//go:build linux`:
   - GetConfigDir → `~/.config`
   - GetAppDataDir → `~/.local/share`
4. Add helper for `~/.mcpmanager/` expansion

**Acceptance**:
- Paths resolve correctly on respective platforms
- Unit test verifies path format per OS (use `runtime.GOOS`)
- No compilation errors on target platforms

---

### T-A007 [P] Implement platform-specific process managers
**Files**: `internal/platform/process_windows.go`, `process_unix.go`
**Effort**: L (2+ hours)
**Dependencies**: T-A005

**Description**:
Implement ProcessManager for Windows (taskkill) and Unix (SIGTERM/SIGKILL).

**Steps**:
1. Create `process_windows.go` (`//go:build windows`):
   - Start: Use `exec.Command`, capture PID
   - Stop: `taskkill /PID {pid} /F` if !graceful, else `/T`
   - IsRunning: Check process existence via WMI or tasklist
2. Create `process_unix.go` (`//go:build darwin || linux`):
   - Start: Use `exec.Command` with `SysProcAttr`
   - Stop: `syscall.Kill(pid, syscall.SIGTERM)` then SIGKILL after timeout
   - IsRunning: `syscall.Kill(pid, 0)` → errno check
3. Handle stdout/stderr capture via `io.Pipe`

**Acceptance**:
- Can start/stop test processes on each platform
- PID tracking works correctly
- Graceful shutdown honors timeout parameter
- Unit tests pass (mock process for testing)

---

### T-A008 [P] Implement single-instance enforcement
**Files**: `internal/platform/singleton_windows.go`, `singleton_unix.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A005

**Description**:
Implement SingleInstance mechanism per FR-051 (bring existing window to foreground).

**Steps**:
1. Windows (`singleton_windows.go`):
   - Use named mutex: `CreateMutex("Global\\MCPManager")`
   - On acquire failure, find window handle, call `SetForegroundWindow`
2. Unix (`singleton_unix.go`):
   - Use Unix socket lock: `/tmp/mcpmanager.lock`
   - On acquire failure, read existing PID, signal to show window

**Acceptance**:
- Second launch brings first window to foreground
- No duplicate processes created (verified via OS task manager)
- Lock released on application exit

---

## Phase B: Domain Models (Tasks B001-B012)

**Objective**: Implement data models with validation per data-model.md

### T-B001 [P] Implement MCPServer model
**File**: `internal/models/server.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Implement MCPServer entity with all attributes from data-model.md §1.

**Steps**:
1. Create `internal/models/server.go`:
   ```go
   type MCPServer struct {
       ID               string              `json:"id"`
       Name             string              `json:"name"`
       Version          string              `json:"version,omitempty"`
       InstallationPath string              `json:"installationPath"`
       Status           ServerStatus        `json:"status"`
       PID              *int                `json:"pid,omitempty"`
       Capabilities     []string            `json:"capabilities,omitempty"`
       Tools            []string            `json:"tools,omitempty"`
       Configuration    ServerConfiguration `json:"configuration"`
       Dependencies     []Dependency        `json:"dependencies,omitempty"`
       DiscoveredAt     time.Time           `json:"discoveredAt"`
       LastSeenAt       time.Time           `json:"lastSeenAt"`
       Source           DiscoverySource     `json:"source"`
   }
   ```
2. Add validation method: `Validate() error`
   - Name unique check (deferred to service layer)
   - PID consistency with status
   - InstallationPath existence check
   - LastSeenAt >= DiscoveredAt

**Acceptance**:
- Struct compiles without errors
- JSON marshaling/unmarshaling works
- Validate() enforces rules from data-model.md

---

### T-B002 [P] Implement ServerStatus model with state machine
**File**: `internal/models/status.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Implement ServerStatus entity with state transition validation per data-model.md §2.

**Steps**:
1. Define `StatusState` enum: `stopped`, `starting`, `running`, `error`
2. Create `ServerStatus` struct with fields from data-model.md
3. Implement `CanTransitionTo(newState StatusState) bool` using state machine diagram
4. Implement `TransitionTo(newState StatusState, reason string) error`
   - Validate transition allowed
   - Update `lastStateChange` timestamp
   - Handle `startupAttempts` increment/reset logic

**Acceptance**:
- Invalid transitions rejected (e.g., stopped → running without starting)
- StartupAttempts increments on starting → error, resets on starting → running
- State machine diagram rules enforced

---

### T-B003 [P] Implement ServerConfiguration model
**File**: `internal/models/configuration.go`
**Effort**: S (<1 hour)
**Dependencies**: T-A001

**Description**:
Implement ServerConfiguration per data-model.md §3.

**Steps**:
1. Create struct with fields:
   - EnvironmentVariables (map[string]string)
   - CommandLineArguments ([]string)
   - WorkingDirectory, AutoStart, RestartOnCrash, MaxRestartAttempts, etc.
2. Add validation:
   - Env var keys match regex `^[A-Z_][A-Z0-9_]*$`
   - WorkingDirectory exists if provided
   - MaxRestartAttempts 0-10

**Acceptance**:
- Struct compiles and marshals to JSON
- Validation catches invalid env var names
- Path validation works

---

### T-B004 [P] Implement LogEntry model with circular buffer
**File**: `internal/models/log.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Implement LogEntry model and CircularLogBuffer per data-model.md §4.

**Steps**:
1. Define `LogSeverity` enum: info, success, warning, error
2. Create `LogEntry` struct (ID, timestamp, severity, source, message, metadata)
3. Implement `CircularLogBuffer`:
   ```go
   type CircularLogBuffer struct {
       entries [1000]LogEntry
       head    int
       size    int
       mu      sync.RWMutex
   }
   func (b *CircularLogBuffer) Add(entry LogEntry)
   func (b *CircularLogBuffer) Get(offset, limit int) []LogEntry
   func (b *CircularLogBuffer) Filter(severity LogSeverity) []LogEntry
   ```

**Acceptance**:
- Buffer stores max 1000 entries, overwrites oldest
- Thread-safe (concurrent Add/Get)
- Get() returns entries in chronological order
- Filter() works by severity

---

### T-B005 [P] Implement Dependency model
**File**: `internal/models/dependency.go`
**Effort**: S (<1 hour)
**Dependencies**: T-A001

**Description**:
Implement Dependency entity per data-model.md §5.

**Steps**:
1. Define `DependencyType` enum: runtime, library, tool, environment
2. Create `Dependency` struct with fields
3. Add `IsInstalled() bool` computed method:
   - Returns true if detectedVersion satisfies requiredVersion
   - Use semver comparison library if needed

**Acceptance**:
- IsInstalled() correctly evaluates version constraints
- Struct marshals to JSON
- InstallationInstructions field accepts markdown

---

### T-B006 [P] Implement ApplicationState model
**File**: `internal/models/appstate.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Implement ApplicationState with sub-entities per data-model.md §6.

**Steps**:
1. Create `ApplicationState` struct with top-level fields
2. Define sub-structs:
   - `UserPreferences` (theme, logRetentionPerServer, etc.)
   - `WindowLayout` (width, height, x, y, maximized, logPanelHeight)
   - `Filters` (selectedServer, selectedSeverity, searchQuery)
3. Add validation:
   - WindowLayout min 640x480
   - DiscoveredServers are valid UUIDs
   - MonitoredConfigPaths are absolute

**Acceptance**:
- All sub-entities defined
- Validation enforces constraints
- JSON marshaling includes nested objects

---

### T-B007 [P] Implement enumerations
**File**: `internal/models/enums.go`
**Effort**: S (<1 hour)
**Dependencies**: T-A001

**Description**:
Define all enumerations from data-model.md in a single file.

**Steps**:
1. Define string constants:
   - `StatusState`: stopped, starting, running, error
   - `LogSeverity`: info, success, warning, error
   - `DependencyType`: runtime, library, tool, environment
   - `DiscoverySource`: client_config, filesystem, process
2. Add validation functions for each enum type

**Acceptance**:
- All enums defined as Go constants
- Validation functions reject invalid values
- Enums used consistently across models

---

### T-B008 [P] Add JSON marshaling with custom timestamp handling
**Files**: Modify all model files
**Effort**: M (1-2 hours)
**Dependencies**: T-B001 through T-B007

**Description**:
Implement custom JSON marshaling for timestamps (ISO 8601 per data-model.md).

**Steps**:
1. For all models with `time.Time` fields:
   - Ensure JSON tag includes RFC3339 format
   - Test marshaling: `json.Marshal(model)` produces ISO 8601 strings
2. Add `MarshalJSON()` and `UnmarshalJSON()` if needed for complex types

**Acceptance**:
- All timestamps serialize to ISO 8601 format
- Deserialization handles ISO 8601 strings
- Round-trip (marshal → unmarshal) preserves data

---

### T-B009 [P] Write unit tests for MCPServer validation
**File**: `internal/models/server_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-B001

**Description**:
Write comprehensive unit tests for MCPServer validation rules.

**Test Cases**:
1. Valid server with all required fields
2. Invalid: empty name
3. Invalid: PID set when status is stopped
4. Invalid: PID null when status is running
5. Invalid: lastSeenAt < discoveredAt
6. Edge case: installation path doesn't exist

**Acceptance**:
- All test cases pass
- Coverage > 80% for server.go
- Tests use table-driven approach

---

### T-B010 [P] Write unit tests for ServerStatus state machine
**File**: `internal/models/status_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-B002

**Description**:
Test all valid and invalid state transitions per data-model.md state machine diagram.

**Test Cases**:
1. Valid: stopped → starting → running
2. Valid: stopped → starting → error
3. Valid: running → stopped
4. Valid: running → error (crash)
5. Valid: error → starting (retry)
6. Invalid: stopped → running (must go through starting)
7. Invalid: error → running (must go through starting)
8. StartupAttempts increment/reset logic

**Acceptance**:
- State machine enforces all diagram rules
- Startup attempts tracked correctly
- Coverage > 90% for status.go

---

### T-B011 [P] Write unit tests for CircularLogBuffer
**File**: `internal/models/log_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-B004

**Description**:
Test circular buffer behavior, thread safety, and capacity limits.

**Test Cases**:
1. Add 1000 entries, verify all stored
2. Add 1001 entries, verify oldest discarded
3. Concurrent Add from multiple goroutines (race detector)
4. Concurrent Add + Get (no deadlocks)
5. Filter by severity returns correct subset
6. Get with offset/limit paginates correctly

**Acceptance**:
- Buffer never exceeds 1000 entries
- `go test -race` passes (no data races)
- Coverage > 85% for log.go

---

### T-B012 [P] Write unit tests for ApplicationState validation
**File**: `internal/models/appstate_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-B006

**Description**:
Test ApplicationState validation rules.

**Test Cases**:
1. Valid state with all fields
2. Invalid: WindowLayout width < 640
3. Invalid: WindowLayout height < 480
4. Invalid: DiscoveredServers contains non-UUID string
5. Invalid: MonitoredConfigPaths contains relative path
6. Default values set correctly (theme=dark, logRetentionPerServer=1000)

**Acceptance**:
- All validation rules enforced
- Defaults apply on struct initialization
- Coverage > 80% for appstate.go

---

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

## Phase D: API Layer (Tasks D001-D018)

**Objective**: Implement REST API handlers + SSE stream per contracts/api-spec.yaml

### T-D001 [P] Write contract test for GET /servers
**File**: `tests/contract/servers_list_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Write failing contract test for GET /api/v1/servers endpoint per api-spec.yaml.

**Test Steps**:
1. Create test HTTP server (no implementation yet, should fail)
2. Send GET /api/v1/servers
3. Assert:
   - Status 200
   - Response schema: `{"servers": [...], "count": int, "lastDiscovery": "ISO 8601"}`
   - Query params: ?status=running, ?source=client_config work

**Acceptance**:
- Test fails initially (no implementation)
- Schema validation correct per OpenAPI spec
- Test uses Go httptest package

---

### T-D002 [P] Write contract test for POST /servers/discover
**File**: `tests/contract/servers_discover_test.go`
**Effort**: S (<1 hour)
**Dependencies**: T-A001

**Description**:
Write failing contract test for POST /api/v1/servers/discover endpoint.

**Test Steps**:
1. Send POST /api/v1/servers/discover
2. Assert:
   - Status 202 Accepted
   - Response: `{"message": "Discovery scan initiated", "scanId": "uuid"}`

**Acceptance**:
- Test fails initially
- Schema matches api-spec.yaml
- UUIDs validated

---

### T-D003 [P] Write contract test for GET /servers/{serverId}
**File**: `tests/contract/servers_get_test.go`
**Effort**: S (<1 hour)
**Dependencies**: T-A001

**Description**:
Write failing contract test for GET /api/v1/servers/{serverId} endpoint.

**Test Steps**:
1. Send GET /api/v1/servers/{valid-uuid}
2. Assert:
   - Status 200
   - Response: MCPServer schema with all fields
3. Send GET /api/v1/servers/{invalid-uuid}
4. Assert:
   - Status 404
   - Error response schema

**Acceptance**:
- Test fails initially
- 200 and 404 cases covered
- UUID validation works

---

### T-D004 [P] Write contract tests for lifecycle endpoints
**File**: `tests/contract/lifecycle_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Write failing contract tests for start/stop/restart/status endpoints.

**Endpoints**:
- POST /api/v1/servers/{serverId}/start → 202
- POST /api/v1/servers/{serverId}/stop → 202 (request body: force, timeout)
- POST /api/v1/servers/{serverId}/restart → 202
- GET /api/v1/servers/{serverId}/status → 200, ServerStatus schema

**Acceptance**:
- All 4 endpoints have contract tests
- Request/response schemas validated
- Tests fail initially

---

### T-D005 [P] Write contract tests for configuration endpoints
**File**: `tests/contract/config_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Write failing contract tests for configuration GET/PUT endpoints.

**Endpoints**:
- GET /api/v1/servers/{serverId}/configuration → 200, ServerConfiguration schema
- PUT /api/v1/servers/{serverId}/configuration → 200 (body: ServerConfiguration), 400 (validation error)

**Acceptance**:
- GET and PUT tests written
- Validation error case (400) covered
- Schema matches data-model.md

---

### T-D006 [P] Write contract tests for monitoring endpoints
**File**: `tests/contract/monitoring_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Write failing contract tests for logs and metrics endpoints.

**Endpoints**:
- GET /api/v1/servers/{serverId}/logs → 200, {logs: [LogEntry], total: int, hasMore: bool}
- GET /api/v1/logs?serverId=x&severity=error&search=query → 200, filtered logs
- GET /api/v1/servers/{serverId}/metrics → 200, {uptimeSeconds, memoryUsageMB, requestCount, cpuPercent}

**Acceptance**:
- All 3 endpoints tested
- Query parameters validated
- Schemas match api-spec.yaml

---

### T-D007 [P] Write contract tests for dependency endpoints
**File**: `tests/contract/dependencies_test.go`
**Effort**: S (<1 hour)
**Dependencies**: T-A001

**Description**:
Write failing contract tests for dependencies and updates endpoints.

**Endpoints**:
- GET /api/v1/servers/{serverId}/dependencies → 200, {dependencies: [Dependency], allSatisfied: bool}
- GET /api/v1/servers/{serverId}/updates → 200, {updateAvailable: bool, currentVersion, latestVersion, releaseNotes}

**Acceptance**:
- Both endpoints tested
- Schemas validated
- Tests fail initially

---

### T-D008 [P] Write contract tests for application state endpoints
**File**: `tests/contract/appstate_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Write failing contract tests for application state GET/PUT endpoints.

**Endpoints**:
- GET /api/v1/application/state → 200, ApplicationState schema
- PUT /api/v1/application/state → 200 (body: ApplicationState), 400 (validation error)

**Acceptance**:
- GET and PUT tested
- Nested objects (userPreferences, windowLayout, filters) validated
- Tests fail initially

---

### T-D009 Implement API handlers for discovery endpoints
**File**: `internal/api/discovery.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-C007, T-D001, T-D002, T-D003

**Description**:
Implement handlers for GET /servers, POST /servers/discover, GET /servers/{serverId}.

**Steps**:
1. Create HTTP handlers using Chi router or similar
2. GET /servers:
   - Call DiscoveryService.GetServers(statusFilter, sourceFilter)
   - Return JSON: {servers, count, lastDiscovery}
3. POST /servers/discover:
   - Trigger discovery in goroutine (non-blocking)
   - Return 202 with scanId
4. GET /servers/{serverId}:
   - Lookup server by ID
   - Return 404 if not found, 200 with server JSON

**Acceptance**:
- Contract tests from T-D001, T-D002, T-D003 now pass
- Query parameters work
- Error handling correct (404 for missing server)

---

### T-D010 Implement API handlers for lifecycle endpoints
**File**: `internal/api/lifecycle.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-C009 through T-C011, T-D004

**Description**:
Implement handlers for start/stop/restart/status endpoints.

**Steps**:
1. POST /servers/{serverId}/start:
   - Call LifecycleService.StartServer()
   - Return 202 with {message, serverId, status: "starting"}
   - Return 400 if server already running
2. POST /servers/{serverId}/stop:
   - Parse request body: {force, timeout}
   - Call LifecycleService.StopServer()
   - Return 202
   - Return 400 if server not running
3. POST /servers/{serverId}/restart:
   - Call LifecycleService.RestartServer()
   - Return 202
4. GET /servers/{serverId}/status:
   - Return current ServerStatus
   - Return 404 if server not found

**Acceptance**:
- Contract tests from T-D004 now pass
- Async operations don't block (202 response immediate)
- Error handling correct (400 for invalid state)

---

### T-D011 Implement API handlers for configuration endpoints
**File**: `internal/api/config.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-C013, T-D005

**Description**:
Implement handlers for configuration GET/PUT.

**Steps**:
1. GET /servers/{serverId}/configuration:
   - Call ConfigService.GetConfiguration(serverID)
   - Return 200 with ServerConfiguration JSON
   - Return 404 if server not found
2. PUT /servers/{serverId}/configuration:
   - Parse request body: ServerConfiguration
   - Call ConfigService.ValidateConfiguration()
   - If valid: Call ConfigService.UpdateConfiguration()
   - Return 200 with updated config
   - Return 400 if validation fails with error details

**Acceptance**:
- Contract tests from T-D005 now pass
- Validation errors return 400 with descriptive message
- Configuration persists (verified via GET after PUT)

---

### T-D012 Implement API handlers for monitoring endpoints
**File**: `internal/api/monitoring.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-C014 through T-C016, T-D006

**Description**:
Implement handlers for logs and metrics endpoints.

**Steps**:
1. GET /servers/{serverId}/logs:
   - Parse query params: severity, limit, offset
   - Call MonitoringService.GetLogs(serverID, offset, limit)
   - Filter by severity if provided
   - Return {logs, total, hasMore}
2. GET /logs:
   - Parse query params: serverId, severity, search, limit
   - Call MonitoringService.FilterLogs()
   - Return {logs, total}
3. GET /servers/{serverId}/metrics:
   - Call MonitoringService.GetMetrics(serverID)
   - Return {uptimeSeconds, memoryUsageMB, requestCount, cpuPercent}
   - Return 404 if server not found

**Acceptance**:
- Contract tests from T-D006 now pass
- Filtering works correctly
- Pagination works (offset + limit)
- Metrics return null for stopped servers

---

### T-D013 Implement API handlers for dependency endpoints
**File**: `internal/api/dependencies.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-C017, T-C018, T-D007

**Description**:
Implement handlers for dependencies and updates endpoints.

**Steps**:
1. GET /servers/{serverId}/dependencies:
   - Call DependencyService.CheckDependencies(serverID)
   - Return {dependencies, allSatisfied}
   - Return 404 if server not found
2. GET /servers/{serverId}/updates:
   - Call DependencyService.CheckForUpdates(serverID)
   - Return {updateAvailable, currentVersion, latestVersion, releaseNotes}
   - Handle network errors gracefully (return "unknown" status)

**Acceptance**:
- Contract tests from T-D007 now pass
- Dependency checks work on all platforms
- Update checks handle network failures

---

### T-D014 Implement API handlers for application state endpoints
**File**: `internal/api/appstate.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-C002, T-D008

**Description**:
Implement handlers for application state GET/PUT.

**Steps**:
1. GET /application/state:
   - Call StorageService.LoadState()
   - Return ApplicationState JSON
   - If file doesn't exist, return defaults
2. PUT /application/state:
   - Parse request body: ApplicationState
   - Validate (window size minimums, UUID formats)
   - Call StorageService.SaveState()
   - Return 200 with {message: "Application state saved"}
   - Return 400 if validation fails

**Acceptance**:
- Contract tests from T-D008 now pass
- State persists across restarts
- Validation errors return 400
- Defaults returned on first load

---

### T-D015 Implement SSE (Server-Sent Events) stream handler
**File**: `internal/api/events.go`
**Effort**: L (2+ hours)
**Dependencies**: T-C001, T-A001

**Description**:
Implement GET /events endpoint for real-time SSE stream per api-spec.yaml.

**Steps**:
1. Create SSE handler:
   - Set headers: `Content-Type: text/event-stream`, `Cache-Control: no-cache`, `Connection: keep-alive`
   - Parse query params: serverIds (comma-separated)
   - Subscribe to EventBus for relevant event types
2. Event loop:
   - For each event received from EventBus:
     - Format as SSE: `id: {uuid}\nevent: {type}\ndata: {json}\n\n`
     - Write to response writer, flush immediately
   - Send heartbeat comment every 15s: `: heartbeat\n\n`
3. Handle client disconnect: Unsubscribe from EventBus, close channels
4. Implement reconnection support:
   - Read `Last-Event-ID` header
   - Resend missed events if available (buffer last 100 events)
   - If Last-Event-ID unknown, send full state snapshot

**Acceptance**:
- SSE stream established (curl -N http://localhost:8080/api/v1/events connects)
- Events formatted correctly per api-spec.yaml
- Heartbeats prevent connection timeout
- Reconnection with Last-Event-ID works
- Unit test: Publish event → received via SSE

---

### T-D016 Implement API middleware (logging, error handling, CORS)
**File**: `internal/api/middleware.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Add middleware for logging, error handling, and CORS.

**Steps**:
1. Request logging middleware:
   - Log: method, path, status, duration
   - Use Go's log/slog package
2. Error handling middleware:
   - Catch panics, return 500 Internal Server Error
   - Log stack trace
3. CORS middleware (for local development):
   - Allow origins: http://localhost:* (Wails dev server)
   - Allow methods: GET, POST, PUT, OPTIONS
   - Allow headers: Content-Type, Last-Event-ID

**Acceptance**:
- All API requests logged
- Panics caught and logged (no server crash)
- CORS headers present (verified with curl -H "Origin: http://localhost:34115")
- Integration test: Panic in handler → 500 response

---

### T-D017 Setup API router and server initialization
**File**: `internal/api/router.go`, `cmd/mcpmanager/main.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-D009 through T-D016

**Description**:
Wire up all API handlers with Chi router and start HTTP server.

**Steps**:
1. Create router in `internal/api/router.go`:
   ```go
   func NewRouter(services *Services) *chi.Mux {
       r := chi.NewRouter()
       r.Use(LoggingMiddleware, ErrorHandlingMiddleware, CORSMiddleware)

       r.Route("/api/v1", func(r chi.Router) {
           r.Get("/servers", handlers.ListServers)
           r.Post("/servers/discover", handlers.DiscoverServers)
           // ... all other endpoints
           r.Get("/events", handlers.SSEStream)
       })

       return r
   }
   ```
2. In `cmd/mcpmanager/main.go`:
   - Initialize all services (DiscoveryService, LifecycleService, etc.)
   - Create router
   - Start HTTP server on localhost:8080
   - Graceful shutdown on SIGINT/SIGTERM

**Acceptance**:
- HTTP server starts on port 8080
- All endpoints accessible
- Graceful shutdown works (Ctrl+C stops cleanly)
- All contract tests pass

---

### T-D018 Generate Wails Go-to-JavaScript bindings
**File**: `frontend/wailsjs/go/main/App.js` (generated)
**Effort**: M (1-2 hours)
**Dependencies**: T-D017, T-A003

**Description**:
Expose backend API methods to Svelte frontend via Wails bindings.

**Steps**:
1. Create `cmd/mcpmanager/app.go`:
   ```go
   type App struct {
       ctx       context.Context
       apiClient *api.Client
   }

   func (a *App) ListServers() ([]models.MCPServer, error)
   func (a *App) StartServer(serverID string) error
   func (a *App) StopServer(serverID string, force bool, timeout int) error
   // ... expose key operations
   ```
2. In `main.go`, pass `app` instance to Wails:
   ```go
   app := NewApp()
   wails.Run(&options.App{
       Bind: []interface{}{app},
       // ... other options
   })
   ```
3. Run `wails dev` → Wails generates bindings in `frontend/wailsjs/go/main/App.js`

**Acceptance**:
- Bindings generated in `frontend/wailsjs/`
- Frontend can import: `import { ListServers, StartServer } from '../wailsjs/go/main/App'`
- Calling bindings from Svelte works (returns Promise)
- TypeScript types available

---

## Phase E: Frontend Implementation (Tasks E001-E030)

**Objective**: Build Svelte UI components and integrate with backend API/SSE

### T-E001 Setup Svelte TypeScript and stores
**File**: `frontend/src/stores/stores.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-A003

**Description**:
Setup Svelte stores for state management and TypeScript configuration.

**Steps**:
1. Configure TypeScript in `frontend/tsconfig.json`:
   - Enable strict mode
   - Add type definitions for Svelte, Wails bindings
2. Create Svelte stores in `stores/stores.ts`:
   ```typescript
   import { writable } from 'svelte/store';

   export const servers = writable<MCPServer[]>([]);
   export const logs = writable<LogEntry[]>([]);
   export const appState = writable<ApplicationState>(defaultState);
   export const selectedServer = writable<string | null>(null);
   export const selectedSeverity = writable<LogSeverity | null>(null);
   ```
3. Define TypeScript interfaces matching Go models:
   - MCPServer, ServerStatus, LogEntry, ApplicationState, etc.

**Acceptance**:
- TypeScript compiles without errors
- Stores defined and exportable
- Type definitions match backend models
- `npm run check` passes (Svelte TypeScript check)

---

### T-E002 Implement API service client (REST wrapper)
**File**: `frontend/src/services/api.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001

**Description**:
Create TypeScript API client for calling backend REST endpoints.

**Steps**:
1. Create `api.ts`:
   ```typescript
   class APIClient {
       private baseURL = 'http://localhost:8080/api/v1';

       async listServers(statusFilter?: string): Promise<{servers: MCPServer[], count: number, lastDiscovery: string}> {
           const response = await fetch(`${this.baseURL}/servers?status=${statusFilter || ''}`);
           return response.json();
       }

       async startServer(serverId: string): Promise<void> {
           await fetch(`${this.baseURL}/servers/${serverId}/start`, { method: 'POST' });
       }

       // ... all other endpoints
   }

   export const apiClient = new APIClient();
   ```
2. Handle errors: Throw descriptive errors on non-200 responses
3. Add request/response logging for debugging

**Acceptance**:
- API client compiles and exports
- All endpoints from api-spec.yaml covered
- Error handling works (network errors, 4xx/5xx responses)
- Unit test (mock fetch): listServers() returns typed data

---

### T-E003 Implement SSE client with auto-reconnect
**File**: `frontend/src/services/sse.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001

**Description**:
Create SSE client for real-time events per api-spec.yaml reconnection strategy.

**Steps**:
1. Create `sse.ts`:
   ```typescript
   class SSEClient {
       private eventSource: EventSource | null = null;
       private lastEventId: string | null = null;
       private reconnectDelay = 1000; // Start at 1s, exponential backoff

       connect(onEvent: (event: Event) => void) {
           const url = `http://localhost:8080/api/v1/events`;
           this.eventSource = new EventSource(url);

           this.eventSource.onmessage = (e) => {
               this.lastEventId = e.lastEventId;
               this.reconnectDelay = 1000; // Reset backoff on success
               onEvent(JSON.parse(e.data));
           };

           this.eventSource.onerror = () => {
               this.eventSource?.close();
               setTimeout(() => this.reconnect(onEvent), this.reconnectDelay);
               this.reconnectDelay = Math.min(this.reconnectDelay * 2, 30000); // Exponential backoff, max 30s
           };
       }

       disconnect() {
           this.eventSource?.close();
       }
   }

   export const sseClient = new SSEClient();
   ```
2. Handle event types: ServerDiscovered, ServerStatusChanged, ServerLogEntry, ConfigFileChanged, ServerMetricsUpdated
3. Update Svelte stores on event received

**Acceptance**:
- SSE connection established
- Events received and parsed
- Auto-reconnect works (kill backend → restart → reconnects)
- Exponential backoff implemented
- Unit test (mock EventSource): Event triggers store update

---

### T-E004 Implement dark theme styling
**File**: `frontend/src/app.css`
**Effort**: M (1-2 hours)
**Dependencies**: T-A003

**Description**:
Implement dark theme per FR-043 using CSS variables.

**Steps**:
1. Create `app.css` with CSS variables:
   ```css
   :root {
       --bg-primary: #1e1e1e;
       --bg-secondary: #2d2d2d;
       --text-primary: #e0e0e0;
       --text-secondary: #a0a0a0;
       --border-color: #3d3d3d;
       --status-running: #4caf50;  /* Green */
       --status-stopped: #f44336;   /* Red */
       --status-starting: #2196f3;  /* Blue */
       --status-error: #ff9800;     /* Yellow/Orange */
       --log-info: #2196f3;         /* Blue */
       --log-success: #4caf50;      /* Green */
       --log-warning: #ff9800;      /* Yellow */
       --log-error: #f44336;        /* Red */
   }

   body {
       background-color: var(--bg-primary);
       color: var(--text-primary);
       font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
   }
   ```
2. Import in `App.svelte`
3. Style all components using variables

**Acceptance**:
- Dark theme applied globally
- Status colors match FR-004 (green/red/blue-gray/yellow)
- Log severity colors match FR-021
- Consistent spacing and alignment (FR-048)

---

### T-E005 Implement main application layout
**File**: `frontend/src/App.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001, T-E004

**Description**:
Create main application layout: header, sidebar, content area, log viewer.

**Steps**:
1. Create `App.svelte` structure:
   ```svelte
   <script lang="ts">
       import ServerTable from './components/ServerTable.svelte';
       import LogViewer from './components/LogViewer.svelte';
       import Sidebar from './components/Sidebar.svelte';
   </script>

   <div class="app-container">
       <header>
           <h1>MCP Manager</h1>
           <button on:click={refreshDiscovery}>Refresh</button>
       </header>

       <div class="main-content">
           <Sidebar />
           <div class="content-area">
               <ServerTable />
           </div>
       </div>

       <LogViewer />
   </div>

   <style>
       .app-container { display: grid; grid-template-rows: auto 1fr auto; height: 100vh; }
       .main-content { display: grid; grid-template-columns: 200px 1fr; }
       /* ... responsive layout */
   </style>
   ```
2. Use CSS Grid for responsive layout (FR-045)
3. Log viewer fixed at bottom, resizable height (FR-020)

**Acceptance**:
- Layout renders correctly
- Responsive resizing works (FR-045)
- Log viewer at bottom (default 200px height)
- Header, sidebar, content area all visible

---

### T-E006 Implement ServerTable component
**File**: `frontend/src/components/ServerTable.svelte`
**Effort**: L (2+ hours)
**Dependencies**: T-E001, T-E002

**Description**:
Display servers in table per FR-003, with Start/Stop/Restart buttons.

**Steps**:
1. Create `ServerTable.svelte`:
   ```svelte
   <script lang="ts">
       import { servers } from '../stores/stores';
       import { apiClient } from '../services/api';

       async function handleStart(serverId: string) {
           await apiClient.startServer(serverId);
       }

       async function handleStop(serverId: string, force: boolean) {
           await apiClient.stopServer(serverId, force, 10);
       }
   </script>

   <table>
       <thead>
           <tr>
               <th>Status</th>
               <th>Name</th>
               <th>Version</th>
               <th>Capabilities</th>
               <th>PID</th>
               <th>Actions</th>
           </tr>
       </thead>
       <tbody>
           {#each $servers as server}
               <tr>
                   <td><span class="status-indicator status-{server.status.state}"></span></td>
                   <td>{server.name}</td>
                   <td>{server.version || 'N/A'}</td>
                   <td>{server.capabilities?.join(', ') || 'N/A'}</td>
                   <td>{server.pid || '-'}</td>
                   <td>
                       {#if server.status.state === 'stopped' || server.status.state === 'error'}
                           <button on:click={() => handleStart(server.id)}>Start</button>
                       {/if}
                       {#if server.status.state === 'running'}
                           <button on:click={() => handleStop(server.id, false)}>Stop</button>
                           <button on:click={() => handleRestart(server.id)}>Restart</button>
                       {/if}
                       <button on:click={() => openConfig(server.id)}>Config</button>
                       <button on:click={() => openLogs(server.id)}>Logs</button>
                   </td>
               </tr>
           {/each}
       </tbody>
   </table>

   <style>
       .status-indicator { display: inline-block; width: 12px; height: 12px; border-radius: 50%; }
       .status-running { background-color: var(--status-running); }
       .status-stopped { background-color: var(--status-stopped); }
       .status-starting { background-color: var(--status-starting); }
       .status-error { background-color: var(--status-error); }
   </style>
   ```
2. Bind to `servers` store (reactive updates)
3. Color-code status per FR-004

**Acceptance**:
- Table displays all servers from store
- Status indicators color-coded correctly
- Buttons enabled/disabled based on state
- Clicking Start calls API and updates UI

---

### T-E007 Implement real-time status updates via SSE
**File**: Modify `App.svelte` and stores
**Effort**: M (1-2 hours)
**Dependencies**: T-E003, T-E006

**Description**:
Connect SSE client to update server table in real-time per FR-005, FR-047.

**Steps**:
1. In `App.svelte` onMount:
   ```typescript
   import { sseClient } from './services/sse';
   import { servers } from './stores/stores';

   onMount(() => {
       sseClient.connect((event) => {
           if (event.type === 'ServerStatusChanged') {
               servers.update(list => {
                   const index = list.findIndex(s => s.id === event.data.serverId);
                   if (index !== -1) {
                       list[index].status.state = event.data.newState;
                       list[index].pid = event.data.pid;
                   }
                   return list;
               });
           }
           // Handle other event types
       });

       return () => sseClient.disconnect();
   });
   ```
2. Update stores on ServerDiscovered, ServerStatusChanged, ServerMetricsUpdated events

**Acceptance**:
- Server status updates in real-time (no manual refresh)
- Start button click → status changes to starting → running
- Backend crash → status updates to error automatically
- SSE reconnects if connection lost

---

### T-E008 Implement LogViewer component
**File**: `frontend/src/components/LogViewer.svelte`
**Effort**: L (2+ hours)
**Dependencies**: T-E001, T-E002

**Description**:
Display real-time logs at bottom of window per FR-020 through FR-023.

**Steps**:
1. Create `LogViewer.svelte`:
   ```svelte
   <script lang="ts">
       import { logs, selectedServer, selectedSeverity } from '../stores/stores';

       let searchQuery = '';

       $: filteredLogs = $logs.filter(log => {
           if ($selectedServer && log.source !== $selectedServer) return false;
           if ($selectedSeverity && log.severity !== $selectedSeverity) return false;
           if (searchQuery && !log.message.toLowerCase().includes(searchQuery.toLowerCase())) return false;
           return true;
       });
   </script>

   <div class="log-viewer">
       <div class="log-toolbar">
           <select bind:value={$selectedServer}>
               <option value={null}>All Servers</option>
               {#each $servers as server}
                   <option value={server.id}>{server.name}</option>
               {/each}
           </select>

           <select bind:value={$selectedSeverity}>
               <option value={null}>All Severities</option>
               <option value="info">INFO</option>
               <option value="success">SUCCESS</option>
               <option value="warning">WARNING</option>
               <option value="error">ERROR</option>
           </select>

           <input type="text" bind:value={searchQuery} placeholder="Search logs..." />
       </div>

       <div class="log-entries">
           {#each filteredLogs as log}
               <div class="log-entry log-{log.severity}">
                   <span class="log-timestamp">{formatTimestamp(log.timestamp)}</span>
                   <span class="log-source">[{log.source}]</span>
                   <span class="log-message">{log.message}</span>
               </div>
           {/each}
       </div>
   </div>

   <style>
       .log-viewer { height: 200px; border-top: 1px solid var(--border-color); overflow-y: auto; }
       .log-info { color: var(--log-info); }
       .log-success { color: var(--log-success); }
       .log-warning { color: var(--log-warning); }
       .log-error { color: var(--log-error); }
   </style>
   ```
2. Implement filtering: server, severity, search (FR-022, FR-023)
3. Color-code by severity (FR-021)
4. Auto-scroll to bottom on new log entry

**Acceptance**:
- Logs display color-coded by severity
- Filter by server dropdown works
- Filter by severity dropdown works
- Search box filters logs in real-time
- Auto-scrolls to newest entry

---

### T-E009 Connect LogViewer to SSE for real-time logs
**File**: Modify `App.svelte` SSE handler
**Effort**: M (1-2 hours)
**Dependencies**: T-E003, T-E008

**Description**:
Update logs store on ServerLogEntry SSE events for real-time log streaming.

**Steps**:
1. In SSE event handler:
   ```typescript
   if (event.type === 'ServerLogEntry') {
       logs.update(list => {
           list.push({
               id: event.id,
               timestamp: event.timestamp,
               severity: event.data.severity,
               source: event.data.serverId,
               message: event.data.message,
               metadata: {}
           });

           // Enforce 1000 entry limit per server (client-side circular buffer)
           // Count entries per source, remove oldest if > 1000
           const sourceEntries = list.filter(l => l.source === event.data.serverId);
           if (sourceEntries.length > 1000) {
               const oldestId = sourceEntries[0].id;
               list = list.filter(l => l.id !== oldestId);
           }

           return list;
       });
   }
   ```

**Acceptance**:
- New log entries appear in real-time
- No page refresh needed
- Client-side buffer enforces 1000 entry limit
- UI performance remains smooth with 50k entries

---

### T-E010 Implement ConfigurationEditor component
**File**: `frontend/src/components/ConfigurationEditor.svelte`
**Effort**: L (2+ hours)
**Dependencies**: T-E001, T-E002

**Description**:
Configuration editor modal/panel per FR-014 through FR-018.

**Steps**:
1. Create `ConfigurationEditor.svelte`:
   ```svelte
   <script lang="ts">
       import { apiClient } from '../services/api';

       export let serverId: string;
       export let onClose: () => void;

       let config: ServerConfiguration;
       let errors: string[] = [];

       onMount(async () => {
           config = await apiClient.getConfiguration(serverId);
       });

       async function saveConfig() {
           try {
               errors = [];
               await apiClient.updateConfiguration(serverId, config);
               alert('Configuration saved successfully');
               onClose();
           } catch (err) {
               errors = [err.message];
           }
       }
   </script>

   <div class="modal">
       <div class="modal-content">
           <h2>Server Configuration</h2>

           <label>
               Environment Variables:
               <table>
                   <thead><tr><th>Key</th><th>Value</th><th></th></tr></thead>
                   <tbody>
                       {#each Object.entries(config.environmentVariables || {}) as [key, value]}
                           <tr>
                               <td><input bind:value={key} /></td>
                               <td><input bind:value={value} /></td>
                               <td><button on:click={() => deleteEnvVar(key)}>Delete</button></td>
                           </tr>
                       {/each}
                   </tbody>
               </table>
               <button on:click={addEnvVar}>Add Variable</button>
           </label>

           <label>
               Command-Line Arguments:
               <textarea bind:value={argsText}></textarea>
           </label>

           <label>
               <input type="checkbox" bind:checked={config.autoStart} />
               Auto-start on launch
           </label>

           <label>
               <input type="checkbox" bind:checked={config.restartOnCrash} />
               Restart on crash
           </label>

           {#if errors.length > 0}
               <div class="errors">
                   {#each errors as error}
                       <p class="error">{error}</p>
                   {/each}
               </div>
           {/if}

           <button on:click={saveConfig}>Save</button>
           <button on:click={onClose}>Cancel</button>
       </div>
   </div>
   ```
2. Validate inputs client-side before submit
3. Display validation errors from API (400 responses)
4. Show read-only client config section (FR-019 - display only, no editing)

**Acceptance**:
- Modal opens when Config button clicked
- Loads current configuration
- Environment variables table editable
- Validation errors displayed
- Save persists changes (verified via reload)
- Client config section clearly marked read-only

---

### T-E011 Implement DetailedLogsView component
**File**: `frontend/src/components/DetailedLogsView.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001, T-E002

**Description**:
Server-specific detailed log view modal per FR-024.

**Steps**:
1. Create `DetailedLogsView.svelte`:
   ```svelte
   <script lang="ts">
       import { apiClient } from '../services/api';

       export let serverId: string;
       export let onClose: () => void;

       let logs: LogEntry[] = [];
       let severityFilter: LogSeverity | null = null;
       let searchQuery = '';

       onMount(async () => {
           await loadLogs();
       });

       async function loadLogs() {
           const response = await apiClient.getServerLogs(serverId, severityFilter, 1000, 0);
           logs = response.logs;
       }

       $: filteredLogs = logs.filter(log => {
           if (severityFilter && log.severity !== severityFilter) return false;
           if (searchQuery && !log.message.toLowerCase().includes(searchQuery.toLowerCase())) return false;
           return true;
       });
   </script>

   <div class="modal">
       <div class="modal-content detailed-logs">
           <h2>Detailed Logs - {serverName}</h2>

           <div class="log-toolbar">
               <select bind:value={severityFilter} on:change={loadLogs}>
                   <option value={null}>All Severities</option>
                   <option value="info">INFO</option>
                   <option value="success">SUCCESS</option>
                   <option value="warning">WARNING</option>
                   <option value="error">ERROR</option>
               </select>

               <input type="text" bind:value={searchQuery} placeholder="Search logs..." />
           </div>

           <div class="log-entries">
               {#each filteredLogs as log}
                   <div class="log-entry log-{log.severity}">
                       <span class="log-timestamp">{log.timestamp}</span>
                       <span class="log-severity">[{log.severity.toUpperCase()}]</span>
                       <span class="log-message">{log.message}</span>
                   </div>
               {/each}
           </div>

           <button on:click={onClose}>Close</button>
       </div>
   </div>
   ```
2. Load up to 1000 most recent entries (FR-053 limit)
3. Client-side filtering and search

**Acceptance**:
- Modal opens when Logs button clicked
- Displays up to 1000 entries
- Color-coded by severity
- Search and filter work
- Scrollable list

---

### T-E012 Implement Sidebar component with utilities
**File**: `frontend/src/components/Sidebar.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001

**Description**:
Sidebar with utility buttons per FR-030 through FR-034.

**Steps**:
1. Create `Sidebar.svelte`:
   ```svelte
   <script lang="ts">
       let activeView = 'servers'; // servers, netstat, shell, explorer, services, help
   </script>

   <aside class="sidebar">
       <button class:active={activeView === 'servers'} on:click={() => activeView = 'servers'}>
           Servers
       </button>
       <button class:active={activeView === 'netstat'} on:click={() => activeView = 'netstat'}>
           Netstat
       </button>
       <button class:active={activeView === 'shell'} on:click={() => activeView = 'shell'}>
           Shell
       </button>
       <button class:active={activeView === 'explorer'} on:click={() => activeView = 'explorer'}>
           Explorer
       </button>
       <button class:active={activeView === 'services'} on:click={() => activeView = 'services'}>
           Services
       </button>
       <button class:active={activeView === 'help'} on:click={() => activeView = 'help'}>
           Help
       </button>
   </aside>

   <style>
       .sidebar { display: flex; flex-direction: column; background-color: var(--bg-secondary); padding: 10px; }
       .sidebar button { margin-bottom: 10px; text-align: left; }
       .sidebar button.active { background-color: var(--status-running); }
   </style>
   ```
2. Clicking utility buttons switches main content area view

**Acceptance**:
- Sidebar displays all utility buttons
- Active button highlighted
- Clicking button changes content area (placeholder views OK for now)

---

### T-E013 Implement Netstat utility view
**File**: `frontend/src/components/NetstatView.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001, T-E012

**Description**:
Display network connections per FR-030 by calling backend API.

**Steps**:
1. Add backend API method: `GET /api/v1/netstat?pids=<comma-separated>`
   - Backend: Parse `netstat` output, filter by PIDs of running MCP servers
   - Return: [{protocol, localAddress, remoteAddress, state, pid}]
2. Create `NetstatView.svelte`:
   ```svelte
   <script lang="ts">
       import { apiClient } from '../services/api';

       let connections: NetstatEntry[] = [];

       onMount(async () => {
           const pids = $servers.filter(s => s.pid).map(s => s.pid).join(',');
           connections = await apiClient.getNetstat(pids);
       });
   </script>

   <div class="netstat-view">
       <h2>Network Connections (MCP Servers)</h2>
       <table>
           <thead>
               <tr><th>Protocol</th><th>Local</th><th>Remote</th><th>State</th><th>PID</th></tr>
           </thead>
           <tbody>
               {#each connections as conn}
                   <tr>
                       <td>{conn.protocol}</td>
                       <td>{conn.localAddress}</td>
                       <td>{conn.remoteAddress}</td>
                       <td>{conn.state}</td>
                       <td>{conn.pid}</td>
                   </tr>
               {/each}
           </tbody>
       </table>
   </div>
   ```

**Acceptance**:
- Netstat view displays connections for running MCP servers
- Table shows protocol, addresses, state, PID
- Works on Windows, macOS, Linux

---

### T-E014 Implement Shell utility view
**File**: `frontend/src/components/ShellView.svelte`
**Effort**: S (<1 hour)
**Dependencies**: T-E001, T-E012

**Description**:
Launch platform shell per FR-031 via backend API.

**Steps**:
1. Add backend API method: `POST /api/v1/shell`
   - Backend: Launch platform shell (cmd.exe, Terminal.app, xterm) via os/exec
   - Return: {success: bool, message: string}
2. Create `ShellView.svelte`:
   ```svelte
   <script lang="ts">
       import { apiClient } from '../services/api';

       async function openShell() {
           await apiClient.launchShell();
           alert('Shell opened');
       }
   </script>

   <div class="shell-view">
       <h2>Quick Shell Access</h2>
       <p>Launch a platform-appropriate shell for quick terminal access.</p>
       <button on:click={openShell}>Open Shell</button>
   </div>
   ```

**Acceptance**:
- Button click launches shell externally
- Correct shell per platform (cmd.exe on Windows, etc.)

---

### T-E015 Implement Explorer utility view
**File**: `frontend/src/components/ExplorerView.svelte`
**Effort**: S (<1 hour)
**Dependencies**: T-E001, T-E012

**Description**:
Open server installation directories per FR-032 via backend API.

**Steps**:
1. Add backend API method: `POST /api/v1/explorer?path=<path>`
   - Backend: Launch file explorer (explorer, open, xdg-open) with path
2. Create `ExplorerView.svelte`:
   ```svelte
   <script lang="ts">
       import { servers } from '../stores/stores';
       import { apiClient } from '../services/api';

       async function openInExplorer(path: string) {
           await apiClient.openExplorer(path);
       }
   </script>

   <div class="explorer-view">
       <h2>Server Installation Directories</h2>
       <ul>
           {#each $servers as server}
               <li>
                   {server.name}: {server.installationPath}
                   <button on:click={() => openInExplorer(server.installationPath)}>Open</button>
               </li>
           {/each}
       </ul>
   </div>
   ```

**Acceptance**:
- Lists all server installation paths
- Open button launches file explorer at path
- Works on all platforms

---

### T-E016 Implement Services utility view
**File**: `frontend/src/components/ServicesView.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001, T-E012

**Description**:
View system service management per FR-033 via backend API.

**Steps**:
1. Add backend API method: `GET /api/v1/services`
   - Backend: Run platform service commands (sc query, launchctl list, systemctl list-units)
   - Return: [{name, status, description}]
2. Create `ServicesView.svelte`:
   ```svelte
   <script lang="ts">
       import { apiClient } from '../services/api';

       let services: Service[] = [];

       onMount(async () => {
           services = await apiClient.getServices();
       });
   </script>

   <div class="services-view">
       <h2>System Services</h2>
       <table>
           <thead><tr><th>Name</th><th>Status</th><th>Description</th></tr></thead>
           <tbody>
               {#each services as service}
                   <tr>
                       <td>{service.name}</td>
                       <td>{service.status}</td>
                       <td>{service.description}</td>
                   </tr>
               {/each}
           </tbody>
       </table>
   </div>
   ```

**Acceptance**:
- Lists system services
- Shows status (running/stopped)
- Works on all platforms

---

### T-E017 Implement Help utility view
**File**: `frontend/src/components/HelpView.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001, T-E012

**Description**:
Display embedded documentation per FR-034.

**Steps**:
1. Create `HelpView.svelte`:
   ```svelte
   <script lang="ts">
       let activeTab = 'quickstart'; // quickstart, keyboard-shortcuts, about
   </script>

   <div class="help-view">
       <div class="help-tabs">
           <button class:active={activeTab === 'quickstart'} on:click={() => activeTab = 'quickstart'}>
               Quickstart
           </button>
           <button class:active={activeTab === 'keyboard-shortcuts'} on:click={() => activeTab = 'keyboard-shortcuts'}>
               Keyboard Shortcuts
           </button>
           <button class:active={activeTab === 'about'} on:click={() => activeTab = 'about'}>
               About
           </button>
       </div>

       <div class="help-content">
           {#if activeTab === 'quickstart'}
               <h2>Quickstart Guide</h2>
               <p>Getting started with MCP Manager...</p>
               <!-- Embed quickstart.md content or simplified version -->
           {:else if activeTab === 'keyboard-shortcuts'}
               <h2>Keyboard Shortcuts</h2>
               <table>
                   <tr><td>Ctrl/Cmd+S</td><td>Start selected server</td></tr>
                   <tr><td>Ctrl/Cmd+T</td><td>Stop selected server</td></tr>
                   <tr><td>Ctrl/Cmd+R</td><td>Restart selected server</td></tr>
                   <tr><td>F5</td><td>Refresh discovery</td></tr>
                   <tr><td>Ctrl/Cmd+F</td><td>Focus search</td></tr>
                   <tr><td>Ctrl/Cmd+L</td><td>Toggle logs panel</td></tr>
               </table>
           {:else if activeTab === 'about'}
               <h2>About MCP Manager</h2>
               <p>Version 1.0.0</p>
               <p>Cross-platform desktop application for managing MCP servers.</p>
               <p>© 2025 Your Organization</p>
           {/if}
       </div>
   </div>
   ```

**Acceptance**:
- Help view displays documentation tabs
- Quickstart guide readable
- Keyboard shortcuts listed (per research.md §14)
- About section shows version

---

### T-E018 Implement keyboard shortcuts
**File**: `frontend/src/App.svelte` global key handler
**Effort**: M (1-2 hours)
**Dependencies**: T-E001

**Description**:
Global keyboard shortcuts per FR-046 and research.md §14.

**Steps**:
1. In `App.svelte`, add global keydown handler:
   ```typescript
   function handleKeydown(event: KeyboardEvent) {
       const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0;
       const modKey = isMac ? event.metaKey : event.ctrlKey;

       if (!modKey) return;

       switch (event.key.toLowerCase()) {
           case 's': // Start server
               event.preventDefault();
               if ($selectedServer) handleStart($selectedServer);
               break;
           case 't': // Stop server
               event.preventDefault();
               if ($selectedServer) handleStop($selectedServer);
               break;
           case 'r': // Restart server
               event.preventDefault();
               if ($selectedServer) handleRestart($selectedServer);
               break;
           case 'f': // Focus search
               event.preventDefault();
               document.querySelector('input[type="text"]')?.focus();
               break;
           case 'l': // Toggle logs
               event.preventDefault();
               toggleLogPanel();
               break;
       }

       if (event.key === 'F5') { // Refresh discovery
           event.preventDefault();
           refreshDiscovery();
       }
   }

   onMount(() => {
       window.addEventListener('keydown', handleKeydown);
       return () => window.removeEventListener('keydown', handleKeydown);
   });
   ```
2. Disable shortcuts when input fields focused

**Acceptance**:
- Keyboard shortcuts work per research.md §14 table
- Platform-aware (Cmd on macOS, Ctrl on Windows/Linux)
- Shortcuts disabled when typing in input fields
- Visual indicators (tooltips show shortcuts)

---

### T-E019 Implement config file change notifications
**File**: `frontend/src/App.svelte` SSE handler
**Effort**: M (1-2 hours)
**Dependencies**: T-E003

**Description**:
Display notification when external config changes detected per clarification (hybrid approach).

**Steps**:
1. In SSE event handler:
   ```typescript
   if (event.type === 'ConfigFileChanged') {
       showNotification({
           message: `Configuration file changed: ${event.data.filePath}. Refresh discovery?`,
           actions: [
               { label: 'Refresh', callback: () => refreshDiscovery() },
               { label: 'Dismiss', callback: () => {} }
           ]
       });
   }
   ```
2. Create `Notification.svelte` component:
   ```svelte
   <script lang="ts">
       export let message: string;
       export let actions: {label: string, callback: () => void}[];
   </script>

   <div class="notification">
       <p>{message}</p>
       <div class="notification-actions">
           {#each actions as action}
               <button on:click={action.callback}>{action.label}</button>
           {/each}
       </div>
   </div>

   <style>
       .notification { position: fixed; top: 20px; right: 20px; background: var(--bg-secondary); border: 1px solid var(--border-color); padding: 15px; border-radius: 5px; z-index: 1000; }
   </style>
   ```

**Acceptance**:
- Notification appears when config file changed
- User can click Refresh or Dismiss
- Refresh triggers discovery scan
- Dismiss closes notification

---

### T-E020 Implement single-instance window activation
**File**: Backend `cmd/mcpmanager/main.go`, Wails config
**Effort**: M (1-2 hours)
**Dependencies**: T-A008, T-E001

**Description**:
Bring existing window to foreground when second instance launched per FR-051.

**Steps**:
1. In `main.go`:
   ```go
   func main() {
       singleton := platform.NewSingleInstance()
       acquired, err := singleton.Acquire()
       if !acquired {
           // Signal existing instance to show window
           // On Windows: Find window by title, call SetForegroundWindow
           // On Unix: Send signal to existing process
           return
       }
       defer singleton.Release()

       // ... start Wails app
   }
   ```
2. Add window show handler in Wails app (listen for signal)

**Acceptance**:
- Second launch brings first window to foreground
- No duplicate processes
- Verified on Windows, macOS, Linux

---

### T-E021 Implement responsive window resizing
**File**: `frontend/src/app.css`, component styles
**Effort**: M (1-2 hours)
**Dependencies**: T-E005

**Description**:
Responsive layout per FR-045.

**Steps**:
1. Use CSS Grid with fr units for flexible sizing:
   ```css
   .app-container {
       display: grid;
       grid-template-rows: auto 1fr auto;
       height: 100vh;
   }

   .main-content {
       display: grid;
       grid-template-columns: 200px 1fr;
   }

   @media (max-width: 1024px) {
       .main-content {
           grid-template-columns: 1fr; /* Stack sidebar on narrow screens */
       }
   }
   ```
2. Resizable log panel:
   - Add drag handle between content and log viewer
   - Save height to ApplicationState.windowLayout.logPanelHeight

**Acceptance**:
- Window resizes smoothly
- Layout reflows without breaking
- Log panel height adjustable via drag
- Saved log panel height persists across restarts

---

### T-E022 Implement window state persistence
**File**: `frontend/src/App.svelte`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001, T-E002

**Description**:
Persist window size/position per FR-041 ApplicationState.windowLayout.

**Steps**:
1. On window resize/move:
   ```typescript
   import { appState } from './stores/stores';
   import { apiClient } from './services/api';

   function saveWindowLayout() {
       const layout = {
           width: window.innerWidth,
           height: window.innerHeight,
           x: window.screenX,
           y: window.screenY,
           maximized: window.outerWidth === screen.width && window.outerHeight === screen.height,
           logPanelHeight: $appState.windowLayout.logPanelHeight
       };

       appState.update(state => ({ ...state, windowLayout: layout }));
       apiClient.updateApplicationState($appState); // Debounced auto-save on backend
   }

   window.addEventListener('resize', saveWindowLayout);
   window.addEventListener('beforeunload', saveWindowLayout);
   ```
2. On app load, restore window layout from ApplicationState

**Acceptance**:
- Window size/position saved on close
- Restored on next launch
- Maximized state preserved

---

### T-E023 Implement log severity color coding
**File**: `frontend/src/components/LogViewer.svelte` styles
**Effort**: S (<1 hour)
**Dependencies**: T-E008

**Description**:
Color-code log entries per FR-021.

**Steps**:
1. Add CSS classes per severity:
   ```css
   .log-entry.log-info { color: var(--log-info); } /* Blue */
   .log-entry.log-success { color: var(--log-success); } /* Green */
   .log-entry.log-warning { color: var(--log-warning); } /* Yellow */
   .log-entry.log-error { color: var(--log-error); } /* Red */
   ```
2. Apply class dynamically: `<div class="log-entry log-{log.severity}">`

**Acceptance**:
- INFO logs blue
- SUCCESS logs green
- WARNING logs yellow
- ERROR logs red
- Colors consistent across LogViewer and DetailedLogsView

---

### T-E024 Implement status indicator color coding
**File**: `frontend/src/components/ServerTable.svelte` styles
**Effort**: S (<1 hour)
**Dependencies**: T-E006

**Description**:
Color-code server status per FR-004.

**Steps**:
1. Add CSS classes per status:
   ```css
   .status-indicator.status-stopped { background-color: var(--status-stopped); } /* Red */
   .status-indicator.status-starting { background-color: var(--status-starting); } /* Blue/Gray */
   .status-indicator.status-running { background-color: var(--status-running); } /* Green */
   .status-indicator.status-error { background-color: var(--status-error); } /* Yellow */
   ```
2. Apply class dynamically: `<span class="status-indicator status-{server.status.state}"></span>`

**Acceptance**:
- Stopped: red
- Starting: blue/gray
- Running: green
- Error: yellow
- Color mapping matches data-model.md §2

---

### T-E025 Implement UI responsiveness optimization
**File**: All Svelte components
**Effort**: M (1-2 hours)
**Dependencies**: T-E006 through T-E024

**Description**:
Optimize for <200ms UI response per FR-038, performance validation tests.

**Steps**:
1. Debounce search input (300ms delay):
   ```typescript
   let searchQuery = '';
   let debouncedSearch = '';

   $: {
       clearTimeout(searchTimeout);
       searchTimeout = setTimeout(() => {
           debouncedSearch = searchQuery;
       }, 300);
   }
   ```
2. Virtualize long log lists (render only visible entries):
   - Use `svelte-virtual-list` or similar for 1000+ entry lists
3. Memoize expensive computations:
   ```typescript
   import { derived } from 'svelte/store';

   const filteredLogs = derived([logs, selectedServer, selectedSeverity], ([$logs, $selectedServer, $selectedSeverity]) => {
       return $logs.filter(log => {
           // ... filtering logic
       });
   });
   ```
4. Profile with Chrome DevTools, optimize hot paths

**Acceptance**:
- Button clicks respond within 200ms (per quickstart.md performance test)
- Log filtering <50ms for 50k entries
- Search <300ms
- UI remains smooth during background operations

---

### T-E026 Write frontend unit tests for stores
**File**: `frontend/tests/stores.test.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E001

**Description**:
Unit tests for Svelte stores using Vitest.

**Test Cases**:
1. servers store: Update works, reactive
2. logs store: Add entry, enforce 1000 limit per server
3. appState store: Defaults set correctly
4. selectedServer store: Null by default, updates
5. Derived stores compute correctly

**Acceptance**:
- All store tests pass
- Coverage > 80% for stores.ts
- Tests use Vitest or similar

---

### T-E027 Write frontend unit tests for API client
**File**: `frontend/tests/api.test.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E002

**Description**:
Unit tests for API client with mocked fetch.

**Test Cases**:
1. listServers() returns typed data
2. startServer() sends POST with correct URL
3. Error handling: 404 throws descriptive error
4. Error handling: Network failure throws error
5. Query parameters encoded correctly

**Acceptance**:
- All API client tests pass
- Uses fetch mock (msw or vitest.mock)
- Coverage > 80% for api.ts

---

### T-E028 Write frontend unit tests for SSE client
**File**: `frontend/tests/sse.test.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E003

**Description**:
Unit tests for SSE client with mocked EventSource.

**Test Cases**:
1. connect() establishes EventSource
2. onmessage handler calls callback with parsed event
3. onerror triggers reconnect with exponential backoff
4. disconnect() closes EventSource
5. Reconnect includes Last-Event-ID header

**Acceptance**:
- All SSE client tests pass
- Uses EventSource mock
- Coverage > 80% for sse.ts

---

### T-E029 Write frontend component tests for ServerTable
**File**: `frontend/tests/ServerTable.test.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E006

**Description**:
Component tests for ServerTable using Svelte Testing Library.

**Test Cases**:
1. Renders list of servers from store
2. Start button enabled when status=stopped
3. Stop button enabled when status=running
4. Clicking Start calls apiClient.startServer()
5. Status color indicator matches server.status.state

**Acceptance**:
- All component tests pass
- Uses @testing-library/svelte
- Coverage > 70% for ServerTable.svelte

---

### T-E030 Write frontend component tests for LogViewer
**File**: `frontend/tests/LogViewer.test.ts`
**Effort**: M (1-2 hours)
**Dependencies**: T-E008

**Description**:
Component tests for LogViewer using Svelte Testing Library.

**Test Cases**:
1. Renders logs from store
2. Filter by server dropdown works
3. Filter by severity dropdown works
4. Search input filters logs
5. Color coding applied correctly

**Acceptance**:
- All component tests pass
- Uses @testing-library/svelte
- Coverage > 70% for LogViewer.svelte

---

## Phase F: Integration & Testing (Tasks F001-F010)

**Objective**: End-to-end tests, performance validation, packaging

### T-F001 [P] Write integration test for quickstart scenario 1
**File**: `tests/integration/quickstart_01_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: All Phase D/E tasks

**Description**:
Automated version of quickstart.md Test Scenario 1: Initial launch & discovery.

**Test Steps**:
1. Create temp directory with mock client config
2. Launch MCP Manager (headless or with test window)
3. Wait for discovery to complete (poll GET /servers until count > 0)
4. Verify:
   - At least 1 server discovered
   - Application state file created (~/.mcpmanager/state.json)
   - Logs directory created

**Acceptance**:
- Test passes consistently
- Startup time < 2 seconds (FR-037)
- Cleanup after test

---

### T-F002 [P] Write integration test for quickstart scenario 2
**File**: `tests/integration/quickstart_02_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: All Phase D/E tasks

**Description**:
Automated version of quickstart.md Test Scenario 2: Start server lifecycle.

**Test Steps**:
1. Create test server: Simple HTTP server script
2. Discover server
3. Call POST /servers/{id}/start
4. Poll GET /servers/{id}/status until state=running (timeout 10s)
5. Verify PID set, HTTP server responding
6. Call POST /servers/{id}/stop
7. Verify state=stopped, PID cleared

**Acceptance**:
- Test passes consistently
- State transitions correct: stopped → starting → running → stopped
- SSE events published for each transition

---

### T-F003 [P] Write integration test for quickstart scenario 3
**File**: `tests/integration/quickstart_03_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: All Phase D/E tasks

**Description**:
Automated version of quickstart.md Test Scenario 3: Log filtering.

**Test Steps**:
1. Start 2 test servers
2. Generate logs from both (write to stdout)
3. Call GET /logs?serverId={server1-id}
4. Verify only server1 logs returned
5. Call GET /logs?severity=error
6. Verify only error logs returned
7. Call GET /logs?search=keyword
8. Verify only matching logs returned

**Acceptance**:
- Test passes consistently
- Filters work correctly (server, severity, search)
- Performance: <50ms filter time for 1000 entries

---

### T-F004 [P] Write integration test for quickstart scenario 4
**File**: `tests/integration/quickstart_04_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: All Phase D/E tasks

**Description**:
Automated version of quickstart.md Test Scenario 4: Configuration editing.

**Test Steps**:
1. Discover test server
2. Call GET /servers/{id}/configuration
3. Modify config: Add env var TEST_VAR=test_value
4. Call PUT /servers/{id}/configuration with updated config
5. Verify 200 response
6. Call GET /servers/{id}/configuration again
7. Verify TEST_VAR persisted
8. Verify client config file UNCHANGED (FR-019 critical check)

**Acceptance**:
- Test passes consistently
- Configuration persists across API calls
- Client config files never modified (assert unchanged)

---

### T-F005 [P] Write integration test for edge case: server crash
**File**: `tests/integration/edge_crash_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: All Phase D/E tasks

**Description**:
Automated version of quickstart.md Edge Case 2: Server crash during operation.

**Test Steps**:
1. Start test server
2. Verify status=running
3. Kill process externally (os.Kill with PID)
4. Poll GET /servers/{id}/status (timeout 10s)
5. Verify status transitions to error within 5 seconds (monitor goroutine detects)
6. Verify error message contains "crashed" or "exited unexpectedly"

**Acceptance**:
- Test passes consistently
- Crash detected within 5 seconds
- Status transitions to error automatically
- Crash logs captured

---

### T-F006 [P] Write integration test for edge case: external config change
**File**: `tests/integration/edge_config_change_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: All Phase D/E tasks

**Description**:
Automated version of quickstart.md Edge Case 6: External config file modification.

**Test Steps**:
1. Start MCP Manager, subscribe to SSE /events
2. Externally modify client config file (add new server)
3. Wait for ConfigFileChanged SSE event (timeout 5s)
4. Verify event received with correct filePath
5. Call POST /servers/discover
6. Verify new server discovered

**Acceptance**:
- Test passes consistently
- File watcher detects change within 2 seconds
- ConfigFileChanged event published
- Discovery picks up new server after manual trigger

---

### T-F007 Write performance benchmark: startup time
**File**: `tests/performance/startup_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: All Phase D/E tasks

**Description**:
Benchmark startup time per FR-037 (<2 seconds).

**Test Steps**:
1. Measure cold start: Launch MCP Manager, time until HTTP server responds
2. Measure warm start: Launch again (filesystem cache warm), time until ready
3. Run 10 iterations, report min/max/avg
4. Assert avg cold start < 2000ms

**Acceptance**:
- Benchmark runs successfully
- Reports timing in milliseconds
- Cold start consistently < 2 seconds

---

### T-F008 Write performance benchmark: memory usage
**File**: `tests/performance/memory_test.go`
**Effort**: M (1-2 hours)
**Dependencies**: All Phase D/E tasks

**Description**:
Benchmark memory usage per FR-039 (<100MB idle), FR-054 (50 servers).

**Test Steps**:
1. Start MCP Manager with no servers
2. Wait 10 seconds for stabilization
3. Measure RSS (Resident Set Size) via runtime.ReadMemStats or ps
4. Assert RSS < 100MB
5. Discover 50 mock servers
6. Start all 50 servers
7. Wait 10 seconds
8. Measure RSS
9. Assert RSS < 300MB (50 servers * ~4MB + 100MB base)

**Acceptance**:
- Benchmark runs successfully
- Idle memory < 100MB
- 50 servers memory < 300MB
- No memory leaks (RSS stable over 5 minutes)

---

### T-F009 Setup CI/CD pipeline with GitHub Actions
**File**: `.github/workflows/ci.yml`
**Effort**: M (1-2 hours)
**Dependencies**: All prior tasks

**Description**:
Setup continuous integration for automated testing.

**Steps**:
1. Create `.github/workflows/ci.yml`:
   ```yaml
   name: CI

   on: [push, pull_request]

   jobs:
     test-backend:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v3
         - uses: actions/setup-go@v4
           with:
             go-version: '1.21'
         - run: go mod download
         - run: go test ./... -race -coverprofile=coverage.txt
         - uses: codecov/codecov-action@v3

     test-frontend:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v3
         - uses: actions/setup-node@v3
           with:
             node-version: '18'
         - run: cd frontend && npm ci
         - run: cd frontend && npm run test
         - run: cd frontend && npm run check # TypeScript check

     build:
       strategy:
         matrix:
           os: [windows-latest, macos-latest, ubuntu-latest]
       runs-on: ${{ matrix.os }}
       steps:
         - uses: actions/checkout@v3
         - uses: actions/setup-go@v4
         - run: go install github.com/wailsapp/wails/v2/cmd/wails@latest
         - run: wails build -clean
         - uses: actions/upload-artifact@v3
           with:
             name: mcpmanager-${{ matrix.os }}
             path: build/bin/
   ```

**Acceptance**:
- CI runs on push/PR
- Tests run on Ubuntu
- Builds succeed on Windows, macOS, Linux
- Artifacts uploaded

---

### T-F010 Create Wails production build and packaging
**File**: `build/` scripts
**Effort**: L (2+ hours)
**Dependencies**: All prior tasks

**Description**:
Create production builds and installers per quickstart.md binary size targets.

**Steps**:
1. Run `wails build -clean -upx` for each platform:
   - Windows: `wails build -platform windows/amd64 -clean -upx`
   - macOS: `wails build -platform darwin/universal -clean -upx`
   - Linux: `wails build -platform linux/amd64 -clean -upx`
2. Measure binary sizes:
   - Target: <50MB uncompressed, <20MB with UPX
3. Create installers:
   - Windows: NSIS installer (target <30MB)
   - macOS: DMG bundle
   - Linux: .deb and .rpm packages
4. Test installers: Install, launch, verify functionality

**Acceptance**:
- Binaries build successfully for all 3 platforms
- Binary size < 50MB (UPX compressed < 20MB)
- Installers created and tested
- Application launches and passes quickstart scenario 1 on all platforms

---

## Execution Summary

**Total Tasks**: 98
**Estimated Effort**: 150-180 hours (3-4 weeks for solo developer)

### Task Dependencies Visualization

```
Phase A (Foundation) [8 tasks, ~12 hours]
  ↓
Phase B (Models) [12 tasks, ~18 hours] ─┐
  ↓                                      │
Phase C (Services) [20 tasks, ~35 hours]├→ Phase D (API) [18 tasks, ~30 hours]
                                         │       ↓
                                         └→ Phase E (Frontend) [30 tasks, ~50 hours]
                                                 ↓
                                           Phase F (Testing) [10 tasks, ~18 hours]
```

### Parallel Execution Opportunities

- **Phase A**: Tasks A001-A008 can run concurrently (8 parallel tasks)
- **Phase B**: Tasks B001-B012 can run concurrently (12 parallel tasks)
- **Phase D**: Tasks D001-D008 (contract tests) can run concurrently (8 parallel tasks)
- **Phase F**: Tasks F001-F006 (integration tests) can run concurrently (6 parallel tasks)

**Maximum parallelism**: ~30% of tasks can run concurrently, reducing wall-clock time to ~105-125 hours with 2-3 developers.

---

## Next Steps

1. **Review tasks.md** with team/stakeholders
2. **Assign tasks** to developers (prefer assigning by phase for focus)
3. **Setup development environment**: Run T-A001 through T-A003 first
4. **Begin TDD cycle**: Write contract tests (Phase D), then implement to pass tests
5. **Track progress**: Mark tasks complete in this document or use project management tool
6. **Run `/implement`** command (if available) to begin automated execution

---

## References

- **Plan**: `specs/001-mcp-manager-specification/plan.md`
- **Research**: `specs/001-mcp-manager-specification/research.md`
- **Data Model**: `specs/001-mcp-manager-specification/data-model.md`
- **API Contracts**: `specs/001-mcp-manager-specification/contracts/api-spec.yaml`
- **Quickstart Tests**: `specs/001-mcp-manager-specification/quickstart.md`
- **Feature Spec**: `specs/001-mcp-manager-specification/spec.md`
- **Constitution**: `.specify/memory/constitution.md`
