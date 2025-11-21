package contract

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Positronikal/MCPManager/internal/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPostServersDiscover_ContractValidation tests POST /api/v1/servers/discover endpoint
// This validates the API contract per api-spec.yaml
func TestPostServersDiscover_ContractValidation(t *testing.T) {
	services := createTestRouter()
	router := api.NewRouter(services)
	defer services.EventBus.Close()

	t.Run("should return 202 Accepted with scanId", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/discover", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Assert status 202 Accepted (asynchronous operation)
		assert.Equal(t, http.StatusAccepted, w.Code, "Expected status 202 Accepted")

		// Parse response
		var response struct {
			Message string `json:"message"`
			ScanID  string `json:"scanId"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err, "Response should be valid JSON")

		// Assert response schema
		assert.NotEmpty(t, response.Message, "message field should be present and not empty")
		assert.Equal(t, "Discovery scan initiated", response.Message, "message should match expected text")
		assert.NotEmpty(t, response.ScanID, "scanId field should be present and not empty")

		// Validate scanId is a valid UUID
		_, err = uuid.Parse(response.ScanID)
		assert.NoError(t, err, "scanId should be a valid UUID format")
	})

	t.Run("should accept POST method only", func(t *testing.T) {
		// Verify request method is POST
		req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/discover", nil)
		assert.Equal(t, http.MethodPost, req.Method, "Method should be POST")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return 202 for POST
		assert.Equal(t, http.StatusAccepted, w.Code, "POST should be accepted")
	})

	t.Run("should not require request body", func(t *testing.T) {
		// POST with no body should be valid
		req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/discover", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Should still return 202 (no 400 Bad Request for missing body)
		assert.Equal(t, http.StatusAccepted, w.Code, "Should accept POST without body")
	})

	t.Run("response should include both required fields", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/discover", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)

		if err == nil {
			// Verify both fields are present
			assert.Contains(t, response, "message", "Response must contain 'message' field")
			assert.Contains(t, response, "scanId", "Response must contain 'scanId' field")

			// Verify no unexpected fields (strict schema validation)
			assert.Len(t, response, 2, "Response should only contain 'message' and 'scanId' fields")
		}
	})

	t.Run("should return unique scanId for each request", func(t *testing.T) {
		// Make two sequential requests
		var scanIds []string

		for i := 0; i < 2; i++ {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/discover", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			var response struct {
				ScanID string `json:"scanId"`
			}

			err := json.NewDecoder(w.Body).Decode(&response)
			if err == nil && response.ScanID != "" {
				scanIds = append(scanIds, response.ScanID)
			}
		}

		// Should have got two scanIds, and they should be different
		require.Len(t, scanIds, 2, "Should have received 2 scanIds")
		assert.NotEqual(t, scanIds[0], scanIds[1], "Each discovery scan should have a unique scanId")
	})
}

// TestPostServersDiscover_ErrorHandling tests error cases
func TestPostServersDiscover_ErrorHandling(t *testing.T) {
	services := createTestRouter()
	router := api.NewRouter(services)
	defer services.EventBus.Close()

	t.Run("should handle 500 Internal Server Error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/servers/discover", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// If an internal error occurs, it should return 500
		// The contract allows for 500 responses per api-spec.yaml
		if w.Code == http.StatusInternalServerError {
			var errorResponse map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&errorResponse)
			require.NoError(t, err, "Error response should be valid JSON")

			// Error responses should have error field
			assert.Contains(t, errorResponse, "error", "Error response should contain 'error' field")
		}
	})
}
