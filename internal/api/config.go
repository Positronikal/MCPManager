package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/Positronikal/MCPManager/internal/core/config"
	"github.com/Positronikal/MCPManager/internal/core/discovery"
	"github.com/Positronikal/MCPManager/internal/models"
)

// ConfigHandlers contains HTTP handlers for configuration endpoints
type ConfigHandlers struct {
	configService    *config.ConfigService
	discoveryService *discovery.DiscoveryService
}

// NewConfigHandlers creates a new ConfigHandlers instance
func NewConfigHandlers(configService *config.ConfigService, discoveryService *discovery.DiscoveryService) *ConfigHandlers {
	return &ConfigHandlers{
		configService:    configService,
		discoveryService: discoveryService,
	}
}

// GetConfiguration handles GET /api/v1/servers/{serverId}/configuration
func (h *ConfigHandlers) GetConfiguration(w http.ResponseWriter, r *http.Request) {
	// Extract server ID from URL
	serverID := chi.URLParam(r, "serverId")

	// Validate UUID format
	if _, err := uuid.Parse(serverID); err != nil {
		respondError(w, http.StatusNotFound, "Invalid server ID format")
		return
	}

	// Check if server exists
	_, exists := h.discoveryService.GetServerByID(serverID)
	if !exists {
		respondError(w, http.StatusNotFound, "Server not found")
		return
	}

	// Get configuration
	config, err := h.configService.GetConfiguration(serverID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to retrieve configuration: "+err.Error())
		return
	}

	respondJSON(w, http.StatusOK, config)
}

// UpdateConfiguration handles PUT /api/v1/servers/{serverId}/configuration
func (h *ConfigHandlers) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	// Extract server ID from URL
	serverID := chi.URLParam(r, "serverId")

	// Validate UUID format
	if _, err := uuid.Parse(serverID); err != nil {
		respondError(w, http.StatusNotFound, "Invalid server ID format")
		return
	}

	// Check if server exists
	server, exists := h.discoveryService.GetServerByID(serverID)
	if !exists {
		respondError(w, http.StatusNotFound, "Server not found")
		return
	}

	// Parse request body
	if r.Body == nil {
		respondError(w, http.StatusBadRequest, "Request body is required")
		return
	}

	var config models.ServerConfiguration
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate configuration
	if err := h.configService.ValidateConfiguration(&config); err != nil {
		respondError(w, http.StatusBadRequest, "Validation error: "+err.Error())
		return
	}

	// Update configuration
	if err := h.configService.UpdateConfiguration(serverID, &config); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update configuration: "+err.Error())
		return
	}

	// Update server's configuration in cache
	server.Configuration = config
	h.discoveryService.UpdateServer(server)

	// Return updated configuration
	respondJSON(w, http.StatusOK, &config)
}
