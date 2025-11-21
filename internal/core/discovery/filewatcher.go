package discovery

import (
	"path/filepath"

	"github.com/Positronikal/MCPManager/internal/core/events"
	"github.com/fsnotify/fsnotify"
)

// ConfigFileWatcher monitors client config files for changes
type ConfigFileWatcher struct {
	watcher  *fsnotify.Watcher
	eventBus *events.EventBus
	paths    []string
	stopChan chan struct{}
}

// NewConfigFileWatcher creates a new config file watcher
func NewConfigFileWatcher(eventBus *events.EventBus, paths []string) (*ConfigFileWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &ConfigFileWatcher{
		watcher:  watcher,
		eventBus: eventBus,
		paths:    paths,
		stopChan: make(chan struct{}),
	}, nil
}

// Start begins watching the configured paths
func (cfw *ConfigFileWatcher) Start() error {
	// Add all paths to watcher
	for _, path := range cfw.paths {
		// Watch the parent directory since the file might not exist yet
		dir := filepath.Dir(path)
		if err := cfw.watcher.Add(dir); err != nil {
			// Log error but continue with other paths
			continue
		}
	}

	// Start watching in background
	go cfw.watch()

	return nil
}

// watch is the main watching loop
func (cfw *ConfigFileWatcher) watch() {
	for {
		select {
		case event, ok := <-cfw.watcher.Events:
			if !ok {
				return
			}

			// Check if this event is for one of our watched files
			if cfw.isWatchedFile(event.Name) {
				// Only care about Write and Create events
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create {
					// Publish config file changed event
					if cfw.eventBus != nil {
						cfw.eventBus.Publish(events.ConfigFileChangedEvent(event.Name))
					}
				}
			}

		case err, ok := <-cfw.watcher.Errors:
			if !ok {
				return
			}
			// Log error but continue watching
			_ = err

		case <-cfw.stopChan:
			return
		}
	}
}

// isWatchedFile checks if the given path matches one of our watched files
func (cfw *ConfigFileWatcher) isWatchedFile(path string) bool {
	for _, watchedPath := range cfw.paths {
		if path == watchedPath {
			return true
		}
	}
	return false
}

// Stop stops watching files
func (cfw *ConfigFileWatcher) Stop() error {
	close(cfw.stopChan)
	return cfw.watcher.Close()
}

// AddPath adds a new path to watch
func (cfw *ConfigFileWatcher) AddPath(path string) error {
	// Add to paths list
	cfw.paths = append(cfw.paths, path)

	// Watch the parent directory
	dir := filepath.Dir(path)
	return cfw.watcher.Add(dir)
}

// RemovePath stops watching a specific path
func (cfw *ConfigFileWatcher) RemovePath(path string) error {
	// Remove from paths list
	for i, p := range cfw.paths {
		if p == path {
			cfw.paths = append(cfw.paths[:i], cfw.paths[i+1:]...)
			break
		}
	}

	// Note: We don't remove the directory from watcher since other files
	// in the same directory might still be watched
	return nil
}
