package fixtures

import (
	"reflect"
	"sync"
)

// MockEventBus is a mock implementation of the event bus for testing
type MockEventBus struct {
	mu              sync.RWMutex
	PublishedEvents []interface{}
	subscribers     map[reflect.Type][]func(interface{})
}

// NewMockEventBus creates a new mock event bus
func NewMockEventBus() *MockEventBus {
	return &MockEventBus{
		PublishedEvents: []interface{}{},
		subscribers:     make(map[reflect.Type][]func(interface{})),
	}
}

// Subscribe registers a handler for a specific event type
func Subscribe[T any](bus *MockEventBus, handler func(T)) {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	eventType := reflect.TypeOf((*T)(nil)).Elem()
	wrapper := func(event interface{}) {
		if typedEvent, ok := event.(T); ok {
			handler(typedEvent)
		}
	}

	bus.subscribers[eventType] = append(bus.subscribers[eventType], wrapper)
}

// Publish publishes an event to all subscribers
func Publish[T any](bus *MockEventBus, event T) {
	bus.mu.Lock()
	bus.PublishedEvents = append(bus.PublishedEvents, event)
	eventType := reflect.TypeOf(event)
	subscribers := bus.subscribers[eventType]
	bus.mu.Unlock()

	// Call subscribers without holding the lock
	for _, handler := range subscribers {
		handler(event)
	}
}

// GetEventsByType returns all published events of a specific type
func GetEventsByType[T any](bus *MockEventBus) []T {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	var result []T
	for _, event := range bus.PublishedEvents {
		if typedEvent, ok := event.(T); ok {
			result = append(result, typedEvent)
		}
	}
	return result
}

// Clear clears all published events
func (m *MockEventBus) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.PublishedEvents = []interface{}{}
}

// EventCount returns the total number of published events
func (m *MockEventBus) EventCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.PublishedEvents)
}

// HasEventType checks if any event of the given type was published
func HasEventType[T any](bus *MockEventBus) bool {
	return len(GetEventsByType[T](bus)) > 0
}
