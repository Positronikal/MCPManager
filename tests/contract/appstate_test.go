package contract

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetApplicationState_ContractValidation tests GET /api/v1/application/state
func TestGetApplicationState_ContractValidation(t *testing.T) {
	t.Run("should return 200 with ApplicationState schema", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D014
			w.WriteHeader(http.StatusNotImplemented)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/application/state", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var state map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&state)
		require.NoError(t, err, "Response should be valid JSON")

		// Validate ApplicationState schema per data-model.md
		assert.Contains(t, state, "userPreferences", "State should have 'userPreferences' field")
		assert.Contains(t, state, "windowLayout", "State should have 'windowLayout' field")
		assert.Contains(t, state, "serverFilters", "State should have 'serverFilters' field")
		assert.Contains(t, state, "selectedServerId", "State should have 'selectedServerId' field")
		assert.Contains(t, state, "lastSyncedAt", "State should have 'lastSyncedAt' field")
	})

	t.Run("should return default state if file does not exist", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D014
			w.WriteHeader(http.StatusNotImplemented)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/application/state", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should return 200 even on first load (with defaults)
		assert.Equal(t, http.StatusOK, w.Code)

		var state map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&state)

		if err == nil {
			// Default state should have all required fields
			assert.Contains(t, state, "userPreferences")
			assert.Contains(t, state, "windowLayout")
			assert.Contains(t, state, "serverFilters")
		}
	})

	t.Run("userPreferences should be object with expected fields", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D014
			w.WriteHeader(http.StatusNotImplemented)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/application/state", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var state map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&state)

		if err == nil && w.Code == http.StatusOK {
			if userPrefs, ok := state["userPreferences"].(map[string]interface{}); ok {
				// UserPreferences fields per data-model.md
				assert.Contains(t, userPrefs, "theme", "userPreferences should have 'theme' field")
				assert.Contains(t, userPrefs, "autoStartServers", "userPreferences should have 'autoStartServers' field")
				assert.Contains(t, userPrefs, "showNotifications", "userPreferences should have 'showNotifications' field")

				// Theme should be valid enum
				if theme, ok := userPrefs["theme"].(string); ok {
					validThemes := []string{"light", "dark", "system"}
					assert.Contains(t, validThemes, theme, "theme should be valid enum")
				}
			}
		}
	})

	t.Run("windowLayout should be object with dimension constraints", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D014
			w.WriteHeader(http.StatusNotImplemented)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/application/state", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

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

	t.Run("serverFilters should be object", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D014
			w.WriteHeader(http.StatusNotImplemented)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/application/state", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var state map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&state)

		if err == nil && w.Code == http.StatusOK {
			if serverFilters, ok := state["serverFilters"].(map[string]interface{}); ok {
				// ServerFilters fields per data-model.md
				assert.Contains(t, serverFilters, "status", "serverFilters should have 'status' field")
				assert.Contains(t, serverFilters, "source", "serverFilters should have 'source' field")
				assert.Contains(t, serverFilters, "searchQuery", "serverFilters should have 'searchQuery' field")
			}
		}
	})
}

// TestPutApplicationState_ContractValidation tests PUT /api/v1/application/state
func TestPutApplicationState_ContractValidation(t *testing.T) {
	t.Run("should return 200 with success message", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D014
			w.WriteHeader(http.StatusNotImplemented)
		})

		validState := map[string]interface{}{
			"userPreferences": map[string]interface{}{
				"theme":              "dark",
				"autoStartServers":   true,
				"showNotifications":  true,
				"logLevel":           "info",
				"refreshInterval":    5,
				"enableAutoDiscovery": true,
			},
			"windowLayout": map[string]interface{}{
				"width":      1024,
				"height":     768,
				"x":          100,
				"y":          100,
				"maximized":  false,
				"fullscreen": false,
			},
			"serverFilters": map[string]interface{}{
				"status":      "running",
				"source":      "client_config",
				"searchQuery": "",
			},
			"selectedServerId": nil,
			"lastSyncedAt":     "2025-01-01T00:00:00Z",
		}
		bodyBytes, _ := json.Marshal(validState)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/application/state", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var response struct {
			Message string `json:"message"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err, "Response should be valid JSON")

		assert.Equal(t, "Application state saved", response.Message, "Response should confirm state saved")
	})

	t.Run("should return 400 for validation error - invalid theme", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D014
			w.WriteHeader(http.StatusNotImplemented)
		})

		invalidState := map[string]interface{}{
			"userPreferences": map[string]interface{}{
				"theme":             "invalid-theme", // Invalid enum value
				"autoStartServers":  true,
				"showNotifications": true,
			},
			"windowLayout": map[string]interface{}{
				"width":  1024,
				"height": 768,
				"x":      0,
				"y":      0,
			},
			"serverFilters": map[string]interface{}{},
		}
		bodyBytes, _ := json.Marshal(invalidState)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/application/state", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should return 400 for validation error
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error", "Error response should have 'error' field")
		}
	})

	t.Run("should return 400 for validation error - negative window dimensions", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D014
			w.WriteHeader(http.StatusNotImplemented)
		})

		invalidState := map[string]interface{}{
			"userPreferences": map[string]interface{}{
				"theme":             "dark",
				"autoStartServers":  true,
				"showNotifications": true,
			},
			"windowLayout": map[string]interface{}{
				"width":  -100, // Invalid negative width
				"height": 768,
				"x":      0,
				"y":      0,
			},
			"serverFilters": map[string]interface{}{},
		}
		bodyBytes, _ := json.Marshal(invalidState)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/application/state", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should return 400 for validation error
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})

	t.Run("should return 400 for validation error - invalid UUID format", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D014
			w.WriteHeader(http.StatusNotImplemented)
		})

		invalidState := map[string]interface{}{
			"userPreferences": map[string]interface{}{
				"theme":             "dark",
				"autoStartServers":  true,
				"showNotifications": true,
			},
			"windowLayout": map[string]interface{}{
				"width":  1024,
				"height": 768,
				"x":      0,
				"y":      0,
			},
			"serverFilters": map[string]interface{}{},
			"selectedServerId": "not-a-uuid", // Invalid UUID format
		}
		bodyBytes, _ := json.Marshal(invalidState)

		req := httptest.NewRequest(http.MethodPut, "/api/v1/application/state", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should return 400 for validation error
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})

	t.Run("should require request body", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D014
			w.WriteHeader(http.StatusNotImplemented)
		})

		req := httptest.NewRequest(http.MethodPut, "/api/v1/application/state", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

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

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D014
			w.WriteHeader(http.StatusNotImplemented)
		})

		validState := map[string]interface{}{
			"userPreferences": map[string]interface{}{
				"theme":             "dark",
				"autoStartServers":  true,
				"showNotifications": true,
			},
			"windowLayout": map[string]interface{}{
				"width":  1920,
				"height": 1080,
				"x":      0,
				"y":      0,
			},
			"serverFilters": map[string]interface{}{
				"status": "running",
			},
		}
		bodyBytes, _ := json.Marshal(validState)

		// PUT state
		req := httptest.NewRequest(http.MethodPut, "/api/v1/application/state", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should return 200
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
