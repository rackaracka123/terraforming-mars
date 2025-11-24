package player

import (
	"context"
	"time"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/types"
)

var logResource = logger.Get()

// PlayerResourceRepository handles player resource and economy operations
type PlayerResourceRepository struct {
	storage  *PlayerStorage
	eventBus *events.EventBusImpl
}

// NewPlayerResourceRepository creates a new player resource repository
func NewPlayerResourceRepository(storage *PlayerStorage, eventBus *events.EventBusImpl) *PlayerResourceRepository {
	return &PlayerResourceRepository{
		storage:  storage,
		eventBus: eventBus,
	}
}

// UpdateResources updates player resources and publishes a batched ResourcesChangedEvent
func (r *PlayerResourceRepository) UpdateResources(ctx context.Context, gameID string, playerID string, resources types.Resources) error {
	// Get current player
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	oldResources := p.Resources

	// Calculate changes (delta for each resource type)
	changes := make(map[string]int)
	if oldResources.Credits != resources.Credits {
		changes["credits"] = resources.Credits - oldResources.Credits
	}
	if oldResources.Steel != resources.Steel {
		changes["steel"] = resources.Steel - oldResources.Steel
	}
	if oldResources.Titanium != resources.Titanium {
		changes["titanium"] = resources.Titanium - oldResources.Titanium
	}
	if oldResources.Plants != resources.Plants {
		changes["plants"] = resources.Plants - oldResources.Plants
	}
	if oldResources.Energy != resources.Energy {
		changes["energy"] = resources.Energy - oldResources.Energy
	}
	if oldResources.Heat != resources.Heat {
		changes["heat"] = resources.Heat - oldResources.Heat
	}

	// Update player resources
	p.Resources = resources

	// Store updated player
	err = r.storage.Set(gameID, playerID, p)
	if err != nil {
		return err
	}

	// Publish single batched event if any resources changed
	if len(changes) > 0 {
		events.Publish(r.eventBus, events.ResourcesChangedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			Changes:   changes,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateProduction updates player production
func (r *PlayerResourceRepository) UpdateProduction(ctx context.Context, gameID string, playerID string, production types.Production) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.Production = production

	return r.storage.Set(gameID, playerID, p)
}

// UpdateTerraformRating updates player terraform rating and publishes TerraformRatingChangedEvent
func (r *PlayerResourceRepository) UpdateTerraformRating(ctx context.Context, gameID string, playerID string, rating int) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	oldRating := p.TerraformRating
	p.TerraformRating = rating

	err = r.storage.Set(gameID, playerID, p)
	if err != nil {
		return err
	}

	// Publish TerraformRatingChangedEvent
	if oldRating != rating {
		logResource.Debug("📡 Publishing TerraformRatingChangedEvent",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.Int("old_rating", oldRating),
			zap.Int("new_rating", rating))

		events.Publish(r.eventBus, events.TerraformRatingChangedEvent{
			GameID:    gameID,
			PlayerID:  playerID,
			OldRating: oldRating,
			NewRating: rating,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateVictoryPoints updates player victory points
func (r *PlayerResourceRepository) UpdateVictoryPoints(ctx context.Context, gameID string, playerID string, victoryPoints int) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.VictoryPoints = victoryPoints

	return r.storage.Set(gameID, playerID, p)
}

// UpdateResourceStorage updates player resource storage (cards that store resources)
func (r *PlayerResourceRepository) UpdateResourceStorage(ctx context.Context, gameID string, playerID string, storage map[string]int) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.ResourceStorage = storage

	return r.storage.Set(gameID, playerID, p)
}

// UpdatePaymentSubstitutes updates player payment substitutes
func (r *PlayerResourceRepository) UpdatePaymentSubstitutes(ctx context.Context, gameID string, playerID string, substitutes []types.PaymentSubstitute) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.PaymentSubstitutes = substitutes

	return r.storage.Set(gameID, playerID, p)
}
