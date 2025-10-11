package events

import (
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// SubscriptionID represents a unique subscription identifier
type SubscriptionID string

// EventHandler is a type-safe event handler function
type EventHandler[T any] func(event T)

// subscription wraps a handler with its type information
type subscription struct {
	id          SubscriptionID
	handler     interface{}     // The actual typed handler
	eventType   string          // Type name for matching
	handlerFunc func(event any) // Type-erased execution wrapper
}

// EventBusImpl implements EventBus with thread-safe operations
type EventBusImpl struct {
	subscriptions map[SubscriptionID]*subscription
	nextID        uint64
	mutex         sync.RWMutex
	logger        *zap.Logger
}

// NewEventBus creates a new type-safe event bus
func NewEventBus() *EventBusImpl {
	return &EventBusImpl{
		subscriptions: make(map[SubscriptionID]*subscription),
		nextID:        1,
		logger:        logger.Get(),
	}
}

// Subscribe registers a type-safe event handler
func Subscribe[T any](eb *EventBusImpl, handler EventHandler[T]) SubscriptionID {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	// Generate unique subscription ID
	id := SubscriptionID(fmt.Sprintf("sub-%d", eb.nextID))
	eb.nextID++

	// Get type name for matching
	var zero T
	eventType := fmt.Sprintf("%T", zero)

	// Create type-erased wrapper that calls the typed handler
	handlerFunc := func(event any) {
		if typedEvent, ok := event.(T); ok {
			handler(typedEvent)
		}
	}

	// Store subscription
	sub := &subscription{
		id:          id,
		handler:     handler,
		eventType:   eventType,
		handlerFunc: handlerFunc,
	}

	eb.subscriptions[id] = sub

	eb.logger.Debug("📬 Event handler subscribed",
		zap.String("subscription_id", string(id)),
		zap.String("event_type", eventType))

	return id
}

// Publish publishes a type-safe event to all matching subscribers
func Publish[T any](eb *EventBusImpl, event T) {
	eb.mutex.RLock()
	defer eb.mutex.RUnlock()

	// Get event type
	eventType := fmt.Sprintf("%T", event)

	// Find all matching subscriptions
	var matchingHandlers []func(any)
	for _, sub := range eb.subscriptions {
		if sub.eventType == eventType {
			matchingHandlers = append(matchingHandlers, sub.handlerFunc)
		}
	}

	if len(matchingHandlers) == 0 {
		eb.logger.Debug("📭 No subscribers for event",
			zap.String("event_type", eventType))
		return
	}

	eb.logger.Debug("📢 Publishing event to subscribers",
		zap.String("event_type", eventType),
		zap.Int("subscriber_count", len(matchingHandlers)))

	// Execute all matching handlers
	// Note: Handlers are executed synchronously for now
	// Future optimization: execute asynchronously with goroutines
	for _, handlerFunc := range matchingHandlers {
		handlerFunc(event)
	}
}

// Unsubscribe removes a subscription by ID
func (eb *EventBusImpl) Unsubscribe(id SubscriptionID) {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()

	if sub, exists := eb.subscriptions[id]; exists {
		delete(eb.subscriptions, id)
		eb.logger.Debug("🗑️ Event handler unsubscribed",
			zap.String("subscription_id", string(id)),
			zap.String("event_type", sub.eventType))
	}
}
