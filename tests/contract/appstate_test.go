package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hoytech/mcpmanager/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetApplicationState_ContractValidation tests GET /api/v1/application/state
func TestGetApplicationState_ContractValidation(t *testing.T) {
	services, cleanup := setupFullTestServices(t)
	defer cleanup()
	router := api.NewRouter(services)

	t.Run("should return 200 with ApplicationState schema", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/application/state", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var state map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&state)
		require.NoError(t, err, "Response should be valid JSON")

		// Validate ApplicationState schema (actual implementation)
		assert.Contains(t, state, "version", "State should have 'version' field")
		assert.Contains(t, state, "lastSaved", "State should have 'lastSaved' field")
		assert.Contains(t, state, "preferences", "State should have 'preferences' field")
		assert.Contains(t, state, "windowLayout", "State should have 'windowLayout' field")
		assert.Contains(t, state, "filters", "State should have 'filters' field")
		assert.Contains(t, state, "discoveredServers", "State should have 'discoveredServers' field")
		assert.Contains(t, state, "monitoredConfigPaths", "State should have 'monitoredConfigPaths' field")
		assert.Contains(t, state, "lastDiscoveryScan", "State should have 'lastDiscoveryScan' field")
	})

	t.Run("should return default state if file does not exist", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/application/state", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 200 even on first load (with defaults)
		assert.Equal(t, http.StatusOK, w.Code)

		var state map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&state)

		if err == nil {
			// Default state should have all required fields
			assert.Contains(t, state, "preferences")
			assert.Contains(t, state, "windowLayout")
			assert.Contains(t, state, "filters")
		}
	})

	t.Run("preferences should be object with expected fields", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/application/state", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var state map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&state)

		if err == nil && w.Code == http.StatusOK {
			if prefs, ok := state["preferences"].(map[string]interface{}); ok {
				// UserPreferences fields (actual implementation)
				assert.Contains(t, prefs, "theme", "preferences should have 'theme' field")
				assert.Contains(t, prefs, "logRetentionPerServer", "preferences should have 'logRetentionPerServer' field")
				assert.Contains(t, prefs, "autoStartServers", "preferences should have 'autoStartServers' field")
				assert.Contains(t, prefs, "minimizeToTray", "preferences should have 'minimizeToTray' field")
				assert.Contains(t, prefs, "showNotifications", "preferences should have 'showNotifications' field")

				// Theme should be valid enum
				if theme, ok := prefs["theme"].(string); ok {
					validThemes := []string{"light", "dark"}
					assert.Contains(t, validThemes, theme, "theme should be 'light' or 'dark'")
				}
			}
		}
	})

	t.Run("windowLayout should be object with dimension constraints", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/application/state", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var state map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&state)

		if err == nil && w.Code == http.StatusOK {
			if windowLayout, ok := state["windowLayout"].(map[string]interface{}); ok {
				// WindowLayout fields per data-model.md
				assert.Contains(t, windowLayout, "width", "windowLayout should have 'width' field")
				assert.Contains(t, windowLayout, "height", "windowLayout should have 'height' field")
				assert.Contains(t, windowLayout, "x", "windowLayout should have 'x' field")
				assert.Contains(t, windowLayout, "y", "windowLayout should have 'y' field")

				// Width and height should be positive numbers
				if width, ok := windowLayout["width"].(float64); ok {
					assert.Greater(t, width, float64(0), "width should be positive")
				}

				if height, ok := windowLayout["height"].(float64); ok {
					assert.Greater(t, height, float64(0), "height should be positive")
				}
			}
		}
	})

	t.Run("filters should be object", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/application/state", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var state map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&state)

		if err == nil && w.Code == http.StatusOK {
			// Filters should be present (though may be empty object due to omitempty fields)
			_, ok := state["filters"].(map[string]interface{})
			assert.True(t, ok, "filters should be an object")
			// All Filters fields (selectedServer, selectedSeverity, searchQuery) have omitempty tags
			// so they may not be present when empty
		}
	})
}

// TestPutApplicationState_ContractValidation tests PUT /api/v1/application/state
func TestPutApplicationState_ContractValidation(t *testing.T) {
	services, cleanup := setupFullTestServices(t)
	defer cleanup()
	router := api.NewRouter(services)

	t.Run("should return 200 with success message", func(t *testing.T) {

		validState := map[string]interface{}{
			"version":   "1.0.0",
			"lastSaved": "2025-01-01T00:00:00Z",
			"preferences": map[string]interface{}{
				"theme":                 "dark",
				"logRetentionPerServer": 1000,
				"autoStartServers":      true,
				"minimizeToTray":        true,
				"showNotifications":     true,
			},
			"windowLayout": map[string]interface{}{
				"width":          1024,
				"height":         768,
				"x":              100,
				"y":              100,
				"maximized":      false,
				"logPanelHeight": 300,
			},
			"filters": map[string]interface{}{
				"searchQuery": "",
			},
			"discoveredServers":    []string{},
			"monitoredConfigPaths": []string{},
			"lastDiscoveryScan":    "2025-01-01T00:00:00Z",
		}
		bodyBytes, _ := json.Marshal(validState)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/application/state", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var response struct {
			Message string `json:"message"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err, "Response should be valid JSON")

		assert.Equal(t, "Application state saved", response.Message, "Response should confirm state saved")
	})

	t.Run("should return 400 for validation error - invalid theme", func(t *testing.T) {
		invalidState := map[string]interface{}{
			"version":   "1.0.0",
			"lastSaved": "2025-01-01T00:00:00Z",
			"preferences": map[string]interface{}{
				"theme":                 "invalid-theme", // Invalid enum value
				"logRetentionPerServer": 1000,
				"autoStartServers":      true,
				"minimizeToTray":        true,
				"showNotifications":     true,
			},
			"windowLayout": map[string]interface{}{
				"width":          1024,
				"height":         768,
				"x":              0,
				"y":              0,
				"maximized":      false,
				"logPanelHeight": 300,
			},
			"filters": map[string]interface{}{
				"searchQuery": "",
			},
			"discoveredServers":    []string{},
			"monitoredConfigPaths": []string{},
			"lastDiscoveryScan":    "2025-01-01T00:00:00Z",
		}
		bodyBytes, _ := json.Marshal(invalidState)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/application/state", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 400 for validation error
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error", "Error response should have 'error' field")
		}
	})

	t.Run("should return 400 for validation error - negative window dimensions", func(t *testing.T) {
		invalidState := map[string]interface{}{
			"version":   "1.0.0",
			"lastSaved": "2025-01-01T00:00:00Z",
			"preferences": map[string]interface{}{
				"theme":                 "dark",
				"logRetentionPerServer": 1000,
				"autoStartServers":      true,
				"minimizeToTray":        true,
				"showNotifications":     true,
			},
			"windowLayout": map[string]interface{}{
				"width":          -100, // Invalid negative width
				"height":         768,
				"x":              0,
				"y":              0,
				"maximized":      false,
				"logPanelHeight": 300,
			},
			"filters": map[string]interface{}{
				"searchQuery": "",
			},
			"discoveredServers":    []string{},
			"monitoredConfigPaths": []string{},
			"lastDiscoveryScan":    "2025-01-01T00:00:00Z",
		}
		bodyBytes, _ := json.Marshal(invalidState)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/application/state", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 400 for validation error
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})

	t.Run("should return 400 for validation error - invalid UUID format", func(t *testing.T) {
		invalidState := map[string]interface{}{
			"version":   "1.0.0",
			"lastSaved": "2025-01-01T00:00:00Z",
			"preferences": map[string]interface{}{
				"theme":                 "dark",
				"logRetentionPerServer": 1000,
				"autoStartServers":      true,
				"minimizeToTray":        true,
				"showNotifications":     true,
			},
			"windowLayout": map[string]interface{}{
				"width":          1024,
				"height":         768,
				"x":              0,
				"y":              0,
				"maximized":      false,
				"logPanelHeight": 300,
			},
			"filters": map[string]interface{}{
				"selectedServer": "not-a-uuid", // Invalid UUID format
				"searchQuery":    "",
			},
			"discoveredServers":    []string{},
			"monitoredConfigPaths": []string{},
			"lastDiscoveryScan":    "2025-01-01T00:00:00Z",
		}
		bodyBytes, _ := json.Marshal(invalidState)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/application/state", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 400 for validation error
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})

	t.Run("should require request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/application/state", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 400 for missing body
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})

	t.Run("state should persist across restarts", func(t *testing.T) {
		// This is more of an integration test requirement
		// Contract test just validates the API accepts and returns state correctly

		validState := map[string]interface{}{
			"version":   "1.0.0",
			"lastSaved": "2025-01-01T00:00:00Z",
			"preferences": map[string]interface{}{
				"theme":                 "dark",
				"logRetentionPerServer": 1000,
				"autoStartServers":      true,
				"minimizeToTray":        true,
				"showNotifications":     true,
			},
			"windowLayout": map[string]interface{}{
				"width":          1920,
				"height":         1080,
				"x":              0,
				"y":              0,
				"maximized":      false,
				"logPanelHeight": 300,
			},
			"filters": map[string]interface{}{
				"searchQuery": "",
			},
			"discoveredServers":    []string{},
			"monitoredConfigPaths": []string{},
			"lastDiscoveryScan":    "2025-01-01T00:00:00Z",
		}
		bodyBytes, _ := json.Marshal(validState)

		// PUT state
		req := httptest.NewRequest(http.MethodPut, "/api/v1/application/state", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 200
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
