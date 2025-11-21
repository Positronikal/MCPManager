# MCP Manager

![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)
![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue)
![Status](https://img.shields.io/badge/status-v1.0--rc-green)

**A cross-platform desktop application for managing Model Context Protocol servers**

---

## What It Does

MCP Manager is a native desktop application for managing Model Context Protocol (MCP) servers. It provides centralized discovery, monitoring, and control of MCP servers across different clients (Claude Desktop, Cursor, etc.). Built with Go and Wails, MCP Manager offers a unified interface for all your MCP server management needs.

### Features at a Glance

| Feature | Description |
|---------|-------------|
| 🔍 **Auto-Discovery** | Finds servers from client configs, extensions, filesystem, and running processes |
| 🎮 **Lifecycle Control** | Start, stop, restart with transport-aware handling (stdio/HTTP/SSE) |
| 📊 **Real-Time Monitoring** | Log aggregation, CPU/memory metrics, status tracking |
| ⚙️ **Configuration** | GUI-based editing of settings, environment variables, arguments |
| 🛠️ **Utilities** | Network analysis (Netstat), system services, file explorer, shell access |
| 📦 **Dependency Management** | Auto-detect Node.js, Python, and other runtime requirements |

## Why It's Useful

Managing multiple MCP servers means juggling config files, terminal windows, and log outputs. MCP Manager consolidates this complexity into a single interface.

**Before MCP Manager:**
- 🔍 Hunt through config files to find which servers are installed
- 📝 Track multiple terminal windows for logs
- 🔧 Manually edit JSON files for configuration changes
- ❓ No visibility into server health or resource usage

**With MCP Manager:**
- ✅ See all servers at a glance with real-time status
- ✅ Unified log viewer with filtering and search
- ✅ GUI-based configuration editing
- ✅ Resource monitoring and dependency validation
- ✅ One-click start/stop/restart operations

**Built for power users**: Efficiently handles up to 50 servers - ideal for developers and teams with complex MCP deployments.

## Prerequisites

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **Node.js 16+** - For frontend development (optional for building from source)
- **Wails v2** - Desktop application framework

**Platform-specific:**
- **Windows**: WebView2 (usually pre-installed on Windows 10+)
- **macOS**: macOS 10.13+ (High Sierra or later)
- **Linux**: webkit2gtk package (`sudo apt install webkit2gtk-4.0` on Debian/Ubuntu)

## Quick Start

```bash
# Install Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Clone and run
git clone https://github.com/Positronikal/MCPManager.git
cd MCPManager
wails dev
```

MCP Manager will automatically discover your installed MCP servers on first launch.

**For detailed installation and usage**, see [USING.md](./USING.md).

## Development Verification

Before committing changes, run local verification to ensure code quality:

```bash
# Unix/macOS/Linux
./scripts/verify-build.sh

# Windows
scripts\verify-build.bat

# Quick check (faster, skips race detection and integration tests)
./scripts/verify-build.sh --quick

# Skip the full Wails build (faster for iterative development)
./scripts/verify-build.sh --skip-build
```

These scripts replace GitHub Actions CI and run the same quality gates locally:
- Backend unit, integration, and contract tests
- Go formatting, vet, and staticcheck linting
- Frontend TypeScript checking and tests
- Build verification

See [USING.md](./USING.md) for more details on the verification workflow.

## Project Status

- ✅ **Core Features**: Complete (98/98 specification tasks)
- ✅ **Cross-Platform**: Windows, macOS, Linux
- ✅ **Transport Support**: stdio, HTTP, SSE
- ✅ **Test Coverage**: All tests passing (unit, contract, integration, performance)
- 📦 **Version**: v1.0.0-rc (Release Candidate)
- 📚 **Documentation**: Complete

## Community & Support

- 📖 **Documentation**: See the [docs/](./docs) directory for development documentation
- 🐛 **Bug Reports**: See [BUGS.md](./BUGS.md) for bug tracking and reporting guidelines
- 🔒 **Security**: See [SECURITY.md](./SECURITY.md) for vulnerability reporting procedures
- 🤝 **Contributing**: See [CONTRIBUTING.md](./CONTRIBUTING.md) for contribution guidelines
- 📋 **Specifications**: See [specs/](./specs) for feature specifications and implementation plans

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

---

*Last Updated: 2025-11-14*
