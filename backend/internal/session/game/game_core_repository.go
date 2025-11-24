package game

import (
	"context"
	"time"

	"terraforming-mars-backend/internal/events"
)

// GameCoreRepository handles core CRUD and game state operations
type GameCoreRepository struct {
	storage  *GameStorage
	eventBus *events.EventBusImpl
}

// NewGameCoreRepository creates a new game core repository
func NewGameCoreRepository(storage *GameStorage, eventBus *events.EventBusImpl) *GameCoreRepository {
	return &GameCoreRepository{
		storage:  storage,
		eventBus: eventBus,
	}
}

// Create creates a new game
func (r *GameCoreRepository) Create(ctx context.Context, game *Game) error {
	return r.storage.Create(game)
}

// GetByID retrieves a game by ID
func (r *GameCoreRepository) GetByID(ctx context.Context, gameID string) (*Game, error) {
	return r.storage.Get(gameID)
}

// List retrieves all games, optionally filtered by status
func (r *GameCoreRepository) List(ctx context.Context, status string) ([]*Game, error) {
	allGames, err := r.storage.GetAll()
	if err != nil {
		return nil, err
	}

	// If no status filter, return all
	if status == "" {
		return allGames, nil
	}

	// Filter by status
	filtered := make([]*Game, 0)
	for _, game := range allGames {
		if string(game.Status) == status {
			filtered = append(filtered, game)
		}
	}

	return filtered, nil
}

// UpdateStatus updates game status and publishes event
func (r *GameCoreRepository) UpdateStatus(ctx context.Context, gameID string, status GameStatus) error {
	game, err := r.storage.Get(gameID)
	if err != nil {
		return err
	}

	oldStatus := game.Status
	game.Status = status

	err = r.storage.Set(gameID, game)
	if err != nil {
		return err
	}

	// Publish event if status changed
	if oldStatus != status {
		events.Publish(r.eventBus, events.GameStatusChangedEvent{
			GameID:    gameID,
			OldStatus: string(oldStatus),
			NewStatus: string(status),
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdatePhase updates game phase and publishes event
func (r *GameCoreRepository) UpdatePhase(ctx context.Context, gameID string, phase GamePhase) error {
	game, err := r.storage.Get(gameID)
	if err != nil {
		return err
	}

	oldPhase := game.CurrentPhase
	game.CurrentPhase = phase

	err = r.storage.Set(gameID, game)
	if err != nil {
		return err
	}

	// Publish event if phase changed
	if oldPhase != phase {
		events.Publish(r.eventBus, events.GamePhaseChangedEvent{
			GameID:    gameID,
			OldPhase:  string(oldPhase),
			NewPhase:  string(phase),
			Timestamp: time.Now(),
		})
	}

	return nil
}
