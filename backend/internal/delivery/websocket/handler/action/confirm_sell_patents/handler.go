package confirm_sell_patents

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles confirm sell patents card selection requests (Phase 2)
type Handler struct {
	confirmSellPatentsAction *action.ConfirmSellPatentsAction
}

// NewHandler creates a new confirm sell patents handler
func NewHandler(confirmSellPatentsAction *action.ConfirmSellPatentsAction) *Handler {
	return &Handler{
		confirmSellPatentsAction: confirmSellPatentsAction,
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := logger.Get().With(
		zap.String("connection_id", connection.ID),
		zap.String("player_id", connection.PlayerID),
		zap.String("game_id", connection.GameID))

	log.Debug("🏛️ Processing confirm sell patents action")

	// Extract selected card IDs from message payload
	payload, ok := message.Payload.(map[string]interface{})
	if !ok {
		log.Error("Invalid payload format")
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": "invalid payload format"},
		}
		return
	}

	selectedCardsInterface, ok := payload["selectedCards"]
	if !ok {
		log.Error("Missing selectedCards in payload")
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": "missing selectedCards"},
		}
		return
	}

	// Convert interface{} array to []string
	selectedCardsArray, ok := selectedCardsInterface.([]interface{})
	if !ok {
		log.Error("Invalid selectedCards format")
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": "invalid selectedCards format"},
		}
		return
	}

	selectedCards := make([]string, len(selectedCardsArray))
	for i, cardInterface := range selectedCardsArray {
		cardID, ok := cardInterface.(string)
		if !ok {
			log.Error("Invalid card ID format in selectedCards")
			connection.Send <- dto.WebSocketMessage{
				Type:    dto.MessageTypeError,
				Payload: map[string]interface{}{"error": "invalid card ID format"},
			}
			return
		}
		selectedCards[i] = cardID
	}

	log.Debug("Confirming sell patents with selected cards", zap.Int("card_count", len(selectedCards)))

	// Execute the confirm sell patents action (Phase 2: process selection)
	err := h.confirmSellPatentsAction.Execute(ctx, connection.GameID, connection.PlayerID, selectedCards)
	if err != nil {
		log.Error("Failed to execute confirm sell patents action", zap.Error(err))
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": err.Error()},
		}
		return
	}

	log.Info("✅ Confirm sell patents action completed successfully", zap.Int("cards_sold", len(selectedCards)))
}
