package connection

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// ReconnectFlow handles player reconnection logic
type ReconnectFlow struct {
	gameService   service.GameService
	playerService service.PlayerService
	broadcaster   *core.Broadcaster
	manager       *core.Manager
	parser        *utils.MessageParser
	errorHandler  *utils.ErrorHandler
	logger        *zap.Logger
}

// NewReconnectFlow creates a new reconnection flow handler
func NewReconnectFlow(
	gameService service.GameService,
	playerService service.PlayerService,
	broadcaster *core.Broadcaster,
	manager *core.Manager,
	parser *utils.MessageParser,
	errorHandler *utils.ErrorHandler,
	logger *zap.Logger,
) *ReconnectFlow {
	return &ReconnectFlow{
		gameService:   gameService,
		playerService: playerService,
		broadcaster:   broadcaster,
		manager:       manager,
		parser:        parser,
		errorHandler:  errorHandler,
		logger:        logger,
	}
}

// HandleReconnect handles explicit player reconnection requests
func (rf *ReconnectFlow) HandleReconnect(ctx context.Context, connection *core.Connection, payload dto.PlayerReconnectPayload) {
	rf.logger.Info("🔄 Player attempting to reconnect",
		zap.String("connection_id", connection.ID),
		zap.String("player_name", payload.PlayerName),
		zap.String("game_id", payload.GameID))

	// Find player by name
	player, err := rf.playerService.GetPlayerByName(ctx, payload.GameID, payload.PlayerName)
	if err != nil {
		rf.logger.Error("Player not found for reconnection", zap.Error(err))
		rf.errorHandler.SendError(connection, utils.ErrPlayerNotFound)
		return
	}

	// Use shared reconnection logic
	_, err = rf.executeReconnectionLogic(ctx, connection, payload.GameID, player.ID, payload.PlayerName)
	if err != nil {
		return
	}

	rf.logger.Info("📢 Player reconnected successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", player.ID),
		zap.String("player_name", player.Name))
}

// HandleExistingPlayerConnect handles when an existing player tries to connect
func (rf *ReconnectFlow) HandleExistingPlayerConnect(ctx context.Context, connection *core.Connection, payload dto.PlayerConnectPayload, playerID string) {
	rf.logger.Debug("🔄 Handling existing player reconnection",
		zap.String("player_id", playerID),
		zap.String("player_name", payload.PlayerName))

	// Setup connection for existing player
	connection.SetPlayer(playerID, payload.GameID)

	gameState, err := rf.executeReconnectionLogic(ctx, connection, payload.GameID, playerID, payload.PlayerName)
	if err != nil {
		return
	}

	// Send game updates to all players
	rf.broadcaster.SendPersonalizedGameUpdates(ctx, payload.GameID)

	rf.logger.Info("🎮 Existing player connected via WebSocket",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameState.ID),
		zap.String("player_name", payload.PlayerName))
}

// executeReconnectionLogic contains the shared logic for both connect and reconnect flows
func (rf *ReconnectFlow) executeReconnectionLogic(ctx context.Context, connection *core.Connection, gameID, playerID, playerName string) (*model.Game, error) {
	// Validate game exists
	game, err := rf.gameService.GetGame(ctx, gameID)
	if err != nil {
		rf.logger.Error("Failed to get game for reconnection", zap.Error(err))
		rf.errorHandler.SendError(connection, utils.ErrGameNotFound)
		return nil, err
	}

	// Clean up any existing connections for this player before adding new one
	existingConnection := rf.manager.RemoveExistingPlayerConnection(playerID, gameID)
	if existingConnection != nil {
		rf.logger.Info("🔄 Replaced existing connection for reconnecting player",
			zap.String("old_connection_id", existingConnection.ID),
			zap.String("new_connection_id", connection.ID),
			zap.String("player_id", playerID),
			zap.String("player_name", playerName))
	}

	// Associate connection with player
	connection.SetPlayer(playerID, gameID)
	rf.manager.AddToGame(connection, gameID)

	// Update player connection status
	err = rf.playerService.UpdatePlayerConnectionStatus(ctx, gameID, playerID, model.ConnectionStatusConnected)
	if err != nil {
		rf.logger.Error("Failed to update player connection status", zap.Error(err))
	}

	// Send reconnection confirmation and broadcast to others
	rf.sendReconnectionMessages(ctx, connection, &game, playerID, playerName, gameID)

	return &game, nil
}

// sendReconnectionMessages handles all messaging for player reconnections
func (rf *ReconnectFlow) sendReconnectionMessages(ctx context.Context, connection *core.Connection, game *model.Game, playerID, playerName, gameID string) {
	// Get all players for the game to create personalized view
	players, err := rf.playerService.GetPlayersForGame(ctx, game.ID)
	if err != nil {
		rf.logger.Error("❌ CRITICAL: Failed to get players for reconnection - this should not happen",
			zap.Error(err), zap.String("game_id", game.ID), zap.String("player_id", playerID))
		// Continue with basic DTO to avoid breaking the reconnection
		rf.sendBasicReconnectionMessage(connection, playerID, playerName, game)
		return
	}

	// Create personalized game DTO for the reconnecting player
	gameDTO := dto.ToGameDto(*game, players, playerID)

	// Send reconnection confirmation to the player
	rf.sendReconnectionConfirmation(connection, playerID, playerName, gameDTO, game.ID)

	// Broadcast reconnection to other players
	rf.broadcastPlayerReconnection(game, playerID, playerName, gameID, connection)
}

// sendBasicReconnectionMessage sends a basic reconnection message when player data can't be fetched
func (rf *ReconnectFlow) sendBasicReconnectionMessage(connection *core.Connection, playerID, playerName string, game *model.Game) {
	gameDTO := dto.ToGameDtoBasic(*game)
	reconnectedMessage := dto.WebSocketMessage{
		Type: dto.MessageTypePlayerReconnected,
		Payload: dto.PlayerReconnectedPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
			Game:       gameDTO,
		},
		GameID: game.ID,
	}
	rf.broadcaster.SendToConnection(connection, reconnectedMessage)
}

// sendReconnectionConfirmation sends full reconnection data to the reconnecting player
func (rf *ReconnectFlow) sendReconnectionConfirmation(connection *core.Connection, playerID, playerName string, gameDTO dto.GameDto, gameID string) {
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
	rf.broadcaster.SendToConnection(connection, reconnectedMessage)

	// Send current game update
	gameUpdateMessage := dto.WebSocketMessage{
		Type: dto.MessageTypeGameUpdated,
		Payload: dto.GameUpdatedPayload{
			Game: gameDTO,
		},
		GameID: gameID,
	}
	rf.broadcaster.SendToConnection(connection, gameUpdateMessage)
}

// broadcastPlayerReconnection notifies other players about the reconnection
func (rf *ReconnectFlow) broadcastPlayerReconnection(game *model.Game, playerID, playerName, gameID string, connection *core.Connection) {
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

	rf.broadcaster.BroadcastToGameExcept(gameID, reconnectedMessage, connection)
}