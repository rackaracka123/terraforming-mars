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

// SendErrorToConnection sends an error message to a connection
func (b *Broadcaster) SendErrorToConnection(connection *Connection, errorMessage string) {
	_, gameID := connection.GetPlayer()

	message := dto.WebSocketMessage{
		Type: dto.MessageTypeError,
		Payload: dto.ErrorPayload{
			Message: errorMessage,
		},
		GameID: gameID,
	}

	b.SendToConnection(connection, message)
}

// SendPersonalizedGameUpdates sends personalized game-updated messages to all connected players
func (b *Broadcaster) SendPersonalizedGameUpdates(ctx context.Context, gameID string) {
	b.logger.Debug("🔍 Getting connected players for personalized broadcast", zap.String("game_id", gameID))

	gameConns := b.manager.GetGameConnections(gameID)

	if gameConns == nil {
		b.logger.Debug("No connections found for game", zap.String("game_id", gameID))
		return
	}

	sentCount := 0
	connectionsWithoutPlayerID := 0

	for connection := range gameConns {
		// Check if context is cancelled before processing each connection
		select {
		case <-ctx.Done():
			b.logger.Warn("Context cancelled during personalized game updates",
				zap.String("game_id", gameID),
				zap.Error(ctx.Err()))
			return
		default:
		}

		playerID, _ := connection.GetPlayer()
		if playerID == "" {
			connectionsWithoutPlayerID++
			continue
		}

		// Skip connections with temporary playerIDs
		if strings.HasPrefix(playerID, "temp-") {
			b.logger.Debug("Skipping temporary connection",
				zap.String("connection_id", connection.ID),
				zap.String("temp_player_id", playerID))
			continue
		}

		// Get game state
		playerGame, err := b.gameService.GetGame(ctx, gameID)
		if err != nil {
			b.logger.Error("❌ Failed to get game state",
				zap.String("game_id", gameID),
				zap.String("player_id", playerID),
				zap.Error(err))
			continue
		}

		// Get all players for the game using PlayerIDs from game
		var gamePlayers []model.Player
		for _, pID := range playerGame.PlayerIDs {
			player, err := b.playerService.GetPlayer(ctx, gameID, pID)
			if err != nil {
				b.logger.Warn("⚠️ Failed to get player data",
					zap.String("game_id", gameID),
					zap.String("missing_player_id", pID),
					zap.Error(err))
				continue
			}
			gamePlayers = append(gamePlayers, player)
		}

		// Convert to personalized DTO and send
		gameDTO := dto.ToGameDto(playerGame, gamePlayers, playerID)
		message := dto.WebSocketMessage{
			Type: dto.MessageTypeGameUpdated,
			Payload: dto.GameUpdatedPayload{
				Game: gameDTO,
			},
		}

		connection.SendMessage(message)
		sentCount++

		b.logger.Debug("📤 Sent personalized game-updated to player",
			zap.String("connection_id", connection.ID),
			zap.String("player_id", playerID))
	}

	b.logger.Info("📢 Sent personalized game-updated messages to players",
		zap.String("game_id", gameID),
		zap.Int("total_connections", len(gameConns)),
		zap.Int("messages_sent", sentCount),
		zap.Int("connections_without_player_id", connectionsWithoutPlayerID))
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
	var allPlayers []model.Player
	for _, pID := range game.PlayerIDs {
		p, err := b.playerService.GetPlayer(ctx, gameID, pID)
		if err != nil {
			b.logger.Warn("⚠️ Failed to get player for disconnection broadcast",
				zap.String("game_id", gameID),
				zap.String("player_id", pID),
				zap.Error(err))
			continue
		}
		allPlayers = append(allPlayers, p)
	}

	// Send personalized disconnection messages to remaining players
	gameConns := b.manager.GetGameConnections(gameID)
	if gameConns != nil {
		for conn := range gameConns {
			if conn == connection { // Skip the disconnected player
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
		}
	}

	b.logger.Info("📢 Player disconnected, broadcasted to other players in game",
		zap.String("player_id", playerID),
		zap.String("player_name", playerName),
		zap.String("game_id", gameID))
}

// SendAvailableCardsToPlayer sends available starting cards to a specific player
func (b *Broadcaster) SendAvailableCardsToPlayer(ctx context.Context, gameID, playerID string, cards []dto.CardDto) {
	// Create available cards payload
	availableCardsPayload := dto.AvailableCardsPayload{
		Cards: cards,
	}

	availableCardsMessage := dto.WebSocketMessage{
		Type:    dto.MessageTypeAvailableCards,
		Payload: availableCardsPayload,
		GameID:  gameID,
	}

	// Send to specific player
	connection := b.manager.GetConnectionByPlayerID(gameID, playerID)
	if connection != nil {
		connection.SendMessage(availableCardsMessage)
		b.logger.Info("📤 Sent available cards to player",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Int("card_count", len(cards)))
	} else {
		b.logger.Warn("⚠️ Player connection not found for available cards",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
	}
}
