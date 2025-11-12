package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/Positronikal/MCPManager/internal/core/discovery"
	"github.com/Positronikal/MCPManager/internal/models"
)

// DiscoveryHandlers contains HTTP handlers for discovery endpoints
type DiscoveryHandlers struct {
	discoveryService *discovery.DiscoveryService
}

// NewDiscoveryHandlers creates a new DiscoveryHandlers instance
func NewDiscoveryHandlers(discoveryService *discovery.DiscoveryService) *DiscoveryHandlers {
	return &DiscoveryHandlers{
		discoveryService: discoveryService,
	}
}

// ListServersResponse is the response structure for GET /servers
type ListServersResponse struct {
	Servers       []models.MCPServer `json:"servers"`
	Count         int                `json:"count"`
	LastDiscovery string             `json:"lastDiscovery"` // ISO 8601 format
}

// DiscoverResponse is the response structure for POST /servers/discover
type DiscoverResponse struct {
	Message string `json:"message"`
	ScanID  string `json:"scanId"`
}

// ListServers handles GET /api/v1/servers
// Returns a list of all discovered servers with optional filters
func (h *DiscoveryHandlers) ListServers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	statusFilter := r.URL.Query().Get("status")
	sourceFilter := r.URL.Query().Get("source")

	// Get all servers from discovery service
	servers, lastDiscovery, err := h.discoveryService.GetServers()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve servers: "+err.Error())
		return
	}

	// Apply filters
	var filteredServers []models.MCPServer
	for _, server := range servers {
		// Filter by status if provided
		if statusFilter != "" {
			if string(server.Status.State) != statusFilter {
				continue
			}
		}

		// Filter by source if provided
		if sourceFilter != "" {
			if string(server.Source) != sourceFilter {
				continue
			}
		}

		filteredServers = append(filteredServers, server)
	}

	// Handle nil slice (return empty array instead of null)
	if filteredServers == nil {
		filteredServers = []models.MCPServer{}
	}

	// Build response
	response := ListServersResponse{
		Servers:       filteredServers,
		Count:         len(filteredServers),
		LastDiscovery: lastDiscovery.Format(time.RFC3339),
	}

	respondJSON(w, http.StatusOK, response)
}

// DiscoverServers handles POST /api/v1/servers/discover
// Triggers a manual discovery scan asynchronously
func (h *DiscoveryHandlers) DiscoverServers(w http.ResponseWriter, r *http.Request) {
	// Generate scan ID
	scanID := uuid.New().String()

	// Trigger discovery in goroutine (non-blocking)
	go func() {
		_, _ = h.discoveryService.Discover()
	}()

	// Return 202 Accepted immediately
	response := DiscoverResponse{
		Message: "Discovery scan initiated",
		ScanID:  scanID,
	}

	respondJSON(w, http.StatusAccepted, response)
}

// GetServerByID handles GET /api/v1/servers/{serverId}
// Returns detailed information about a specific server
func (h *DiscoveryHandlers) GetServerByID(w http.ResponseWriter, r *http.Request) {
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

	respondJSON(w, http.StatusOK, server)
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError writes an error response
func respondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
