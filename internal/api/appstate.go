package api

import (
	"encoding/json"
	"net/http"

	"github.com/hoytech/mcpmanager/internal/models"
	"github.com/hoytech/mcpmanager/internal/storage"
)

// AppStateHandlers contains HTTP handlers for application state endpoints
type AppStateHandlers struct {
	storageService storage.StorageService
}

// NewAppStateHandlers creates a new AppStateHandlers instance
func NewAppStateHandlers(storageService storage.StorageService) *AppStateHandlers {
	return &AppStateHandlers{
		storageService: storageService,
	}
}

// SaveStateResponse is the response structure for PUT /application/state
type SaveStateResponse struct {
	Message string `json:"message"`
}

// GetApplicationState handles GET /api/v1/application/state
func (h *AppStateHandlers) GetApplicationState(w http.ResponseWriter, r *http.Request) {
	// Load state from storage
	state, err := h.storageService.LoadState()
	if err != nil {
		// If error loading, return defaults
		state = models.NewApplicationState()
	}

	respondJSON(w, http.StatusOK, state)
}

// UpdateApplicationState handles PUT /api/v1/application/state
func (h *AppStateHandlers) UpdateApplicationState(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	if r.Body == nil {
		respondError(w, http.StatusBadRequest, "Request body is required")
		return
	}

	var state models.ApplicationState
	if err := json.NewDecoder(r.Body).Decode(&state); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Validate state
	if err := state.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, "Validation error: "+err.Error())
		return
	}

	// Save state to storage
	if err := h.storageService.SaveState(&state); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to save state: "+err.Error())
		return
	}

	// Return success message
	response := SaveStateResponse{
		Message: "Application state saved",
	}

	respondJSON(w, http.StatusOK, response)
}
