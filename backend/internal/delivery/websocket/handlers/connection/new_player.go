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

// NewPlayerFlow handles new player joining logic
type NewPlayerFlow struct {
	gameService   service.GameService
	playerService service.PlayerService
	broadcaster   *core.Broadcaster
	parser        *utils.MessageParser
	errorHandler  *utils.ErrorHandler
	logger        *zap.Logger
}

// NewNewPlayerFlow creates a new player flow handler
func NewNewPlayerFlow(
	gameService service.GameService,
	playerService service.PlayerService,
	broadcaster *core.Broadcaster,
	parser *utils.MessageParser,
	errorHandler *utils.ErrorHandler,
	logger *zap.Logger,
) *NewPlayerFlow {
	return &NewPlayerFlow{
		gameService:   gameService,
		playerService: playerService,
		broadcaster:   broadcaster,
		parser:        parser,
		errorHandler:  errorHandler,
		logger:        logger,
	}
}

// HandleNewPlayer handles the complete flow for a new player joining
func (npf *NewPlayerFlow) HandleNewPlayer(ctx context.Context, connection *core.Connection, payload dto.PlayerConnectPayload) {
	// Setup temporary connection
	tempPlayerID := "temp-" + connection.ID
	connection.SetPlayer(tempPlayerID, payload.GameID)

	// Join the game
	game, playerID := npf.joinGame(ctx, connection, payload)
	if game == nil || playerID == "" {
		return
	}

	// Update connection with real player ID
	connection.SetPlayer(playerID, payload.GameID)

	// Send confirmations and updates
	npf.sendConnectionConfirmation(connection, playerID, payload.PlayerName, *game)

	// Broadcast game updates
	npf.broadcaster.SendPersonalizedGameUpdates(ctx, payload.GameID)

	npf.logger.Info("🎮 New player connected via WebSocket",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", game.ID),
		zap.String("player_name", payload.PlayerName))
}

// joinGame handles the game joining logic
func (npf *NewPlayerFlow) joinGame(ctx context.Context, connection *core.Connection, payload dto.PlayerConnectPayload) (*model.Game, string) {
	game, err := npf.gameService.JoinGame(ctx, payload.GameID, payload.PlayerName)
	if err != nil {
		npf.logger.Error("Failed to join game via WebSocket", zap.Error(err))
		npf.errorHandler.SendError(connection, utils.ErrConnectionFailed)
		return nil, ""
	}

	// Find the player ID using player service
	player, err := npf.playerService.GetPlayerByName(ctx, payload.GameID, payload.PlayerName)
	if err != nil {
		npf.logger.Error("❌ Player not found in game after join", zap.Error(err))
		npf.errorHandler.SendError(connection, "Player not found in game")
		return nil, ""
	}

	return &game, player.ID
}

// sendConnectionConfirmation sends confirmation to the new player
func (npf *NewPlayerFlow) sendConnectionConfirmation(connection *core.Connection, playerID, playerName string, game model.Game) {
	// Get all players for the game to create personalized view
	players, err := npf.playerService.GetPlayersForGame(context.Background(), game.ID)
	if err != nil {
		npf.logger.Error("❌ CRITICAL: Failed to get players for game - this should not happen",
			zap.Error(err), zap.String("game_id", game.ID), zap.String("player_id", playerID))
		// Continue with basic DTO to avoid breaking the connection
		npf.sendBasicConnectionMessage(connection, playerID, playerName, game)
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
	npf.broadcaster.SendToConnection(connection, playerConnectedMsg)

	// Also broadcast to other players in the game
	npf.broadcaster.BroadcastToGameExcept(game.ID, playerConnectedMsg, connection)
}

// sendBasicConnectionMessage sends a basic message when player data can't be fetched
func (npf *NewPlayerFlow) sendBasicConnectionMessage(_ *core.Connection, playerID, playerName string, game model.Game) {
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
	npf.broadcaster.BroadcastToGame(game.ID, playerConnectedMsg)
}
