package standard_project

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// SellPatentsHandler handles sell patents standard project requests
type SellPatentsHandler struct {
	action      *action.SellPatentsAction
	broadcaster Broadcaster
	logger      *zap.Logger
}

// NewSellPatentsHandler creates a new sell patents handler
func NewSellPatentsHandler(action *action.SellPatentsAction, broadcaster Broadcaster) *SellPatentsHandler {
	return &SellPatentsHandler{
		action:      action,
		broadcaster: broadcaster,
		logger:      logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *SellPatentsHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🏛️ Processing sell patents request (migrated)")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute sell patents action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Sell patents action completed successfully")

	// Explicitly broadcast game state after action completes
	h.broadcaster.BroadcastGameState(connection.GameID, nil)
	log.Debug("📡 Broadcasted game state to all players")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "sell-patents",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *SellPatentsHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
