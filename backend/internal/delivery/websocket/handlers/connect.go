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

// ConnectionHandler handles player connection and reconnection logic
type ConnectionHandler struct {
	gameService   service.GameService
	playerService service.PlayerService
	broadcaster   *core.Broadcaster
	manager       *core.Manager
	parser        *utils.MessageParser
	errorHandler  *utils.ErrorHandler
	logger        *zap.Logger
}

// NewConnectionHandler creates a new connection handler
func NewConnectionHandler(
	gameService service.GameService,
	playerService service.PlayerService,
	broadcaster *core.Broadcaster,
	manager *core.Manager,
) *ConnectionHandler {
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

// HandleMessage implements the MessageHandler interface
func (ch *ConnectionHandler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	switch message.Type {
	case dto.MessageTypePlayerConnect:
		ch.handlePlayerConnect(ctx, connection, message)
	case dto.MessageTypePlayerReconnect:
		ch.handlePlayerReconnect(ctx, connection, message)
	}
}

// handlePlayerConnect handles player connection requests
func (ch *ConnectionHandler) handlePlayerConnect(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	ch.logger.Debug("🚪 Starting player connect handler", zap.String("connection_id", connection.ID))

	// Parse payload
	var payload dto.PlayerConnectPayload
	if err := ch.parser.ParsePayload(message.Payload, &payload); err != nil {
		ch.logger.Error("❌ Failed to parse player connect payload",
			zap.Error(err), zap.String("connection_id", connection.ID))
		ch.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return
	}

	ch.logger.Debug("✅ Payload parsed successfully",
		zap.String("game_id", payload.GameID),
		zap.String("player_name", payload.PlayerName))

	// Validate game exists
	if !ch.validateGameExists(ctx, connection, payload.GameID) {
		return
	}

	// Check if player already exists
	playerID := ch.findExistingPlayer(ctx, payload.GameID, payload.PlayerName)
	isNewPlayer := playerID == ""

	// Setup connection
	if !ch.setupConnection(connection, payload.GameID, playerID, isNewPlayer) {
		return
	}

	// Handle existing vs new player logic
	var game *model.Game
	var finalPlayerID string

	if !isNewPlayer {
		// Existing player reconnecting - use shared reconnection logic
		ch.logger.Debug("🔄 Handling existing player reconnection",
			zap.String("player_id", playerID),
			zap.String("player_name", payload.PlayerName))

		gameState, err := ch.handlePlayerReconnectionLogic(ctx, connection, payload.GameID, playerID, payload.PlayerName)
		if err != nil {
			return
		}

		game = gameState
		finalPlayerID = playerID
	} else {
		// New player joining - use existing join game logic
		gameState, newPlayerID := ch.joinGame(ctx, connection, payload)
		if gameState == nil || newPlayerID == "" {
			return
		}
		game = gameState
		finalPlayerID = newPlayerID
	}

	// Update connection with final player ID
	connection.SetPlayer(finalPlayerID, payload.GameID)

	// Send confirmations and updates
	ch.sendConnectionConfirmation(connection, finalPlayerID, payload.PlayerName, *game)

	// Send game-updated to all players in the game (including the new player)
	// This ensures all clients have the updated game state when any player connects/reconnects
	ch.broadcaster.SendPersonalizedGameUpdates(ctx, payload.GameID)

	ch.logger.Info("🎮 Player connected via WebSocket",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", finalPlayerID),
		zap.String("game_id", game.ID),
		zap.String("player_name", payload.PlayerName))
}

// handlePlayerReconnect handles player reconnection requests
func (ch *ConnectionHandler) handlePlayerReconnect(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	var payload dto.PlayerReconnectPayload
	if err := ch.parser.ParsePayload(message.Payload, &payload); err != nil {
		ch.logger.Error("Failed to parse player reconnect payload", zap.Error(err))
		ch.errorHandler.SendError(connection, utils.ErrInvalidPayload)
		return
	}

	ch.logger.Info("🔄 Player attempting to reconnect",
		zap.String("connection_id", connection.ID),
		zap.String("player_name", payload.PlayerName),
		zap.String("game_id", payload.GameID))

	// Find player by name
	player, err := ch.playerService.GetPlayerByName(ctx, payload.GameID, payload.PlayerName)
	if err != nil {
		ch.logger.Error("Player not found for reconnection", zap.Error(err))
		ch.errorHandler.SendError(connection, utils.ErrPlayerNotFound)
		return
	}

	// Use shared reconnection logic
	_, err = ch.handlePlayerReconnectionLogic(ctx, connection, payload.GameID, player.ID, payload.PlayerName)
	if err != nil {
		return
	}

	ch.logger.Info("📢 Player reconnected successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", player.ID),
		zap.String("player_name", player.Name))
}

// Helper methods

func (ch *ConnectionHandler) validateGameExists(ctx context.Context, connection *core.Connection, gameID string) bool {
	_, err := ch.gameService.GetGame(ctx, gameID)
	if err != nil {
		ch.logger.Error("Failed to get game for WebSocket connection", zap.Error(err))
		ch.errorHandler.SendError(connection, utils.ErrGameNotFound)
		return false
	}
	return true
}

func (ch *ConnectionHandler) findExistingPlayer(ctx context.Context, gameID, playerName string) string {
	// Use player service to find player by name
	player, err := ch.playerService.GetPlayerByName(ctx, gameID, playerName)
	if err != nil {
		return ""
	}
	return player.ID
}

func (ch *ConnectionHandler) setupConnection(connection *core.Connection, gameID, playerID string, isNewPlayer bool) bool {
	if isNewPlayer {
		// New player - use temporary player ID
		tempPlayerID := "temp-" + connection.ID
		connection.SetPlayer(tempPlayerID, gameID)
		ch.manager.AddToGame(connection, gameID)
		ch.logger.Debug("🔗 Connection set up for new player (temporary)",
			zap.String("connection_id", connection.ID),
			zap.String("temp_player_id", tempPlayerID))
	} else {
		// Existing player - use real player ID
		connection.SetPlayer(playerID, gameID)
		ch.manager.AddToGame(connection, gameID)
		ch.logger.Debug("🔗 Connection set up for existing player",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID))
	}
	return true
}

func (ch *ConnectionHandler) joinGame(ctx context.Context, connection *core.Connection, payload dto.PlayerConnectPayload) (*model.Game, string) {
	game, err := ch.gameService.JoinGame(ctx, payload.GameID, payload.PlayerName)
	if err != nil {
		ch.logger.Error("Failed to join game via WebSocket", zap.Error(err))
		ch.errorHandler.SendError(connection, utils.ErrConnectionFailed)
		return nil, ""
	}

	// Find the player ID using player service
	player, err := ch.playerService.GetPlayerByName(ctx, payload.GameID, payload.PlayerName)
	if err != nil {
		ch.logger.Error("❌ Player not found in game after join", zap.Error(err))
		ch.errorHandler.SendError(connection, "Player not found in game")
		return nil, ""
	}

	return &game, player.ID
}

func (ch *ConnectionHandler) sendConnectionConfirmation(connection *core.Connection, playerID, playerName string, game model.Game) {
	// Get all players for the game to create personalized view
	players, err := ch.playerService.GetPlayersForGame(context.Background(), game.ID)
	if err != nil {
		ch.logger.Error("❌ CRITICAL: Failed to get players for game - this should not happen",
			zap.Error(err), zap.String("game_id", game.ID), zap.String("player_id", playerID))
		// Continue with basic DTO to avoid breaking the connection
		gameDTO := dto.ToGameDtoBasic(game)
		playerConnectedMsg := dto.WebSocketMessage{
			Type: dto.MessageTypePlayerConnected,
			Payload: dto.PlayerConnectedPayload{
				PlayerID:   playerID,
				PlayerName: playerName,
				Game:       gameDTO,
			},
			GameID: game.ID,
		}
		ch.broadcaster.BroadcastToGame(game.ID, playerConnectedMsg)
		return
	}

	// Create personalized game DTO for the connecting player
	gameDTO := dto.ToGameDto(game, players, playerID)

	playerConnectedMsg := dto.WebSocketMessage{
		Type: dto.MessageTypePlayerConnected,
		Payload: dto.PlayerConnectedPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
			Game:       gameDTO,
		},
		GameID: game.ID,
	}

	// Send direct confirmation to the connecting player
	ch.broadcaster.SendToConnection(connection, playerConnectedMsg)

	// Also broadcast to other players in the game
	ch.broadcaster.BroadcastToGameExcept(game.ID, playerConnectedMsg, connection)
}

// handlePlayerReconnectionLogic contains the shared logic for both connect and reconnect flows
func (ch *ConnectionHandler) handlePlayerReconnectionLogic(ctx context.Context, connection *core.Connection, gameID, playerID, playerName string) (*model.Game, error) {
	// Validate game exists
	game, err := ch.gameService.GetGame(ctx, gameID)
	if err != nil {
		ch.logger.Error("Failed to get game for reconnection", zap.Error(err))
		ch.errorHandler.SendError(connection, utils.ErrGameNotFound)
		return nil, err
	}

	// Associate connection with player
	connection.SetPlayer(playerID, gameID)
	ch.manager.AddToGame(connection, gameID)

	// Update player connection status
	err = ch.playerService.UpdatePlayerConnectionStatus(ctx, gameID, playerID, model.ConnectionStatusConnected)
	if err != nil {
		ch.logger.Error("Failed to update player connection status", zap.Error(err))
	}

	// Send reconnection confirmation and broadcast to others
	ch.sendPlayerReconnectionMessages(ctx, connection, &game, playerID, playerName, gameID)

	return &game, nil
}

// sendPlayerReconnectionMessages handles all messaging for player reconnections
func (ch *ConnectionHandler) sendPlayerReconnectionMessages(ctx context.Context, connection *core.Connection, game *model.Game, playerID, playerName, gameID string) {
	// Get all players for the game to create personalized view
	players, err := ch.playerService.GetPlayersForGame(ctx, game.ID)
	if err != nil {
		ch.logger.Error("❌ CRITICAL: Failed to get players for reconnection - this should not happen",
			zap.Error(err), zap.String("game_id", game.ID), zap.String("player_id", playerID))
		// Continue with basic DTO to avoid breaking the reconnection
		gameDTO := dto.ToGameDtoBasic(*game)
		ch.sendBasicReconnectionMessage(connection, playerID, playerName, gameDTO, game.ID)
		return
	}

	// Create personalized game DTO for the reconnecting player
	gameDTO := dto.ToGameDto(*game, players, playerID)

	// Send reconnection confirmation to the player
	ch.sendReconnectionConfirmation(connection, playerID, playerName, gameDTO, game.ID)

	// Broadcast reconnection to other players
	ch.broadcastPlayerReconnection(game, playerID, playerName, gameID, connection)
}

// sendBasicReconnectionMessage sends a basic reconnection message when player data can't be fetched
func (ch *ConnectionHandler) sendBasicReconnectionMessage(connection *core.Connection, playerID, playerName string, gameDTO dto.GameDto, gameID string) {
	reconnectedMessage := dto.WebSocketMessage{
		Type: dto.MessageTypePlayerReconnected,
		Payload: dto.PlayerReconnectedPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
			Game:       gameDTO,
		},
		GameID: gameID,
	}
	ch.broadcaster.SendToConnection(connection, reconnectedMessage)
}

// sendReconnectionConfirmation sends full reconnection data to the reconnecting player
func (ch *ConnectionHandler) sendReconnectionConfirmation(connection *core.Connection, playerID, playerName string, gameDTO dto.GameDto, gameID string) {
	// Send reconnection confirmation
	reconnectedMessage := dto.WebSocketMessage{
		Type: dto.MessageTypePlayerReconnected,
		Payload: dto.PlayerReconnectedPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
			Game:       gameDTO,
		},
		GameID: gameID,
	}
	ch.broadcaster.SendToConnection(connection, reconnectedMessage)

	// Send current game update
	gameUpdateMessage := dto.WebSocketMessage{
		Type: dto.MessageTypeGameUpdated,
		Payload: dto.GameUpdatedPayload{
			Game: gameDTO,
		},
		GameID: gameID,
	}
	ch.broadcaster.SendToConnection(connection, gameUpdateMessage)
}

// broadcastPlayerReconnection notifies other players about the reconnection
func (ch *ConnectionHandler) broadcastPlayerReconnection(game *model.Game, playerID, playerName, gameID string, connection *core.Connection) {
	// For broadcasting to other players, use basic DTO since each client gets their own personalized view
	reconnectedPayload := dto.PlayerReconnectedPayload{
		PlayerID:   playerID,
		PlayerName: playerName,
		Game:       dto.ToGameDtoBasic(*game),
	}

	reconnectedMessage := dto.WebSocketMessage{
		Type:    dto.MessageTypePlayerReconnected,
		Payload: reconnectedPayload,
		GameID:  gameID,
	}

	ch.broadcaster.BroadcastToGameExcept(gameID, reconnectedMessage, connection)
}
