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
// choiceIndex is optional and used when the card has choices between different effects
// cardStorageTarget is optional and used when outputs target "any-card" storage
func (cp *CardProcessor) ApplyCardEffects(ctx context.Context, gameID, playerID string, card *model.Card, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🎭 Applying card effects", zap.String("card_name", card.Name))

	// Apply production effects if the card has them
	if err := cp.applyProductionEffects(ctx, gameID, playerID, card, choiceIndex); err != nil {
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
	if err := cp.extractAndAddManualActions(ctx, gameID, playerID, card, choiceIndex); err != nil {
		return fmt.Errorf("failed to extract manual actions: %w", err)
	}

	// Extract and add passive effects from card behaviors
	if err := cp.extractAndAddEffects(ctx, gameID, playerID, card); err != nil {
		return fmt.Errorf("failed to extract effects: %w", err)
	}

	// Apply victory point conditions
	if err := cp.applyVictoryPointConditions(ctx, gameID, playerID, card); err != nil {
		return fmt.Errorf("failed to apply victory point conditions: %w", err)
	}

	// Apply immediate resource effects including card storage
	if err := cp.applyResourceEffects(ctx, gameID, playerID, card, choiceIndex, cardStorageTarget); err != nil {
		return fmt.Errorf("failed to apply resource effects: %w", err)
	}

	// Apply tile placement effects
	if err := cp.applyTileEffects(ctx, gameID, playerID, card); err != nil {
		return fmt.Errorf("failed to apply tile effects: %w", err)
	}

	// Apply global parameter effects (temperature, oxygen, oceans)
	if err := cp.applyGlobalParameterEffects(ctx, gameID, playerID, card, choiceIndex); err != nil {
		return fmt.Errorf("failed to apply global parameter effects: %w", err)
	}

	log.Debug("✅ Card effects applied successfully")
	return nil
}

// applyProductionEffects applies production changes from a card's behaviors
// choiceIndex is optional and used when the card has choices between different effects
func (cp *CardProcessor) applyProductionEffects(ctx context.Context, gameID, playerID string, card *model.Card, choiceIndex *int) error {
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
		// Only process auto triggers WITHOUT conditions (immediate effects when card is played)
		// Auto triggers WITH conditions are passive effects, not immediate production effects
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == model.ResourceTriggerAuto && behavior.Triggers[0].Condition == nil {
			// Aggregate all outputs: behavior.Outputs + choice[choiceIndex].Outputs
			allOutputs := behavior.Outputs

			// If choiceIndex is provided and this behavior has choices, add choice outputs
			if choiceIndex != nil && len(behavior.Choices) > 0 && *choiceIndex < len(behavior.Choices) {
				selectedChoice := behavior.Choices[*choiceIndex]
				allOutputs = append(allOutputs, selectedChoice.Outputs...)
				log.Debug("🎯 Applying choice outputs for production",
					zap.Int("choice_index", *choiceIndex),
					zap.Int("choice_outputs_count", len(selectedChoice.Outputs)))
			}

			// Process all aggregated outputs
			for _, output := range allOutputs {
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
// choiceIndex is optional - manual actions can also have choices that need to be resolved when the action is played
func (cp *CardProcessor) extractAndAddManualActions(ctx context.Context, gameID, playerID string, card *model.Card, choiceIndex *int) error {
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
		// Note: Manual actions can have their own choices, which are independent from the choice made when playing the card
		// The choiceIndex parameter here is for auto-triggered effects, not for manual actions
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

// extractAndAddEffects extracts passive effects (auto-triggered with conditions) from card behaviors and adds them to the player
// These are effects like "when a city is placed, gain 2 MC" that trigger automatically on game events
func (cp *CardProcessor) extractAndAddEffects(ctx context.Context, gameID, playerID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)

	// Track passive effects found
	var passiveEffects []model.PlayerEffect

	// Process all behaviors to find auto triggers with conditions
	for behaviorIndex, behavior := range card.Behaviors {
		// Check if this behavior has auto triggers WITH a condition
		// (auto triggers without conditions are immediate effects, not passive effects)
		hasConditionalAutoTrigger := false
		for _, trigger := range behavior.Triggers {
			if trigger.Type == model.ResourceTriggerAuto && trigger.Condition != nil {
				hasConditionalAutoTrigger = true
				break
			}
		}

		// If behavior has conditional auto triggers, create a PlayerEffect
		if hasConditionalAutoTrigger {
			effect := model.PlayerEffect{
				CardID:        card.ID,
				CardName:      card.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			}
			passiveEffects = append(passiveEffects, effect)

			log.Debug("✨ Found passive effect",
				zap.String("card_name", card.Name),
				zap.Int("behavior_index", behaviorIndex),
				zap.String("trigger_type", string(behavior.Triggers[0].Condition.Type)))
		}
	}

	// If passive effects were found, add them to the player
	if len(passiveEffects) > 0 {
		// Get current player to read current effects
		player, err := cp.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for effects update: %w", err)
		}

		// Create new effects slice with existing effects plus new passive effects
		newEffects := make([]model.PlayerEffect, len(player.Effects)+len(passiveEffects))
		copy(newEffects, player.Effects)
		copy(newEffects[len(player.Effects):], passiveEffects)

		// Update player effects via repository
		if err := cp.playerRepo.UpdateEffects(ctx, gameID, playerID, newEffects); err != nil {
			log.Error("Failed to update player passive effects", zap.Error(err))
			return fmt.Errorf("failed to update player passive effects: %w", err)
		}

		log.Debug("🎆 Passive effects added",
			zap.Int("effects_count", len(passiveEffects)),
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
// choiceIndex is optional and used when the card has choices between different effects
// cardStorageTarget is optional and used when outputs target "any-card" storage
func (cp *CardProcessor) applyResourceEffects(ctx context.Context, gameID, playerID string, card *model.Card, choiceIndex *int, cardStorageTarget *string) error {
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
	var trChange int

	// Process all behaviors to find resource effects
	for _, behavior := range card.Behaviors {
		// Only process auto triggers WITHOUT conditions (immediate effects when card is played)
		// Auto triggers WITH conditions are passive effects, not immediate resource effects
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == model.ResourceTriggerAuto && behavior.Triggers[0].Condition == nil {
			// Aggregate all outputs: behavior.Outputs + choice[choiceIndex].Outputs
			allOutputs := behavior.Outputs

			// If choiceIndex is provided and this behavior has choices, add choice outputs
			if choiceIndex != nil && len(behavior.Choices) > 0 && *choiceIndex < len(behavior.Choices) {
				selectedChoice := behavior.Choices[*choiceIndex]
				allOutputs = append(allOutputs, selectedChoice.Outputs...)
				log.Debug("🎯 Applying choice outputs for resources",
					zap.Int("choice_index", *choiceIndex),
					zap.Int("choice_outputs_count", len(selectedChoice.Outputs)))
			}

			// Process all aggregated outputs
			for _, output := range allOutputs {
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
				case model.ResourceTR:
					trChange += output.Amount

				// Card storage resources (animals, microbes, floaters, science, asteroid)
				case model.ResourceAnimals, model.ResourceMicrobes, model.ResourceFloaters, model.ResourceScience, model.ResourceAsteroid:
					// Handle card storage resources
					if err := cp.ApplyCardStorageResource(ctx, gameID, playerID, card.ID, output, cardStorageTarget, log); err != nil {
						return fmt.Errorf("failed to apply card storage resource: %w", err)
					}
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

	// Apply TR changes separately (not part of resources)
	if trChange != 0 {
		newTR := player.TerraformRating + trChange
		if err := cp.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR); err != nil {
			log.Error("Failed to update terraform rating", zap.Error(err))
			return fmt.Errorf("failed to update terraform rating: %w", err)
		}

		log.Info("⭐ Terraform Rating changed",
			zap.String("card_name", card.Name),
			zap.Int("change", trChange),
			zap.Int("old_tr", player.TerraformRating),
			zap.Int("new_tr", newTR))
	}

	return nil
}

// applyTileEffects handles tile placement effects from card behaviors
func (cp *CardProcessor) applyTileEffects(ctx context.Context, gameID, playerID string, card *model.Card) error {
	// Collect all tile placements from card behaviors
	var tilePlacementQueue []string

	// Process all behaviors to find tile placement effects
	for _, behavior := range card.Behaviors {
		// Only process auto triggers WITHOUT conditions (immediate effects when card is played)
		// Auto triggers WITH conditions are passive effects, not immediate tile placement effects
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == model.ResourceTriggerAuto && behavior.Triggers[0].Condition == nil {
			for _, output := range behavior.Outputs {
				switch output.Type {
				case model.ResourceCityPlacement:
					// Add each city placement to the queue individually
					for i := 0; i < output.Amount; i++ {
						tilePlacementQueue = append(tilePlacementQueue, "city")
					}
				case model.ResourceOceanPlacement:
					// Add each ocean placement to the queue individually
					for i := 0; i < output.Amount; i++ {
						tilePlacementQueue = append(tilePlacementQueue, "ocean")
					}
				case model.ResourceGreeneryPlacement:
					// Add each greenery placement to the queue individually
					for i := 0; i < output.Amount; i++ {
						tilePlacementQueue = append(tilePlacementQueue, "greenery")
					}
				}
			}
		}
	}

	// Delegate to PlayerRepository to handle the queue creation and processing
	return cp.playerRepo.CreateTileQueue(ctx, gameID, playerID, card.ID, tilePlacementQueue)
}

// ApplyCardStorageResource handles adding resources to card storage (animals, microbes, floaters, science)
// This is exported so it can be used by CardService for both card play and card actions
func (cp *CardProcessor) ApplyCardStorageResource(ctx context.Context, gameID, playerID, playedCardID string, output model.ResourceCondition, cardStorageTarget *string, log *zap.Logger) error {
	// Get current player to access resource storage
	player, err := cp.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for card storage update: %w", err)
	}

	// Initialize resource storage map if nil
	if player.ResourceStorage == nil {
		player.ResourceStorage = make(map[string]int)
	}

	// Determine target card based on output.Target
	var targetCardID string
	switch output.Target {
	case model.TargetSelfCard:
		// Target is the card itself (the card being played)
		targetCardID = playedCardID
		log.Debug("💾 Applying card storage to self-card",
			zap.String("target_card_id", targetCardID),
			zap.String("resource_type", string(output.Type)),
			zap.Int("amount", output.Amount))

	case model.TargetAnyCard:
		// Target is any card - if not provided or empty, resources will be discarded
		if cardStorageTarget == nil || *cardStorageTarget == "" {
			log.Info("⚠️ No card storage target provided - resources will be lost",
				zap.String("resource_type", string(output.Type)),
				zap.Int("amount", output.Amount))
			return nil // Successfully handled - resources are discarded
		}

		targetCardID = *cardStorageTarget

		// Validate that target card exists in player's played cards
		targetCardExists := false
		for _, playedCardID := range player.PlayedCards {
			if playedCardID == targetCardID {
				targetCardExists = true
				break
			}
		}
		if !targetCardExists {
			return fmt.Errorf("target card %s not found in player's played cards", targetCardID)
		}

		// Note: We skip storage type validation here as we'd need cardRepo access
		// The frontend filters cards to show only valid storage targets
		log.Debug("💾 Applying card storage to any-card",
			zap.String("target_card_id", targetCardID),
			zap.String("resource_type", string(output.Type)),
			zap.Int("amount", output.Amount))

	default:
		// Other targets are not valid for card storage
		return fmt.Errorf("invalid target type for card storage: %s", output.Target)
	}

	// Update the resource storage for the target card
	player.ResourceStorage[targetCardID] += output.Amount

	// Persist the updated resource storage
	if err := cp.playerRepo.UpdateResourceStorage(ctx, gameID, playerID, player.ResourceStorage); err != nil {
		log.Error("Failed to update card resource storage", zap.Error(err))
		return fmt.Errorf("failed to update card resource storage: %w", err)
	}

	log.Info("✅ Card storage resource applied",
		zap.String("target_card_id", targetCardID),
		zap.String("resource_type", string(output.Type)),
		zap.Int("amount", output.Amount),
		zap.Int("new_storage_amount", player.ResourceStorage[targetCardID]))

	return nil
}

// applyGlobalParameterEffects applies global parameter changes (temperature, oxygen, oceans) from a card's behaviors
// choiceIndex is optional and used when the card has choices between different effects
func (cp *CardProcessor) applyGlobalParameterEffects(ctx context.Context, gameID, playerID string, card *model.Card, choiceIndex *int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current game to read current global parameters
	game, err := cp.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game for global parameter update: %w", err)
	}

	// Track changes for logging
	var temperatureChange, oxygenChange, oceansChange int

	// Process all behaviors to find global parameter effects
	for _, behavior := range card.Behaviors {
		// Only process auto triggers WITHOUT conditions (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == model.ResourceTriggerAuto && behavior.Triggers[0].Condition == nil {
			// Aggregate all outputs: behavior.Outputs + choice[choiceIndex].Outputs
			allOutputs := behavior.Outputs

			// If choiceIndex is provided and this behavior has choices, add choice outputs
			if choiceIndex != nil && len(behavior.Choices) > 0 && *choiceIndex < len(behavior.Choices) {
				selectedChoice := behavior.Choices[*choiceIndex]
				allOutputs = append(allOutputs, selectedChoice.Outputs...)
			}

			// Process all aggregated outputs
			for _, output := range allOutputs {
				switch output.Type {
				case model.ResourceTemperature:
					temperatureChange += output.Amount
				case model.ResourceOxygen:
					oxygenChange += output.Amount
				case model.ResourceOceans:
					oceansChange += output.Amount
				}
			}
		}
	}

	// Apply temperature changes
	if temperatureChange != 0 {
		newTemperature := game.GlobalParameters.Temperature + temperatureChange
		// Clamp to valid range
		if newTemperature > model.MaxTemperature {
			newTemperature = model.MaxTemperature
		}
		if newTemperature < model.MinTemperature {
			newTemperature = model.MinTemperature
		}

		if err := cp.gameRepo.UpdateTemperature(ctx, gameID, newTemperature); err != nil {
			log.Error("Failed to update temperature", zap.Error(err))
			return fmt.Errorf("failed to update temperature: %w", err)
		}

		log.Info("🌡️ Temperature changed",
			zap.String("card_name", card.Name),
			zap.Int("change", temperatureChange),
			zap.Int("old_temperature", game.GlobalParameters.Temperature),
			zap.Int("new_temperature", newTemperature))
	}

	// Apply oxygen changes
	if oxygenChange != 0 {
		newOxygen := game.GlobalParameters.Oxygen + oxygenChange
		// Clamp to valid range
		if newOxygen > model.MaxOxygen {
			newOxygen = model.MaxOxygen
		}
		if newOxygen < model.MinOxygen {
			newOxygen = model.MinOxygen
		}

		if err := cp.gameRepo.UpdateOxygen(ctx, gameID, newOxygen); err != nil {
			log.Error("Failed to update oxygen", zap.Error(err))
			return fmt.Errorf("failed to update oxygen: %w", err)
		}

		log.Info("💨 Oxygen changed",
			zap.String("card_name", card.Name),
			zap.Int("change", oxygenChange),
			zap.Int("old_oxygen", game.GlobalParameters.Oxygen),
			zap.Int("new_oxygen", newOxygen))
	}

	// Apply oceans changes
	if oceansChange != 0 {
		newOceans := game.GlobalParameters.Oceans + oceansChange
		// Clamp to valid range
		if newOceans > model.MaxOceans {
			newOceans = model.MaxOceans
		}
		if newOceans < model.MinOceans {
			newOceans = model.MinOceans
		}

		if err := cp.gameRepo.UpdateOceans(ctx, gameID, newOceans); err != nil {
			log.Error("Failed to update oceans", zap.Error(err))
			return fmt.Errorf("failed to update oceans: %w", err)
		}

		log.Info("🌊 Oceans changed",
			zap.String("card_name", card.Name),
			zap.Int("change", oceansChange),
			zap.Int("old_oceans", game.GlobalParameters.Oceans),
			zap.Int("new_oceans", newOceans))
	}

	return nil
}
