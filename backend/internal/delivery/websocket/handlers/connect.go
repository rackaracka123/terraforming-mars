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

	// Join/rejoin game
	game, finalPlayerID := ch.joinGame(ctx, connection, payload, isNewPlayer)
	if game == nil || finalPlayerID == "" {
		return
	}

	// Update connection with final player ID
	connection.SetPlayer(finalPlayerID, payload.GameID)

	// Send confirmations and updates
	ch.sendConnectionConfirmation(connection, finalPlayerID, payload.PlayerName, *game)
	
	if isNewPlayer {
		ch.sendPersonalizedUpdate(ctx, connection, finalPlayerID, payload.GameID)
	}

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

	// Validate game exists
	game, err := ch.gameService.GetGame(ctx, payload.GameID)
	if err != nil {
		ch.logger.Error("Failed to get game for reconnection", zap.Error(err))
		ch.errorHandler.SendError(connection, utils.ErrGameNotFound)
		return
	}

	// Find player by name
	player, err := ch.playerService.GetPlayerByName(ctx, payload.GameID, payload.PlayerName)
	if err != nil {
		ch.logger.Error("Player not found for reconnection", zap.Error(err))
		ch.errorHandler.SendError(connection, utils.ErrPlayerNotFound)
		return
	}

	// Handle existing connections
	ch.handleExistingConnection(player.ID, payload.GameID)

	// Associate connection with player
	connection.SetPlayer(player.ID, payload.GameID)
	ch.manager.AddToGame(connection, payload.GameID)

	// Update player connection status
	err = ch.playerService.UpdatePlayerConnectionStatus(ctx, payload.GameID, player.ID, model.ConnectionStatusConnected)
	if err != nil {
		ch.logger.Error("Failed to update player connection status", zap.Error(err))
	}

	// Send comprehensive reconnection data
	ch.sendReconnectionData(ctx, connection, &game, &player)

	// Notify other players
	ch.broadcastPlayerReconnection(ctx, &game, &player, payload.GameID, connection)

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
	currentGame, _ := ch.gameService.GetGame(ctx, gameID)
	for _, player := range currentGame.Players {
		if player.Name == playerName {
			return player.ID
		}
	}
	return ""
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

func (ch *ConnectionHandler) joinGame(ctx context.Context, connection *core.Connection, payload dto.PlayerConnectPayload, isNewPlayer bool) (*model.Game, string) {
	game, err := ch.gameService.JoinGame(ctx, payload.GameID, payload.PlayerName)
	if err != nil {
		ch.logger.Error("Failed to join game via WebSocket", zap.Error(err))
		ch.errorHandler.SendError(connection, utils.ErrConnectionFailed)
		return nil, ""
	}

	// Find the player ID in the updated game
	for _, player := range game.Players {
		if player.Name == payload.PlayerName {
			return &game, player.ID
		}
	}

	ch.logger.Error("❌ Player not found in game after join")
	ch.errorHandler.SendError(connection, "Player not found in game")
	return nil, ""
}

func (ch *ConnectionHandler) sendConnectionConfirmation(connection *core.Connection, playerID, playerName string, game model.Game) {
	gameDTO := dto.ToGameDto(game)
	
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
}

func (ch *ConnectionHandler) sendPersonalizedUpdate(ctx context.Context, connection *core.Connection, playerID, gameID string) {
	playerGame, err := ch.gameService.GetGameForPlayer(ctx, gameID, playerID)
	if err != nil {
		ch.logger.Error("Failed to get personalized game state for new player", zap.Error(err))
		return
	}

	personalizedGameDTO := dto.ToGameDto(playerGame)
	gameUpdateMessage := dto.WebSocketMessage{
		Type: dto.MessageTypeGameUpdated,
		Payload: dto.GameUpdatedPayload{
			Game: personalizedGameDTO,
		},
	}
	
	ch.broadcaster.SendToConnection(connection, gameUpdateMessage)
}

func (ch *ConnectionHandler) handleExistingConnection(playerID, gameID string) {
	// This will be handled by the manager's FindConnectionByPlayer method
	// If there's an existing connection, it will be closed appropriately
}

func (ch *ConnectionHandler) sendReconnectionData(ctx context.Context, connection *core.Connection, game *model.Game, player *model.Player) {
	gameDTO := dto.ToGameDto(*game)

	// Send primary reconnection confirmation
	primaryMessage := dto.WebSocketMessage{
		Type: dto.MessageTypePlayerConnected,
		Payload: dto.PlayerConnectedPayload{
			PlayerID:   player.ID,
			PlayerName: player.Name,
			Game:       gameDTO,
		},
		GameID: game.ID,
	}

	ch.broadcaster.SendToConnection(connection, primaryMessage)

	// Send current game update
	gameUpdateMessage := dto.WebSocketMessage{
		Type: dto.MessageTypeGameUpdated,
		Payload: dto.GameUpdatedPayload{
			Game: gameDTO,
		},
		GameID: game.ID,
	}

	ch.broadcaster.SendToConnection(connection, gameUpdateMessage)
}

func (ch *ConnectionHandler) broadcastPlayerReconnection(ctx context.Context, game *model.Game, player *model.Player, gameID string, connection *core.Connection) {
	reconnectedPayload := dto.PlayerReconnectedPayload{
		PlayerID:   player.ID,
		PlayerName: player.Name,
		Game:       dto.ToGameDto(*game),
	}

	reconnectedMessage := dto.WebSocketMessage{
		Type:    dto.MessageTypePlayerReconnected,
		Payload: reconnectedPayload,
		GameID:  gameID,
	}

	ch.broadcaster.BroadcastToGameExcept(gameID, reconnectedMessage, connection)
}