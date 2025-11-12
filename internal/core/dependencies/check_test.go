package dependencies

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/Positronikal/MCPManager/internal/models"
)

// MockCommandExecutor implements CommandExecutor for testing
type MockCommandExecutor struct {
	commands map[string]string // command -> output
	errors   map[string]error  // command -> error
}

func NewMockCommandExecutor() *MockCommandExecutor {
	return &MockCommandExecutor{
		commands: make(map[string]string),
		errors:   make(map[string]error),
	}
}

func (m *MockCommandExecutor) AddCommand(command string, output string) {
	m.commands[command] = output
}

func (m *MockCommandExecutor) AddError(command string, err error) {
	m.errors[command] = err
}

func (m *MockCommandExecutor) Execute(command string, args ...string) (string, error) {
	key := command
	if len(args) > 0 {
		key = command + " " + strings.Join(args, " ")
	}

	if err, exists := m.errors[key]; exists {
		return "", err
	}

	if output, exists := m.commands[key]; exists {
		return output, nil
	}

	return "", fmt.Errorf("command not found: %s", key)
}

func TestNewDependencyService(t *testing.T) {
	ds := NewDependencyService()

	if ds == nil {
		t.Fatal("Expected dependency service to be created")
	}

	if ds.executor == nil {
		t.Error("Expected executor to be set")
	}
}

func TestNewDependencyServiceWithExecutor(t *testing.T) {
	mock := NewMockCommandExecutor()
	ds := NewDependencyServiceWithExecutor(mock)

	if ds == nil {
		t.Fatal("Expected dependency service to be created")
	}

	if ds.executor != mock {
		t.Error("Expected custom executor to be set")
	}
}

func TestCheckDependencies_NilServer(t *testing.T) {
	ds := NewDependencyService()

	_, err := ds.CheckDependencies(nil)
	if err == nil {
		t.Error("Expected error for nil server")
	}
}

func TestCheckDependencies_NoDependencies(t *testing.T) {
	ds := NewDependencyService()

	server := &models.MCPServer{
		Name:         "test-server",
		Dependencies: []models.Dependency{},
	}

	deps, err := ds.CheckDependencies(server)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies, got %d", len(deps))
	}
}

func TestCheckRuntime_NodeJS(t *testing.T) {
	mock := NewMockCommandExecutor()
	mock.AddCommand("node --version", "v18.17.0")

	ds := NewDependencyServiceWithExecutor(mock)

	dep := models.Dependency{
		Name:            "node",
		Type:            models.DependencyRuntime,
		RequiredVersion: ">=16.0.0",
	}

	result := ds.CheckSingleDependency(dep)

	if result.DetectedVersion != "18.17.0" {
		t.Errorf("Expected detected version '18.17.0', got '%s'", result.DetectedVersion)
	}

	if !result.IsInstalled() {
		t.Error("Expected Node.js to be installed")
	}
}

func TestCheckRuntime_Python(t *testing.T) {
	mock := NewMockCommandExecutor()
	mock.AddCommand("python3 --version", "Python 3.11.4")

	ds := NewDependencyServiceWithExecutor(mock)

	dep := models.Dependency{
		Name:            "python",
		Type:            models.DependencyRuntime,
		RequiredVersion: ">=3.8",
	}

	result := ds.CheckSingleDependency(dep)

	if result.DetectedVersion != "3.11.4" {
		t.Errorf("Expected detected version '3.11.4', got '%s'", result.DetectedVersion)
	}

	if !result.IsInstalled() {
		t.Error("Expected Python to be installed")
	}
}

func TestCheckRuntime_PythonFallback(t *testing.T) {
	mock := NewMockCommandExecutor()
	mock.AddError("python3 --version", fmt.Errorf("not found"))
	mock.AddCommand("python --version", "Python 3.9.0")

	ds := NewDependencyServiceWithExecutor(mock)

	dep := models.Dependency{
		Name: "python",
		Type: models.DependencyRuntime,
	}

	result := ds.CheckSingleDependency(dep)

	if result.DetectedVersion != "3.9.0" {
		t.Errorf("Expected detected version '3.9.0', got '%s'", result.DetectedVersion)
	}
}

func TestCheckRuntime_Go(t *testing.T) {
	mock := NewMockCommandExecutor()
	mock.AddCommand("go version", "go version go1.21.0 linux/amd64")

	ds := NewDependencyServiceWithExecutor(mock)

	dep := models.Dependency{
		Name:            "go",
		Type:            models.DependencyRuntime,
		RequiredVersion: ">=1.20",
	}

	result := ds.CheckSingleDependency(dep)

	if result.DetectedVersion != "1.21.0" {
		t.Errorf("Expected detected version '1.21.0', got '%s'", result.DetectedVersion)
	}

	if !result.IsInstalled() {
		t.Error("Expected Go to be installed")
	}
}

func TestCheckRuntime_NotInstalled(t *testing.T) {
	mock := NewMockCommandExecutor()
	mock.AddError("node --version", fmt.Errorf("command not found"))

	ds := NewDependencyServiceWithExecutor(mock)

	dep := models.Dependency{
		Name:            "node",
		Type:            models.DependencyRuntime,
		RequiredVersion: ">=16.0.0",
	}

	result := ds.CheckSingleDependency(dep)

	if result.DetectedVersion != "" {
		t.Errorf("Expected no detected version, got '%s'", result.DetectedVersion)
	}

	if result.IsInstalled() {
		t.Error("Expected Node.js to not be installed")
	}

	if result.InstallationInstructions == "" {
		t.Error("Expected installation instructions to be provided")
	}
}

func TestCheckTool_npm(t *testing.T) {
	mock := NewMockCommandExecutor()
	mock.AddCommand("npm --version", "9.8.1")

	ds := NewDependencyServiceWithExecutor(mock)

	dep := models.Dependency{
		Name: "npm",
		Type: models.DependencyTool,
	}

	result := ds.CheckSingleDependency(dep)

	if result.DetectedVersion != "9.8.1" {
		t.Errorf("Expected detected version '9.8.1', got '%s'", result.DetectedVersion)
	}
}

func TestCheckTool_VersionFlag(t *testing.T) {
	mock := NewMockCommandExecutor()
	mock.AddError("git --version", fmt.Errorf("not found"))
	mock.AddCommand("git -v", "git version 2.40.0")

	ds := NewDependencyServiceWithExecutor(mock)

	dep := models.Dependency{
		Name: "git",
		Type: models.DependencyTool,
	}

	result := ds.CheckSingleDependency(dep)

	if result.DetectedVersion != "2.40.0" {
		t.Errorf("Expected detected version '2.40.0', got '%s'", result.DetectedVersion)
	}
}

func TestCheckTool_WhichFallback(t *testing.T) {
	mock := NewMockCommandExecutor()
	mock.AddError("mytool --version", fmt.Errorf("not found"))
	mock.AddError("mytool -v", fmt.Errorf("not found"))
	mock.AddError("mytool version", fmt.Errorf("not found"))

	// Mock 'which' or 'where' command
	if runtime.GOOS == "windows" {
		mock.AddCommand("where mytool", "C:\\path\\to\\mytool.exe")
	} else {
		mock.AddCommand("which mytool", "/usr/bin/mytool")
	}

	ds := NewDependencyServiceWithExecutor(mock)

	dep := models.Dependency{
		Name: "mytool",
		Type: models.DependencyTool,
	}

	result := ds.CheckSingleDependency(dep)

	if result.DetectedVersion != "unknown" {
		t.Errorf("Expected detected version 'unknown', got '%s'", result.DetectedVersion)
	}
}

func TestCheckEnvironment_Set(t *testing.T) {
	// Set environment variable for test
	envVar := "TEST_DEPENDENCY_VAR"
	os.Setenv(envVar, "test-value")
	defer os.Unsetenv(envVar)

	ds := NewDependencyService()

	dep := models.Dependency{
		Name: envVar,
		Type: models.DependencyEnvironment,
	}

	result := ds.CheckSingleDependency(dep)

	if result.DetectedVersion != "set" {
		t.Errorf("Expected detected version 'set', got '%s'", result.DetectedVersion)
	}

	if !result.IsInstalled() {
		t.Error("Expected environment variable to be installed")
	}
}

func TestCheckEnvironment_NotSet(t *testing.T) {
	ds := NewDependencyService()

	dep := models.Dependency{
		Name: "NONEXISTENT_VAR",
		Type: models.DependencyEnvironment,
	}

	result := ds.CheckSingleDependency(dep)

	if result.DetectedVersion != "" {
		t.Errorf("Expected no detected version, got '%s'", result.DetectedVersion)
	}

	if result.IsInstalled() {
		t.Error("Expected environment variable to not be installed")
	}
}

func TestExtractVersion(t *testing.T) {
	ds := NewDependencyService()

	tests := []struct {
		input    string
		expected string
	}{
		{"v18.17.0", "18.17.0"},
		{"18.17.0", "18.17.0"},
		{"Python 3.11.4", "3.11.4"},
		{"go version go1.21.0 linux/amd64", "1.21.0"},
		{"version 9.8.1", "9.8.1"},
		{"npm 9.8.1", "9.8.1"},
		{"git version 2.40.0", "2.40.0"},
		{"1.2.3-beta.1", "1.2.3-beta.1"},
		{"no version here", ""},
	}

	for _, tt := range tests {
		result := ds.extractVersion(tt.input)
		if result != tt.expected {
			t.Errorf("Input '%s': expected '%s', got '%s'", tt.input, tt.expected, result)
		}
	}
}

func TestGetInstallationInstructions_Runtime(t *testing.T) {
	ds := NewDependencyService()

	dep := models.Dependency{
		Name: "node",
		Type: models.DependencyRuntime,
	}

	instructions := ds.getInstallationInstructions(&dep)

	if instructions == "" {
		t.Error("Expected installation instructions to be provided")
	}

	// Check that instructions are platform-appropriate
	platform := runtime.GOOS
	switch platform {
	case "windows":
		if !strings.Contains(instructions, "winget") && !strings.Contains(instructions, "Download") {
			t.Error("Expected Windows-specific instructions")
		}
	case "darwin":
		if !strings.Contains(instructions, "brew") {
			t.Error("Expected macOS-specific instructions")
		}
	case "linux":
		if !strings.Contains(instructions, "apt-get") && !strings.Contains(instructions, "dnf") {
			t.Error("Expected Linux-specific instructions")
		}
	}
}

func TestGetInstallationInstructions_Tool(t *testing.T) {
	ds := NewDependencyService()

	dep := models.Dependency{
		Name: "npm",
		Type: models.DependencyTool,
	}

	instructions := ds.getInstallationInstructions(&dep)

	if instructions == "" {
		t.Error("Expected installation instructions to be provided")
	}

	// npm comes with Node.js
	if !strings.Contains(strings.ToLower(instructions), "node") {
		t.Error("Expected instructions to mention Node.js")
	}
}

func TestGetInstallationInstructions_Environment(t *testing.T) {
	ds := NewDependencyService()

	dep := models.Dependency{
		Name: "MY_ENV_VAR",
		Type: models.DependencyEnvironment,
	}

	instructions := ds.getInstallationInstructions(&dep)

	if instructions == "" {
		t.Error("Expected installation instructions to be provided")
	}

	if !strings.Contains(instructions, "MY_ENV_VAR") {
		t.Error("Expected instructions to mention the environment variable name")
	}
}

func TestCheckDependencies_Multiple(t *testing.T) {
	mock := NewMockCommandExecutor()
	mock.AddCommand("node --version", "v18.17.0")
	mock.AddCommand("npm --version", "9.8.1")
	mock.AddError("python3 --version", fmt.Errorf("not found"))
	mock.AddError("python --version", fmt.Errorf("not found"))

	ds := NewDependencyServiceWithExecutor(mock)

	// Set environment variable
	os.Setenv("API_KEY", "test-key")
	defer os.Unsetenv("API_KEY")

	server := &models.MCPServer{
		Name: "test-server",
		Dependencies: []models.Dependency{
			{
				Name:            "node",
				Type:            models.DependencyRuntime,
				RequiredVersion: ">=16.0.0",
			},
			{
				Name: "npm",
				Type: models.DependencyTool,
			},
			{
				Name:            "python",
				Type:            models.DependencyRuntime,
				RequiredVersion: ">=3.8",
			},
			{
				Name: "API_KEY",
				Type: models.DependencyEnvironment,
			},
		},
	}

	deps, err := ds.CheckDependencies(server)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(deps) != 4 {
		t.Fatalf("Expected 4 dependencies, got %d", len(deps))
	}

	// Check Node.js
	if !deps[0].IsInstalled() {
		t.Error("Expected Node.js to be installed")
	}
	if deps[0].DetectedVersion != "18.17.0" {
		t.Errorf("Expected Node.js version '18.17.0', got '%s'", deps[0].DetectedVersion)
	}

	// Check npm
	if !deps[1].IsInstalled() {
		t.Error("Expected npm to be installed")
	}

	// Check Python (not installed)
	if deps[2].IsInstalled() {
		t.Error("Expected Python to not be installed")
	}
	if deps[2].InstallationInstructions == "" {
		t.Error("Expected installation instructions for Python")
	}

	// Check API_KEY
	if !deps[3].IsInstalled() {
		t.Error("Expected API_KEY to be set")
	}
}

func TestCheckDependencies_VersionComparison(t *testing.T) {
	tests := []struct {
		name            string
		detectedVersion string
		requiredVersion string
		shouldSatisfy   bool
	}{
		{"Exact match", "18.17.0", "18.17.0", true},
		{"Greater than", "18.17.0", ">=16.0.0", true},
		{"Less than fails", "14.0.0", ">=16.0.0", false},
		{"Constraint satisfied", "3.11.4", ">=3.8", true},
		{"Constraint not satisfied", "3.7.0", ">=3.8", false},
		{"No requirement", "1.2.3", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh mock for each test
			mock := NewMockCommandExecutor()
			mock.AddCommand("testruntime --version", "v"+tt.detectedVersion)
			ds := NewDependencyServiceWithExecutor(mock)

			dep := models.Dependency{
				Name:            "testruntime",
				Type:            models.DependencyRuntime,
				RequiredVersion: tt.requiredVersion,
			}

			result := ds.CheckSingleDependency(dep)

			if result.IsInstalled() != tt.shouldSatisfy {
				t.Errorf("Version %s with requirement '%s': expected IsInstalled=%v, got %v (detected: %s)",
					tt.detectedVersion, tt.requiredVersion, tt.shouldSatisfy, result.IsInstalled(), result.DetectedVersion)
			}
		})
	}
}

func TestCheckDependencies_PreservesOriginal(t *testing.T) {
	mock := NewMockCommandExecutor()
	mock.AddCommand("node --version", "v18.17.0")

	ds := NewDependencyServiceWithExecutor(mock)

	server := &models.MCPServer{
		Name: "test-server",
		Dependencies: []models.Dependency{
			{
				Name:            "node",
				Type:            models.DependencyRuntime,
				RequiredVersion: ">=16.0.0",
			},
		},
	}

	// Check that original dependency is not modified
	originalRequired := server.Dependencies[0].RequiredVersion

	deps, err := ds.CheckDependencies(server)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Original should be unchanged
	if server.Dependencies[0].RequiredVersion != originalRequired {
		t.Error("Original dependency was modified")
	}

	// Result should have detected version
	if deps[0].DetectedVersion == "" {
		t.Error("Expected detected version in result")
	}
}
