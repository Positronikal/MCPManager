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

## Menu Bar

**Rationale:**
Users expect menu bars at the top of application windows (Windows) or or displayed as global menu bars at the top of the desktop (Unix-like systems, i.e. Top Bar/Panel - GNOME/Unity, Application Menu Bar - KDE Plasma, Menu Bar - macOS).

**Proposed Features:**
Add typical menu bar items, e.g. File, Edit, View, Window, Help, etc.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** Unix Rule of Least Surprise.
- **Priority:** < ... >

## Navigation Pane

**Rationale:**
Current UI presents the app name "MCP Manager" in three locations: the top window bar, the app header pane, and the left-side navigation pane header. Both the top bar and app header pane locations are expected locations for this type of branding, but the navigation header pane isn't since a user would expect this to describe what is found in the pane below it, which is the "UTILITIES" separator followed by each of the accompanying utilities, e.g. Netstat, Shell, Explorer, etc.

This pane also has only one category of item: UTILITIES. Adding a new category, "DIAGNOSTICS", allows segregation by use case and helps a user decide when to use any item intuitively without having to provide explicit instructions. Netstat and Services are purely monitoring and diagnostics tools, where Shell and Explorer provide utilitarian potential to make changes. Help is in a separate category itself and can be segregated from the others with a simple horizontal graphic.

**Proposed Features:**
Change header from "MCP Manager" to "Views" (see .\nav_pane_change.png) and add a new .

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** Unix Rules of Least Surprise and Extensibility. This also leaves open the possibility that this pane may included additional categories in the future.
- **Priority:** QUICK WIN!

## UI Scaling

**Rationale:**
MCP Manager is intended to serve a utility purpose. Utilities like this, e.g. XAMPP's control panel, are typically scaled small so that when the UI is open, it's as non-blocking and desktop-space conservative as possible. This is different from an application where interactive usage is the core expectation, e.g. Microsoft Word. Bewing small and unobtrusive allows a user to "dock" the UI somewhere on the desktop to keep an eye on it while the rest of the desktop area is used for interactive apps the user is actively working in.

**Proposed Features:**
Scale app down to "utility size" similar to XAMPP's UI while maintaining HiDPI compatibility (see .\app_scaling_comparison.png).

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** Unix Rule of Least Surprise.
- **Priority:** < ... >

## Theme Selector

**Rationale:**
User-selectable themes have become ubiquitous and are expected options for today's apps.

**Proposed Features:**
Add a "Themes" option to [View menu](#menubar), if only the options Light, Dark, and System.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** Unix Rule of Least Surprise.
- **Priority:** < ... >

## Server Table View

**Rationale:**
Defining a user-accessible standard MCP installation location[^1] for MCP servers allows MCP Manager to discover them automatically. Users clone/install MCP servers to this standard directory and MCP Manager automatically discovers them (new filesystem discovery source). Servers appear in the table view regardless of whether any MCP client is configured to use them. This becomes the "canonical" place for MCP servers on a system in addition to any existing standard locations, e.g. "%APPDATA%\Claude\Claude Extensions".

**Proposed Features:**
Add a defined standard MCP server installation directory that MCP Manager looks in during discovery to those it's already designed to inspect so that new servers appear in the Server Table View:
- Linux & Mac: "~/.local/mcp-servers"
- Windows: "%USERPROFILE%\.local\mcp-servers"

These servers should appear in the list regardless whether an MCP client is configured to use them. Pertinent documentation, such as README.md, must be updated to instruct users to create this directory and place unpacked MCP servers here. Optionally, the directory can be created as part of an installer solution.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:**
  - Decouples server discovery from client configuration
  - Gives users a clear, predictable place to install servers
  - Allows MCP Manager to function as a standalone server management tool
  - Mirrors established patterns (like ~/.local/bin for executables)
- **Priority:** < ... >

## Shell View

**Rationale:**
< ... >

**Proposed Features:**
Left-align the Platform terminal list and Quick Tips section.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** QUICK WIN!

## Explorer View

**Rationale:**
< ... >

**Proposed Features:**
Open Directory buttons should open the standard MCP server installation directory (see [Server Table View](#serverttableview).

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** QUICK WIN!

## Log Viewer

**Rationale:**
< ... >

**Proposed Features:**
Add information to identify that the logs to be displayed are those logs related to server operations run from MCP Manager, not server run logs generally.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** QUICK WIN!

## Application Icon

**Rationale:**
< ... >

**Proposed Features:**
Design and deploy an appropriate app icon set to replace the Wails icons currently in use. These should be made part of the program executable in the normal fashion for that executable type, e.g. Linux ELF, Mac Mach-O, or WIN32 PE.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** QUICK WIN!

## Continuous Operations

**Rationale:**
< ... >

**Proposed Features:**
Allow MCP Manager to minimize to and restore from the System Tray allowing it to continue running in the background and having an appropriate context menu for the systray icon. App notifications should use the host OS user notification system when running in the background.

**Cost/Benefit:**
- **Cost:** < ... >
- **Benefit:** < ... >
- **Priority:** < ... >

## Binary Release Distribution

**Rationale:**
MCP has incredible potential but is stuck in "developer early adopter" mode. Non-technical users hit barriers at "install Node.js, clone repos, npm install, configure JSON files" and give up. This leads to negative perception ("MCP sucks") spreading on social media, unfairly maligning the technology. MCP Manager can be the gateway to MCP adoption by making it accessible to non-technical users - but only if they can actually run it without building from source.

**Proposed Features:**

### Release Packaging
- [ ] **Windows Distribution**
  - `.exe` installer (NSIS or WiX tooling)
  - Portable `.exe` (single file, no installer required)
  - Code signing certificate (prevents "Unknown Publisher" warnings)
  - WebView2 runtime bundled or auto-installed

- [ ] **macOS Distribution**
  - `.dmg` disk image (standard Mac distribution format)
  - Code signed and notarized (required for Gatekeeper compliance)
  - Universal binary (Intel + Apple Silicon support)

- [ ] **Linux Distribution**
  - `.AppImage` (single-file, works on most distros - highest priority)
  - `.deb` package (Debian/Ubuntu)
  - `.rpm` package (Fedora/RHEL)
  - Flatpak (optional, increasingly popular)

### Release Infrastructure
- [ ] **Automated Build Pipeline**
  - GitHub Actions workflow for multi-platform builds
  - Automated testing on all target platforms
  - Version tagging and release note generation
  - Asset upload to GitHub Releases

- [ ] **Release Checklist**
  - Semantic versioning (v1.0.0 format)
  - User-focused release notes (not developer-focused)
  - Dead-simple installation instructions (assume zero technical knowledge)
  - Minimum system requirements clearly stated
  - Screenshots/video showing functionality in 30 seconds
  - Quick start guide: "Install → Launch → See your servers"

### Distribution Channels
- [ ] **Primary**: GitHub Releases (free, reliable, version history)
- [ ] **Secondary**: Direct downloads from project website (enables metrics)
- [ ] **Future**: Package managers (Homebrew, Chocolatey, winget, apt repositories)
- [ ] **Ecosystem**: Get listed on modelcontextprotocol.io as official tool

### Documentation Requirements
- [ ] Installation guide with platform-specific instructions
- [ ] Troubleshooting guide for common issues
- [ ] Video walkthrough (2-3 minutes showing install → first use)
- [ ] FAQ covering non-technical user questions

### Marketing/Positioning
Position as **"The Missing Control Panel for MCP"**:
- "Finally understand what MCP servers you have"
- "Start, stop, and monitor servers without touching config files"
- "See logs and errors in one place instead of scattered terminal windows"
- "No command line required - just install and run"

Sample announcement text:
> "Tried MCP but got frustrated with configuration? MCP Manager gives you a simple desktop app to discover, manage, and monitor all your MCP servers. No command line required. Download for Windows/Mac/Linux →"

**Cost/Benefit:**
- **Cost:** Medium-High (2-4 weeks initial setup)
  - GitHub Actions workflow configuration
  - Code signing certificates ($100-300/year for Windows, $99/year for macOS)
  - Testing on multiple platforms and OS versions
  - Documentation creation (guides, videos, screenshots)
  - Community engagement and support preparation
- **Benefit:** Very High (ecosystem impact)
  - Dramatically lowers barrier to MCP adoption
  - Expands user base beyond developers
  - Reduces "MCP is too hard" negative perception
  - Positions MCP Manager as essential MCP ecosystem tool
  - Enables organic growth through user recommendations
  - Provides usage data and feedback for future improvements
- **Priority:** High - Should be completed before or immediately after v1.0.0 release

**Success Metrics:**
- Download count across platforms
- User retention (downloads vs active users)
- Community feedback (GitHub issues, discussions)
- Reduction in "MCP is too complicated" sentiment
- Increase in MCP server usage generally (tracked via telemetry if added)

**Dependencies:**
- Code signing certificates acquired
- GitHub Actions workflow tested and validated
- Installation documentation completed
- At least one promotional video/walkthrough created

**Why High Priority:**
The current MCP ecosystem suffers from an adoption problem, not a technology problem. MCP Manager is positioned to solve this by being the first truly user-friendly tool for MCP management. Without binary releases, it remains a developer tool. With binary releases, it becomes the gateway that brings MCP to mainstream users - which benefits the entire ecosystem, not just this project.

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
