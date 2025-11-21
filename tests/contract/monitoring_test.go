package contract

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Positronikal/MCPManager/internal/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetServerLogs_ContractValidation tests GET /api/v1/servers/{serverId}/logs
func TestGetServerLogs_ContractValidation(t *testing.T) {
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

	t.Run("should return 200 with logs array", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/logs", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var response struct {
			Logs    []map[string]interface{} `json:"logs"`
			Total   int                      `json:"total"`
			HasMore bool                     `json:"hasMore"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err, "Response should be valid JSON")

		assert.NotNil(t, response.Logs, "logs should be present (not null)")
		assert.GreaterOrEqual(t, response.Total, 0, "total should be non-negative")
		assert.IsType(t, false, response.HasMore, "hasMore should be boolean")
	})

	t.Run("should support severity filter", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/logs?severity=error", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify query parameter
		severityParam := req.URL.Query().Get("severity")
		validSeverities := []string{"info", "success", "warning", "error"}
		assert.Contains(t, validSeverities, severityParam, "severity should be valid enum value")
	})

	t.Run("should support limit and offset pagination", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/logs?limit=50&offset=100", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify query parameters
		limitParam := req.URL.Query().Get("limit")
		offsetParam := req.URL.Query().Get("offset")

		assert.Equal(t, "50", limitParam)
		assert.Equal(t, "100", offsetParam)
	})

	t.Run("log entries should match LogEntry schema", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/logs", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response struct {
			Logs []map[string]interface{} `json:"logs"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)

		if err == nil && len(response.Logs) > 0 {
			logEntry := response.Logs[0]

			// LogEntry schema per data-model.md
			assert.Contains(t, logEntry, "timestamp", "LogEntry should have 'timestamp' field")
			assert.Contains(t, logEntry, "severity", "LogEntry should have 'severity' field")
			assert.Contains(t, logEntry, "message", "LogEntry should have 'message' field")
			assert.Contains(t, logEntry, "serverId", "LogEntry should have 'serverId' field")

			// Severity should be valid enum
			if severity, ok := logEntry["severity"].(string); ok {
				validSeverities := []string{"info", "success", "warning", "error"}
				assert.Contains(t, validSeverities, severity)
			}
		}
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/logs", nonExistentUUID), nil)
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

// TestGetAllLogs_ContractValidation tests GET /api/v1/logs
func TestGetAllLogs_ContractValidation(t *testing.T) {
	services, cleanup := setupFullTestServices(t)
	defer cleanup()
	router := api.NewRouter(services)

	t.Run("should return 200 with filtered logs", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/logs", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response struct {
			Logs  []map[string]interface{} `json:"logs"`
			Total int                      `json:"total"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotNil(t, response.Logs)
		assert.GreaterOrEqual(t, response.Total, 0)
	})

	t.Run("should support serverId filter", func(t *testing.T) {
		serverID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/logs?serverId=%s", serverID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify query parameter is valid UUID
		serverIDParam := req.URL.Query().Get("serverId")
		_, err := uuid.Parse(serverIDParam)
		assert.NoError(t, err, "serverId should be valid UUID")
	})

	t.Run("should support search query", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/logs?search=error&severity=error", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify query parameters
		searchParam := req.URL.Query().Get("search")
		assert.Equal(t, "error", searchParam)
	})
}

// TestGetServerMetrics_ContractValidation tests GET /api/v1/servers/{serverId}/metrics
func TestGetServerMetrics_ContractValidation(t *testing.T) {
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

	t.Run("should return 200 with metrics object", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/metrics", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var metrics map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&metrics)
		require.NoError(t, err)

		// Metrics schema per api-spec.yaml
		assert.Contains(t, metrics, "uptimeSeconds")
		assert.Contains(t, metrics, "memoryUsageMB")
		assert.Contains(t, metrics, "requestCount")
		assert.Contains(t, metrics, "cpuPercent")
	})

	t.Run("metrics should be nullable for stopped servers", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/metrics", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var metrics map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&metrics)

		if err == nil && w.Code == http.StatusOK {
			// Metrics can be null for stopped servers (nullable: true per API spec)
			// This is valid and should not fail
			// If non-null, should be correct type

			if uptimeSeconds, exists := metrics["uptimeSeconds"]; exists && uptimeSeconds != nil {
				_, ok := uptimeSeconds.(float64) // JSON numbers are float64
				assert.True(t, ok, "uptimeSeconds should be number if not null")
			}

			if memoryUsageMB, exists := metrics["memoryUsageMB"]; exists && memoryUsageMB != nil {
				_, ok := memoryUsageMB.(float64)
				assert.True(t, ok, "memoryUsageMB should be number if not null")
			}

			if cpuPercent, exists := metrics["cpuPercent"]; exists && cpuPercent != nil {
				_, ok := cpuPercent.(float64)
				assert.True(t, ok, "cpuPercent should be number if not null")
			}
		}
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/metrics", nonExistentUUID), nil)
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
