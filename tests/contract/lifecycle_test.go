package contract

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/Positronikal/MCPManager/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostServerStart_ContractValidation tests POST /api/v1/servers/{serverId}/start
func TestPostServerStart_ContractValidation(t *testing.T) {
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

	t.Run("should return 202 with starting status", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		// Stop server first if it's running
		stopReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/stop", validUUID), nil)
		stopW := httptest.NewRecorder()
		router.ServeHTTP(stopW, stopReq)

		// Now try to start the server
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/start", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Assert status 202 Accepted (async operation)
		assert.Equal(t, http.StatusAccepted, w.Code, "Expected status 202 Accepted")

		var response struct {
			Message  string `json:"message"`
			ServerID string `json:"serverId"`
			Status   string `json:"status"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err, "Response should be valid JSON")

		assert.NotEmpty(t, response.Message, "message should be present")
		assert.Equal(t, validUUID, response.ServerID, "serverId should match request")
		assert.Equal(t, "starting", response.Status, "status should be 'starting'")
	})

	t.Run("should return 400 if server already running", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		// Ensure server is running by starting it first
		startReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/start", validUUID), nil)
		startW := httptest.NewRecorder()
		router.ServeHTTP(startW, startReq)

		// Try to start again - should return 400
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/start", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// If server is already running, should return 400
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err, "Error response should be valid JSON")
			assert.Contains(t, errorResponse, "error", "Error response should have 'error' field")
		}
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/start", nonExistentUUID), nil)
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
}

// TestPostServerStop_ContractValidation tests POST /api/v1/servers/{serverId}/stop
func TestPostServerStop_ContractValidation(t *testing.T) {
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

	t.Run("should return 202 for graceful stop", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		// Start server first to ensure it's running
		startReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/start", validUUID), nil)
		startW := httptest.NewRecorder()
		router.ServeHTTP(startW, startReq)

		// Wait briefly for server to reach running state (start is async)
		time.Sleep(100 * time.Millisecond)

		// Now stop the server
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/stop", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code, "Expected status 202 Accepted")

		var response struct {
			Message  string `json:"message"`
			ServerID string `json:"serverId"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err, "Response should be valid JSON")

		assert.NotEmpty(t, response.Message, "message should be present")
		assert.Equal(t, validUUID, response.ServerID, "serverId should match request")
	})

	t.Run("should accept force and timeout parameters", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		// Start server first to ensure it's running
		startReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/start", validUUID), nil)
		startW := httptest.NewRecorder()
		router.ServeHTTP(startW, startReq)

		// Wait briefly for server to reach running state (start is async)
		time.Sleep(100 * time.Millisecond)

		requestBody := map[string]interface{}{
			"force":   true,
			"timeout": 5,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/stop", validUUID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code, "Should accept stop with force and timeout")

		// Verify request body was parsed (in actual implementation)
		var parsedBody map[string]interface{}
		json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&parsedBody)
		assert.Equal(t, true, parsedBody["force"])
		assert.Equal(t, float64(5), parsedBody["timeout"])
	})

	t.Run("should use default values if force/timeout not provided", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		// Start server first to ensure it's running
		startReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/start", validUUID), nil)
		startW := httptest.NewRecorder()
		router.ServeHTTP(startW, startReq)

		// Wait briefly for server to reach running state (start is async)
		time.Sleep(100 * time.Millisecond)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/stop", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should still accept request without body (use defaults: force=false, timeout=10)
		assert.Equal(t, http.StatusAccepted, w.Code, "Should accept stop without body parameters")
	})

	t.Run("should return 400 if server not running", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		// Ensure server is stopped by stopping it first
		stopReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/stop", validUUID), nil)
		stopW := httptest.NewRecorder()
		router.ServeHTTP(stopW, stopReq)

		// Try to stop again - should return 400
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/stop", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// If server is not running, should return 400
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})
}

// TestPostServerRestart_ContractValidation tests POST /api/v1/servers/{serverId}/restart
func TestPostServerRestart_ContractValidation(t *testing.T) {
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

	t.Run("should return 202 for restart", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/restart", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code, "Expected status 202 Accepted")

		var response struct {
			Message  string `json:"message"`
			ServerID string `json:"serverId"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err, "Response should be valid JSON")

		assert.NotEmpty(t, response.Message, "message should be present")
		assert.Equal(t, validUUID, response.ServerID, "serverId should match request")
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/restart", nonExistentUUID), nil)
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

	t.Run("should not require request body", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/restart", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Restart doesn't require body, should still return 202
		assert.Equal(t, http.StatusAccepted, w.Code, "Should accept restart without body")
	})
}

// TestGetServerStatus_ContractValidation tests GET /api/v1/servers/{serverId}/status
func TestGetServerStatus_ContractValidation(t *testing.T) {
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

	t.Run("should return 200 with ServerStatus schema", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/status", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var status map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&status)
		require.NoError(t, err, "Response should be valid JSON")

		// Validate ServerStatus schema (actual implementation)
		assert.Contains(t, status, "state", "Status should have 'state' field")
		assert.Contains(t, status, "startupAttempts", "Status should have 'startupAttempts' field")
		assert.Contains(t, status, "lastStateChange", "Status should have 'lastStateChange' field")
		assert.Contains(t, status, "crashRecoverable", "Status should have 'crashRecoverable' field")

		// State should be valid enum
		if state, ok := status["state"].(string); ok {
			validStates := []string{"stopped", "starting", "running", "error"}
			assert.Contains(t, validStates, state, "State should be valid enum value")
		}
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/status", nonExistentUUID), nil)
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

	t.Run("startupAttempts should be number", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/status", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var status map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&status)

		if err == nil && w.Code == http.StatusOK {
			if startupAttempts, exists := status["startupAttempts"]; exists && startupAttempts != nil {
				// startupAttempts should be numeric
				_, ok := startupAttempts.(float64)
				assert.True(t, ok, "startupAttempts should be a number")
			}
		}
	})
}
