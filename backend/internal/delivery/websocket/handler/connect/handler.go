package connect

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/internal/session"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConnectionHandler handles player connection and reconnection logic
type ConnectionHandler struct {
	hub            *core.Hub
	sessionManager session.SessionManager
	gameService    service.GameService
	playerService  service.PlayerService
	joinGameAction *action.JoinGameAction
	parser         *utils.MessageParser
	errorHandler   *utils.ErrorHandler
	logger         *zap.Logger
}

// connectionContext holds the context for a connection operation
type connectionContext struct {
	ctx        context.Context
	connection *core.Connection
	payload    dto.PlayerConnectPayload
	playerID   string
	isNew      bool
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(
	hub *core.Hub,
	sessionManager session.SessionManager,
	gameService service.GameService,
	playerService service.PlayerService,
	joinGameAction *action.JoinGameAction,
) *ConnectionHandler {
	return &ConnectionHandler{
		hub:            hub,
		sessionManager: sessionManager,
		gameService:    gameService,
		playerService:  playerService,
		joinGameAction: joinGameAction,
		parser:         utils.NewMessageParser(),
		errorHandler:   utils.NewErrorHandler(),
		logger:         logger.Get(),
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

	// Parse and validate the connection request
	connCtx, err := ch.parseAndValidate(ctx, connection, message)
	if err != nil {
		return // Error already logged and sent to client
	}

	// Process the connection (new or reconnection)
	if err := ch.processConnection(connCtx); err != nil {
		return // Error already handled
	}

	// For new players, finalize the connection and send state updates
	// For reconnections, the service handles all state sending
	if connCtx.isNew {
		ch.finalizeConnection(connCtx)
	}
}

// parseAndValidate parses the payload and validates the connection request
func (ch *ConnectionHandler) parseAndValidate(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) (*connectionContext, error) {
	// Parse payload
	var payload dto.PlayerConnectPayload
	if err := ch.parser.ParsePayload(message.Payload, &payload); err != nil {
		ch.logger.Error("❌ Failed to parse player connect payload",
			zap.Error(err),
			zap.String("connection_id", connection.ID))
		ch.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return nil, err
	}

	ch.logger.Debug("✅ Payload parsed successfully",
		zap.String("game_id", payload.GameID),
		zap.String("player_name", payload.PlayerName))

	// Note: Game and player validation is handled by actions
	// No need to check existence here - let the action layer handle business logic

	// Determine if this is a new connection or reconnection based on playerID presence
	isNew := payload.PlayerID == ""
	playerID := payload.PlayerID

	if isNew {
		ch.logger.Debug("🆕 New player connection detected")
	} else {
		ch.logger.Debug("🔄 Reconnection detected", zap.String("player_id", playerID))
	}

	return &connectionContext{
		ctx:        ctx,
		connection: connection,
		payload:    payload,
		playerID:   playerID,
		isNew:      isNew,
	}, nil
}

// processConnection handles the connection setup for new or existing players
func (ch *ConnectionHandler) processConnection(connCtx *connectionContext) error {
	if connCtx.isNew {
		// For new players, don't set up connection until we have the real player ID
		// This prevents the temporary ID from interfering with broadcasts
		return ch.processNewPlayer(connCtx)
	}

	// For reconnections, clean up existing connections BEFORE setting up the new one
	return ch.processReconnection(connCtx)
}

// processNewPlayer handles new player joining
func (ch *ConnectionHandler) processNewPlayer(connCtx *connectionContext) error {
	ch.logger.Debug("✨ Handling new player connection",
		zap.String("player_name", connCtx.payload.PlayerName))

	// CRITICAL: Generate playerID upfront and set up connection BEFORE calling action
	// This ensures the connection is registered when PlayerJoinedEvent broadcasts
	playerID := uuid.New().String()

	// Set up connection FIRST (before action triggers PlayerJoinedEvent)
	connCtx.connection.SetPlayer(playerID, connCtx.payload.GameID)

	// CRITICAL: Register connection with Hub's Manager so broadcasts can find it
	// This MUST happen before JoinGameAction executes, which publishes PlayerJoinedEvent
	ch.hub.RegisterConnectionWithGame(connCtx.connection, connCtx.payload.GameID)

	ch.logger.Debug("🔗 Connection set up and registered with game",
		zap.String("connection_id", connCtx.connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", connCtx.payload.GameID))

	// Now join game - broadcasts will work because connection is already registered
	returnedPlayerID, err := ch.joinGameAction.Execute(connCtx.ctx, connCtx.payload.GameID, connCtx.payload.PlayerName, playerID)
	if err != nil {
		ch.logger.Error("Failed to join game via action",
			zap.Error(err))
		ch.errorHandler.SendError(connCtx.connection, utils.ErrConnectionFailed)
		return err
	}

	ch.logger.Debug("✅ Player joined game via action",
		zap.String("player_id", returnedPlayerID),
		zap.String("player_name", connCtx.payload.PlayerName))

	connCtx.playerID = returnedPlayerID

	// Broadcast game state to all players now that join is complete
	// NOTE: PlayerJoinedEvent subscription removed from SessionManager to prevent race condition
	// Connection handler now controls broadcast timing after all setup is complete
	ch.logger.Debug("📡 Broadcasting game state after player join",
		zap.String("game_id", connCtx.payload.GameID),
		zap.String("player_id", returnedPlayerID))

	err = ch.sessionManager.Broadcast(connCtx.payload.GameID)
	if err != nil {
		ch.logger.Error("Failed to broadcast game state after join",
			zap.Error(err),
			zap.String("game_id", connCtx.payload.GameID))
		// Non-fatal - player joined successfully, broadcast can be retried
	}

	return nil
}

// processReconnection handles existing player reconnection
func (ch *ConnectionHandler) processReconnection(connCtx *connectionContext) error {
	ch.logger.Debug("🔄 Handling existing player reconnection",
		zap.String("player_id", connCtx.playerID),
		zap.String("player_name", connCtx.payload.PlayerName))

	// Set up the current connection with the real player ID (not temporary)
	connCtx.connection.SetPlayer(connCtx.playerID, connCtx.payload.GameID)

	// Register connection with Hub's Manager for reconnections as well
	ch.hub.RegisterConnectionWithGame(connCtx.connection, connCtx.payload.GameID)

	ch.logger.Debug("🔗 Connection set up and registered for reconnecting player",
		zap.String("connection_id", connCtx.connection.ID),
		zap.String("player_id", connCtx.playerID),
		zap.String("game_id", connCtx.payload.GameID))

	// Let the service handle the complete reconnection process including:
	// - Updating connection status
	// - Retrieving complete game state
	// - Sending personalized state to the player
	err := ch.gameService.PlayerReconnected(connCtx.ctx, connCtx.payload.GameID, connCtx.playerID)
	if err != nil {
		ch.logger.Error("Failed to process player reconnection via service",
			zap.Error(err))
		return err
	}

	ch.logger.Info("✅ Player reconnection completed via service",
		zap.String("player_id", connCtx.playerID),
		zap.String("player_name", connCtx.payload.PlayerName),
		zap.String("game_id", connCtx.payload.GameID))

	return nil
}

// finalizeConnection logs connection completion (connection already updated with player ID)
func (ch *ConnectionHandler) finalizeConnection(connCtx *connectionContext) {
	// Connection has already been updated with the real player ID in processNewPlayer
	// Services have already handled broadcasting the game state during join/reconnect
	ch.logger.Info("🎮 Player connected via WebSocket",
		zap.String("connection_id", connCtx.connection.ID),
		zap.String("player_id", connCtx.playerID),
		zap.String("game_id", connCtx.payload.GameID),
		zap.String("player_name", connCtx.payload.PlayerName),
		zap.Bool("is_new_player", connCtx.isNew))
}
