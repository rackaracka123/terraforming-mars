package player

import (
	"context"

	"terraforming-mars-backend/internal/session/types"
)

// PlayerCorporationRepository handles player corporation operations
type PlayerCorporationRepository struct {
	storage *PlayerStorage
}

// NewPlayerCorporationRepository creates a new player corporation repository
func NewPlayerCorporationRepository(storage *PlayerStorage) *PlayerCorporationRepository {
	return &PlayerCorporationRepository{
		storage: storage,
	}
}

// SetCorporation sets the player's corporation ID
func (r *PlayerCorporationRepository) SetCorporation(ctx context.Context, gameID string, playerID string, corporationID string) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.CorporationID = corporationID

	return r.storage.Set(gameID, playerID, p)
}

// UpdateCorporation updates the player's corporation with full card data
func (r *PlayerCorporationRepository) UpdateCorporation(ctx context.Context, gameID string, playerID string, corporation types.Card) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.Corporation = &corporation
	p.CorporationID = corporation.ID

	return r.storage.Set(gameID, playerID, p)
}
