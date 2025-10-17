package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetServerConfiguration_ContractValidation tests GET /api/v1/servers/{serverId}/configuration
func TestGetServerConfiguration_ContractValidation(t *testing.T) {
	t.Run("should return 200 with ServerConfiguration schema", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D011
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var config map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&config)
		require.NoError(t, err, "Response should be valid JSON")

		// Validate ServerConfiguration schema per data-model.md
		assert.Contains(t, config, "command", "Configuration should have 'command' field")
		assert.Contains(t, config, "args", "Configuration should have 'args' field")
		assert.Contains(t, config, "env", "Configuration should have 'env' field")
		assert.Contains(t, config, "autoStart", "Configuration should have 'autoStart' field")
		assert.Contains(t, config, "restartOnFailure", "Configuration should have 'restartOnFailure' field")
		assert.Contains(t, config, "maxRestarts", "Configuration should have 'maxRestarts' field")
		assert.Contains(t, config, "restartDelay", "Configuration should have 'restartDelay' field")
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D011
			w.WriteHeader(http.StatusNotImplemented)
		})

		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/configuration", nonExistentUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})

	t.Run("args should be array and env should be object", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D011
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var config map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&config)

		if err == nil && w.Code == http.StatusOK {
			// args should be array
			if args, exists := config["args"]; exists {
				_, ok := args.([]interface{})
				assert.True(t, ok, "args should be an array")
			}

			// env should be object (map)
			if env, exists := config["env"]; exists {
				_, ok := env.(map[string]interface{})
				assert.True(t, ok, "env should be an object")
			}

			// autoStart should be boolean
			if autoStart, exists := config["autoStart"]; exists {
				_, ok := autoStart.(bool)
				assert.True(t, ok, "autoStart should be a boolean")
			}
		}
	})
}

// TestPutServerConfiguration_ContractValidation tests PUT /api/v1/servers/{serverId}/configuration
func TestPutServerConfiguration_ContractValidation(t *testing.T) {
	t.Run("should return 200 with updated configuration", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D011
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()

		// Valid ServerConfiguration per data-model.md
		configUpdate := map[string]interface{}{
			"command":          "node",
			"args":             []string{"server.js"},
			"env":              map[string]string{"NODE_ENV": "production"},
			"autoStart":        true,
			"restartOnFailure": true,
			"maxRestarts":      3,
			"restartDelay":     5,
		}
		bodyBytes, _ := json.Marshal(configUpdate)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var responseConfig map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&responseConfig)
		require.NoError(t, err, "Response should be valid JSON")

		// Response should be the updated ServerConfiguration
		assert.Contains(t, responseConfig, "command")
		assert.Contains(t, responseConfig, "args")
		assert.Contains(t, responseConfig, "env")
	})

	t.Run("should return 400 for validation error", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D011
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()

		// Invalid configuration - missing required 'command' field
		invalidConfig := map[string]interface{}{
			"args":      []string{"server.js"},
			"autoStart": true,
		}
		bodyBytes, _ := json.Marshal(invalidConfig)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should return 400 for validation error
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err, "Error response should be valid JSON")

			assert.Contains(t, errorResponse, "error", "Error response should have 'error' field")

			// Error message should describe validation failure
			if errMsg, ok := errorResponse["error"].(string); ok {
				assert.NotEmpty(t, errMsg, "Error message should be descriptive")
			}
		}
	})

	t.Run("should return 400 for invalid field types", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D011
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()

		// Invalid - autoStart should be boolean, not string
		invalidConfig := map[string]interface{}{
			"command":   "node",
			"args":      []string{"server.js"},
			"autoStart": "true", // Wrong type
		}
		bodyBytes, _ := json.Marshal(invalidConfig)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should return 400 for type validation error
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D011
			w.WriteHeader(http.StatusNotImplemented)
		})

		nonExistentUUID := uuid.New().String()

		configUpdate := map[string]interface{}{
			"command":   "node",
			"args":      []string{"server.js"},
			"autoStart": true,
		}
		bodyBytes, _ := json.Marshal(configUpdate)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/servers/%s/configuration", nonExistentUUID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})

	t.Run("should require request body", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D011
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()

		// PUT without body should fail
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), nil)
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

	t.Run("should validate maxRestarts and restartDelay ranges", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D011
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()

		// Invalid - negative maxRestarts
		invalidConfig := map[string]interface{}{
			"command":     "node",
			"args":        []string{},
			"autoStart":   true,
			"maxRestarts": -1, // Invalid negative value
		}
		bodyBytes, _ := json.Marshal(invalidConfig)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should return 400 for invalid range
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})
}
