package disconnect

import (
	"context"

	"terraforming-mars-backend/internal/actions"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// DisconnectHandler handles player disconnection events
// Delegates all business logic to DisconnectPlayerAction
type DisconnectHandler struct {
	disconnectAction *actions.DisconnectPlayerAction
	parser           *utils.MessageParser
	errorHandler     *utils.ErrorHandler
	logger           *zap.Logger
}

// NewDisconnectHandler creates a new disconnect handler
func NewDisconnectHandler(disconnectAction *actions.DisconnectPlayerAction) *DisconnectHandler {
	return &DisconnectHandler{
		disconnectAction: disconnectAction,
		parser:           utils.NewMessageParser(),
		errorHandler:     utils.NewErrorHandler(),
		logger:           logger.Get(),
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

	// Execute the disconnect action (handles all business logic)
	if err := dh.disconnectAction.Execute(ctx, disconnectPayload.GameID, disconnectPayload.PlayerID); err != nil {
		log.Error("Failed to execute disconnect player action",
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
