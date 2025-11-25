package player

import (
	"context"

	"terraforming-mars-backend/internal/session/types"
)

// CorporationRepository handles corporation operations for a specific player
// Auto-saves changes after every operation
type CorporationRepository struct {
	player *Player // Reference to parent player
}

// NewCorporationRepository creates a new corporation repository for a specific player
func NewCorporationRepository(player *Player) *CorporationRepository {
	return &CorporationRepository{
		player: player,
	}
}

// Set sets the player's corporation ID
// Auto-saves changes to the player
func (r *CorporationRepository) Set(ctx context.Context, corporationID string) error {
	r.player.CorporationID = corporationID
	return nil
}

// Update updates the player's corporation with full card data
// Auto-saves changes to the player
func (r *CorporationRepository) Update(ctx context.Context, corporation types.Card) error {
	r.player.Player.Corporation = &corporation
	r.player.CorporationID = corporation.ID
	return nil
}
