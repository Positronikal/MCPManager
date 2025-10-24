# Bug Tracking

## Active Bugs

*No active bugs*

---

## Fixed Bugs

### BUG-001: Server State Desynchronization After Stop Operation ✅ FIXED
**Date Reported:** 2025-10-23  
**Date Fixed:** 2025-10-23  
**Fixed In:** Phase 1 implementation by Claude Code  
**Severity:** High  
**Components:** Discovery Service, Lifecycle Service

#### Problem Description
When a server was stopped via MCP Manager's Stop button, the cache retained stale "Running" state instead of updating to "Stopped". This caused Refresh to not update state correctly and Restart operations to fail.

#### Root Cause
Lifecycle service modified server state in memory but never synchronized changes to the discovery service cache. The cache remained out of sync until next discovery scan.

#### Solution Implemented
Added synchronous cache updates in `internal/core/lifecycle/lifecycle.go`:
1. Added `DiscoveryService` interface to avoid circular dependency
2. Modified `LifecycleService` to accept `discoveryService` parameter
3. Added `discoveryService.UpdateServer(server)` calls after all state transitions:
   - `StopServer()` - both code paths
   - `StartServer()` - after setting PID
   - `monitorProcess()` - all state transitions (error, stopped, running)

#### Verification
Terminal logs confirm fix working:
```
time=2025-10-23T20:53:23.267 level=INFO msg="StopServer: Cache synchronized with stopped state"
```

**Test Evidence:** 
- Original bug: `etc/MCPMANAGER_button_test_20251023T1908.txt`
- Fixed version: `etc/MCPMANAGER_button_test_20251023T2102.txt`

#### Related Issues
During testing, discovered that UI table doesn't refresh in real-time after state changes. This is **NOT** part of this bug - it's a missing feature from the original specifications:
- **FR-005**: "System MUST update server status in real-time without requiring manual page refresh"
- **FR-047**: "System MUST display status changes immediately without manual refresh"

**Note:** The backend (cache) is now correctly synchronized. The UI reactivity issue should be tracked as a separate implementation task for FR-005/FR-047, not as a bug.

#### Files Modified
- `internal/core/lifecycle/lifecycle.go` - Added cache synchronization
- `app.go` - Updated dependency injection
- `tests/integration/lifecycle_test.go` - Updated test signatures

#### Specification References
- Feature Spec: `specs/001-mcp-manager-specification/spec.md` (FR-006, FR-009)
- Data Model: `specs/001-mcp-manager-specification/data-model.md` (MCPServer, ServerStatus)

---

## Bug Reporting Guidelines

### Before Reporting
- Check existing bugs to avoid duplicates
- Ensure you're using the latest version
- Test in development mode (`wails dev`) for detailed logs

### Bug Report Template
When reporting bugs, include:
- **Environment**: OS, Go version, Wails version
- **Configuration**: Relevant server configs
- **Steps to Reproduce**: Clear, minimal reproduction steps
- **Expected vs Actual**: What should happen vs what does happen
- **Logs**: Terminal output from `wails dev`
- **Spec References**: Which FR/requirements are violated

### Severity Levels
- **Critical**: System unusable, data loss risk
- **High**: Major feature broken, workaround exists
- **Medium**: Feature impaired, minor impact
- **Low**: Cosmetic issue, no functional impact

### Priority
- **P0**: Blocks release, fix immediately
- **P1**: High impact, fix in current sprint
- **P2**: Medium impact, schedule for next sprint
- **P3**: Low impact, fix when convenient

---

## Known Limitations

These are intentional design decisions per specifications, not bugs:

1. **No MCP Client Config Modification** (FR-019)
   - MCP Manager displays but never modifies Claude Desktop/Cursor configs
   - Users must edit client configs manually

2. **No Remote Server Installation**
   - MCP Manager only manages already-installed servers
   - Users must install servers via npm/pip/etc. first

3. **Single Instance Only** (FR-051)
   - Only one MCP Manager instance per machine
   - Second launch shows existing window

4. **UI Real-Time Updates** (FR-005, FR-047)
   - Currently, UI table requires manual refresh after lifecycle operations
   - Backend cache updates correctly, but UI doesn't subscribe to state change events
   - This is a missing feature implementation, not a bug
   - Should be implemented as part of normal spec-kit development workflow

---

## Contact

- **Security Issues**: See SECURITY.md for private disclosure
- **General Bugs**: GitHub Issues (this file guides reporting)
- **Feature Requests**: specs/ directory (add new spec-kit feature)
