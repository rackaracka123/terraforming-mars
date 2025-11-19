package sell_patents

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Handler handles sell patents standard project action requests
type Handler struct {
	sellPatentsAction *action.SellPatentsAction
}

// NewHandler creates a new sell patents handler
func NewHandler(sellPatentsAction *action.SellPatentsAction) *Handler {
	return &Handler{
		sellPatentsAction: sellPatentsAction,
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := logger.Get().With(
		zap.String("connection_id", connection.ID),
		zap.String("player_id", connection.PlayerID),
		zap.String("game_id", connection.GameID))

	log.Debug("🏛️ Processing sell patents action")

	// Execute the sell patents action (Phase 1: initiate card selection)
	err := h.sellPatentsAction.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute sell patents action", zap.Error(err))
		connection.Send <- dto.WebSocketMessage{
			Type:    dto.MessageTypeError,
			Payload: map[string]interface{}{"error": err.Error()},
		}
		return
	}

	log.Info("✅ Sell patents action completed successfully - awaiting card selection")
}
