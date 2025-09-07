package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GameRepository manages game metadata and state
type GameRepository interface {
	// Create a new game
	Create(ctx context.Context, settings model.GameSettings) (*model.Game, error)

	// Get game by ID
	Get(ctx context.Context, gameID string) (*model.Game, error)

	// Update game state
	Update(ctx context.Context, game *model.Game) error

	// List games with optional status filter
	List(ctx context.Context, status string) ([]*model.Game, error)

	// Delete game
	Delete(ctx context.Context, gameID string) error
}

// GameRepositoryImpl implements GameRepository interface
type GameRepositoryImpl struct {
	games    map[string]*model.Game
	mutex    sync.RWMutex
	eventBus events.EventBus
}

// NewGameRepository creates a new game repository
func NewGameRepository(eventBus events.EventBus) GameRepository {
	return &GameRepositoryImpl{
		games:    make(map[string]*model.Game),
		eventBus: eventBus,
	}
}

// Create creates a new game with the given settings
func (r *GameRepositoryImpl) Create(ctx context.Context, settings model.GameSettings) (*model.Game, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.Get()
	log.Debug("Creating new game")

	// Generate unique game ID
	gameID := uuid.New().String()

	// Create the game
	game := model.NewGame(gameID, settings)

	// Store in repository
	r.games[gameID] = game

	log.Debug("Game created",
		zap.String("game_id", gameID),
	)

	// Publish game created event
	if r.eventBus != nil {
		gameCreatedEvent := events.NewGameCreatedEvent(gameID, settings)
		if err := r.eventBus.Publish(ctx, gameCreatedEvent); err != nil {
			log.Warn("Failed to publish game created event", zap.Error(err))
		}
	}

	return game, nil
}

// Get retrieves a game by ID
func (r *GameRepositoryImpl) Get(ctx context.Context, gameID string) (*model.Game, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if gameID == "" {
		return nil, fmt.Errorf("game ID cannot be empty")
	}

	game, exists := r.games[gameID]
	if !exists {
		return nil, fmt.Errorf("game with ID %s not found", gameID)
	}

	return game.DeepCopy(), nil
}

// Update updates a game in the repository
func (r *GameRepositoryImpl) Update(ctx context.Context, game *model.Game) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(game.ID, "")

	if game == nil {
		return fmt.Errorf("game cannot be nil")
	}

	if _, exists := r.games[game.ID]; !exists {
		log.Error("Attempted to update non-existent game")
		return fmt.Errorf("game with ID %s not found", game.ID)
	}

	// Capture old state for event
	oldGame := *r.games[game.ID]

	// Update timestamp
	game.UpdatedAt = time.Now()

	// Store updated game
	r.games[game.ID] = game

	log.Debug("Game updated")

	// Publish game updated event
	if r.eventBus != nil {
		gameUpdatedEvent := events.NewGameStateChangedEvent(game.ID, &oldGame, game)
		if err := r.eventBus.Publish(ctx, gameUpdatedEvent); err != nil {
			log.Warn("Failed to publish game updated event", zap.Error(err))
		}
	}

	return nil
}

// List returns all games, optionally filtered by status
func (r *GameRepositoryImpl) List(ctx context.Context, status string) ([]*model.Game, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	games := make([]*model.Game, 0)

	for _, game := range r.games {
		if status == "" || string(game.Status) == status {
			games = append(games, game.DeepCopy())
		}
	}

	return games, nil
}

// Delete removes a game from the repository
func (r *GameRepositoryImpl) Delete(ctx context.Context, gameID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	log := logger.WithGameContext(gameID, "")

	if gameID == "" {
		return fmt.Errorf("game ID cannot be empty")
	}

	_, exists := r.games[gameID]
	if !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	delete(r.games, gameID)

	log.Info("Game deleted")

	// Publish game deleted event
	if r.eventBus != nil {
		gameDeletedEvent := events.NewGameDeletedEvent(gameID)
		if err := r.eventBus.Publish(ctx, gameDeletedEvent); err != nil {
			log.Warn("Failed to publish game deleted event", zap.Error(err))
		}
	}

	return nil
}
