package game

import (
	"context"
	"fmt"
	"sync"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// Repository manages game data with event-driven updates
type Repository interface {
	// Create creates a new game
	Create(ctx context.Context, game *Game) error

	// GetByID retrieves a game by ID
	GetByID(ctx context.Context, gameID string) (*Game, error)

	// AddPlayer adds a player to a game (event-driven)
	AddPlayer(ctx context.Context, gameID string, playerID string) error

	// UpdateStatus updates game status (event-driven)
	UpdateStatus(ctx context.Context, gameID string, status GameStatus) error

	// UpdatePhase updates game phase (event-driven)
	UpdatePhase(ctx context.Context, gameID string, phase GamePhase) error

	// SetHostPlayer sets the host player for a game
	SetHostPlayer(ctx context.Context, gameID string, playerID string) error

	// SetCurrentTurn sets the current turn player
	SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error
}

// RepositoryImpl implements the Repository interface with in-memory storage
type RepositoryImpl struct {
	mu       sync.RWMutex
	games    map[string]*Game // gameID -> Game
	eventBus *events.EventBusImpl
}

// NewRepository creates a new game repository
func NewRepository(eventBus *events.EventBusImpl) Repository {
	return &RepositoryImpl{
		games:    make(map[string]*Game),
		eventBus: eventBus,
	}
}

// Create creates a new game
func (r *RepositoryImpl) Create(ctx context.Context, game *Game) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.games[game.ID]; exists {
		return fmt.Errorf("game %s already exists", game.ID)
	}

	r.games[game.ID] = game

	// Event publishing can be added here if needed
	// For now, game creation doesn't need an event

	return nil
}

// GetByID retrieves a game by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, gameID string) (*Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	game, exists := r.games[gameID]
	if !exists {
		return nil, &model.NotFoundError{Resource: "game", ID: gameID}
	}

	return game, nil
}

// AddPlayer adds a player to a game (event-driven)
func (r *RepositoryImpl) AddPlayer(ctx context.Context, gameID string, playerID string) error {
	r.mu.Lock()

	game, exists := r.games[gameID]
	if !exists {
		r.mu.Unlock()
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	// Check if player already in game
	for _, pid := range game.PlayerIDs {
		if pid == playerID {
			r.mu.Unlock()
			return fmt.Errorf("player %s already in game %s", playerID, gameID)
		}
	}

	game.PlayerIDs = append(game.PlayerIDs, playerID)

	// Release lock BEFORE publishing event to avoid deadlock
	// Event subscribers may need to read from this repository
	r.mu.Unlock()

	// Publish PlayerJoinedEvent for event-driven broadcasting
	// This is safe to do after unlocking because the event only contains IDs (not object references)
	events.Publish(r.eventBus, repository.PlayerJoinedEvent{
		GameID:     gameID,
		PlayerID:   playerID,
		PlayerName: "", // Name not available at repository level, subscribers can look up if needed
		Timestamp:  time.Now(),
	})

	return nil
}

// UpdateStatus updates game status (event-driven)
func (r *RepositoryImpl) UpdateStatus(ctx context.Context, gameID string, status GameStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	game, exists := r.games[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	oldStatus := game.Status
	game.Status = status

	// Event publishing can be added here if needed
	_ = oldStatus // Use oldStatus if needed for event payload

	return nil
}

// UpdatePhase updates game phase (event-driven)
func (r *RepositoryImpl) UpdatePhase(ctx context.Context, gameID string, phase GamePhase) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	game, exists := r.games[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	oldPhase := game.CurrentPhase
	game.CurrentPhase = phase

	// Event publishing can be added here if needed
	_ = oldPhase // Use oldPhase if needed for event payload

	return nil
}

// SetHostPlayer sets the host player for a game
func (r *RepositoryImpl) SetHostPlayer(ctx context.Context, gameID string, playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	game, exists := r.games[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	game.HostPlayerID = playerID

	return nil
}

// SetCurrentTurn sets the current turn player
func (r *RepositoryImpl) SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	game, exists := r.games[gameID]
	if !exists {
		return &model.NotFoundError{Resource: "game", ID: gameID}
	}

	game.CurrentTurn = playerID

	return nil
}
