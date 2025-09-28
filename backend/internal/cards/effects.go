package cards

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
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

	// Apply discount effects if the card has them
	if err := e.applyDiscountEffects(ctx, gameID, playerID, card); err != nil {
		return fmt.Errorf("failed to apply discount effects: %w", err)
	}

	// Apply global parameter lenience effects if the card has them
	if err := e.applyGlobalParameterLenienceEffects(ctx, gameID, playerID, card); err != nil {
		return fmt.Errorf("failed to apply global parameter lenience effects: %w", err)
	}

	// Extract and add manual actions from card behaviors
	if err := e.extractAndAddManualActions(ctx, gameID, playerID, card); err != nil {
		return fmt.Errorf("failed to extract manual actions: %w", err)
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

// applyProductionEffects applies production changes from a card's behaviors
func (e *EffectProcessor) applyProductionEffects(ctx context.Context, gameID, playerID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to read current production
	player, err := e.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for production update: %w", err)
	}

	// Calculate new production values from card behaviors
	newProduction := player.Production

	// Track changes for logging
	var creditsChange, steelChange, titaniumChange, plantsChange, energyChange, heatChange int

	// Process all behaviors to find production effects
	for _, behavior := range card.Behaviors {
		// Only process auto triggers (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == model.ResourceTriggerAuto {
			for _, output := range behavior.Outputs {
				switch output.Type {
				case model.ResourceCreditsProduction:
					newProduction.Credits += output.Amount
					creditsChange += output.Amount
				case model.ResourceSteelProduction:
					newProduction.Steel += output.Amount
					steelChange += output.Amount
				case model.ResourceTitaniumProduction:
					newProduction.Titanium += output.Amount
					titaniumChange += output.Amount
				case model.ResourcePlantsProduction:
					newProduction.Plants += output.Amount
					plantsChange += output.Amount
				case model.ResourceEnergyProduction:
					newProduction.Energy += output.Amount
					energyChange += output.Amount
				case model.ResourceHeatProduction:
					newProduction.Heat += output.Amount
					heatChange += output.Amount
				}
			}
		}
	}

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
		zap.Int("credits_change", creditsChange),
		zap.Int("steel_change", steelChange),
		zap.Int("titanium_change", titaniumChange),
		zap.Int("plants_change", plantsChange),
		zap.Int("energy_change", energyChange),
		zap.Int("heat_change", heatChange))

	return nil
}

// applyDiscountEffects applies discount effects from a card's behaviors to the player's effects list
func (e *EffectProcessor) applyDiscountEffects(ctx context.Context, gameID, playerID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to read current effects
	player, err := e.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for discount effects update: %w", err)
	}

	// Track if any discount effects were found
	var discountEffectsFound []model.PlayerEffect

	// Process all behaviors to find discount effects
	for behaviorIndex, behavior := range card.Behaviors {
		// Only process auto triggers (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == model.ResourceTriggerAuto {
			for _, output := range behavior.Outputs {
				if output.Type == model.ResourceDiscount {
					// Create effect behavior that represents this discount
					effectBehavior := model.CardBehavior{
						Triggers: []model.Trigger{
							{
								Type: model.ResourceTriggerAuto,
							},
						},
						Outputs: []model.ResourceCondition{
							{
								Type:         model.ResourceDiscount,
								Amount:       output.Amount,
								Target:       model.TargetSelfPlayer,
								AffectedTags: output.AffectedTags,
							},
						},
					}

					discountEffect := model.PlayerEffect{
						CardID:        card.ID,
						CardName:      card.Name,
						BehaviorIndex: behaviorIndex,
						Behavior:      effectBehavior,
					}
					discountEffectsFound = append(discountEffectsFound, discountEffect)

					log.Debug("💰 Found discount effect",
						zap.String("card_name", card.Name),
						zap.Int("behavior_index", behaviorIndex),
						zap.Int("amount", output.Amount),
						zap.Any("affected_tags", output.AffectedTags))
				}
			}
		}
	}

	// If discount effects were found, add them to player's effects list
	if len(discountEffectsFound) > 0 {
		// Create new effects slice with existing effects plus new discount effects
		newEffects := make([]model.PlayerEffect, len(player.Effects)+len(discountEffectsFound))
		copy(newEffects, player.Effects)
		copy(newEffects[len(player.Effects):], discountEffectsFound)

		// Update player effects via repository
		if err := e.playerRepo.UpdateEffects(ctx, gameID, playerID, newEffects); err != nil {
			log.Error("Failed to update player discount effects", zap.Error(err))
			return fmt.Errorf("failed to update player discount effects: %w", err)
		}

		log.Debug("✨ Discount effects applied",
			zap.Int("total_effects_count", len(discountEffectsFound)))
	}

	return nil
}

// applyGlobalParameterLenienceEffects applies global parameter lenience effects from a card's behaviors to the player's effects list
func (e *EffectProcessor) applyGlobalParameterLenienceEffects(ctx context.Context, gameID, playerID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to read current effects
	player, err := e.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for global parameter lenience effects update: %w", err)
	}

	// Track if any global parameter lenience effects were found
	var lenienceEffectsFound []model.PlayerEffect

	// Process all behaviors to find global parameter lenience effects
	for behaviorIndex, behavior := range card.Behaviors {
		// Only process auto triggers (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == model.ResourceTriggerAuto {
			for _, output := range behavior.Outputs {
				if output.Type == model.ResourceGlobalParameterLenience {
					// Create effect behavior that represents this global parameter lenience
					effectBehavior := model.CardBehavior{
						Triggers: []model.Trigger{
							{
								Type: model.ResourceTriggerAuto,
							},
						},
						Outputs: []model.ResourceCondition{
							{
								Type:         model.ResourceGlobalParameterLenience,
								Amount:       output.Amount,
								Target:       model.TargetSelfPlayer,
								AffectedTags: output.AffectedTags,
							},
						},
					}

					lenienceEffect := model.PlayerEffect{
						CardID:        card.ID,
						CardName:      card.Name,
						BehaviorIndex: behaviorIndex,
						Behavior:      effectBehavior,
					}
					lenienceEffectsFound = append(lenienceEffectsFound, lenienceEffect)

					log.Debug("🎯 Found global parameter lenience effect",
						zap.String("card_name", card.Name),
						zap.Int("behavior_index", behaviorIndex),
						zap.Int("amount", output.Amount),
						zap.Any("affected_tags", output.AffectedTags))
				}
			}
		}
	}

	// If global parameter lenience effects were found, add them to player's effects list
	if len(lenienceEffectsFound) > 0 {
		// Create new effects slice with existing effects plus new lenience effects
		newEffects := make([]model.PlayerEffect, len(player.Effects)+len(lenienceEffectsFound))
		copy(newEffects, player.Effects)
		copy(newEffects[len(player.Effects):], lenienceEffectsFound)

		// Update player effects via repository
		if err := e.playerRepo.UpdateEffects(ctx, gameID, playerID, newEffects); err != nil {
			log.Error("Failed to update player global parameter lenience effects", zap.Error(err))
			return fmt.Errorf("failed to update player global parameter lenience effects: %w", err)
		}

		log.Debug("✨ Global parameter lenience effects applied",
			zap.Int("total_effects_count", len(lenienceEffectsFound)))
	}

	return nil
}

// extractAndAddManualActions extracts manual actions from card behaviors and adds them to the player
func (e *EffectProcessor) extractAndAddManualActions(ctx context.Context, gameID, playerID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)

	// Track manual actions found
	var manualActions []model.PlayerAction

	// Process all behaviors to find manual triggers
	for behaviorIndex, behavior := range card.Behaviors {
		// Check if this behavior has manual triggers
		hasManualTrigger := false
		for _, trigger := range behavior.Triggers {
			if trigger.Type == model.ResourceTriggerManual {
				hasManualTrigger = true
				break
			}
		}

		// If behavior has manual triggers, create a PlayerAction
		if hasManualTrigger {
			action := model.PlayerAction{
				CardID:        card.ID,
				CardName:      card.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			}
			manualActions = append(manualActions, action)

			log.Debug("🎯 Found manual action",
				zap.String("card_name", card.Name),
				zap.Int("behavior_index", behaviorIndex))
		}
	}

	// If manual actions were found, add them to the player
	if len(manualActions) > 0 {
		// Get current player to read current actions
		player, err := e.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for manual actions update: %w", err)
		}

		// Create new actions slice with existing actions plus new manual actions
		newActions := make([]model.PlayerAction, len(player.Actions)+len(manualActions))
		copy(newActions, player.Actions)
		copy(newActions[len(player.Actions):], manualActions)

		// Update player actions via repository
		if err := e.playerRepo.UpdatePlayerActions(ctx, gameID, playerID, newActions); err != nil {
			log.Error("Failed to update player manual actions", zap.Error(err))
			return fmt.Errorf("failed to update player manual actions: %w", err)
		}

		log.Debug("⚡ Manual actions added",
			zap.Int("actions_count", len(manualActions)),
			zap.String("card_name", card.Name))
	}

	return nil
}

// Future expansion: Additional effect types can be implemented here
// Examples:
// - applyResourceEffects: immediate resource gains/losses
// - applyGlobalParameterEffects: temperature, oxygen, ocean changes
// - applySpecialEffects: unique card abilities
// - applyTileEffects: board tile placements
// These will be added when the card model is extended with the corresponding fields
