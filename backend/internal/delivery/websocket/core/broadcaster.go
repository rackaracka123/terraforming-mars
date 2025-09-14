package core

import (
	"context"
	"strings"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// Broadcaster handles sending messages to WebSocket connections
type Broadcaster struct {
	manager       *Manager
	gameService   service.GameService
	playerService service.PlayerService
	logger        *zap.Logger
}

// NewBroadcaster creates a new message broadcaster
func NewBroadcaster(manager *Manager, gameService service.GameService, playerService service.PlayerService) *Broadcaster {
	return &Broadcaster{
		manager:       manager,
		gameService:   gameService,
		playerService: playerService,
		logger:        logger.Get(),
	}
}

// BroadcastToGame sends a message to all connections in a game
func (b *Broadcaster) BroadcastToGame(gameID string, message dto.WebSocketMessage) {
	gameConns := b.manager.GetGameConnections(gameID)

	if gameConns == nil || len(gameConns) == 0 {
		b.logger.Warn("❌ No connections found for game", zap.String("game_id", gameID))
		return
	}

	sentCount := 0
	for connection := range gameConns {
		playerID, _ := connection.GetPlayer()
		b.logger.Debug("📤 Sending message to individual connection",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID),
			zap.String("message_type", string(message.Type)))

		connection.SendMessage(message)
		sentCount++
	}

	b.logger.Info("📢 Server broadcasted to game clients",
		zap.String("game_id", gameID),
		zap.String("message_type", string(message.Type)),
		zap.Int("messages_sent", sentCount))
}

// BroadcastToGameExcept sends a message to all connections in a game except the excluded connection
func (b *Broadcaster) BroadcastToGameExcept(gameID string, message dto.WebSocketMessage, excludeConnection *Connection) {
	gameConns := b.manager.GetGameConnections(gameID)

	if gameConns == nil {
		return
	}

	sentCount := 0
	for connection := range gameConns {
		if connection != excludeConnection {
			connection.SendMessage(message)
			sentCount++
		}
	}

	b.logger.Debug("📢 Server broadcasting to game clients (excluding one)",
		zap.String("game_id", gameID),
		zap.String("message_type", string(message.Type)),
		zap.Int("sent_to_count", sentCount))
}

// SendToConnection sends a message to a specific connection
func (b *Broadcaster) SendToConnection(connection *Connection, message dto.WebSocketMessage) {
	connection.SendMessage(message)

	b.logger.Debug("💬 Server message sent to client",
		zap.String("connection_id", connection.ID),
		zap.String("message_type", string(message.Type)))
}

// SendPersonalizedGameUpdates sends personalized game-updated messages to all connected players
func (b *Broadcaster) SendPersonalizedGameUpdates(ctx context.Context, gameID string) {
	b.logger.Debug("🔍 Getting connected players for personalized broadcast", zap.String("game_id", gameID))

	gameConns := b.manager.GetGameConnections(gameID)
	if gameConns == nil {
		b.logger.Debug("No connections found for game", zap.String("game_id", gameID))
		return
	}

	b.logConnectionState(gameConns, gameID)

	sentCount := 0
	connectionsWithoutPlayerID := 0

	for connection := range gameConns {
		if ctx.Err() != nil {
			b.logger.Warn("Context cancelled during personalized game updates",
				zap.String("game_id", gameID),
				zap.Error(ctx.Err()))
			return
		}

		playerID, validConnection := b.validateConnection(connection, gameID)
		if !validConnection {
			connectionsWithoutPlayerID++
			continue
		}

		if b.sendPersonalizedMessage(ctx, connection, playerID, gameID) {
			sentCount++
		}
	}

	b.logger.Info("📢 Sent personalized game-updated messages to players",
		zap.String("game_id", gameID),
		zap.Int("total_connections", len(gameConns)),
		zap.Int("messages_sent", sentCount),
		zap.Int("connections_without_player_id", connectionsWithoutPlayerID))
}

// logConnectionState logs the current state of connections for debugging
func (b *Broadcaster) logConnectionState(gameConns map[*Connection]bool, gameID string) {
	connectionList := make([]string, 0, len(gameConns))
	playerIDList := make([]string, 0, len(gameConns))

	for connection := range gameConns {
		connectionList = append(connectionList, connection.ID)
		playerID, _ := connection.GetPlayer()
		playerIDList = append(playerIDList, playerID)
	}

	b.logger.Debug("📊 Connection state before personalized broadcast",
		zap.String("game_id", gameID),
		zap.Int("total_connections", len(gameConns)),
		zap.Strings("connection_ids", connectionList),
		zap.Strings("player_ids", playerIDList))
}

// validateConnection checks if a connection is valid for sending personalized updates
func (b *Broadcaster) validateConnection(connection *Connection, gameID string) (string, bool) {
	playerID, _ := connection.GetPlayer()
	if playerID == "" {
		b.logger.Debug("⚠️ Skipping connection without player ID",
			zap.String("connection_id", connection.ID),
			zap.String("game_id", gameID))
		return "", false
	}

	// Skip connections with temporary playerIDs
	if strings.HasPrefix(playerID, "temp-") {
		b.logger.Debug("Skipping temporary connection",
			zap.String("connection_id", connection.ID),
			zap.String("temp_player_id", playerID))
		return "", false
	}

	return playerID, true
}

// sendPersonalizedMessage sends a personalized game update to a specific connection
func (b *Broadcaster) sendPersonalizedMessage(ctx context.Context, connection *Connection, playerID, gameID string) bool {
	gameData, err := b.getGameData(ctx, gameID)
	if err != nil {
		b.logger.Error("❌ Failed to get game data",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Error(err))
		return false
	}

	gamePlayers, err := b.getGamePlayers(ctx, gameID, gameData.PlayerIDs)
	if err != nil {
		b.logger.Error("❌ Failed to get player data",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Error(err))
		return false
	}

	gameDTO := dto.ToGameDto(gameData, gamePlayers, playerID)
	message := dto.WebSocketMessage{
		Type: dto.MessageTypeGameUpdated,
		Payload: dto.GameUpdatedPayload{
			Game: gameDTO,
		},
	}

	connection.SendMessage(message)

	b.logger.Debug("📤 Sent personalized game-updated to player",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID))

	return true
}

// getGameData retrieves game state
func (b *Broadcaster) getGameData(ctx context.Context, gameID string) (model.Game, error) {
	return b.gameService.GetGame(ctx, gameID)
}

// getGamePlayers retrieves all players for the game using PlayerIDs from game
func (b *Broadcaster) getGamePlayers(ctx context.Context, gameID string, playerIDs []string) ([]model.Player, error) {
	b.logger.Debug("🔍 Getting players for personalized DTO",
		zap.String("game_id", gameID),
		zap.Strings("game_player_ids", playerIDs))

	var gamePlayers []model.Player
	for _, pID := range playerIDs {
		player, err := b.playerService.GetPlayer(ctx, gameID, pID)
		if err != nil {
			b.logger.Warn("⚠️ Failed to get player data",
				zap.String("game_id", gameID),
				zap.String("missing_player_id", pID),
				zap.Error(err))
			continue
		}
		b.logger.Debug("✅ Retrieved player for DTO",
			zap.String("player_id", player.ID),
			zap.String("player_name", player.Name))
		gamePlayers = append(gamePlayers, player)
	}

	b.logger.Debug("📋 Players retrieved for DTO conversion",
		zap.Int("total_players", len(gamePlayers)))

	return gamePlayers, nil
}

// BroadcastPlayerDisconnection handles player disconnection broadcasting
func (b *Broadcaster) BroadcastPlayerDisconnection(ctx context.Context, playerID, gameID string, connection *Connection) {
	// Get game info for the broadcast
	game, err := b.gameService.GetGame(ctx, gameID)
	if err != nil {
		b.logger.Error("Failed to get game for player disconnection broadcast",
			zap.String("game_id", gameID),
			zap.Error(err))
		return
	}

	// Get player info using player service
	player, err := b.playerService.GetPlayer(ctx, gameID, playerID)
	if err != nil {
		b.logger.Error("Failed to get player for disconnection broadcast",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Error(err))
		return
	}
	playerName := player.Name

	// Get all players for personalized messages
	allPlayers, err := b.getGamePlayers(ctx, gameID, game.PlayerIDs)
	if err != nil {
		b.logger.Error("Failed to get players for disconnection broadcast",
			zap.String("game_id", gameID),
			zap.Error(err))
		return
	}

	// Send personalized disconnection messages to remaining players
	gameConns := b.manager.GetGameConnections(gameID)
	if gameConns != nil {
		for conn := range gameConns {
			if conn == connection { // Skip the disconnected player
				continue
			}

			connPlayerID, validConnection := b.validateConnection(conn, gameID)
			if !validConnection {
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
		}
	}

	b.logger.Info("📢 Player disconnected, broadcasted to other players in game",
		zap.String("player_id", playerID),
		zap.String("player_name", playerName),
		zap.String("game_id", gameID))
}

// BroadcastProductionPhaseStarted sends production phase started messages to all players in the game
func (b *Broadcaster) BroadcastProductionPhaseStarted(ctx context.Context, gameID string, playersData []dto.PlayerProductionData) {
	// Create production phase started payload
	productionPayload := dto.ProductionPhaseStartedPayload{
		PlayersData: playersData,
	}

	productionMessage := dto.WebSocketMessage{
		Type:    dto.MessageTypeProductionPhaseStarted,
		Payload: productionPayload,
		GameID:  gameID,
	}

	// Broadcast to all players in the game
	b.BroadcastToGame(gameID, productionMessage)

	b.logger.Info("📢 Broadcasted production phase started to all players",
		zap.String("game_id", gameID),
		zap.Int("players_data_count", len(playersData)))
}
