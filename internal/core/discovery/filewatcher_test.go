package discovery

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Positronikal/MCPManager/internal/core/events"
)

func TestNewConfigFileWatcher(t *testing.T) {
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	paths := []string{"/path/to/config.json"}
	watcher, err := NewConfigFileWatcher(eventBus, paths)
	if err != nil {
		t.Fatalf("Expected watcher to be created, got error: %v", err)
	}

	if watcher == nil {
		t.Fatal("Expected non-nil watcher")
	}

	if len(watcher.paths) != 1 {
		t.Errorf("Expected 1 path, got %d", len(watcher.paths))
	}

	if watcher.eventBus != eventBus {
		t.Error("Event bus should be set")
	}

	if watcher.watcher == nil {
		t.Error("fsnotify watcher should be initialized")
	}

	if watcher.stopChan == nil {
		t.Error("stopChan should be initialized")
	}

	// Clean up
	watcher.Stop()
}

func TestConfigFileWatcher_DetectFileModification(t *testing.T) {
	// Create a temporary directory and file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.json")

	// Create the file
	if err := os.WriteFile(testFile, []byte(`{"test": "data"}`), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set up event bus and watcher
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	watcher, err := NewConfigFileWatcher(eventBus, []string{testFile})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Subscribe to config file changed events
	eventChan := eventBus.Subscribe(events.EventConfigFileChanged)

	// Start watching
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Give the watcher time to initialize
	time.Sleep(100 * time.Millisecond)

	// Modify the file
	if err := os.WriteFile(testFile, []byte(`{"test": "modified"}`), 0644); err != nil {
		t.Fatalf("Failed to modify file: %v", err)
	}

	// Wait for the event
	select {
	case event := <-eventChan:
		if event == nil {
			t.Fatal("Received nil event")
		}
		if event.Type != events.EventConfigFileChanged {
			t.Errorf("Expected EventConfigFileChanged, got %s", event.Type)
		}
		// Check that the data contains the file path
		if filePath, ok := event.Data["filePath"].(string); !ok || filePath != testFile {
			t.Errorf("Expected file path %s in data, got %v", testFile, event.Data["filePath"])
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for file modification event")
	}
}

func TestConfigFileWatcher_DetectFileCreation(t *testing.T) {
	// Create a temporary directory (no file yet)
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "newconfig.json")

	// Set up event bus and watcher
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	watcher, err := NewConfigFileWatcher(eventBus, []string{testFile})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Subscribe to config file changed events
	eventChan := eventBus.Subscribe(events.EventConfigFileChanged)

	// Start watching
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Give the watcher time to initialize
	time.Sleep(100 * time.Millisecond)

	// Create the file
	if err := os.WriteFile(testFile, []byte(`{"test": "created"}`), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Wait for the event
	select {
	case event := <-eventChan:
		if event == nil {
			t.Fatal("Received nil event")
		}
		if event.Type != events.EventConfigFileChanged {
			t.Errorf("Expected EventConfigFileChanged, got %s", event.Type)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for file creation event")
	}
}

func TestConfigFileWatcher_IgnoreNonWatchedFiles(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()
	watchedFile := filepath.Join(tmpDir, "watched.json")
	unwatchedFile := filepath.Join(tmpDir, "unwatched.json")

	// Create both files
	os.WriteFile(watchedFile, []byte(`{"watched": true}`), 0644)
	os.WriteFile(unwatchedFile, []byte(`{"watched": false}`), 0644)

	// Set up event bus and watcher (only watch one file)
	eventBus := events.NewEventBus()
	defer eventBus.Close()

	watcher, err := NewConfigFileWatcher(eventBus, []string{watchedFile})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Subscribe to config file changed events
	eventChan := eventBus.Subscribe(events.EventConfigFileChanged)

	// Start watching
	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	// Give the watcher time to initialize
	time.Sleep(100 * time.Millisecond)

	// Modify the unwatched file
	if err := os.WriteFile(unwatchedFile, []byte(`{"watched": "still false"}`), 0644); err != nil {
		t.Fatalf("Failed to modify unwatched file: %v", err)
	}

	// Should NOT receive an event
	select {
	case event := <-eventChan:
		t.Errorf("Should not receive event for unwatched file, got: %v", event)
	case <-time.After(500 * time.Millisecond):
		// Expected - no event received
	}

	// Now modify the watched file
	if err := os.WriteFile(watchedFile, []byte(`{"watched": "modified"}`), 0644); err != nil {
		t.Fatalf("Failed to modify watched file: %v", err)
	}

	// Should receive an event
	select {
	case event := <-eventChan:
		if event == nil {
			t.Fatal("Received nil event")
		}
		if filePath, ok := event.Data["filePath"].(string); !ok || filePath != watchedFile {
			t.Errorf("Expected watched file path, got %v", event.Data["filePath"])
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for watched file event")
	}
}

func TestConfigFileWatcher_StartStop(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.json")
	os.WriteFile(testFile, []byte(`{"test": "data"}`), 0644)

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	watcher, err := NewConfigFileWatcher(eventBus, []string{testFile})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	// Start
	if err := watcher.Start(); err != nil {
		t.Errorf("Start() should not error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Stop
	if err := watcher.Stop(); err != nil {
		t.Errorf("Stop() should not error: %v", err)
	}

	// Modifying file after stop should not trigger events
	eventChan := eventBus.Subscribe(events.EventConfigFileChanged)

	os.WriteFile(testFile, []byte(`{"test": "after stop"}`), 0644)

	select {
	case event := <-eventChan:
		t.Errorf("Should not receive event after stop, got: %v", event)
	case <-time.After(500 * time.Millisecond):
		// Expected - no event
	}
}

func TestConfigFileWatcher_AddPath(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "config1.json")
	file2 := filepath.Join(tmpDir, "config2.json")

	os.WriteFile(file1, []byte(`{"file": 1}`), 0644)
	os.WriteFile(file2, []byte(`{"file": 2}`), 0644)

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	// Start with only file1
	watcher, err := NewConfigFileWatcher(eventBus, []string{file1})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Add file2
	if err := watcher.AddPath(file2); err != nil {
		t.Errorf("AddPath() should not error: %v", err)
	}

	// Verify file2 is now watched
	if len(watcher.paths) != 2 {
		t.Errorf("Expected 2 paths after AddPath, got %d", len(watcher.paths))
	}

	eventChan := eventBus.Subscribe(events.EventConfigFileChanged)

	// Modify file2
	os.WriteFile(file2, []byte(`{"file": "2 modified"}`), 0644)

	// Should receive event
	select {
	case event := <-eventChan:
		if event == nil {
			t.Fatal("Received nil event")
		}
		if filePath, ok := event.Data["filePath"].(string); !ok || filePath != file2 {
			t.Errorf("Expected file2 path, got %v", event.Data["filePath"])
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for file2 event after AddPath")
	}
}

func TestConfigFileWatcher_RemovePath(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "config1.json")
	file2 := filepath.Join(tmpDir, "config2.json")

	os.WriteFile(file1, []byte(`{"file": 1}`), 0644)
	os.WriteFile(file2, []byte(`{"file": 2}`), 0644)

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	// Start with both files
	watcher, err := NewConfigFileWatcher(eventBus, []string{file1, file2})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Remove file2
	if err := watcher.RemovePath(file2); err != nil {
		t.Errorf("RemovePath() should not error: %v", err)
	}

	// Verify file2 is removed
	if len(watcher.paths) != 1 {
		t.Errorf("Expected 1 path after RemovePath, got %d", len(watcher.paths))
	}

	eventChan := eventBus.Subscribe(events.EventConfigFileChanged)

	// Modify file2 (should not trigger event)
	os.WriteFile(file2, []byte(`{"file": "2 modified"}`), 0644)

	select {
	case event := <-eventChan:
		t.Errorf("Should not receive event for removed file, got: %v", event)
	case <-time.After(500 * time.Millisecond):
		// Expected - no event
	}

	// Modify file1 (should still trigger event)
	os.WriteFile(file1, []byte(`{"file": "1 modified"}`), 0644)

	select {
	case event := <-eventChan:
		if event == nil {
			t.Fatal("Received nil event")
		}
		if filePath, ok := event.Data["filePath"].(string); !ok || filePath != file1 {
			t.Errorf("Expected file1 path, got %v", event.Data["filePath"])
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for file1 event")
	}
}

func TestConfigFileWatcher_MultiplePaths(t *testing.T) {
	// Test that we can watch multiple files at once
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "config1.json")
	file2 := filepath.Join(tmpDir, "config2.json")

	os.WriteFile(file1, []byte(`{"file": 1}`), 0644)
	os.WriteFile(file2, []byte(`{"file": 2}`), 0644)

	eventBus := events.NewEventBus()
	defer eventBus.Close()

	watcher, err := NewConfigFileWatcher(eventBus, []string{file1, file2})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Verify both paths are registered
	if len(watcher.paths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(watcher.paths))
	}

	if err := watcher.Start(); err != nil {
		t.Fatalf("Failed to start watcher: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	eventChan := eventBus.Subscribe(events.EventConfigFileChanged)

	// Modify file1 and verify we get an event
	os.WriteFile(file1, []byte(`{"file": "1 modified"}`), 0644)

	select {
	case event := <-eventChan:
		if event == nil {
			t.Fatal("Received nil event")
		}
		if filePath, ok := event.Data["filePath"].(string); !ok || filePath != file1 {
			t.Errorf("Expected file1 path, got %v", event.Data["filePath"])
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for file1 event")
	}

	// Drain any remaining events from file1 modification
	time.Sleep(200 * time.Millisecond)
drainLoop1:
	for {
		select {
		case <-eventChan:
			// Drain any duplicate events
		default:
			break drainLoop1
		}
	}

	// Modify file2 and verify we get an event
	os.WriteFile(file2, []byte(`{"file": "2 modified"}`), 0644)

	select {
	case event := <-eventChan:
		if event == nil {
			t.Fatal("Received nil event")
		}
		if filePath, ok := event.Data["filePath"].(string); !ok || filePath != file2 {
			t.Errorf("Expected file2 path, got %v", event.Data["filePath"])
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for file2 event")
	}
}

func TestConfigFileWatcher_NoEventBus(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "config.json")
	os.WriteFile(testFile, []byte(`{"test": "data"}`), 0644)

	// Create watcher with nil event bus
	watcher, err := NewConfigFileWatcher(nil, []string{testFile})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Stop()

	// Should not crash when starting
	if err := watcher.Start(); err != nil {
		t.Errorf("Start() should not error with nil event bus: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Modify file - should not crash even though event bus is nil
	os.WriteFile(testFile, []byte(`{"test": "modified"}`), 0644)

	time.Sleep(200 * time.Millisecond)

	// If we got here without crashing, test passes
}
