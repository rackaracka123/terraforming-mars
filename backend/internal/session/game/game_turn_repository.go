package game

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
)

// GameTurnRepository handles turn and player management operations
type GameTurnRepository struct {
	storage  *GameStorage
	eventBus *events.EventBusImpl
}

// NewGameTurnRepository creates a new game turn repository
func NewGameTurnRepository(storage *GameStorage, eventBus *events.EventBusImpl) *GameTurnRepository {
	return &GameTurnRepository{
		storage:  storage,
		eventBus: eventBus,
	}
}

// SetHostPlayer sets the host player for a game
func (r *GameTurnRepository) SetHostPlayer(ctx context.Context, gameID string, playerID string) error {
	game, err := r.storage.Get(gameID)
	if err != nil {
		return err
	}

	game.HostPlayerID = playerID

	return r.storage.Set(gameID, game)
}

// SetCurrentTurn sets the current turn player
func (r *GameTurnRepository) SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	game, err := r.storage.Get(gameID)
	if err != nil {
		return err
	}

	game.CurrentTurn = playerID

	return r.storage.Set(gameID, game)
}

// AddPlayer adds a player to a game and publishes event
func (r *GameTurnRepository) AddPlayer(ctx context.Context, gameID string, playerID string) error {
	game, err := r.storage.Get(gameID)
	if err != nil {
		return err
	}

	// Check if player already exists
	for _, pid := range game.PlayerIDs {
		if pid == playerID {
			return fmt.Errorf("player %s already in game %s", playerID, gameID)
		}
	}

	game.PlayerIDs = append(game.PlayerIDs, playerID)

	err = r.storage.Set(gameID, game)
	if err != nil {
		return err
	}

	// Publish PlayerJoinedEvent
	log := logger.Get().With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID))
	log.Info("📡 Publishing PlayerJoinedEvent")

	events.Publish(r.eventBus, events.PlayerJoinedEvent{
		GameID:     gameID,
		PlayerID:   playerID,
		PlayerName: "", // Name not available at repository level
		Timestamp:  time.Now(),
	})

	log.Info("✅ PlayerJoinedEvent published")

	return nil
}
