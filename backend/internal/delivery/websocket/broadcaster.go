package websocket

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// Broadcaster handles automatic broadcasting for the new architecture
// Subscribes to BroadcastEvent and sends personalized game states to clients
type Broadcaster struct {
	gameRepo game.GameRepository
	eventBus *events.EventBusImpl
	hub      *core.Hub
	logger   *zap.Logger
}

// NewBroadcaster creates a broadcaster for the migration architecture
// Automatically subscribes to BroadcastEvent during initialization
func NewBroadcaster(
	gameRepo game.GameRepository,
	eventBus *events.EventBusImpl,
	hub *core.Hub,
) *Broadcaster {
	broadcaster := &Broadcaster{
		gameRepo: gameRepo,
		eventBus: eventBus,
		hub:      hub,
		logger:   logger.Get(),
	}

	// Subscribe to BroadcastEvent
	events.Subscribe(eventBus, broadcaster.OnBroadcastEvent)

	broadcaster.logger.Info("📡 Broadcaster initialized and subscribed to BroadcastEvent")

	return broadcaster
}

// OnBroadcastEvent handles BroadcastEvent by fetching game state and sending personalized DTOs
// This is the core of the event-driven broadcasting system
func (b *Broadcaster) OnBroadcastEvent(event events.BroadcastEvent) {
	ctx := context.Background()
	log := b.logger.With(zap.String("game_id", event.GameID))

	// Fetch game from repository
	game, err := b.gameRepo.Get(ctx, event.GameID)
	if err != nil {
		log.Error("Failed to get game for broadcast", zap.Error(err))
		return
	}

	// Determine which players to notify
	var playerIDs []string
	if event.PlayerIDs == nil {
		// Broadcast to all players in the game
		players := game.GetAllPlayers()
		playerIDs = make([]string, len(players))
		for i, player := range players {
			playerIDs[i] = player.ID()
		}
		log.Debug("📢 Broadcasting to all players", zap.Int("player_count", len(playerIDs)))
	} else {
		// Broadcast to specific players
		playerIDs = event.PlayerIDs
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

	log.Debug("✅ Broadcast completed")
}

// sendToPlayer creates a personalized DTO for a player and sends it via WebSocket
func (b *Broadcaster) sendToPlayer(ctx context.Context, game *game.Game, playerID string) error {
	log := b.logger.With(
		zap.String("game_id", game.ID()),
		zap.String("player_id", playerID),
	)

	// Create DTO from game state using migration mapper
	// TODO: Implement personalization based on playerID
	gameDto := dto.ToGameDto(game)

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
