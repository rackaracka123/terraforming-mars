package events

import (
	"context"
	"sync"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// EventListener represents a function that handles an event
type EventListener func(ctx context.Context, event Event) error

// EventBus defines the interface for event publishing and subscription
type EventBus interface {
	// Subscribe registers a listener for events of the specified type
	Subscribe(eventType string, listener EventListener)
	// Publish sends an event to all registered listeners for its type
	Publish(ctx context.Context, event Event) error
	// Unsubscribe removes a listener (if needed for testing)
	Unsubscribe(eventType string, listener EventListener)
}

// InMemoryEventBus implements EventBus using in-memory subscription storage
type InMemoryEventBus struct {
	listeners map[string][]EventListener
	mutex     sync.RWMutex
}

// NewInMemoryEventBus creates a new in-memory event bus
func NewInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		listeners: make(map[string][]EventListener),
	}
}

// Subscribe registers a listener for events of the specified type
func (bus *InMemoryEventBus) Subscribe(eventType string, listener EventListener) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	if bus.listeners[eventType] == nil {
		bus.listeners[eventType] = make([]EventListener, 0)
	}

	bus.listeners[eventType] = append(bus.listeners[eventType], listener)

	log := logger.Get()
	log.Info("Event listener registered",
		zap.String("event_type", eventType),
		zap.Int("listener_count", len(bus.listeners[eventType])),
	)
}

// Publish sends an event to all registered listeners for its type
func (bus *InMemoryEventBus) Publish(ctx context.Context, event Event) error {
	bus.mutex.RLock()
	listeners := bus.listeners[event.GetType()]
	bus.mutex.RUnlock()

	log := logger.WithGameContext(event.GetGameID(), "")

	if len(listeners) == 0 {
		log.Debug("No listeners registered for event type",
			zap.String("event_type", event.GetType()),
		)
		return nil
	}

	log.Info("Publishing event",
		zap.String("event_type", event.GetType()),
		zap.Int("listener_count", len(listeners)),
	)

	// Execute listeners concurrently
	var wg sync.WaitGroup
	errorChan := make(chan error, len(listeners))

	for _, listener := range listeners {
		wg.Add(1)
		go func(l EventListener) {
			defer wg.Done()
			if err := l(ctx, event); err != nil {
				log.Error("Event listener failed",
					zap.String("event_type", event.GetType()),
					zap.Error(err),
				)
				errorChan <- err
			}
		}(listener)
	}

	wg.Wait()
	close(errorChan)

	// Check for errors
	for err := range errorChan {
		if err != nil {
			return err
		}
	}

	log.Info("Event published",
		zap.String("event_type", event.GetType()),
		zap.Int("listeners_executed", len(listeners)),
	)

	return nil
}

// Unsubscribe removes a listener from the event type (used mainly for testing)
func (bus *InMemoryEventBus) Unsubscribe(eventType string, listener EventListener) {
	bus.mutex.Lock()
	defer bus.mutex.Unlock()

	listeners := bus.listeners[eventType]
	if listeners == nil {
		return
	}

	// Note: This is a simple implementation that removes all instances
	// In production, you might want a more sophisticated approach
	bus.listeners[eventType] = make([]EventListener, 0)
}
