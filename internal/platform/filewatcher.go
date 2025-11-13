package platform

// Event represents a file system event
type Event struct {
	Path string
	Op   string
}

// FileWatcher defines the interface for monitoring file system changes
type FileWatcher interface {
	// Watch starts monitoring the specified path
	Watch(path string) error

	// Events returns a channel for receiving file system events
	Events() <-chan Event

	// Close stops the watcher and releases resources
	Close() error
}
