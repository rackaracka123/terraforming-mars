package disconnect

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// DisconnectHandler handles player disconnection events
type DisconnectHandler struct {
	playerDisconnectedAction *action.PlayerDisconnectedAction
	parser                   *utils.MessageParser
	errorHandler             *utils.ErrorHandler
	logger                   *zap.Logger
}

// NewDisconnectHandler creates a new disconnect handler
func NewDisconnectHandler(playerDisconnectedAction *action.PlayerDisconnectedAction) *DisconnectHandler {
	return &DisconnectHandler{
		playerDisconnectedAction: playerDisconnectedAction,
		parser:                   utils.NewMessageParser(),
		errorHandler:             utils.NewErrorHandler(),
		logger:                   logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (dh *DisconnectHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	log := dh.logger.With(zap.String("handler", "disconnect"))

	// Parse the disconnect payload
	var disconnectPayload dto.PlayerDisconnectedPayload
	if err := dh.parser.ParsePayload(message.Payload, &disconnectPayload); err != nil {
		log.Error("Failed to parse player disconnected payload", zap.Error(err))
		dh.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return
	}

	log.Info("🔌 Processing player disconnection",
		zap.String("player_id", disconnectPayload.PlayerID),
		zap.String("game_id", disconnectPayload.GameID))

	// Execute the player disconnected action
	err := dh.playerDisconnectedAction.Execute(ctx, disconnectPayload.GameID, disconnectPayload.PlayerID)
	if err != nil {
		log.Error("Failed to process player disconnection",
			zap.String("player_id", disconnectPayload.PlayerID),
			zap.String("game_id", disconnectPayload.GameID),
			zap.Error(err))
		dh.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	log.Info("✅ Player disconnection processed successfully",
		zap.String("player_id", disconnectPayload.PlayerID),
		zap.String("game_id", disconnectPayload.GameID))
}
