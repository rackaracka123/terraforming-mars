package handlers

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

// ConnectionHandler handles WebSocket connection and event management
type ConnectionHandler struct {
	gameService   service.GameService
	playerService service.PlayerService
	broadcaster   *core.Broadcaster
	manager       *core.Manager
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
func NewConnectionHandler(gameService service.GameService, playerService service.PlayerService, broadcaster *core.Broadcaster, manager *core.Manager) *ConnectionHandler {
	return &ConnectionHandler{
		gameService:   gameService,
		playerService: playerService,
		broadcaster:   broadcaster,
		manager:       manager,
		parser:        utils.NewMessageParser(),
		errorHandler:  utils.NewErrorHandler(),
		logger:        logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface with connection routing
func (h *ConnectionHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	switch message.Type {
	case dto.MessageTypePlayerConnect:
		h.handleConnection(ctx, connection, message)
	default:
		h.logger.Warn("Unknown connection message type", zap.String("type", string(message.Type)))
		h.errorHandler.SendError(connection, "Unknown connection message type")
	}
}

// handleConnection handles player connection requests (both new connections and reconnections)
func (h *ConnectionHandler) handleConnection(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	h.logger.Debug("🚪 Starting player connection handler", zap.String("connection_id", connection.ID))

	// Parse and validate the connection request
	connCtx, err := h.parseAndValidate(ctx, connection, message)
	if err != nil {
		return // Error already logged and sent to client
	}

	// Process the connection (new or reconnection)
	if err := h.processConnection(connCtx); err != nil {
		return // Error already handled
	}

	// Finalize and broadcast the connection
	h.finalizeConnection(connCtx)
}

// parseAndValidate parses the payload and validates the connection request
func (h *ConnectionHandler) parseAndValidate(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) (*connectionContext, error) {
	// Parse payload
	var payload dto.PlayerConnectPayload
	if err := h.parser.ParsePayload(message.Payload, &payload); err != nil {
		h.logger.Error("❌ Failed to parse player connect payload",
			zap.Error(err),
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return nil, err
	}

	h.logger.Debug("✅ Payload parsed successfully",
		zap.String("game_id", payload.GameID),
		zap.String("player_name", payload.PlayerName))

	// Validate game exists
	game, err := h.gameService.GetGame(ctx, payload.GameID)
	if err != nil {
		h.logger.Error("Failed to get game for WebSocket connection", zap.Error(err))
		h.errorHandler.SendError(connection, utils.ErrGameNotFound)
		return nil, err
	}

	// Check if player already exists - prioritize ID for reconnection
	var playerID string
	if payload.PlayerID != "" {
		// Try to find by ID first (for reconnection)
		playerID = h.findExistingPlayerByID(ctx, payload.GameID, payload.PlayerID)
		// If player found by ID, this is a reconnection
		if playerID != "" {
			h.logger.Debug("🔄 Found player by ID - treating as reconnection",
				zap.String("player_id", playerID),
				zap.String("player_name", payload.PlayerName))
		}
	}

	// Only check by name if no playerID was provided or found
	// This prevents creating duplicate players when reconnecting
	if playerID == "" && payload.PlayerID == "" {
		// Fall back to finding by name only for new connections
		playerID = h.findExistingPlayer(ctx, payload.GameID, payload.PlayerName)
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
func (h *ConnectionHandler) processConnection(connCtx *connectionContext) error {
	if connCtx.isNew {
		// Setup temporary connection for new players
		h.setupTemporaryConnection(connCtx)
		return h.processNewPlayer(connCtx)
	}

	// For reconnections, clean up existing connections BEFORE setting up the new one
	return h.processReconnection(connCtx)
}

// processNewPlayer handles new player joining
func (h *ConnectionHandler) processNewPlayer(connCtx *connectionContext) error {
	h.logger.Debug("✨ Handling new player connection", zap.String("player_name", connCtx.payload.PlayerName))

	// Join game
	game, err := h.gameService.JoinGame(connCtx.ctx, connCtx.payload.GameID, connCtx.payload.PlayerName)
	if err != nil {
		h.logger.Error("Failed to join game via WebSocket", zap.Error(err))
		h.errorHandler.SendError(connCtx.connection, utils.ErrConnectionFailed)
		return err
	}

	// Get the newly created player
	player, err := h.playerService.GetPlayerByName(connCtx.ctx, connCtx.payload.GameID, connCtx.payload.PlayerName)
	if err != nil {
		h.logger.Error("❌ Player not found in game after join", zap.Error(err))
		h.errorHandler.SendError(connCtx.connection, "Player not found in game")
		return err
	}

	connCtx.game = &game
	connCtx.playerID = player.ID
	return nil
}

// processReconnection handles existing player reconnection
func (h *ConnectionHandler) processReconnection(connCtx *connectionContext) error {
	h.logger.Debug("🔄 Handling existing player reconnection",
		zap.String("player_id", connCtx.playerID),
		zap.String("player_name", connCtx.payload.PlayerName))

	// Clean up any existing connections for this player BEFORE setting up the new one
	h.cleanupExistingConnection(connCtx)

	// Set up the current connection with the real player ID (not temporary)
	connCtx.connection.SetPlayer(connCtx.playerID, connCtx.payload.GameID)
	h.manager.AddToGame(connCtx.connection, connCtx.payload.GameID)

	h.logger.Debug("🔗 Connection set up for reconnecting player",
		zap.String("connection_id", connCtx.connection.ID),
		zap.String("player_id", connCtx.playerID))

	// Update player connection status BEFORE getting game state
	// This ensures the player is marked as connected in the database
	err := h.playerService.UpdatePlayerConnectionStatus(
		connCtx.ctx,
		connCtx.payload.GameID,
		connCtx.playerID,
		true,
	)
	if err != nil {
		h.logger.Error("Failed to update player connection status", zap.Error(err))
		// Non-critical error, continue
	}

	// Get the latest game state for reconnection
	// This should now include the reconnected player with correct status
	game, err := h.gameService.GetGame(connCtx.ctx, connCtx.payload.GameID)
	if err != nil {
		h.logger.Error("Failed to get game state for reconnection", zap.Error(err))
		return err
	}
	connCtx.game = &game

	h.logger.Info("✅ Player reconnection processed successfully",
		zap.String("player_id", connCtx.playerID),
		zap.String("player_name", connCtx.payload.PlayerName),
		zap.String("game_id", connCtx.game.ID))

	return nil
}

// setupTemporaryConnection sets up initial connection state
func (h *ConnectionHandler) setupTemporaryConnection(connCtx *connectionContext) {
	if connCtx.isNew {
		// New player - use temporary player ID
		tempPlayerID := "temp-" + connCtx.connection.ID
		connCtx.connection.SetPlayer(tempPlayerID, connCtx.payload.GameID)
		h.manager.AddToGame(connCtx.connection, connCtx.payload.GameID)

		h.logger.Debug("🔗 Connection set up for new player (temporary)",
			zap.String("connection_id", connCtx.connection.ID),
			zap.String("temp_player_id", tempPlayerID))
	} else {
		// Existing player - use real player ID
		connCtx.connection.SetPlayer(connCtx.playerID, connCtx.payload.GameID)
		h.manager.AddToGame(connCtx.connection, connCtx.payload.GameID)

		h.logger.Debug("🔗 Connection set up for existing player",
			zap.String("connection_id", connCtx.connection.ID),
			zap.String("player_id", connCtx.playerID))
	}
}

// cleanupExistingConnection removes any existing connection for the player
func (h *ConnectionHandler) cleanupExistingConnection(connCtx *connectionContext) {
	existingConnection := h.manager.RemoveExistingPlayerConnection(
		connCtx.playerID,
		connCtx.payload.GameID,
		connCtx.connection, // Pass current connection to avoid cleaning it up
	)

	if existingConnection != nil {
		h.logger.Info("🔄 Replaced existing connection for reconnecting player",
			zap.String("old_connection_id", existingConnection.ID),
			zap.String("new_connection_id", connCtx.connection.ID),
			zap.String("player_id", connCtx.playerID),
			zap.String("player_name", connCtx.payload.PlayerName))
	}
}

// finalizeConnection updates the connection with final player ID and sends state updates
func (h *ConnectionHandler) finalizeConnection(connCtx *connectionContext) {
	// Update connection with final player ID
	connCtx.connection.SetPlayer(connCtx.playerID, connCtx.payload.GameID)
	h.manager.AddToGame(connCtx.connection, connCtx.payload.GameID)

	// Send state updates
	h.sendStateUpdates(connCtx)

	h.logger.Info("🎮 Player connected via WebSocket",
		zap.String("connection_id", connCtx.connection.ID),
		zap.String("player_id", connCtx.playerID),
		zap.String("game_id", connCtx.game.ID),
		zap.String("player_name", connCtx.payload.PlayerName),
		zap.Bool("is_new_player", connCtx.isNew))
}

// sendStateUpdates sends all necessary state updates for the connection
func (h *ConnectionHandler) sendStateUpdates(connCtx *connectionContext) {
	// Get personalized game state
	gameDTO := h.getPersonalizedGameState(connCtx)

	// Send connection confirmation
	h.sendConnectionConfirmation(connCtx, gameDTO)

	// Send game state update to the connecting player
	h.sendGameStateUpdate(connCtx, gameDTO)

	// For reconnections, we need to ensure all players get the updated state
	if !connCtx.isNew {
		h.logger.Info("🔄 Broadcasting game state to all players after reconnection",
			zap.String("player_id", connCtx.playerID),
			zap.String("game_id", connCtx.game.ID))

		// Trigger personalized updates for ALL players including the reconnected one
		// This ensures everyone has the correct player list
		h.broadcaster.SendPersonalizedGameUpdates(connCtx.ctx, connCtx.game.ID)
	} else {
		// For new connections, just notify other players
		h.broadcastConnectionEvent(connCtx)

		// Then send personalized updates to all
		h.broadcaster.SendPersonalizedGameUpdates(connCtx.ctx, connCtx.game.ID)
	}
}

// getPersonalizedGameState creates a personalized game DTO for the player
func (h *ConnectionHandler) getPersonalizedGameState(connCtx *connectionContext) dto.GameDto {
	players, err := h.playerService.GetPlayersForGame(connCtx.ctx, connCtx.game.ID)
	if err != nil {
		h.logger.Error("❌ Failed to get players for personalized state",
			zap.Error(err),
			zap.String("game_id", connCtx.game.ID))
		// Return basic DTO as fallback
		return dto.ToGameDtoBasic(*connCtx.game)
	}

	return dto.ToGameDto(*connCtx.game, players, connCtx.playerID)
}

// sendConnectionConfirmation sends the connection/reconnection confirmation message
func (h *ConnectionHandler) sendConnectionConfirmation(connCtx *connectionContext, gameDTO dto.GameDto) {
	messageType := h.getConnectionMessageType(connCtx.isNew)

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

	h.broadcaster.SendToConnection(connCtx.connection, message)
}

// sendGameStateUpdate sends a game-updated message to the connected player
func (h *ConnectionHandler) sendGameStateUpdate(connCtx *connectionContext, gameDTO dto.GameDto) {
	message := dto.WebSocketMessage{
		Type: dto.MessageTypeGameUpdated,
		Payload: dto.GameUpdatedPayload{
			Game: gameDTO,
		},
		GameID: connCtx.game.ID,
	}

	h.broadcaster.SendToConnection(connCtx.connection, message)
}

// broadcastConnectionEvent notifies other players about the connection
func (h *ConnectionHandler) broadcastConnectionEvent(connCtx *connectionContext) {
	messageType := h.getConnectionMessageType(connCtx.isNew)

	var payload any
	if connCtx.isNew {
		payload = dto.PlayerConnectedPayload{
			PlayerID:   connCtx.playerID,
			PlayerName: connCtx.payload.PlayerName,
			Game:       dto.ToGameDtoBasic(*connCtx.game),
		}
	} else {
		// For reconnection, use PlayerReconnectedPayload
		payload = dto.PlayerReconnectedPayload{
			PlayerID:   connCtx.playerID,
			PlayerName: connCtx.payload.PlayerName,
			Game:       dto.ToGameDtoBasic(*connCtx.game),
		}
	}

	message := dto.WebSocketMessage{
		Type:    messageType,
		Payload: payload,
		GameID:  connCtx.game.ID,
	}

	h.broadcaster.BroadcastToGameExcept(connCtx.game.ID, message, connCtx.connection)
}

// Helper methods

// findExistingPlayer checks if a player with the given name exists in the game
func (h *ConnectionHandler) findExistingPlayer(ctx context.Context, gameID, playerName string) string {
	player, err := h.playerService.GetPlayerByName(ctx, gameID, playerName)
	if err != nil {
		return ""
	}
	return player.ID
}

// findExistingPlayerByID checks if a player with the given ID exists in the game
func (h *ConnectionHandler) findExistingPlayerByID(ctx context.Context, gameID, playerID string) string {
	player, err := h.playerService.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		h.logger.Debug("Player not found by ID",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Error(err))
		return ""
	}
	h.logger.Debug("Found existing player by ID",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("player_name", player.Name))
	return player.ID
}

// getConnectionMessageType returns the appropriate message type based on connection type
func (h *ConnectionHandler) getConnectionMessageType(isNew bool) dto.MessageType {
	if isNew {
		return dto.MessageTypePlayerConnected
	}
	return dto.MessageTypePlayerReconnected
}