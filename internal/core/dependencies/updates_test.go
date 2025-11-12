package dependencies

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/Positronikal/MCPManager/internal/models"
)

// MockHTTPClient implements HTTPClient for testing
type MockHTTPClient struct {
	responses map[string]*http.Response
	errors    map[string]error
}

func NewMockHTTPClient() *MockHTTPClient {
	return &MockHTTPClient{
		responses: make(map[string]*http.Response),
		errors:    make(map[string]error),
	}
}

func (m *MockHTTPClient) AddResponse(url string, statusCode int, body string) {
	m.responses[url] = &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}
}

func (m *MockHTTPClient) AddError(url string, err error) {
	m.errors[url] = err
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	if err, exists := m.errors[url]; exists {
		return nil, err
	}

	if resp, exists := m.responses[url]; exists {
		return resp, nil
	}

	return &http.Response{
		StatusCode: 404,
		Body:       io.NopCloser(bytes.NewBufferString("")),
	}, nil
}

func TestNewUpdateChecker(t *testing.T) {
	uc := NewUpdateChecker()

	if uc == nil {
		t.Fatal("Expected update checker to be created")
	}

	if uc.httpClient == nil {
		t.Error("Expected HTTP client to be set")
	}

	if uc.executor == nil {
		t.Error("Expected executor to be set")
	}
}

func TestCheckForUpdates_NilServer(t *testing.T) {
	uc := NewUpdateChecker()

	_, err := uc.CheckForUpdates(nil)
	if err == nil {
		t.Error("Expected error for nil server")
	}
}

func TestCheckForUpdates_NPM(t *testing.T) {
	mockHTTP := NewMockHTTPClient()
	mockExec := NewMockCommandExecutor()
	mockExec.AddCommand("npm view test-package version", "2.5.0\n")

	uc := NewUpdateCheckerWithClients(mockHTTP, mockExec)

	server := &models.MCPServer{
		ID:               "server1",
		Name:             "test-package",
		Version:          "2.3.0",
		InstallationPath: "/path/to/node_modules/test-package",
	}

	info, err := uc.CheckForUpdates(server)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if info.LatestVersion != "2.5.0" {
		t.Errorf("Expected latest version '2.5.0', got '%s'", info.LatestVersion)
	}

	if info.CurrentVersion != "2.3.0" {
		t.Errorf("Expected current version '2.3.0', got '%s'", info.CurrentVersion)
	}

	if !info.UpdateAvailable {
		t.Error("Expected update to be available")
	}

	if info.Status != UpdateStatusAvailable {
		t.Errorf("Expected status 'available', got '%s'", info.Status)
	}

	if info.PackageType != "npm" {
		t.Errorf("Expected package type 'npm', got '%s'", info.PackageType)
	}
}

func TestCheckForUpdates_NPM_Current(t *testing.T) {
	mockHTTP := NewMockHTTPClient()
	mockExec := NewMockCommandExecutor()
	mockExec.AddCommand("npm view test-package version", "2.3.0\n")

	uc := NewUpdateCheckerWithClients(mockHTTP, mockExec)

	server := &models.MCPServer{
		ID:               "server1",
		Name:             "test-package",
		Version:          "2.3.0",
		InstallationPath: "/path/to/node_modules/test-package",
	}

	info, err := uc.CheckForUpdates(server)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if info.UpdateAvailable {
		t.Error("Expected no update to be available")
	}

	if info.Status != UpdateStatusCurrent {
		t.Errorf("Expected status 'current', got '%s'", info.Status)
	}
}

func TestCheckForUpdates_PyPI(t *testing.T) {
	mockHTTP := NewMockHTTPClient()
	mockHTTP.AddResponse("https://pypi.org/pypi/test-package/json", 200, `{
		"info": {
			"version": "1.5.2"
		}
	}`)

	mockExec := NewMockCommandExecutor()

	uc := NewUpdateCheckerWithClients(mockHTTP, mockExec)

	server := &models.MCPServer{
		ID:               "server1",
		Name:             "test-package",
		Version:          "1.4.0",
		InstallationPath: "/path/to/site-packages/test-package",
	}

	info, err := uc.CheckForUpdates(server)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if info.LatestVersion != "1.5.2" {
		t.Errorf("Expected latest version '1.5.2', got '%s'", info.LatestVersion)
	}

	if !info.UpdateAvailable {
		t.Error("Expected update to be available")
	}

	if info.PackageType != "pypi" {
		t.Errorf("Expected package type 'pypi', got '%s'", info.PackageType)
	}
}

func TestCheckForUpdates_Go(t *testing.T) {
	mockHTTP := NewMockHTTPClient()
	mockHTTP.AddResponse("https://proxy.golang.org/github.com/example/module/@latest", 200, `{
		"Version": "v1.3.0"
	}`)

	mockExec := NewMockCommandExecutor()

	uc := NewUpdateCheckerWithClients(mockHTTP, mockExec)

	server := &models.MCPServer{
		ID:               "server1",
		Name:             "github.com/example/module",
		Version:          "1.2.0",
		InstallationPath: "/go/pkg/mod/github.com/example/module",
	}

	info, err := uc.CheckForUpdates(server)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if info.LatestVersion != "1.3.0" {
		t.Errorf("Expected latest version '1.3.0', got '%s'", info.LatestVersion)
	}

	if !info.UpdateAvailable {
		t.Error("Expected update to be available")
	}

	if info.PackageType != "go" {
		t.Errorf("Expected package type 'go', got '%s'", info.PackageType)
	}
}

func TestCheckForUpdates_NetworkError(t *testing.T) {
	mockHTTP := NewMockHTTPClient()
	mockHTTP.AddError("https://pypi.org/pypi/test-package/json", fmt.Errorf("network error"))

	mockExec := NewMockCommandExecutor()

	uc := NewUpdateCheckerWithClients(mockHTTP, mockExec)

	server := &models.MCPServer{
		ID:               "server1",
		Name:             "test-package",
		Version:          "1.0.0",
		InstallationPath: "/path/to/site-packages/test-package",
	}

	info, err := uc.CheckForUpdates(server)

	// Should not return error, but status should be unknown
	if err != nil {
		t.Errorf("Expected no error (graceful handling), got %v", err)
	}

	if info.Status != UpdateStatusUnknown {
		t.Errorf("Expected status 'unknown', got '%s'", info.Status)
	}

	if info.UpdateAvailable {
		t.Error("Expected update not available on network error")
	}
}

func TestCheckForUpdates_HTTPError(t *testing.T) {
	mockHTTP := NewMockHTTPClient()
	mockHTTP.AddResponse("https://pypi.org/pypi/test-package/json", 404, "Not Found")

	mockExec := NewMockCommandExecutor()

	uc := NewUpdateCheckerWithClients(mockHTTP, mockExec)

	server := &models.MCPServer{
		ID:               "server1",
		Name:             "test-package",
		Version:          "1.0.0",
		InstallationPath: "/path/to/site-packages/test-package",
	}

	info, err := uc.CheckForUpdates(server)

	// Should not return error, but status should be unknown
	if err != nil {
		t.Errorf("Expected no error (graceful handling), got %v", err)
	}

	if info.Status != UpdateStatusUnknown {
		t.Errorf("Expected status 'unknown', got '%s'", info.Status)
	}
}

func TestCheckForUpdates_NoPackageName(t *testing.T) {
	uc := NewUpdateChecker()

	server := &models.MCPServer{
		ID:               "server1",
		Name:             "", // Empty name
		Version:          "1.0.0",
		InstallationPath: "/some/path",
	}

	info, err := uc.CheckForUpdates(server)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if info.Status != UpdateStatusUnknown {
		t.Errorf("Expected status 'unknown', got '%s'", info.Status)
	}
}

func TestDetectPackageType(t *testing.T) {
	uc := NewUpdateChecker()

	tests := []struct {
		name         string
		server       *models.MCPServer
		expectedType string
	}{
		{
			name: "NPM from path",
			server: &models.MCPServer{
				InstallationPath: "/path/to/node_modules/package",
			},
			expectedType: "npm",
		},
		{
			name: "PyPI from path",
			server: &models.MCPServer{
				InstallationPath: "/path/to/site-packages/package",
			},
			expectedType: "pypi",
		},
		{
			name: "Go from path",
			server: &models.MCPServer{
				InstallationPath: "/go/pkg/mod/example.com/module",
			},
			expectedType: "go",
		},
		{
			name: "NPM from dependency",
			server: &models.MCPServer{
				InstallationPath: "/some/path",
				Dependencies: []models.Dependency{
					{Name: "node", Type: models.DependencyRuntime},
				},
			},
			expectedType: "npm",
		},
		{
			name: "PyPI from dependency",
			server: &models.MCPServer{
				InstallationPath: "/some/path",
				Dependencies: []models.Dependency{
					{Name: "python", Type: models.DependencyRuntime},
				},
			},
			expectedType: "pypi",
		},
		{
			name: "Unknown",
			server: &models.MCPServer{
				InstallationPath: "/unknown/path",
			},
			expectedType: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packageType := uc.detectPackageType(tt.server)
			if packageType != tt.expectedType {
				t.Errorf("Expected package type '%s', got '%s'", tt.expectedType, packageType)
			}
		})
	}
}

func TestIsUpdateAvailable(t *testing.T) {
	uc := NewUpdateChecker()

	tests := []struct {
		current  string
		latest   string
		expected bool
	}{
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "1.0.1", true},
		{"1.0.0", "2.0.0", true},
		{"2.0.0", "1.0.0", false},
		{"v1.0.0", "v1.1.0", true},
		{"1.0.0-beta", "1.0.0", true},
		{"1.0.0", "1.0.0-beta", false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s to %s", tt.current, tt.latest), func(t *testing.T) {
			result, err := uc.isUpdateAvailable(tt.current, tt.latest)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestCheckForUpdatesMultiple(t *testing.T) {
	mockHTTP := NewMockHTTPClient()
	mockHTTP.AddResponse("https://pypi.org/pypi/package1/json", 200, `{
		"info": {"version": "2.0.0"}
	}`)
	mockHTTP.AddResponse("https://pypi.org/pypi/package2/json", 200, `{
		"info": {"version": "1.0.0"}
	}`)

	mockExec := NewMockCommandExecutor()

	uc := NewUpdateCheckerWithClients(mockHTTP, mockExec)

	servers := []*models.MCPServer{
		{
			ID:               "server1",
			Name:             "package1",
			Version:          "1.5.0",
			InstallationPath: "/site-packages/package1",
		},
		{
			ID:               "server2",
			Name:             "package2",
			Version:          "1.0.0",
			InstallationPath: "/site-packages/package2",
		},
	}

	results := uc.CheckForUpdatesMultiple(servers)

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results["server1"].UpdateAvailable != true {
		t.Error("Expected update available for server1")
	}

	if results["server2"].UpdateAvailable != false {
		t.Error("Expected no update for server2")
	}
}

func TestCheckForUpdatesMultiple_WithNil(t *testing.T) {
	uc := NewUpdateChecker()

	servers := []*models.MCPServer{
		{
			ID:               "server1",
			Name:             "package1",
			Version:          "1.0.0",
			InstallationPath: "/some/path",
		},
		nil, // Nil server should be skipped
		{
			ID:               "server2",
			Name:             "package2",
			Version:          "1.0.0",
			InstallationPath: "/some/path",
		},
	}

	results := uc.CheckForUpdatesMultiple(servers)

	// Should only have 2 results (skipping nil)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestGetUpdateSummary(t *testing.T) {
	mockHTTP := NewMockHTTPClient()
	mockHTTP.AddResponse("https://pypi.org/pypi/package1/json", 200, `{
		"info": {"version": "2.0.0"}
	}`)
	mockHTTP.AddResponse("https://pypi.org/pypi/package2/json", 200, `{
		"info": {"version": "1.0.0"}
	}`)
	mockHTTP.AddResponse("https://pypi.org/pypi/package3/json", 404, "Not Found")

	mockExec := NewMockCommandExecutor()

	uc := NewUpdateCheckerWithClients(mockHTTP, mockExec)

	servers := []*models.MCPServer{
		{
			ID:               "server1",
			Name:             "package1",
			Version:          "1.5.0",
			InstallationPath: "/site-packages/package1",
		},
		{
			ID:               "server2",
			Name:             "package2",
			Version:          "1.0.0",
			InstallationPath: "/site-packages/package2",
		},
		{
			ID:               "server3",
			Name:             "package3",
			Version:          "1.0.0",
			InstallationPath: "/site-packages/package3",
		},
	}

	summary := uc.GetUpdateSummary(servers)

	if summary[UpdateStatusAvailable] != 1 {
		t.Errorf("Expected 1 available update, got %d", summary[UpdateStatusAvailable])
	}

	if summary[UpdateStatusCurrent] != 1 {
		t.Errorf("Expected 1 current, got %d", summary[UpdateStatusCurrent])
	}

	if summary[UpdateStatusUnknown] != 1 {
		t.Errorf("Expected 1 unknown, got %d", summary[UpdateStatusUnknown])
	}
}

func TestCheckForUpdates_VersionParsing(t *testing.T) {
	mockHTTP := NewMockHTTPClient()
	mockExec := NewMockCommandExecutor()

	tests := []struct {
		name            string
		current         string
		latest          string
		shouldUpdate    bool
	}{
		{"Patch update", "1.2.3", "1.2.4", true},
		{"Minor update", "1.2.3", "1.3.0", true},
		{"Major update", "1.2.3", "2.0.0", true},
		{"No update", "1.2.3", "1.2.3", false},
		{"Downgrade", "2.0.0", "1.9.9", false},
		{"With v prefix", "v1.0.0", "v1.1.0", true},
		{"Mixed prefix", "1.0.0", "v1.1.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockExec.AddCommand("npm view test-pkg version", tt.latest+"\n")
			uc := NewUpdateCheckerWithClients(mockHTTP, mockExec)

			server := &models.MCPServer{
				ID:               "server1",
				Name:             "test-pkg",
				Version:          tt.current,
				InstallationPath: "/node_modules/test-pkg",
			}

			info, err := uc.CheckForUpdates(server)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if info.UpdateAvailable != tt.shouldUpdate {
				t.Errorf("Expected UpdateAvailable=%v, got %v", tt.shouldUpdate, info.UpdateAvailable)
			}
		})
	}
}
