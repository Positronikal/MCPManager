# Specification Updates - 2025-10-15

## Summary of Changes

This document tracks the enhancements made to the MCP Manager specification based on developer feedback and clarifications.

---

## 1. Technology Stack Correction ✅

### Change: Fyne → Wails + Svelte

**Rationale:**
- **Developer Preference**: Hoyt explicitly prefers Wails + Svelte for this project
- **Toolchain Alignment**: Wails and Svelte already installed in development environment
- **Workflow Integration**: Enables Figma → Google Stitch → Svelte component pipeline
- **Constitutional Compliance**: Backend remains pure Go; Svelte frontend is consumer of Go API (API-first architecture)

**Files Updated:**
- `research.md` § 1 (GUI Framework Selection)
- `research.md` § 12 (Dependencies)
- `research.md` § 14 (Keyboard Shortcuts)
- `plan.md` (Technical Context, Project Structure, Task Planning)

**Impact:**
- Frontend tasks increased from 20-25 to 25-30 (Svelte components + stores + SSE client)
- Total task estimate: 107-133 (up from 95-120)
- Wails bindings generation added to Phase D
- Frontend testing with Svelte Testing Library added

---

## 2. MCP SDK Path Clarification ✅

### Change: "When Available" → Local Path Confirmed

**Clarification:**
- MCP Go SDK available locally at: `D:\dev\ARTIFICIAL_INTELLIGENCE\MCP\_MCP-Tools-Dev\go-sdk\`
- Fully up-to-date and ready for use
- No conditional logic needed ("when available" removed)

**Files Updated:**
- `research.md` § 2 (MCP Protocol Integration)
- `plan.md` (Constitution Check, Technical Context)

**Impact:**
- `go.mod` will use local module replacement: `replace github.com/modelcontextprotocol/go-sdk => ../../../_MCP-Tools-Dev/go-sdk`
- Setup task added: "Setup local MCP SDK module reference"
- Removed fallback implementation strategy

---

## 3. Architecture Diagrams Added ✅

### New Sections in research.md:

#### § 15: System Architecture (Component Diagram)
- Visual representation of system layers:
  - Frontend (Svelte UI) with components
  - Wails Bindings layer
  - Backend API (Go) with services
  - Event Bus (Pub/Sub)
  - External Integrations

**Purpose:**
- Clarifies separation of concerns
- Shows data flow between layers
- Illustrates event-driven architecture
- Helps developers understand system structure before implementation

#### § 16: Server Discovery Flow (Sequence Diagram)
- Step-by-step discovery process:
  - User triggers refresh
  - API initiates discovery
  - Config parsing
  - Filesystem scanning
  - Process matching
  - SSE event delivery
  - File watching setup

**Purpose:**
- Documents critical discovery algorithm
- Shows interaction between components
- Clarifies timing and sequencing
- Reduces implementation ambiguity

**Impact:**
- Developers can implement discovery service without reverse-engineering requirements
- Clear contract between Discovery Service and Config Parser
- File watching strategy explicitly documented

---

## 4. SSE Reconnection Strategy Documented ✅

### Enhancement to api-spec.yaml

**Added to `/events` endpoint:**

**Connection Management:**
- Client establishes persistent HTTP connection
- Server sends events as they occur (no polling)
- Connection kept alive with periodic heartbeat comments (every 15s)
- Automatic reconnection on disconnect (client responsibility)

**Reconnection Strategy:**
1. Client detects disconnect (connection close, timeout, error)
2. Client waits exponential backoff: 1s, 2s, 4s, 8s (max 30s)
3. Client reconnects with `Last-Event-ID` header if available
4. Server resends missed events based on `Last-Event-ID`
5. If `Last-Event-ID` unknown, server sends full state snapshot

**Event Format Specification:**
```
id: <event-uuid>
event: <event-type>
data: <json-payload>

```

**Heartbeat:**
- Server sends `: heartbeat` comment every 15s
- Client should close + reconnect if no data received for 45s

**Event Type Examples Added:**
- ServerDiscovered (with full JSON schema)
- ServerStatusChanged (with state transitions)
- ServerLogEntry (with severity)
- ConfigFileChanged (with file path)
- ServerMetricsUpdated (with metrics)

**Impact:**
- Frontend developers have complete SSE implementation guide
- No ambiguity about reconnection logic
- Prevents duplicate event processing
- Ensures robust connection management

---

## 5. Performance Benchmarking Tasks Added ✅

### Major Enhancement to quickstart.md

**New Performance Validation Sections:**

#### Startup Performance (Enhanced)
- **Methodology:** Cold start vs warm start, statistical significance (5 runs)
- **Pass Criteria:** Cold <2.5s, warm <1.5s, UI interactive <1s

#### Memory Usage (Enhanced)
- **Methodology:** RSS/VSZ measurement, baseline vs loaded states
- **Pass Criteria:** Baseline <50MB, 10 servers <75MB, 50 servers <100MB

#### Memory with 50 Servers (New)
- **Methodology:** Mock server generation, memory isolation
- **Pass Criteria:** Manager <150MB, total system <500MB, no leaks

#### UI Responsiveness (New)
- **Test Cases Table:**
  - Click Start button: <200ms
  - Switch log filter (server): <200ms
  - Switch log filter (severity): <100ms
  - Search logs (50k entries): <300ms
  - Resize window: <50ms (60 FPS)
  - Scroll log view: ≥60 FPS
  - Receive SSE event: <100ms
- **Automated Test Example:** Svelte Testing Library code snippet

#### Log Filtering Performance (New)
- **Methodology:** Generate 50k logs, measure filter operations
- **Pass Criteria:** Server filter <50ms, severity <30ms, full-text <300ms

#### Discovery Performance (New)
- **Methodology:** Populate test environment, measure scan time
- **Pass Criteria:** 50 servers <5s, config parsing <100ms, filesystem <3s

#### Event Stream Performance (New)
- **Methodology:** High-frequency event generation, latency measurement
- **Pass Criteria:** Delivery <100ms, UI update <50ms, no dropped events

#### Stress Test: 50 Simultaneous Server Starts (New)
- **Methodology:** Start all servers in parallel, monitor resources
- **Pass Criteria:** All started <30s, UI responsive, memory <200MB, CPU <50%

#### Long-Running Stability Test (New)
- **Methodology:** 24-hour test with continuous activity
- **Pass Criteria:** No crashes, memory growth <10MB, stable operations

#### Binary Size (New)
- **Methodology:** Measure production build size
- **Pass Criteria:** Binary <50MB, UPX compressed <20MB, installer <30MB

### Performance Regression Testing (New Section)

#### Continuous Benchmarking
- GitHub Actions workflow template
- Automated performance comparison
- Fail build if >10% degradation

#### Benchmark Baselines
- Reference hardware specification
- Baseline JSON format
- Storage location: `tests/performance/baselines/`

**Impact:**
- Comprehensive performance validation framework
- Automated CI/CD integration
- Clear acceptance criteria for optimization work
- Prevents performance regressions

---

## Files Changed Summary

| File | Sections Modified | Lines Added/Changed |
|------|-------------------|---------------------|
| `research.md` | §1, §2, §12, §14, +§15, +§16, References | ~250 lines |
| `api-spec.yaml` | `/events` endpoint | ~140 lines |
| `quickstart.md` | Performance Validation, +Regression Testing | ~350 lines |
| `plan.md` | Technical Context, Constitution Check, Project Structure, Task Planning | ~80 lines |
| `UPDATES.md` | *(this file)* | New file |

**Total Changes:** ~820 lines added/modified

---

## Constitutional Compliance Review

All changes maintain alignment with project constitution:

| Principle | Compliance Status |
|-----------|------------------|
| **Go Primary** | ✅ Backend in Go; Svelte is UI consumer (API-first) |
| **API-First** | ✅ Wails bindings call Go API; clear separation |
| **File-Based Storage** | ✅ No changes to storage strategy |
| **Cross-Platform** | ✅ Wails uses native WebView per platform |
| **Existing Libraries** | ✅ Local MCP SDK confirmed and integrated |
| **Performance Standards** | ✅ Enhanced with comprehensive benchmarks |

**Conclusion:** No constitutional violations introduced. All changes strengthen the specification.

---

## Impact on Phase 2 (Task Generation)

### Task Count Adjustment
- **Previous Estimate:** 95-120 tasks
- **New Estimate:** 107-133 tasks
- **Increase:** +12-13 tasks (~11% increase)

### New Task Categories
1. **Setup Tasks:**
   - Setup Wails v2.x project structure
   - Setup Svelte 4.x with TypeScript
   - Configure Wails bindings
   - Setup local MCP SDK module reference

2. **Frontend Tasks:**
   - Svelte component development (table, logs, config editor)
   - Svelte stores for state management
   - SSE client with exponential backoff reconnection
   - Theme system implementation (dark mode)
   - Frontend tests with Svelte Testing Library

3. **Performance Tasks:**
   - Implement performance benchmarks
   - Setup CI/CD performance validation
   - Baseline establishment on reference hardware
   - Optimization based on benchmark results

4. **Integration Tasks:**
   - Wails build configuration
   - Cross-platform packaging (Windows, macOS, Linux)
   - NSIS installer generation
   - UPX binary compression

### Dependencies Updated
- Frontend now depends on Wails bindings (Phase D)
- SSE client depends on SSE stream implementation
- Local MCP SDK must be set up before Discovery Service

---

## Next Steps

1. ✅ **Specification Update Complete** - All documents updated
2. ⏭️ **Ready for `/tasks` Command** - Generate implementation tasks
3. 📋 **Task Generation Will Produce:**
   - ~110-130 granular, actionable tasks
   - Proper dependency ordering (TDD + incremental)
   - Parallel execution markers where applicable
   - Acceptance criteria per task

---

## Notes for Task Generation

When running `/tasks`, ensure:
- Wails + Svelte setup tasks are in Phase A (foundation)
- Local MCP SDK setup is early in Phase A
- Frontend tasks are properly sequenced (stores → services → components)
- SSE reconnection logic is explicit in frontend tasks
- Performance benchmarks are integrated throughout (not just at end)
- Wails build + packaging tasks are in Phase F

---

## Revision History

- **2025-10-15 16:30 UTC** - Initial specification updates
  - Technology stack correction (Fyne → Wails + Svelte)
  - MCP SDK path clarification
  - Architecture diagrams added
  - SSE reconnection strategy documented
  - Performance benchmarking framework added

---

*This document serves as a change log for the specification enhancement phase. All changes were made collaboratively based on developer feedback and technical clarifications.*
