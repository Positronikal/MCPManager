# Research & Technology Decisions

**Feature**: MCP Manager - Cross-Platform Server Management Application
**Date**: 2025-10-15
**Status**: Phase 0 Complete

## Research Overview

This document captures research findings and technology decisions for implementing MCP Manager following constitutional principles and technical requirements.

---

## 1. Cross-Platform GUI Framework Selection

### Decision: **Wails + Svelte**

### Rationale:
- **API-First Architecture**: Go backend with web frontend maintains separation of concerns (constitutional compliance)
- **Toolchain Alignment**: Wails and Svelte already installed and preferred for this project
- **Development Workflow**: Enables Figma → Google Stitch → Svelte component pipeline
- **Svelte Compiler**: Produces highly optimized, minimal JavaScript (meets <100MB memory constraint)
- **Native Performance**: Wails uses WebView2 (Windows), WebKit (macOS/Linux) - no Chromium overhead
- **Fast Startup**: Meets <2s startup requirement (FR-037)
- **Modern UI**: Svelte's reactive paradigm simplifies real-time status updates (FR-005, FR-047)
- **Constitutional Compliance**: Backend remains pure Go; UI is decoupled consumer of backend API

### Alternatives Considered:

| Framework | Pros | Cons | Why Not Selected |
|-----------|------|------|------------------|
| **Fyne** | Pure Go, no web stack, lightweight | Manual UI construction, less design flexibility, no Figma workflow integration | Doesn't leverage existing Svelte/Stitch toolchain |
| **Gio** | Pure Go, immediate mode, very lightweight | Steeper learning curve, less mature ecosystem, manual UI construction | Higher development effort; no design tool integration |
| **Qt (Go bindings)** | Mature, feature-rich, professional appearance | C++ dependency (violates constitution), large binary size | Requires C++ (constitutional violation for non-critical path) |
| **Electron (Tauri)** | Web technologies, rich ecosystem | Non-Go, massive binary size (>100MB baseline), slow startup | Electron violates memory constraints; Tauri uses Rust (non-Go backend) |

### Implementation Notes:
- **Project Initialization**: `wails init -n mcpmanager -t svelte` (NOT `npx sv create`)
  - Wails provides: Svelte 4.x compiler + Vite bundler (NOT SvelteKit framework)
  - No SvelteKit: Desktop app doesn't need SSR, file-based routing, or Node.js server runtime
  - Wails template includes: Svelte + Vite + TypeScript + Go bindings generation
- Use Wails v2.x with Go 1.21+ backend
- Implement Svelte 4.x frontend with TypeScript for type safety
- Backend exposes REST API + SSE stream (per api-spec.yaml) at localhost:8080
- Frontend can call Go directly via Wails bindings: `import { StartServer } from '../wailsjs/go/main/App'`
- Alternative: Frontend can call REST API (localhost:8080) for testing independence
- Dark theme per FR-043 using Svelte theming system (CSS variables or Tailwind)
- Leverage Svelte stores for reactive UI updates (FR-005, FR-047)
- CSS Grid/Flexbox for responsive layouts (FR-045)
- **Development Workflow**: `wails dev` (starts Go backend + Vite dev server with hot reload)
- **Production Build**: `wails build` → Single executable with embedded Svelte UI (no Node.js required at runtime)
- **Platform Builds**: `wails build -platform windows/amd64,darwin/universal,linux/amd64`

---

## 2. MCP Protocol Integration

### Decision: **Use local MCP Go SDK at `D:\dev\ARTIFICIAL_INTELLIGENCE\MCP\_MCP-Tools-Dev\go-sdk\`**

### Rationale:
- **Constitutional requirement**: "Rely on existing MCP libraries" - don't implement protocol
- **Local SDK available**: Fully up-to-date Go SDK already present in development environment
- **Minimal coupling**: MCP Manager monitors/controls servers, doesn't need full client implementation
- **Read-only server introspection**: Query server capabilities/tools via MCP protocol
- **Process management primary**: Focus on lifecycle control (start/stop/restart) rather than protocol communication

### Alternatives Considered:

| Approach | Pros | Cons | Why Not Selected |
|----------|------|------|------------------|
| **Full MCP client library** | Complete protocol support, future-proof | Overengineered for monitoring use case | Violates "do one thing well" - we manage servers, not act as clients |
| **Protocol reimplementation** | Full control, no dependencies | Constitutional violation, maintenance burden | Explicitly forbidden by constitution |
| **No MCP integration** | Simplest approach | Cannot query server capabilities/tools | Doesn't meet FR-003 (display capabilities) |

### Implementation Notes:
- Reference local SDK in `go.mod`: `replace github.com/modelcontextprotocol/go-sdk => ../../../_MCP-Tools-Dev/go-sdk`
- Use SDK for server introspection (capabilities, tools metadata)
- Focus on server metadata discovery, not full client operations
- Keep MCP integration isolated in `pkg/mcpclient/` wrapper

---

## 3. Server Discovery Strategy

### Decision: **Multi-source discovery with platform-specific configuration paths**

### Rationale:
- **FR-001**: Scan common installation locations (NPM global, Python site-packages, Go bin)
- **FR-002**: Read MCP client configuration files (Claude Desktop, Cursor, Zed, etc.)
- **FR-050**: Monitor client config files for external changes

### Discovery Sources (Priority Order):

1. **MCP Client Configuration Files** (Primary):
   - Windows: `%APPDATA%/Claude/claude_desktop_config.json`, `%APPDATA%/Cursor/mcp_config.json`
   - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`, `~/Library/Application Support/Cursor/mcp_config.json`
   - Linux: `~/.config/Claude/claude_desktop_config.json`, `~/.config/Cursor/mcp_config.json`

2. **Standard Installation Paths** (Secondary):
   - NPM global: `npm root -g` + known MCP server packages
   - Python: `site-packages` + MCP server pattern matching
   - Go binaries: `$GOPATH/bin` or `~/go/bin`
   - Custom paths: User-defined directories

3. **Process Discovery** (Runtime):
   - Scan running processes for known MCP server executables
   - Match PIDs to discovered servers for status tracking

### Implementation Notes:
- Use `fsnotify` library for file system watching (FR-050)
- Parse JSON config files read-only (FR-019)
- Cache discovery results with manual refresh (FR-006)
- Implement platform-specific path resolution in `internal/platform/`

---

## 4. Process Lifecycle Management

### Decision: **os/exec with PID tracking and graceful shutdown signals**

### Rationale:
- **Native Go support**: `os/exec` package provides cross-platform process management
- **PID tracking**: Required by FR-010 for lifecycle control
- **Signal handling**: Platform-appropriate termination (SIGTERM on Unix, taskkill on Windows)
- **Non-blocking**: Launch processes without blocking UI thread (FR-038)

### Process States (per clarification):
- **stopped** (red): Not running
- **starting** (blue/gray): Launch initiated, awaiting confirmation
- **running** (green): Process active with valid PID
- **error** (yellow): Startup failed or unexpected termination

### Implementation Notes:
- Use `exec.CommandContext` for cancellable operations
- Monitor process via goroutine, update state on exit
- Capture stdout/stderr to log buffer (1000 entries per server, FR-053)
- Implement graceful shutdown timeout (SIGTERM → SIGKILL fallback)
- Handle platform differences in `internal/platform/process*.go`

---

## 5. State Persistence

### Decision: **JSON files with atomic writes**

### Rationale:
- **Constitutional requirement**: File-based storage, no database needed
- **Simplicity**: Easy debugging, human-readable, version control friendly
- **Atomic writes**: Write to temp file + rename prevents corruption
- **Fast startup**: Plain file reads faster than database connection overhead

### State Files:

```
~/.mcpmanager/
├── state.json           # Application state (FR-041)
│   ├── discovered_servers[]
│   ├── user_preferences{}
│   ├── window_layout{}
│   └── last_discovery_timestamp
└── logs/
    ├── server_{id}_logs.json  # Per-server log history (FR-053)
    └── app.log                # Application-level logs
```

### Implementation Notes:
- Use `encoding/json` standard library
- Implement `storage.StateManager` interface for future backend swapping
- Atomic writes: `ioutil.TempFile` → `os.Rename`
- Auto-save on state changes with debouncing (max 1 write/sec)

---

## 6. Real-Time Updates Architecture

### Decision: **Event-driven architecture with pub/sub pattern**

### Rationale:
- **FR-005, FR-047**: Real-time status updates without manual refresh
- **FR-040**: Avoid constant polling that degrades performance
- **Reactive UI**: Event-driven updates trigger UI re-renders efficiently
- **Decoupled**: Backend publishes events, UI subscribes (API-first architecture)

### Event Types:
- `ServerDiscovered`
- `ServerStatusChanged` (stopped → starting → running/error)
- `ServerLogEntry` (new log line with severity)
- `ConfigFileChanged` (external modification detected, FR-050)
- `ServerMetricsUpdated` (memory, uptime, request count)

### Implementation Notes:
- Implement lightweight pub/sub in `internal/core/events/`
- Use Go channels for event delivery to UI
- Rate-limit metrics updates (1Hz for UI responsiveness)
- Buffer events during UI updates to prevent blocking

---

## 7. Configuration Management

### Decision: **Structured config editor with validation**

### Rationale:
- **FR-014 through FR-018**: View/edit server configs, env vars, CLI args with validation
- **FR-019**: NEVER modify client configs (Claude Desktop, Cursor, etc.)
- **Validation before apply**: Prevent invalid configs (edge case requirement)

### Configuration Scope:
- **Server-Specific** (MCP Manager managed):
  - Environment variables
  - Command-line arguments
  - Custom launch parameters
  - Working directory
- **Client Configurations** (Read-Only):
  - Claude Desktop config (display only, no writes)
  - Other MCP client configs (observability, no modification)

### Implementation Notes:
- Store MCP Manager-specific configs in `~/.mcpmanager/servers/{id}/config.json`
- Implement JSON schema validation before applying changes
- Provide config templates for common server types
- Clear visual distinction: "MCP Manager Settings" vs "Client Configuration (Read-Only)"

---

## 8. Cross-Platform Considerations

### Decision: **Platform abstraction layer with conditional compilation**

### Rationale:
- **FR-035, FR-036**: Windows, macOS, Linux support with platform-appropriate conventions
- **Maintainability**: Isolate platform-specific code for testability
- **Go build tags**: Compile-time platform selection

### Platform-Specific Concerns:

| Concern | Windows | macOS | Linux |
|---------|---------|-------|-------|
| **Config Paths** | `%APPDATA%` | `~/Library/Application Support` | `~/.config` |
| **Process Signals** | `taskkill` | SIGTERM/SIGKILL | SIGTERM/SIGKILL |
| **File Watching** | Windows API | FSEvents | inotify |
| **Single Instance** | Mutex | Mach ports | Unix socket lock |

### Implementation Notes:
- Use build tags: `//go:build windows`, `//go:build darwin`, `//go:build linux`
- Implement interfaces in `internal/platform/`:
  - `PathResolver` - platform-appropriate config locations
  - `ProcessManager` - platform-specific process control
  - `FileWatcher` - native file system monitoring
  - `SingleInstance` - prevent duplicate app launches (FR-051)

---

## 9. Logging & Monitoring

### Decision: **Structured logging with severity levels and circular buffer**

### Rationale:
- **FR-020 through FR-026**: Real-time log viewer, color-coded severity, filtering
- **FR-053**: 1000 log entries per server (rolling retention)
- **Memory management**: Circular buffer prevents unbounded growth

### Log Levels (per FR-021):
- **INFO** (blue): Normal operations
- **SUCCESS** (green): Successful operations
- **WARN** (yellow): Non-fatal issues
- **ERROR** (red): Failures requiring attention

### Implementation Notes:
- Use Go's `log/slog` (structured logging, Go 1.21+)
- Implement circular buffer in `internal/core/monitoring/logbuffer.go`
- Capture server stdout/stderr via `io.Pipe`
- Parse log lines for severity (keywords: "error", "warn", "success", etc.)
- Filtering: server name + severity (FR-022, FR-023)

---

## 10. Performance Optimization Strategy

### Decision: **Lazy loading + goroutine pooling + efficient data structures**

### Rationale:
- **FR-037**: <2s startup time
- **FR-038**: Non-blocking UI operations
- **FR-039**: <100MB memory at idle
- **FR-054**: Efficiently handle 50 servers simultaneously

### Optimization Techniques:

1. **Lazy Discovery**: Background goroutine for server scanning, show UI immediately
2. **Goroutine Pool**: Limit concurrent operations (max 10 simultaneous server starts)
3. **Efficient Buffering**: Ring buffer for logs (O(1) insertion, bounded memory)
4. **Debounced UI Updates**: Batch status changes (max 60 FPS UI refresh)
5. **Minimal Memory**: Limit log retention (1000 * 50 servers = 50k entries max)

### Performance Budget:
- **Startup**: <2s (FR-037)
- **Memory Idle**: <100MB (FR-039)
- **Memory 50 Servers**: <300MB (50 servers * ~4MB each + 100MB base)
- **UI Response**: <200ms for button clicks (FR-038)
- **Log Filtering**: <10ms for 50k entries

### Implementation Notes:
- Profile with `pprof` to identify bottlenecks
- Use `sync.Pool` for frequently allocated objects
- Implement backpressure for event streams
- Cache server metadata to avoid repeated MCP queries

---

## 11. Testing Strategy

### Decision: **Multi-layer testing: unit + contract + integration**

### Rationale:
- **Constitutional requirement**: Unit tests for business logic, integration tests for lifecycle
- **TDD approach**: Tests before implementation (enforced by /tasks workflow)
- **Contract tests**: Ensure API stability for UI-backend decoupling

### Test Layers:

1. **Unit Tests** (`tests/unit/`):
   - Domain models (validation, state transitions)
   - Business logic (discovery, lifecycle, config management)
   - Platform abstractions (mock external dependencies)
   - Coverage target: >80% for `internal/core/`

2. **Contract Tests** (`tests/contract/`):
   - API endpoint schemas (request/response validation)
   - Event payload structures
   - State persistence format
   - Fail initially (no implementation yet)

3. **Integration Tests** (`tests/integration/`):
   - End-to-end server lifecycle (discover → start → monitor → stop)
   - File system interactions (config reading, state persistence)
   - Process management (launch server, track PID, terminate)
   - Cross-platform compatibility (CI for Windows, macOS, Linux)

### Implementation Notes:
- Use Go's `testing` package + `testify` for assertions
- Mock external dependencies: file system, process execution, MCP server responses
- CI/CD: GitHub Actions with matrix builds (OS * Go version)
- Quickstart test: Automate acceptance scenarios from spec.md

---

## 12. Dependency Management & Utilities

### Decision: **Minimal dependencies with clear justification**

### Rationale:
- **Constitutional preference**: Minimize external dependencies, prefer standard library
- **FR-027, FR-028**: Check server prerequisites, actionable error messages
- **FR-029**: Check for server updates

### Approved Dependencies:

| Dependency | Justification | Alternatives Considered |
|------------|---------------|-------------------------|
| **github.com/wailsapp/wails/v2** | Cross-platform GUI framework (constitutional requirement) | Fyne, Gio, Qt (see Research §1) |
| **fsnotify/fsnotify** | File system watching for FR-050 | Manual polling (violates FR-040) |
| **MCP SDK** | Constitutional requirement (use existing libraries) | DIY implementation (forbidden) |

### Utility Implementation:

- **FR-030 (Netstat)**: Parse `netstat` output via `os/exec` (no library needed)
- **FR-031 (Shell)**: Launch platform shell via `os/exec` (cmd.exe, Terminal.app, xterm)
- **FR-032 (Explorer)**: `os/exec` with `explorer`, `open`, `xdg-open`
- **FR-033 (Services)**: Platform-specific service commands (sc, launchctl, systemctl)
- **FR-034 (Help)**: Embedded markdown + Svelte markdown component

### Dependency Checking:
- Detect Node.js for NPM-based servers (`node --version`)
- Detect Python for Python-based servers (`python --version`)
- Detect Go for Go-based servers (`go version`)
- Clear error messages with installation instructions (FR-028)

---

## 13. Security Considerations

### Decision: **Principle of least privilege + no credential handling**

### Rationale:
- **Constitutional requirement**: Secure config data, no hardcoded secrets
- **Scope boundary**: MCP Manager monitors servers, doesn't authenticate users
- **Server responsibility**: Servers handle their own auth (MCP Manager just starts/stops them)

### Security Measures:

1. **No Credential Storage**: MCP Manager never stores API keys, passwords, tokens
2. **Read-Only Client Configs**: FR-019 prevents accidental secret modification
3. **File Permissions**: Config files readable only by user (`0600` on Unix, ACLs on Windows)
4. **Process Isolation**: Each server runs as separate process (inherit user permissions)
5. **Input Validation**: Sanitize user-provided paths, commands, environment variables
6. **No Remote Execution**: Only manage local servers (no network server control)

### Implementation Notes:
- Validate file paths to prevent directory traversal
- Sanitize environment variables (no injection attacks)
- Warn users if server configs contain sensitive data in logs
- Document security model: "MCP Manager has same permissions as user account"

---

## 14. Keyboard Shortcuts

### Decision: **Standard shortcuts with platform awareness**

### Rationale:
- **FR-046**: Keyboard shortcuts for Start, Stop, Refresh
- **Accessibility**: Power users prefer keyboard navigation
- **Platform conventions**: Cmd on macOS, Ctrl on Windows/Linux

### Shortcut Mapping:

| Action | Windows/Linux | macOS | Notes |
|--------|---------------|-------|-------|
| **Start Server** | Ctrl+S | Cmd+S | Conflicts with Save (context: no unsaved data) |
| **Stop Server** | Ctrl+T | Cmd+T | T for Terminate |
| **Restart Server** | Ctrl+R | Cmd+R | R for Restart |
| **Refresh Discovery** | F5 | Cmd+R | Standard refresh |
| **Focus Search** | Ctrl+F | Cmd+F | Filter servers |
| **Toggle Logs** | Ctrl+L | Cmd+L | Show/hide log panel |

### Implementation Notes:
- Implement keyboard shortcuts via Svelte `on:keydown` handlers with platform detection
- Visual indicators: Show shortcuts in tooltip/menu
- Conflict resolution: Disable shortcuts when text fields focused

---

## Research Completion Checklist

- [x] GUI framework selected (Wails + Svelte)
- [x] MCP protocol integration strategy defined
- [x] Server discovery approach documented
- [x] Process lifecycle management planned
- [x] State persistence mechanism chosen
- [x] Real-time updates architecture designed
- [x] Configuration management scoped
- [x] Cross-platform considerations addressed
- [x] Logging & monitoring approach defined
- [x] Performance optimization strategy established
- [x] Testing strategy documented
- [x] Dependencies justified and minimized
- [x] Security considerations evaluated
- [x] Keyboard shortcuts defined

**No NEEDS CLARIFICATION items remain. Proceeding to Phase 1 (Design & Contracts).**

---

---

## 15. System Architecture

### Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        MCP Manager Application                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────────────────────────────────────────┐   │
│  │              Frontend (Svelte UI)                       │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │   │
│  │  │ Server Table │  │  Log Viewer  │  │Config Editor│ │   │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬──────┘ │   │
│  │         │                  │                  │         │   │
│  │         └──────────────────┼──────────────────┘         │   │
│  │                            │                            │   │
│  │                    ┌───────▼────────┐                  │   │
│  │                    │  API Service   │                  │   │
│  │                    │  (REST/SSE)    │                  │   │
│  │                    └───────┬────────┘                  │   │
│  └────────────────────────────┼───────────────────────────┘   │
│                                │ Wails Bindings                │
│  ══════════════════════════════╪═══════════════════════════   │
│                                │                                │
│  ┌────────────────────────────▼───────────────────────────┐   │
│  │         Backend API (Go)                               │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │   │
│  │  │  Discovery   │  │  Lifecycle   │  │ Config Mgmt │ │   │
│  │  │   Service    │  │   Service    │  │   Service   │ │   │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬──────┘ │   │
│  │         │                  │                  │         │   │
│  │  ┌──────▼──────────────────▼──────────────────▼──────┐ │   │
│  │  │           Event Bus (Pub/Sub)                     │ │   │
│  │  └──────┬────────────────────────────────────────────┘ │   │
│  │         │                                               │   │
│  │  ┌──────▼───────┐  ┌──────────────┐  ┌─────────────┐ │   │
│  │  │  Monitoring  │  │   Storage    │  │  Platform   │ │   │
│  │  │   Service    │  │   Service    │  │ Abstraction │ │   │
│  │  └──────────────┘  └──────────────┘  └─────────────┘ │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                │                                │
│  ══════════════════════════════╪═══════════════════════════   │
│                                │                                │
│  ┌────────────────────────────▼───────────────────────────┐   │
│  │              External Integrations                      │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐ │   │
│  │  │  MCP Servers │  │ Client Configs│ │ File System │ │   │
│  │  │  (Processes) │  │(Claude/Cursor)│ │  (fsnotify) │ │   │
│  │  └──────────────┘  └──────────────┘  └─────────────┘ │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### Layer Responsibilities:

**Frontend (Svelte)**:
- User interaction handling
- Real-time UI updates via SSE
- Component-based UI architecture
- State management with Svelte stores

**Backend API (Go)**:
- REST endpoints for CRUD operations
- SSE stream for real-time events
- Business logic execution
- State orchestration

**Core Services**:
- **Discovery**: Scan filesystem, parse configs, detect running processes
- **Lifecycle**: Start/stop/restart servers, PID tracking, process monitoring
- **Config Mgmt**: CRUD for server configurations, validation
- **Monitoring**: Log capture, circular buffer, metrics collection
- **Storage**: JSON persistence, atomic writes, state management
- **Platform**: OS-specific abstractions (paths, process control, file watching)

**Event Bus**:
- Decouple services via pub/sub
- Enable real-time UI updates
- Prevent circular dependencies

---

## 16. Server Discovery Flow

### Sequence Diagram

```
┌──────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  UI  │     │   API    │     │Discovery │     │  Config  │     │   File   │
│      │     │ Service  │     │ Service  │     │  Parser  │     │ Watcher  │
└───┬──┘     └────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
    │             │                 │                 │                 │
    │ User clicks │                 │                 │                 │
    │  "Refresh"  │                 │                 │                 │
    │─────────────>                 │                 │                 │
    │             │                 │                 │                 │
    │             │ POST /discover  │                 │                 │
    │             │─────────────────>                 │                 │
    │             │                 │                 │                 │
    │             │  202 Accepted   │                 │                 │
    │<────────────│<─────────────────│                 │                 │
    │             │                 │                 │                 │
    │             │                 │ Scan Client     │                 │
    │             │                 │ Config Paths    │                 │
    │             │                 │─────────────────>                 │
    │             │                 │                 │                 │
    │             │                 │                 │ Read JSON files │
    │             │                 │                 │─────────────────>
    │             │                 │                 │                 │
    │             │                 │                 │  Parsed Servers │
    │             │                 │<─────────────────────────────────│
    │             │                 │                 │                 │
    │             │                 │ For each server:│                 │
    │             │                 │ - Create MCPServer entity         │
    │             │                 │ - Set source=client_config        │
    │             │                 │ - Check if process running        │
    │             │                 │                 │                 │
    │             │                 │ Scan Filesystem │                 │
    │             │                 │ (npm, pip, go)  │                 │
    │             │                 │─────────────────>                 │
    │             │                 │                 │                 │
    │             │                 │  Found Servers  │                 │
    │             │                 │<─────────────────                 │
    │             │                 │                 │                 │
    │             │                 │ For each found: │                 │
    │             │                 │ - Create MCPServer entity         │
    │             │                 │ - Set source=filesystem           │
    │             │                 │                 │                 │
    │             │                 │ Scan Running    │                 │
    │             │                 │ Processes       │                 │
    │             │                 │─────────────────>                 │
    │             │                 │                 │                 │
    │             │                 │  Active PIDs    │                 │
    │             │                 │<─────────────────                 │
    │             │                 │                 │                 │
    │             │                 │ Match PIDs to   │                 │
    │             │                 │ discovered      │                 │
    │             │                 │ servers         │                 │
    │             │                 │                 │                 │
    │             │ ServerDiscovered│                 │                 │
    │             │  events (SSE)   │                 │                 │
    │<──────────────────────────────│                 │                 │
    │             │                 │                 │                 │
    │ UI updates  │                 │                 │                 │
    │ server table│                 │                 │                 │
    │             │                 │                 │                 │
    │             │                 │ Persist State   │                 │
    │             │                 │ (state.json)    │                 │
    │             │                 │─────────────────>                 │
    │             │                 │                 │                 │
    │             │                 │ Setup File Watch│                 │
    │             │                 │ on Client Configs                 │
    │             │                 │─────────────────────────────────> │
    │             │                 │                 │                 │
    │             │                 │                 │ Watch Active    │
    │             │                 │                 │ (FR-050)        │
    │             │                 │                 │                 │
```

### Discovery Sources Priority:

1. **Client Configuration Files** (Primary):
   - Highest reliability (explicit user configuration)
   - Includes command + args + environment
   - Monitored for external changes (FR-050)

2. **Filesystem Scanning** (Secondary):
   - Finds installed servers not yet configured
   - Scans standard package manager locations
   - Lower confidence (needs validation)

3. **Process Discovery** (Runtime):
   - Detects already-running servers
   - Associates PIDs with discovered servers
   - Enables status tracking for externally-launched servers

### File Watching Strategy:

Once discovery completes, file watchers are established on:
- `~/.config/Claude/claude_desktop_config.json`
- `~/.config/Cursor/mcp_config.json`
- Other monitored client config paths

On file modification:
1. Event triggers `ConfigFileChanged` notification
2. UI shows "Configuration changed externally. Refresh?"
3. User clicks notification → triggers new discovery scan
4. Diff detected: new servers added, removed servers marked inactive

---

## References

- [Wails Documentation](https://wails.io/docs/introduction)
- [Svelte Documentation](https://svelte.dev/docs)
- [MCP Specification](https://modelcontextprotocol.io/)
- [MCP Go SDK](file://D:/dev/ARTIFICIAL_INTELLIGENCE/MCP/_MCP-Tools-Dev/go-sdk/)
- [Go os/exec Package](https://pkg.go.dev/os/exec)
- [fsnotify Library](https://github.com/fsnotify/fsnotify)
- Project Constitution: `.specify/memory/constitution.md`
- Feature Specification: `specs/001-mcp-manager-specification/spec.md`
