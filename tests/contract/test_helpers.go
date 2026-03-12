package contract

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/api"
	"github.com/Positronikal/MCPManager/internal/core/discovery"
	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/Positronikal/MCPManager/internal/platform"
)

// createTestRouter creates a router with test services for contract testing
func createTestRouter() *api.Services {
	pathResolver := platform.NewPathResolver()
	eventBus := events.NewEventBus()
	discoveryService := discovery.NewDiscoveryService(pathResolver, eventBus)

	// Run initial discovery to have test data
	discoveryService.Discover()

	return &api.Services{
		DiscoveryService: discoveryService,
		EventBus:         eventBus,
	}
}

// waitForServerState polls the server status endpoint until the server
// reaches the desired state or the timeout is exceeded.
// Returns the final state observed, or fails the test on timeout.
func waitForServerState(t *testing.T, router http.Handler, serverID string, desiredState string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var lastState string

	for time.Now().Before(deadline) {
		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/api/v1/servers/%s/status", serverID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			var status struct {
				State string `json:"state"`
			}
			if err := json.NewDecoder(w.Body).Decode(&status); err == nil {
				lastState = status.State
				if lastState == desiredState {
					return lastState
				}
			}
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("Timed out waiting for server %s to reach state %q (last observed: %q)",
		serverID, desiredState, lastState)
	return lastState
}

// waitForStableState polls the server status endpoint until the server
// reaches any stable (non-transitional) state: "running", "stopped", or "error".
// Returns the final stable state. Fails the test only on timeout.
func waitForStableState(t *testing.T, router http.Handler, serverID string, timeout time.Duration) string {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var lastState string

	for time.Now().Before(deadline) {
		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/api/v1/servers/%s/status", serverID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			var status struct {
				State string `json:"state"`
			}
			if err := json.NewDecoder(w.Body).Decode(&status); err == nil {
				lastState = status.State
				if lastState == "running" || lastState == "stopped" || lastState == "error" {
					return lastState
				}
			}
		}

		time.Sleep(50 * time.Millisecond)
	}

	t.Fatalf("Timed out waiting for server %s to reach a stable state (last observed: %q)",
		serverID, lastState)
	return lastState
}

// ensureServerStopped attempts to stop the server and waits for it to
// reach the stopped state. Safe to call even if server is already stopped.
func ensureServerStopped(t *testing.T, router http.Handler, serverID string) {
	t.Helper()

	// Fire a stop request (ignore response — server may already be stopped)
	stopReq := httptest.NewRequest(http.MethodPost,
		fmt.Sprintf("/api/v1/servers/%s/stop", serverID), nil)
	stopW := httptest.NewRecorder()
	router.ServeHTTP(stopW, stopReq)

	// Wait for stopped state (or error, which is also a terminal state we can work with)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		req := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/api/v1/servers/%s/status", serverID), nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			var status struct {
				State string `json:"state"`
			}
			if err := json.NewDecoder(w.Body).Decode(&status); err == nil {
				if status.State == "stopped" || status.State == "error" {
					return
				}
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
}
