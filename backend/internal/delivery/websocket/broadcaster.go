package websocket

import (
	"context"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Broadcaster handles game state broadcasting to WebSocket clients
// Called explicitly by WebSocket handlers after actions complete
type Broadcaster struct {
	gameRepo     game.GameRepository
	hub          *core.Hub
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewBroadcaster creates a broadcaster for explicit broadcasting
func NewBroadcaster(
	gameRepo game.GameRepository,
	hub *core.Hub,
	cardRegistry cards.CardRegistry,
) *Broadcaster {
	broadcaster := &Broadcaster{
		gameRepo:     gameRepo,
		hub:          hub,
		cardRegistry: cardRegistry,
		logger:       logger.Get(),
	}

	broadcaster.logger.Info("📡 Broadcaster initialized")

	return broadcaster
}

// BroadcastGameState broadcasts game state to specified players (nil = all players)
// Called explicitly by WebSocket handlers after action execution completes
func (b *Broadcaster) BroadcastGameState(gameID string, playerIDs []string) {
	ctx := context.Background()
	log := b.logger.With(zap.String("game_id", gameID))

	// Fetch game from repository
	game, err := b.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for broadcast", zap.Error(err))
		return
	}

	// Determine which players to notify
	if playerIDs == nil {
		// Broadcast to all players in the game
		players := game.GetAllPlayers()
		playerIDs = make([]string, len(players))
		for i, player := range players {
			playerIDs[i] = player.ID()
		}
		log.Debug("📢 Broadcasting to all players", zap.Int("player_count", len(playerIDs)))
	} else {
		// Broadcast to specific players
		log.Debug("📢 Broadcasting to specific players", zap.Strings("player_ids", playerIDs))
	}

	// Send personalized game state to each player
	for _, playerID := range playerIDs {
		if err := b.sendToPlayer(ctx, game, playerID); err != nil {
			log.Error("Failed to send game state to player",
				zap.String("player_id", playerID),
				zap.Error(err))
			// Continue with other players even if one fails
		}
	}

	log.Debug("✅ Broadcast completed", zap.Int("player_count", len(playerIDs)))
}

// sendToPlayer creates a personalized DTO for a player and sends it via WebSocket
func (b *Broadcaster) sendToPlayer(ctx context.Context, game *game.Game, playerID string) error {
	log := b.logger.With(
		zap.String("game_id", game.ID()),
		zap.String("player_id", playerID),
	)

	// Create personalized DTO from game state using migration mapper
	// playerID determines which player is "currentPlayer" vs "otherPlayers"
	gameDto := dto.ToGameDto(game, b.cardRegistry, playerID)

	// Create game updated message
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypeGameUpdated,
		GameID: game.ID(),
		Payload: dto.GameUpdatedPayload{
			Game: gameDto,
		},
	}

	// Send via Hub
	if err := b.hub.SendToPlayer(game.ID(), playerID, message); err != nil {
		return err
	}

	log.Debug("✅ Sent personalized game state to player")
	return nil
}
