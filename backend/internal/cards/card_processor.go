package cards

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// CardProcessor handles the complete card processing including validation and effect application
type CardProcessor struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
}

// NewCardProcessor creates a new card processor
func NewCardProcessor(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) *CardProcessor {
	return &CardProcessor{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
	}
}

// ApplyCardEffects applies all effects when a card is played to the game state
// This function assumes all requirements and affordability checks have already been validated
func (cp *CardProcessor) ApplyCardEffects(ctx context.Context, gameID, playerID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🎭 Applying card effects", zap.String("card_name", card.Name))

	// Apply production effects if the card has them
	if err := cp.applyProductionEffects(ctx, gameID, playerID, card); err != nil {
		return fmt.Errorf("failed to apply production effects: %w", err)
	}

	// TODO: Apply discount effects if the card has them (implement later)
	// if err := cp.applyDiscountEffects(ctx, gameID, playerID, card); err != nil {
	//     return fmt.Errorf("failed to apply discount effects: %w", err)
	// }

	// TODO: Apply global parameter lenience effects (implement later - too complex for now)
	// if err := cp.applyGlobalParameterLenienceEffects(ctx, gameID, playerID, card); err != nil {
	//     return fmt.Errorf("failed to apply global parameter lenience effects: %w", err)
	// }

	// Extract and add manual actions from card behaviors
	if err := cp.extractAndAddManualActions(ctx, gameID, playerID, card); err != nil {
		return fmt.Errorf("failed to extract manual actions: %w", err)
	}

	// Apply victory point conditions
	if err := cp.applyVictoryPointConditions(ctx, gameID, playerID, card); err != nil {
		return fmt.Errorf("failed to apply victory point conditions: %w", err)
	}

	// Apply immediate resource effects
	if err := cp.applyResourceEffects(ctx, gameID, playerID, card); err != nil {
		return fmt.Errorf("failed to apply resource effects: %w", err)
	}

	// Future implementation: Apply global parameter effects
	// if err := cp.applyGlobalParameterEffects(ctx, gameID, card); err != nil {
	//     return fmt.Errorf("failed to apply global parameter effects: %w", err)
	// }

	log.Debug("✅ Card effects applied successfully")
	return nil
}

// applyProductionEffects applies production changes from a card's behaviors
func (cp *CardProcessor) applyProductionEffects(ctx context.Context, gameID, playerID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to read current production
	player, err := cp.playerRepo.GetByID(ctx, gameID, playerID)
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

	// Note: Validation that production reductions don't go below minimum values
	// should be done by the requirements validator before this function is called

	// Update player production
	if err := cp.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
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



// extractAndAddManualActions extracts manual actions from card behaviors and adds them to the player
func (cp *CardProcessor) extractAndAddManualActions(ctx context.Context, gameID, playerID string, card *model.Card) error {
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
		player, err := cp.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for manual actions update: %w", err)
		}

		// Create new actions slice with existing actions plus new manual actions
		newActions := make([]model.PlayerAction, len(player.Actions)+len(manualActions))
		copy(newActions, player.Actions)
		copy(newActions[len(player.Actions):], manualActions)

		// Update player actions via repository
		if err := cp.playerRepo.UpdatePlayerActions(ctx, gameID, playerID, newActions); err != nil {
			log.Error("Failed to update player manual actions", zap.Error(err))
			return fmt.Errorf("failed to update player manual actions: %w", err)
		}

		log.Debug("⚡ Manual actions added",
			zap.Int("actions_count", len(manualActions)),
			zap.String("card_name", card.Name))
	}

	return nil
}

// applyVictoryPointConditions applies victory point conditions from a card
func (cp *CardProcessor) applyVictoryPointConditions(ctx context.Context, gameID, playerID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)

	// Check if card has VP conditions
	if len(card.VPConditions) == 0 {
		return nil // No VP conditions to process
	}

	// Get current player
	player, err := cp.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for VP update: %w", err)
	}

	var totalVPAwarded int

	// Process each VP condition
	for _, vpCondition := range card.VPConditions {
		var vpAwarded int

		switch vpCondition.Condition {
		case model.VPConditionFixed:
			// Fixed VP - award immediately
			vpAwarded = vpCondition.Amount
			log.Debug("🏆 Fixed VP condition found",
				zap.Int("vp_amount", vpAwarded),
				zap.String("card_name", card.Name))

		case model.VPConditionOnce:
			// TODO: Implement once conditions (triggered when condition is met)
			log.Debug("⚠️ VP Once condition not yet implemented",
				zap.String("card_name", card.Name))
			continue

		case model.VPConditionPer:
			// TODO: Implement per conditions (VP per resource/tag/etc)
			log.Debug("⚠️ VP Per condition not yet implemented",
				zap.String("card_name", card.Name))
			continue

		default:
			log.Warn("❌ Unknown VP condition type",
				zap.String("condition", string(vpCondition.Condition)),
				zap.String("card_name", card.Name))
			continue
		}

		totalVPAwarded += vpAwarded
	}

	// Update player's victory points if any were awarded
	if totalVPAwarded > 0 {
		newVictoryPoints := player.VictoryPoints + totalVPAwarded
		if err := cp.playerRepo.UpdateVictoryPoints(ctx, gameID, playerID, newVictoryPoints); err != nil {
			log.Error("Failed to update player victory points", zap.Error(err))
			return fmt.Errorf("failed to update player victory points: %w", err)
		}

		log.Info("🏆 Victory Points awarded",
			zap.String("card_name", card.Name),
			zap.Int("vp_awarded", totalVPAwarded),
			zap.Int("total_vp", newVictoryPoints))
	}

	return nil
}

// applyResourceEffects applies immediate resource gains/losses from a card's behaviors
func (cp *CardProcessor) applyResourceEffects(ctx context.Context, gameID, playerID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to read current resources
	player, err := cp.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for resource update: %w", err)
	}

	// Calculate new resource values from card behaviors
	newResources := player.Resources

	// Track changes for logging
	var creditsChange, steelChange, titaniumChange, plantsChange, energyChange, heatChange int

	// Process all behaviors to find resource effects
	for _, behavior := range card.Behaviors {
		// Only process auto triggers (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == model.ResourceTriggerAuto {
			for _, output := range behavior.Outputs {
				switch output.Type {
				case model.ResourceCredits:
					newResources.Credits += output.Amount
					creditsChange += output.Amount
				case model.ResourceSteel:
					newResources.Steel += output.Amount
					steelChange += output.Amount
				case model.ResourceTitanium:
					newResources.Titanium += output.Amount
					titaniumChange += output.Amount
				case model.ResourcePlants:
					newResources.Plants += output.Amount
					plantsChange += output.Amount
				case model.ResourceEnergy:
					newResources.Energy += output.Amount
					energyChange += output.Amount
				case model.ResourceHeat:
					newResources.Heat += output.Amount
					heatChange += output.Amount
				}
			}
		}
	}

	// Note: Validation that player can afford resource deductions should be done
	// by the requirements validator before this function is called

	// Only update if there are any changes
	if creditsChange != 0 || steelChange != 0 || titaniumChange != 0 || plantsChange != 0 || energyChange != 0 || heatChange != 0 {
		// Update player resources
		if err := cp.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
			log.Error("Failed to update player resources", zap.Error(err))
			return fmt.Errorf("failed to update player resources: %w", err)
		}

		log.Debug("💰 Resource effects applied",
			zap.String("card_name", card.Name),
			zap.Int("credits_change", creditsChange),
			zap.Int("steel_change", steelChange),
			zap.Int("titanium_change", titaniumChange),
			zap.Int("plants_change", plantsChange),
			zap.Int("energy_change", energyChange),
			zap.Int("heat_change", heatChange))
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
