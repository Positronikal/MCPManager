# Next Session: Complete Remaining APIs

## Status Summary

**Tests**: ✅ All passing (100%)
- Discovery tests: Fixed nil slice issue
- Lifecycle restart test: Fixed mock to use `StartWithOutputFunc`
- Contract tests: All passing

**Specification Progress**: 96/98 tasks complete (98%)

## Remaining Work (2 Tasks)

### T-E013: Netstat Backend API

**File**: `internal/api/monitoring.go` (new handler needed)

**Requirement**:
- Endpoint: `GET /api/v1/netstat?pids=<comma-separated-pids>`
- Response: `{ connections: NetstatEntry[] }`
- NetstatEntry: `{ protocol, localAddress, remoteAddress, state, pid }`

**Frontend Status**: ✅ Complete (showing mock data with notice)
- File: `frontend/src/components/NetstatView.svelte` (lines 65-74)
- Currently returns mock data and notifies user API not implemented

**Implementation Steps**:
1. Create platform-specific netstat utilities:
   - `internal/platform/netstat_windows.go` - Parse `netstat -ano`
   - `internal/platform/netstat_darwin.go` - Parse `netstat -anvp tcp`
   - `internal/platform/netstat_linux.go` - Parse `netstat -antp`

2. Add handler to `internal/api/monitoring.go`:
   ```go
   type NetstatEntry struct {
       Protocol      string `json:"protocol"`
       LocalAddress  string `json:"localAddress"`
       RemoteAddress string `json:"remoteAddress"`
       State         string `json:"state"`
       PID           int    `json:"pid"`
   }

   func (h *MonitoringHandlers) GetNetstat(w http.ResponseWriter, r *http.Request) {
       pidsParam := r.URL.Query().Get("pids")
       // Parse PIDs, call platform.GetNetstat(pids), return JSON
   }
   ```

3. Wire endpoint in `internal/api/router.go`:
   ```go
   r.Get("/netstat", monitoringHandlers.GetNetstat)
   ```

4. Update frontend to remove mock notification (line 74 in NetstatView.svelte)

**Testing**:
- Start MCP Manager with running servers
- Navigate to Netstat view
- Verify real network connections shown
- Test with no running servers (should show empty list)

---

### T-E016: Services Backend API

**File**: `internal/api/monitoring.go` (new handler needed)

**Requirement**:
- Endpoint: `GET /api/v1/services`
- Response: `{ services: Service[] }`
- Service: `{ name, status, description, pid? }`

**Frontend Status**: ✅ Complete (showing mock data with notice)
- File: `frontend/src/components/ServicesView.svelte` (lines 57-68)
- Currently returns mock data and notifies user API not implemented

**Implementation Steps**:
1. Create platform-specific service utilities:
   - `internal/platform/services_windows.go` - Run `sc query` and parse output
   - `internal/platform/services_darwin.go` - Run `launchctl list` and parse output
   - `internal/platform/services_linux.go` - Run `systemctl list-units --type=service` and parse output

2. Add handler to `internal/api/monitoring.go`:
   ```go
   type Service struct {
       Name        string `json:"name"`
       Status      string `json:"status"`
       Description string `json:"description"`
       PID         *int   `json:"pid,omitempty"`
   }

   func (h *MonitoringHandlers) GetServices(w http.ResponseWriter, r *http.Request) {
       // Call platform.GetServices(), return JSON
   }
   ```

3. Wire endpoint in `internal/api/router.go`:
   ```go
   r.Get("/services", monitoringHandlers.GetServices)
   ```

4. Update frontend to remove mock notification (line 68 in ServicesView.svelte)

**Testing**:
- Start MCP Manager
- Navigate to Services view
- Verify real system services shown
- Test filtering by status and search

---

## Files Changed This Session (Test Fixes)

- `internal/core/discovery/discovery.go:399` - Initialize empty slice instead of nil
- `internal/core/discovery/filesystem.go:31` - Initialize empty slice instead of nil
- `internal/core/lifecycle/lifecycle_test.go:289` - Use `StartWithOutputFunc` in RestartServer test
- `internal/core/lifecycle/lifecycle_test.go:318` - Add wait for async operation

## Commit Message Template

```
fix: Resolve 5 test failures (discovery, lifecycle, contract)

Fixed nil slice returns in discovery causing test failures:
- discovery.go: GetCachedServers() now returns empty slice
- filesystem.go: DiscoverFromFilesystem() now returns empty slice

Fixed lifecycle restart test mock configuration:
- Use StartWithOutputFunc instead of StartFunc
- Add IsRunningFunc for both old and new PIDs
- Add wait for async start operation to complete

All tests now passing (100% pass rate).

Remaining work: 2 backend APIs (T-E013 Netstat, T-E016 Services)

🤖 Generated with [Claude Code](https://claude.com/claude-code)

Co-Authored-By: Claude <noreply@anthropic.com>
```

## Quick Start for Next Session

1. **Read this file first** for context
2. **Implement T-E013 (Netstat)**:
   - Create platform utilities
   - Add API handler
   - Wire endpoint
   - Remove frontend mock notice
3. **Implement T-E016 (Services)**:
   - Same pattern as Netstat
4. **Test both features** manually
5. **Update CLAUDE.md** with 98/98 complete
6. **Commit and merge to main**
7. **Tag v1.0.0** 🎉

## Notes

- Both frontends are fully functional with mock data
- Users are clearly notified APIs aren't implemented yet
- No breaking changes - APIs can be added incrementally
- Platform utilities pattern already established (see `internal/platform/utilities_*.go`)
