package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/Positronikal/MCPManager/internal/core/discovery"
	"github.com/Positronikal/MCPManager/internal/core/lifecycle"
	"github.com/Positronikal/MCPManager/internal/models"
)

// LifecycleHandlers contains HTTP handlers for lifecycle endpoints
type LifecycleHandlers struct {
	lifecycleService *lifecycle.LifecycleService
	discoveryService *discovery.DiscoveryService
}

// NewLifecycleHandlers creates a new LifecycleHandlers instance
func NewLifecycleHandlers(lifecycleService *lifecycle.LifecycleService, discoveryService *discovery.DiscoveryService) *LifecycleHandlers {
	return &LifecycleHandlers{
		lifecycleService: lifecycleService,
		discoveryService: discoveryService,
	}
}

// StartServerResponse is the response structure for POST /servers/{serverId}/start
type StartServerResponse struct {
	Message  string `json:"message"`
	ServerID string `json:"serverId"`
	Status   string `json:"status"`
}

// StopServerRequest is the request structure for POST /servers/{serverId}/stop
type StopServerRequest struct {
	Force   bool `json:"force"`
	Timeout int  `json:"timeout"`
}

// StopServerResponse is the response structure for POST /servers/{serverId}/stop
type StopServerResponse struct {
	Message  string `json:"message"`
	ServerID string `json:"serverId"`
}

// RestartServerResponse is the response structure for POST /servers/{serverId}/restart
type RestartServerResponse struct {
	Message  string `json:"message"`
	ServerID string `json:"serverId"`
}

// StartServer handles POST /api/v1/servers/{serverId}/start
func (h *LifecycleHandlers) StartServer(w http.ResponseWriter, r *http.Request) {
	// Extract server ID from URL
	serverID := chi.URLParam(r, "serverId")

	// Validate UUID format
	if _, err := uuid.Parse(serverID); err != nil {
		respondError(w, http.StatusNotFound, "Invalid server ID format")
		return
	}

	// Get server from discovery service
	server, exists := h.discoveryService.GetServerByID(serverID)
	if !exists {
		respondError(w, http.StatusNotFound, "Server not found")
		return
	}

	// Check if server is already running
	if server.Status.State == models.StatusRunning || server.Status.State == models.StatusStarting {
		respondError(w, http.StatusBadRequest, "Server is already running or starting")
		return
	}

	// Start server asynchronously
	go func() {
		err := h.lifecycleService.StartServer(server)
		if err == nil {
			// Update server in cache
			h.discoveryService.UpdateServer(server)
		}
	}()

	// Return 202 Accepted immediately
	response := StartServerResponse{
		Message:  "Server starting",
		ServerID: serverID,
		Status:   "starting",
	}

	respondJSON(w, http.StatusAccepted, response)
}

// StopServer handles POST /api/v1/servers/{serverId}/stop
func (h *LifecycleHandlers) StopServer(w http.ResponseWriter, r *http.Request) {
	// Extract server ID from URL
	serverID := chi.URLParam(r, "serverId")

	// Validate UUID format
	if _, err := uuid.Parse(serverID); err != nil {
		respondError(w, http.StatusNotFound, "Invalid server ID format")
		return
	}

	// Get server from discovery service
	server, exists := h.discoveryService.GetServerByID(serverID)
	if !exists {
		respondError(w, http.StatusNotFound, "Server not found")
		return
	}

	// Check if server is running
	if server.Status.State != models.StatusRunning && server.Status.State != models.StatusStarting {
		respondError(w, http.StatusBadRequest, "Server is not running")
		return
	}

	// Parse request body (optional parameters)
	var req StopServerRequest
	req.Force = false   // Default
	req.Timeout = 10    // Default

	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// Ignore decode errors - use defaults
		}
	}

	// Stop server asynchronously
	go func() {
		err := h.lifecycleService.StopServer(server, req.Force, req.Timeout)
		if err == nil {
			// Update server in cache
			h.discoveryService.UpdateServer(server)
		}
	}()

	// Return 202 Accepted immediately
	response := StopServerResponse{
		Message:  "Server stopping",
		ServerID: serverID,
	}

	respondJSON(w, http.StatusAccepted, response)
}

// RestartServer handles POST /api/v1/servers/{serverId}/restart
func (h *LifecycleHandlers) RestartServer(w http.ResponseWriter, r *http.Request) {
	// Extract server ID from URL
	serverID := chi.URLParam(r, "serverId")

	// Validate UUID format
	if _, err := uuid.Parse(serverID); err != nil {
		respondError(w, http.StatusNotFound, "Invalid server ID format")
		return
	}

	// Get server from discovery service
	server, exists := h.discoveryService.GetServerByID(serverID)
	if !exists {
		respondError(w, http.StatusNotFound, "Server not found")
		return
	}

	// Restart server asynchronously
	go func() {
		err := h.lifecycleService.RestartServer(server)
		if err == nil {
			// Update server in cache
			h.discoveryService.UpdateServer(server)
		}
	}()

	// Return 202 Accepted immediately
	response := RestartServerResponse{
		Message:  "Server restarting",
		ServerID: serverID,
	}

	respondJSON(w, http.StatusAccepted, response)
}

// GetServerStatus handles GET /api/v1/servers/{serverId}/status
func (h *LifecycleHandlers) GetServerStatus(w http.ResponseWriter, r *http.Request) {
	// Extract server ID from URL
	serverID := chi.URLParam(r, "serverId")

	// Validate UUID format
	if _, err := uuid.Parse(serverID); err != nil {
		respondError(w, http.StatusNotFound, "Invalid server ID format")
		return
	}

	// Get server from discovery service
	server, exists := h.discoveryService.GetServerByID(serverID)
	if !exists {
		respondError(w, http.StatusNotFound, "Server not found")
		return
	}

	// Return current server status
	respondJSON(w, http.StatusOK, server.Status)
}
