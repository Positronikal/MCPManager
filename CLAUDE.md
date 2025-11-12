# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MCPManager is a desktop application for managing Model Context Protocol (MCP) servers. Built with Wails (Go backend + Svelte frontend), it discovers, monitors, and manages MCP servers across different clients (Claude Desktop, Cursor, etc.).

## Current Status

**Last Updated**: 2025-11-12

### Implementation Progress
- **Phase A-E**: Complete (88/88 tasks) ✅
- **Phase F**: 4/10 tasks complete 🔄
  - ✅ Initial setup and builds working
  - ❌ T-F001-F006: Integration tests (quickstart scenarios, edge cases)
  - ❌ T-F007: Performance benchmark - startup time (<2s target)
  - ❌ T-F008: Performance benchmark - memory usage (<100MB idle, 50 servers)
  - ❌ T-F009: CI/CD pipeline with GitHub Actions
  - ❌ T-F010: Production packaging for Windows/macOS/Linux

**Tasks**: See [tasks/](specs/001-mcp-manager-specification/tasks/) for modular phase files (90KB → 6 files)

### Known Incomplete Features
- **Netstat** (`NetstatView.svelte`): Shows mock data, backend API pending (T-E013)
- **Services** (`ServicesView.svelte`): Shows mock data, backend API pending (T-E016)
- Both components properly notify users that backend APIs aren't ready

### Recent Commits
- `ffc14e0` (2025-11-12): Fixed 7 manual test failures - stdio detection, Explorer/Shell APIs, UI fixes
- `55f0067`: Phase F Tasks 1-4 complete, build clean, ready for Option D UI
- `a18e35e`: Argument parsing and Svelte reactivity fixes

### Known Issues
- Explorer sometimes opens to Documents folder on Windows (functional but could be refined)
- Some contract tests failing (need investigation)

## Specification-Driven Development (Spec Kit)

This project follows a **specification-driven development** methodology using [Spec Kit](https://github.com/github/spec-kit), which emphasizes defining the "what" and "why" before the "how". Specifications are executable documents that directly guide implementation.

### Project Constitution

The project's governing principles are defined in `.specify/memory/constitution.md`. Key principles include:

- **Unix Philosophy**: Do one thing well, modularity, simplicity, composability
- **Language Selection**: Go as primary language for user-facing software
- **Architecture Principles**: API-first, cross-platform, state persistence, separation of concerns
- **MCP Server Expectations**: Self-managing servers with state isolation and robust lifecycle management
- **Scope Boundaries**: Focus on server management; avoid becoming an MCP client or protocol implementation
- **Quality Standards**: Testing requirements, documentation expectations, security practices
- **Performance Targets**: <2s startup, <100MB memory at idle, non-blocking UI operations

All development decisions must align with these constitutional principles. Use `/constitution` to review or update these principles.

### Specification Structure

Specifications live in `specs/[###-feature-name]/` directories:

```
specs/001-mcp-manager-specification/
├── spec.md              # Functional requirements (what & why)
├── plan.md              # Technical implementation plan (how)
├── tasks.md             # Task index (links to modular task files)
├── tasks/               # Modular task breakdown by phase
│   ├── phase-a-foundation.md
│   ├── phase-b-models.md
│   ├── phase-c-services.md
│   ├── phase-d-api.md
│   ├── phase-e-frontend.md
│   └── phase-f-testing.md
├── data-model.md        # Data structures and relationships
├── research.md          # Technology research and decisions
├── quickstart.md        # Getting started guide
├── contracts/           # API specifications (OpenAPI, etc.)
└── UPDATES.md           # Change history
```

### Spec Kit Workflow

The development workflow follows these phases:

1. **Constitution** (`/constitution`): Establish or review project principles
2. **Specify** (`/specify`): Define functional requirements and user stories (focus on WHAT and WHY)
3. **Clarify** (`/clarify`): Ask structured questions to resolve ambiguities before planning
4. **Plan** (`/plan`): Create technical implementation plan with chosen tech stack (focus on HOW)
5. **Tasks** (`/tasks`): Generate dependency-ordered, actionable task breakdown
6. **Implement** (`/implement`): Execute tasks systematically with checkpoints
7. **Analyze** (`/analyze`): Cross-artifact consistency and quality analysis

### Current Feature: 001-mcp-manager-specification

The main feature specification defines the core MCP Manager application:

- **User Story**: Centralized management for up to 50 MCP servers
- **Key Requirements**: Discovery (FR-001 to FR-020), Lifecycle Management (FR-021 to FR-030), Monitoring (FR-031 to FR-036), Configuration, Dependencies
- **Tech Stack**: Go 1.21+ backend, Wails v2.x for cross-platform GUI, Svelte 4.x frontend
- **Scale Target**: Support 50 servers efficiently (power user/team environment)
- **Clarifications**: Session notes in `spec.md` document key decisions (stdio server handling, config file watching, error states, log retention)

### Using Spec Kit Commands

When working on new features or enhancements:

1. Review the constitution first: `/constitution`
2. Create a new feature specification: `/specify <feature description>`
3. Clarify ambiguities: `/clarify` (before planning)
4. Generate technical plan: `/plan <technical details>`
5. Break down into tasks: `/tasks`
6. Validate consistency: `/analyze` (before implementing)
7. Execute implementation: `/implement`

The `.specify/` directory contains:
- `memory/`: Constitution and long-term project memory
- `templates/`: Spec, plan, and task templates
- `scripts/`: Automation scripts for feature management

### Feature Branches

Spec Kit uses a feature branch workflow:
- Each feature gets a numbered branch: `###-feature-name` (e.g., `001-mcp-manager-specification`)
- The branch number matches the spec directory: `specs/001-mcp-manager-specification/`
- Scripts in `.specify/scripts/` automate branch creation and feature setup
- All development for a feature happens on its branch
- Pull requests merge feature branches back to main

## Architecture

### Backend (Go)

The backend follows a service-oriented architecture with clear separation of concerns:

**Core Services** (`internal/core/`):
- **DiscoveryService**: Orchestrates multiple discovery sources (client configs, extensions, filesystem, running processes) to find MCP servers
- **LifecycleService**: Manages server start/stop/restart operations with graceful shutdown
- **ConfigService**: Handles server configuration management
- **MonitoringService**: Collects and manages server logs
- **MetricsCollector**: Gathers resource metrics (CPU, memory) for running servers
- **DependencyService**: Checks for required dependencies (Node.js, Python, etc.)
- **UpdateChecker**: Monitors for server updates

**EventBus** (`internal/core/events/`):
- Central pub/sub system connecting all services
- Events flow from services → EventBus → Wails runtime → Frontend
- Key events: `server.discovered`, `server.status.changed`, `server.log.entry`, `server.metrics.updated`, `config.file.changed`

**Platform Abstraction** (`internal/platform/`):
- Cross-platform implementations for Windows/macOS/Linux
- PathResolver, ProcessManager, ProcessInfo interfaces

**Models** (`internal/models/`):
- Core data structures: MCPServer, ServerStatus, ServerConfiguration, LogEntry, ServerMetrics
- TransportType enum: stdio (requires client), http/sse (standalone), unknown
- Deterministic UUID generation based on name+path+source for stable server IDs

**App Layer** (`app.go`, `cmd/mcpmanager/app.go`):
- Wails bindings exposing Go methods to frontend
- Pattern: App methods → Services → EventBus → Frontend events
- Startup sequence: EventBus → Storage → Discovery → Monitoring → Lifecycle → Config → Initial discovery
- Shutdown sequence: Stop servers → Close discovery (stop file watcher) → Close EventBus

### Frontend (Svelte)

**Structure** (`frontend/src/`):
- `App.svelte`: Main application component with routing
- `components/`: Reusable UI components (ServerTable, modals, etc.)
- `services/`: API wrappers and event handlers
- `stores/`: Svelte stores for state management
- `types/`: TypeScript type definitions

**Wails Integration**:
- Backend methods available via `wailsjs/go/main/App`
- Real-time events via `runtime.EventsOn()` from `@wailsapp/runtime`
- Event naming: Backend `"server:status:changed"` → Frontend listener

### Discovery Sources

1. **ClientConfigDiscovery**: Parses client config files (Claude Desktop, Cursor)
2. **ClaudeExtensionsDiscovery**: Finds servers from Claude Extensions
3. **FilesystemDiscovery**: Scans for MCP server installations
4. **ProcessDiscovery**: Detects running MCP processes
5. **ConfigFileWatcher**: Monitors config files for external changes (FR-050)

### Transport Handling

- **stdio servers**: Cannot be started directly; require client configuration. StartServer returns error `"stdio_requires_client"` with guidance to use config editor
- **http/sse/unknown servers**: Can be started/stopped directly by MCPManager

## Development Commands

### Backend (Go)

```bash
# Build the application
wails build

# Development mode (hot reload)
wails dev

# Run Go tests
go test ./...                                    # All tests
go test ./internal/core/discovery/...            # Specific package
go test -v ./tests/contract/...                  # Contract tests (verbose)
go test -run TestServerLifecycle ./tests/...     # Specific test

# Lint and format
go fmt ./...
go vet ./...

# Manage dependencies
go mod tidy
go mod download
```

### Frontend (Svelte)

```bash
cd frontend

# Install dependencies
npm install

# Development server (used by wails dev)
npm run dev

# Build for production
npm run build

# TypeScript check
npm run check

# Run tests
npm test                  # Run all tests
npm run test:ui           # Interactive test UI
npm run test:coverage     # With coverage report
```

### Testing Strategy

- **Unit tests**: Individual package functionality (`internal/*/..._test.go`)
- **Integration tests**: Service interactions (`tests/integration/`)
- **Contract tests**: API contract validation (`tests/contract/`)
- **Frontend tests**: Component testing with Vitest + Testing Library (`frontend/src/test/`)

## Key Implementation Patterns

### Event Flow
```
Service → EventBus.Publish() → EventBus.Subscribe() → Wails runtime.EventsEmit() → Frontend runtime.EventsOn()
```

### Service Initialization
All services receive dependencies via constructor injection. EventBus is injected into all services for event publishing.

### Server State Management
- Services update server objects directly
- After state change, call `discoveryService.UpdateServer()` to update cache
- Emit `runtime.EventsEmit("server:status:changed", ...)` to notify frontend

### Graceful Shutdown
The shutdown sequence ensures:
1. Running servers are stopped gracefully
2. File watchers are closed
3. EventBus channels are closed
4. No goroutine leaks

## Important Notes

- **Local Go SDK**: The project uses a local replace directive for `github.com/modelcontextprotocol/go-sdk` (see `go.mod` line 46)
- **Wails JSON Config**: Frontend commands defined in `wails.json`
- **Cross-platform**: Use platform abstraction layer (`internal/platform/`) for OS-specific code
- **Config watching**: FR-050 requires monitoring client config files for external changes
- **Deterministic IDs**: Server IDs are deterministic (hash of name+path+source) to survive app restarts

## Spec Kit Commands Reference

This repository uses **Spec Kit** for specification-driven development. Commands are available via `.claude/commands/`:

### Core Spec Kit Commands
- `/constitution`: Create or update project governing principles (`.specify/memory/constitution.md`)
- `/specify`: Define feature requirements and user stories (creates `specs/###-feature/spec.md`)
- `/clarify`: Structured clarification workflow to resolve ambiguities before planning
- `/plan`: Create technical implementation plan with tech stack decisions (creates `plan.md`, `data-model.md`, `research.md`, etc.)
- `/tasks`: Generate actionable, dependency-ordered task breakdown (creates `tasks.md`)
- `/implement`: Execute all tasks systematically according to the plan
- `/analyze`: Cross-artifact consistency and quality analysis
- `/audit`: Code quality audit (custom addition)

## Coding Standards

This project adheres to the [Positronikal Coding Standards](https://github.com/positronikal/coding-standards/tree/main/standards).
