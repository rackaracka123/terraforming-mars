package turn_management

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// StartGameHandler handles start game requests
type StartGameHandler struct {
	action *action.StartGameAction
	logger *zap.Logger
}

// NewStartGameHandler creates a new start game handler
func NewStartGameHandler(action *action.StartGameAction) *StartGameHandler {
	return &StartGameHandler{
		action: action,
		logger: logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *StartGameHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("🚀 Processing start game request (migrated)")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context")
		h.sendError(connection, "Not connected to a game")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute start game action", zap.Error(err))
		h.sendError(connection, err.Error())
		return
	}

	log.Info("✅ Start game action completed successfully")

	response := dto.WebSocketMessage{
		Type:   "action-success",
		GameID: connection.GameID,
		Payload: map[string]interface{}{
			"action":  "start-game",
			"success": true,
		},
	}

	connection.Send <- response
}

func (h *StartGameHandler) sendError(connection *core.Connection, errorMessage string) {
	connection.Send <- dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: map[string]interface{}{
			"error": errorMessage,
		},
	}
}
