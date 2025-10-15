package events

import (
	"sync"
	"testing"
	"time"

	"github.com/hoytech/mcpmanager/internal/models"
)

func TestNewEventBus(t *testing.T) {
	bus := NewEventBus()
	if bus == nil {
		t.Fatal("Expected event bus to be created")
	}
	if bus.closed {
		t.Error("Event bus should not be closed initially")
	}
}

func TestEventBus_Subscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	ch := bus.Subscribe(EventServerDiscovered)
	if ch == nil {
		t.Fatal("Expected subscriber channel to be created")
	}

	if bus.SubscriberCount(EventServerDiscovered) != 1 {
		t.Errorf("Expected 1 subscriber, got %d", bus.SubscriberCount(EventServerDiscovered))
	}
}

func TestEventBus_PublishAndReceive(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	ch := bus.Subscribe(EventServerDiscovered)

	event := NewEvent(EventServerDiscovered, map[string]interface{}{
		"serverID": "test-123",
	})

	bus.Publish(event)

	select {
	case received := <-ch:
		if received.Type != EventServerDiscovered {
			t.Errorf("Expected event type %s, got %s", EventServerDiscovered, received.Type)
		}
		if received.Data["serverID"] != "test-123" {
			t.Error("Event data not preserved")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Timeout waiting for event")
	}
}

func TestEventBus_MultipleSubscribers(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	ch1 := bus.Subscribe(EventServerStatusChanged)
	ch2 := bus.Subscribe(EventServerStatusChanged)
	ch3 := bus.Subscribe(EventServerStatusChanged)

	if bus.SubscriberCount(EventServerStatusChanged) != 3 {
		t.Errorf("Expected 3 subscribers, got %d", bus.SubscriberCount(EventServerStatusChanged))
	}

	event := NewEvent(EventServerStatusChanged, map[string]interface{}{
		"test": "data",
	})

	bus.Publish(event)

	// All subscribers should receive the event
	channels := []<-chan *Event{ch1, ch2, ch3}
	for i, ch := range channels {
		select {
		case received := <-ch:
			if received.Type != EventServerStatusChanged {
				t.Errorf("Subscriber %d received wrong event type", i)
			}
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Subscriber %d timeout waiting for event", i)
		}
	}
}

func TestEventBus_DifferentEventTypes(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	ch1 := bus.Subscribe(EventServerDiscovered)
	ch2 := bus.Subscribe(EventServerStatusChanged)

	// Publish to first type
	event1 := NewEvent(EventServerDiscovered, map[string]interface{}{})
	bus.Publish(event1)

	// ch1 should receive, ch2 should not
	select {
	case <-ch1:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("ch1 should have received event")
	}

	select {
	case <-ch2:
		t.Error("ch2 should not have received event")
	case <-time.After(50 * time.Millisecond):
		// Expected - timeout means no event received
	}

	// Publish to second type
	event2 := NewEvent(EventServerStatusChanged, map[string]interface{}{})
	bus.Publish(event2)

	select {
	case <-ch2:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("ch2 should have received event")
	}
}

func TestEventBus_ConcurrentPublish(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	ch := bus.Subscribe(EventServerLogEntry)

	var wg sync.WaitGroup
	numPublishers := 10
	eventsPerPublisher := 10

	// Start multiple publishers
	for i := 0; i < numPublishers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < eventsPerPublisher; j++ {
				event := NewEvent(EventServerLogEntry, map[string]interface{}{
					"publisher": id,
					"seq":       j,
				})
				bus.Publish(event)
			}
		}(i)
	}

	// Collect events
	receivedCount := 0
	done := make(chan struct{})

	go func() {
		for range ch {
			receivedCount++
			if receivedCount >= numPublishers*eventsPerPublisher {
				close(done)
				return
			}
		}
	}()

	wg.Wait()

	select {
	case <-done:
		// All events received
	case <-time.After(2 * time.Second):
		t.Errorf("Expected %d events, got %d", numPublishers*eventsPerPublisher, receivedCount)
	}
}

func TestEventBus_BufferOverflow(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	ch := bus.Subscribe(EventServerLogEntry)

	// Publish more events than buffer size (100)
	for i := 0; i < 150; i++ {
		event := NewEvent(EventServerLogEntry, map[string]interface{}{
			"seq": i,
		})
		bus.Publish(event)
	}

	// Should not block even though we exceeded buffer
	// Some events may be dropped, but we should still receive some
	receivedCount := 0
	timeout := time.After(100 * time.Millisecond)

drainLoop:
	for {
		select {
		case <-ch:
			receivedCount++
		case <-timeout:
			break drainLoop
		}
	}

	if receivedCount == 0 {
		t.Error("Should have received at least some events")
	}
	if receivedCount > 150 {
		t.Errorf("Received more events than published: %d", receivedCount)
	}
}

func TestEventBus_Close(t *testing.T) {
	bus := NewEventBus()

	ch1 := bus.Subscribe(EventServerDiscovered)
	ch2 := bus.Subscribe(EventServerStatusChanged)

	bus.Close()

	// Channels should be closed
	_, ok1 := <-ch1
	if ok1 {
		t.Error("Channel should be closed")
	}

	_, ok2 := <-ch2
	if ok2 {
		t.Error("Channel should be closed")
	}

	// Subscribe after close should return closed channel
	ch3 := bus.Subscribe(EventServerDiscovered)
	_, ok3 := <-ch3
	if ok3 {
		t.Error("Channel should be closed")
	}

	// Publish after close should not panic
	event := NewEvent(EventServerDiscovered, map[string]interface{}{})
	bus.Publish(event) // Should not panic
}

func TestEventBus_Unsubscribe(t *testing.T) {
	bus := NewEventBus()
	defer bus.Close()

	ch := bus.Subscribe(EventServerDiscovered)

	if bus.SubscriberCount(EventServerDiscovered) != 1 {
		t.Error("Expected 1 subscriber")
	}

	bus.Unsubscribe(EventServerDiscovered, ch)

	if bus.SubscriberCount(EventServerDiscovered) != 0 {
		t.Error("Expected 0 subscribers after unsubscribe")
	}

	// Channel should be closed
	_, ok := <-ch
	if ok {
		t.Error("Channel should be closed after unsubscribe")
	}
}

func TestEventHelpers(t *testing.T) {
	// Test ServerDiscoveredEvent
	server := models.NewMCPServer("test", "/path", models.DiscoveryClientConfig)
	event := ServerDiscoveredEvent(server)
	if event.Type != EventServerDiscovered {
		t.Error("Wrong event type")
	}
	if event.Data["serverID"] != server.ID {
		t.Error("Server ID not in event data")
	}

	// Test ServerStatusChangedEvent
	event = ServerStatusChangedEvent("server-123", models.StatusStopped, models.StatusRunning)
	if event.Type != EventServerStatusChanged {
		t.Error("Wrong event type")
	}

	// Test ConfigFileChangedEvent
	event = ConfigFileChangedEvent("/path/to/config.json")
	if event.Type != EventConfigFileChanged {
		t.Error("Wrong event type")
	}
	if event.Data["filePath"] != "/path/to/config.json" {
		t.Error("File path not in event data")
	}
}
