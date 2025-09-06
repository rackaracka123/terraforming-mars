package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestEvent for testing purposes
type TestEvent struct {
	BaseEvent
	Data string
}

// NewTestEvent creates a new test event
func NewTestEvent(gameID, data string) *TestEvent {
	return &TestEvent{
		BaseEvent: NewBaseEvent("test-event", gameID, data),
		Data:      data,
	}
}

func TestInMemoryEventBus_PublishSubscribe(t *testing.T) {
	bus := NewInMemoryEventBus()
	ctx := context.Background()

	// Create a channel to receive events
	eventsCh := make(chan Event, 10)

	// Subscribe to test events
	bus.Subscribe("test-event", func(ctx context.Context, event Event) error {
		eventsCh <- event
		return nil
	})

	// Publish a test event
	testEvent := NewTestEvent("game1", "test data")

	err := bus.Publish(ctx, testEvent)
	assert.NoError(t, err)

	// Wait for the event
	select {
	case receivedEvent := <-eventsCh:
		assert.Equal(t, "game1", receivedEvent.GetGameID())
		assert.Equal(t, "test-event", receivedEvent.GetType())

		// Cast back to TestEvent to check data
		if testEv, ok := receivedEvent.(*TestEvent); ok {
			assert.Equal(t, "test data", testEv.Data)
		} else {
			t.Error("Event should be of type TestEvent")
		}
	case <-time.After(1 * time.Second):
		t.Error("Expected to receive event within 1 second")
	}
}

func TestInMemoryEventBus_NoSubscribers(t *testing.T) {
	bus := NewInMemoryEventBus()
	ctx := context.Background()

	// Publish an event with no subscribers
	testEvent := NewTestEvent("game1", "no subscribers")

	// Should not return an error even if no one is subscribed
	err := bus.Publish(ctx, testEvent)
	assert.NoError(t, err)
}

func TestInMemoryEventBus_Subscribe(t *testing.T) {
	bus := NewInMemoryEventBus()

	// Test that Subscribe doesn't panic or return error
	bus.Subscribe("test", func(ctx context.Context, event Event) error {
		return nil
	})

	// Test subscribing multiple listeners
	bus.Subscribe("test", func(ctx context.Context, event Event) error {
		return nil
	})

	// Verify internal state (listeners should be added)
	assert.NotNil(t, bus.listeners)
	assert.Len(t, bus.listeners["test"], 2)
}
