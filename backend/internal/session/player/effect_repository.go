package player

import (
	"context"

	"terraforming-mars-backend/internal/session/types"
)

// EffectRepository handles effects and modifiers for a specific player
// Auto-saves changes after every operation
type EffectRepository struct {
	player *Player // Reference to parent player
}

// NewEffectRepository creates a new effect repository for a specific player
func NewEffectRepository(player *Player) *EffectRepository {
	return &EffectRepository{
		player: player,
	}
}

// UpdateRequirementModifiers updates player requirement modifiers
// Auto-saves changes to the player
func (r *EffectRepository) UpdateRequirementModifiers(ctx context.Context, modifiers []types.RequirementModifier) error {
	r.player.RequirementModifiers = modifiers
	return nil
}

// UpdateEffects updates player active effects
// Auto-saves changes to the player
func (r *EffectRepository) UpdateEffects(ctx context.Context, effects []types.PlayerEffect) error {
	r.player.Effects = effects
	return nil
}
