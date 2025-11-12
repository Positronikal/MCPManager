## Phase A: Foundation & Setup (Tasks A001-A008)

**Objective**: Establish project structure, dependencies, and platform abstractions

### T-A001 [P] Initialize Go module and project structure
**File**: `go.mod`, project directories
**Effort**: M (1-2 hours)
**Dependencies**: None

**Description**:
Initialize Go module for MCP Manager and create the complete directory structure per plan.md.

**Steps**:
1. Run `go mod init github.com/your-org/mcpmanager` from repo root
2. Create directories:
   - `cmd/mcpmanager/`
   - `internal/api/`, `internal/core/{discovery,lifecycle,config,monitoring}/`
   - `internal/models/`, `internal/storage/`, `internal/platform/`
   - `frontend/src/{components,stores,services,routes}/`
   - `tests/{contract,integration,unit,frontend}/`
   - `pkg/mcpclient/`
3. Create `.gitignore` for Go + Node.js + Wails

**Acceptance**:
- `go.mod` exists with module path
- All directories created and committed
- `go mod tidy` runs without errors

---

### T-A002 [P] Setup local MCP SDK module reference
**File**: `go.mod`
**Effort**: S (<1 hour)
**Dependencies**: T-A001

**Description**:
Configure go.mod to use local MCP Go SDK per research.md decision.

**Steps**:
1. Add replace directive to `go.mod`:
   ```go
   replace github.com/modelcontextprotocol/go-sdk => ../../../_MCP-Tools-Dev/go-sdk
   ```
2. Add requirement: `require github.com/modelcontextprotocol/go-sdk v0.0.0`
3. Run `go mod tidy`

**Acceptance**:
- go.mod contains replace directive
- SDK imports resolve without errors
- No network fetch attempts for MCP SDK

---

### T-A003 [P] Initialize Wails project with Svelte frontend
**File**: `wails.json`, `frontend/` structure
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Initialize Wails v2 project with Svelte 4.x frontend per research.md.

**Steps**:
1. Run: `wails init -n mcpmanager -t svelte` (use existing structure, merge if needed)
2. Configure `wails.json`:
   - Set `outputfilename: "mcpmanager"`
   - Enable `devServer` for frontend hot reload
   - Configure `frontend:build` command: `npm run build`
3. Initialize frontend dependencies: `cd frontend && npm install`
4. Verify Wails CLI: `wails doctor`

**Acceptance**:
- `wails.json` configured correctly
- `frontend/package.json` contains Svelte 4.x dependencies
- `wails dev` starts without errors (Ctrl+C to stop)
- Frontend displays default Wails + Svelte template

---

### T-A004 [P] Add fsnotify dependency for file watching
**File**: `go.mod`
**Effort**: S (<1 hour)
**Dependencies**: T-A001

**Description**:
Add fsnotify library for monitoring client config file changes (FR-050).

**Steps**:
1. Run: `go get github.com/fsnotify/fsnotify`
2. Verify installation: `go mod tidy`
3. Create placeholder: `internal/platform/filewatcher.go` with interface:
   ```go
   type FileWatcher interface {
       Watch(path string) error
       Events() <-chan Event
       Close() error
   }
   ```

**Acceptance**:
- fsnotify appears in `go.mod`
- FileWatcher interface defined
- No build errors

---

### T-A005 [P] Implement platform abstraction interfaces
**File**: `internal/platform/interfaces.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A001

**Description**:
Define platform-specific abstraction interfaces per research.md §8.

**Steps**:
1. Create `internal/platform/interfaces.go`:
   ```go
   type PathResolver interface {
       GetConfigDir() string
       GetAppDataDir() string
       GetUserHomeDir() string
   }

   type ProcessManager interface {
       Start(cmd string, args []string, env map[string]string) (pid int, err error)
       Stop(pid int, graceful bool, timeout int) error
       IsRunning(pid int) bool
   }

   type SingleInstance interface {
       Acquire() (bool, error)
       Release() error
   }
   ```

**Acceptance**:
- All interfaces defined in `internal/platform/interfaces.go`
- Interfaces exported (capitalized)
- Go build succeeds

---

### T-A006 [P] Implement platform-specific path resolvers
**Files**: `internal/platform/paths_windows.go`, `paths_darwin.go`, `paths_linux.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A005

**Description**:
Implement PathResolver for Windows, macOS, and Linux using build tags.

**Steps**:
1. Create `paths_windows.go` with `//go:build windows`:
   - GetConfigDir → `%APPDATA%`
   - GetAppDataDir → `%LOCALAPPDATA%`
2. Create `paths_darwin.go` with `//go:build darwin`:
   - GetConfigDir → `~/Library/Application Support`
   - GetAppDataDir → `~/Library/Application Support`
3. Create `paths_linux.go` with `//go:build linux`:
   - GetConfigDir → `~/.config`
   - GetAppDataDir → `~/.local/share`
4. Add helper for `~/.mcpmanager/` expansion

**Acceptance**:
- Paths resolve correctly on respective platforms
- Unit test verifies path format per OS (use `runtime.GOOS`)
- No compilation errors on target platforms

---

### T-A007 [P] Implement platform-specific process managers
**Files**: `internal/platform/process_windows.go`, `process_unix.go`
**Effort**: L (2+ hours)
**Dependencies**: T-A005

**Description**:
Implement ProcessManager for Windows (taskkill) and Unix (SIGTERM/SIGKILL).

**Steps**:
1. Create `process_windows.go` (`//go:build windows`):
   - Start: Use `exec.Command`, capture PID
   - Stop: `taskkill /PID {pid} /F` if !graceful, else `/T`
   - IsRunning: Check process existence via WMI or tasklist
2. Create `process_unix.go` (`//go:build darwin || linux`):
   - Start: Use `exec.Command` with `SysProcAttr`
   - Stop: `syscall.Kill(pid, syscall.SIGTERM)` then SIGKILL after timeout
   - IsRunning: `syscall.Kill(pid, 0)` → errno check
3. Handle stdout/stderr capture via `io.Pipe`

**Acceptance**:
- Can start/stop test processes on each platform
- PID tracking works correctly
- Graceful shutdown honors timeout parameter
- Unit tests pass (mock process for testing)

---

### T-A008 [P] Implement single-instance enforcement
**Files**: `internal/platform/singleton_windows.go`, `singleton_unix.go`
**Effort**: M (1-2 hours)
**Dependencies**: T-A005

**Description**:
Implement SingleInstance mechanism per FR-051 (bring existing window to foreground).

**Steps**:
1. Windows (`singleton_windows.go`):
   - Use named mutex: `CreateMutex("Global\\MCPManager")`
   - On acquire failure, find window handle, call `SetForegroundWindow`
2. Unix (`singleton_unix.go`):
   - Use Unix socket lock: `/tmp/mcpmanager.lock`
   - On acquire failure, read existing PID, signal to show window

**Acceptance**:
- Second launch brings first window to foreground
- No duplicate processes created (verified via OS task manager)
- Lock released on application exit

---

