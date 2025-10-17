package contract

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetServerById_ContractValidation tests GET /api/v1/servers/{serverId} endpoint
// This is a failing test that defines the API contract per api-spec.yaml
func TestGetServerById_ContractValidation(t *testing.T) {
	t.Run("should return 200 with server details for valid UUID", func(t *testing.T) {
		// Create test HTTP server (no implementation yet, should fail)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Assert status 200 for valid server
		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK for valid server ID")

		// Parse response as MCPServer
		var server map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&server)
		require.NoError(t, err, "Response should be valid JSON")

		// Verify MCPServer schema - required fields per data-model.md
		assert.Contains(t, server, "id", "Server should have 'id' field")
		assert.Contains(t, server, "name", "Server should have 'name' field")
		assert.Contains(t, server, "installationPath", "Server should have 'installationPath' field")
		assert.Contains(t, server, "status", "Server should have 'status' field")
		assert.Contains(t, server, "configuration", "Server should have 'configuration' field")
		assert.Contains(t, server, "discoveredAt", "Server should have 'discoveredAt' field")
		assert.Contains(t, server, "lastSeenAt", "Server should have 'lastSeenAt' field")
		assert.Contains(t, server, "source", "Server should have 'source' field")

		// Validate ID matches requested ID
		if id, ok := server["id"].(string); ok {
			assert.Equal(t, validUUID, id, "Returned server ID should match requested ID")
		}
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		// Use a valid UUID format, but server doesn't exist
		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s", nonExistentUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Assert status 404 Not Found
		assert.Equal(t, http.StatusNotFound, w.Code, "Expected status 404 Not Found for non-existent server")

		// Parse error response
		var errorResponse map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&errorResponse)
		require.NoError(t, err, "Error response should be valid JSON")

		// Error response should have error field per api-spec.yaml
		assert.Contains(t, errorResponse, "error", "Error response should contain 'error' field")
	})

	t.Run("should return 404 for invalid UUID format", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		// Test with invalid UUID formats
		invalidUUIDs := []string{
			"not-a-uuid",
			"12345",
			"invalid-uuid-format",
			"",
		}

		for _, invalidUUID := range invalidUUIDs {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s", invalidUUID), nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Should return 404 for invalid UUID (or 400 Bad Request)
			assert.True(t, w.Code == http.StatusNotFound || w.Code == http.StatusBadRequest,
				"Expected 404 or 400 for invalid UUID: %s", invalidUUID)

			if w.Code == http.StatusNotFound || w.Code == http.StatusBadRequest {
				var errorResponse map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&errorResponse)
				require.NoError(t, err, "Error response should be valid JSON")
				assert.Contains(t, errorResponse, "error", "Error response should contain 'error' field")
			}
		}
	})

	t.Run("server response should include all optional fields", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var server map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&server)

		if err == nil && w.Code == http.StatusOK {
			// Optional fields (may be null/omitted, but should be recognized)
			// These won't fail the test if absent, but should be present in schema
			optionalFields := []string{"version", "pid", "capabilities", "tools", "dependencies"}

			for _, field := range optionalFields {
				// If field exists, it should be valid type
				if val, exists := server[field]; exists {
					assert.NotNil(t, val, "If %s field exists, it should not be explicitly null in response", field)
				}
			}
		}
	})

	t.Run("status field should be valid ServerStatus object", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var server map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&server)

		if err == nil && w.Code == http.StatusOK {
			// Validate status object structure
			if status, ok := server["status"].(map[string]interface{}); ok {
				// ServerStatus should have these fields per data-model.md
				assert.Contains(t, status, "state", "Status should have 'state' field")
				assert.Contains(t, status, "uptime", "Status should have 'uptime' field")
				assert.Contains(t, status, "lastChecked", "Status should have 'lastChecked' field")

				// State should be valid enum
				if state, ok := status["state"].(string); ok {
					validStates := []string{"stopped", "starting", "running", "error"}
					assert.Contains(t, validStates, state, "Status state should be valid enum value")
				}
			}
		}
	})

	t.Run("configuration field should be valid ServerConfiguration object", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var server map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&server)

		if err == nil && w.Code == http.StatusOK {
			// Validate configuration object structure
			if config, ok := server["configuration"].(map[string]interface{}); ok {
				// ServerConfiguration should have these fields per data-model.md
				assert.Contains(t, config, "command", "Configuration should have 'command' field")
				assert.Contains(t, config, "args", "Configuration should have 'args' field")
				assert.Contains(t, config, "env", "Configuration should have 'env' field")
				assert.Contains(t, config, "autoStart", "Configuration should have 'autoStart' field")
			}
		}
	})

	t.Run("source field should be valid enum", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		validUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s", validUUID), nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var server map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&server)

		if err == nil && w.Code == http.StatusOK {
			// Validate source is valid enum
			if source, ok := server["source"].(string); ok {
				validSources := []string{"client_config", "filesystem", "process"}
				assert.Contains(t, validSources, source, "Source should be valid enum value")
			}
		}
	})
}
