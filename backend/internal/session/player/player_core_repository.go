package player

import (
	"context"
)

// PlayerCoreRepository handles core CRUD operations for players
type PlayerCoreRepository struct {
	storage *PlayerStorage
}

// NewPlayerCoreRepository creates a new player core repository
func NewPlayerCoreRepository(storage *PlayerStorage) *PlayerCoreRepository {
	return &PlayerCoreRepository{
		storage: storage,
	}
}

// Create creates a new player in a game
func (r *PlayerCoreRepository) Create(ctx context.Context, gameID string, p *Player) error {
	return r.storage.Create(gameID, p)
}

// GetByID retrieves a player by ID from a specific game
func (r *PlayerCoreRepository) GetByID(ctx context.Context, gameID string, playerID string) (*Player, error) {
	return r.storage.Get(gameID, playerID)
}

// ListByGameID retrieves all players in a game
func (r *PlayerCoreRepository) ListByGameID(ctx context.Context, gameID string) ([]*Player, error) {
	return r.storage.GetAll(gameID)
}
