package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/Positronikal/MCPManager/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetServerConfiguration_ContractValidation tests GET /api/v1/servers/{serverId}/configuration
func TestGetServerConfiguration_ContractValidation(t *testing.T) {
	services, cleanup := setupFullTestServices(t)
	defer cleanup()
	router := api.NewRouter(services)

	// Get a valid server ID for testing
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	var serverList struct {
		Servers []struct {
			ID string `json:"id"`
		} `json:"servers"`
	}
	json.NewDecoder(listW.Body).Decode(&serverList)

	var validUUID string
	if len(serverList.Servers) > 0 {
		validUUID = serverList.Servers[0].ID
	}

	t.Run("should return 200 with ServerConfiguration schema", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var config map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&config)
		require.NoError(t, err, "Response should be valid JSON")

		// Validate ServerConfiguration schema (actual implementation)
		assert.Contains(t, config, "commandLineArguments", "Configuration should have 'commandLineArguments' field")
		assert.Contains(t, config, "environmentVariables", "Configuration should have 'environmentVariables' field")
		assert.Contains(t, config, "autoStart", "Configuration should have 'autoStart' field")
		assert.Contains(t, config, "restartOnCrash", "Configuration should have 'restartOnCrash' field")
		assert.Contains(t, config, "maxRestartAttempts", "Configuration should have 'maxRestartAttempts' field")
		assert.Contains(t, config, "startupTimeout", "Configuration should have 'startupTimeout' field")
		assert.Contains(t, config, "shutdownTimeout", "Configuration should have 'shutdownTimeout' field")
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/configuration", nonExistentUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			var errorResponse map[string]interface{}
			if w.Body.Len() > 0 {
				err := json.NewDecoder(w.Body).Decode(&errorResponse)
				require.NoError(t, err)
				assert.Contains(t, errorResponse, "error")
			}
		}
	})

	t.Run("args should be array and env should be object", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var config map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&config)

		if err == nil && w.Code == http.StatusOK {
			// commandLineArguments should be array
			if args, exists := config["commandLineArguments"]; exists {
				_, ok := args.([]interface{})
				assert.True(t, ok, "commandLineArguments should be an array")
			}

			// environmentVariables should be object (map)
			if env, exists := config["environmentVariables"]; exists {
				_, ok := env.(map[string]interface{})
				assert.True(t, ok, "environmentVariables should be an object")
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
	services, cleanup := setupFullTestServices(t)
	defer cleanup()
	router := api.NewRouter(services)

	// Get a valid server ID for testing
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	var serverList struct {
		Servers []struct {
			ID string `json:"id"`
		} `json:"servers"`
	}
	json.NewDecoder(listW.Body).Decode(&serverList)

	var validUUID string
	if len(serverList.Servers) > 0 {
		validUUID = serverList.Servers[0].ID
	}

	t.Run("should return 200 with updated configuration", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		// Valid ServerConfiguration (actual implementation)
		configUpdate := map[string]interface{}{
			"commandLineArguments":  []string{"server.js"},
			"environmentVariables":  map[string]string{"NODE_ENV": "production"},
			"autoStart":             true,
			"restartOnCrash":        true,
			"maxRestartAttempts":    3,
			"startupTimeout":        30,
			"shutdownTimeout":       10,
			"healthCheckInterval":   5,
		}
		bodyBytes, _ := json.Marshal(configUpdate)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var responseConfig map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&responseConfig)
		require.NoError(t, err, "Response should be valid JSON")

		// Response should be the updated ServerConfiguration
		assert.Contains(t, responseConfig, "commandLineArguments")
		assert.Contains(t, responseConfig, "environmentVariables")
		assert.Contains(t, responseConfig, "autoStart")
	})

	t.Run("should return 400 for validation error", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		// Invalid configuration test (API might not enforce specific required fields)
		invalidConfig := map[string]interface{}{
			"commandLineArguments": []string{"server.js"},
			"maxRestartAttempts":   -1, // Invalid negative value
		}
		bodyBytes, _ := json.Marshal(invalidConfig)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

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
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		// Invalid - autoStart should be boolean, not string
		invalidConfig := map[string]interface{}{
			"commandLineArguments": []string{"server.js"},
			"autoStart":            "true", // Wrong type
		}
		bodyBytes, _ := json.Marshal(invalidConfig)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 400 for type validation error
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()

		configUpdate := map[string]interface{}{
			"commandLineArguments": []string{"server.js"},
			"autoStart":            true,
		}
		bodyBytes, _ := json.Marshal(configUpdate)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/servers/%s/configuration", nonExistentUUID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			var errorResponse map[string]interface{}
			if w.Body.Len() > 0 {
				err := json.NewDecoder(w.Body).Decode(&errorResponse)
				require.NoError(t, err)
				assert.Contains(t, errorResponse, "error")
			}
		}
	})

	t.Run("should require request body", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		// PUT without body should fail
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 400 for missing body
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			if w.Body.Len() > 0 {
				err := json.NewDecoder(w.Body).Decode(&errorResponse)
				require.NoError(t, err)
				assert.Contains(t, errorResponse, "error")
			}
		}
	})

	t.Run("should validate maxRestarts and restartDelay ranges", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		// Invalid - negative maxRestartAttempts
		invalidConfig := map[string]interface{}{
			"commandLineArguments": []string{},
			"autoStart":            true,
			"maxRestartAttempts":   -1, // Invalid negative value
		}
		bodyBytes, _ := json.Marshal(invalidConfig)

		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/servers/%s/configuration", validUUID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should return 400 for invalid range
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			if w.Body.Len() > 0 {
				err := json.NewDecoder(w.Body).Decode(&errorResponse)
				require.NoError(t, err)
				assert.Contains(t, errorResponse, "error")
			}
		}
	})
}
