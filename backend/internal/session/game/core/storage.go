package core

import (
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/session/types"
)

// GameStorage manages the in-memory storage for all game data
// This is shared across all game sub-repositories
type GameStorage struct {
	mu    sync.RWMutex
	games map[string]*Game // gameID -> Game
}

// NewGameStorage creates a new game storage instance
func NewGameStorage() *GameStorage {
	return &GameStorage{
		games: make(map[string]*Game),
	}
}

// Get retrieves a game by ID
func (s *GameStorage) Get(gameID string) (*Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, exists := s.games[gameID]
	if !exists {
		return nil, &types.NotFoundError{Resource: "game", ID: gameID}
	}

	return game, nil
}

// GetAll retrieves all games
func (s *GameStorage) GetAll() ([]*Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Game, 0, len(s.games))
	for _, game := range s.games {
		result = append(result, game)
	}

	return result, nil
}

// Set stores or updates a game
func (s *GameStorage) Set(gameID string, game *Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.games[gameID] = game
	return nil
}

// Create creates a new game
func (s *GameStorage) Create(game *Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.games[game.ID]; exists {
		return fmt.Errorf("game %s already exists", game.ID)
	}

	s.games[game.ID] = game
	return nil
}

// WithLock executes a function with write lock held
// Useful for complex operations that need to atomically read and update
func (s *GameStorage) WithLock(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn()
}

// WithRLock executes a function with read lock held
// Useful for complex read operations
func (s *GameStorage) WithRLock(fn func()) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fn()
}
