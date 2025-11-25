package core

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
)

// GameTurnRepository handles turn and player management operations
// This repository is scoped to a specific game instance
type GameTurnRepository struct {
	gameID   string // Bound to specific game
	storage  *GameStorage
	eventBus *events.EventBusImpl
}

// NewGameTurnRepository creates a new game turn repository bound to a specific game
func NewGameTurnRepository(gameID string, storage *GameStorage, eventBus *events.EventBusImpl) *GameTurnRepository {
	return &GameTurnRepository{
		gameID:   gameID,
		storage:  storage,
		eventBus: eventBus,
	}
}

// SetHostPlayer sets the host player for a game
func (r *GameTurnRepository) SetHostPlayer(ctx context.Context, playerID string) error {
	game, err := r.storage.Get(r.gameID)
	if err != nil {
		return err
	}

	game.HostPlayerID = playerID

	return r.storage.Set(r.gameID, game)
}

// SetCurrentTurn sets the current turn player
func (r *GameTurnRepository) SetCurrentTurn(ctx context.Context, playerID *string) error {
	game, err := r.storage.Get(r.gameID)
	if err != nil {
		return err
	}

	game.CurrentTurn = playerID

	return r.storage.Set(r.gameID, game)
}

// AddPlayer adds a player to a game and publishes event
func (r *GameTurnRepository) AddPlayer(ctx context.Context, playerID string) error {
	game, err := r.storage.Get(r.gameID)
	if err != nil {
		return err
	}

	// Check if player already exists
	for _, pid := range game.PlayerIDs {
		if pid == playerID {
			return fmt.Errorf("player %s already in game %s", playerID, r.gameID)
		}
	}

	game.PlayerIDs = append(game.PlayerIDs, playerID)

	err = r.storage.Set(r.gameID, game)
	if err != nil {
		return err
	}

	// Publish PlayerJoinedEvent
	log := logger.Get().With(
		zap.String("game_id", r.gameID),
		zap.String("player_id", playerID))
	log.Info("📡 Publishing PlayerJoinedEvent")

	events.Publish(r.eventBus, events.PlayerJoinedEvent{
		GameID:     r.gameID,
		PlayerID:   playerID,
		PlayerName: "", // Name not available at repository level
		Timestamp:  time.Now(),
	})

	log.Info("✅ PlayerJoinedEvent published")

	return nil
}
