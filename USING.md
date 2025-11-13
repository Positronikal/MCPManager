# Using MCP Manager

This document provides instructions for developers who want to modify or extend MCP Manager, as well as users who want to build the application from source.

## Prerequisites

### For Users (Building from Source)
- **Go 1.21+**: [Download Go](https://go.dev/dl/)
- **Node.js 18+**: [Download Node.js](https://nodejs.org/) (for frontend build)
- **Wails v2**: Install via `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Additional for Developers
- **Git**: For version control
- **Code Editor**: VS Code, GoLand, or similar with Go/Svelte support
- **Platform-specific tools**:
  - **Windows**: Build Tools, MinGW-w64
  - **macOS**: Xcode Command Line Tools
  - **Linux**: gcc, pkg-config, gtk3-dev, webkit2gtk-dev

## Quick Start

### Clone the Repository
```bash
git clone https://github.com/Positronikal/MCPManager.git
cd MCPManager
```

### Install Dependencies
```bash
# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..
```

### Development Mode
Run with hot reload for both backend and frontend:
```bash
wails dev
```

This will:
- Compile the Go backend
- Start the Svelte development server
- Launch the application with live reload
- Output logs to the terminal

### Production Build
Build standalone executables:
```bash
# Clean build (recommended)
wails build -clean

# Output will be in: build/bin/mcpmanager.exe (Windows)
#                    build/bin/mcpmanager (macOS/Linux)
```

For platform-specific builds and packaging instructions, see [docs/PACKAGING.md](./docs/PACKAGING.md).

## Project Structure

### Backend (Go)
- **`cmd/mcpmanager/`**: Application entry point (`main.go`, Wails bindings)
- **`internal/`**: Private application code
  - `api/`: REST API handlers and routing
  - `core/`: Business logic services (discovery, lifecycle, monitoring, config, events)
  - `models/`: Data structures
  - `platform/`: Cross-platform abstractions (Windows/macOS/Linux)
  - `storage/`: Application state persistence
- **`pkg/`**: Public library code (currently unused, reserved for future)

### Frontend (Svelte)
- **`frontend/src/`**:
  - `components/`: UI components (ServerTable, modals, utility views)
  - `services/`: API client and event handlers
  - `stores/`: Svelte stores for state management
  - `types/`: TypeScript type definitions
- **`frontend/public/`**: Static assets

### Tests
- **`tests/unit/`**: Unit tests for individual packages
- **`tests/contract/`**: API contract validation
- **`tests/integration/`**: Service interaction tests
- **`tests/performance/`**: Benchmarks (startup time, memory usage)

## Development Workflow

### Running Tests

**Backend Tests:**
```bash
# All tests
go test ./...

# Specific package
go test ./internal/core/discovery/...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./tests/contract/...
```

**Frontend Tests:**
```bash
cd frontend

# Run all tests
npm test

# Interactive test UI
npm run test:ui

# With coverage
npm run test:coverage
```

### Code Quality

**Backend (Go):**
```bash
# Format code
go fmt ./...

# Lint
go vet ./...

# Check dependencies
go mod tidy
```

**Frontend (Svelte/TypeScript):**
```bash
cd frontend

# TypeScript check
npm run check

# Format code (if configured)
npm run format

# Lint (if configured)
npm run lint
```

### Debugging

**Development Mode Logs:**
When running `wails dev`, both backend and frontend logs appear in the terminal:
- Backend: Structured logging via `slog` (INFO level by default)
- Frontend: Browser console logs

**Production Build Debugging:**
```bash
# Build with debug mode
wails build -debug

# This enables:
# - DevTools in production
# - Verbose logging
# - Source maps
```

## Architecture Overview

### Communication Flow
```
Frontend (Svelte) ←→ Wails IPC Bridge ←→ Go Backend
                                          ↓
                                     Services → EventBus → Real-time Events
```

- **Wails Bindings**: Frontend calls Go methods directly via `wailsjs/go/main/App`
- **Server-Sent Events (SSE)**: Real-time updates via `runtime.EventsOn()`
- **EventBus**: Central pub/sub system connecting all backend services

### Key Services
1. **DiscoveryService**: Finds MCP servers from multiple sources
2. **LifecycleService**: Manages start/stop/restart operations
3. **MonitoringService**: Collects logs and emits events
4. **MetricsCollector**: Gathers resource usage (CPU, memory)
5. **ConfigService**: Manages server configuration
6. **DependencyService**: Checks for Node.js, Python, etc.

### State Management
- **Backend**: Services maintain state, synchronized via `DiscoveryService.UpdateServer()`
- **Frontend**: Svelte stores (`$servers`, `$logs`, `$appState`)
- **Persistence**: Application state saved to `~/.mcpmanager/state.json`

## Common Development Tasks

### Adding a New API Endpoint

1. **Add handler** in `internal/api/` (e.g., `discovery.go`)
2. **Wire route** in `internal/api/router.go`
3. **Add frontend client method** in `frontend/src/services/api.ts`
4. **Add contract test** in `tests/contract/`
5. **Test manually** with `wails dev`

### Adding a New Frontend Component

1. **Create component** in `frontend/src/components/`
2. **Import in App.svelte** and add to routing
3. **Add any new types** to `frontend/src/types/`
4. **Write tests** in `frontend/src/test/`
5. **Run tests**: `cd frontend && npm test`

### Adding Platform-Specific Code

1. **Create interface** in `internal/platform/` (e.g., `platform.go`)
2. **Implement per-platform**:
   - `*_windows.go`
   - `*_darwin.go`
   - `*_linux.go`
3. **Use build tags** if needed (e.g., `//go:build windows`)
4. **Test on all platforms** via CI/CD or manual testing

## Troubleshooting

### Build Issues

**"wails: command not found"**
```bash
# Install Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Add to PATH (if needed)
export PATH=$PATH:$(go env GOPATH)/bin
```

**Frontend build fails**
```bash
# Clean and reinstall
cd frontend
rm -rf node_modules package-lock.json
npm install
cd ..
wails build -clean
```

**"go.mod" issues**
```bash
# Update dependencies
go mod tidy
go mod download
```

### Runtime Issues

**Application won't start**
- Check logs in terminal for errors
- Verify all dependencies installed: `go mod verify`
- Try clean build: `wails build -clean`

**Discovery not finding servers**
- Check client config files exist:
  - Windows: `%APPDATA%\Claude\claude_desktop_config.json`
  - macOS: `~/Library/Application Support/Claude/claude_desktop_config.json`
  - Linux: `~/.config/Claude/claude_desktop_config.json`
- Run with debug logging: `wails dev` (logs show discovery process)

**Tests failing**
- Ensure all dependencies installed: `go mod download && cd frontend && npm install`
- Check test output for specific errors
- Run individual tests: `go test -v ./path/to/package`

## Additional Resources

- **Developer Documentation**: See [docs/](./docs) directory (Doxygen-generated after completion)
- **Feature Specifications**: See [specs/](./specs) for detailed requirements
- **Bug Reporting**: See [BUGS.md](./BUGS.md)
- **Security**: See [SECURITY.md](./SECURITY.md)
- **Contributing**: See [CONTRIBUTING.md](./CONTRIBUTING.md)
- **Wails Documentation**: https://wails.io/docs/introduction
- **Go Documentation**: https://go.dev/doc/
- **Svelte Documentation**: https://svelte.dev/docs

## Building for Distribution

For instructions on creating installers and release packages, see [docs/PACKAGING.md](./docs/PACKAGING.md).

---

**Questions or Issues?** Open an issue on GitHub or see [BUGS.md](./BUGS.md) for reporting guidelines.
