package monitoring

import (
	"fmt"
	"sync"
	"time"

	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/Positronikal/MCPManager/internal/models"
	"github.com/Positronikal/MCPManager/internal/platform"
)

// MetricsCollector manages metrics collection for MCP servers
type MetricsCollector struct {
	processInfo    platform.ProcessInfo
	eventBus       *events.EventBus
	metricsCache   map[string]*cachedMetrics
	mu             sync.RWMutex
	rateLimitCache map[string]time.Time // serverID -> last update time
	rateLimitMu    sync.RWMutex
}

// cachedMetrics stores metrics with their server context
type cachedMetrics struct {
	metrics   *models.ServerMetrics
	pid       int
	startTime time.Time
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(processInfo platform.ProcessInfo, eventBus *events.EventBus) *MetricsCollector {
	return &MetricsCollector{
		processInfo:    processInfo,
		eventBus:       eventBus,
		metricsCache:   make(map[string]*cachedMetrics),
		rateLimitCache: make(map[string]time.Time),
	}
}

// GetMetrics retrieves current metrics for a server
// Returns metrics for running servers, nil metrics for stopped servers
// Rate-limited to 1Hz per server as per research.md §6
func (mc *MetricsCollector) GetMetrics(serverID string, status *models.ServerStatus, pid int) (*models.ServerMetrics, error) {
	// Check if server is running
	if status.State != models.StatusRunning {
		// Return empty metrics for non-running servers
		metrics := models.NewServerMetrics(serverID)
		return metrics, nil
	}

	// Check rate limit (1Hz = 1 update per second)
	if !mc.shouldUpdate(serverID) {
		// Return cached metrics if within rate limit
		return mc.getCachedMetrics(serverID), nil
	}

	// Collect fresh metrics
	metrics := models.NewServerMetrics(serverID)

	// Calculate uptime from LastStateChange
	if status.State == models.StatusRunning {
		metrics.Uptime = time.Since(status.LastStateChange)
	}

	// Get memory usage if PID is available
	if pid > 0 {
		memBytes, err := mc.processInfo.GetMemoryUsage(pid)
		if err == nil && memBytes > 0 {
			metrics.MemoryBytes = &memBytes
		}
		// Silently ignore errors - memory might not be available
	}

	// Request count is not implemented yet (would require MCP protocol support)
	// Set to nil for now

	// Cache the metrics
	mc.cacheMetrics(serverID, metrics, pid, status.LastStateChange)

	// Update rate limit timestamp
	mc.updateRateLimit(serverID)

	// Publish metrics updated event
	if mc.eventBus != nil {
		event := events.ServerMetricsUpdatedEvent(serverID, map[string]interface{}{
			"uptime":      metrics.Uptime.Seconds(),
			"memoryBytes": metrics.MemoryBytes,
		})
		mc.eventBus.Publish(event)
	}

	return metrics, nil
}

// shouldUpdate checks if we should update metrics based on rate limit
func (mc *MetricsCollector) shouldUpdate(serverID string) bool {
	mc.rateLimitMu.RLock()
	defer mc.rateLimitMu.RUnlock()

	lastUpdate, exists := mc.rateLimitCache[serverID]
	if !exists {
		return true
	}

	// Allow update if more than 1 second has passed
	return time.Since(lastUpdate) >= time.Second
}

// updateRateLimit updates the last update time for a server
func (mc *MetricsCollector) updateRateLimit(serverID string) {
	mc.rateLimitMu.Lock()
	defer mc.rateLimitMu.Unlock()

	mc.rateLimitCache[serverID] = time.Now()
}

// cacheMetrics stores metrics in the cache
func (mc *MetricsCollector) cacheMetrics(serverID string, metrics *models.ServerMetrics, pid int, startTime time.Time) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metricsCache[serverID] = &cachedMetrics{
		metrics:   metrics,
		pid:       pid,
		startTime: startTime,
	}
}

// getCachedMetrics retrieves cached metrics for a server
func (mc *MetricsCollector) getCachedMetrics(serverID string) *models.ServerMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	cached, exists := mc.metricsCache[serverID]
	if !exists {
		return models.NewServerMetrics(serverID)
	}

	// Update uptime in cached metrics
	metrics := *cached.metrics
	metrics.Uptime = time.Since(cached.startTime)
	metrics.Timestamp = time.Now()

	return &metrics
}

// ClearMetrics removes cached metrics for a server
func (mc *MetricsCollector) ClearMetrics(serverID string) {
	mc.mu.Lock()
	delete(mc.metricsCache, serverID)
	mc.mu.Unlock()

	mc.rateLimitMu.Lock()
	delete(mc.rateLimitCache, serverID)
	mc.rateLimitMu.Unlock()
}

// GetAllMetrics returns metrics for all cached servers
func (mc *MetricsCollector) GetAllMetrics() map[string]*models.ServerMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	result := make(map[string]*models.ServerMetrics)
	for serverID, cached := range mc.metricsCache {
		// Update uptime in cached metrics
		metrics := *cached.metrics
		metrics.Uptime = time.Since(cached.startTime)
		metrics.Timestamp = time.Now()
		result[serverID] = &metrics
	}

	return result
}

// UpdatePID updates the PID for a server's metrics collection
func (mc *MetricsCollector) UpdatePID(serverID string, pid int, startTime time.Time) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if cached, exists := mc.metricsCache[serverID]; exists {
		cached.pid = pid
		cached.startTime = startTime
	} else {
		// Create new cache entry
		metrics := models.NewServerMetrics(serverID)
		mc.metricsCache[serverID] = &cachedMetrics{
			metrics:   metrics,
			pid:       pid,
			startTime: startTime,
		}
	}
}

// ForceUpdate forces an immediate metrics update, bypassing rate limiting
func (mc *MetricsCollector) ForceUpdate(serverID string, status *models.ServerStatus, pid int) (*models.ServerMetrics, error) {
	// Clear rate limit for this server
	mc.rateLimitMu.Lock()
	delete(mc.rateLimitCache, serverID)
	mc.rateLimitMu.Unlock()

	return mc.GetMetrics(serverID, status, pid)
}

// GetLastUpdateTime returns the last time metrics were updated for a server
func (mc *MetricsCollector) GetLastUpdateTime(serverID string) (time.Time, error) {
	mc.rateLimitMu.RLock()
	defer mc.rateLimitMu.RUnlock()

	lastUpdate, exists := mc.rateLimitCache[serverID]
	if !exists {
		return time.Time{}, fmt.Errorf("no metrics update recorded for server %s", serverID)
	}

	return lastUpdate, nil
}
