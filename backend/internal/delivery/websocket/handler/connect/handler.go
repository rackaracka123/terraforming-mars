package connect

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ConnectionHandler handles player connection and reconnection logic
type ConnectionHandler struct {
	lobbyService  lobby.Service
	gameService   service.GameService
	playerService service.PlayerService
	parser        *utils.MessageParser
	errorHandler  *utils.ErrorHandler
	logger        *zap.Logger
}

// connectionContext holds the context for a connection operation
type connectionContext struct {
	ctx        context.Context
	connection *core.Connection
	payload    dto.PlayerConnectPayload
	playerID   string
	game       *model.Game
	isNew      bool
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(
	lobbyService lobby.Service,
	gameService service.GameService,
	playerService service.PlayerService,
) *ConnectionHandler {
	return &ConnectionHandler{
		lobbyService:  lobbyService,
		gameService:   gameService,
		playerService: playerService,
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

	// Validate game exists
	game, err := ch.lobbyService.GetGame(ctx, payload.GameID)
	if err != nil {
		ch.logger.Error("Failed to get game for WebSocket connection",
			zap.Error(err))
		ch.errorHandler.SendError(connection, utils.ErrGameNotFound)
		return nil, err
	}

	// Check if player already exists - prioritize ID for reconnection
	var playerID string
	if payload.PlayerID != "" {
		// Try to find by ID first (for reconnection)
		playerID = ch.findExistingPlayerByID(ctx, payload.GameID, payload.PlayerID)
		// If player found by ID, this is a reconnection
		if playerID != "" {
			ch.logger.Debug("🔄 Found player by ID - treating as reconnection",
				zap.String("player_id", playerID),
				zap.String("player_name", payload.PlayerName))
		}
	}

	// Only check by name if no playerID was provided or found
	// This prevents creating duplicate players when reconnecting
	if playerID == "" && payload.PlayerID == "" {
		// Fall back to finding by name only for new connections
		playerID = ch.findExistingPlayer(ctx, payload.GameID, payload.PlayerName)
	}

	return &connectionContext{
		ctx:        ctx,
		connection: connection,
		payload:    payload,
		playerID:   playerID,
		game:       &game,
		isNew:      playerID == "",
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

	// First, create a new player ID that will be used consistently
	playerID := uuid.New().String()

	// CRITICAL: Set up connection with real player ID BEFORE calling JoinGame
	// This ensures the connection is properly registered when the service broadcasts
	connCtx.connection.SetPlayer(playerID, connCtx.payload.GameID)

	ch.logger.Debug("🔗 Connection set up with pre-generated player ID",
		zap.String("connection_id", connCtx.connection.ID),
		zap.String("player_id", playerID))

	// Join game using the pre-generated player ID
	game, err := ch.lobbyService.JoinGameWithPlayerID(connCtx.ctx, connCtx.payload.GameID, connCtx.payload.PlayerName, playerID)
	if err != nil {
		ch.logger.Error("Failed to join game via WebSocket",
			zap.Error(err))
		ch.errorHandler.SendError(connCtx.connection, utils.ErrConnectionFailed)
		return err
	}

	ch.logger.Debug("✅ Player joined game with pre-registered connection",
		zap.String("connection_id", connCtx.connection.ID),
		zap.String("player_id", playerID))

	connCtx.game = &game
	connCtx.playerID = playerID
	return nil
}

// processReconnection handles existing player reconnection
func (ch *ConnectionHandler) processReconnection(connCtx *connectionContext) error {
	ch.logger.Debug("🔄 Handling existing player reconnection",
		zap.String("player_id", connCtx.playerID),
		zap.String("player_name", connCtx.payload.PlayerName))

	// Set up the current connection with the real player ID (not temporary)
	connCtx.connection.SetPlayer(connCtx.playerID, connCtx.payload.GameID)

	ch.logger.Debug("🔗 Connection set up for reconnecting player",
		zap.String("connection_id", connCtx.connection.ID),
		zap.String("player_id", connCtx.playerID))

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
		zap.String("game_id", connCtx.game.ID),
		zap.String("player_name", connCtx.payload.PlayerName),
		zap.Bool("is_new_player", connCtx.isNew))
}

// Helper methods

// findExistingPlayer checks if a player with the given name exists in the game
func (ch *ConnectionHandler) findExistingPlayer(ctx context.Context, gameID, playerName string) string {
	player, err := ch.playerService.GetPlayerByName(ctx, gameID, playerName)
	if err != nil {
		return ""
	}
	return player.ID
}

// findExistingPlayerByID checks if a player with the given ID exists in the game
func (ch *ConnectionHandler) findExistingPlayerByID(ctx context.Context, gameID, playerID string) string {
	player, err := ch.playerService.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		ch.logger.Debug("Player not found by ID",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Error(err))
		return ""
	}
	ch.logger.Debug("Found existing player by ID",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("player_name", player.Name))
	return player.ID
}
