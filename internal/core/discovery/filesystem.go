package discovery

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
	"github.com/hoytech/mcpmanager/internal/platform"
)

// FilesystemDiscovery discovers MCP servers from the filesystem
type FilesystemDiscovery struct {
	pathResolver platform.PathResolver
	eventBus     *events.EventBus
}

// NewFilesystemDiscovery creates a new filesystem discovery instance
func NewFilesystemDiscovery(pathResolver platform.PathResolver, eventBus *events.EventBus) *FilesystemDiscovery {
	return &FilesystemDiscovery{
		pathResolver: pathResolver,
		eventBus:     eventBus,
	}
}

// DiscoverFromFilesystem discovers servers from NPM, Python, and Go installations
func (fd *FilesystemDiscovery) DiscoverFromFilesystem() ([]models.MCPServer, error) {
	var allServers []models.MCPServer

	// Discover from NPM global packages
	npmServers, err := fd.discoverNPMServers()
	if err == nil {
		allServers = append(allServers, npmServers...)
	}

	// Discover from Python site-packages
	pythonServers, err := fd.discoverPythonServers()
	if err == nil {
		allServers = append(allServers, pythonServers...)
	}

	// Discover from Go binaries
	goServers, err := fd.discoverGoServers()
	if err == nil {
		allServers = append(allServers, goServers...)
	}

	// Publish discovery events
	for i := range allServers {
		if fd.eventBus != nil {
			fd.eventBus.Publish(events.ServerDiscoveredEvent(&allServers[i]))
		}
	}

	return allServers, nil
}

// discoverNPMServers discovers MCP servers from NPM global packages
func (fd *FilesystemDiscovery) discoverNPMServers() ([]models.MCPServer, error) {
	var servers []models.MCPServer

	// Get NPM global root
	cmd := exec.Command("npm", "root", "-g")
	output, err := cmd.Output()
	if err != nil {
		// NPM not installed or not in PATH
		return servers, err
	}

	npmRoot := strings.TrimSpace(string(output))
	if npmRoot == "" {
		return servers, fmt.Errorf("npm root returned empty")
	}

	// Check if directory exists
	if _, err := os.Stat(npmRoot); os.IsNotExist(err) {
		return servers, nil
	}

	// Scan for MCP server packages
	entries, err := os.ReadDir(npmRoot)
	if err != nil {
		return servers, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Match patterns for MCP servers
		if strings.Contains(strings.ToLower(name), "mcp") {
			serverPath := filepath.Join(npmRoot, name)

			// Try to find the main entry point
			packageJSONPath := filepath.Join(serverPath, "package.json")
			if _, err := os.Stat(packageJSONPath); err == nil {
				// Create server entry
				server := models.NewMCPServer(name, serverPath, models.DiscoveryFilesystem)
				server.Configuration.CommandLineArguments = []string{}
				servers = append(servers, *server)
			}
		}
	}

	return servers, nil
}

// discoverPythonServers discovers MCP servers from Python site-packages
func (fd *FilesystemDiscovery) discoverPythonServers() ([]models.MCPServer, error) {
	var servers []models.MCPServer

	// Get Python site-packages directory
	cmd := exec.Command("python", "-m", "site", "--user-site")
	output, err := cmd.Output()
	if err != nil {
		// Try python3
		cmd = exec.Command("python3", "-m", "site", "--user-site")
		output, err = cmd.Output()
		if err != nil {
			// Python not installed or not in PATH
			return servers, err
		}
	}

	sitePackages := strings.TrimSpace(string(output))
	if sitePackages == "" {
		return servers, fmt.Errorf("python site returned empty")
	}

	// Check if directory exists
	if _, err := os.Stat(sitePackages); os.IsNotExist(err) {
		return servers, nil
	}

	// Scan for MCP server packages
	entries, err := os.ReadDir(sitePackages)
	if err != nil {
		return servers, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Match patterns for MCP servers (excluding .dist-info directories)
		if strings.Contains(strings.ToLower(name), "mcp") && !strings.HasSuffix(name, ".dist-info") {
			serverPath := filepath.Join(sitePackages, name)

			// Create server entry
			server := models.NewMCPServer(name, serverPath, models.DiscoveryFilesystem)
			server.Configuration.CommandLineArguments = []string{"-m", name}
			servers = append(servers, *server)
		}
	}

	return servers, nil
}

// discoverGoServers discovers MCP servers from Go binaries
func (fd *FilesystemDiscovery) discoverGoServers() ([]models.MCPServer, error) {
	var servers []models.MCPServer

	// Determine GOPATH/bin or ~/go/bin
	goBinPath := os.Getenv("GOPATH")
	if goBinPath != "" {
		goBinPath = filepath.Join(goBinPath, "bin")
	} else {
		// Use default ~/go/bin
		homeDir := fd.pathResolver.GetUserHomeDir()
		if homeDir == "" {
			return servers, fmt.Errorf("could not determine home directory")
		}
		goBinPath = filepath.Join(homeDir, "go", "bin")
	}

	// Check if directory exists
	if _, err := os.Stat(goBinPath); os.IsNotExist(err) {
		return servers, nil
	}

	// Scan for MCP binaries
	entries, err := os.ReadDir(goBinPath)
	if err != nil {
		return servers, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()

		// Match patterns for MCP servers
		if strings.Contains(strings.ToLower(name), "mcp") {
			binaryPath := filepath.Join(goBinPath, name)

			// Create server entry
			server := models.NewMCPServer(name, binaryPath, models.DiscoveryFilesystem)
			server.Configuration.CommandLineArguments = []string{}
			servers = append(servers, *server)
		}
	}

	return servers, nil
}
