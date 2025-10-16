package monitoring

import (
	"bufio"
	"context"
	"io"
	"strings"
	"sync"

	"github.com/hoytech/mcpmanager/internal/core/events"
	"github.com/hoytech/mcpmanager/internal/models"
)

// MonitoringService manages log capture and monitoring for MCP servers
type MonitoringService struct {
	logBuffers map[string]*models.CircularLogBuffer // serverID -> buffer
	eventBus   *events.EventBus
	mu         sync.RWMutex
}

// NewMonitoringService creates a new monitoring service
func NewMonitoringService(eventBus *events.EventBus) *MonitoringService {
	return &MonitoringService{
		logBuffers: make(map[string]*models.CircularLogBuffer),
		eventBus:   eventBus,
	}
}

// CaptureOutput captures stdout/stderr from a server process
// Reads lines from the provided reader, parses severity, and stores in buffer
// This method blocks until the reader is closed or context is cancelled
func (ms *MonitoringService) CaptureOutput(ctx context.Context, serverID string, reader io.Reader) {
	// Ensure buffer exists for this server
	ms.ensureBuffer(serverID)

	scanner := bufio.NewScanner(reader)
	// Set a large max token size to handle long log lines
	const maxCapacity = 1024 * 1024 // 1MB
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			// Context cancelled, stop capturing
			return
		default:
			line := scanner.Text()
			if line == "" {
				continue
			}

			// Parse severity from line content
			severity := ms.parseSeverity(line)

			// Create log entry
			entry := models.NewLogEntry(severity, serverID, line)

			// Add to buffer
			ms.addLog(serverID, *entry)

			// Publish event for real-time UI updates
			if ms.eventBus != nil {
				event := events.ServerLogEntryEvent(serverID, entry)
				ms.eventBus.Publish(event)
			}
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil && err != io.EOF {
		// Log the error itself
		entry := models.NewLogEntry(models.LogError, serverID, "Error reading output: "+err.Error())
		ms.addLog(serverID, *entry)

		if ms.eventBus != nil {
			event := events.ServerLogEntryEvent(serverID, entry)
			ms.eventBus.Publish(event)
		}
	}
}

// parseSeverity determines log severity from line content
// Looks for keywords like "error", "warn", "warning", "success"
func (ms *MonitoringService) parseSeverity(line string) models.LogSeverity {
	lower := strings.ToLower(line)

	// Check for error keywords first (highest priority)
	if strings.Contains(lower, "error") || strings.Contains(lower, "fatal") ||
	   strings.Contains(lower, "panic") || strings.Contains(lower, "exception") {
		return models.LogError
	}

	// Check for warning keywords
	if strings.Contains(lower, "warn") || strings.Contains(lower, "warning") {
		return models.LogWarning
	}

	// Check for success keywords
	if strings.Contains(lower, "success") || strings.Contains(lower, "successful") ||
	   strings.Contains(lower, "completed") || strings.Contains(lower, "started") {
		return models.LogSuccess
	}

	// Default to info
	return models.LogInfo
}

// GetLogs retrieves logs for a specific server with pagination
func (ms *MonitoringService) GetLogs(serverID string, offset, limit int) []models.LogEntry {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	buffer, exists := ms.logBuffers[serverID]
	if !exists {
		return []models.LogEntry{}
	}

	return buffer.Get(offset, limit)
}

// GetAllLogs retrieves all logs for a specific server
func (ms *MonitoringService) GetAllLogs(serverID string) []models.LogEntry {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	buffer, exists := ms.logBuffers[serverID]
	if !exists {
		return []models.LogEntry{}
	}

	return buffer.GetAll()
}

// FilterLogs retrieves logs for a specific server filtered by severity
func (ms *MonitoringService) FilterLogs(serverID string, severity models.LogSeverity) []models.LogEntry {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	buffer, exists := ms.logBuffers[serverID]
	if !exists {
		return []models.LogEntry{}
	}

	return buffer.Filter(severity)
}

// ClearLogs clears all logs for a specific server
func (ms *MonitoringService) ClearLogs(serverID string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if buffer, exists := ms.logBuffers[serverID]; exists {
		buffer.Clear()
	}
}

// RemoveServer removes all logs and buffer for a server
func (ms *MonitoringService) RemoveServer(serverID string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	delete(ms.logBuffers, serverID)
}

// ensureBuffer ensures a log buffer exists for the given server
func (ms *MonitoringService) ensureBuffer(serverID string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.logBuffers[serverID]; !exists {
		ms.logBuffers[serverID] = models.NewCircularLogBuffer()
	}
}

// addLog adds a log entry to the server's buffer
func (ms *MonitoringService) addLog(serverID string, entry models.LogEntry) {
	// Ensure buffer exists first
	ms.ensureBuffer(serverID)

	ms.mu.RLock()
	buffer := ms.logBuffers[serverID]
	ms.mu.RUnlock()

	buffer.Add(entry)
}

// GetLogCount returns the number of logs for a server
func (ms *MonitoringService) GetLogCount(serverID string) int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	buffer, exists := ms.logBuffers[serverID]
	if !exists {
		return 0
	}

	return buffer.Size()
}
