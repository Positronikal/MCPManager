# Implementation Plan: MCP Manager - Cross-Platform Server Management Application

**Branch**: `001-mcp-manager-specification` | **Date**: 2025-10-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/001-mcp-manager-specification/spec.md`

## Execution Flow (/plan command scope)
```
1. Load feature spec from Input path
   → If not found: ERROR "No feature spec at {path}"
2. Fill Technical Context (scan for NEEDS CLARIFICATION)
   → Detect Project Type from file system structure or context (web=frontend+backend, mobile=app+api)
   → Set Structure Decision based on project type
3. Fill the Constitution Check section based on the content of the constitution document.
4. Evaluate Constitution Check section below
   → If violations exist: Document in Complexity Tracking
   → If no justification possible: ERROR "Simplify approach first"
   → Update Progress Tracking: Initial Constitution Check
5. Execute Phase 0 → research.md
   → If NEEDS CLARIFICATION remain: ERROR "Resolve unknowns"
6. Execute Phase 1 → contracts, data-model.md, quickstart.md, agent-specific template file (e.g., `CLAUDE.md` for Claude Code, `.github/copilot-instructions.md` for GitHub Copilot, `GEMINI.md` for Gemini CLI, `QWEN.md` for Qwen Code or `AGENTS.md` for opencode).
7. Re-evaluate Constitution Check section
   → If new violations: Refactor design, return to Phase 1
   → Update Progress Tracking: Post-Design Constitution Check
8. Plan Phase 2 → Describe task generation approach (DO NOT create tasks.md)
9. STOP - Ready for /tasks command
```

**IMPORTANT**: The /plan command STOPS at step 7. Phases 2-4 are executed by other commands:
- Phase 2: /tasks command creates tasks.md
- Phase 3-4: Implementation execution (manual or via tools)

## Summary
MCP Manager is a cross-platform desktop application that provides centralized management for Model Context Protocol (MCP) servers. It enables developers and system administrators to discover, configure, monitor, and control up to 50 MCP servers from a unified interface. The system uses a backend API written in Go with a cross-platform GUI frontend, following an API-first architecture that decouples server management logic from the user interface. Key capabilities include automatic server discovery from client configuration files, real-time status monitoring, lifecycle control (start/stop/restart), configuration management, log viewing with severity filtering, and dependency checking.

## Technical Context
**Language/Version**: Go 1.21+ (backend), Svelte 4.x + TypeScript (frontend)
**Primary Dependencies**: 
- Backend: MCP Go SDK (local at `D:\dev\ARTIFICIAL_INTELLIGENCE\MCP\_MCP-Tools-Dev\go-sdk\`), fsnotify (file watching)
- Frontend: Wails v2.x (Go-to-web bridge), Svelte 4.x (UI framework), TypeScript (type safety)
**Storage**: File-based JSON persistence for application state, no external database required
**Testing**: Go standard testing (testing package) for backend, Svelte Testing Library for frontend, contract tests for API, integration tests for server lifecycle
**Target Platform**: Cross-platform desktop - Windows 10+, macOS 12+, Linux (Ubuntu 20.04+ LTS)
**Project Type**: Single project with API-first architecture (Go backend core + Svelte frontend decoupled via Wails)
**Performance Goals**: <2 second startup time, real-time status updates for 50 servers, <200ms UI response time
**Constraints**: <100MB memory at idle, non-blocking UI operations, 1000 log entries per server rolling retention
**Scale/Scope**: Efficiently manage up to 50 MCP servers simultaneously (power user/team environment)

## Constitution Check
*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Alignment with Core Principles

| Principle | Status | Evidence |
|-----------|--------|----------|
| **Unix Philosophy - Do One Thing Well** | ✅ PASS | MCP Manager focuses solely on server management; does not attempt to be an MCP client or implement the protocol |
| **Modularity** | ✅ PASS | API-first architecture ensures backend and frontend are independently testable and replaceable; Wails provides clean separation |
| **Language Selection - Go Primary** | ✅ PASS | Go 1.21+ backend (primary business logic); Svelte frontend is consumer of Go API (constitutional compliance) |
| **API-First Architecture** | ✅ PASS | Backend core exposes well-defined REST/SSE APIs before GUI implementation; Wails bindings decouple frontend from backend |
| **Cross-Platform Support** | ✅ PASS | Windows, macOS, and Linux support from day one; Wails uses native WebView (WebView2/WebKit) per platform |
| **State Persistence** | ✅ PASS | All application state persisted to file-based storage; no in-memory-only state (FR-041, FR-042) |
| **Performance Standards** | ✅ PASS | <2s startup (FR-037), <100MB idle memory (FR-039), non-blocking UI (FR-038) |
| **Scope Boundaries - Client Configs** | ✅ PASS | Explicitly out-of-scope per specification; read-only access to client configs (FR-019) |
| **Use Existing MCP Libraries** | ✅ PASS | Uses local MCP Go SDK at `D:\dev\ARTIFICIAL_INTELLIGENCE\MCP\_MCP-Tools-Dev\go-sdk\`; not implementing protocol from scratch |

### Initial Gate Result: **PASS**

No constitutional violations detected. Proceeding to Phase 0 research.

---

### Post-Design Gate Result: **PASS**

Re-evaluated after Phase 1 design artifacts (data-model.md, contracts, quickstart.md):

| Principle | Post-Design Status | Evidence |
|-----------|-------------------|----------|
| **Go Language Usage** | ✅ PASS | Backend in Go 1.21+; all contracts use Go-friendly types; data model maps to Go structs |
| **API-First Architecture** | ✅ PASS | OpenAPI 3.0 spec defines complete backend API with 20+ endpoints; Svelte frontend decoupled via Wails + REST/SSE |
| **File-Based Storage** | ✅ PASS | ApplicationState persists to JSON files (~/.mcpmanager/state.json); no database |
| **Cross-Platform** | ✅ PASS | Platform abstraction layer in design (internal/platform/); Wails uses native WebView per platform; contracts platform-agnostic |
| **Performance Targets** | ✅ PASS | Data model optimized (circular log buffer); quickstart validates <2s startup, <100MB memory with extensive benchmarks |
| **Read-Only Client Configs** | ✅ PASS | ServerConfiguration explicitly separates editable (MCP Manager) vs read-only (client configs) |
| **Local MCP SDK** | ✅ PASS | go.mod references local SDK path; pkg/mcpclient/ wraps SDK for server introspection |

**Design Validation**: No new constitutional violations introduced. Architecture remains aligned with all core principles. Ready to proceed to Phase 2 task planning.

## Project Structure

### Documentation (this feature)
```
specs/[###-feature]/
├── plan.md              # This file (/plan command output)
├── research.md          # Phase 0 output (/plan command)
├── data-model.md        # Phase 1 output (/plan command)
├── quickstart.md        # Phase 1 output (/plan command)
├── contracts/           # Phase 1 output (/plan command)
└── tasks.md             # Phase 2 output (/tasks command - NOT created by /plan)
```

### Source Code (repository root)
```
cmd/
├── mcpmanager/          # Main application entry point (Wails)
└── mcpmanagerd/         # Optional background daemon (future)

internal/
├── api/                 # Backend API layer (REST/SSE handlers)
├── core/                # Core business logic
│   ├── discovery/       # Server discovery logic
│   ├── lifecycle/       # Server lifecycle management
│   ├── config/          # Configuration management
│   └── monitoring/      # Monitoring & logging
├── models/              # Domain models
├── storage/             # State persistence layer
└── platform/            # Platform-specific abstractions

frontend/                # Svelte UI (Wails frontend)
├── src/
│   ├── components/      # Reusable UI components
│   ├── stores/          # Svelte stores (state management)
│   ├── services/        # API client services
│   ├── routes/          # Views/pages
│   └── App.svelte       # Root component
├── wailsjs/             # Wails-generated Go bindings
└── package.json         # NPM dependencies

tests/
├── contract/            # API contract tests (backend)
├── integration/         # Integration tests (backend)
├── unit/                # Unit tests (backend)
└── frontend/            # Frontend tests (Svelte Testing Library)

pkg/                     # Public reusable packages
└── mcpclient/           # MCP client wrapper (uses local SDK)
```

**Structure Decision**: Single project with API-first architecture using Wails. The `internal/` directory contains the Go backend (API + business logic), while `frontend/` contains the Svelte UI. Wails provides the bridge between Go and JavaScript, enabling native performance with web technologies. The `cmd/` directory provides application entry points following Go conventions.

## Phase 0: Outline & Research
1. **Extract unknowns from Technical Context** above:
   - For each NEEDS CLARIFICATION → research task
   - For each dependency → best practices task
   - For each integration → patterns task

2. **Generate and dispatch research agents**:
   ```
   For each unknown in Technical Context:
     Task: "Research {unknown} for {feature context}"
   For each technology choice:
     Task: "Find best practices for {tech} in {domain}"
   ```

3. **Consolidate findings** in `research.md` using format:
   - Decision: [what was chosen]
   - Rationale: [why chosen]
   - Alternatives considered: [what else evaluated]

**Output**: research.md with all NEEDS CLARIFICATION resolved

## Phase 1: Design & Contracts
*Prerequisites: research.md complete*

1. **Extract entities from feature spec** → `data-model.md`:
   - Entity name, fields, relationships
   - Validation rules from requirements
   - State transitions if applicable

2. **Generate API contracts** from functional requirements:
   - For each user action → endpoint
   - Use standard REST/GraphQL patterns
   - Output OpenAPI/GraphQL schema to `/contracts/`

3. **Generate contract tests** from contracts:
   - One test file per endpoint
   - Assert request/response schemas
   - Tests must fail (no implementation yet)

4. **Extract test scenarios** from user stories:
   - Each story → integration test scenario
   - Quickstart test = story validation steps

5. **Update agent file incrementally** (O(1) operation):
   - Run `.specify/scripts/powershell/update-agent-context.ps1 -AgentType claude`
     **IMPORTANT**: Execute it exactly as specified above. Do not add or remove any arguments.
   - If exists: Add only NEW tech from current plan
   - Preserve manual additions between markers
   - Update recent changes (keep last 3)
   - Keep under 150 lines for token efficiency
   - Output to repository root

**Output**: data-model.md, /contracts/*, failing tests, quickstart.md, agent-specific file

## Phase 2: Task Planning Approach
*This section describes what the /tasks command will do - DO NOT execute during /plan*

### Task Generation Strategy

The `/tasks` command will:

1. **Load Base Template**: Start with `.specify/templates/tasks-template.md`

2. **Extract Task Sources**:
   - **From `contracts/api-spec.yaml`**: Generate contract test tasks for each endpoint (20+ endpoints)
   - **From `data-model.md`**: Generate model implementation tasks for each entity (6 entities)
   - **From `quickstart.md`**: Generate integration test tasks for each acceptance scenario (8 scenarios)
   - **From `research.md`**: Generate infrastructure tasks (Fyne setup, file watcher, platform abstraction)

3. **Task Categories & Mapping**:

   | Source | Task Type | Example | Parallelizable |
   |--------|-----------|---------|----------------|
   | API Spec Endpoints | Contract Test | "Write failing contract test for GET /servers" | ✅ [P] |
   | Data Model Entities | Model Implementation | "Implement MCPServer model with validation" | ✅ [P] |
   | API Spec Endpoints | Service Implementation | "Implement discovery service (GET /servers)" | ❌ (depends on models) |
   | Quickstart Scenarios | Integration Test | "Integration test: Start server lifecycle" | ❌ (depends on implementation) |
   | Research Decisions | Infrastructure | "Setup Fyne GUI framework with dark theme" | ✅ [P] |
   | Cross-Cutting | Platform Layer | "Implement platform-specific path resolution" | ✅ [P] |

4. **Task Ordering Strategy** (TDD + Dependency):

   **Phase A: Foundation (Parallel)**
   - Setup Go module structure
   - Setup Wails v2.x + Svelte 4.x project
   - Implement platform abstraction interfaces
   - Write all contract tests (failing)
   - Setup local MCP SDK module reference

   **Phase B: Domain Models (Parallel)**
   - Implement each entity from data-model.md
   - Implement state machine logic (ServerStatus transitions)
   - Implement validation rules per entity

   **Phase C: Core Services (Sequential Dependencies)**
   - Discovery service (filesystem scanning, config parsing)
   - Lifecycle service (process management)
   - Configuration service (CRUD operations)
   - Monitoring service (log buffering, metrics)
   - Storage service (JSON persistence)

   **Phase D: API Layer (Parallel per Service)**
   - REST endpoint handlers (map to services)
   - SSE event stream implementation
   - API middleware (logging, error handling)
   - Wails bindings generation

   **Phase E: Frontend Implementation (Sequential)**
   - Svelte project setup (TypeScript, stores)
   - Main layout component
   - Server table component (reactive updates)
   - Log viewer component (filtering, search)
   - Configuration editor component (forms, validation)
   - Utility panels (Netstat, Shell, etc.)
   - SSE client integration (auto-reconnect)
   - Theme system (dark mode per FR-043)

   **Phase F: Integration & Testing**
   - End-to-end integration tests (from quickstart.md)
   - Performance validation tests
   - Cross-platform compatibility tests (CI matrix)
   - Wails build + packaging (Windows, macOS, Linux)

5. **Parallel Execution Markers**:
   - `[P]` denotes tasks that can run concurrently
   - Tasks without `[P]` must run after their dependencies complete
   - Example:
     ```
     1. [P] Setup Go module (go mod init)
     2. [P] Install Fyne (go get fyne.io/fyne/v2)
     3. [P] Write contract test for GET /servers
     4. Implement MCPServer model (depends on 1)
     5. Implement discovery service (depends on 4)
     ```

### Estimated Task Breakdown

| Category | Estimated Tasks | Notes |
|----------|----------------|-------|
| **Setup & Infrastructure** | 6-8 | Go module, Wails + Svelte, platform layer, file watcher, local SDK setup |
| **Contract Tests** | 20-25 | One per API endpoint (all failing initially) |
| **Domain Models** | 10-12 | 6 entities + validation + state machine |
| **Core Services** | 15-20 | Discovery, lifecycle, config, monitoring, storage |
| **API Layer** | 15-18 | REST handlers, SSE stream, middleware, Wails bindings |
| **Frontend Implementation** | 25-30 | Svelte components, stores, services, SSE client, theming |
| **Integration Tests** | 8-10 | From quickstart scenarios |
| **Performance & Polish** | 8-10 | Optimization, CI setup, Wails packaging, documentation |
| **TOTAL** | **107-133 tasks** | Large feature with full TDD coverage + frontend |

### Task Granularity Guidelines

- **Granular tasks**: Each task should be completable in 30 minutes to 2 hours
- **Atomic**: One clear deliverable per task (one test file, one model, one endpoint)
- **Testable**: Each task produces verifiable output (passing test, working endpoint, visible UI component)

### Dependencies to Enforce

```
Contract Tests → (can fail initially)
Models → Services → API Layer → Wails Bindings → Frontend Components
Platform Abstractions → Services (that use platform-specific features)
Storage Service → Application State Management
Monitoring Service → Log Viewer UI Component
SSE Stream → Frontend SSE Client → Svelte Stores (reactive updates)
Local MCP SDK → pkg/mcpclient wrapper → Discovery Service
```

### Output Format (tasks.md)

The `/tasks` command will generate `tasks.md` with:
- Numbered tasks (1 through ~100)
- Dependency relationships explicit (e.g., "Task 25 depends on tasks 10, 15")
- Parallel markers `[P]` for independent tasks
- Acceptance criteria per task (how to verify completion)
- Estimated effort (S/M/L: Small <1hr, Medium 1-2hr, Large >2hr)

**IMPORTANT**: This phase is executed by the `/tasks` command, NOT by /plan. The above is the PLAN for task generation, not the tasks themselves.

## Phase 3+: Future Implementation
*These phases are beyond the scope of the /plan command*

**Phase 3**: Task execution (/tasks command creates tasks.md)  
**Phase 4**: Implementation (execute tasks.md following constitutional principles)  
**Phase 5**: Validation (run tests, execute quickstart.md, performance validation)

## Complexity Tracking
*Fill ONLY if Constitution Check has violations that must be justified*

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |


## Progress Tracking
*This checklist is updated during execution flow*

**Phase Status**:
- [x] Phase 0: Research complete (/plan command) - ✅ `research.md` generated
- [x] Phase 1: Design complete (/plan command) - ✅ `data-model.md`, `contracts/api-spec.yaml`, `quickstart.md`, `CLAUDE.md` generated
- [x] Phase 2: Task planning complete (/plan command - describe approach only) - ✅ Task generation strategy documented above
- [x] Phase 3: Tasks generated (/tasks command) - ✅ `tasks.md` (98 tasks) generated
- [ ] Phase 4: Implementation complete
- [ ] Phase 5: Validation passed

**Gate Status**:
- [x] Initial Constitution Check: PASS
- [x] Post-Design Constitution Check: PASS
- [x] All NEEDS CLARIFICATION resolved (5 clarifications in spec.md)
- [x] Complexity deviations documented (none - no violations)

**Artifacts Generated**:
- ✅ `plan.md` (this file)
- ✅ `research.md` (14 research topics, 0 unresolved)
- ✅ `data-model.md` (6 entities, 4 enums, validation rules, state machines)
- ✅ `contracts/api-spec.yaml` (OpenAPI 3.0, 20+ endpoints, complete schemas)
- ✅ `quickstart.md` (8 acceptance scenarios, 8 edge cases, performance validation)
- ✅ `CLAUDE.md` (agent context for Claude Code)
- ✅ `tasks.md` (98 tasks across 6 phases)

**Ready for Next Step**: ✅ **Begin Phase 4: Implementation**

---
*Based on Constitution v1.0.0 - See `.specify/memory/constitution.md`*
