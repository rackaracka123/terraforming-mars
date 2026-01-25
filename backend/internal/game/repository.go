package game

import (
	"context"
	"fmt"
	"sync"
)

// GameRepository manages the collection of active games
type GameRepository interface {
	Get(ctx context.Context, gameID string) (*Game, error)
	Create(ctx context.Context, game *Game) error
	Delete(ctx context.Context, gameID string) error
	List(ctx context.Context, status *GameStatus) ([]*Game, error)
	Exists(ctx context.Context, gameID string) bool
}

// InMemoryGameRepository implements GameRepository using in-memory storage
type InMemoryGameRepository struct {
	mu    sync.RWMutex
	games map[string]*Game
}

// NewInMemoryGameRepository creates a new in-memory game repository
func NewInMemoryGameRepository() *InMemoryGameRepository {
	return &InMemoryGameRepository{
		games: make(map[string]*Game),
	}
}

// Get retrieves a game by ID
func (r *InMemoryGameRepository) Get(ctx context.Context, gameID string) (*Game, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	game, exists := r.games[gameID]
	if !exists {
		return nil, fmt.Errorf("game %s not found", gameID)
	}

	return game, nil
}

// Create adds a new game to the repository
func (r *InMemoryGameRepository) Create(ctx context.Context, game *Game) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if game == nil {
		return fmt.Errorf("game cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.games[game.ID()]; exists {
		return fmt.Errorf("game %s already exists", game.ID())
	}

	r.games[game.ID()] = game
	return nil
}

// Delete removes a game from the repository
func (r *InMemoryGameRepository) Delete(ctx context.Context, gameID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.games[gameID]; !exists {
		return fmt.Errorf("game %s not found", gameID)
	}

	delete(r.games, gameID)
	return nil
}

// List returns all games, optionally filtered by status
func (r *InMemoryGameRepository) List(ctx context.Context, status *GameStatus) ([]*Game, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	games := make([]*Game, 0, len(r.games))
	for _, game := range r.games {
		// Filter by status if provided
		if status != nil && game.Status() != *status {
			continue
		}
		games = append(games, game)
	}

	return games, nil
}

// Exists checks if a game exists
func (r *InMemoryGameRepository) Exists(ctx context.Context, gameID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.games[gameID]
	return exists
}
