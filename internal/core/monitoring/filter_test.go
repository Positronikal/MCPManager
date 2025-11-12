package monitoring

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/models"
)

// setupTestMonitoringService creates a monitoring service with test data
func setupTestMonitoringService() *MonitoringService {
	ms := NewMonitoringService(nil)

	// Add logs for server1
	ms.addLog("server1", *models.NewLogEntry(models.LogInfo, "server1", "Server started"))
	ms.addLog("server1", *models.NewLogEntry(models.LogError, "server1", "Database connection error"))
	ms.addLog("server1", *models.NewLogEntry(models.LogWarning, "server1", "High memory usage"))
	ms.addLog("server1", *models.NewLogEntry(models.LogInfo, "server1", "Processing request"))
	ms.addLog("server1", *models.NewLogEntry(models.LogSuccess, "server1", "Request completed successfully"))

	// Add logs for server2
	ms.addLog("server2", *models.NewLogEntry(models.LogInfo, "server2", "Server started"))
	ms.addLog("server2", *models.NewLogEntry(models.LogError, "server2", "Failed to load configuration"))
	ms.addLog("server2", *models.NewLogEntry(models.LogWarning, "server2", "Deprecated API used"))
	ms.addLog("server2", *models.NewLogEntry(models.LogSuccess, "server2", "Configuration reloaded"))

	return ms
}

func TestNewLogFilter(t *testing.T) {
	filter := NewLogFilter()

	if filter.Limit != DefaultLogLimit {
		t.Errorf("Expected default limit %d, got %d", DefaultLogLimit, filter.Limit)
	}

	if filter.ServerID != nil {
		t.Error("Expected ServerID to be nil")
	}

	if filter.Severity != nil {
		t.Error("Expected Severity to be nil")
	}

	if filter.Search != "" {
		t.Error("Expected Search to be empty")
	}
}

func TestLogFilter_FluentAPI(t *testing.T) {
	serverID := "test-server"
	severity := models.LogError

	filter := NewLogFilter().
		WithServerID(serverID).
		WithSeverity(severity).
		WithSearch("error").
		WithLimit(50)

	if filter.ServerID == nil || *filter.ServerID != serverID {
		t.Errorf("Expected ServerID %s, got %v", serverID, filter.ServerID)
	}

	if filter.Severity == nil || *filter.Severity != severity {
		t.Errorf("Expected Severity %s, got %v", severity, filter.Severity)
	}

	if filter.Search != "error" {
		t.Errorf("Expected Search 'error', got '%s'", filter.Search)
	}

	if filter.Limit != 50 {
		t.Errorf("Expected Limit 50, got %d", filter.Limit)
	}
}

func TestLogFilter_Validate(t *testing.T) {
	tests := []struct {
		name          string
		limit         int
		expectedLimit int
	}{
		{"Zero limit uses default", 0, DefaultLogLimit},
		{"Negative limit uses default", -1, DefaultLogLimit},
		{"Valid limit preserved", 500, 500},
		{"Exceeds max limit clamped", 2000, MaxLogLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := &LogFilter{Limit: tt.limit}
			filter.Validate()

			if filter.Limit != tt.expectedLimit {
				t.Errorf("Expected limit %d, got %d", tt.expectedLimit, filter.Limit)
			}
		})
	}
}

func TestLogFilter_ValidateInvalidSeverity(t *testing.T) {
	invalidSeverity := models.LogSeverity("invalid")
	filter := &LogFilter{
		Severity: &invalidSeverity,
		Limit:    100,
	}

	filter.Validate()

	if filter.Severity != nil {
		t.Error("Expected invalid severity to be set to nil")
	}
}

func TestFilterLogs_ByServerID(t *testing.T) {
	ms := setupTestMonitoringService()

	// Filter by server1
	filter := NewLogFilter().WithServerID("server1")
	logs := ms.FilterLogs(filter)

	if len(logs) != 5 {
		t.Errorf("Expected 5 logs for server1, got %d", len(logs))
	}

	// Verify all logs are from server1
	for _, log := range logs {
		if log.Source != "server1" {
			t.Errorf("Expected all logs from server1, got log from %s", log.Source)
		}
	}
}

func TestFilterLogs_BySeverity(t *testing.T) {
	ms := setupTestMonitoringService()

	// Filter by error severity across all servers
	filter := NewLogFilter().WithSeverity(models.LogError)
	logs := ms.FilterLogs(filter)

	if len(logs) != 2 {
		t.Errorf("Expected 2 error logs, got %d", len(logs))
	}

	// Verify all logs are errors
	for _, log := range logs {
		if log.Severity != models.LogError {
			t.Errorf("Expected all logs to be errors, got %s", log.Severity)
		}
	}
}

func TestFilterLogs_BySearch(t *testing.T) {
	ms := setupTestMonitoringService()

	// Search for "error" (case-insensitive)
	filter := NewLogFilter().WithSearch("error")
	logs := ms.FilterLogs(filter)

	// Should match only "Database connection error"
	// Note: We're searching message content, not severity
	if len(logs) != 1 {
		t.Errorf("Expected 1 log containing 'error', got %d", len(logs))
	}

	// Verify all logs contain "error" (case-insensitive)
	for _, log := range logs {
		if !containsCaseInsensitive(log.Message, "error") {
			t.Errorf("Expected log message to contain 'error', got '%s'", log.Message)
		}
	}
}

func TestFilterLogs_CaseInsensitiveSearch(t *testing.T) {
	ms := setupTestMonitoringService()

	// Search with different cases
	searches := []string{"ERROR", "Error", "error", "eRrOr"}

	for _, search := range searches {
		filter := NewLogFilter().WithSearch(search)
		logs := ms.FilterLogs(filter)

		if len(logs) != 1 {
			t.Errorf("Search '%s': Expected 1 log, got %d", search, len(logs))
		}
	}
}

func TestFilterLogs_CombinedFilters(t *testing.T) {
	ms := setupTestMonitoringService()

	// Filter server1 + error severity
	filter := NewLogFilter().
		WithServerID("server1").
		WithSeverity(models.LogError)

	logs := ms.FilterLogs(filter)

	if len(logs) != 1 {
		t.Errorf("Expected 1 log (server1 errors), got %d", len(logs))
	}

	if len(logs) > 0 {
		log := logs[0]
		if log.Source != "server1" || log.Severity != models.LogError {
			t.Errorf("Expected server1 error log, got source=%s severity=%s", log.Source, log.Severity)
		}
	}
}

func TestFilterLogs_CombinedFiltersWithSearch(t *testing.T) {
	ms := setupTestMonitoringService()

	// Filter server1 + search for "memory"
	filter := NewLogFilter().
		WithServerID("server1").
		WithSearch("memory")

	logs := ms.FilterLogs(filter)

	if len(logs) != 1 {
		t.Errorf("Expected 1 log matching criteria, got %d", len(logs))
	}

	if len(logs) > 0 {
		log := logs[0]
		if log.Source != "server1" || !containsCaseInsensitive(log.Message, "memory") {
			t.Errorf("Expected log from server1 containing 'memory', got source=%s message=%s", log.Source, log.Message)
		}
	}
}

func TestFilterLogs_AllServers(t *testing.T) {
	ms := setupTestMonitoringService()

	// No server filter = all servers
	filter := NewLogFilter()
	logs := ms.FilterLogs(filter)

	// Total logs from both servers (limited by default limit)
	if len(logs) != 9 {
		t.Errorf("Expected 9 total logs, got %d", len(logs))
	}
}

func TestFilterLogs_Limit(t *testing.T) {
	ms := setupTestMonitoringService()

	// Apply limit of 3
	filter := NewLogFilter().WithLimit(3)
	logs := ms.FilterLogs(filter)

	if len(logs) != 3 {
		t.Errorf("Expected 3 logs (limited), got %d", len(logs))
	}
}

func TestFilterLogs_NoMatches(t *testing.T) {
	ms := setupTestMonitoringService()

	// Search for non-existent string
	filter := NewLogFilter().WithSearch("nonexistent")
	logs := ms.FilterLogs(filter)

	if len(logs) != 0 {
		t.Errorf("Expected 0 logs for non-matching search, got %d", len(logs))
	}
}

func TestFilterLogs_NonExistentServer(t *testing.T) {
	ms := setupTestMonitoringService()

	filter := NewLogFilter().WithServerID("nonexistent")
	logs := ms.FilterLogs(filter)

	if len(logs) != 0 {
		t.Errorf("Expected 0 logs for non-existent server, got %d", len(logs))
	}
}

func TestFilterLogs_NilFilter(t *testing.T) {
	ms := setupTestMonitoringService()

	// Passing nil should use default filter
	logs := ms.FilterLogs(nil)

	if len(logs) != 9 {
		t.Errorf("Expected 9 logs with nil filter, got %d", len(logs))
	}
}

func TestFilterLogsBySeverity_Deprecated(t *testing.T) {
	ms := setupTestMonitoringService()

	logs := ms.FilterLogsBySeverity("server1", models.LogError)

	if len(logs) != 1 {
		t.Errorf("Expected 1 error log from server1, got %d", len(logs))
	}
}

func TestSearchLogs_Convenience(t *testing.T) {
	ms := setupTestMonitoringService()

	logs := ms.SearchLogs("configuration", 10)

	if len(logs) != 2 {
		t.Errorf("Expected 2 logs containing 'configuration', got %d", len(logs))
	}
}

func TestGetServerLogs_Convenience(t *testing.T) {
	ms := setupTestMonitoringService()

	logs := ms.GetServerLogs("server2", 100)

	if len(logs) != 4 {
		t.Errorf("Expected 4 logs from server2, got %d", len(logs))
	}

	for _, log := range logs {
		if log.Source != "server2" {
			t.Errorf("Expected all logs from server2, got log from %s", log.Source)
		}
	}
}

func TestGetRecentErrors(t *testing.T) {
	ms := setupTestMonitoringService()

	logs := ms.GetRecentErrors(10)

	if len(logs) != 2 {
		t.Errorf("Expected 2 error logs, got %d", len(logs))
	}

	for _, log := range logs {
		if log.Severity != models.LogError {
			t.Errorf("Expected all errors, got %s", log.Severity)
		}
	}
}

func TestGetRecentWarnings(t *testing.T) {
	ms := setupTestMonitoringService()

	logs := ms.GetRecentWarnings(10)

	if len(logs) != 2 {
		t.Errorf("Expected 2 warning logs, got %d", len(logs))
	}

	for _, log := range logs {
		if log.Severity != models.LogWarning {
			t.Errorf("Expected all warnings, got %s", log.Severity)
		}
	}
}

func TestCountLogsBySeverity_SingleServer(t *testing.T) {
	ms := setupTestMonitoringService()

	serverID := "server1"
	counts := ms.CountLogsBySeverity(&serverID)

	expected := map[models.LogSeverity]int{
		models.LogInfo:    2,
		models.LogError:   1,
		models.LogWarning: 1,
		models.LogSuccess: 1,
	}

	for severity, expectedCount := range expected {
		if counts[severity] != expectedCount {
			t.Errorf("Server1 %s: expected %d, got %d", severity, expectedCount, counts[severity])
		}
	}
}

func TestCountLogsBySeverity_AllServers(t *testing.T) {
	ms := setupTestMonitoringService()

	counts := ms.CountLogsBySeverity(nil)

	expected := map[models.LogSeverity]int{
		models.LogInfo:    3,
		models.LogError:   2,
		models.LogWarning: 2,
		models.LogSuccess: 2,
	}

	for severity, expectedCount := range expected {
		if counts[severity] != expectedCount {
			t.Errorf("All servers %s: expected %d, got %d", severity, expectedCount, counts[severity])
		}
	}
}

func TestFilterLogs_EmptySearch(t *testing.T) {
	ms := setupTestMonitoringService()

	// Empty search should not filter
	filter := NewLogFilter().WithSearch("")
	logs := ms.FilterLogs(filter)

	if len(logs) != 9 {
		t.Errorf("Expected 9 logs with empty search, got %d", len(logs))
	}
}

func TestFilterLogs_PartialMatch(t *testing.T) {
	ms := setupTestMonitoringService()

	// Search for partial word "config"
	filter := NewLogFilter().WithSearch("config")
	logs := ms.FilterLogs(filter)

	// Should match "configuration"
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs with partial match, got %d", len(logs))
	}
}

// Performance test: Filter 50k entries in under 50ms
func TestFilterLogs_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	ms := NewMonitoringService(nil)

	// Add 50k log entries across 5 servers
	severities := []models.LogSeverity{models.LogInfo, models.LogError, models.LogWarning, models.LogSuccess}
	messages := []string{
		"Processing request",
		"Database query error",
		"High memory warning",
		"Operation completed successfully",
		"Connection timeout",
	}

	for i := 0; i < 50000; i++ {
		serverID := fmt.Sprintf("server%d", i%5)
		severity := severities[i%len(severities)]
		message := messages[i%len(messages)]

		ms.addLog(serverID, *models.NewLogEntry(severity, serverID, message))
	}

	// Test filtering performance
	start := time.Now()

	filter := NewLogFilter().
		WithSeverity(models.LogError).
		WithSearch("database")

	logs := ms.FilterLogs(filter)

	elapsed := time.Since(start)

	t.Logf("Filtered 50k entries in %v", elapsed)
	t.Logf("Found %d matching logs", len(logs))

	// Should complete in under 50ms
	if elapsed > 50*time.Millisecond {
		t.Errorf("Performance requirement not met: took %v (expected < 50ms)", elapsed)
	}

	// Verify results are correct
	for _, log := range logs {
		if log.Severity != models.LogError {
			t.Error("Expected all logs to be errors")
		}
		if !containsCaseInsensitive(log.Message, "database") {
			t.Error("Expected all logs to contain 'database'")
		}
	}
}

// Benchmark for filter performance
func BenchmarkFilterLogs_NoFilters(b *testing.B) {
	ms := setupTestMonitoringService()
	filter := NewLogFilter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.FilterLogs(filter)
	}
}

func BenchmarkFilterLogs_WithSeverity(b *testing.B) {
	ms := setupTestMonitoringService()
	filter := NewLogFilter().WithSeverity(models.LogError)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.FilterLogs(filter)
	}
}

func BenchmarkFilterLogs_WithSearch(b *testing.B) {
	ms := setupTestMonitoringService()
	filter := NewLogFilter().WithSearch("error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.FilterLogs(filter)
	}
}

func BenchmarkFilterLogs_Combined(b *testing.B) {
	ms := setupTestMonitoringService()
	filter := NewLogFilter().
		WithServerID("server1").
		WithSeverity(models.LogError).
		WithSearch("database")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.FilterLogs(filter)
	}
}

// Helper function for case-insensitive substring matching
func containsCaseInsensitive(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
