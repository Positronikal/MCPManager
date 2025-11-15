// Package main implements the MCP Manager desktop application.
//
// # MCP Manager
//
// MCP Manager is a desktop application for discovering, monitoring, and managing
// Model Context Protocol (MCP) servers across different MCP clients.
//
// # Architecture
//
// The application follows a service-oriented architecture with clear separation:
//
//   - **EventBus**: Central pub/sub system connecting all services
//   - **Discovery**: Finds MCP servers from client configs, extensions, filesystem, and processes
//   - **Lifecycle**: Manages server start/stop/restart operations
//   - **Monitoring**: Collects logs and metrics from running servers
//   - **Configuration**: Handles server configuration management
//   - **Storage**: Persists application state
//
// # Key Features
//
//   - Multi-source discovery (client configs, extensions, filesystem, processes)
//   - Lifecycle management for standalone servers (http/sse transports)
//   - Real-time monitoring with logs and metrics
//   - Client configuration editor for stdio servers
//   - Network connection monitoring (netstat integration)
//   - System services monitoring
//   - Cross-platform support (Windows, macOS, Linux)
//
// # Transport Types
//
// MCP servers use different transport mechanisms:
//
//   - **stdio**: Requires MCP client to start (managed via client config)
//   - **http/sse**: Standalone servers that can be started directly
//   - **unknown**: Transport not yet determined
//
// # Event Flow
//
// Services communicate via EventBus:
//
//	Service → EventBus.Publish() → EventBus.Subscribe() → Wails runtime.EventsEmit() → Frontend
//
// # Lifecycle
//
//   - **Startup**: Initialize EventBus → Storage → Discovery → Monitoring → Lifecycle → Config
//   - **Shutdown**: Stop servers → Close discovery (file watcher) → Close EventBus
//
// @author Positronikal
// @version 1.0.0
package main
