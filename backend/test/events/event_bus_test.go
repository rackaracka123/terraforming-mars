package events_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"terraforming-mars-backend/internal/events"
)

// TestEvent is a simple event for testing
type TestEvent struct {
	Message string
	Value   int
}

// AnotherTestEvent is another event type for testing
type AnotherTestEvent struct {
	Data string
}

func TestEventBus_Subscribe_Success(t *testing.T) {
	bus := events.NewEventBus()
	called := false

	events.Subscribe(bus, func(event TestEvent) {
		called = true
	})

	events.Publish(bus, TestEvent{Message: "test", Value: 42})

	assert.True(t, called, "Handler should have been called")
}

func TestEventBus_Publish_SingleSubscriber(t *testing.T) {
	bus := events.NewEventBus()
	var receivedEvent TestEvent

	events.Subscribe(bus, func(event TestEvent) {
		receivedEvent = event
	})

	expectedEvent := TestEvent{Message: "hello", Value: 123}
	events.Publish(bus, expectedEvent)

	assert.Equal(t, expectedEvent.Message, receivedEvent.Message)
	assert.Equal(t, expectedEvent.Value, receivedEvent.Value)
}

func TestEventBus_Publish_MultipleSubscribers(t *testing.T) {
	bus := events.NewEventBus()
	callCount := 0
	var mu sync.Mutex

	// Subscribe three handlers
	for i := 0; i < 3; i++ {
		events.Subscribe(bus, func(event TestEvent) {
			mu.Lock()
			callCount++
			mu.Unlock()
		})
	}

	events.Publish(bus, TestEvent{Message: "broadcast", Value: 1})

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 3, callCount, "All three handlers should have been called")
}

func TestEventBus_Publish_NoSubscribers(t *testing.T) {
	bus := events.NewEventBus()

	// Publishing with no subscribers should not panic
	assert.NotPanics(t, func() {
		events.Publish(bus, TestEvent{Message: "orphan", Value: 0})
	})
}

func TestEventBus_Publish_DifferentEventTypes(t *testing.T) {
	bus := events.NewEventBus()
	testEventCalled := false
	anotherEventCalled := false

	events.Subscribe(bus, func(event TestEvent) {
		testEventCalled = true
	})

	events.Subscribe(bus, func(event AnotherTestEvent) {
		anotherEventCalled = true
	})

	// Publish TestEvent - only TestEvent handler should be called
	events.Publish(bus, TestEvent{Message: "test", Value: 1})
	assert.True(t, testEventCalled)
	assert.False(t, anotherEventCalled)

	// Reset and publish AnotherTestEvent
	testEventCalled = false
	anotherEventCalled = false

	events.Publish(bus, AnotherTestEvent{Data: "other"})
	assert.False(t, testEventCalled)
	assert.True(t, anotherEventCalled)
}

func TestEventBus_TypeSafety(t *testing.T) {
	bus := events.NewEventBus()
	var receivedValue int

	events.Subscribe(bus, func(event TestEvent) {
		receivedValue = event.Value
	})

	// This should work - correct type
	events.Publish(bus, TestEvent{Message: "correct", Value: 100})
	assert.Equal(t, 100, receivedValue)

	// This should not call the TestEvent handler - different type
	events.Publish(bus, AnotherTestEvent{Data: "wrong type"})
	assert.Equal(t, 100, receivedValue, "Value should not change from different event type")
}

func TestEventBus_ThreadSafety_ConcurrentPublish(t *testing.T) {
	bus := events.NewEventBus()
	var mu sync.Mutex
	callCount := 0

	events.Subscribe(bus, func(event TestEvent) {
		mu.Lock()
		callCount++
		mu.Unlock()
	})

	// Publish 100 events concurrently
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(value int) {
			defer wg.Done()
			events.Publish(bus, TestEvent{Message: "concurrent", Value: value})
		}(i)
	}

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 100, callCount, "All 100 concurrent events should have been processed")
}

func TestEventBus_ThreadSafety_ConcurrentSubscribe(t *testing.T) {
	bus := events.NewEventBus()
	var mu sync.Mutex
	callCount := 0

	// Subscribe 50 handlers concurrently
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			events.Subscribe(bus, func(event TestEvent) {
				mu.Lock()
				callCount++
				mu.Unlock()
			})
		}()
	}

	wg.Wait()

	// Publish one event - all 50 handlers should be called
	events.Publish(bus, TestEvent{Message: "test", Value: 1})

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 50, callCount, "All 50 concurrently subscribed handlers should have been called")
}

func TestEventBus_MultipleEventsInSequence(t *testing.T) {
	bus := events.NewEventBus()
	var receivedValues []int

	events.Subscribe(bus, func(event TestEvent) {
		receivedValues = append(receivedValues, event.Value)
	})

	// Publish multiple events in sequence
	for i := 1; i <= 5; i++ {
		events.Publish(bus, TestEvent{Message: "sequence", Value: i})
	}

	assert.Equal(t, []int{1, 2, 3, 4, 5}, receivedValues, "Events should be received in order")
}

func TestEventBus_HandlerCanPublishNewEvent(t *testing.T) {
	bus := events.NewEventBus()
	testEventCalled := false
	anotherEventCalled := false

	// Handler that publishes another event
	events.Subscribe(bus, func(event TestEvent) {
		testEventCalled = true
		events.Publish(bus, AnotherTestEvent{Data: "triggered"})
	})

	events.Subscribe(bus, func(event AnotherTestEvent) {
		anotherEventCalled = true
	})

	events.Publish(bus, TestEvent{Message: "trigger", Value: 1})

	assert.True(t, testEventCalled, "First handler should be called")
	assert.True(t, anotherEventCalled, "Second event handler should be called")
}

func TestEventBus_Clear_RemovesAllSubscriptions(t *testing.T) {
	bus := events.NewEventBus()
	called := false

	events.Subscribe(bus, func(event TestEvent) {
		called = true
	})

	// Clear all subscriptions
	bus.Clear()

	// Publish event - handler should not be called
	events.Publish(bus, TestEvent{Message: "cleared", Value: 1})

	assert.False(t, called, "Handler should not be called after Clear()")
}
