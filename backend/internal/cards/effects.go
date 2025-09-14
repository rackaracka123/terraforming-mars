package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// EffectProcessor handles applying card effects to the game state
type EffectProcessor struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
}

// NewEffectProcessor creates a new card effect processor
func NewEffectProcessor(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) *EffectProcessor {
	return &EffectProcessor{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
	}
}

// ApplyCardEffects applies the effects of a played card to the game state
func (e *EffectProcessor) ApplyCardEffects(ctx context.Context, gameID, playerID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🎭 Applying card effects", zap.String("card_name", card.Name))

	// Apply production effects if the card has them
	if err := e.applyProductionEffects(ctx, gameID, playerID, card); err != nil {
		return fmt.Errorf("failed to apply production effects: %w", err)
	}

	// Future implementation: Apply immediate resource effects
	// if err := e.applyResourceEffects(ctx, gameID, playerID, card); err != nil {
	//     return fmt.Errorf("failed to apply resource effects: %w", err)
	// }

	// Future implementation: Apply global parameter effects
	// if err := e.applyGlobalParameterEffects(ctx, gameID, card); err != nil {
	//     return fmt.Errorf("failed to apply global parameter effects: %w", err)
	// }

	log.Debug("✅ Card effects applied successfully")
	return nil
}

// applyProductionEffects applies production changes from a card
func (e *EffectProcessor) applyProductionEffects(ctx context.Context, gameID, playerID string, card *model.Card) error {
	if card.ProductionEffects == nil {
		return nil // No production effects to apply
	}

	log := logger.WithGameContext(gameID, playerID)

	// Get current player to read current production
	player, err := e.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for production update: %w", err)
	}

	// Calculate new production values
	newProduction := player.Production
	newProduction.Credits += card.ProductionEffects.Credits
	newProduction.Steel += card.ProductionEffects.Steel
	newProduction.Titanium += card.ProductionEffects.Titanium
	newProduction.Plants += card.ProductionEffects.Plants
	newProduction.Energy += card.ProductionEffects.Energy
	newProduction.Heat += card.ProductionEffects.Heat

	// Ensure production values don't go below zero
	if newProduction.Credits < 0 {
		newProduction.Credits = 0
	}
	if newProduction.Steel < 0 {
		newProduction.Steel = 0
	}
	if newProduction.Titanium < 0 {
		newProduction.Titanium = 0
	}
	if newProduction.Plants < 0 {
		newProduction.Plants = 0
	}
	if newProduction.Energy < 0 {
		newProduction.Energy = 0
	}
	if newProduction.Heat < 0 {
		newProduction.Heat = 0
	}

	// Update player production
	if err := e.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
		log.Error("Failed to update player production", zap.Error(err))
		return fmt.Errorf("failed to update player production: %w", err)
	}

	log.Debug("📈 Production effects applied",
		zap.Int("credits_change", card.ProductionEffects.Credits),
		zap.Int("steel_change", card.ProductionEffects.Steel),
		zap.Int("titanium_change", card.ProductionEffects.Titanium),
		zap.Int("plants_change", card.ProductionEffects.Plants),
		zap.Int("energy_change", card.ProductionEffects.Energy),
		zap.Int("heat_change", card.ProductionEffects.Heat))

	return nil
}

// Future expansion: Additional effect types can be implemented here
// Examples:
// - applyResourceEffects: immediate resource gains/losses
// - applyGlobalParameterEffects: temperature, oxygen, ocean changes
// - applySpecialEffects: unique card abilities
// - applyTileEffects: board tile placements
// These will be added when the card model is extended with the corresponding fields
