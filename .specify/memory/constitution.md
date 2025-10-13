# MCP Manager Constitution

## Project Identity
MCP Manager is a client-agnostic management utility for Model Context Protocol (MCP) servers, providing centralized discovery, configuration, monitoring, and lifecycle management.

## Core Engineering Principles

### Unix Philosophy Adherence
- **Do one thing well**: MCP Manager manages MCP servers; it does not attempt to be an MCP client or replace existing tooling
- **Modularity**: Each component should be independently testable and replaceable
- **Simplicity**: Prefer straightforward solutions over clever ones
- **Composability**: Design for integration with other tools and workflows

### Code Standards
- Follow GNU Coding Standards for all code contributions
- Adhere to procedural and structured programming paradigms
- Prioritize readability and maintainability over premature optimization

### Language Selection
- **Go**: Primary language for all user-facing software and high-level functionality
- **C**: Reserved only for low-level system interfaces if absolutely necessary
- **Python**: Scripting and tooling only (build scripts, utilities)
- **Bash/PowerShell**: Platform-specific automation scripts

### Architecture Principles
- **Separation of Concerns**: Backend (server management logic) and frontend (GUI) must be decoupled
- **Cross-Platform**: Support Windows, macOS, and Linux from day one
- **API-First**: Core functionality exposed via well-defined APIs before GUI implementation
- **State Management**: All state must be persisted reliably; avoid in-memory-only state

## MCP Server Expectations

### Server Design Imperatives
All MCP servers (including those we build) must be self-managing with:
- **State Isolation**: Multiple clients can connect without state interference
- **Context Preservation**: Session context maintained appropriately within connections
- **Robust Connection Management**: Handle connection failures, timeouts, and reconnections gracefully
- **Clean Lifecycle**: Proper startup, shutdown, and resource cleanup without external intervention

### MCP Manager's Role
MCP Manager **coordinates** servers but does not **compensate** for poorly-designed servers:
- Provides visibility and management capabilities
- Does not work around broken server implementations
- Encourages and enables proper server design through tooling and examples

## Scope Boundaries

### In Scope
1. **Discovery & Configuration**: GUI to view available servers, capabilities, and configure options
2. **Control, Monitoring, & Diagnostics**: Manual override for startup/shutdown, health checks, metrics, logging
3. **Dependency Management**: Installation, updates, prerequisite checking for servers

### Out of Scope (Initially)
1. **Client Configuration Management**: Individual MCP client configs (Claude Desktop, Cursor, etc.) are not managed to avoid coupling to external implementation changes
2. **MCP Protocol Implementation**: We use existing MCP libraries; we don't implement the protocol
3. **Server Development Tools**: Use Anthropic's MCP Inspector for server development/testing

## Quality Standards

### Testing
- Unit tests required for all business logic
- Integration tests for server discovery and lifecycle management
- Manual testing procedures documented for GUI components

### Documentation
- Every public API must have clear documentation
- User-facing features require end-user documentation
- Architecture decisions documented in ADR (Architecture Decision Records) format

### Security
- No hardcoded credentials or secrets
- Secure handling of server configuration data
- Principle of least privilege for all operations

## Development Workflow

### Version Control
- Git-based workflow with feature branches
- Meaningful commit messages following conventional commits format
- Pull requests required for all changes (even solo development for discipline)

### Release Management
- Semantic versioning (MAJOR.MINOR.PATCH)
- Changelog maintained for all releases
- Tagged releases with compiled binaries

## Constraints and Preferences

### Dependencies
- Minimize external dependencies
- Prefer standard library solutions when available
- All dependencies must be actively maintained and well-documented

### Performance
- Startup time: < 2 seconds
- UI responsiveness: No blocking operations on main thread
- Memory footprint: Reasonable for a management utility (target < 100MB at idle)

### Compatibility
- Support current LTS versions of target platforms
- Graceful degradation when features unavailable on older systems
- Clear communication of minimum requirements

## Governance

This constitution supersedes all other development practices and guides every technical decision in the MCP Manager project. All specifications, plans, and implementations must align with these principles.

**Version**: 1.0.0 | **Ratified**: 2025-10-11 | **Last Amended**: 2025-10-11
