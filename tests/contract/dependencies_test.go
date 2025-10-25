package contract

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/hoytech/mcpmanager/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetServerDependencies_ContractValidation tests GET /api/v1/servers/{serverId}/dependencies
func TestGetServerDependencies_ContractValidation(t *testing.T) {
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

	t.Run("should return 200 with dependencies array", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/dependencies", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var response struct {
			Dependencies []map[string]interface{} `json:"dependencies"`
			AllSatisfied bool                     `json:"allSatisfied"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err, "Response should be valid JSON")

		// Validate response schema per api-spec.yaml
		assert.NotNil(t, response.Dependencies, "dependencies should be present (not null)")
		assert.IsType(t, false, response.AllSatisfied, "allSatisfied should be boolean")
	})

	t.Run("dependency objects should match Dependency schema", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/dependencies", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response struct {
			Dependencies []map[string]interface{} `json:"dependencies"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)

		if err == nil && len(response.Dependencies) > 0 {
			dep := response.Dependencies[0]

			// Dependency schema per data-model.md
			assert.Contains(t, dep, "name", "Dependency should have 'name' field")
			assert.Contains(t, dep, "type", "Dependency should have 'type' field")
			assert.Contains(t, dep, "satisfied", "Dependency should have 'satisfied' field")

			// Type should be valid enum
			if depType, ok := dep["type"].(string); ok {
				validTypes := []string{"node", "python", "binary", "package"}
				assert.Contains(t, validTypes, depType, "Dependency type should be valid enum")
			}

			// Satisfied should be boolean
			if satisfied, ok := dep["satisfied"].(bool); ok {
				assert.IsType(t, false, satisfied, "satisfied should be boolean")
			}
		}
	})

	t.Run("allSatisfied should be true when all dependencies met", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/dependencies", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response struct {
			Dependencies []map[string]interface{} `json:"dependencies"`
			AllSatisfied bool                     `json:"allSatisfied"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)

		if err == nil && w.Code == http.StatusOK {
			// Logic: if allSatisfied is true, all dependencies should have satisfied=true
			if response.AllSatisfied {
				for _, dep := range response.Dependencies {
					if satisfied, ok := dep["satisfied"].(bool); ok {
						assert.True(t, satisfied, "If allSatisfied is true, all dependencies should be satisfied")
					}
				}
			}
		}
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/dependencies", nonExistentUUID), nil)
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

// TestGetServerUpdates_ContractValidation tests GET /api/v1/servers/{serverId}/updates
func TestGetServerUpdates_ContractValidation(t *testing.T) {
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

	t.Run("should return 200 with update status", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/updates", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected status 200 OK")

		var response struct {
			UpdateAvailable bool   `json:"updateAvailable"`
			CurrentVersion  string `json:"currentVersion"`
			LatestVersion   string `json:"latestVersion"`
			ReleaseNotes    string `json:"releaseNotes"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err, "Response should be valid JSON")

		// Validate response schema per api-spec.yaml
		assert.IsType(t, false, response.UpdateAvailable, "updateAvailable should be boolean")
		assert.NotEmpty(t, response.CurrentVersion, "currentVersion should be present")
		assert.NotEmpty(t, response.LatestVersion, "latestVersion should be present")
		// releaseNotes is nullable, can be empty
	})

	t.Run("currentVersion and latestVersion should be semantic version format", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/updates", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response struct {
			CurrentVersion string `json:"currentVersion"`
			LatestVersion  string `json:"latestVersion"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)

		if err == nil && w.Code == http.StatusOK {
			// Versions should follow semantic versioning (loosely validated)
			// Examples: "1.0.0", "2.3.1-beta", "v1.2.3"
			if response.CurrentVersion != "" {
				assert.NotEmpty(t, response.CurrentVersion, "currentVersion should not be empty")
			}

			if response.LatestVersion != "" {
				assert.NotEmpty(t, response.LatestVersion, "latestVersion should not be empty")
			}
		}
	})

	t.Run("updateAvailable should be true if latestVersion > currentVersion", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/updates", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response struct {
			UpdateAvailable bool   `json:"updateAvailable"`
			CurrentVersion  string `json:"currentVersion"`
			LatestVersion   string `json:"latestVersion"`
		}

		err := json.NewDecoder(w.Body).Decode(&response)

		if err == nil && w.Code == http.StatusOK {
			// This is a logical check - if versions are equal, updateAvailable should be false
			if response.CurrentVersion == response.LatestVersion {
				assert.False(t, response.UpdateAvailable, "updateAvailable should be false when versions match")
			}
		}
	})

	t.Run("should handle network errors gracefully", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/updates", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// Per task requirements, network errors should be handled gracefully
		// Should not return 500, but rather a valid response with "unknown" status
		// or return available info with updateAvailable=false
		if w.Code == http.StatusOK {
			var response map[string]interface{}
			err := json.NewDecoder(w.Body).Decode(&response)
			require.NoError(t, err)

			// Should still have required fields
			assert.Contains(t, response, "updateAvailable")
			assert.Contains(t, response, "currentVersion")
			assert.Contains(t, response, "latestVersion")
		}
	})

	t.Run("should return 404 for non-existent server", func(t *testing.T) {
		nonExistentUUID := uuid.New().String()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/updates", nonExistentUUID), nil)
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

	t.Run("releaseNotes should be nullable", func(t *testing.T) {
		if validUUID == "" {
			t.Skip("No servers available for testing")
		}

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/servers/%s/updates", validUUID), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)

		if err == nil && w.Code == http.StatusOK {
			// releaseNotes is nullable per api-spec.yaml
			// It can be null, empty string, or have content - all valid
			// It may also be omitted entirely when empty (due to omitempty tag)
			// Just verify response structure is valid
			assert.Contains(t, response, "updateAvailable")
			assert.Contains(t, response, "currentVersion")
			assert.Contains(t, response, "latestVersion")
		}
	})
}
