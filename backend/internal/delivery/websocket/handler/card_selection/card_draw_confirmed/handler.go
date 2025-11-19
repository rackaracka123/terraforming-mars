package card_draw_confirmed

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles card draw confirmation requests
type Handler struct {
	confirmCardDrawAction *action.ConfirmCardDrawAction
	parser                *utils.MessageParser
}

// NewHandler creates a new card draw confirmation handler
func NewHandler(confirmCardDrawAction *action.ConfirmCardDrawAction, parser *utils.MessageParser) *Handler {
	return &Handler{
		confirmCardDrawAction: confirmCardDrawAction,
		parser:                parser,
	}
}

// CardDrawConfirmedRequest represents the payload for card draw confirmation
type CardDrawConfirmedRequest struct {
	CardsToTake []string `json:"cardsToTake" ts:"string[]"` // Card IDs to take for free
	CardsToBuy  []string `json:"cardsToBuy" ts:"string[]"`  // Card IDs to buy (pay for each)
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := logger.Get().With(
		zap.String("connection_id", connection.ID),
		zap.String("player_id", connection.PlayerID),
		zap.String("game_id", connection.GameID))

	log.Debug("🃏 Processing card draw confirmation")

	// Parse the action payload
	var request CardDrawConfirmedRequest
	if err := h.parser.ParsePayload(message.Payload, &request); err != nil {
		log.Error("Failed to parse card draw confirmation payload", zap.Error(err))
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": "invalid payload format"},
		}
		return
	}

	// Execute the confirm card draw action
	err := h.confirmCardDrawAction.Execute(ctx, connection.GameID, connection.PlayerID, request.CardsToTake, request.CardsToBuy)
	if err != nil {
		log.Error("Failed to execute confirm card draw action", zap.Error(err))
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": err.Error()},
		}
		return
	}

	log.Info("✅ Card draw confirmation completed successfully",
		zap.Int("cards_taken", len(request.CardsToTake)),
		zap.Int("cards_bought", len(request.CardsToBuy)))
}
