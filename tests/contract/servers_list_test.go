package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetServers_ContractValidation tests GET /api/v1/servers endpoint
// This is a failing test that defines the API contract per api-spec.yaml
func TestGetServers_ContractValidation(t *testing.T) {
	t.Run("should return 200 with servers list", func(t *testing.T) {
		// Create test HTTP server (no implementation yet, should fail)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		// Assert status 200
		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		// Parse response
		var response struct {
			Servers       []map[string]interface{} `json:"servers"`
			Count         int                      `json:"count"`
			LastDiscovery string                   `json:"lastDiscovery"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err, "Response should be valid JSON")

		// Assert response schema
		assert.NotNil(t, response.Servers, "servers field should be present")
		assert.GreaterOrEqual(t, response.Count, 0, "count should be non-negative")
		assert.NotEmpty(t, response.LastDiscovery, "lastDiscovery should be present")

		// Validate lastDiscovery is ISO 8601 format
		_, err = time.Parse(time.RFC3339, response.LastDiscovery)
		assert.NoError(t, err, "lastDiscovery should be in ISO 8601 format (RFC3339)")
	})

	t.Run("should support status filter query parameter", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		// Test with status=running query parameter
		req := httptest.NewRequest(http.MethodGet, "/api/v1/servers?status=running", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		// Verify query parameter was parsed
		statusParam := req.URL.Query().Get("status")
		assert.Equal(t, "running", statusParam, "status query parameter should be 'running'")

		// Valid status values per api-spec.yaml
		validStatuses := []string{"stopped", "starting", "running", "error"}
		assert.Contains(t, validStatuses, statusParam, "status should be one of the valid enum values")
	})

	t.Run("should support source filter query parameter", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		// Test with source=client_config query parameter
		req := httptest.NewRequest(http.MethodGet, "/api/v1/servers?source=client_config", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		// Verify query parameter was parsed
		sourceParam := req.URL.Query().Get("source")
		assert.Equal(t, "client_config", sourceParam, "source query parameter should be 'client_config'")

		// Valid source values per api-spec.yaml
		validSources := []string{"client_config", "filesystem", "process"}
		assert.Contains(t, validSources, sourceParam, "source should be one of the valid enum values")
	})

	t.Run("should support combined status and source filters", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		// Test with both query parameters
		req := httptest.NewRequest(http.MethodGet, "/api/v1/servers?status=running&source=filesystem", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		// Verify both query parameters were parsed
		statusParam := req.URL.Query().Get("status")
		sourceParam := req.URL.Query().Get("source")

		assert.Equal(t, "running", statusParam)
		assert.Equal(t, "filesystem", sourceParam)
	})

	t.Run("should return empty array when no servers found", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK even with no servers")

		var response struct {
			Servers       []interface{} `json:"servers"`
			Count         int           `json:"count"`
			LastDiscovery string        `json:"lastDiscovery"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		// Empty array is valid, should not be null
		assert.NotNil(t, response.Servers, "servers should be an array, not null")
		assert.Equal(t, 0, response.Count, "count should be 0 when no servers")
	})
}

// TestGetServers_SchemaValidation validates the MCPServer schema in responses
func TestGetServers_SchemaValidation(t *testing.T) {
	t.Run("each server should match MCPServer schema", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// This will be implemented in T-D009
			w.WriteHeader(http.StatusNotImplemented)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		var response struct {
			Servers []map[string]interface{} `json:"servers"`
			Count   int                      `json:"count"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		if err == nil && len(response.Servers) > 0 {
			// Validate first server has required fields per data-model.md
			server := response.Servers[0]

			// Required fields
			assert.Contains(t, server, "id", "Server should have 'id' field")
			assert.Contains(t, server, "name", "Server should have 'name' field")
			assert.Contains(t, server, "installationPath", "Server should have 'installationPath' field")
			assert.Contains(t, server, "status", "Server should have 'status' field")
			assert.Contains(t, server, "configuration", "Server should have 'configuration' field")
			assert.Contains(t, server, "discoveredAt", "Server should have 'discoveredAt' field")
			assert.Contains(t, server, "lastSeenAt", "Server should have 'lastSeenAt' field")
			assert.Contains(t, server, "source", "Server should have 'source' field")

			// Validate ID is UUID format
			if id, ok := server["id"].(string); ok {
				assert.Regexp(t, `^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
					id, "Server ID should be valid UUID")
			}

			// Validate source is valid enum
			if source, ok := server["source"].(string); ok {
				validSources := []string{"client_config", "filesystem", "process"}
				assert.Contains(t, validSources, source, "Source should be valid enum value")
			}
		}
	})
}
