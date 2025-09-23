package handlers

import (
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/service"
)

// EventHandler handles domain events for WebSocket broadcasting
type EventHandler struct {
	broadcaster *core.Broadcaster
	cardService service.CardService
}

// NewEventHandler creates a new event handler
func NewEventHandler(broadcaster *core.Broadcaster, cardService service.CardService) *EventHandler {
	return &EventHandler{
		broadcaster: broadcaster,
		cardService: cardService,
	}
}