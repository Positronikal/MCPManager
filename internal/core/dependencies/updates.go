package dependencies

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/Positronikal/MCPManager/internal/models"
)

// UpdateStatus represents the update availability status
type UpdateStatus string

const (
	UpdateStatusAvailable UpdateStatus = "available"
	UpdateStatusCurrent   UpdateStatus = "current"
	UpdateStatusUnknown   UpdateStatus = "unknown"
)

// UpdateInfo contains information about available updates
type UpdateInfo struct {
	UpdateAvailable bool         `json:"updateAvailable"`
	Status          UpdateStatus `json:"status"`
	CurrentVersion  string       `json:"currentVersion"`
	LatestVersion   string       `json:"latestVersion"`
	ReleaseNotes    string       `json:"releaseNotes,omitempty"`
	PackageName     string       `json:"packageName"`
	PackageType     string       `json:"packageType"` // npm, pypi, go
}

// HTTPClient interface for mocking HTTP requests
type HTTPClient interface {
	Get(url string) (*http.Response, error)
}

// DefaultHTTPClient implements HTTPClient using http.DefaultClient
type DefaultHTTPClient struct {
	client *http.Client
}

// NewDefaultHTTPClient creates a new HTTP client with timeout
func NewDefaultHTTPClient() *DefaultHTTPClient {
	return &DefaultHTTPClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Get performs an HTTP GET request
func (c *DefaultHTTPClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

// UpdateChecker checks for package updates
type UpdateChecker struct {
	httpClient HTTPClient
	executor   CommandExecutor
}

// NewUpdateChecker creates a new update checker
func NewUpdateChecker() *UpdateChecker {
	return &UpdateChecker{
		httpClient: NewDefaultHTTPClient(),
		executor:   &DefaultCommandExecutor{},
	}
}

// NewUpdateCheckerWithClients creates an update checker with custom clients (for testing)
func NewUpdateCheckerWithClients(httpClient HTTPClient, executor CommandExecutor) *UpdateChecker {
	return &UpdateChecker{
		httpClient: httpClient,
		executor:   executor,
	}
}

// CheckForUpdates checks if updates are available for a server
func (uc *UpdateChecker) CheckForUpdates(server *models.MCPServer) (*UpdateInfo, error) {
	if server == nil {
		return nil, fmt.Errorf("server cannot be nil")
	}

	// Determine package type from installation path or metadata
	packageType := uc.detectPackageType(server)

	info := &UpdateInfo{
		CurrentVersion: server.Version,
		PackageName:    uc.extractPackageName(server),
		PackageType:    packageType,
		Status:         UpdateStatusUnknown,
	}

	if info.PackageName == "" {
		return info, nil
	}

	var err error
	switch packageType {
	case "npm":
		err = uc.checkNPMUpdate(info)
	case "pypi":
		err = uc.checkPyPIUpdate(info)
	case "go":
		err = uc.checkGoUpdate(info)
	default:
		// Unknown package type
		return info, nil
	}

	if err != nil {
		// Network or API error - set status to unknown but don't fail
		info.Status = UpdateStatusUnknown
		return info, nil
	}

	// Compare versions
	if info.CurrentVersion != "" && info.LatestVersion != "" {
		updateAvailable, err := uc.isUpdateAvailable(info.CurrentVersion, info.LatestVersion)
		if err == nil {
			info.UpdateAvailable = updateAvailable
			if updateAvailable {
				info.Status = UpdateStatusAvailable
			} else {
				info.Status = UpdateStatusCurrent
			}
		}
	}

	return info, nil
}

// detectPackageType determines the package type from server metadata
func (uc *UpdateChecker) detectPackageType(server *models.MCPServer) string {
	// Check installation path for clues
	path := strings.ToLower(server.InstallationPath)

	if strings.Contains(path, "node_modules") || strings.Contains(path, "npm") {
		return "npm"
	}
	if strings.Contains(path, "site-packages") || strings.Contains(path, "pip") || strings.Contains(path, "python") {
		return "pypi"
	}
	if strings.Contains(path, "go/pkg") || strings.Contains(path, "GOPATH") {
		return "go"
	}

	// Check runtime dependencies
	for _, dep := range server.Dependencies {
		if dep.Type == models.DependencyRuntime {
			name := strings.ToLower(dep.Name)
			if strings.Contains(name, "node") {
				return "npm"
			}
			if strings.Contains(name, "python") {
				return "pypi"
			}
			if strings.Contains(name, "go") {
				return "go"
			}
		}
	}

	return "unknown"
}

// extractPackageName extracts the package name from server metadata
func (uc *UpdateChecker) extractPackageName(server *models.MCPServer) string {
	// For now, use the server name as package name
	// In a real implementation, this might come from package.json, setup.py, go.mod, etc.
	return server.Name
}

// checkNPMUpdate checks for NPM package updates
func (uc *UpdateChecker) checkNPMUpdate(info *UpdateInfo) error {
	// Try using npm view command
	output, err := uc.executor.Execute("npm", "view", info.PackageName, "version")
	if err != nil {
		return fmt.Errorf("npm view failed: %w", err)
	}

	version := strings.TrimSpace(output)
	if version != "" {
		info.LatestVersion = version
		return nil
	}

	return fmt.Errorf("no version found")
}

// checkPyPIUpdate checks for PyPI package updates
func (uc *UpdateChecker) checkPyPIUpdate(info *UpdateInfo) error {
	url := fmt.Sprintf("https://pypi.org/pypi/%s/json", info.PackageName)

	resp, err := uc.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("pypi request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pypi returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var pypiResp struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
	}

	if err := json.Unmarshal(body, &pypiResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	info.LatestVersion = pypiResp.Info.Version
	return nil
}

// checkGoUpdate checks for Go module updates
func (uc *UpdateChecker) checkGoUpdate(info *UpdateInfo) error {
	// Go modules use the format: example.com/user/module
	// For simplicity, assume the package name is already in this format
	url := fmt.Sprintf("https://proxy.golang.org/%s/@latest", info.PackageName)

	resp, err := uc.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("go proxy request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("go proxy returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var goResp struct {
		Version string `json:"Version"`
	}

	if err := json.Unmarshal(body, &goResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Go proxy returns versions like "v1.2.3", strip the 'v' prefix
	info.LatestVersion = strings.TrimPrefix(goResp.Version, "v")
	return nil
}

// isUpdateAvailable compares two versions using semver
func (uc *UpdateChecker) isUpdateAvailable(current, latest string) (bool, error) {
	// Clean version strings
	current = strings.TrimPrefix(strings.TrimSpace(current), "v")
	latest = strings.TrimPrefix(strings.TrimSpace(latest), "v")

	if current == latest {
		return false, nil
	}

	// Try semver comparison
	currentVer, err := semver.NewVersion(current)
	if err != nil {
		// Not valid semver, do string comparison
		return current != latest, nil
	}

	latestVer, err := semver.NewVersion(latest)
	if err != nil {
		// Latest not valid semver, do string comparison
		return current != latest, nil
	}

	return latestVer.GreaterThan(currentVer), nil
}

// CheckForUpdatesMultiple checks updates for multiple servers
func (uc *UpdateChecker) CheckForUpdatesMultiple(servers []*models.MCPServer) map[string]*UpdateInfo {
	results := make(map[string]*UpdateInfo)

	for _, server := range servers {
		if server == nil {
			continue
		}

		info, err := uc.CheckForUpdates(server)
		if err != nil {
			// Log error but continue with other servers
			results[server.ID] = &UpdateInfo{
				Status:         UpdateStatusUnknown,
				CurrentVersion: server.Version,
				PackageName:    server.Name,
			}
			continue
		}

		results[server.ID] = info
	}

	return results
}

// GetUpdateSummary returns a summary of available updates
func (uc *UpdateChecker) GetUpdateSummary(servers []*models.MCPServer) map[UpdateStatus]int {
	summary := map[UpdateStatus]int{
		UpdateStatusAvailable: 0,
		UpdateStatusCurrent:   0,
		UpdateStatusUnknown:   0,
	}

	for _, server := range servers {
		info, err := uc.CheckForUpdates(server)
		if err != nil || info == nil {
			summary[UpdateStatusUnknown]++
			continue
		}

		summary[info.Status]++
	}

	return summary
}
