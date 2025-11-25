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

// ResourceRepository handles resource and economy operations for a specific player
// Auto-saves changes after every operation
type ResourceRepository struct {
	player   *Player // Reference to parent player
	eventBus *events.EventBusImpl
}

// NewResourceRepository creates a new resource repository for a specific player
func NewResourceRepository(player *Player, eventBus *events.EventBusImpl) *ResourceRepository {
	return &ResourceRepository{
		player:   player,
		eventBus: eventBus,
	}
}

// Update updates player resources and publishes a batched ResourcesChangedEvent
// Auto-saves changes to the player
func (r *ResourceRepository) Update(ctx context.Context, resources types.Resources) error {
	oldResources := r.player.Resources

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

	// Update player resources (auto-saved, no explicit save call needed)
	r.player.Resources = resources

	// Publish single batched event if any resources changed
	if len(changes) > 0 {
		events.Publish(r.eventBus, events.ResourcesChangedEvent{
			GameID:    r.player.GameID,
			PlayerID:  r.player.ID,
			Changes:   changes,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateProduction updates player production
// Auto-saves changes to the player
func (r *ResourceRepository) UpdateProduction(ctx context.Context, production types.Production) error {
	r.player.Production = production
	return nil
}

// UpdateTerraformRating updates player terraform rating and publishes TerraformRatingChangedEvent
// Auto-saves changes to the player
func (r *ResourceRepository) UpdateTerraformRating(ctx context.Context, rating int) error {
	oldRating := r.player.TerraformRating
	r.player.TerraformRating = rating

	// Publish TerraformRatingChangedEvent
	if oldRating != rating {
		logResource.Debug("📡 Publishing TerraformRatingChangedEvent",
			zap.String("game_id", r.player.GameID),
			zap.String("player_id", r.player.ID),
			zap.Int("old_rating", oldRating),
			zap.Int("new_rating", rating))

		events.Publish(r.eventBus, events.TerraformRatingChangedEvent{
			GameID:    r.player.GameID,
			PlayerID:  r.player.ID,
			OldRating: oldRating,
			NewRating: rating,
			Timestamp: time.Now(),
		})
	}

	return nil
}

// UpdateVictoryPoints updates player victory points
// Auto-saves changes to the player
func (r *ResourceRepository) UpdateVictoryPoints(ctx context.Context, victoryPoints int) error {
	r.player.VictoryPoints = victoryPoints
	return nil
}

// UpdateStorage updates player resource storage (cards that store resources)
// Auto-saves changes to the player
func (r *ResourceRepository) UpdateStorage(ctx context.Context, storage map[string]int) error {
	r.player.ResourceStorage = storage
	return nil
}

// UpdatePaymentSubstitutes updates player payment substitutes
// Auto-saves changes to the player
func (r *ResourceRepository) UpdatePaymentSubstitutes(ctx context.Context, substitutes []types.PaymentSubstitute) error {
	r.player.PaymentSubstitutes = substitutes
	return nil
}
