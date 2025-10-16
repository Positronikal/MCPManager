package monitoring

import (
	"fmt"
	"testing"
	"time"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
)

// MockProcessInfo implements ProcessInfo for testing
type MockProcessInfo struct {
	memoryUsage map[int]uint64
	shouldError bool
}

func (m *MockProcessInfo) GetMemoryUsage(pid int) (uint64, error) {
	if m.shouldError {
		return 0, fmt.Errorf("mock error")
	}
	if mem, exists := m.memoryUsage[pid]; exists {
		return mem, nil
	}
	return 0, fmt.Errorf("process not found")
}

func TestNewMetricsCollector(t *testing.T) {
	mockPI := &MockProcessInfo{memoryUsage: make(map[int]uint64)}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	mc := NewMetricsCollector(mockPI, eventBus)

	if mc == nil {
		t.Fatal("Expected metrics collector to be created")
	}

	if mc.processInfo != mockPI {
		t.Error("Expected processInfo to be set")
	}

	if mc.eventBus != eventBus {
		t.Error("Expected eventBus to be set")
	}

	if mc.metricsCache == nil {
		t.Error("Expected metricsCache to be initialized")
	}

	if mc.rateLimitCache == nil {
		t.Error("Expected rateLimitCache to be initialized")
	}
}

func TestGetMetrics_StoppedServer(t *testing.T) {
	mockPI := &MockProcessInfo{memoryUsage: make(map[int]uint64)}
	mc := NewMetricsCollector(mockPI, nil)

	status := models.NewServerStatus()
	status.State = models.StatusStopped

	metrics, err := mc.GetMetrics("test-server", status, 0)

	if err != nil {
		t.Errorf("Expected no error for stopped server, got %v", err)
	}

	if metrics == nil {
		t.Fatal("Expected metrics to be returned")
	}

	if metrics.Uptime != 0 {
		t.Error("Expected uptime to be 0 for stopped server")
	}

	if metrics.MemoryBytes != nil {
		t.Error("Expected memory to be nil for stopped server")
	}
}

func TestGetMetrics_RunningServer(t *testing.T) {
	mockPI := &MockProcessInfo{
		memoryUsage: map[int]uint64{
			1234: 1024 * 1024 * 100, // 100 MB
		},
	}
	mc := NewMetricsCollector(mockPI, nil)

	status := models.NewServerStatus()
	status.State = models.StatusRunning
	status.LastStateChange = time.Now().Add(-5 * time.Minute)

	metrics, err := mc.GetMetrics("test-server", status, 1234)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if metrics == nil {
		t.Fatal("Expected metrics to be returned")
	}

	// Check uptime (should be around 5 minutes)
	if metrics.Uptime < 4*time.Minute || metrics.Uptime > 6*time.Minute {
		t.Errorf("Expected uptime around 5 minutes, got %v", metrics.Uptime)
	}

	// Check memory
	if metrics.MemoryBytes == nil {
		t.Fatal("Expected memory to be set")
	}

	expectedMem := uint64(1024 * 1024 * 100)
	if *metrics.MemoryBytes != expectedMem {
		t.Errorf("Expected memory %d, got %d", expectedMem, *metrics.MemoryBytes)
	}
}

func TestGetMetrics_MemoryError(t *testing.T) {
	mockPI := &MockProcessInfo{
		memoryUsage: make(map[int]uint64),
		shouldError: true,
	}
	mc := NewMetricsCollector(mockPI, nil)

	status := models.NewServerStatus()
	status.State = models.StatusRunning
	status.LastStateChange = time.Now().Add(-5 * time.Second)

	metrics, err := mc.GetMetrics("test-server", status, 1234)

	if err != nil {
		t.Errorf("Expected no error (memory errors should be silent), got %v", err)
	}

	if metrics == nil {
		t.Fatal("Expected metrics to be returned")
	}

	// Memory should be nil due to error
	if metrics.MemoryBytes != nil {
		t.Error("Expected memory to be nil when retrieval fails")
	}

	// But uptime should still be available
	if metrics.Uptime < 4*time.Second || metrics.Uptime > 6*time.Second {
		t.Errorf("Expected uptime around 5 seconds, got %v", metrics.Uptime)
	}
}

func TestGetMetrics_RateLimiting(t *testing.T) {
	mockPI := &MockProcessInfo{
		memoryUsage: map[int]uint64{
			1234: 1024 * 1024 * 50, // 50 MB
		},
	}
	mc := NewMetricsCollector(mockPI, nil)

	status := models.NewServerStatus()
	status.State = models.StatusRunning
	status.LastStateChange = time.Now().Add(-1 * time.Minute)

	// First call should update metrics
	metrics1, err := mc.GetMetrics("test-server", status, 1234)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Second call immediately after should return cached metrics
	metrics2, err := mc.GetMetrics("test-server", status, 1234)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Timestamps should be very close (cached)
	timeDiff := metrics2.Timestamp.Sub(metrics1.Timestamp)
	if timeDiff > 100*time.Millisecond {
		t.Errorf("Expected cached metrics, but timestamps differ by %v", timeDiff)
	}

	// Wait for rate limit to expire (1 second)
	time.Sleep(1100 * time.Millisecond)

	// Third call should update metrics again
	metrics3, err := mc.GetMetrics("test-server", status, 1234)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Timestamp should be fresh
	timeDiff = metrics3.Timestamp.Sub(metrics1.Timestamp)
	if timeDiff < 1*time.Second {
		t.Errorf("Expected fresh metrics after rate limit, but timestamps differ by only %v", timeDiff)
	}
}

func TestClearMetrics(t *testing.T) {
	mockPI := &MockProcessInfo{memoryUsage: make(map[int]uint64)}
	mc := NewMetricsCollector(mockPI, nil)

	status := models.NewServerStatus()
	status.State = models.StatusRunning
	status.LastStateChange = time.Now()

	// Get metrics to populate cache
	mc.GetMetrics("test-server", status, 1234)

	// Verify cache exists
	cached := mc.getCachedMetrics("test-server")
	if cached == nil {
		t.Error("Expected cached metrics to exist")
	}

	// Clear metrics
	mc.ClearMetrics("test-server")

	// Verify cache cleared
	cached = mc.getCachedMetrics("test-server")
	if cached.Uptime != 0 {
		t.Error("Expected cached metrics to be cleared")
	}
}

func TestGetAllMetrics(t *testing.T) {
	mockPI := &MockProcessInfo{
		memoryUsage: map[int]uint64{
			1234: 1024 * 1024 * 100,
			5678: 1024 * 1024 * 200,
		},
	}
	mc := NewMetricsCollector(mockPI, nil)

	status1 := models.NewServerStatus()
	status1.State = models.StatusRunning
	status1.LastStateChange = time.Now().Add(-5 * time.Minute)

	status2 := models.NewServerStatus()
	status2.State = models.StatusRunning
	status2.LastStateChange = time.Now().Add(-10 * time.Minute)

	// Collect metrics for two servers
	mc.GetMetrics("server-1", status1, 1234)
	mc.GetMetrics("server-2", status2, 5678)

	// Get all metrics
	allMetrics := mc.GetAllMetrics()

	if len(allMetrics) != 2 {
		t.Errorf("Expected 2 metrics, got %d", len(allMetrics))
	}

	if _, exists := allMetrics["server-1"]; !exists {
		t.Error("Expected metrics for server-1")
	}

	if _, exists := allMetrics["server-2"]; !exists {
		t.Error("Expected metrics for server-2")
	}
}

func TestUpdatePID(t *testing.T) {
	mockPI := &MockProcessInfo{memoryUsage: make(map[int]uint64)}
	mc := NewMetricsCollector(mockPI, nil)

	serverID := "test-server"
	startTime := time.Now().Add(-1 * time.Hour)

	// Update PID for non-existent server (should create new entry)
	mc.UpdatePID(serverID, 1234, startTime)

	// Verify cache entry exists
	mc.mu.RLock()
	cached, exists := mc.metricsCache[serverID]
	mc.mu.RUnlock()

	if !exists {
		t.Fatal("Expected cache entry to be created")
	}

	if cached.pid != 1234 {
		t.Errorf("Expected PID 1234, got %d", cached.pid)
	}

	if !cached.startTime.Equal(startTime) {
		t.Errorf("Expected start time %v, got %v", startTime, cached.startTime)
	}

	// Update PID again (should update existing entry)
	newStartTime := time.Now()
	mc.UpdatePID(serverID, 5678, newStartTime)

	mc.mu.RLock()
	cached, exists = mc.metricsCache[serverID]
	mc.mu.RUnlock()

	if !exists {
		t.Fatal("Expected cache entry to exist")
	}

	if cached.pid != 5678 {
		t.Errorf("Expected PID 5678, got %d", cached.pid)
	}

	if !cached.startTime.Equal(newStartTime) {
		t.Errorf("Expected start time %v, got %v", newStartTime, cached.startTime)
	}
}

func TestForceUpdate(t *testing.T) {
	mockPI := &MockProcessInfo{
		memoryUsage: map[int]uint64{
			1234: 1024 * 1024 * 100,
		},
	}
	mc := NewMetricsCollector(mockPI, nil)

	status := models.NewServerStatus()
	status.State = models.StatusRunning
	status.LastStateChange = time.Now()

	// First update
	metrics1, _ := mc.GetMetrics("test-server", status, 1234)

	// Immediate second update should be rate-limited
	metrics2, _ := mc.GetMetrics("test-server", status, 1234)

	// Timestamps should be close (cached)
	timeDiff := metrics2.Timestamp.Sub(metrics1.Timestamp)
	if timeDiff > 100*time.Millisecond {
		t.Errorf("Expected cached metrics, but timestamps differ by %v", timeDiff)
	}

	// Force update should bypass rate limit
	time.Sleep(100 * time.Millisecond)
	metrics3, _ := mc.ForceUpdate("test-server", status, 1234)

	// Timestamp should be fresh
	timeDiff = metrics3.Timestamp.Sub(metrics1.Timestamp)
	if timeDiff < 50*time.Millisecond {
		t.Errorf("Expected fresh metrics, but timestamps differ by only %v", timeDiff)
	}
}

func TestGetLastUpdateTime(t *testing.T) {
	mockPI := &MockProcessInfo{memoryUsage: make(map[int]uint64)}
	mc := NewMetricsCollector(mockPI, nil)

	status := models.NewServerStatus()
	status.State = models.StatusRunning
	status.LastStateChange = time.Now()

	// Should return error for server with no updates
	_, err := mc.GetLastUpdateTime("test-server")
	if err == nil {
		t.Error("Expected error for server with no updates")
	}

	// Get metrics to trigger update
	beforeUpdate := time.Now()
	mc.GetMetrics("test-server", status, 1234)
	afterUpdate := time.Now()

	// Should return update time
	lastUpdate, err := mc.GetLastUpdateTime("test-server")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if lastUpdate.Before(beforeUpdate) || lastUpdate.After(afterUpdate) {
		t.Errorf("Expected update time between %v and %v, got %v", beforeUpdate, afterUpdate, lastUpdate)
	}
}

func TestGetMetrics_EventPublishing(t *testing.T) {
	mockPI := &MockProcessInfo{
		memoryUsage: map[int]uint64{
			1234: 1024 * 1024 * 100,
		},
	}
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	mc := NewMetricsCollector(mockPI, eventBus)

	// Subscribe to metrics events
	metricsChan := eventBus.Subscribe(events.EventServerMetricsUpdated)

	status := models.NewServerStatus()
	status.State = models.StatusRunning
	status.LastStateChange = time.Now()

	// Get metrics should trigger event
	mc.GetMetrics("test-server", status, 1234)

	// Wait for event
	select {
	case event := <-metricsChan:
		if event.Type != events.EventServerMetricsUpdated {
			t.Errorf("Expected EventServerMetricsUpdated, got %s", event.Type)
		}
		if event.Data["serverID"] != "test-server" {
			t.Errorf("Expected serverID 'test-server', got %v", event.Data["serverID"])
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Expected metrics event but none received")
	}
}

func TestGetMetrics_UptimeCalculation(t *testing.T) {
	mockPI := &MockProcessInfo{memoryUsage: make(map[int]uint64)}
	mc := NewMetricsCollector(mockPI, nil)

	status := models.NewServerStatus()
	status.State = models.StatusRunning
	status.LastStateChange = time.Now().Add(-2 * time.Hour)

	metrics, err := mc.GetMetrics("test-server", status, 0)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Uptime should be around 2 hours
	expectedUptime := 2 * time.Hour
	tolerance := 1 * time.Second

	if metrics.Uptime < expectedUptime-tolerance || metrics.Uptime > expectedUptime+tolerance {
		t.Errorf("Expected uptime around %v, got %v", expectedUptime, metrics.Uptime)
	}
}

func TestGetMetrics_CachedUptimeUpdate(t *testing.T) {
	mockPI := &MockProcessInfo{memoryUsage: make(map[int]uint64)}
	mc := NewMetricsCollector(mockPI, nil)

	status := models.NewServerStatus()
	status.State = models.StatusRunning
	status.LastStateChange = time.Now().Add(-1 * time.Minute)

	// First call
	metrics1, _ := mc.GetMetrics("test-server", status, 0)

	// Wait a bit
	time.Sleep(500 * time.Millisecond)

	// Second call (cached, but uptime should be updated)
	metrics2, _ := mc.GetMetrics("test-server", status, 0)

	// Uptime should have increased
	if metrics2.Uptime <= metrics1.Uptime {
		t.Errorf("Expected uptime to increase from %v to %v", metrics1.Uptime, metrics2.Uptime)
	}

	uptimeDiff := metrics2.Uptime - metrics1.Uptime
	if uptimeDiff < 400*time.Millisecond || uptimeDiff > 600*time.Millisecond {
		t.Errorf("Expected uptime difference around 500ms, got %v", uptimeDiff)
	}
}
