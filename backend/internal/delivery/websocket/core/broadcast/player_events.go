package broadcast

import (
	"context"
	"strings"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// PlayerEvents handles player connection/disconnection events
type PlayerEvents struct {
	manager       *core.Manager
	gameService   service.GameService
	playerService service.PlayerService
	logger        *zap.Logger
}

// NewPlayerEvents creates a new player events broadcaster
func NewPlayerEvents(manager *core.Manager, gameService service.GameService, playerService service.PlayerService, logger *zap.Logger) *PlayerEvents {
	return &PlayerEvents{
		manager:       manager,
		gameService:   gameService,
		playerService: playerService,
		logger:        logger,
	}
}

// BroadcastPlayerDisconnection handles player disconnection broadcasting
func (pe *PlayerEvents) BroadcastPlayerDisconnection(ctx context.Context, playerID, gameID string, connection *core.Connection) {
	// Get game info for the broadcast
	game, err := pe.gameService.GetGame(ctx, gameID)
	if err != nil {
		pe.logger.Error("Failed to get game for player disconnection broadcast",
			zap.String("game_id", gameID),
			zap.Error(err))
		return
	}

	// Get player info using player service
	player, err := pe.playerService.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		pe.logger.Error("Failed to get player for disconnection broadcast",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Error(err))
		return
	}
	playerName := player.Name

	// Get all players for personalized messages
	allPlayers := pe.getAllPlayersForGame(ctx, game, gameID)

	// Send personalized disconnection messages to remaining players
	pe.sendDisconnectionMessages(ctx, gameID, playerID, playerName, game, allPlayers, connection)

	pe.logger.Info("📢 Player disconnected, broadcasted to other players in game",
		zap.String("player_id", playerID),
		zap.String("player_name", playerName),
		zap.String("game_id", gameID))
}

// getAllPlayersForGame retrieves all players for a game
func (pe *PlayerEvents) getAllPlayersForGame(ctx context.Context, game model.Game, gameID string) []model.Player {
	var allPlayers []model.Player
	for _, pID := range game.PlayerIDs {
		p, err := pe.playerService.GetPlayer(ctx, gameID, pID)
		if err != nil {
			pe.logger.Warn("⚠️ Failed to get player for disconnection broadcast",
				zap.String("game_id", gameID),
				zap.String("player_id", pID),
				zap.Error(err))
			continue
		}
		allPlayers = append(allPlayers, p)
	}
	return allPlayers
}

// sendDisconnectionMessages sends personalized disconnection messages to remaining players
func (pe *PlayerEvents) sendDisconnectionMessages(ctx context.Context, gameID, playerID, playerName string, game model.Game, allPlayers []model.Player, disconnectedConnection *core.Connection) {
	gameConns := pe.manager.GetGameConnections(gameID)
	if gameConns == nil {
		return
	}

	sentCount := 0
	for conn := range gameConns {
		if conn == disconnectedConnection { // Skip the disconnected player
			continue
		}

		connPlayerID, _ := conn.GetPlayer()
		if connPlayerID == "" || strings.HasPrefix(connPlayerID, "temp-") {
			continue
		}

		// Create personalized disconnection payload for this player
		personalizedGame := dto.ToGameDto(game, allPlayers, connPlayerID)
		disconnectedPayload := dto.PlayerDisconnectedPayload{
			PlayerID:   playerID,
			PlayerName: playerName,
			Game:       personalizedGame,
		}

		disconnectedMessage := dto.WebSocketMessage{
			Type:    dto.MessageTypePlayerDisconnected,
			Payload: disconnectedPayload,
			GameID:  gameID,
		}

		conn.SendMessage(disconnectedMessage)
		sentCount++
	}

	pe.logger.Debug("📢 Sent disconnection messages to remaining players",
		zap.String("game_id", gameID),
		zap.String("disconnected_player_id", playerID),
		zap.Int("messages_sent", sentCount))
}