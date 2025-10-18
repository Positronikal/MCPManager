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
}

// sseSubscription represents an active SSE connection
type sseSubscription struct {
	id          string
	serverIDs   map[string]bool // Filter by server IDs (empty = all)
	cancelFunc  context.CancelFunc
	connectedAt time.Time
	lastEventID int // Simple counter for event IDs
}

// NewSSEHandlers creates a new SSEHandlers instance
func NewSSEHandlers(eventBus *events.EventBus) *SSEHandlers {
	return &SSEHandlers{
		eventBus:      eventBus,
		subscriptions: make(map[string]*sseSubscription),
	}
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
		cancelFunc:  cancel,
		connectedAt: time.Now(),
		lastEventID: 0,
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
	}()

	// Subscribe to all event types from EventBus
	eventTypes := []events.EventType{
		events.EventServerDiscovered,
		events.EventServerStatusChanged,
		events.EventServerLogEntry,
		events.EventConfigFileChanged,
		events.EventServerMetricsUpdated,
	}

	// Create a combined channel for all events
	combinedChan := make(chan *events.Event, 100)
	var channels []<-chan *events.Event

	// Subscribe to each event type
	for _, eventType := range eventTypes {
		ch := h.eventBus.Subscribe(eventType)
		channels = append(channels, ch)
	}

	// Start goroutine to merge all channels into one
	go func() {
		for _, ch := range channels {
			go func(eventChan <-chan *events.Event) {
				for event := range eventChan {
					if event != nil {
						select {
						case combinedChan <- event:
						case <-ctx.Done():
							return
						}
					}
				}
			}(ch)
		}
	}()

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

		case event := <-combinedChan:
			// Filter event by server ID if needed
			if !h.shouldSendEvent(sub, event) {
				continue
			}

			// Write event
			sub.lastEventID++
			h.writeSSEEvent(w, event, sub.lastEventID)

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
func (h *SSEHandlers) shouldSendEvent(sub *sseSubscription, event *events.Event) bool {
	// If no server filter, send all events
	if len(sub.serverIDs) == 0 {
		return true
	}

	// Check if event's server ID matches filter
	serverID, ok := event.Data["serverID"].(string)
	if !ok || serverID == "" {
		return true // Send events without server ID
	}

	return sub.serverIDs[serverID]
}

// writeSSEEvent writes an event in SSE format
func (h *SSEHandlers) writeSSEEvent(w http.ResponseWriter, event *events.Event, eventID int) {
	// SSE format:
	// id: <event-id>
	// event: <event-type>
	// data: <json-payload>
	// (blank line)

	fmt.Fprintf(w, "id: %d\n", eventID)
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
