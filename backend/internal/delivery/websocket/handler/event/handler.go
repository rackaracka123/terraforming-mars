package event

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// EventHandler handles domain events that need WebSocket broadcasting
type EventHandler struct {
	broadcaster     *core.Broadcaster
	cardDataService service.CardDataService
	logger          *zap.Logger
}

// NewEventHandler creates a new event handler
func NewEventHandler(broadcaster *core.Broadcaster, cardDataService service.CardDataService) *EventHandler {
	return &EventHandler{
		broadcaster:     broadcaster,
		cardDataService: cardDataService,
		logger:          logger.Get(),
	}
}

// HandlePlayerStartingCardOptions handles starting card options events
func (h *EventHandler) HandlePlayerStartingCardOptions(ctx context.Context, event events.Event) error {
	payload := event.GetPayload().(events.CardDealtEventData)
	gameID := payload.GameID
	playerID := payload.PlayerID
	cardIDs := payload.CardOptions

	h.logger.Info("🃏 Processing starting card options broadcast",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Strings("card_ids", cardIDs))

	// Get card details from card data service
	cardModels := make([]model.Card, 0, len(cardIDs))
	for _, cardID := range cardIDs {
		if card, err := h.cardDataService.GetCardByID(cardID); err == nil {
			cardModels = append(cardModels, *card)
		} else {
			h.logger.Warn("Card not found", zap.String("card_id", cardID), zap.Error(err))
		}
	}

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
