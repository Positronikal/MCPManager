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

