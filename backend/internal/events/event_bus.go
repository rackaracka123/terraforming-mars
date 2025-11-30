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

// BroadcastFunc is a callback function for automatic broadcasting
// Called after every event is published to notify clients of state changes
type BroadcastFunc func(gameID string, playerIDs []string)

// EventBusImpl implements EventBus with thread-safe operations
type EventBusImpl struct {
	subscriptions map[SubscriptionID]*subscription
	nextID        uint64
	mutex         sync.RWMutex
	logger        *zap.Logger
	gameID        string        // Game ID for automatic broadcasting
	broadcaster   BroadcastFunc // Callback for automatic broadcasting (optional)
}

// NewEventBus creates a new type-safe event bus
// gameID: The game this event bus belongs to (empty string for non-game event buses)
// broadcaster: Optional callback for automatic broadcasting (nil to disable)
func NewEventBus(gameID string, broadcaster BroadcastFunc) *EventBusImpl {
	return &EventBusImpl{
		subscriptions: make(map[SubscriptionID]*subscription),
		nextID:        1,
		logger:        logger.Get(),
		gameID:        gameID,
		broadcaster:   broadcaster,
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

// PlayerTargetable is an optional interface events can implement
// to specify which players should receive broadcast updates
type PlayerTargetable interface {
	GetPlayerIDs() []string // nil = broadcast to all players
}

// Publish publishes a type-safe event to all matching subscribers
// Automatically triggers broadcast if broadcaster is configured
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
	} else {
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

	// Automatic broadcasting: Call broadcaster after event handlers execute
	// Skip if this IS a BroadcastEvent (avoid infinite recursion)
	if eb.broadcaster != nil && eventType != "events.BroadcastEvent" {
		// Extract player IDs if event implements PlayerTargetable
		var playerIDs []string
		if targetable, ok := any(event).(PlayerTargetable); ok {
			playerIDs = targetable.GetPlayerIDs()
		}

		eb.logger.Debug("🔄 Auto-broadcasting after event",
			zap.String("event_type", eventType),
			zap.String("game_id", eb.gameID))

		eb.broadcaster(eb.gameID, playerIDs)
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

// Clear removes all subscriptions from the event bus
func (eb *EventBusImpl) Clear() {
	eb.mutex.Lock()
	defer eb.mutex.Unlock()
	eb.subscriptions = make(map[SubscriptionID]*subscription)
	eb.nextID = 1
}
