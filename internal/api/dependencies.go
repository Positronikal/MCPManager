package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/hoytech/mcpmanager/internal/core/dependencies"
	"github.com/hoytech/mcpmanager/internal/core/discovery"
	"github.com/hoytech/mcpmanager/internal/models"
)

// DependencyHandlers contains HTTP handlers for dependency endpoints
type DependencyHandlers struct {
	dependencyService *dependencies.DependencyService
	updateChecker     *dependencies.UpdateChecker
	discoveryService  *discovery.DiscoveryService
}

// NewDependencyHandlers creates a new DependencyHandlers instance
func NewDependencyHandlers(dependencyService *dependencies.DependencyService, updateChecker *dependencies.UpdateChecker, discoveryService *discovery.DiscoveryService) *DependencyHandlers {
	return &DependencyHandlers{
		dependencyService: dependencyService,
		updateChecker:     updateChecker,
		discoveryService:  discoveryService,
	}
}

// DependenciesResponse is the response structure for GET /servers/{serverId}/dependencies
type DependenciesResponse struct {
	Dependencies []models.Dependency `json:"dependencies"`
	AllSatisfied bool                `json:"allSatisfied"`
}

// UpdatesResponse is the response structure for GET /servers/{serverId}/updates
type UpdatesResponse struct {
	UpdateAvailable bool   `json:"updateAvailable"`
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	ReleaseNotes    string `json:"releaseNotes,omitempty"`
}

// GetServerDependencies handles GET /api/v1/servers/{serverId}/dependencies
func (h *DependencyHandlers) GetServerDependencies(w http.ResponseWriter, r *http.Request) {
	// Extract server ID from URL
	serverID := chi.URLParam(r, "serverId")

	// Validate UUID format
	if _, err := uuid.Parse(serverID); err != nil {
		respondError(w, http.StatusNotFound, "Invalid server ID format")
		return
	}

	// Get server
	server, exists := h.discoveryService.GetServerByID(serverID)
	if !exists {
		respondError(w, http.StatusNotFound, "Server not found")
		return
	}

	// Check dependencies
	deps, err := h.dependencyService.CheckDependencies(server)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to check dependencies: "+err.Error())
		return
	}

	// Handle nil slice
	if deps == nil {
		deps = []models.Dependency{}
	}

	// Check if all dependencies are satisfied
	allSatisfied := true
	for _, dep := range deps {
		if !dep.IsInstalled() {
			allSatisfied = false
			break
		}
	}

	response := DependenciesResponse{
		Dependencies: deps,
		AllSatisfied: allSatisfied,
	}

	respondJSON(w, http.StatusOK, response)
}

// GetServerUpdates handles GET /api/v1/servers/{serverId}/updates
func (h *DependencyHandlers) GetServerUpdates(w http.ResponseWriter, r *http.Request) {
	// Extract server ID from URL
	serverID := chi.URLParam(r, "serverId")

	// Validate UUID format
	if _, err := uuid.Parse(serverID); err != nil {
		respondError(w, http.StatusNotFound, "Invalid server ID format")
		return
	}

	// Get server
	server, exists := h.discoveryService.GetServerByID(serverID)
	if !exists {
		respondError(w, http.StatusNotFound, "Server not found")
		return
	}

	// Check for updates
	updateInfo, err := h.updateChecker.CheckForUpdates(server)

	// Handle network errors gracefully - return available info with unknown status
	if err != nil {
		// Still return a response with current version
		response := UpdatesResponse{
			UpdateAvailable: false,
			CurrentVersion:  server.Version,
			LatestVersion:   server.Version,
			ReleaseNotes:    "",
		}
		respondJSON(w, http.StatusOK, response)
		return
	}

	// Build response
	response := UpdatesResponse{
		UpdateAvailable: updateInfo.UpdateAvailable,
		CurrentVersion:  updateInfo.CurrentVersion,
		LatestVersion:   updateInfo.LatestVersion,
		ReleaseNotes:    updateInfo.ReleaseNotes,
	}

	// Ensure versions are not empty
	if response.CurrentVersion == "" {
		response.CurrentVersion = server.Version
		if response.CurrentVersion == "" {
			response.CurrentVersion = "unknown"
		}
	}

	if response.LatestVersion == "" {
		response.LatestVersion = response.CurrentVersion
	}

	respondJSON(w, http.StatusOK, response)
}
