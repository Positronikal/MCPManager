# Future Enhancements & Optimizations

This document tracks potential feature improvements, optimizations, and value adds for MCP Manager that should be considered after initial release and burn-in period. These items are in no particular order.

## Implementation

**Requirements:**
1. Both enhancements require extending `models.UserPreferences` struct
2. Frontend needs settings/preferences UI panel
3. Consider adding "Advanced Settings" section to avoid cluttering main UI

**Validation Needed:**
- Measure actual performance impact in production use
- Gather user feedback on whether these views are frequently used
- Determine if resource usage justifies adding preference complexity

## Netstat View Preferences

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

## Services View Optimizations

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

## End-to-End Process/IPC Testing

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

## Main Menu

**Rationale:**
< ... >

**Proposed Features:**
Add typical main menu items, e.g. File, Edit, View, Window, Help.Add typical main menu items, e.g. File, Edit, View, Window, Help.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** < ... >

## Tools Pane

**Rationale:**
< ... >

**Proposed Features:**
Change header from "MCP Manager" to "Tools".

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** < ... >

## UI Scaling

**Rationale:**
< ... >

**Proposed Features:**
Scale app down to "utility size" similar to XAMPP's UI while maintaining HiDPI compatibility (see .\app_scaling_comparison.png).

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** < ... >

## Theme Selector

**Rationale:**
< ... >

**Proposed Features:**
Add theming, if only the options Light, Dark, and System.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** < ... >

## Server Table View

**Rationale:**
< ... >

**Proposed Features:**
Define an easily user-accessible standard MCP installation location[^1] in addition to those locations MCP Manager already looks to find MCP servers that a user can clone to and MCP Manager can discover so that new servers appear in the Server Table View like the others and regardless whether an MCP client is configured to use them or not. Pertinent documentation, such as README.md, must be updated to instruct users to create this directory and place unpacked MCP servers here. Optionally, the directory can be created as part of an installer solution.

[^1]: Linux & Mac: "~/.local/mcp-servers"
      Windows: "%USERPROFILE%\.local\mcp-servers"

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** < ... >

## Shell View

**Rationale:**
< ... >

**Proposed Features:**
Left-align the Platform terminal list and Quick Tips section.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** < ... >

## Explorer View

**Rationale:**
< ... >

**Proposed Features:**
Open Directory buttons should open the standard MCP server installation directory (see [Server Table View](#serverttableview).

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** < ... >

## Log Viewer

**Rationale:**
< ... >

**Proposed Features:**
Add information to identify that the logs to be displayed are those logs related to server operations run from MCP Manager, not server run logs generally.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** < ... >

## Application Icon

**Rationale:**
< ... >

**Proposed Features:**
Design and deploy an appropriate app icon set to replace the Wails icons currently in use. These should be made part of the program executable in the normal fashion for that executable type, e.g. Linux ELF, Mac Mach-O, or WIN32 PE.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** < ... >

## Continuous Operations

**Rationale:**
< ... >

**Proposed Features:**
Allow MCP Manager to minimize to and restore from the System Tray allowing it to continue running in the background and having an appropriate context menu for the systray icon. App notifications should use the host OS user notification system when running in the background.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** < ... >

## Related Issues and Additional notes

- None yet - document as issues arise during burn-in period
- See also .\Issues.md

---
## <entry item stub>

**Rationale:**
< ... >

**Proposed Features:**
< ... >

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** < ... >
---

*Last Updated: 2025-11-14*
*Status: Deferred pending real-world usage data*
