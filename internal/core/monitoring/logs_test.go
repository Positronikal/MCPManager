package monitoring

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/Positronikal/MCPManager/internal/models"
)

func TestNewMonitoringService(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	ms := NewMonitoringService(eventBus)

	if ms == nil {
		t.Fatal("Expected monitoring service to be created")
	}

	if ms.eventBus != eventBus {
		t.Error("Expected event bus to be set")
	}

	if ms.logBuffers == nil {
		t.Error("Expected log buffers map to be initialized")
	}
}

func TestCaptureOutput_BasicCapture(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	ms := NewMonitoringService(eventBus)
	serverID := "test-server"

	// Create a pipe to simulate server output
	reader, writer := io.Pipe()

	// Start capturing in a goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan bool)
	go func() {
		ms.CaptureOutput(ctx, serverID, reader)
		done <- true
	}()

	// Write some log lines
	writer.Write([]byte("Starting server...\n"))
	writer.Write([]byte("Server is running\n"))
	writer.Write([]byte("Connection established\n"))

	// Give it time to process
	time.Sleep(50 * time.Millisecond)

	// Close writer to signal EOF
	writer.Close()

	// Wait for capture to complete
	select {
	case <-done:
		// Capture completed
	case <-time.After(1 * time.Second):
		t.Fatal("Capture did not complete in time")
	}

	// Verify logs were captured
	logs := ms.GetAllLogs(serverID)
	if len(logs) != 3 {
		t.Errorf("Expected 3 log entries, got %d", len(logs))
	}

	if logs[0].Message != "Starting server..." {
		t.Errorf("Expected first message to be 'Starting server...', got '%s'", logs[0].Message)
	}
}

func TestCaptureOutput_EmptyLines(t *testing.T) {
	ms := NewMonitoringService(nil)
	serverID := "test-server"

	// Create reader with empty lines
	input := "Line 1\n\n\nLine 2\n"
	reader := strings.NewReader(input)

	ctx := context.Background()
	ms.CaptureOutput(ctx, serverID, reader)

	// Verify only non-empty lines were captured
	logs := ms.GetAllLogs(serverID)
	if len(logs) != 2 {
		t.Errorf("Expected 2 log entries (empty lines should be skipped), got %d", len(logs))
	}
}

func TestCaptureOutput_ContextCancellation(t *testing.T) {
	ms := NewMonitoringService(nil)
	serverID := "test-server"

	// Create a pipe
	reader, writer := io.Pipe()

	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan bool)
	go func() {
		ms.CaptureOutput(ctx, serverID, reader)
		done <- true
	}()

	// Write one line
	writer.Write([]byte("First line\n"))
	time.Sleep(50 * time.Millisecond)

	// Cancel context and close writer to unblock scanner
	cancel()
	writer.Close()

	// Wait for capture to stop
	select {
	case <-done:
		// Capture completed
	case <-time.After(1 * time.Second):
		t.Fatal("Capture did not stop after context cancellation")
	}

	// Verify we got the first line
	logs := ms.GetAllLogs(serverID)
	if len(logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(logs))
	}
}

func TestParseSeverity(t *testing.T) {
	ms := NewMonitoringService(nil)

	tests := []struct {
		line     string
		expected models.LogSeverity
	}{
		{"This is an error message", models.LogError},
		{"ERROR: Something went wrong", models.LogError},
		{"Fatal error occurred", models.LogError},
		{"Panic: system failure", models.LogError},
		{"Exception thrown", models.LogError},
		{"Warning: low disk space", models.LogWarning},
		{"WARN: deprecated API", models.LogWarning},
		{"Success: operation completed", models.LogSuccess},
		{"Successful connection", models.LogSuccess},
		{"Server started successfully", models.LogSuccess},
		{"Task completed", models.LogSuccess},
		{"Normal info message", models.LogInfo},
		{"Processing request...", models.LogInfo},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			severity := ms.parseSeverity(tt.line)
			if severity != tt.expected {
				t.Errorf("Line '%s': expected severity %s, got %s", tt.line, tt.expected, severity)
			}
		})
	}
}

func TestParseSeverity_CaseInsensitive(t *testing.T) {
	ms := NewMonitoringService(nil)

	tests := []struct {
		line     string
		expected models.LogSeverity
	}{
		{"ERROR", models.LogError},
		{"error", models.LogError},
		{"ErRoR", models.LogError},
		{"WARNING", models.LogWarning},
		{"warning", models.LogWarning},
		{"WaRnInG", models.LogWarning},
	}

	for _, tt := range tests {
		severity := ms.parseSeverity(tt.line)
		if severity != tt.expected {
			t.Errorf("Case insensitive test failed for '%s': expected %s, got %s", tt.line, tt.expected, severity)
		}
	}
}

func TestGetLogs_Pagination(t *testing.T) {
	ms := NewMonitoringService(nil)
	serverID := "test-server"

	// Add 10 log entries
	for i := 0; i < 10; i++ {
		entry := models.NewLogEntry(models.LogInfo, serverID, "Message "+string(rune('0'+i)))
		ms.addLog(serverID, *entry)
	}

	// Test pagination
	logs := ms.GetLogs(serverID, 0, 5)
	if len(logs) != 5 {
		t.Errorf("Expected 5 logs with offset 0 and limit 5, got %d", len(logs))
	}

	logs = ms.GetLogs(serverID, 5, 5)
	if len(logs) != 5 {
		t.Errorf("Expected 5 logs with offset 5 and limit 5, got %d", len(logs))
	}

	logs = ms.GetLogs(serverID, 8, 5)
	if len(logs) != 2 {
		t.Errorf("Expected 2 logs with offset 8 and limit 5, got %d", len(logs))
	}
}

func TestGetLogs_NonExistentServer(t *testing.T) {
	ms := NewMonitoringService(nil)

	logs := ms.GetLogs("nonexistent", 0, 10)
	if len(logs) != 0 {
		t.Errorf("Expected empty array for nonexistent server, got %d logs", len(logs))
	}
}

func TestFilterLogsBySeverityOld(t *testing.T) {
	ms := NewMonitoringService(nil)
	serverID := "test-server"

	// Add logs with different severities
	ms.addLog(serverID, *models.NewLogEntry(models.LogInfo, serverID, "Info 1"))
	ms.addLog(serverID, *models.NewLogEntry(models.LogError, serverID, "Error 1"))
	ms.addLog(serverID, *models.NewLogEntry(models.LogInfo, serverID, "Info 2"))
	ms.addLog(serverID, *models.NewLogEntry(models.LogWarning, serverID, "Warning 1"))
	ms.addLog(serverID, *models.NewLogEntry(models.LogError, serverID, "Error 2"))

	// Filter by error
	errorLogs := ms.FilterLogsBySeverityOld(serverID, models.LogError)
	if len(errorLogs) != 2 {
		t.Errorf("Expected 2 error logs, got %d", len(errorLogs))
	}

	// Filter by info
	infoLogs := ms.FilterLogsBySeverityOld(serverID, models.LogInfo)
	if len(infoLogs) != 2 {
		t.Errorf("Expected 2 info logs, got %d", len(infoLogs))
	}

	// Filter by warning
	warningLogs := ms.FilterLogsBySeverityOld(serverID, models.LogWarning)
	if len(warningLogs) != 1 {
		t.Errorf("Expected 1 warning log, got %d", len(warningLogs))
	}
}

func TestClearLogs(t *testing.T) {
	ms := NewMonitoringService(nil)
	serverID := "test-server"

	// Add logs
	ms.addLog(serverID, *models.NewLogEntry(models.LogInfo, serverID, "Message 1"))
	ms.addLog(serverID, *models.NewLogEntry(models.LogInfo, serverID, "Message 2"))

	// Verify logs exist
	if ms.GetLogCount(serverID) != 2 {
		t.Error("Expected 2 logs before clear")
	}

	// Clear logs
	ms.ClearLogs(serverID)

	// Verify logs cleared
	if ms.GetLogCount(serverID) != 0 {
		t.Error("Expected 0 logs after clear")
	}
}

func TestRemoveServer(t *testing.T) {
	ms := NewMonitoringService(nil)
	serverID := "test-server"

	// Add logs
	ms.addLog(serverID, *models.NewLogEntry(models.LogInfo, serverID, "Message"))

	// Verify buffer exists
	if ms.GetLogCount(serverID) != 1 {
		t.Error("Expected buffer to exist")
	}

	// Remove server
	ms.RemoveServer(serverID)

	// Verify buffer removed
	logs := ms.GetAllLogs(serverID)
	if len(logs) != 0 {
		t.Error("Expected empty logs after server removal")
	}
}

func TestGetLogCount(t *testing.T) {
	ms := NewMonitoringService(nil)
	serverID := "test-server"

	// Initially should be 0
	if ms.GetLogCount(serverID) != 0 {
		t.Error("Expected 0 logs initially")
	}

	// Add logs
	for i := 0; i < 5; i++ {
		ms.addLog(serverID, *models.NewLogEntry(models.LogInfo, serverID, "Message"))
	}

	// Should be 5
	if ms.GetLogCount(serverID) != 5 {
		t.Errorf("Expected 5 logs, got %d", ms.GetLogCount(serverID))
	}
}

func TestCaptureOutput_EventPublishing(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	ms := NewMonitoringService(eventBus)
	serverID := "test-server"

	// Subscribe to log events
	logChan := eventBus.Subscribe(events.EventServerLogEntry)

	// Create reader with test data
	input := "Test log message\n"
	reader := strings.NewReader(input)

	// Capture in goroutine
	done := make(chan bool)
	go func() {
		ms.CaptureOutput(context.Background(), serverID, reader)
		done <- true
	}()

	// Wait for event
	select {
	case event := <-logChan:
		if event.Type != events.EventServerLogEntry {
			t.Errorf("Expected EventServerLogEntry, got %s", event.Type)
		}
		if event.Data["serverID"] != serverID {
			t.Errorf("Expected serverID %s, got %v", serverID, event.Data["serverID"])
		}
		if event.Data["message"] != "Test log message" {
			t.Errorf("Expected message 'Test log message', got %v", event.Data["message"])
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Expected log event but none received")
	}

	// Wait for capture to complete
	<-done
}

func TestCaptureOutput_LongLines(t *testing.T) {
	ms := NewMonitoringService(nil)
	serverID := "test-server"

	// Create a very long line
	longLine := strings.Repeat("A", 100000) + "\n"
	reader := strings.NewReader(longLine)

	ctx := context.Background()
	ms.CaptureOutput(ctx, serverID, reader)

	logs := ms.GetAllLogs(serverID)
	if len(logs) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(logs))
	}

	if len(logs[0].Message) != 100000 {
		t.Errorf("Expected message length 100000, got %d", len(logs[0].Message))
	}
}

func TestCaptureOutput_CircularBufferLimit(t *testing.T) {
	ms := NewMonitoringService(nil)
	serverID := "test-server"

	// Create reader with more than 1000 lines
	var buf bytes.Buffer
	for i := 0; i < 1500; i++ {
		buf.WriteString("Line " + string(rune('0'+i%10)) + "\n")
	}

	reader := bytes.NewReader(buf.Bytes())

	ctx := context.Background()
	ms.CaptureOutput(ctx, serverID, reader)

	// Verify buffer respects 1000 entry limit
	count := ms.GetLogCount(serverID)
	if count != 1000 {
		t.Errorf("Expected log count to be 1000 (circular buffer limit), got %d", count)
	}
}

func TestCaptureOutput_MultipleServers(t *testing.T) {
	ms := NewMonitoringService(nil)

	server1 := "server-1"
	server2 := "server-2"

	// Capture for server 1
	reader1 := strings.NewReader("Server 1 log\n")
	ms.CaptureOutput(context.Background(), server1, reader1)

	// Capture for server 2
	reader2 := strings.NewReader("Server 2 log\n")
	ms.CaptureOutput(context.Background(), server2, reader2)

	// Verify logs are separate
	logs1 := ms.GetAllLogs(server1)
	logs2 := ms.GetAllLogs(server2)

	if len(logs1) != 1 || len(logs2) != 1 {
		t.Error("Expected 1 log for each server")
	}

	if logs1[0].Message != "Server 1 log" {
		t.Errorf("Expected 'Server 1 log', got '%s'", logs1[0].Message)
	}

	if logs2[0].Message != "Server 2 log" {
		t.Errorf("Expected 'Server 2 log', got '%s'", logs2[0].Message)
	}
}

func TestCaptureOutput_SeverityPriority(t *testing.T) {
	ms := NewMonitoringService(nil)
	serverID := "test-server"

	// Create reader with mixed severity indicators
	// Error should take precedence over success/warning
	input := "Success: but also error occurred\n"
	reader := strings.NewReader(input)

	ms.CaptureOutput(context.Background(), serverID, reader)

	logs := ms.GetAllLogs(serverID)
	if len(logs) != 1 {
		t.Fatalf("Expected 1 log entry, got %d", len(logs))
	}

	// Error should take precedence
	if logs[0].Severity != models.LogError {
		t.Errorf("Expected severity Error (highest priority), got %s", logs[0].Severity)
	}
}
