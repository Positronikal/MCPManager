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

