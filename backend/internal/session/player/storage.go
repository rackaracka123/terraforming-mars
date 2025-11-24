package player

import (
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/session/types"
)

// PlayerStorage manages the in-memory storage for all player data
// This is shared across all player sub-repositories (core, resource, card, state)
type PlayerStorage struct {
	mu      sync.RWMutex
	players map[string]map[string]*Player // gameID -> playerID -> Player
}

// NewPlayerStorage creates a new player storage instance
func NewPlayerStorage() *PlayerStorage {
	return &PlayerStorage{
		players: make(map[string]map[string]*Player),
	}
}

// Get retrieves a player by game ID and player ID
func (s *PlayerStorage) Get(gameID string, playerID string) (*Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	gamePlayers, exists := s.players[gameID]
	if !exists {
		return nil, &types.NotFoundError{Resource: "game", ID: gameID}
	}

	player, exists := gamePlayers[playerID]
	if !exists {
		return nil, &types.NotFoundError{Resource: "player", ID: playerID}
	}

	return player, nil
}

// GetAll retrieves all players in a game
func (s *PlayerStorage) GetAll(gameID string) ([]*Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	gamePlayers, exists := s.players[gameID]
	if !exists {
		return []*Player{}, nil
	}

	result := make([]*Player, 0, len(gamePlayers))
	for _, player := range gamePlayers {
		result = append(result, player)
	}

	return result, nil
}

// Set stores or updates a player
func (s *PlayerStorage) Set(gameID string, playerID string, player *Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.players[gameID]; !exists {
		s.players[gameID] = make(map[string]*Player)
	}

	s.players[gameID][playerID] = player
	return nil
}

// Create creates a new player in a game
func (s *PlayerStorage) Create(gameID string, player *Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.players[gameID]; !exists {
		s.players[gameID] = make(map[string]*Player)
	}

	if _, exists := s.players[gameID][player.ID]; exists {
		return fmt.Errorf("player %s already exists in game %s", player.ID, gameID)
	}

	s.players[gameID][player.ID] = player
	return nil
}

// WithLock executes a function with write lock held
// Useful for complex operations that need to atomically read and update
func (s *PlayerStorage) WithLock(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn()
}

// WithRLock executes a function with read lock held
// Useful for complex read operations
func (s *PlayerStorage) WithRLock(fn func()) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fn()
}
