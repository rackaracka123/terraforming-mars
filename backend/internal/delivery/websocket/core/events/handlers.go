package events

import (
	"context"

	"terraforming-mars-backend/internal/delivery/websocket/core/broadcast"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handlers contains all event handlers for the WebSocket hub
type Handlers struct {
	broadcaster  *broadcast.Broadcaster
	eventHandler EventHandler
	logger       *zap.Logger
}

// NewHandlers creates a new handlers instance
func NewHandlers(broadcaster *broadcast.Broadcaster, eventHandler EventHandler) *Handlers {
	return &Handlers{
		broadcaster:  broadcaster,
		eventHandler: eventHandler,
		logger:       logger.Get(),
	}
}

// HandleGameUpdated processes game updated events
func (h *Handlers) HandleGameUpdated(ctx context.Context, event events.Event) error {
	payload := event.GetPayload().(events.GameUpdatedEventData)
	gameID := payload.GameID

	h.logger.Info("🎮 Processing game updated broadcast",
		zap.String("game_id", gameID))

	// Delegate to broadcaster
	h.broadcaster.SendPersonalizedGameUpdates(ctx, gameID)

	h.logger.Info("✅ Game updated broadcast completed", zap.String("game_id", gameID))
	return nil
}

// HandlePlayerStartingCardOptions handles card option events
func (h *Handlers) HandlePlayerStartingCardOptions(ctx context.Context, event events.Event) error {
	h.logger.Debug("🃏 Card options event received - delegating to event handler")

	// Delegate to the proper event handler
	if h.eventHandler != nil {
		return h.eventHandler.HandlePlayerStartingCardOptions(ctx, event)
	}

	h.logger.Warn("⚠️ No event handler configured")
	return nil
}

// HandleGlobalParameterChange handles global parameter changes (temperature, oceans, etc.)
func (h *Handlers) HandleGlobalParameterChange(ctx context.Context, event events.Event) error {
	// Extract game ID from the event payload
	var gameID string

	// Handle consolidated global parameter event
	switch event.GetType() {
	case events.EventTypeGlobalParametersChanged:
		payload := event.GetPayload().(events.GlobalParametersChangedEventData)
		gameID = payload.GameID
		h.logger.Debug("🌍 Processing global parameters change event",
			zap.String("game_id", gameID),
			zap.Strings("change_types", payload.ChangeTypes))
	default:
		h.logger.Warn("⚠️ Unknown global parameter event type", zap.String("event_type", event.GetType()))
		return nil
	}

	// Trigger game update broadcast to notify clients of parameter changes
	h.broadcaster.SendPersonalizedGameUpdates(ctx, gameID)

	h.logger.Debug("✅ Global parameter change broadcast completed", zap.String("game_id", gameID))
	return nil
}