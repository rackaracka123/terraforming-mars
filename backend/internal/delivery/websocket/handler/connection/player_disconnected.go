package connection

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// PlayerDisconnectedHandler handles player disconnection requests
type PlayerDisconnectedHandler struct {
	action *action.PlayerDisconnectedAction
	logger *zap.Logger
}

// NewPlayerDisconnectedHandler creates a new player disconnected handler
func NewPlayerDisconnectedHandler(action *action.PlayerDisconnectedAction) *PlayerDisconnectedHandler {
	return &PlayerDisconnectedHandler{
		action: action,
		logger: logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *PlayerDisconnectedHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := h.logger.With(
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)),
	)

	log.Info("⛓️‍💥 Processing player disconnected request (migrated)")

	if connection.GameID == "" || connection.PlayerID == "" {
		log.Error("Missing connection context - connection closing anyway")
		return
	}

	err := h.action.Execute(ctx, connection.GameID, connection.PlayerID)
	if err != nil {
		log.Error("Failed to execute player disconnected action (connection closing anyway)", zap.Error(err))
		return
	}

	log.Info("✅ Player disconnected action completed successfully")

	// NOTE: Do NOT send response on connection.Send - the connection is being closed
	// The BroadcastEvent from PlayerDisconnectedAction will notify other players
}
