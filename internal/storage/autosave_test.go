package storage

import (
	"sync"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/models"
)

// MockStorage is a mock storage service for testing
type MockStorage struct {
	saveCount int
	mu        sync.Mutex
	lastState *models.ApplicationState
}

func (m *MockStorage) SaveState(state *models.ApplicationState) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.saveCount++
	m.lastState = state
	return nil
}

func (m *MockStorage) LoadState() (*models.ApplicationState, error) {
	return models.NewApplicationState(), nil
}

func (m *MockStorage) LoadServerLogs(serverID string) ([]models.LogEntry, error) {
	return []models.LogEntry{}, nil
}

func (m *MockStorage) SaveServerLogs(serverID string, logs []models.LogEntry) error {
	return nil
}

func (m *MockStorage) GetSaveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveCount
}

func TestNewAutoSaver(t *testing.T) {
	storage := &MockStorage{}
	state := models.NewApplicationState()
	autoSaver := NewAutoSaver(storage, state)

	if autoSaver == nil {
		t.Fatal("Expected auto-saver to be created")
	}
	if autoSaver.dirty {
		t.Error("Auto-saver should not be dirty initially")
	}
}

func TestAutoSaver_MarkDirty(t *testing.T) {
	storage := &MockStorage{}
	state := models.NewApplicationState()
	autoSaver := NewAutoSaver(storage, state)

	if autoSaver.IsDirty() {
		t.Error("Should not be dirty initially")
	}

	autoSaver.MarkDirty()

	if !autoSaver.IsDirty() {
		t.Error("Should be dirty after MarkDirty()")
	}
}

func TestAutoSaver_SingleSaveAfterMultipleMarkDirty(t *testing.T) {
	storage := &MockStorage{}
	state := models.NewApplicationState()
	autoSaver := NewAutoSaver(storage, state)

	autoSaver.Start()
	defer autoSaver.Stop()

	// Mark dirty multiple times rapidly
	for i := 0; i < 10; i++ {
		autoSaver.MarkDirty()
		time.Sleep(50 * time.Millisecond)
	}

	// Wait for at least one tick (1 second + buffer)
	time.Sleep(1500 * time.Millisecond)

	saveCount := storage.GetSaveCount()

	// Should have saved once (debounced)
	// Note: Due to timing, we might get 1 or 2 saves depending on when marks occurred
	if saveCount == 0 {
		t.Error("Should have saved at least once")
	}
	if saveCount > 3 {
		t.Errorf("Expected ~1-2 saves due to debouncing, got %d", saveCount)
	}
}

func TestAutoSaver_NoSaveWhenNotDirty(t *testing.T) {
	storage := &MockStorage{}
	state := models.NewApplicationState()
	autoSaver := NewAutoSaver(storage, state)

	autoSaver.Start()
	defer autoSaver.Stop()

	// Don't mark dirty - just wait
	time.Sleep(2500 * time.Millisecond)

	saveCount := storage.GetSaveCount()
	if saveCount != 0 {
		t.Errorf("Should not have saved when not dirty, got %d saves", saveCount)
	}
}

func TestAutoSaver_StopFlushesPendingChanges(t *testing.T) {
	storage := &MockStorage{}
	state := models.NewApplicationState()
	autoSaver := NewAutoSaver(storage, state)

	autoSaver.Start()

	// Give goroutine time to start
	time.Sleep(100 * time.Millisecond)

	// Mark dirty
	autoSaver.MarkDirty()

	// Stop immediately (before next tick)
	autoSaver.Stop()

	// Should have flushed the pending change
	saveCount := storage.GetSaveCount()
	if saveCount != 1 {
		t.Errorf("Expected 1 save on stop, got %d", saveCount)
	}
}

func TestAutoSaver_UpdateState(t *testing.T) {
	storage := &MockStorage{}
	state1 := models.NewApplicationState()
	autoSaver := NewAutoSaver(storage, state1)

	// Update state
	state2 := models.NewApplicationState()
	state2.Preferences.Theme = "light"
	autoSaver.UpdateState(state2)

	// Should be marked dirty
	if !autoSaver.IsDirty() {
		t.Error("Should be dirty after UpdateState")
	}

	// Get state should return the new state
	retrievedState := autoSaver.GetState()
	if retrievedState.Preferences.Theme != "light" {
		t.Error("Should return updated state")
	}
}

func TestAutoSaver_ConcurrentMarkDirty(t *testing.T) {
	storage := &MockStorage{}
	state := models.NewApplicationState()
	autoSaver := NewAutoSaver(storage, state)

	autoSaver.Start()
	defer autoSaver.Stop()

	var wg sync.WaitGroup
	numGoroutines := 10

	// Mark dirty from multiple goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				autoSaver.MarkDirty()
				time.Sleep(10 * time.Millisecond)
			}
		}()
	}

	wg.Wait()

	// Wait for save
	time.Sleep(1500 * time.Millisecond)

	// Should have saved without crashing
	saveCount := storage.GetSaveCount()
	if saveCount == 0 {
		t.Error("Should have saved at least once")
	}
}

func TestAutoSaver_StopIdempotent(t *testing.T) {
	storage := &MockStorage{}
	state := models.NewApplicationState()
	autoSaver := NewAutoSaver(storage, state)

	autoSaver.Start()
	autoSaver.Stop()

	// Second stop should not panic
	autoSaver.Stop()

	// Mark dirty after stop should not panic
	autoSaver.MarkDirty()
}

func TestAutoSaver_GetStateThreadSafe(t *testing.T) {
	storage := &MockStorage{}
	state := models.NewApplicationState()
	autoSaver := NewAutoSaver(storage, state)

	autoSaver.Start()
	defer autoSaver.Stop()

	var wg sync.WaitGroup

	// Read state from multiple goroutines
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = autoSaver.GetState()
			}
		}()
	}

	// Write state from another goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < 100; j++ {
			newState := models.NewApplicationState()
			autoSaver.UpdateState(newState)
			time.Sleep(time.Millisecond)
		}
	}()

	wg.Wait()
	// If we get here without deadlock or race, test passes
}

func TestAutoSaver_SavedStateIsLatest(t *testing.T) {
	storage := &MockStorage{}
	state := models.NewApplicationState()
	autoSaver := NewAutoSaver(storage, state)

	autoSaver.Start()
	defer autoSaver.Stop()

	// Update state with specific value
	newState := models.NewApplicationState()
	newState.Preferences.Theme = "custom-theme"
	autoSaver.UpdateState(newState)

	// Wait for save
	time.Sleep(1500 * time.Millisecond)

	// Check that the saved state has the correct value
	storage.mu.Lock()
	savedTheme := storage.lastState.Preferences.Theme
	storage.mu.Unlock()

	if savedTheme != "custom-theme" {
		t.Errorf("Expected saved theme 'custom-theme', got '%s'", savedTheme)
	}
}
