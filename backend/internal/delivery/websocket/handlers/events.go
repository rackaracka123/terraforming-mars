package handlers

import (
	"context"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// EventHandler handles domain events that need WebSocket broadcasting
type EventHandler struct {
	broadcaster *core.Broadcaster
	logger      *zap.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(broadcaster *core.Broadcaster) *EventHandler {
	return &EventHandler{
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandlePlayerStartingCardOptions handles starting card options events
func (h *EventHandler) HandlePlayerStartingCardOptions(ctx context.Context, event events.Event) error {
	payload := event.GetPayload().(events.PlayerStartingCardOptionsEventData)
	gameID := payload.GameID
	playerID := payload.PlayerID
	cardIDs := payload.CardOptions

	h.logger.Info("🃏 Processing starting card options broadcast",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Strings("card_ids", cardIDs))

	// Get card details from mock database
	cardModels := cards.GetCardsByIDs(cardIDs)
	
	// Convert to DTOs
	cardDtos := make([]dto.CardDto, len(cardModels))
	for i, card := range cardModels {
		cardDtos[i] = dto.ToCardDto(card)
	}

	// Send available-cards message to the specific player
	h.broadcaster.SendAvailableCardsToPlayer(ctx, gameID, playerID, cardDtos)

	h.logger.Info("✅ Starting card options broadcast completed",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Int("cards_sent", len(cardDtos)))
	return nil
}