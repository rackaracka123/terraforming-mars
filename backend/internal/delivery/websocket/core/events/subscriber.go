package events

import (
	"context"

	"terraforming-mars-backend/internal/delivery/websocket/core/broadcast"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Subscriber manages event subscriptions for WebSocket hub
type Subscriber struct {
	eventBus    events.EventBus
	broadcaster *broadcast.Broadcaster
	handlers    *Handlers
	logger      *zap.Logger
}

// NewSubscriber creates a new event subscriber
func NewSubscriber(eventBus events.EventBus, broadcaster *broadcast.Broadcaster, eventHandler EventHandler) *Subscriber {
	return &Subscriber{
		eventBus:    eventBus,
		broadcaster: broadcaster,
		handlers:    NewHandlers(broadcaster, eventHandler),
		logger:      logger.Get(),
	}
}

// EventHandler interface for handling domain events
type EventHandler interface {
	HandlePlayerStartingCardOptions(ctx context.Context, event events.Event) error
}

// SubscribeToEvents sets up event listeners
func (s *Subscriber) SubscribeToEvents() {
	// Subscribe to game updates for broadcasting updates
	s.eventBus.Subscribe(events.EventTypeGameUpdated, s.handlers.HandleGameUpdated)

	// Subscribe to card events (using new consolidated event names)
	s.eventBus.Subscribe(events.EventTypeCardDealt, s.handlers.HandlePlayerStartingCardOptions) // Renamed from PlayerStartingCardOptions

	// Subscribe to global parameter changes to trigger game updates (consolidated event only)
	s.eventBus.Subscribe(events.EventTypeGlobalParametersChanged, s.handlers.HandleGlobalParameterChange)

	s.logger.Info("📡 WebSocket hub subscribed to events")
}