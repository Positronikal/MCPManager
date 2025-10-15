package models

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// LogEntry represents a single log entry
type LogEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Severity  LogSeverity            `json:"severity"`
	Source    string                 `json:"source"` // Server ID or component name
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewLogEntry creates a new log entry
func NewLogEntry(severity LogSeverity, source, message string) *LogEntry {
	return &LogEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Severity:  severity,
		Source:    source,
		Message:   message,
		Metadata:  make(map[string]interface{}),
	}
}

// CircularLogBuffer is a thread-safe circular buffer for storing log entries
type CircularLogBuffer struct {
	entries [1000]LogEntry
	head    int
	size    int
	mu      sync.RWMutex
}

// NewCircularLogBuffer creates a new circular log buffer
func NewCircularLogBuffer() *CircularLogBuffer {
	return &CircularLogBuffer{
		head: 0,
		size: 0,
	}
}

// Add adds a log entry to the buffer
func (b *CircularLogBuffer) Add(entry LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.entries[b.head] = entry
	b.head = (b.head + 1) % 1000

	if b.size < 1000 {
		b.size++
	}
}

// Get retrieves log entries with offset and limit
// Returns entries in chronological order (oldest first)
func (b *CircularLogBuffer) Get(offset, limit int) []LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if offset >= b.size {
		return []LogEntry{}
	}

	// Calculate the start index (oldest entry)
	startIdx := 0
	if b.size == 1000 {
		// Buffer is full, oldest entry is at head
		startIdx = b.head
	}
	// Otherwise, oldest entry is at index 0

	// Collect entries in chronological order
	result := make([]LogEntry, 0, limit)
	for i := 0; i < b.size; i++ {
		if i < offset {
			continue
		}
		if len(result) >= limit {
			break
		}

		idx := (startIdx + i) % 1000
		result = append(result, b.entries[idx])
	}

	return result
}

// GetAll returns all entries in chronological order
func (b *CircularLogBuffer) GetAll() []LogEntry {
	b.mu.RLock()
	size := b.size
	b.mu.RUnlock()
	return b.Get(0, size)
}

// Filter returns entries matching the specified severity
func (b *CircularLogBuffer) Filter(severity LogSeverity) []LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// Calculate the start index (oldest entry)
	startIdx := 0
	if b.size == 1000 {
		startIdx = b.head
	}

	result := make([]LogEntry, 0)
	for i := 0; i < b.size; i++ {
		idx := (startIdx + i) % 1000
		entry := b.entries[idx]
		if entry.Severity == severity {
			result = append(result, entry)
		}
	}

	return result
}

// Size returns the current number of entries in the buffer
func (b *CircularLogBuffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

// Clear removes all entries from the buffer
func (b *CircularLogBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.head = 0
	b.size = 0
}
