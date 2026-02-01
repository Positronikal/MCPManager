# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MCP Manager is a cross-platform desktop application for managing Model Context Protocol (MCP) servers. Built with Go (backend) and Wails v2 (desktop framework with Svelte frontend), it provides centralized discovery, monitoring, and lifecycle management of MCP servers across different clients (Claude Desktop, Cursor, etc.).

**Key Technologies:**
- **Backend:** Go 1.21+ (service-oriented architecture)
- **Frontend:** Svelte 4.x with TypeScript
- **Desktop Framework:** Wails v2 (Go + Web UI)
- **Cross-Platform:** Windows, macOS, Linux

## Development Commands

### Building and Running

```bash
# Development mode with hot reload
wails dev

# Production build
wails build

# Clean build (recommended after major changes)
wails build -clean

# Output: build/bin/mcpmanager.exe (Windows) or build/bin/mcpmanager (Unix)
```

### Testing

**Backend Tests:**
```bash
# All tests
go test ./...

# Specific package
go test ./internal/core/discovery/...

# With coverage
go test -cover ./...

# Specific test types
go test -v ./tests/unit/...          # Unit tests
go test -v ./tests/integration/...   # Integration tests
go test -v ./tests/contract/...      # API contract tests
go test -v ./tests/performance/...   # Performance benchmarks

# Run specific test
go test -run TestServerLifecycle ./tests/...
```

**Frontend Tests:**
```bash
cd frontend

# Run all tests
npm test

# Interactive test UI
npm run test:ui

# With coverage
npm run test:coverage

# TypeScript check
npm run check
```

### Pre-Commit Verification

Run local verification scripts before committing to ensure all quality gates pass:

```bash
# Unix/macOS/Linux
./scripts/verify-build.sh

# Windows
scripts\verify-build.bat

# Quick mode (faster, skips race detection and integration tests)
./scripts/verify-build.sh --quick

# Skip Wails build (for iterative development)
./scripts/verify-build.sh --skip-build

# Combine options
./scripts/verify-build.sh --quick --skip-build
```

**Recommended workflow:**
1. During development: `--quick --skip-build`
2. Before committing: `--skip-build`
3. Before pushing: full verification (no flags)

### Code Quality

```bash
# Backend
go fmt ./...       # Format code
go vet ./...       # Static analysis
go mod tidy        # Clean dependencies

# Frontend
cd frontend
npm run check      # TypeScript checking
```

## Architecture Overview

### High-Level Structure

```
Frontend (Svelte) ←→ Wails IPC Bridge ←→ Go Backend
                                          ↓
                                    Services → EventBus → Real-time Events
```

### Backend Services Architecture

The backend follows a **service-oriented architecture** with clear separation of concerns:

**Core Services** (`internal/core/`):

1. **DiscoveryService** - Multi-source MCP server discovery
   - ClientConfigDiscovery: Parses Claude Desktop/Cursor configs
   - ClaudeExtensionsDiscovery: Finds Claude Extension servers
   - FilesystemDiscovery: Scans installation directories
   - ProcessDiscovery: Matches running processes to servers
   - ConfigFileWatcher: Monitors config files for external changes

2. **LifecycleService** - Process lifecycle management (start/stop/restart)
   - State machine: Stopped → Starting → Running → Error
   - Process health monitoring with crash detection
   - Output capture integration
   - PID validation (detects stale processes)

3. **MonitoringService** - Log capture and buffering
   - CircularLogBuffer: 1000-entry ring buffer per server
   - Real-time log streaming via EventBus
   - Severity parsing (error/warning/success/info)

4. **MetricsCollector** - Resource usage monitoring
   - CPU and memory metrics
   - Rate-limited collection (1Hz per server)
   - Uptime tracking

5. **ConfigService** - Server configuration persistence
   - Per-server config storage: `~/.mcpmanager/servers/{serverID}/config.json`
   - Atomic writes with backup
   - Configuration validation

6. **DependencyService** - Dependency checking and validation
   - Runtime detection (Node.js, Python, Go)
   - Tool detection (npm, pip, git)
   - Platform-aware installation instructions

7. **EventBus** - Lightweight pub/sub system
   - Non-blocking event publishing
   - Key events: `server.discovered`, `server.status.changed`, `server.log.entry`, `config.file.changed`, `server.metrics.updated`

**Platform Abstraction Layer** (`internal/platform/`):
- PathResolver, ProcessManager, ProcessInfo interfaces
- Platform-specific implementations: `*_windows.go`, `*_darwin.go`, `*_linux.go`
- Use build tags for platform-specific code

**Models** (`internal/models/`):
- MCPServer: Core server entity with deterministic UUID (hash of name+path+source)
- ServerStatus: State machine for server lifecycle
- ServerConfiguration: Per-server settings and preferences
- TransportType: stdio (requires client) vs http/sse/unknown (standalone)

### Communication Flow

**Service Coordination:**
```
Discovery → finds servers → caches results
    ↓
Lifecycle → starts/stops → updates Discovery cache → emits events
    ↓
Monitoring → captures logs → emits events
    ↓
EventBus → publishes events → Wails runtime → Frontend
```

**Critical Synchronization:** LifecycleService always calls `DiscoveryService.UpdateServer()` after state changes to keep the cache synchronized (fixes BUG-001).

### Frontend Architecture

**Structure** (`frontend/src/`):
- `App.svelte`: Main component with routing
- `components/`: Reusable UI components (ServerTable, modals, utilities)
- `services/`: API wrappers and event handlers
- `stores/`: Svelte stores for state management
- `types/`: TypeScript definitions

**Wails Integration:**
- Backend methods: `wailsjs/go/main/App.*`
- Real-time events: `runtime.EventsOn()` from `@wailsapp/runtime`
- Event flow: Backend emits → Frontend receives via runtime

### Transport Handling

- **stdio servers:** Cannot be started directly by MCP Manager; require client configuration. `StartServer()` returns error `"stdio_requires_client"` with guidance to use config editor.
- **http/sse/unknown servers:** Can be started/stopped directly by MCP Manager via LifecycleService.

## Key Implementation Patterns

### Event-Driven Architecture
```go
// Services publish events
eventBus.Publish(events.Event{
    Type: events.EventServerStatusChanged,
    Data: statusData,
})

// Frontend subscribes via Wails runtime
runtime.EventsEmit(ctx, "server:status:changed", statusData)
```

### Service Initialization Sequence
```go
// app.go startup()
1. EventBus initialization
2. Storage service
3. Discovery service
4. Monitoring service
5. Lifecycle service (depends on Discovery + Monitoring)
6. Config service
7. Initial discovery scan
```

### Graceful Shutdown Sequence
```go
// app.go shutdown()
1. Stop all managed servers (LifecycleService.StopAll())
2. Close file watchers (DiscoveryService.Close())
3. Close EventBus (prevents goroutine leaks)
```

### Thread Safety
- Services use `sync.RWMutex` for cache access
- CircularLogBuffer is thread-safe
- EventBus channels are non-blocking (drops events if full)

## Specification-Driven Development

This project follows **Spec Kit** methodology for feature development. All specifications live in `specs/[###-feature-name]/`.

**Key Documents:**
- `spec.md`: Functional requirements (what & why)
- `plan.md`: Technical implementation plan (how)
- `tasks.md`: Task index with links to modular phase files
- `tasks/`: Modular task breakdown by phase
- `.specify/memory/constitution.md`: Project governing principles

**Development Workflow:**
1. Review constitution: `/constitution`
2. Define requirements: `/specify <feature description>`
3. Clarify ambiguities: `/clarify`
4. Create technical plan: `/plan`
5. Generate task breakdown: `/tasks`
6. Validate consistency: `/analyze`
7. Execute implementation: `/implement`

**Project Constitution Highlights:**
- Unix Philosophy: Modularity, simplicity, composability
- Primary Language: Go for system software
- Architecture: API-first, cross-platform, separation of concerns
- Performance Targets: <2s startup, <100MB memory idle
- See `.specify/memory/constitution.md` for full principles

## Important Notes

### Server ID Generation
Server IDs are deterministic (hash of name+path+source) to maintain stability across app restarts and re-discoveries.

### Discovery Priority
Three-tier strategy:
1. **PRIMARY:** Client configs (Claude Desktop, Cursor)
2. **SECONDARY:** Filesystem scanning (npm, pip, Go installations)
3. **TERTIARY:** Process matching (running processes)

Servers are merged by name with priority-based deduplication.

### State Machine Validation
ServerStatus follows strict state transitions. Invalid transitions are rejected to prevent inconsistent state.

### Platform-Specific Code
Always use the platform abstraction layer (`internal/platform/`) for OS-specific functionality. Use build tags (`//go:build windows`) when needed.

### Local Go SDK
The project uses a local replace directive for `github.com/modelcontextprotocol/go-sdk` (see `go.mod` line 46).

### Config File Watching
FR-050 requires monitoring client config files for external changes via ConfigFileWatcher.

## Testing Strategy

**Test Types:**
- **Unit tests:** Individual package functionality (`internal/*/..._test.go`)
- **Integration tests:** Service interactions (`tests/integration/`)
- **Contract tests:** API contract validation (`tests/contract/`)
- **Performance tests:** Benchmarks for startup time and memory usage (`tests/performance/`)
- **Frontend tests:** Component testing with Vitest + Testing Library (`frontend/src/test/`)

**Test Coverage:** All tests must pass before committing. Run verification scripts to ensure compliance.

## Adding New Features

### Adding a New API Endpoint
1. Create handler in `internal/api/` (e.g., `discovery.go`)
2. Wire route in `internal/api/router.go`
3. Add frontend client method in `frontend/src/services/api.ts`
4. Add contract test in `tests/contract/`
5. Test manually with `wails dev`

### Adding Platform-Specific Code
1. Define interface in `internal/platform/` (e.g., `platform.go`)
2. Implement per-platform:
   - `*_windows.go`
   - `*_darwin.go`
   - `*_linux.go`
3. Use build tags if needed: `//go:build windows`
4. Test on all target platforms

### Adding a New Service
1. Create service in `internal/core/`
2. Define interfaces for dependencies
3. Inject EventBus for event publishing
4. Update `app.go` to initialize service
5. Add corresponding API handlers
6. Add tests (unit + integration)

---

## CURRENT TASK: Fix Issue #1 - Explorer Directory Path Bug

**Branch:** `fix-explorer-directory-path`  
**GitHub Issue:** #1  
**Priority:** P2 (Medium)

### Problem Summary
The "Open Directory" button in the Explorer view opens the user's Documents directory for ALL servers instead of opening each server's actual installation directory.

**Expected Behavior:**
- Figma server should open: `%APPDATA%\Claude\Claude Extensions\ant.dir.ant.figma.figma\`
- Filesystem server should open: `%APPDATA%\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem\`
- Each server opens its own unique installation directory

**Actual Behavior:**
- All servers open: `C:\Users\hoyth\Documents\`

### Root Cause Analysis (from Architect)
The `OpenExplorer` method in `app.go` accepts a `path` parameter from the frontend and blindly passes it to `platform.OpenFileExplorer()`. The frontend is passing the Documents path instead of using `server.installationPath`.

**Key Finding:** The `MCPServer` struct already has the `InstallationPath` field properly defined in `internal/models/server.go`.

### Architectural Decision
Refactor `OpenExplorer` to accept `serverID` instead of `path` parameter. This provides:
- **Type Safety:** Frontend cannot pass wrong path
- **Single Source of Truth:** Server data lives in backend
- **Better Validation:** Backend verifies server exists before opening directory
- **Future-Proof:** Any UI implementation will work correctly

### Implementation Tasks

#### Task 1: Update Backend API (`app.go`)
**File:** `app.go` (around line 669)

**Current Signature:**
```go
func (a *App) OpenExplorer(path string) (*OpenExplorerResponse, error)
```

**New Implementation:**
```go
func (a *App) OpenExplorer(serverID string) (*OpenExplorerResponse, error) {
	slog.Info("OpenExplorer called", "serverId", serverID)

	// Validate serverID is not empty
	if serverID == "" {
		return &OpenExplorerResponse{
			Success: false,
			Message: "Server ID cannot be empty",
		}, nil
	}

	// Get server by ID from discovery service
	server, exists := a.discoveryService.GetServerByID(serverID)
	if !exists {
		return &OpenExplorerResponse{
			Success: false,
			Message: fmt.Sprintf("Server not found: %s", serverID),
		}, nil
	}

	// Validate InstallationPath exists and is not empty
	if server.InstallationPath == "" {
		return &OpenExplorerResponse{
			Success: false,
			Message: "Server installation path not available",
		}, nil
	}

	// Use server's installation path (not a parameter!)
	err := platform.OpenFileExplorer(server.InstallationPath)
	if err != nil {
		return &OpenExplorerResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to open explorer: %v", err),
		}, nil
	}

	return &OpenExplorerResponse{
		Success: true,
		Message: "File explorer opened successfully",
	}, nil
}
```

**Checklist:**
- [ ] Change method signature from `path string` to `serverID string`
- [ ] Add serverID validation (not empty)
- [ ] Add server lookup via `discoveryService.GetServerByID()`
- [ ] Add validation for server existence
- [ ] Add validation for non-empty `InstallationPath`
- [ ] Use `server.InstallationPath` instead of path parameter
- [ ] Update slog calls to use `serverId` field
- [ ] Ensure all error messages are specific and helpful

#### Task 2: Find Frontend Explorer Component
**Action:** Locate the component with "Open Directory" button

**Likely Locations:**
```bash
# Search strategies:
grep -r "OpenExplorer" frontend/src/
grep -r "Open Directory" frontend/src/
grep -r "open.*directory" frontend/src/ -i

# Likely files:
frontend/src/components/Explorer.svelte
frontend/src/views/Explorer.svelte
frontend/src/components/ServerCard.svelte
frontend/src/components/ServerTable.svelte
```

**What to Look For:**
- Button with text "Open Directory" or similar
- Click handler calling `OpenExplorer(...)`
- Any hardcoded paths (Documents, home directory, etc.)

#### Task 3: Update Frontend Call
**Current Pattern** (expected to find something like):
```javascript
// BAD - passing hardcoded or wrong path
import { OpenExplorer } from '../wailsjs/go/main/App';

function handleOpenDirectory() {
    OpenExplorer(documentsPath);  // or some hardcoded path
}
```

**New Implementation:**
```javascript
// GOOD - passing server ID
import { OpenExplorer } from '../wailsjs/go/main/App';

function handleOpenDirectory() {
    // Assuming 'server' object is in scope with 'id' field
    OpenExplorer(server.id);
}
```

**Checklist:**
- [ ] Find the component with "Open Directory" button
- [ ] Locate the `OpenExplorer` function call
- [ ] Change from `OpenExplorer(path)` to `OpenExplorer(server.id)`
- [ ] Verify `server.id` is available in component scope
- [ ] Remove any hardcoded path variables (Documents, home, etc.)
- [ ] Ensure TypeScript types are updated if needed

#### Task 4: Verify Discovery Populates InstallationPath (Optional)
**Purpose:** Confirm that discovery services are setting `InstallationPath` correctly

**Files to Check:**
```bash
internal/core/discovery/extensions.go  # Claude Extensions discovery
internal/core/discovery/filesystem.go  # Filesystem discovery
internal/core/discovery/clientconfig.go # Config-based discovery
```

**What to Verify:**
- Extension discovery sets full path like: `C:\Users\hoyth\AppData\Roaming\Claude\Claude Extensions\ant.dir.ant.figma.figma`
- Not just the extension name or partial path
- Path includes the full directory structure

**Add Debug Logging** (if needed):
```go
slog.Info("Server discovered",
    "name", server.Name,
    "installationPath", server.InstallationPath,
    "source", server.Source)
```

#### Task 5: Manual Testing
**Test Cases:**

1. **Figma Server:**
   - [ ] Click "Open Directory" button
   - [ ] Verify opens: `%APPDATA%\Claude\Claude Extensions\ant.dir.ant.figma.figma\`
   - [ ] NOT Documents folder

2. **Filesystem Server:**
   - [ ] Click "Open Directory" button
   - [ ] Verify opens: `%APPDATA%\Claude\Claude Extensions\ant.dir.ant.anthropic.filesystem\`
   - [ ] NOT Documents folder

3. **Different Servers Open Different Directories:**
   - [ ] Click "Open Directory" for Figma → correct path
   - [ ] Click "Open Directory" for Filesystem → different correct path
   - [ ] Each server opens its own unique directory

4. **Error Handling:**
   - [ ] Test with non-existent server ID → shows error message
   - [ ] Test with server missing InstallationPath → shows appropriate error

#### Task 6: Automated Testing
```bash
# Backend compile check
go build ./...

# Frontend compile check
cd frontend && npm run check

# Run full verification
scripts\verify-build.bat

# Or quick verification during development
scripts\verify-build.bat --quick --skip-build
```

**Checklist:**
- [ ] No Go compilation errors
- [ ] No TypeScript errors
- [ ] All existing tests pass
- [ ] Consider adding a contract test for OpenExplorer API

### Acceptance Criteria
✅ **Primary Requirements:**
- Figma server opens correct extension directory (not Documents)
- Filesystem server opens correct extension directory (not Documents)
- Each server opens its own unique directory
- No hardcoded Documents paths remain in code

✅ **Error Handling:**
- Empty serverID returns appropriate error
- Non-existent server returns "Server not found" error
- Missing InstallationPath returns "path not available" error

✅ **Quality Gates:**
- All verification tests pass (`scripts\verify-build.bat`)
- No regression in existing functionality
- Code follows project standards

### Commit Message Template
```
fix: Explorer opens correct server installation directory

- Refactored OpenExplorer to accept serverID instead of path
- Backend now uses server.InstallationPath from discovery service
- Frontend passes server.id instead of hardcoded Documents path
- Added validation for server existence and path availability
- Improved error messages for better debugging

Fixes #1
```

### Notes from Architect (Claude Web)
- The `InstallationPath` field already exists in `MCPServer` struct - no model changes needed
- This is purely a data flow issue: frontend → backend → platform layer
- The fix improves architectural integrity by enforcing single source of truth
- Discovery services should already be populating `InstallationPath` correctly, but verify if issues persist

---

## Coding Standards

This project adheres to the [Positronikal Coding Standards](https://github.com/positronikal/coding-standards/tree/main/standards).

## Directory Structure Notes

This project follows **Go and Wails conventions** for better tooling support:

**Go/Wails Conventions:**
- `cmd/`: Application entry points
- `internal/`: Private application code
- `pkg/`: Public library code
- `tests/`: Test files (note plural, Go convention)
- `frontend/`: Svelte frontend (Wails requirement at root)
- `build/`: Wails build output (single source of truth for executables)
- Root `.go` files: `app.go`, `main.go` (Wails bindings)

**Positronikal Standards Maintained:**
- `docs/`: Development documentation
- `etc/`: Scratch workspace
- `rel/`: Release packages (production installers)
- `ref/`: Reference materials

**Special Directories:**
- `.specify/`: Spec Kit framework
- `specs/`: Feature specifications
- `.claude/`: Claude Code configuration
- `scripts/`: Build and verification scripts

## Additional Resources

- **Documentation:** `docs/` directory
- **Feature Specifications:** `specs/` directory
- **Bug Reporting:** `BUGS.md`
- **Security:** `SECURITY.md`
- **Contributing:** `CONTRIBUTING.md`
- **Installation & Usage:** `USING.md`
- **Packaging:** `docs/PACKAGING.md`
- **Wails Documentation:** https://wails.io/docs/introduction
