package player

import (
	"context"

	"terraforming-mars-backend/internal/session/types"
)

// PlayerEffectRepository handles player effects and modifiers
type PlayerEffectRepository struct {
	storage *PlayerStorage
}

// NewPlayerEffectRepository creates a new player effect repository
func NewPlayerEffectRepository(storage *PlayerStorage) *PlayerEffectRepository {
	return &PlayerEffectRepository{
		storage: storage,
	}
}

// UpdateRequirementModifiers updates player requirement modifiers
func (r *PlayerEffectRepository) UpdateRequirementModifiers(ctx context.Context, gameID string, playerID string, modifiers []types.RequirementModifier) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.RequirementModifiers = modifiers

	return r.storage.Set(gameID, playerID, p)
}

// UpdatePlayerEffects updates player active effects
func (r *PlayerEffectRepository) UpdatePlayerEffects(ctx context.Context, gameID string, playerID string, effects []types.PlayerEffect) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.Effects = effects

	return r.storage.Set(gameID, playerID, p)
}
