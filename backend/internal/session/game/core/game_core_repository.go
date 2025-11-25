package core

import (
	"context"
	"time"

	"terraforming-mars-backend/internal/events"
)

// GameCoreRepository handles core game state operations
// This repository is scoped to a specific game instance
type GameCoreRepository struct {
	gameID   string // Bound to specific game
	storage  *GameStorage
	eventBus *events.EventBusImpl
}

// NewGameCoreRepository creates a new game core repository bound to a specific game
func NewGameCoreRepository(gameID string, storage *GameStorage, eventBus *events.EventBusImpl) *GameCoreRepository {
	return &GameCoreRepository{
		gameID:   gameID,
		storage:  storage,
		eventBus: eventBus,
	}
}

// Get retrieves the game data for this repository's bound game
func (r *GameCoreRepository) Get(ctx context.Context) (*Game, error) {
	return r.storage.Get(r.gameID)
}

// UpdateStatus updates game status and publishes event
func (r *GameCoreRepository) UpdateStatus(ctx context.Context, status GameStatus) error {
	game, err := r.storage.Get(r.gameID)
	if err != nil {
		return err
	}

	oldStatus := game.Status
	game.Status = status

	err = r.storage.Set(r.gameID, game)
	if err != nil {
		return err
	}

	// Publish event if status changed
	if oldStatus != status {
		events.Publish(r.eventBus, events.GameStatusChangedEvent{
			GameID:    r.gameID,
			OldStatus: string(oldStatus),
			NewStatus: string(status),
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdatePhase updates game phase and publishes event
func (r *GameCoreRepository) UpdatePhase(ctx context.Context, phase GamePhase) error {
	game, err := r.storage.Get(r.gameID)
	if err != nil {
		return err
	}

	oldPhase := game.CurrentPhase
	game.CurrentPhase = phase

	err = r.storage.Set(r.gameID, game)
	if err != nil {
		return err
	}

	// Publish event if phase changed
	if oldPhase != phase {
		events.Publish(r.eventBus, events.GamePhaseChangedEvent{
			GameID:    r.gameID,
			OldPhase:  string(oldPhase),
			NewPhase:  string(phase),
			Timestamp: time.Now(),
		})
	}

	return nil
}
