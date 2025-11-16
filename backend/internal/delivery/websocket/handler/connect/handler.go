package connect

import (
	"context"

	"terraforming-mars-backend/internal/actions"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ConnectionHandler handles player connection and reconnection requests
// Delegates all business logic to ConnectPlayerAction
type ConnectionHandler struct {
	connectAction *actions.ConnectPlayerAction
	parser        *utils.MessageParser
	errorHandler  *utils.ErrorHandler
	logger        *zap.Logger
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(
	connectAction *actions.ConnectPlayerAction,
) *ConnectionHandler {
	return &ConnectionHandler{
		connectAction: connectAction,
		parser:        utils.NewMessageParser(),
		errorHandler:  utils.NewErrorHandler(),
		logger:        logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (ch *ConnectionHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	switch message.Type {
	case dto.MessageTypePlayerConnect:
		ch.handleConnection(ctx, connection, message)
	}
}

// handleConnection handles player connection requests (both new connections and reconnections)
func (ch *ConnectionHandler) handleConnection(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	ch.logger.Debug("🚪 Starting player connection handler",
		zap.String("connection_id", connection.ID))

	// Parse payload
	var payload dto.PlayerConnectPayload
	if err := ch.parser.ParsePayload(message.Payload, &payload); err != nil {
		ch.logger.Error("❌ Failed to parse player connect payload",
			zap.Error(err),
			zap.String("connection_id", connection.ID))
		ch.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return
	}

	ch.logger.Debug("✅ Payload parsed successfully",
		zap.String("game_id", payload.GameID),
		zap.String("player_name", payload.PlayerName))

	// Execute the connect action (handles all business logic)
	result, err := ch.connectAction.Execute(ctx, payload.GameID, payload.PlayerName, payload.PlayerID)
	if err != nil {
		ch.logger.Error("Failed to execute connect player action", zap.Error(err))
		ch.errorHandler.SendError(connection, utils.ErrConnectionFailed)
		return
	}

	// CRITICAL: Set up the connection with the player ID BEFORE sending confirmation
	// This ensures the Hub can route messages to this player
	connection.SetPlayer(result.PlayerID, result.GameID)

	ch.logger.Debug("🔗 Connection registered with player ID",
		zap.String("player_id", result.PlayerID),
		zap.String("game_id", result.GameID))

	// Now send connection confirmation - the connection is ready to receive messages
	err = ch.connectAction.SendConfirmation(ctx, result)
	if err != nil {
		ch.logger.Error("Failed to send connection confirmation", zap.Error(err))
		// Don't fail - connection is established, just log the error
	}

	ch.logger.Info("🎮 Player connected via WebSocket",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", result.PlayerID),
		zap.String("game_id", result.GameID),
		zap.String("player_name", payload.PlayerName),
		zap.Bool("is_new_player", result.IsNew))
}
