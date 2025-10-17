package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hoytech/mcpmanager/internal/core/events"
)

// SSEHandlers contains HTTP handlers for Server-Sent Events
type SSEHandlers struct {
	eventBus        *events.EventBus
	subscriptions   map[string]*sseSubscription
	subscriptionsMu sync.RWMutex
	eventBuffer     *eventBuffer
}

// sseSubscription represents an active SSE connection
type sseSubscription struct {
	id            string
	serverIDs     map[string]bool // Filter by server IDs (empty = all)
	events        chan events.Event
	cancelFunc    context.CancelFunc
	lastEventID   string
	connectedAt   time.Time
}

// eventBuffer stores recent events for reconnection support
type eventBuffer struct {
	events    []events.Event
	maxSize   int
	mu        sync.RWMutex
}

// NewSSEHandlers creates a new SSEHandlers instance
func NewSSEHandlers(eventBus *events.EventBus) *SSEHandlers {
	handler := &SSEHandlers{
		eventBus:      eventBus,
		subscriptions: make(map[string]*sseSubscription),
		eventBuffer: &eventBuffer{
			events:  make([]events.Event, 0, 100),
			maxSize: 100,
		},
	}

	// Subscribe to all events from event bus and buffer them
	handler.startEventBuffering()

	return handler
}

// startEventBuffering subscribes to all events and buffers them for reconnection
func (h *SSEHandlers) startEventBuffering() {
	// Subscribe to all event types
	eventTypes := []events.EventType{
		events.EventServerDiscovered,
		events.EventServerStatusChanged,
		events.EventServerLogEntry,
		events.EventServerConfigChanged,
	}

	for _, eventType := range eventTypes {
		h.eventBus.Subscribe(eventType, func(event events.Event) {
			h.eventBuffer.Add(event)
		})
	}
}

// Add adds an event to the buffer
func (eb *eventBuffer) Add(event events.Event) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.events = append(eb.events, event)

	// Keep only the last maxSize events
	if len(eb.events) > eb.maxSize {
		eb.events = eb.events[len(eb.events)-eb.maxSize:]
	}
}

// GetSince retrieves events since a given event ID
func (eb *eventBuffer) GetSince(eventID string) []events.Event {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// Find the event with the given ID
	startIdx := -1
	for i, event := range eb.events {
		if event.ID == eventID {
			startIdx = i + 1 // Start from next event
			break
		}
	}

	if startIdx == -1 || startIdx >= len(eb.events) {
		return []events.Event{}
	}

	// Return events after the found ID
	result := make([]events.Event, len(eb.events)-startIdx)
	copy(result, eb.events[startIdx:])
	return result
}

// SSEStream handles GET /api/v1/events
func (h *SSEHandlers) SSEStream(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Parse query parameters
	serverIDsParam := r.URL.Query().Get("serverIds")
	lastEventID := r.Header.Get("Last-Event-ID")

	// Parse server IDs filter
	serverIDsFilter := make(map[string]bool)
	if serverIDsParam != "" {
		ids := strings.Split(serverIDsParam, ",")
		for _, id := range ids {
			trimmedID := strings.TrimSpace(id)
			if _, err := uuid.Parse(trimmedID); err == nil {
				serverIDsFilter[trimmedID] = true
			}
		}
	}

	// Create subscription
	ctx, cancel := context.WithCancel(r.Context())
	sub := &sseSubscription{
		id:          uuid.New().String(),
		serverIDs:   serverIDsFilter,
		events:      make(chan events.Event, 100),
		cancelFunc:  cancel,
		lastEventID: lastEventID,
		connectedAt: time.Now(),
	}

	// Register subscription
	h.subscriptionsMu.Lock()
	h.subscriptions[sub.id] = sub
	h.subscriptionsMu.Unlock()

	// Cleanup on disconnect
	defer func() {
		h.subscriptionsMu.Lock()
		delete(h.subscriptions, sub.id)
		h.subscriptionsMu.Unlock()
		cancel()
		close(sub.events)
	}()

	// Resend missed events if Last-Event-ID provided
	if lastEventID != "" {
		missedEvents := h.eventBuffer.GetSince(lastEventID)
		for _, event := range missedEvents {
			if h.shouldSendEvent(sub, event) {
				h.writeSSEEvent(w, event)
			}
		}
	}

	// Subscribe to event bus
	eventTypes := []events.EventType{
		events.EventServerDiscovered,
		events.EventServerStatusChanged,
		events.EventServerLogEntry,
		events.EventServerConfigChanged,
	}

	for _, eventType := range eventTypes {
		h.eventBus.Subscribe(eventType, func(event events.Event) {
			// Forward event to subscription if it matches filter
			if h.shouldSendEvent(sub, event) {
				select {
				case sub.events <- event:
				default:
					// Channel full, skip event
				}
			}
		})
	}

	// Create heartbeat ticker
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	// Flush immediately to establish connection
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Event loop
	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return

		case event := <-sub.events:
			// Write event
			h.writeSSEEvent(w, event)

			// Flush
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}

		case <-heartbeat.C:
			// Send heartbeat comment
			fmt.Fprintf(w, ": heartbeat\n\n")

			// Flush
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// shouldSendEvent checks if an event should be sent to a subscription based on filters
func (h *SSEHandlers) shouldSendEvent(sub *sseSubscription, event events.Event) bool {
	// If no server filter, send all events
	if len(sub.serverIDs) == 0 {
		return true
	}

	// Check if event's server ID matches filter
	serverID := event.Metadata["serverId"]
	if serverID == "" {
		return true // Send events without server ID
	}

	return sub.serverIDs[serverID]
}

// writeSSEEvent writes an event in SSE format
func (h *SSEHandlers) writeSSEEvent(w http.ResponseWriter, event events.Event) {
	// SSE format:
	// id: <event-id>
	// event: <event-type>
	// data: <json-payload>
	// (blank line)

	fmt.Fprintf(w, "id: %s\n", event.ID)
	fmt.Fprintf(w, "event: %s\n", event.Type)

	// Marshal event data to JSON
	dataJSON, err := json.Marshal(event)
	if err != nil {
		// Log error and send error event
		fmt.Fprintf(w, "data: {\"error\": \"failed to marshal event\"}\n\n")
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", string(dataJSON))
}
