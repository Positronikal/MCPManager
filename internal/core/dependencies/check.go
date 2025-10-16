package dependencies

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"

	"github.com/hoytech/mcpmanager/internal/models"
)

// CommandExecutor allows for mocking command execution in tests
type CommandExecutor interface {
	Execute(command string, args ...string) (string, error)
}

// DefaultCommandExecutor implements CommandExecutor using exec.Command
type DefaultCommandExecutor struct{}

// Execute runs a command and returns its output
func (e *DefaultCommandExecutor) Execute(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

// DependencyService checks and validates server dependencies
type DependencyService struct {
	executor CommandExecutor
}

// NewDependencyService creates a new dependency service
func NewDependencyService() *DependencyService {
	return &DependencyService{
		executor: &DefaultCommandExecutor{},
	}
}

// NewDependencyServiceWithExecutor creates a service with a custom executor (for testing)
func NewDependencyServiceWithExecutor(executor CommandExecutor) *DependencyService {
	return &DependencyService{
		executor: executor,
	}
}

// CheckDependencies checks all dependencies for a server
func (ds *DependencyService) CheckDependencies(server *models.MCPServer) ([]models.Dependency, error) {
	if server == nil {
		return nil, fmt.Errorf("server cannot be nil")
	}

	if server.Dependencies == nil || len(server.Dependencies) == 0 {
		return []models.Dependency{}, nil
	}

	results := make([]models.Dependency, len(server.Dependencies))

	for i, dep := range server.Dependencies {
		// Create a copy of the dependency
		result := models.Dependency{
			Name:            dep.Name,
			Type:            dep.Type,
			RequiredVersion: dep.RequiredVersion,
		}

		// Check the dependency based on its type
		switch dep.Type {
		case models.DependencyRuntime:
			ds.checkRuntime(&result)
		case models.DependencyTool:
			ds.checkTool(&result)
		case models.DependencyEnvironment:
			ds.checkEnvironment(&result)
		case models.DependencyLibrary:
			ds.checkLibrary(&result)
		default:
			// Unknown type, leave as not installed
		}

		// Set installation instructions if not installed
		if result.DetectedVersion == "" {
			result.InstallationInstructions = ds.getInstallationInstructions(&result)
		}

		results[i] = result
	}

	return results, nil
}

// checkRuntime checks runtime dependencies (node, python, go)
func (ds *DependencyService) checkRuntime(dep *models.Dependency) {
	var command string
	var args []string

	name := strings.ToLower(dep.Name)

	switch {
	case strings.Contains(name, "node"):
		command = "node"
		args = []string{"--version"}
	case strings.Contains(name, "python"):
		// Try python3 first, then python
		version, err := ds.executor.Execute("python3", "--version")
		if err == nil {
			dep.DetectedVersion = ds.extractVersion(version)
			return
		}
		command = "python"
		args = []string{"--version"}
	case strings.Contains(name, "go"):
		command = "go"
		args = []string{"version"}
	default:
		// Unknown runtime - try with --version flag
		command = name
		args = []string{"--version"}
	}

	output, err := ds.executor.Execute(command, args...)
	if err == nil {
		dep.DetectedVersion = ds.extractVersion(output)
	}
}

// checkTool checks tool dependencies (npm, pip, etc)
func (ds *DependencyService) checkTool(dep *models.Dependency) {
	name := strings.ToLower(dep.Name)

	// First try to get version
	var versionOutput string
	var err error

	// Try common version flags
	versionFlags := []string{"--version", "-v", "version"}
	for _, flag := range versionFlags {
		output, e := ds.executor.Execute(name, flag)
		if e == nil {
			versionOutput = output
			err = nil
			break
		}
		err = e
	}

	if err == nil && versionOutput != "" {
		dep.DetectedVersion = ds.extractVersion(versionOutput)
		return
	}

	// If version check failed, try 'which' or 'where' to see if it exists
	var checkCmd string
	if runtime.GOOS == "windows" {
		checkCmd = "where"
	} else {
		checkCmd = "which"
	}

	output, err := ds.executor.Execute(checkCmd, name)
	if err == nil && output != "" {
		// Tool exists but version unknown
		dep.DetectedVersion = "unknown"
	}
}

// checkEnvironment checks environment variable dependencies
func (ds *DependencyService) checkEnvironment(dep *models.Dependency) {
	value := os.Getenv(dep.Name)
	if value != "" {
		dep.DetectedVersion = "set"
	}
}

// checkLibrary checks library dependencies (platform-specific)
func (ds *DependencyService) checkLibrary(dep *models.Dependency) {
	// This is platform-specific and complex
	// For now, we'll do a simple check using ldconfig on Linux
	// or checking common library paths

	name := strings.ToLower(dep.Name)

	switch runtime.GOOS {
	case "linux":
		// Try ldconfig -p to search for library
		output, err := ds.executor.Execute("ldconfig", "-p")
		if err == nil && strings.Contains(strings.ToLower(output), name) {
			dep.DetectedVersion = "found"
		}
	case "darwin":
		// On macOS, check common library paths
		paths := []string{
			"/usr/lib",
			"/usr/local/lib",
			"/opt/homebrew/lib",
		}
		for _, path := range paths {
			entries, err := os.ReadDir(path)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				if strings.Contains(strings.ToLower(entry.Name()), name) {
					dep.DetectedVersion = "found"
					return
				}
			}
		}
	case "windows":
		// On Windows, check System32 and common paths
		// This is simplified; real implementation would be more complex
		dep.DetectedVersion = "unknown"
	}
}

// extractVersion extracts version number from command output
func (ds *DependencyService) extractVersion(output string) string {
	// Common version patterns
	patterns := []string{
		`v?(\d+\.\d+\.\d+[-\w\.]*)`,     // v1.2.3 or 1.2.3
		`version\s+v?(\d+\.\d+\.\d+)`,   // "version 1.2.3"
		`(\d+\.\d+\.\d+)`,                // plain 1.2.3
		`v?(\d+\.\d+)`,                   // v1.2 or 1.2
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 1 {
			return strings.TrimPrefix(matches[1], "v")
		}
	}

	// If no version pattern found, return empty
	return ""
}

// getInstallationInstructions returns platform-specific installation instructions
func (ds *DependencyService) getInstallationInstructions(dep *models.Dependency) string {
	name := strings.ToLower(dep.Name)
	platform := runtime.GOOS

	switch dep.Type {
	case models.DependencyRuntime:
		return ds.getRuntimeInstructions(name, platform)
	case models.DependencyTool:
		return ds.getToolInstructions(name, platform)
	case models.DependencyEnvironment:
		return ds.getEnvironmentInstructions(dep.Name, platform)
	case models.DependencyLibrary:
		return ds.getLibraryInstructions(name, platform)
	default:
		return fmt.Sprintf("Please install %s manually.", dep.Name)
	}
}

// getRuntimeInstructions returns instructions for installing runtimes
func (ds *DependencyService) getRuntimeInstructions(name, platform string) string {
	instructions := map[string]map[string]string{
		"node": {
			"windows": "Download and install Node.js from https://nodejs.org/\n\nOr use a package manager:\n```\nwinget install OpenJS.NodeJS\n```",
			"darwin":  "Install using Homebrew:\n```bash\nbrew install node\n```\n\nOr download from https://nodejs.org/",
			"linux":   "Install using your package manager:\n```bash\n# Ubuntu/Debian\nsudo apt-get update\nsudo apt-get install nodejs\n\n# Fedora\nsudo dnf install nodejs\n```",
		},
		"python": {
			"windows": "Download and install Python from https://python.org/\n\nOr use a package manager:\n```\nwinget install Python.Python.3.12\n```",
			"darwin":  "Install using Homebrew:\n```bash\nbrew install python3\n```\n\nOr download from https://python.org/",
			"linux":   "Install using your package manager:\n```bash\n# Ubuntu/Debian\nsudo apt-get update\nsudo apt-get install python3\n\n# Fedora\nsudo dnf install python3\n```",
		},
		"go": {
			"windows": "Download and install Go from https://go.dev/dl/\n\nOr use a package manager:\n```\nwinget install GoLang.Go\n```",
			"darwin":  "Install using Homebrew:\n```bash\nbrew install go\n```\n\nOr download from https://go.dev/dl/",
			"linux":   "Install using your package manager:\n```bash\n# Ubuntu/Debian\nsudo apt-get update\nsudo apt-get install golang\n\n# Or download from https://go.dev/dl/\n```",
		},
	}

	// Find matching runtime
	for runtime, platformInstructions := range instructions {
		if strings.Contains(name, runtime) {
			if instr, ok := platformInstructions[platform]; ok {
				return instr
			}
			// Default to Linux instructions
			return platformInstructions["linux"]
		}
	}

	return fmt.Sprintf("Please install %s from its official website.", name)
}

// getToolInstructions returns instructions for installing tools
func (ds *DependencyService) getToolInstructions(name, platform string) string {
	instructions := map[string]map[string]string{
		"npm": {
			"windows": "npm is included with Node.js. Install Node.js from https://nodejs.org/",
			"darwin":  "npm is included with Node.js:\n```bash\nbrew install node\n```",
			"linux":   "npm is included with Node.js:\n```bash\nsudo apt-get install nodejs npm\n```",
		},
		"pip": {
			"windows": "pip is included with Python. Install Python from https://python.org/",
			"darwin":  "pip is included with Python:\n```bash\nbrew install python3\n```",
			"linux":   "Install pip:\n```bash\nsudo apt-get install python3-pip\n```",
		},
		"git": {
			"windows": "Download Git from https://git-scm.com/\n\nOr use:\n```\nwinget install Git.Git\n```",
			"darwin":  "Install Git:\n```bash\nbrew install git\n```",
			"linux":   "Install Git:\n```bash\nsudo apt-get install git\n```",
		},
	}

	for tool, platformInstructions := range instructions {
		if strings.Contains(name, tool) {
			if instr, ok := platformInstructions[platform]; ok {
				return instr
			}
		}
	}

	return fmt.Sprintf("Please install %s using your system's package manager.", name)
}

// getEnvironmentInstructions returns instructions for setting environment variables
func (ds *DependencyService) getEnvironmentInstructions(name, platform string) string {
	switch platform {
	case "windows":
		return fmt.Sprintf("Set the `%s` environment variable:\n\n1. Open System Properties > Environment Variables\n2. Add a new variable named `%s`\n3. Restart your terminal/application", name, name)
	case "darwin", "linux":
		return fmt.Sprintf("Set the `%s` environment variable:\n\n```bash\nexport %s=<value>\n```\n\nAdd to your shell profile (~/.bashrc or ~/.zshrc) to make it permanent.", name, name)
	default:
		return fmt.Sprintf("Please set the %s environment variable.", name)
	}
}

// getLibraryInstructions returns instructions for installing libraries
func (ds *DependencyService) getLibraryInstructions(name, platform string) string {
	switch platform {
	case "windows":
		return fmt.Sprintf("Install the %s library using vcpkg or download from the vendor's website.", name)
	case "darwin":
		return fmt.Sprintf("Install the %s library using Homebrew:\n```bash\nbrew install %s\n```", name, name)
	case "linux":
		return fmt.Sprintf("Install the %s library:\n```bash\nsudo apt-get install lib%s-dev\n```\n\nOr use your distribution's package manager.", name, name)
	default:
		return fmt.Sprintf("Please install the %s library.", name)
	}
}

// CheckSingleDependency checks a single dependency
func (ds *DependencyService) CheckSingleDependency(dep models.Dependency) models.Dependency {
	result := models.Dependency{
		Name:            dep.Name,
		Type:            dep.Type,
		RequiredVersion: dep.RequiredVersion,
	}

	switch dep.Type {
	case models.DependencyRuntime:
		ds.checkRuntime(&result)
	case models.DependencyTool:
		ds.checkTool(&result)
	case models.DependencyEnvironment:
		ds.checkEnvironment(&result)
	case models.DependencyLibrary:
		ds.checkLibrary(&result)
	}

	if result.DetectedVersion == "" {
		result.InstallationInstructions = ds.getInstallationInstructions(&result)
	}

	return result
}
