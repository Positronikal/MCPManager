package monitoring

import (
	"strings"

	"github.com/hoytech/mcpmanager/internal/models"
)

const (
	// DefaultLogLimit is the default maximum number of logs to return
	DefaultLogLimit = 100
	// MaxLogLimit is the absolute maximum number of logs that can be returned
	MaxLogLimit = 1000
)

// LogFilter represents filtering criteria for logs
type LogFilter struct {
	ServerID *string               // Filter by specific server ID (nil = all servers)
	Severity *models.LogSeverity   // Filter by severity level (nil = all severities)
	Search   string                // Full-text search in message (case-insensitive)
	Limit    int                   // Maximum number of results (0 = use default)
}

// NewLogFilter creates a new log filter with default values
func NewLogFilter() *LogFilter {
	return &LogFilter{
		Limit: DefaultLogLimit,
	}
}

// WithServerID sets the server ID filter
func (f *LogFilter) WithServerID(serverID string) *LogFilter {
	f.ServerID = &serverID
	return f
}

// WithSeverity sets the severity filter
func (f *LogFilter) WithSeverity(severity models.LogSeverity) *LogFilter {
	f.Severity = &severity
	return f
}

// WithSearch sets the search query
func (f *LogFilter) WithSearch(search string) *LogFilter {
	f.Search = search
	return f
}

// WithLimit sets the result limit
func (f *LogFilter) WithLimit(limit int) *LogFilter {
	f.Limit = limit
	return f
}

// Validate ensures the filter has valid values
func (f *LogFilter) Validate() {
	// Ensure limit is within bounds
	if f.Limit <= 0 {
		f.Limit = DefaultLogLimit
	}
	if f.Limit > MaxLogLimit {
		f.Limit = MaxLogLimit
	}

	// Validate severity if provided
	if f.Severity != nil && !f.Severity.IsValid() {
		f.Severity = nil
	}
}

// FilterLogs retrieves and filters logs based on multiple criteria
// Supports filtering by server ID, severity, and full-text search
// Returns up to 'limit' entries (default 100, max 1000)
func (ms *MonitoringService) FilterLogs(filter *LogFilter) []models.LogEntry {
	if filter == nil {
		filter = NewLogFilter()
	}

	// Validate filter parameters
	filter.Validate()

	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Collect entries from relevant buffers
	var allEntries []models.LogEntry

	if filter.ServerID != nil {
		// Filter by specific server
		buffer, exists := ms.logBuffers[*filter.ServerID]
		if exists {
			allEntries = buffer.GetAll()
		}
	} else {
		// Aggregate logs from all servers
		for _, buffer := range ms.logBuffers {
			entries := buffer.GetAll()
			allEntries = append(allEntries, entries...)
		}
	}

	// Apply filters
	filtered := ms.applyFilters(allEntries, filter)

	// Apply limit
	if len(filtered) > filter.Limit {
		filtered = filtered[:filter.Limit]
	}

	return filtered
}

// applyFilters applies severity and search filters to log entries
func (ms *MonitoringService) applyFilters(entries []models.LogEntry, filter *LogFilter) []models.LogEntry {
	if filter.Severity == nil && filter.Search == "" {
		// No filters to apply
		return entries
	}

	result := make([]models.LogEntry, 0, len(entries))
	searchLower := strings.ToLower(filter.Search)

	for _, entry := range entries {
		// Check severity filter
		if filter.Severity != nil && entry.Severity != *filter.Severity {
			continue
		}

		// Check search filter (case-insensitive substring match)
		if filter.Search != "" {
			messageLower := strings.ToLower(entry.Message)
			if !strings.Contains(messageLower, searchLower) {
				continue
			}
		}

		result = append(result, entry)
	}

	return result
}

// FilterLogsBySeverity is a convenience method for filtering by severity only
// Deprecated: Use FilterLogs with LogFilter instead
func (ms *MonitoringService) FilterLogsBySeverity(serverID string, severity models.LogSeverity) []models.LogEntry {
	filter := NewLogFilter().WithServerID(serverID).WithSeverity(severity)
	return ms.FilterLogs(filter)
}

// SearchLogs is a convenience method for full-text search across all servers
func (ms *MonitoringService) SearchLogs(query string, limit int) []models.LogEntry {
	filter := NewLogFilter().WithSearch(query).WithLimit(limit)
	return ms.FilterLogs(filter)
}

// GetServerLogs is a convenience method for getting all logs from a specific server
func (ms *MonitoringService) GetServerLogs(serverID string, limit int) []models.LogEntry {
	filter := NewLogFilter().WithServerID(serverID).WithLimit(limit)
	return ms.FilterLogs(filter)
}

// GetRecentErrors retrieves recent error logs across all servers
func (ms *MonitoringService) GetRecentErrors(limit int) []models.LogEntry {
	filter := NewLogFilter().WithSeverity(models.LogError).WithLimit(limit)
	return ms.FilterLogs(filter)
}

// GetRecentWarnings retrieves recent warning logs across all servers
func (ms *MonitoringService) GetRecentWarnings(limit int) []models.LogEntry {
	filter := NewLogFilter().WithSeverity(models.LogWarning).WithLimit(limit)
	return ms.FilterLogs(filter)
}

// CountLogsBySeverity counts logs by severity for a specific server or all servers
func (ms *MonitoringService) CountLogsBySeverity(serverID *string) map[models.LogSeverity]int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	counts := make(map[models.LogSeverity]int)

	var buffers []*models.CircularLogBuffer
	if serverID != nil {
		if buffer, exists := ms.logBuffers[*serverID]; exists {
			buffers = append(buffers, buffer)
		}
	} else {
		for _, buffer := range ms.logBuffers {
			buffers = append(buffers, buffer)
		}
	}

	for _, buffer := range buffers {
		entries := buffer.GetAll()
		for _, entry := range entries {
			counts[entry.Severity]++
		}
	}

	return counts
}
