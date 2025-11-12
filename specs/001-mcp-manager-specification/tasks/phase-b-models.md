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

