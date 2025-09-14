package event

import (
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// EventHandler handles domain events that need WebSocket broadcasting
type EventHandler struct {
	broadcaster *core.Broadcaster
	cardService service.CardService
	logger      *zap.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(broadcaster *core.Broadcaster, cardService service.CardService) *EventHandler {
	return &EventHandler{
		broadcaster: broadcaster,
		cardService: cardService,
		logger:      logger.Get(),
	}
}
