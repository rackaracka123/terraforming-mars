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

// GameUpdates handles game state broadcasting
type GameUpdates struct {
	manager       *core.Manager
	gameService   service.GameService
	playerService service.PlayerService
	logger        *zap.Logger
}

// NewGameUpdates creates a new game updates broadcaster
func NewGameUpdates(manager *core.Manager, gameService service.GameService, playerService service.PlayerService, logger *zap.Logger) *GameUpdates {
	return &GameUpdates{
		manager:       manager,
		gameService:   gameService,
		playerService: playerService,
		logger:        logger,
	}
}

// SendPersonalizedGameUpdates sends personalized game-updated messages to all connected players
func (gu *GameUpdates) SendPersonalizedGameUpdates(ctx context.Context, gameID string) {
	gu.logger.Debug("🔍 Getting connected players for personalized broadcast", zap.String("game_id", gameID))

	gameConns := gu.manager.GetGameConnections(gameID)

	if gameConns == nil {
		gu.logger.Debug("No connections found for game", zap.String("game_id", gameID))
		return
	}

	// Enhanced logging: track all connections before processing
	connectionList := make([]string, 0, len(gameConns))
	playerIDList := make([]string, 0, len(gameConns))
	for connection := range gameConns {
		connectionList = append(connectionList, connection.ID)
		playerID, _ := connection.GetPlayer()
		playerIDList = append(playerIDList, playerID)
	}

	gu.logger.Debug("📊 Connection state before personalized broadcast",
		zap.String("game_id", gameID),
		zap.Int("total_connections", len(gameConns)),
		zap.Strings("connection_ids", connectionList),
		zap.Strings("player_ids", playerIDList))

	sentCount := 0
	connectionsWithoutPlayerID := 0

	for connection := range gameConns {
		// Check if context is cancelled before processing each connection
		select {
		case <-ctx.Done():
			gu.logger.Warn("Context cancelled during personalized game updates",
				zap.String("game_id", gameID),
				zap.Error(ctx.Err()))
			return
		default:
		}

		playerID, _ := connection.GetPlayer()
		if playerID == "" {
			connectionsWithoutPlayerID++
			gu.logger.Debug("⚠️ Skipping connection without player ID",
				zap.String("connection_id", connection.ID),
				zap.String("game_id", gameID))
			continue
		}

		// Skip connections with temporary playerIDs
		if strings.HasPrefix(playerID, "temp-") {
			gu.logger.Debug("Skipping temporary connection",
				zap.String("connection_id", connection.ID),
				zap.String("temp_player_id", playerID))
			continue
		}

		// Send personalized update
		if gu.sendPersonalizedUpdate(ctx, connection, gameID, playerID) {
			sentCount++
		}
	}

	gu.logger.Info("📢 Sent personalized game-updated messages to players",
		zap.String("game_id", gameID),
		zap.Int("total_connections", len(gameConns)),
		zap.Int("messages_sent", sentCount),
		zap.Int("connections_without_player_id", connectionsWithoutPlayerID))
}

// sendPersonalizedUpdate sends a personalized update to a single player
func (gu *GameUpdates) sendPersonalizedUpdate(ctx context.Context, connection *core.Connection, gameID, playerID string) bool {
	// Get game state
	playerGame, err := gu.gameService.GetGame(ctx, gameID)
	if err != nil {
		gu.logger.Error("❌ Failed to get game state",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Error(err))
		return false
	}

	// Get all players for the game using PlayerIDs from game
	gu.logger.Debug("🔍 Getting players for personalized DTO",
		zap.String("game_id", gameID),
		zap.String("viewing_player_id", playerID),
		zap.Strings("game_player_ids", playerGame.PlayerIDs))

	var gamePlayers []model.Player
	for _, pID := range playerGame.PlayerIDs {
		player, err := gu.playerService.GetPlayer(ctx, gameID, pID)
		if err != nil {
			gu.logger.Warn("⚠️ Failed to get player data",
				zap.String("game_id", gameID),
				zap.String("missing_player_id", pID),
				zap.Error(err))
			continue
		}
		gu.logger.Debug("✅ Retrieved player for DTO",
			zap.String("player_id", player.ID),
			zap.String("player_name", player.Name))
		gamePlayers = append(gamePlayers, player)
	}

	gu.logger.Debug("📋 Players retrieved for DTO conversion",
		zap.String("viewing_player_id", playerID),
		zap.Int("total_players", len(gamePlayers)))

	// Convert to personalized DTO and send
	gameDTO := dto.ToGameDto(playerGame, gamePlayers, playerID)
	message := dto.WebSocketMessage{
		Type: dto.MessageTypeGameUpdated,
		Payload: dto.GameUpdatedPayload{
			Game: gameDTO,
		},
	}

	connection.SendMessage(message)

	gu.logger.Debug("📤 Sent personalized game-updated to player",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID))
	return true
}

// SendAvailableCardsToPlayer sends available starting cards to a specific player
func (gu *GameUpdates) SendAvailableCardsToPlayer(ctx context.Context, gameID, playerID string, cards []dto.CardDto) {
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
	connection := gu.manager.GetConnectionByPlayerID(gameID, playerID)
	if connection != nil {
		connection.SendMessage(availableCardsMessage)
		gu.logger.Info("📤 Sent available cards to player",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Int("card_count", len(cards)))
	} else {
		gu.logger.Warn("⚠️ Player connection not found for available cards",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
	}
}

// BroadcastProductionPhaseStarted sends production phase started messages to all players in the game
func (gu *GameUpdates) BroadcastProductionPhaseStarted(ctx context.Context, gameID string, playersData []dto.PlayerProductionData) {
	// Create production phase started payload
	productionPayload := dto.ProductionPhaseStartedPayload{
		PlayersData: playersData,
	}

	productionMessage := dto.WebSocketMessage{
		Type:    dto.MessageTypeProductionPhaseStarted,
		Payload: productionPayload,
		GameID:  gameID,
	}

	// Broadcast to all players in the game using system messages
	gu.broadcastToGame(gameID, productionMessage)

	gu.logger.Info("📢 Broadcasted production phase started to all players",
		zap.String("game_id", gameID),
		zap.Int("players_data_count", len(playersData)))
}

// broadcastToGame is a helper method for internal broadcasting
func (gu *GameUpdates) broadcastToGame(gameID string, message dto.WebSocketMessage) {
	gameConns := gu.manager.GetGameConnections(gameID)

	if gameConns == nil || len(gameConns) == 0 {
		gu.logger.Warn("❌ No connections found for game", zap.String("game_id", gameID))
		return
	}

	sentCount := 0
	for connection := range gameConns {
		connection.SendMessage(message)
		sentCount++
	}

	gu.logger.Info("📢 Game update broadcasted to game clients",
		zap.String("game_id", gameID),
		zap.String("message_type", string(message.Type)),
		zap.Int("messages_sent", sentCount))
}
