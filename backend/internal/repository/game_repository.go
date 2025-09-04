package repository

import (
	"fmt"
	"sync"
	"terraforming-mars-backend/internal/model"

	"github.com/google/uuid"
)

// GameRepository handles game storage and retrieval
type GameRepository struct {
	games map[string]*model.Game
	mutex sync.RWMutex
}

// NewGameRepository creates a new game repository
func NewGameRepository() *GameRepository {
	return &GameRepository{
		games: make(map[string]*model.Game),
	}
}

// CreateGame creates a new game with the given settings
func (r *GameRepository) CreateGame(settings model.GameSettings) (*model.Game, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Generate unique game ID
	gameID := uuid.New().String()

	// Create the game
	game := model.NewGame(gameID, settings)

	// Store in repository
	r.games[gameID] = game

	return game, nil
}

// GetGame retrieves a game by ID
func (r *GameRepository) GetGame(gameID string) (*model.Game, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	game, exists := r.games[gameID]
	if !exists {
		return nil, fmt.Errorf("game with ID %s not found", gameID)
	}

	return game, nil
}

// UpdateGame updates a game in the repository
func (r *GameRepository) UpdateGame(game *model.Game) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.games[game.ID]; !exists {
		return fmt.Errorf("game with ID %s not found", game.ID)
	}

	r.games[game.ID] = game
	return nil
}

// ListGames returns all games in the repository
func (r *GameRepository) ListGames() ([]*model.Game, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	games := make([]*model.Game, 0, len(r.games))
	for _, game := range r.games {
		games = append(games, game)
	}

	return games, nil
}

// DeleteGame removes a game from the repository
func (r *GameRepository) DeleteGame(gameID string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.games[gameID]; !exists {
		return fmt.Errorf("game with ID %s not found", gameID)
	}

	delete(r.games, gameID)
	return nil
}

// GetGamesByStatus returns games with a specific status
func (r *GameRepository) GetGamesByStatus(status string) ([]*model.Game, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	games := make([]*model.Game, 0)
	for _, game := range r.games {
		if string(game.Status) == status {
			games = append(games, game)
		}
	}

	return games, nil
}
