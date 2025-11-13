# Future Enhancements & Optimizations

This document tracks potential feature improvements and optimizations for MCP Manager that should be considered after initial release and burn-in period.

## Priority: Medium - Post-Release Optimizations

### 1. Netstat View Preferences

**Rationale:** Not all users need network connection monitoring, especially when running only stdio transport servers.

**Proposed Features:**
- [ ] Add user preference to disable Netstat view entirely
- [ ] Make auto-refresh interval configurable (currently hardcoded to 5 seconds)
  - Default: 5s
  - Range: 1s - 60s
  - Option to disable auto-refresh by default

**Cost/Benefit:**
- **Cost:** Low (preference storage already exists via ApplicationState)
- **Benefit:** Reduces unnecessary system calls for users who don't need network monitoring
- **Priority:** Wait for user feedback - stdio-only users may not use this view at all

### 2. Services View Optimizations

**Rationale:** System services change infrequently, but the view currently queries on every refresh.

**Proposed Features:**
- [ ] Add caching layer with configurable TTL
  - Default TTL: 30 seconds
  - Invalidate on manual refresh
- [ ] Add user preference to disable Services view entirely
- [ ] Consider lazy-loading (only fetch when view is first accessed)

**Cost/Benefit:**
- **Cost:** Low-Medium (requires cache implementation + invalidation logic)
- **Benefit:** Reduces subprocess overhead (313 services on Windows = ~1-10MB per query)
- **Priority:** Monitor real-world usage patterns first

## Implementation Notes

**Requirements:**
1. Both enhancements require extending `models.UserPreferences` struct
2. Frontend needs settings/preferences UI panel
3. Consider adding "Advanced Settings" section to avoid cluttering main UI

**Validation Needed:**
- Measure actual performance impact in production use
- Gather user feedback on whether these views are frequently used
- Determine if resource usage justifies adding preference complexity

## Priority: Low - Testing Infrastructure

### 3. End-to-End Process/IPC Testing

**Rationale:** Current test suite lacks true end-to-end tests that exercise the Wails desktop application as users would interact with it.

**Background:**
- MCP Manager is a Wails desktop application using IPC/bridge communication
- Initial integration tests (deleted in commit af6af34+) were HTTP-based and incompatible with Wails architecture
- Current test coverage:
  - ✅ Unit tests for all core services
  - ✅ Contract tests for API handlers (in-memory httptest)
  - ✅ Integration tests for discovery/lifecycle services
  - ❌ No end-to-end tests of full Wails application

**Proposed Testing Approach:**
- [ ] **Wails Process Tests**: Launch actual `mcpmanager.exe` and interact via IPC
  - Test application startup and initialization
  - Verify Wails event emission and frontend updates
  - Test full user workflows (discover → start → stop → monitor)
  - Validate window state persistence

- [ ] **Playwright/WebDriver Tests**: Automate UI interactions
  - Requires headless Wails support or virtual display
  - Test button clicks, form inputs, navigation
  - Verify UI state updates from backend events

- [ ] **Test Server Infrastructure**: Build minimal MCP test servers
  - Stdio transport test server (pipes)
  - HTTP/SSE transport test server (network)
  - Crash/error scenario servers for edge case testing

**Cost/Benefit:**
- **Cost:** High (2-3 weeks of work)
  - Research Wails testing best practices
  - Set up process launching and IPC communication
  - Implement test fixtures and utilities
  - Write comprehensive test scenarios
- **Benefit:** High (confidence in production deployments)
  - Catch UI bugs before release
  - Validate cross-platform behavior
  - Test actual user workflows end-to-end
- **Priority:** Defer until after v1.0 release and real-world usage validation

**Alternative Approach:**
- Manual testing checklist for releases
- Community beta testing program
- Telemetry/crash reporting to identify issues in production

**Why Deferred:**
1. Current unit and contract tests provide good coverage
2. Manual testing has validated core functionality
3. Wails E2E testing is non-trivial and tooling is immature
4. Better to gather real-world usage data first before investing in complex test infrastructure

## Related Issues

- None yet - document as issues arise during burn-in period

---

*Last Updated: 2025-11-12*
*Status: Deferred pending real-world usage data*
