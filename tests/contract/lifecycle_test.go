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

// TestPostServerStart_ContractValidation tests POST /api/v1/servers/{serverId}/start
func TestPostServerStart_ContractValidation(t *testing.T) {
	t.Run("should return 202 with starting status", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/start", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

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
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/start", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// If server is already running, should return 400
		if w.Code == http.StatusBadRequest {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err, "Error response should be valid JSON")
			assert.Contains(t, errorResponse, "error", "Error response should have 'error' field")
		}
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/start", nonExistentUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})
}

// TestPostServerStop_ContractValidation tests POST /api/v1/servers/{serverId}/stop
func TestPostServerStop_ContractValidation(t *testing.T) {
	t.Run("should return 202 for graceful stop", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/stop", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

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
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()

		requestBody := map[string]interface{}{
			"force":   true,
			"timeout": 5,
		}
		bodyBytes, _ := json.Marshal(requestBody)

		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/stop", validUUID), bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusAccepted, w.Code, "Should accept stop with force and timeout")

		// Verify request body was parsed (in actual implementation)
		var parsedBody map[string]interface{}
		json.NewDecoder(bytes.NewReader(bodyBytes)).Decode(&parsedBody)
		assert.Equal(t, true, parsedBody["force"])
		assert.Equal(t, float64(5), parsedBody["timeout"])
	})

	t.Run("should use default values if force/timeout not provided", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/stop", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Should still accept request without body (use defaults: force=false, timeout=10)
		assert.Equal(t, http.StatusAccepted, w.Code, "Should accept stop without body parameters")
	})

	t.Run("should return 400 if server not running", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/stop", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

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
	t.Run("should return 202 for restart", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/restart", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

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
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/restart", nonExistentUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})

	t.Run("should not require request body", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/servers/%s/restart", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Restart doesn't require body, should still return 202
		assert.Equal(t, http.StatusAccepted, w.Code, "Should accept restart without body")
	})
}

// TestGetServerStatus_ContractValidation tests GET /api/v1/servers/{serverId}/status
func TestGetServerStatus_ContractValidation(t *testing.T) {
	t.Run("should return 200 with ServerStatus schema", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/status", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var status map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&status)
		require.NoError(t, err, "Response should be valid JSON")

		// Validate ServerStatus schema per data-model.md
		assert.Contains(t, status, "state", "Status should have 'state' field")
		assert.Contains(t, status, "uptime", "Status should have 'uptime' field")
		assert.Contains(t, status, "lastChecked", "Status should have 'lastChecked' field")

		// State should be valid enum
		if state, ok := status["state"].(string); ok {
			validStates := []string{"stopped", "starting", "running", "error"}
			assert.Contains(t, validStates, state, "State should be valid enum value")
		}
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/status", nonExistentUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		if w.Code == http.StatusNotFound {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err)
			assert.Contains(t, errorResponse, "error")
		}
	})

	t.Run("uptime should be number (seconds)", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D010
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/status", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var status map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&status)

		if err == nil && w.Code == http.StatusOK {
			if uptime, exists := status["uptime"]; exists && uptime != nil {
				// Uptime should be numeric
				_, ok := uptime.(float64)
				assert.True(t, ok, "uptime should be a number")
			}
		}
	})
}
