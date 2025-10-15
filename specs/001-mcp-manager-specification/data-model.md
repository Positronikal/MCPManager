# Data Model

**Feature**: MCP Manager - Cross-Platform Server Management Application
**Date**: 2025-10-15
**Phase**: 1 - Design & Contracts

---

## Overview

This document defines the domain entities, their attributes, relationships, validation rules, and state transitions for MCP Manager. All entities are derived from the feature specification's Key Entities section and clarification decisions.

---

## Entity Definitions

### 1. MCPServer

Represents an installed Model Context Protocol server discovered by the system.

#### Attributes

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `id` | string (UUID) | Yes | Valid UUID v4 | Unique identifier for the server |
| `name` | string | Yes | 1-100 chars, alphanumeric + spaces | Human-readable server name |
| `version` | string | No | Semver format (e.g., "1.2.3") | Server version if detectable |
| `installationPath` | string | Yes | Valid absolute path | File system path to server executable/module |
| `status` | ServerStatus enum | Yes | One of: stopped, starting, running, error | Current operational state |
| `pid` | int | No | >0 when running | Process ID when server is active |
| `capabilities` | []string | No | Array of capability names | MCP capabilities exposed by server (e.g., ["tools", "prompts"]) |
| `tools` | []string | No | Array of tool names | MCP tools provided by server |
| `configuration` | ServerConfiguration | Yes | Valid config object | Server-specific configuration |
| `dependencies` | []Dependency | No | Array of dependency objects | Required prerequisites for server |
| `discoveredAt` | timestamp | Yes | ISO 8601 | When server was first discovered |
| `lastSeenAt` | timestamp | Yes | ISO 8601 | Last time server was detected in discovery |
| `source` | DiscoverySource enum | Yes | One of: client_config, filesystem, process | How server was discovered |

#### Relationships
- **1:1** with `ServerStatus` (embedded)
- **1:1** with `ServerConfiguration` (embedded)
- **1:N** with `Dependency` (composition)
- **1:N** with `LogEntry` (via server ID reference)

#### Validation Rules
- `name` must be unique within the system
- `pid` must be null when `status` is `stopped` or `error`
- `pid` must be non-null when `status` is `running`
- `installationPath` must exist on file system at discovery time
- `lastSeenAt` >= `discoveredAt`

---

### 2. ServerStatus

Represents the real-time operational state of a server.

#### Attributes

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `state` | StatusState enum | Yes | One of: stopped, starting, running, error | Current lifecycle state |
| `uptimeSeconds` | int | No | >=0, null when not running | Seconds since server started |
| `memoryUsageMB` | float | No | >=0, null when not running | Current memory consumption in MB |
| `requestCount` | int | No | >=0, null if unavailable | Total requests handled (if server exposes metric) |
| `lastStateChange` | timestamp | Yes | ISO 8601 | When status last transitioned |
| `errorMessage` | string | No | Max 500 chars | Error details when state is `error` |
| `startupAttempts` | int | Yes | >=0 | Number of start attempts since last successful run |

#### State Transition Rules

```
[stopped] ──(start)──> [starting] ──(success)──> [running]
                            │
                            └──(failure)──> [error]

[running] ──(stop)────> [stopped]
          ──(crash)───> [error]

[error] ────(start)───> [starting]
        ────(reset)───> [stopped]
```

#### Validation Rules
- `uptimeSeconds`, `memoryUsageMB`, `requestCount` must be null when `state` is `stopped` or `error`
- `errorMessage` required when `state` is `error`, null otherwise
- `startupAttempts` resets to 0 on successful transition to `running`
- `startupAttempts` increments on transition from `starting` to `error`

#### Color Mapping (per FR-004)
- `stopped` → Red
- `starting` → Blue/Gray
- `running` → Green
- `error` → Yellow

---

### 3. ServerConfiguration

Represents editable configuration data for a server managed by MCP Manager.

#### Attributes

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `environmentVariables` | map[string]string | No | Keys: valid env var names; Values: any string | Environment variables passed to server process |
| `commandLineArguments` | []string | No | Array of strings | CLI arguments appended to server command |
| `workingDirectory` | string | No | Valid absolute path | Working directory for server process |
| `autoStart` | bool | Yes | Default: false | Whether to start server automatically on app launch |
| `restartOnCrash` | bool | Yes | Default: false | Whether to auto-restart server on unexpected termination |
| `maxRestartAttempts` | int | Yes | 0-10, default: 3 | Max restart attempts before giving up |
| `configFilePath` | string | No | Valid absolute path | Path to server-specific config file (if applicable) |
| `customCommand` | string | No | Max 500 chars | Override default server launch command |

#### Validation Rules
- `environmentVariables` keys must match regex: `^[A-Z_][A-Z0-9_]*$`
- `workingDirectory` must exist on file system if provided
- `configFilePath` must exist on file system if provided
- `maxRestartAttempts` only applies when `restartOnCrash` is true
- `customCommand` if provided, must be valid executable path or shell command

#### Read-Only vs Editable
- **Editable by MCP Manager**: All fields in this entity (managed by MCP Manager)
- **Read-Only**: MCP client configuration files (e.g., Claude Desktop config) - displayed but never modified (FR-019)

---

### 4. LogEntry

Represents a single log message from a server or the application.

#### Attributes

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `id` | string (UUID) | Yes | Valid UUID v4 | Unique log entry identifier |
| `timestamp` | timestamp | Yes | ISO 8601 | When log message was generated |
| `severity` | LogSeverity enum | Yes | One of: info, success, warning, error | Log level |
| `source` | string | Yes | Server ID or "mcpmanager" | Which server generated the log |
| `message` | string | Yes | Max 10,000 chars | Log message content |
| `metadata` | map[string]any | No | JSON-serializable object | Additional structured data (stack trace, request ID, etc.) |

#### Validation Rules
- `message` content is plain text or JSON-formatted (parsed for severity if no explicit level)
- `source` must reference valid server ID or be "mcpmanager" for app-level logs
- Per-server log retention: max 1000 entries (FR-053) - oldest entries auto-deleted

#### Severity Color Mapping (per FR-021)
- `info` → Blue
- `success` → Green
- `warning` → Yellow
- `error` → Red

#### Log Retention Policy (per Clarification)
- **Circular buffer**: 1000 entries per server
- **Rolling deletion**: Oldest entry deleted when 1001st entry added
- **Memory bound**: ~50k entries total (50 servers * 1000 entries)
- **Persistence**: Logs saved to `~/.mcpmanager/logs/server_{id}_logs.json`

---

### 5. Dependency

Represents a required prerequisite for a server to run successfully.

#### Attributes

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `id` | string (UUID) | Yes | Valid UUID v4 | Unique dependency identifier |
| `name` | string | Yes | 1-100 chars | Dependency name (e.g., "Node.js", "Python 3.11") |
| `type` | DependencyType enum | Yes | One of: runtime, library, tool, environment | Dependency category |
| `requiredVersion` | string | No | Version string or range (e.g., ">=3.11", "^1.2.0") | Required version constraint |
| `detectedVersion` | string | No | Version string | Currently installed version if detected |
| `installed` | bool | Yes | Computed: detectedVersion != null && matches requiredVersion | Whether dependency is satisfied |
| `installationInstructions` | string | No | Max 1000 chars, markdown | How to install if missing (FR-028) |
| `checkCommand` | string | No | Shell command | Command to check if dependency exists (e.g., "node --version") |

#### Validation Rules
- `requiredVersion` if provided, must be valid semver or version range
- `detectedVersion` if provided, must be valid semver
- `installed` computed field: true if `detectedVersion` satisfies `requiredVersion` constraint
- `installationInstructions` should be actionable (URLs, commands, package manager instructions)

#### Dependency Types
- **runtime**: Language runtime (Node.js, Python, Go)
- **library**: System library (libssl, glibc)
- **tool**: External tool (npm, pip, gcc)
- **environment**: Environment setup (PATH variable, API keys)

---

### 6. ApplicationState

Represents persistent application data that survives restarts.

#### Attributes

| Field | Type | Required | Validation | Description |
|-------|------|----------|------------|-------------|
| `discoveredServers` | []string (Server IDs) | Yes | Array of valid UUIDs | List of all discovered server IDs |
| `userPreferences` | UserPreferences | Yes | Valid preferences object | User settings |
| `windowLayout` | WindowLayout | Yes | Valid layout object | Window size/position |
| `selectedFilters` | Filters | Yes | Valid filter object | Active log/server filters |
| `lastDiscoveryTimestamp` | timestamp | Yes | ISO 8601 | Last time discovery scan ran |
| `monitoredConfigPaths` | []string | Yes | Array of valid paths | Client config files being monitored (FR-050) |

#### Sub-Entities

**UserPreferences**:
```go
{
  theme: "dark" | "light" | "system",  // Default: "dark" per FR-043
  logRetentionPerServer: int,           // Default: 1000 per FR-053
  autoRefreshDiscovery: bool,           // Default: false (manual refresh per clarification)
  showSystemNotifications: bool,        // Default: true
  confirmServerStop: bool               // Default: true (safety)
}
```

**WindowLayout**:
```go
{
  width: int,         // Default: 1280
  height: int,        // Default: 800
  x: int,             // Window X position
  y: int,             // Window Y position
  maximized: bool,    // Default: false
  logPanelHeight: int // Default: 200px
}
```

**Filters**:
```go
{
  selectedServer: string | null,        // Server ID or null for all (FR-022)
  selectedSeverity: LogSeverity | null, // Severity filter or null for all (FR-023)
  searchQuery: string                   // Free-text search in logs
}
```

#### Validation Rules
- `discoveredServers` references must point to valid server records
- `monitoredConfigPaths` must be absolute paths
- `windowLayout.width` and `height` must be >=640x480 (minimum usable size)

#### Persistence
- File location: `~/.mcpmanager/state.json`
- Write strategy: Atomic (write to temp, rename on success)
- Auto-save: Debounced (max 1 write/second)
- Backup: Previous state.json saved as state.json.backup on each write

---

## Enumerations

### StatusState
```go
const (
    StatusStopped  = "stopped"   // Server not running (red)
    StatusStarting = "starting"  // Launch initiated (blue/gray)
    StatusRunning  = "running"   // Process active (green)
    StatusError    = "error"     // Failed start or crashed (yellow)
)
```

### LogSeverity
```go
const (
    LogInfo    = "info"     // Informational (blue)
    LogSuccess = "success"  // Successful operation (green)
    LogWarning = "warning"  // Non-fatal issue (yellow)
    LogError   = "error"    // Failure (red)
)
```

### DependencyType
```go
const (
    DepRuntime     = "runtime"     // Language runtime
    DepLibrary     = "library"     // System library
    DepTool        = "tool"        // External tool
    DepEnvironment = "environment" // Environment variable/config
)
```

### DiscoverySource
```go
const (
    SourceClientConfig = "client_config" // Found in MCP client config file
    SourceFilesystem   = "filesystem"    // Found by scanning install paths
    SourceProcess      = "process"       // Detected as running process
)
```

---

## Relationships Diagram

```
ApplicationState
  │
  ├──> discoveredServers [1:N] ──> MCPServer
  ├──> userPreferences [1:1] ──> UserPreferences
  ├──> windowLayout [1:1] ──> WindowLayout
  └──> selectedFilters [1:1] ──> Filters

MCPServer
  ├──> status [1:1] ──> ServerStatus
  ├──> configuration [1:1] ──> ServerConfiguration
  ├──> dependencies [1:N] ──> Dependency
  └──> (via id reference) [1:N] ──> LogEntry

LogEntry
  └──> source [N:1] ──> MCPServer.id
```

---

## Data Validation Summary

| Entity | Critical Validations |
|--------|---------------------|
| **MCPServer** | Unique name, valid installation path, PID consistency with status |
| **ServerStatus** | State transition rules enforced, metrics null when not running |
| **ServerConfiguration** | Environment var name format, path existence, restart attempt limits |
| **LogEntry** | Max 1000 per server (circular buffer), valid source reference |
| **Dependency** | Version matching logic, actionable installation instructions |
| **ApplicationState** | Valid server ID references, window size minimums, atomic file writes |

---

## Implementation Notes

### State Machine Implementation
- Implement `ServerStatus` state transitions with explicit validation
- Log all state changes with timestamp and reason
- Emit `ServerStatusChanged` event on each transition

### Persistence Strategy
- Serialize all entities to JSON
- Use Go's `encoding/json` with struct tags
- Implement `json.Marshaler` and `json.Unmarshaler` for custom serialization (e.g., timestamps)
- Validate on load; discard corrupted records with error log

### Memory Management
- Limit total log entries: 50 servers * 1000 entries = 50k max
- Estimated memory per log entry: ~1KB → 50MB for logs
- Total idle memory budget: ~100MB (50MB logs + 50MB app + GUI)

---

## References

- Feature Specification: `specs/001-mcp-manager-specification/spec.md`
- Research Decisions: `specs/001-mcp-manager-specification/research.md`
- Clarifications: `specs/001-mcp-manager-specification/spec.md` § Clarifications
