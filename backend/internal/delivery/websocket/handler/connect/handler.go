package connect

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// ConnectionHandler handles player connection and reconnection logic
type ConnectionHandler struct {
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
	gameService service.GameService,
	playerService service.PlayerService,
) *ConnectionHandler {
	return &ConnectionHandler{
		gameService:  gameService,
		playerService: playerService,
		parser:       utils.NewMessageParser(),
		errorHandler: utils.NewErrorHandler(),
		logger:       logger.Get(),
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
	game, err := ch.gameService.GetGame(ctx, payload.GameID)
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
		// Setup temporary connection for new players
		ch.setupTemporaryConnection(connCtx)
		return ch.processNewPlayer(connCtx)
	}

	// For reconnections, clean up existing connections BEFORE setting up the new one
	return ch.processReconnection(connCtx)
}

// processNewPlayer handles new player joining
func (ch *ConnectionHandler) processNewPlayer(connCtx *connectionContext) error {
	ch.logger.Debug("✨ Handling new player connection",
		zap.String("player_name", connCtx.payload.PlayerName))

	// Join game
	game, err := ch.gameService.JoinGame(connCtx.ctx, connCtx.payload.GameID, connCtx.payload.PlayerName)
	if err != nil {
		ch.logger.Error("Failed to join game via WebSocket",
			zap.Error(err))
		ch.errorHandler.SendError(connCtx.connection, utils.ErrConnectionFailed)
		return err
	}

	// Get the newly created player
	player, err := ch.playerService.GetPlayerByName(connCtx.ctx, connCtx.payload.GameID, connCtx.payload.PlayerName)
	if err != nil {
		ch.logger.Error("❌ Player not found in game after join",
			zap.Error(err))
		ch.errorHandler.SendError(connCtx.connection, "Player not found in game")
		return err
	}

	connCtx.game = &game
	connCtx.playerID = player.ID
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

// setupTemporaryConnection sets up initial connection state
func (ch *ConnectionHandler) setupTemporaryConnection(connCtx *connectionContext) {
	if connCtx.isNew {
		// New player - use temporary player ID
		tempPlayerID := "temp-" + connCtx.connection.ID
		connCtx.connection.SetPlayer(tempPlayerID, connCtx.payload.GameID)
		// Note: Connection is automatically managed by the Hub

		ch.logger.Debug("🔗 Connection set up for new player (temporary)",
			zap.String("connection_id", connCtx.connection.ID),
			zap.String("temp_player_id", tempPlayerID))
	} else {
		// Existing player - use real player ID
		connCtx.connection.SetPlayer(connCtx.playerID, connCtx.payload.GameID)
		// Note: Connection is automatically managed by the Hub

		ch.logger.Debug("🔗 Connection set up for existing player",
			zap.String("connection_id", connCtx.connection.ID),
			zap.String("player_id", connCtx.playerID))
	}
}


// finalizeConnection updates the connection with final player ID and sends state updates
func (ch *ConnectionHandler) finalizeConnection(connCtx *connectionContext) {
	// Update connection with final player ID
	connCtx.connection.SetPlayer(connCtx.playerID, connCtx.payload.GameID)

	// Send state updates
	ch.sendStateUpdates(connCtx)

	ch.logger.Info("🎮 Player connected via WebSocket",
		zap.String("connection_id", connCtx.connection.ID),
		zap.String("player_id", connCtx.playerID),
		zap.String("game_id", connCtx.game.ID),
		zap.String("player_name", connCtx.payload.PlayerName),
		zap.Bool("is_new_player", connCtx.isNew))
}

// sendStateUpdates sends all necessary state updates for the connection
func (ch *ConnectionHandler) sendStateUpdates(connCtx *connectionContext) {
	// Get personalized game state
	gameDTO := ch.getPersonalizedGameState(connCtx)

	// Send connection confirmation
	ch.sendConnectionConfirmation(connCtx, gameDTO)

	// Send game state update to the connecting player
	ch.sendGameStateUpdate(connCtx, gameDTO)

	// Let services handle broadcasting - they know when and what to broadcast
	ch.logger.Debug("Connection finalized - services will handle any needed broadcasting",
		zap.String("player_id", connCtx.playerID),
		zap.String("game_id", connCtx.game.ID))
}

// getPersonalizedGameState creates a personalized game DTO for the player
func (ch *ConnectionHandler) getPersonalizedGameState(connCtx *connectionContext) dto.GameDto {
	players, err := ch.playerService.GetPlayersForGame(connCtx.ctx, connCtx.game.ID)
	if err != nil {
		ch.logger.Error("❌ Failed to get players for personalized state",
			zap.Error(err),
			zap.String("game_id", connCtx.game.ID))
		// Return basic DTO as fallback
		return dto.ToGameDtoBasic(*connCtx.game)
	}

	return dto.ToGameDto(*connCtx.game, players, connCtx.playerID)
}

// sendConnectionConfirmation sends the connection/reconnection confirmation message
func (ch *ConnectionHandler) sendConnectionConfirmation(connCtx *connectionContext, gameDTO dto.GameDto) {
	messageType := ch.getConnectionMessageType(connCtx.isNew)

	var payload any
	if connCtx.isNew {
		payload = dto.PlayerConnectedPayload{
			PlayerID:   connCtx.playerID,
			PlayerName: connCtx.payload.PlayerName,
			Game:       gameDTO,
		}
	} else {
		// For reconnection, use PlayerReconnectedPayload
		payload = dto.PlayerReconnectedPayload{
			PlayerID:   connCtx.playerID,
			PlayerName: connCtx.payload.PlayerName,
			Game:       gameDTO,
		}
	}

	message := dto.WebSocketMessage{
		Type:    messageType,
		Payload: payload,
		GameID:  connCtx.game.ID,
	}

	connCtx.connection.SendMessage(message)
}

// sendGameStateUpdate sends a game-updated message to the connected player
func (ch *ConnectionHandler) sendGameStateUpdate(connCtx *connectionContext, gameDTO dto.GameDto) {
	message := dto.WebSocketMessage{
		Type: dto.MessageTypeGameUpdated,
		Payload: dto.GameUpdatedPayload{
			Game: gameDTO,
		},
		GameID: connCtx.game.ID,
	}

	connCtx.connection.SendMessage(message)
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

// getConnectionMessageType returns the appropriate message type based on connection type
func (ch *ConnectionHandler) getConnectionMessageType(isNew bool) dto.MessageType {
	if isNew {
		return dto.MessageTypePlayerConnected
	}
	return dto.MessageTypePlayerReconnected
}
