package events

import (
	"sync"
	"time"

	"github.com/Positronikal/MCPManager/internal/models"
)

// EventType represents the type of event
type EventType string

const (
	EventServerDiscovered      EventType = "server.discovered"
	EventServerStatusChanged   EventType = "server.status.changed"
	EventServerLogEntry        EventType = "server.log.entry"
	EventConfigFileChanged     EventType = "config.file.changed"
	EventServerMetricsUpdated  EventType = "server.metrics.updated"
)

// Event represents a generic event in the system
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewEvent creates a new event with the given type and data
func NewEvent(eventType EventType, data map[string]interface{}) *Event {
	return &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
}

// ServerDiscoveredEvent creates a server discovered event
func ServerDiscoveredEvent(server *models.MCPServer) *Event {
	return NewEvent(EventServerDiscovered, map[string]interface{}{
		"serverID": server.ID,
		"name":     server.Name,
		"source":   server.Source,
	})
}

// ServerStatusChangedEvent creates a server status changed event
func ServerStatusChangedEvent(serverID string, oldState, newState models.StatusState) *Event {
	return NewEvent(EventServerStatusChanged, map[string]interface{}{
		"serverID": serverID,
		"oldState": oldState,
		"newState": newState,
	})
}

// ServerLogEntryEvent creates a server log entry event
func ServerLogEntryEvent(serverID string, entry *models.LogEntry) *Event {
	return NewEvent(EventServerLogEntry, map[string]interface{}{
		"serverID": serverID,
		"severity": entry.Severity,
		"message":  entry.Message,
	})
}

// ConfigFileChangedEvent creates a config file changed event
func ConfigFileChangedEvent(filePath string) *Event {
	return NewEvent(EventConfigFileChanged, map[string]interface{}{
		"filePath": filePath,
	})
}

// ServerMetricsUpdatedEvent creates a server metrics updated event
func ServerMetricsUpdatedEvent(serverID string, metrics map[string]interface{}) *Event {
	return NewEvent(EventServerMetricsUpdated, map[string]interface{}{
		"serverID": serverID,
		"metrics":  metrics,
	})
}

// EventBus is a lightweight pub/sub event bus
type EventBus struct {
	subscribers map[EventType][]chan *Event
	mu          sync.RWMutex
	closed      bool
}

// NewEventBus creates a new event bus
func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[EventType][]chan *Event),
		closed:      false,
	}
}

// Subscribe subscribes to events of a specific type
// Returns a buffered channel that will receive events
func (eb *EventBus) Subscribe(eventType EventType) <-chan *Event {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		// Return a closed channel if the bus is closed
		ch := make(chan *Event)
		close(ch)
		return ch
	}

	// Create a buffered channel to prevent blocking
	ch := make(chan *Event, 100)
	eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)

	return ch
}

// Publish publishes an event to all subscribers
func (eb *EventBus) Publish(event *Event) {
	if event == nil {
		return
	}

	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if eb.closed {
		return
	}

	// Get subscribers for this event type
	subscribers, exists := eb.subscribers[event.Type]
	if !exists {
		return
	}

	// Send to all subscribers (non-blocking)
	for _, ch := range subscribers {
		select {
		case ch <- event:
			// Event sent successfully
		default:
			// Channel full, drop event to prevent blocking
			// In production, you might want to log this
		}
	}
}

// Unsubscribe removes a subscriber channel
func (eb *EventBus) Unsubscribe(eventType EventType, ch <-chan *Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subscribers, exists := eb.subscribers[eventType]
	if !exists {
		return
	}

	// Find and remove the channel
	for i, subscriber := range subscribers {
		if subscriber == ch {
			// Remove by swapping with last element and truncating
			subscribers[i] = subscribers[len(subscribers)-1]
			eb.subscribers[eventType] = subscribers[:len(subscribers)-1]
			close(subscriber)
			break
		}
	}
}

// Close closes the event bus and all subscriber channels
func (eb *EventBus) Close() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.closed {
		return
	}

	eb.closed = true

	// Close all subscriber channels
	for _, subscribers := range eb.subscribers {
		for _, ch := range subscribers {
			close(ch)
		}
	}

	// Clear subscribers map
	eb.subscribers = make(map[EventType][]chan *Event)
}

// SubscriberCount returns the number of subscribers for a given event type
func (eb *EventBus) SubscriberCount(eventType EventType) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	return len(eb.subscribers[eventType])
}
