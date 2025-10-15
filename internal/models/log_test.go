package models

import (
	"sync"
	"testing"
	"time"
)

func TestNewLogEntry(t *testing.T) {
	entry := NewLogEntry(LogInfo, "test-source", "test message")

	if entry.ID == "" {
		t.Error("ID should be generated")
	}
	if entry.Severity != LogInfo {
		t.Errorf("Expected severity %s, got %s", LogInfo, entry.Severity)
	}
	if entry.Source != "test-source" {
		t.Errorf("Expected source 'test-source', got %s", entry.Source)
	}
	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", entry.Message)
	}
	if entry.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

func TestCircularLogBuffer_Add(t *testing.T) {
	buffer := NewCircularLogBuffer()

	// Add 1000 entries (exactly fill the buffer)
	for i := 0; i < 1000; i++ {
		entry := NewLogEntry(LogInfo, "test", "message")
		buffer.Add(*entry)
	}

	if buffer.Size() != 1000 {
		t.Errorf("Expected size 1000, got %d", buffer.Size())
	}

	// Add one more entry (should overwrite oldest)
	entry1001 := NewLogEntry(LogInfo, "test", "message 1001")
	buffer.Add(*entry1001)

	if buffer.Size() != 1000 {
		t.Errorf("Expected size to remain 1000, got %d", buffer.Size())
	}

	// The newest entry should be retrievable
	entries := buffer.GetAll()
	if len(entries) != 1000 {
		t.Errorf("Expected 1000 entries, got %d", len(entries))
	}
}

func TestCircularLogBuffer_GetWithOffsetAndLimit(t *testing.T) {
	buffer := NewCircularLogBuffer()

	// Add 100 entries
	for i := 0; i < 100; i++ {
		entry := NewLogEntry(LogInfo, "test", "message")
		buffer.Add(*entry)
	}

	// Get first 10
	entries := buffer.Get(0, 10)
	if len(entries) != 10 {
		t.Errorf("Expected 10 entries, got %d", len(entries))
	}

	// Get entries 50-60
	entries = buffer.Get(50, 10)
	if len(entries) != 10 {
		t.Errorf("Expected 10 entries, got %d", len(entries))
	}

	// Get with offset beyond size
	entries = buffer.Get(1000, 10)
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries, got %d", len(entries))
	}

	// Get more than available
	entries = buffer.Get(90, 20)
	if len(entries) != 10 {
		t.Errorf("Expected 10 entries (90-99), got %d", len(entries))
	}
}

func TestCircularLogBuffer_Filter(t *testing.T) {
	buffer := NewCircularLogBuffer()

	// Add entries with different severities
	severities := []LogSeverity{LogInfo, LogSuccess, LogWarning, LogError}
	for i := 0; i < 100; i++ {
		severity := severities[i%4]
		entry := NewLogEntry(severity, "test", "message")
		buffer.Add(*entry)
	}

	// Filter by each severity
	for _, severity := range severities {
		filtered := buffer.Filter(severity)
		if len(filtered) != 25 {
			t.Errorf("Expected 25 entries for %s, got %d", severity, len(filtered))
		}

		// Verify all entries have correct severity
		for _, entry := range filtered {
			if entry.Severity != severity {
				t.Errorf("Expected severity %s, got %s", severity, entry.Severity)
			}
		}
	}
}

func TestCircularLogBuffer_ChronologicalOrder(t *testing.T) {
	buffer := NewCircularLogBuffer()

	// Add entries with delays to ensure timestamp ordering
	entries := make([]*LogEntry, 10)
	for i := 0; i < 10; i++ {
		entry := NewLogEntry(LogInfo, "test", "message")
		entries[i] = entry
		buffer.Add(*entry)
		time.Sleep(time.Millisecond)
	}

	// Get all entries
	retrieved := buffer.GetAll()

	// Verify chronological order
	for i := 1; i < len(retrieved); i++ {
		if retrieved[i].Timestamp.Before(retrieved[i-1].Timestamp) {
			t.Error("Entries should be in chronological order")
		}
	}
}

func TestCircularLogBuffer_OverwriteOldest(t *testing.T) {
	buffer := NewCircularLogBuffer()

	// Fill buffer completely
	firstEntry := NewLogEntry(LogInfo, "first", "first message")
	buffer.Add(*firstEntry)

	for i := 1; i < 1000; i++ {
		entry := NewLogEntry(LogInfo, "test", "message")
		buffer.Add(*entry)
	}

	// At this point, buffer is full and first entry is still there
	entries := buffer.GetAll()
	if entries[0].Source != "first" {
		t.Error("First entry should still be present")
	}

	// Add one more to overwrite the first
	newEntry := NewLogEntry(LogInfo, "new", "new message")
	buffer.Add(*newEntry)

	// First entry should be gone
	entries = buffer.GetAll()
	if entries[0].Source == "first" {
		t.Error("First entry should have been overwritten")
	}
	if entries[len(entries)-1].Source != "new" {
		t.Error("New entry should be at the end")
	}
}

func TestCircularLogBuffer_ConcurrentAdd(t *testing.T) {
	buffer := NewCircularLogBuffer()

	// Add entries concurrently from multiple goroutines
	var wg sync.WaitGroup
	numGoroutines := 10
	entriesPerGoroutine := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < entriesPerGoroutine; j++ {
				entry := NewLogEntry(LogInfo, "concurrent", "message")
				buffer.Add(*entry)
			}
		}(i)
	}

	wg.Wait()

	// Verify buffer size (should be at most 1000)
	size := buffer.Size()
	expectedSize := numGoroutines * entriesPerGoroutine
	if expectedSize > 1000 {
		expectedSize = 1000
	}

	if size != expectedSize {
		t.Errorf("Expected size %d, got %d", expectedSize, size)
	}
}

func TestCircularLogBuffer_ConcurrentAddAndGet(t *testing.T) {
	buffer := NewCircularLogBuffer()

	// Start adding entries in background
	stopChan := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopChan:
				return
			default:
				entry := NewLogEntry(LogInfo, "concurrent", "message")
				buffer.Add(*entry)
				time.Sleep(time.Millisecond)
			}
		}
	}()

	// Concurrently read entries
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = buffer.GetAll()
			_ = buffer.Filter(LogInfo)
			time.Sleep(time.Millisecond)
		}
	}()

	// Let it run for a bit
	time.Sleep(200 * time.Millisecond)
	close(stopChan)
	wg.Wait()

	// If we get here without deadlock, test passes
}

func TestCircularLogBuffer_Clear(t *testing.T) {
	buffer := NewCircularLogBuffer()

	// Add entries
	for i := 0; i < 100; i++ {
		entry := NewLogEntry(LogInfo, "test", "message")
		buffer.Add(*entry)
	}

	if buffer.Size() != 100 {
		t.Errorf("Expected size 100, got %d", buffer.Size())
	}

	// Clear buffer
	buffer.Clear()

	if buffer.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", buffer.Size())
	}

	entries := buffer.GetAll()
	if len(entries) != 0 {
		t.Errorf("Expected no entries after clear, got %d", len(entries))
	}
}

// Test with race detector: go test -race
func TestCircularLogBuffer_RaceConditions(t *testing.T) {
	buffer := NewCircularLogBuffer()

	var wg sync.WaitGroup
	numGoroutines := 5

	// Multiple writers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				entry := NewLogEntry(LogInfo, "test", "message")
				buffer.Add(*entry)
			}
		}()
	}

	// Multiple readers
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				_ = buffer.GetAll()
				_ = buffer.Size()
				_ = buffer.Filter(LogInfo)
			}
		}()
	}

	wg.Wait()
}
