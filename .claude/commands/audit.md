# Audit Command

**Purpose**: Verify implementation quality and spec compliance for completed tasks

**Usage**: `@audit [phase] [task-range]`

## Instructions

When auditing, follow this process:

### 1. Read Complete Task Specifications

**CRITICAL**: Always read the FULL tasks.md file first:
```
Read file: D:\dev\ARTIFICIAL_INTELLIGENCE\MCP\MCPManager\specs\001-mcp-manager-specification\tasks.md
```

**Do NOT use search patterns.** Read the entire file to see complete specifications.

### 2. Verify Each Task

For each task in the specified range:

**Check Files:**
- ✅ All files listed in "File:" field exist
- ✅ Files are in correct locations
- ✅ No placeholder/stub implementations

**Check Implementation:**
- ✅ All steps in "Steps:" section completed
- ✅ Code matches specification examples
- ✅ No deviations from architecture

**Check Acceptance Criteria:**
- ✅ All items in "Acceptance:" section verified
- ✅ Tests pass (if tests specified)
- ✅ Edge cases handled

### 3. Run Tests

For phases with tests:
```bash
# Backend tests
go test ./internal/models/... -v
go test ./internal/core/... -v
go test ./tests/integration/... -v
go test ./tests/contract/... -v

# Frontend tests
cd frontend && npm run test
cd frontend && npm run check
```

### 4. Report Format

Use this exact format:

```markdown
# AUDIT REPORT - Phase [X]
**Date**: [YYYY-MM-DD]
**Scope**: Tasks [X-Y]
**Auditor**: Claude Code

## Executive Summary
- Total Tasks: [N]
- Complete: [N] (X%)
- Partial: [N] (X%)
- Incomplete: [N] (X%)

## Task Results

### T-X001 [Status] - Task Name
**Files**: ✅/⚠️/❌
**Implementation**: ✅/⚠️/❌
**Acceptance**: ✅/⚠️/❌
**Notes**: [Details if not complete]

[Repeat for each task]

## Critical Findings

### 🔴 HIGH PRIORITY
[List issues that block progress or violate requirements]

### 🟡 MEDIUM PRIORITY  
[List issues that should be fixed but don't block]

### 🟢 LOW PRIORITY
[List minor issues or optional improvements]

## Quality Assessment

- Code Organization: [Assessment]
- Test Coverage: [Percentage/Status]
- Spec Compliance: [Assessment]
- Architecture: [Assessment]

## Recommendations

1. [Immediate actions needed]
2. [Follow-up items]
3. [Optional improvements]

## Test Results

[Include test output summaries]
```

### 5. Status Indicators

**✅ COMPLETE**: All criteria met, production-ready
**⚠️ PARTIAL**: Functionality works but missing some requirements
**❌ INCOMPLETE**: Critical gaps, needs implementation

## Audit Triggers

Run audits at these milestones:

1. **End of each phase** (A→B, B→C, C→D, D→E, E→F)
2. **Mid-phase for long phases** (e.g., Phase E at task 15/30)
3. **Before critical dependencies** (e.g., before frontend if backend incomplete)
4. **After major architectural changes**
5. **At 25%, 50%, 75%, 100% project completion**

## Example Usage

```
@audit Phase-E T-E001:T-E015
```
This audits Phase E tasks 1 through 15.

```
@audit All-Phases
```
This audits all completed tasks across all phases.

## Critical Rules

1. **Always read full tasks.md** - Never rely on search patterns
2. **Verify files exist** - Check actual filesystem, not assumptions
3. **Run tests** - Don't just check if tests exist, run them
4. **Be thorough** - Quality > speed for audits
5. **Report honestly** - Partial is better than claiming complete

## Output Location

Save audit reports to:
```
D:\dev\ARTIFICIAL_INTELLIGENCE\MCP\MCPManager\etc\audit-report-YYYYMMDD.md
```

## Post-Audit Actions

After completing audit:
1. **Critical issues**: Fix immediately before proceeding
2. **Medium issues**: Add to backlog, fix before phase completion
3. **Low issues**: Document, fix if time permits

Never proceed to next phase with critical issues unresolved.
