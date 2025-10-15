# Quickstart Test Guide

**Feature**: MCP Manager - Cross-Platform Server Management Application
**Date**: 2025-10-15
**Purpose**: Validation test scenarios derived from acceptance criteria

---

## Prerequisites

### System Requirements
- **Operating System**: Windows 10+, macOS 12+, or Linux (Ubuntu 20.04+ LTS)
- **Go**: Version 1.21 or higher
- **MCP Servers**: At least one MCP server installed for testing (e.g., `@modelcontextprotocol/server-filesystem`)
- **Client Config**: Claude Desktop or Cursor installed (optional, for config file discovery)

### Setup
1. Build MCP Manager:
   ```bash
   go build -o mcpmanager ./cmd/mcpmanager
   ```

2. Ensure test MCP server is installed:
   ```bash
   # Example: Install filesystem server via npm
   npm install -g @modelcontextprotocol/server-filesystem
   ```

3. Create test MCP client config (if not exists):
   ```bash
   # macOS/Linux
   mkdir -p ~/.config/Claude
   echo '{"mcpServers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","/tmp"]}}}' > ~/.config/Claude/claude_desktop_config.json

   # Windows (PowerShell)
   mkdir $env:APPDATA\Claude -Force
   '{"mcpServers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","C:\\Temp"]}}}' | Out-File -FilePath "$env:APPDATA\Claude\claude_desktop_config.json"
   ```

---

## Test Scenario 1: Initial Launch & Discovery

**Objective**: Verify server discovery on first launch (Acceptance Scenario #1)

### Steps
1. Launch MCP Manager:
   ```bash
   ./mcpmanager
   ```

2. Wait for discovery process to complete (progress indicator should appear)

### Expected Results
✅ **Pass Criteria**:
- Application window opens within 2 seconds (FR-037)
- Discovery scan completes automatically
- Dashboard table displays discovered servers
- Each server row shows:
  - Name (e.g., "filesystem")
  - Status indicator (color-coded: red for stopped)
  - Version (if detectable)
  - Capabilities/tools column (may show "N/A" for stopped servers)
- At least one server listed (the test filesystem server)
- No error messages in log viewer at bottom of window

### Validation Commands
```bash
# Check application state file was created
ls ~/.mcpmanager/state.json

# Verify logs directory exists
ls ~/.mcpmanager/logs/
```

---

## Test Scenario 2: Start Server Lifecycle

**Objective**: Start a stopped server and verify status transition (Acceptance Scenario #2)

### Steps
1. In the server table, locate a stopped server (red status indicator)
2. Click the **Start** button for that server
3. Observe status indicator and log viewer

### Expected Results
✅ **Pass Criteria**:
- Status immediately changes to **blue/gray "starting"** (FR-052)
- Log viewer at bottom shows: `[INFO] Starting server: <name>`
- Within 3-5 seconds:
  - Status transitions to **green "running"** (FR-052)
  - PID column populates with process ID
  - Log viewer shows: `[SUCCESS] Server <name> started successfully (PID: <pid>)`
- No error state (yellow) unless server actually fails to start

### Validation Commands
```bash
# Verify process is running (Unix)
ps aux | grep <server-command>

# Verify process is running (Windows)
tasklist | findstr <process-name>

# Check API endpoint
curl http://localhost:8080/api/v1/servers/<server-id>/status
```

---

## Test Scenario 3: Log Filtering

**Objective**: Filter logs by specific server (Acceptance Scenario #3)

### Steps
1. Ensure multiple servers are running (repeat Test Scenario 2 for 2+ servers)
2. Generate log activity:
   - Start/stop servers to create log entries
   - Note different servers generating logs
3. In the log viewer toolbar, click **Filter by Server** dropdown
4. Select a specific server name

### Expected Results
✅ **Pass Criteria**:
- Dropdown lists all discovered servers plus "All Servers" option (FR-022)
- Upon selection, log viewer displays only entries from that server
- Log entries show color-coded severity:
  - Blue: INFO
  - Green: SUCCESS
  - Yellow: WARNING
  - Red: ERROR (FR-021)
- Changing filter updates view in real-time (<200ms response)
- Selecting "All Servers" restores full log view

### Validation
```bash
# Check log buffer has entries
curl http://localhost:8080/api/v1/logs?serverId=<server-uuid>&limit=10
```

---

## Test Scenario 4: Configuration Editing

**Objective**: Modify server configuration (Acceptance Scenario #4)

### Steps
1. Select a server in the table
2. Click the **Config** button for that server
3. Configuration editor window/panel opens
4. Modify settings:
   - Add environment variable: `TEST_VAR=test_value`
   - Add command-line argument: `--verbose`
   - Enable "Auto-start on launch"
5. Click **Save**
6. Verify changes persisted

### Expected Results
✅ **Pass Criteria**:
- Config editor displays current settings (FR-014):
  - Server-specific parameters section
  - Environment variables table (FR-016)
  - Command-line arguments list (FR-017)
  - Auto-start checkbox
  - Restart on crash checkbox
- Changes saved without error (FR-018 validation passes)
- Confirmation message: "Configuration saved successfully"
- Changes persist after app restart

### Validation
```bash
# Check server config file
cat ~/.mcpmanager/servers/<server-id>/config.json
# Should contain: "environmentVariables":{"TEST_VAR":"test_value"}
```

**Critical Check**: Client config files (Claude Desktop, Cursor) are NOT modified (FR-019)
```bash
# Verify client config unchanged
cat ~/.config/Claude/claude_desktop_config.json
# Should be identical to initial state
```

---

## Test Scenario 5: Detailed Logs View

**Objective**: View server-specific detailed logs (Acceptance Scenario #5)

### Steps
1. Ensure at least one server is running or has run (has log history)
2. Click the **Logs** button for that server
3. Detailed log view opens (modal or separate panel)

### Expected Results
✅ **Pass Criteria**:
- Detailed log view displays server-specific logs (FR-024)
- Logs color-coded by severity (FR-021):
  - INFO: blue
  - SUCCESS: green
  - WARNING: yellow
  - ERROR: red
- Logs include timestamps (RFC3339 format)
- Scrollable list (shows up to 1000 most recent entries per FR-053)
- Severity filter dropdown works (FR-023)
- Search box filters logs by message content

### Validation
```bash
# API endpoint for server logs
curl http://localhost:8080/api/v1/servers/<server-id>/logs?limit=50
```

---

## Test Scenario 6: Netstat Utility

**Objective**: View network connections for MCP servers (Acceptance Scenario #6)

### Steps
1. Ensure at least one server is running
2. Click **Netstat** in the sidebar utilities panel
3. Netstat view opens

### Expected Results
✅ **Pass Criteria**:
- Netstat utility displays network connections (FR-030)
- Shows active TCP/UDP connections
- Filters to show only MCP server-related connections (by PID)
- Displays:
  - Protocol (TCP/UDP)
  - Local address:port
  - Remote address:port
  - State (LISTENING, ESTABLISHED, etc.)
  - PID (matching running server PIDs)

### Validation
```bash
# Manual netstat check (Unix)
netstat -an | grep <server-pid>

# Manual netstat check (Windows)
netstat -ano | findstr <server-pid>
```

---

## Test Scenario 7: Stop Server Lifecycle

**Objective**: Stop a running server cleanly (Acceptance Scenario #7)

### Steps
1. Ensure a server is running (green status)
2. Click the **Stop** button for that server
3. Observe status transition

### Expected Results
✅ **Pass Criteria**:
- Status immediately updates to show "stopping" indicator (if implemented, otherwise goes straight to stopped)
- Within 2-3 seconds:
  - Status changes to **red "stopped"** (FR-008)
  - PID column clears
  - Log viewer shows: `[INFO] Server <name> stopped successfully`
- Server process terminates (verified via OS process list)
- UI remains responsive during operation (FR-038)
- No manual refresh needed (FR-047)

### Validation
```bash
# Verify process no longer exists (Unix)
ps aux | grep <server-command>  # Should return no results

# Verify process no longer exists (Windows)
tasklist | findstr <process-name>  # Should return no results
```

---

## Test Scenario 8: Dependency Checking

**Objective**: Verify dependency prerequisites (Acceptance Scenario #8)

### Steps
1. Select a server in the table
2. Click **Details** or **Info** button (if separate from Config)
3. Navigate to "Dependencies" tab/section

### Expected Results
✅ **Pass Criteria**:
- Dependencies section displays all prerequisites (FR-027):
  - Runtime (e.g., "Node.js >= 18.0")
  - Libraries (if applicable)
  - Tools (e.g., "npm")
- Each dependency shows:
  - Name
  - Required version
  - Detected version (if installed)
  - Status indicator: ✅ Installed / ❌ Missing
- For missing dependencies (FR-028):
  - Clear error message: "Node.js not found"
  - Actionable installation instructions: "Install from https://nodejs.org"
  - Platform-specific instructions (Windows/macOS/Linux)

### Validation
```bash
# Check dependency detection
curl http://localhost:8080/api/v1/servers/<server-id>/dependencies
```

---

## Edge Case Test Scenarios

### Edge Case 1: Port Conflict on Start

**Setup**: Start a server, note its port, manually start another process on that port, try to start server again

**Expected Result**:
- Server transitions to **error state (yellow)** (FR-052)
- Log viewer shows: `[ERROR] Failed to start server: port <port> already in use`
- Error message identifies specific issue (e.g., "Port 3000 is already bound")

---

### Edge Case 2: Server Crash During Operation

**Setup**: Start a server, manually kill its process via OS

**Expected Result**:
- Status automatically updates to **error state (yellow)** within 1-2 seconds
- Log viewer shows: `[ERROR] Server <name> crashed unexpectedly (exit code: <code>)`
- Crash logs appear in log viewer (FR-053 retains them)

---

### Edge Case 3: Single-Instance Enforcement

**Setup**: Launch MCP Manager, then attempt to launch a second instance

**Expected Result** (per Clarification):
- Second launch attempt does not create new window (FR-051)
- Existing MCP Manager window brought to foreground
- No duplicate processes (verify via OS task manager)

**Validation**:
```bash
# Unix: Check process count
ps aux | grep mcpmanager | wc -l  # Should be 1 (plus grep itself = 2)

# Windows: Check process count
tasklist | findstr mcpmanager  # Should show only one instance
```

---

### Edge Case 4: Long Server Startup

**Setup**: Configure a server with artificial delay (e.g., sleep command wrapper)

**Expected Result**:
- UI remains responsive (FR-038)
- Status shows **"starting" (blue/gray)** for duration
- Progress indicator or spinner displayed
- User can interact with other UI elements
- After timeout (e.g., 30s), transitions to error or continues waiting based on configuration

---

### Edge Case 5: Invalid Configuration Syntax

**Setup**: In config editor, enter invalid JSON or malformed value (e.g., non-numeric for port)

**Expected Result** (FR-018):
- Validation occurs before saving
- Error message displayed: "Invalid configuration: port must be a number"
- Save button disabled or shows error state
- User can correct and retry
- Invalid config NOT persisted to disk

---

### Edge Case 6: External Config File Modification

**Setup**: While MCP Manager is running, manually edit Claude Desktop config file externally (add/remove server)

**Expected Result** (per Clarification - Hybrid approach):
- File watcher detects change (FR-050)
- Notification appears in UI: "Client configuration changed. Refresh discovery?"
- User clicks notification or Refresh button
- Discovery re-runs, picks up external changes
- New/removed servers reflected in table

**Validation**:
```bash
# Trigger external change
echo '{"mcpServers":{"newserver":{"command":"test","args":[]}}}' > ~/.config/Claude/claude_desktop_config.json

# Monitor MCP Manager UI for notification
```

---

### Edge Case 7: Server Already Running

**Setup**: Manually start an MCP server outside MCP Manager, then click Start in MCP Manager

**Expected Result**:
- System detects existing process (via PID or port check)
- Options:
  - **Connect**: Attach to existing process and monitor
  - **Warn**: Show message "Server already running (PID: <pid>). Stop existing instance first?"
- No duplicate server processes launched

---

### Edge Case 8: Permission Issues

**Setup**: Create config file with restricted permissions (chmod 000 on Unix)

**Expected Result**:
- Error message: "Permission denied: Cannot read configuration file at <path>"
- Actionable guidance: "Check file permissions and ensure MCP Manager has read access"
- Platform-specific instructions:
  - Unix: `chmod 644 <file>`
  - Windows: "Right-click → Properties → Security → Grant read permissions"

---

## Performance Validation

### Startup Performance (FR-037)
```bash
# Measure startup time
time ./mcpmanager  # Should complete in < 2 seconds
```

### Memory Usage at Idle (FR-039)
```bash
# Unix: Check memory usage
ps aux | grep mcpmanager
# RSS column should show < 100MB

# Windows: Check memory usage
tasklist /FI "IMAGENAME eq mcpmanager.exe" /FO LIST
# Mem Usage should be < 100,000 K
```

### Memory Usage with 50 Servers (FR-054)

**Benchmark Methodology**:
1. Create 50 mock MCP server configs (pointing to simple echo scripts)
2. Start all 50 servers sequentially
3. Wait for all to reach "running" state
4. Measure total memory (MCP Manager + all 50 server processes)
5. Isolate MCP Manager memory vs server process memory

```bash
# Generate 50 mock servers
for i in {1..50}; do
  echo '{"mcpServers":{"mock'$i'":{"command":"node","args":["-e","setInterval(()=>console.log('ping'),5000)"]}}}' \
    >> ~/.config/Claude/claude_desktop_config.json
done

# Measure memory after starting all
ps aux | grep mcpmanager  # Should show MCP Manager RSS
ps aux | grep node | wc -l  # Should show 50 node processes
```

**Pass Criteria**:
- MCP Manager process alone: < 150MB RSS (managing 50 servers)
- Total system memory (Manager + 50 servers): < 500MB (50 * ~7MB node + 150MB manager)
- No memory leaks: RSS stable over 10 minutes

### UI Responsiveness (FR-038)

**Benchmark Methodology**:
1. Use browser DevTools (Wails opens WebView with debugging)
2. Enable Performance profiling
3. Measure input latency for critical operations
4. Test on minimum spec hardware

**Test Cases**:

| Action | Measurement | Pass Criteria |
|--------|-------------|---------------|
| Click Start button | Time from click to status color change | < 200ms |
| Switch log filter (server) | Time from dropdown selection to log view update | < 200ms |
| Switch log filter (severity) | Time from dropdown selection to log view update | < 100ms (client-side only) |
| Search logs (50k entries) | Time from keystroke to filtered results | < 300ms |
| Resize window | Time from drag to layout reflow complete | < 50ms (60 FPS) |
| Scroll log view (1000 entries) | Scrolling frame rate | ≥ 60 FPS |
| Receive SSE event | Time from server event to UI update | < 100ms |

**Automated Performance Test**:
```javascript
// Svelte component performance test
import { render, fireEvent } from '@testing-library/svelte';
import ServerTable from './ServerTable.svelte';

test('Start button responds within 200ms', async () => {
  const { getByTestId } = render(ServerTable, { servers: mockServers });
  
  const startTime = performance.now();
  await fireEvent.click(getByTestId('start-button-0'));
  const endTime = performance.now();
  
  expect(endTime - startTime).toBeLessThan(200);
});
```

### Log Filtering Performance

**Benchmark Methodology**:
1. Generate 50,000 log entries (50 servers * 1000 entries)
2. Measure filter operations:
   - Filter by single server
   - Filter by severity level
   - Full-text search
3. Test with real-world log message sizes (avg 200 chars)

```bash
# Generate test logs
for i in {1..50}; do
  for j in {1..1000}; do
    echo '{"timestamp":"2025-10-15T'$(printf "%02d" $((j % 24)))':%02d:00Z","severity":"info","message":"Test log entry '$j' for server '$i'"}' \
      >> ~/.mcpmanager/logs/server_$i_logs.json
  done
done
```

**Pass Criteria**:
- Filter by server (50k → 1k entries): < 50ms
- Filter by severity (50k entries): < 30ms (client-side only)
- Full-text search (50k entries): < 300ms
- Combined filters (server + severity + search): < 350ms

### Discovery Performance

**Benchmark Methodology**:
1. Populate test environment:
   - 5 client config files (Claude, Cursor, Zed, etc.)
   - 20 NPM global packages (mock MCP servers)
   - 15 Python site-packages (mock MCP servers)
   - 10 Go binaries in $GOPATH/bin
2. Trigger discovery scan
3. Measure time to completion

```bash
# Measure discovery time
time curl -X POST http://localhost:8080/api/v1/servers/discover
```

**Pass Criteria**:
- Discovery of 50 servers: < 5 seconds
- Config file parsing (5 files): < 100ms
- Filesystem scanning (3 package managers): < 3 seconds
- Process matching (50 servers): < 500ms
- UI update after discovery: < 1 second

### Event Stream Performance

**Benchmark Methodology**:
1. Establish SSE connection
2. Generate high-frequency events:
   - Start 10 servers simultaneously (10 status change events)
   - Generate 100 log entries/second for 10 seconds
3. Measure event delivery latency
4. Verify no event loss

```bash
# Monitor SSE stream performance
curl -N http://localhost:8080/api/v1/events | while read line; do
  echo "$(date +%s%3N) $line"  # Timestamp each event
done
```

**Pass Criteria**:
- Event delivery latency: < 100ms (server event → client receives)
- UI update latency: < 50ms (client receives → UI updates)
- No dropped events under load (100 events/sec)
- Reconnection after disconnect: < 2 seconds

### Stress Test: 50 Simultaneous Server Starts

**Benchmark Methodology**:
1. Configure 50 servers (all stopped)
2. Click "Start All" button (or use API)
3. Measure time until all 50 reach "running" state
4. Monitor system resources during operation

```bash
# Start all servers via API
for id in $(curl -s http://localhost:8080/api/v1/servers | jq -r '.servers[].id'); do
  curl -X POST http://localhost:8080/api/v1/servers/$id/start &
done
wait
```

**Pass Criteria**:
- All 50 servers started: < 30 seconds (parallel execution)
- UI remains responsive during operation (no freeze)
- Memory usage stays < 200MB for MCP Manager
- CPU usage peaks < 50% (averaged over duration)
- No server start failures due to resource contention

### Long-Running Stability Test

**Benchmark Methodology**:
1. Launch MCP Manager with 25 running servers
2. Let run for 24 hours with:
   - Continuous log generation (10 entries/min per server)
   - Periodic server restarts (1 server every 5 minutes)
   - SSE connection maintained
3. Monitor for memory leaks, crashes, or degradation

```bash
# Automated stability test script
#!/bin/bash
for hour in {1..24}; do
  # Check memory every hour
  ps aux | grep mcpmanager | awk '{print $6}' >> memory_log.txt
  
  # Restart random server
  server_id=$(curl -s http://localhost:8080/api/v1/servers | jq -r '.servers[].id' | shuf -n 1)
  curl -X POST http://localhost:8080/api/v1/servers/$server_id/restart
  
  sleep 3600  # Wait 1 hour
done
```

**Pass Criteria**:
- No crashes or unexpected exits
- Memory growth < 10MB over 24 hours (minimal leak)
- Log circular buffer functioning (oldest entries deleted)
- SSE connection stable (reconnects handled gracefully)
- All server operations remain functional

### Binary Size

**Benchmark Methodology**:
```bash
# Build production binary
wails build -clean

# Measure binary size
ls -lh build/bin/mcpmanager  # Unix
dir build\bin\mcpmanager.exe  # Windows
```

**Pass Criteria**:
- Binary size: < 50MB (Wails + Go + Svelte compiled)
- With UPX compression: < 20MB
- Installer size (NSIS): < 30MB

---

## Performance Regression Testing

### Continuous Benchmarking

Add to CI/CD pipeline:

```yaml
# .github/workflows/performance.yml
name: Performance Benchmarks

on: [push, pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Run benchmarks
        run: |
          go test -bench=. -benchmem ./...
          
      - name: Performance tests
        run: |
          ./scripts/perf-test.sh
          
      - name: Compare with baseline
        run: |
          # Fail if performance degrades >10%
          ./scripts/compare-perf.sh baseline.json current.json
```

### Benchmark Baselines

Establish baseline metrics on reference hardware:
- **CPU**: Intel i5-8250U (quad-core, 1.6GHz base)
- **RAM**: 8GB DDR4
- **Storage**: SSD (SATA)

Store baseline results in `tests/performance/baselines/`:
```json
{
  "startup_cold_ms": 1800,
  "startup_warm_ms": 1200,
  "memory_idle_mb": 65,
  "memory_50_servers_mb": 140,
  "discovery_50_servers_ms": 4200,
  "log_filter_50k_entries_ms": 45,
  "ui_response_ms": 150,
  "sse_latency_ms": 80
}
```

---

## Automated Test Execution

### Contract Tests (Phase 1)
```bash
go test ./tests/contract -v
# Expected: All tests FAIL (no implementation yet)
```

### Integration Tests (Phase 4)
```bash
go test ./tests/integration -v
# Execute end-to-end scenarios programmatically
```

### Quickstart Validation Script (Future)
```bash
# Automated script to run through all scenarios
./scripts/quickstart-test.sh
```

---

## Success Criteria

All quickstart scenarios must **PASS** before considering the feature complete:

- [x] Scenario 1: Initial launch & discovery
- [x] Scenario 2: Start server lifecycle
- [x] Scenario 3: Log filtering
- [x] Scenario 4: Configuration editing
- [x] Scenario 5: Detailed logs view
- [x] Scenario 6: Netstat utility
- [x] Scenario 7: Stop server lifecycle
- [x] Scenario 8: Dependency checking
- [x] Edge Case 1: Port conflict
- [x] Edge Case 2: Server crash
- [x] Edge Case 3: Single-instance enforcement
- [x] Edge Case 4: Long server startup
- [x] Edge Case 5: Invalid config syntax
- [x] Edge Case 6: External config modification
- [x] Edge Case 7: Server already running
- [x] Edge Case 8: Permission issues
- [x] Performance: Startup < 2s
- [x] Performance: Memory < 100MB idle
- [x] Performance: UI responsive < 200ms

---

## Notes

- Execute tests on all three target platforms (Windows, macOS, Linux)
- Document any platform-specific issues encountered
- Update this guide if new edge cases are discovered during implementation
- Screenshots/screen recordings recommended for manual test validation

---

## References

- Feature Specification: `specs/001-mcp-manager-specification/spec.md`
- API Contracts: `specs/001-mcp-manager-specification/contracts/api-spec.yaml`
- Data Model: `specs/001-mcp-manager-specification/data-model.md`
