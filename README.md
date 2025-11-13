# MCP Manager
**GPLv3**

## What It Does
MCP Manager is a cross-platform desktop application for managing Model Context Protocol (MCP) servers. It provides centralized discovery, monitoring, and control of MCP servers across different clients (Claude Desktop, Cursor, etc.). Built with Go and Wails, MCP Manager offers a native desktop interface with real-time server status monitoring, log aggregation, configuration management, and lifecycle control.

Key capabilities:
- **Discovery**: Automatically finds MCP servers from client configurations, Claude Extensions, filesystem scans, and running processes
- **Lifecycle Management**: Start, stop, and restart MCP servers with transport-aware handling (stdio vs HTTP/SSE)
- **Monitoring**: Real-time log aggregation, resource metrics (CPU, memory), and status tracking
- **Configuration**: View and manage server settings, environment variables, and command-line arguments
- **Utilities**: Built-in tools for network analysis (Netstat), system services viewing, file explorer integration, and shell access
- **Dependencies**: Automatic detection of required runtime dependencies (Node.js, Python, etc.) and update monitoring

## Why It's Useful
Working with multiple MCP servers typically requires managing separate configuration files, terminal windows, and log outputs. MCP Manager consolidates this complexity into a single interface, allowing developers and system administrators to:
- See all MCP servers at a glance with their current status
- Monitor logs from multiple servers in one unified view
- Troubleshoot issues without switching between terminals
- Verify dependencies and server health before use
- Manage server configurations without manual file editing
- Track network connections and resource usage per server

Designed to handle up to 50 servers efficiently, MCP Manager is ideal for power users and team environments with complex MCP server deployments.

## How To Get Started
See **[USING.md](./USING.md)** for detailed installation and usage instructions.

Quick start:
1. Ensure you have Go 1.21+ installed
2. Install Wails v2: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
3. Clone this repository
4. Run `wails dev` for development mode, or `wails build` for production binary
5. Launch MCP Manager - it will automatically discover your installed MCP servers

## Where To Get Help
- **Documentation**: See the **[docs/](./docs)** directory for development documentation
- **Bug Reports**: See **[BUGS.md](./BUGS.md)** for bug tracking and reporting guidelines
- **Security Issues**: See **[SECURITY.md](./SECURITY.md)** for vulnerability reporting procedures
- **Contributing**: See **[CONTRIBUTING.md](./CONTRIBUTING.md)** for contribution guidelines
- **GitHub Issues**: Report bugs and request features via GitHub Issues
- **Specifications**: See **[specs/](./specs)** for feature specifications and implementation plans

## Adherence to Standards
This project adheres to the [Positronikal Coding Standards](https://github.com/positronikal/coding-standards/tree/main/standards). All contributors are expected to be familiar with these standards.

## Repository Structure Notes
This project follows **Go and Wails framework conventions** where they provide better tooling support and developer experience. While we maintain compatibility with Positronikal standards for documentation and security practices, the directory structure reflects idiomatic Go project layout and Wails requirements:

### Go/Wails Conventions Used
- **`cmd/`**: Application entry points (Go standard)
- **`internal/`**: Private application code (Go standard)
- **`pkg/`**: Public library code (Go standard)
- **`tests/`**: Test files and fixtures (Go convention - note the plural)
- **`frontend/`**: Svelte frontend code (Wails requirement - must be at root)
- **`build/`**: Wails build output and assets (Wails requirement)
- **Root `.go` files**: `app.go`, `main.go` for Wails bindings (Wails requirement)

### Positronikal Standard Directories Maintained
- **`docs/`**: Development documentation
- **`etc/`**: Scratch workspace for developers
- **`rel/`**: Release packages (alpha, beta, stable)
- **`ref/`**: Reference materials and future user manual content

### Special Directories
- **`.specify/`**: Spec Kit framework for specification-driven development
- **`specs/`**: Feature specifications and implementation plans
- **`.claude/`**: Claude Code configuration and slash commands
- **`.github/`**: GitHub Actions workflows and configuration

This hybrid approach allows the project to benefit from Go's excellent tooling (go modules, go test, gopls) and Wails' build system while maintaining Positronikal documentation and security standards.

## Repository Map
The following directories and files comprise the MCP Manager repository:

### Go Application Structure
**[cmd/](./cmd 'cmd/')**
- Application entry points and command-line interfaces
- `mcpmanager/` - Main application entry point

**[internal/](./internal 'internal/')**
- Private application code (not importable by other projects)
- `api/` - REST API handlers and routing
- `core/` - Core business logic services (discovery, lifecycle, monitoring, config, events)
- `models/` - Data structures and domain models
- `platform/` - Cross-platform abstractions (Windows, macOS, Linux)
- `storage/` - Application state persistence

**[pkg/](./pkg 'pkg/')**
- Public library code (potentially importable by other projects)

**[tests/](./tests 'tests/')**
- Test files organized by type:
- `unit/` - Unit tests for individual packages
- `contract/` - API contract validation tests
- `integration/` - Service interaction tests
- `performance/` - Benchmarks for startup time and memory usage

### Frontend Application
**[frontend/](./frontend 'frontend/')**
- Svelte 4.x frontend application
- `src/` - Source code (components, services, stores, types)
- `node_modules/` - NPM dependencies
- `public/` - Static assets
- `package.json` - Frontend dependencies and scripts

**[build/](./build 'build/')**
- Wails build output and platform-specific assets (gitignored)
- `bin/` - Compiled executables (development and testing)
- `windows/`, `darwin/`, `linux/` - Platform-specific build artifacts
- **Note**: This is the single source of truth for built executables during development

### Documentation & Standards
**[docs/](./docs 'docs/')**
- Development documentation for understanding and maintaining the project

**[specs/](./specs 'specs/')**
- Feature specifications using Spec Kit methodology
- `001-mcp-manager-specification/` - Main feature spec with requirements, plans, and tasks

**[ref/](./ref 'ref/')**
- Reference materials and external documentation

**[etc/](./etc 'etc/')**
- Developer scratch workspace, session notes, test reports, and temporary files

**[rel/](./rel 'rel/')**
- Release packages (alpha, beta, stable versions) for distribution
- Production installers and packaged executables belong here

**[scripts/](./scripts 'scripts/')**
- Build scripts, automation tools, and helper utilities

### Special Directories
**[.specify/](./.specify '.specify/')**
- Spec Kit framework for specification-driven development
- `memory/` - Project constitution and long-term memory
- `templates/` - Specification and plan templates
- `scripts/` - Automation for feature management

**[.claude/](./.claude '.claude/')**
- Claude Code configuration and custom slash commands
- Contains project-specific commands for AI-assisted development

**[.github/](./.github '.github/')**
- GitHub Actions CI/CD workflows
- Issue templates and repository configuration

### Go/Wails Build Files
**[app.go](./app.go 'app.go')**
- Wails application bindings exposing Go methods to frontend

**[main.go](./main.go 'main.go')**
- Application entry point

**[wails.json](./wails.json 'wails.json')**
- Wails configuration (frontend commands, build settings, app metadata)

**[go.mod](./go.mod 'go.mod')** / **[go.sum](./go.sum 'go.sum')**
- Go module dependencies

**[.editorconfig](./.editorconfig '.editorconfig')**
- Editor configuration for consistent code formatting

**[.gitattributes](./.gitattributes '.gitattributes')** / **[.gitignore](./.gitignore '.gitignore')**
- Git configuration for line endings and ignored files

### Standard Repository Files
**[ATTRIBUTION.md](./ATTRIBUTION.md 'ATTRIBUTION.md')**
- Credit to upstream projects and dependencies

**[AUTHORS.md](./AUTHORS.md 'AUTHORS.md')**
- Project contributors (human and AI)

**[BUGS.md](./BUGS.md 'BUGS.md')**
- Bug tracking and reporting process

**[CONTRIBUTING.md](./CONTRIBUTING.md 'CONTRIBUTING.md')**
- Contribution guidelines

**[COPYING.md](./COPYING.md 'COPYING.md')**
- GPLv3 license terms

**[SECURITY.md](./SECURITY.md 'SECURITY.md')**
- Security policy and vulnerability reporting

**[USING.md](./USING.md 'USING.md')**
- Installation and usage instructions

**[docs/PACKAGING.md](./docs/PACKAGING.md 'docs/PACKAGING.md')**
- Production build and packaging instructions (for maintainers)

**[docs/CLAUDE.md](./docs/CLAUDE.md 'docs/CLAUDE.md')**
- Guidance for Claude Code when working with this repository (for AI-assisted development)

**[README.md](./README.md 'README.md')**
- This document - project overview and navigation
