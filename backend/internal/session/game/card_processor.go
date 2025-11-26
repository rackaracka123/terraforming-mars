package game

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/game/deck"
	"terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// CardProcessor handles the complete card processing including validation and effect application in session-scoped architecture
type CardProcessor struct {
	deckRepo deck.Repository
}

// NewCardProcessor creates a new card processor with session-scoped repositories
func NewCardProcessor(deckRepo deck.Repository) *CardProcessor {
	return &CardProcessor{
		deckRepo: deckRepo,
	}
}

// ApplyCardEffects applies all effects when a card is played to the game state
// This function assumes all requirements and affordability checks have already been validated
// choiceIndex is optional and used when the card has choices between different effects
// cardStorageTarget is optional and used when outputs target "any-card" storage
func (cp *CardProcessor) ApplyCardEffects(ctx context.Context, game *Game, p *player.Player, card *card.Card, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(p.GameID(), p.ID())
	log.Debug("🎭 Applying card effects", zap.String("card_name", card.Name))

	// Apply production effects if the card has them
	if err := cp.applyProductionEffects(ctx, p, card, choiceIndex); err != nil {
		return fmt.Errorf("failed to apply production effects: %w", err)
	}

	// Note: Discount effects are handled via Player.RequirementModifiers system
	// card.Card discounts are automatically applied during card play validation via CardManager

	// TODO: Apply global parameter lenience effects (implement later - complex feature)
	// Global parameter leniences allow cards to modify temperature/oxygen/ocean requirements
	// This is deferred until the requirements system is enhanced to handle leniences
	// if err := cp.applyGlobalParameterLenienceEffects(ctx, p, card); err != nil {
	//     return fmt.Errorf("failed to apply global parameter lenience effects: %w", err)
	// }

	// Extract and add manual actions from card behaviors
	if err := cp.extractAndAddManualActions(ctx, p, card, choiceIndex); err != nil {
		return fmt.Errorf("failed to extract manual actions: %w", err)
	}

	// Apply victory point conditions
	if err := cp.applyVictoryPointConditions(ctx, p, card); err != nil {
		return fmt.Errorf("failed to apply victory point conditions: %w", err)
	}

	// Apply immediate resource effects including card storage
	if err := cp.applyResourceEffects(ctx, p, card, choiceIndex, cardStorageTarget); err != nil {
		return fmt.Errorf("failed to apply resource effects: %w", err)
	}

	// Apply tile placement effects
	if err := cp.applyTileEffects(ctx, p, card); err != nil {
		return fmt.Errorf("failed to apply tile effects: %w", err)
	}

	// Apply card draw/peek/take/buy effects
	if err := cp.applyCardDrawPeekEffects(ctx, p, card); err != nil {
		return fmt.Errorf("failed to apply card draw/peek effects: %w", err)
	}

	// Apply global parameter effects (temperature, oxygen, oceans)
	if err := cp.applyGlobalParameterEffects(ctx, game, p, card, choiceIndex); err != nil {
		return fmt.Errorf("failed to apply global parameter effects: %w", err)
	}

	log.Debug("✅ card.Card effects applied successfully")
	return nil
}

// applyProductionEffects applies production changes from a card's behaviors
// choiceIndex is optional and used when the card has choices between different effects
func (cp *CardProcessor) applyProductionEffects(ctx context.Context, p *player.Player, c *card.Card, choiceIndex *int) error {
	log := logger.WithGameContext(p.GameID(), p.ID())

	// Calculate new production values from card behaviors
	newProduction := p.Production()

	// Track changes for logging
	var creditsChange, steelChange, titaniumChange, plantsChange, energyChange, heatChange int

	// Process all behaviors to find production effects
	for _, behavior := range c.Behaviors {
		// Only process auto triggers WITHOUT conditions (immediate effects when card is played)
		// Auto triggers WITH conditions are passive effects, not immediate production effects
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == card.ResourceTriggerAuto && behavior.Triggers[0].Condition == nil {
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
				case types.ResourceCreditsProduction, types.ResourceSteelProduction, types.ResourceTitaniumProduction,
					types.ResourcePlantsProduction, types.ResourceEnergyProduction, types.ResourceHeatProduction:
					// Apply production change directly based on type
					switch output.Type {
					case types.ResourceCreditsProduction:
						newProduction.Credits += output.Amount
					case types.ResourceSteelProduction:
						newProduction.Steel += output.Amount
					case types.ResourceTitaniumProduction:
						newProduction.Titanium += output.Amount
					case types.ResourcePlantsProduction:
						newProduction.Plants += output.Amount
					case types.ResourceEnergyProduction:
						newProduction.Energy += output.Amount
					case types.ResourceHeatProduction:
						newProduction.Heat += output.Amount
					}

					// Track changes for logging
					switch output.Type {
					case types.ResourceCreditsProduction:
						creditsChange += output.Amount
					case types.ResourceSteelProduction:
						steelChange += output.Amount
					case types.ResourceTitaniumProduction:
						titaniumChange += output.Amount
					case types.ResourcePlantsProduction:
						plantsChange += output.Amount
					case types.ResourceEnergyProduction:
						energyChange += output.Amount
					case types.ResourceHeatProduction:
						heatChange += output.Amount
					}
				}
			}
		}
	}

	// Note: Validation that production reductions don't go below minimum values
	// should be done by the requirements validator before this function is called

	// Update player production
	if err := p.SetProduction(ctx, newProduction); err != nil {
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
func (cp *CardProcessor) extractAndAddManualActions(ctx context.Context, p *player.Player, c *card.Card, choiceIndex *int) error {
	log := logger.WithGameContext(p.GameID(), p.ID())

	// Track manual actions found
	var manualActions []player.PlayerAction

	// Process all behaviors to find manual triggers
	for behaviorIndex, behavior := range c.Behaviors {
		// Check if this behavior has manual triggers
		hasManualTrigger := false
		for _, trigger := range behavior.Triggers {
			if trigger.Type == card.ResourceTriggerManual {
				hasManualTrigger = true
				break
			}
		}

		// If behavior has manual triggers, create a PlayerAction
		// Note: Manual actions can have their own choices, which are independent from the choice made when playing the card
		// The choiceIndex parameter here is for auto-triggered effects, not for manual actions
		if hasManualTrigger {
			action := player.PlayerAction{
				CardID:        c.ID,
				CardName:      c.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			}
			manualActions = append(manualActions, action)

			log.Debug("🎯 Found manual action",
				zap.String("card_name", c.Name),
				zap.Int("behavior_index", behaviorIndex))
		}
	}

	// If manual actions were found, add them to the player
	if len(manualActions) > 0 {
		// Get current player actions
		currentActions := p.Actions()

		// Create new actions slice with existing actions plus new manual actions
		newActions := make([]player.PlayerAction, len(currentActions)+len(manualActions))
		copy(newActions, currentActions)
		copy(newActions[len(currentActions):], manualActions)

		// Update player actions
		if err := p.SetActions(ctx, newActions); err != nil {
			log.Error("Failed to update player manual actions", zap.Error(err))
			return fmt.Errorf("failed to update player manual actions: %w", err)
		}

		log.Debug("⚡ Manual actions added",
			zap.Int("actions_count", len(manualActions)),
			zap.String("card_name", c.Name))
	}

	return nil
}

// applyVictoryPointConditions applies victory point conditions from a card
func (cp *CardProcessor) applyVictoryPointConditions(ctx context.Context, p *player.Player, c *card.Card) error {
	log := logger.WithGameContext(p.GameID(), p.ID())

	// Check if card has VP conditions
	if len(c.VPConditions) == 0 {
		return nil // No VP conditions to process
	}

	// Get current player

	var totalVPAwarded int

	// Process each VP condition
	for _, vpCondition := range c.VPConditions {
		var vpAwarded int

		switch vpCondition.Condition {
		case card.VPConditionFixed:
			// Fixed VP - award immediately
			vpAwarded = vpCondition.Amount
			log.Debug("🏆 Fixed VP condition found",
				zap.Int("vp_amount", vpAwarded),
				zap.String("card_name", c.Name))

		case card.VPConditionOnce:
			// TODO: Implement once conditions (triggered when condition is met)
			// Examples: "3 VP when temperature reaches 0°C", "2 VP when you have 8 plant production"
			// Requires: Event subscription system to detect when conditions are met
			log.Debug("⚠️  VP Once condition not yet implemented",
				zap.String("card_name", c.Name))
			continue

		case card.VPConditionPer:
			// TODO: Implement per conditions (VP per resource/tag/tile/etc)
			// Examples: "1 VP per jovian tag", "2 VP per ocean tile you own", "1 VP per 3 animals on this card"
			// Requires: Dynamic VP calculation at game end based on current game state
			log.Debug("⚠️  VP Per condition not yet implemented",
				zap.String("card_name", c.Name))
			continue

		default:
			log.Warn("❌ Unknown VP condition type",
				zap.String("condition", string(vpCondition.Condition)),
				zap.String("card_name", c.Name))
			continue
		}

		totalVPAwarded += vpAwarded
	}

	// Update player's victory points if any were awarded
	if totalVPAwarded > 0 {
		newVictoryPoints := p.VictoryPoints() + totalVPAwarded
		if err := p.SetVictoryPoints(ctx, newVictoryPoints); err != nil {
			log.Error("Failed to update player victory points", zap.Error(err))
			return fmt.Errorf("failed to update player victory points: %w", err)
		}

		log.Info("🏆 Victory Points awarded",
			zap.String("card_name", c.Name),
			zap.Int("vp_awarded", totalVPAwarded),
			zap.Int("total_vp", newVictoryPoints))
	}

	return nil
}

// applyResourceEffects applies immediate resource gains/losses from a card's behaviors
// choiceIndex is optional and used when the card has choices between different effects
// cardStorageTarget is optional and used when outputs target "any-card" storage
func (cp *CardProcessor) applyResourceEffects(ctx context.Context, p *player.Player, c *card.Card, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(p.GameID(), p.ID())

	// Get current player to read current resources

	// Calculate new resource values from card behaviors
	newResources := p.Resources()

	// Track changes for logging
	var creditsChange, steelChange, titaniumChange, plantsChange, energyChange, heatChange int
	var trChange int

	// Process all behaviors to find resource effects
	for _, behavior := range c.Behaviors {
		// Only process auto triggers WITHOUT conditions (immediate effects when card is played)
		// Auto triggers WITH conditions are passive effects, not immediate resource effects
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == card.ResourceTriggerAuto && behavior.Triggers[0].Condition == nil {
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
				case types.ResourceCredits, types.ResourceSteel, types.ResourceTitanium,
					types.ResourcePlants, types.ResourceEnergy, types.ResourceHeat:
					// Apply resource change directly based on type
					switch output.Type {
					case types.ResourceCredits:
						newResources.Credits += output.Amount
					case types.ResourceSteel:
						newResources.Steel += output.Amount
					case types.ResourceTitanium:
						newResources.Titanium += output.Amount
					case types.ResourcePlants:
						newResources.Plants += output.Amount
					case types.ResourceEnergy:
						newResources.Energy += output.Amount
					case types.ResourceHeat:
						newResources.Heat += output.Amount
					}

					// Track changes for logging
					switch output.Type {
					case types.ResourceCredits:
						creditsChange += output.Amount
					case types.ResourceSteel:
						steelChange += output.Amount
					case types.ResourceTitanium:
						titaniumChange += output.Amount
					case types.ResourcePlants:
						plantsChange += output.Amount
					case types.ResourceEnergy:
						energyChange += output.Amount
					case types.ResourceHeat:
						heatChange += output.Amount
					}

				case types.ResourceTR:
					trChange += output.Amount

				// card.Card storage resources (animals, microbes, floaters, science, asteroid)
				case types.ResourceAnimals, types.ResourceMicrobes, types.ResourceFloaters, types.ResourceScience, types.ResourceAsteroid:
					// Handle card storage resources
					if err := cp.ApplyCardStorageResource(ctx, p, c.ID, output, cardStorageTarget, log); err != nil {
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
		if err := p.SetResources(ctx, newResources); err != nil {
			log.Error("Failed to update player resources", zap.Error(err))
			return fmt.Errorf("failed to update player resources: %w", err)
		}

		log.Debug("💰 Resource effects applied",
			zap.String("card_name", c.Name),
			zap.Int("credits_change", creditsChange),
			zap.Int("steel_change", steelChange),
			zap.Int("titanium_change", titaniumChange),
			zap.Int("plants_change", plantsChange),
			zap.Int("energy_change", energyChange),
			zap.Int("heat_change", heatChange))
	}

	// Apply TR changes separately (not part of resources)
	if trChange != 0 {
		newTR := p.TerraformRating() + trChange
		if err := p.SetTerraformRating(ctx, newTR); err != nil {
			log.Error("Failed to update terraform rating", zap.Error(err))
			return fmt.Errorf("failed to update terraform rating: %w", err)
		}

		log.Info("⭐ Terraform Rating changed",
			zap.String("card_name", c.Name),
			zap.Int("change", trChange),
			zap.Int("terraform_rating", p.TerraformRating()),
			zap.Int("new_tr", newTR))
	}

	return nil
}

// applyTileEffects handles tile placement effects from card behaviors
func (cp *CardProcessor) applyTileEffects(ctx context.Context, p *player.Player, c *card.Card) error {
	// Collect all tile placements from card behaviors
	var tilePlacementQueue []string

	// Process all behaviors to find tile placement effects
	for _, behavior := range c.Behaviors {
		// Only process auto triggers WITHOUT conditions (immediate effects when card is played)
		// Auto triggers WITH conditions are passive effects, not immediate tile placement effects
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == card.ResourceTriggerAuto && behavior.Triggers[0].Condition == nil {
			for _, output := range behavior.Outputs {
				switch output.Type {
				case types.ResourceCityPlacement:
					// Add each city placement to the queue individually
					for i := 0; i < output.Amount; i++ {
						tilePlacementQueue = append(tilePlacementQueue, "city")
					}
				case types.ResourceOceanPlacement:
					// Add each ocean placement to the queue individually
					for i := 0; i < output.Amount; i++ {
						tilePlacementQueue = append(tilePlacementQueue, "ocean")
					}
				case types.ResourceGreeneryPlacement:
					// Add each greenery placement to the queue individually
					for i := 0; i < output.Amount; i++ {
						tilePlacementQueue = append(tilePlacementQueue, "greenery")
					}
				}
			}
		}
	}

	// Delegate to Player to handle the queue creation and processing
	return p.CreateTileQueue(ctx, c.ID, tilePlacementQueue)
}

// applyCardDrawPeekEffects handles card draw/peek/take/buy effects from card behaviors
func (cp *CardProcessor) applyCardDrawPeekEffects(ctx context.Context, p *player.Player, c *card.Card) error {
	log := logger.WithGameContext(p.GameID(), p.ID())

	// Scan for card-draw, card-peek, card-take, card-buy outputs
	var cardDrawAmount, cardPeekAmount, cardTakeAmount, cardBuyAmount int
	var cardBuyCost int = 3 // Default cost for buying cards in Terraforming Mars

	// Process all behaviors to find card draw/peek effects
	for _, behavior := range c.Behaviors {
		// Only process auto triggers WITHOUT conditions (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == card.ResourceTriggerAuto && behavior.Triggers[0].Condition == nil {
			for _, output := range behavior.Outputs {
				switch output.Type {
				case types.ResourceCardDraw:
					cardDrawAmount += output.Amount
				case types.ResourceCardPeek:
					cardPeekAmount += output.Amount
				case types.ResourceCardTake:
					cardTakeAmount += output.Amount
				case types.ResourceCardBuy:
					cardBuyAmount += output.Amount
				}
			}
		}
	}

	// If no card effects found, return early
	if cardDrawAmount == 0 && cardPeekAmount == 0 && cardTakeAmount == 0 && cardBuyAmount == 0 {
		return nil
	}

	// Determine the scenario and create appropriate PendingCardDrawSelection
	var cardsToShow []string
	var freeTakeCount, maxBuyCount int

	if cardDrawAmount > 0 && cardPeekAmount == 0 && cardTakeAmount == 0 && cardBuyAmount == 0 {
		// Scenario 1: Simple card-draw (e.g., "Draw 2 cards")
		// Draw cards from deck and auto-select all
		drawnCards, err := cp.deckRepo.DrawProjectCards(ctx, cardDrawAmount)
		if err != nil {
			log.Error("Failed to draw cards from deck", zap.Error(err))
			return fmt.Errorf("failed to draw card: %w", err)
		}
		cardsToShow = drawnCards

		// For card-draw, player must take all cards (freeTakeCount = number of cards)
		freeTakeCount = len(drawnCards)
		maxBuyCount = 0

		log.Info("🃏 card.Card draw effect detected",
			zap.String("card_name", c.Name),
			zap.Int("cards_to_draw", len(drawnCards)))

	} else if cardPeekAmount > 0 {
		// Scenario 2/3/4: Peek-based scenarios (card-peek + card-take/card-buy)
		// Draw cards from deck to peek at them (they won't be returned)
		peekedCards, err := cp.deckRepo.DrawProjectCards(ctx, cardPeekAmount)
		if err != nil {
			log.Error("Failed to draw cards from deck for peek", zap.Error(err))
			return fmt.Errorf("failed to peek card: %w", err)
		}
		cardsToShow = peekedCards

		// If card-draw is combined with card-peek, the draw amount becomes mandatory takes
		// card-take adds optional takes on top
		freeTakeCount = cardDrawAmount + cardTakeAmount
		maxBuyCount = cardBuyAmount

		log.Info("🃏 card.Card peek effect detected",
			zap.String("card_name", c.Name),
			zap.Int("cards_to_peek", len(peekedCards)),
			zap.Int("card_draw_amount", cardDrawAmount),
			zap.Int("card_take_amount", cardTakeAmount),
			zap.Int("free_take_count", freeTakeCount),
			zap.Int("max_buy_count", cardBuyAmount))
	} else {
		// Invalid combination (e.g., card-take without card-peek, or card-buy without card-peek)
		log.Warn("⚠️ Invalid card effect combination",
			zap.String("card_name", c.Name),
			zap.Int("card_draw", cardDrawAmount),
			zap.Int("card_peek", cardPeekAmount),
			zap.Int("card_take", cardTakeAmount),
			zap.Int("card_buy", cardBuyAmount))
		return fmt.Errorf("invalid card effect combination: must have either card-draw or card-peek")
	}

	// Create PendingCardDrawSelection
	selection := &player.PendingCardDrawSelection{
		AvailableCards: cardsToShow,
		FreeTakeCount:  freeTakeCount,
		MaxBuyCount:    maxBuyCount,
		CardBuyCost:    cardBuyCost,
		Source:         c.ID,
	}

	// Store in player
	if err := p.SetPendingCardDrawSelection(ctx, selection); err != nil {
		log.Error("Failed to create pending card draw selection", zap.Error(err))
		return fmt.Errorf("failed to create pending card draw selection: %w", err)
	}

	log.Info("✅ Pending card draw selection created",
		zap.String("card_name", c.Name),
		zap.Int("available_cards", len(cardsToShow)),
		zap.Int("free_take_count", freeTakeCount),
		zap.Int("max_buy_count", maxBuyCount),
		zap.Int("card_buy_cost", cardBuyCost))

	return nil
}

// ApplyCardStorageResource handles adding resources to card storage (animals, microbes, floaters, science)
// This is exported so it can be used by actions for both card play and card ability execution
func (cp *CardProcessor) ApplyCardStorageResource(ctx context.Context, p *player.Player, playedCardID string, output card.ResourceCondition, cardStorageTarget *string, log *zap.Logger) error {
	// Get current resource storage
	resourceStorage := p.ResourceStorage()

	// Initialize resource storage map if nil
	if resourceStorage == nil {
		resourceStorage = make(map[string]int)
	}

	// Determine target card based on output.Target
	var targetCardID string
	switch output.Target {
	case card.TargetSelfCard:
		// Target is the card itself (the card being played)
		targetCardID = playedCardID
		log.Debug("💾 Applying card storage to self-card",
			zap.String("target_card_id", targetCardID),
			zap.String("resource_type", string(output.Type)),
			zap.Int("amount", output.Amount))

	case card.TargetAnyCard:
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
		for _, playedCardID := range p.PlayedCards() {
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
	resourceStorage[targetCardID] += output.Amount

	// Persist the updated resource storage
	if err := p.SetResourceStorage(ctx, resourceStorage); err != nil {
		log.Error("Failed to update card resource storage", zap.Error(err))
		return fmt.Errorf("failed to update card resource storage: %w", err)
	}

	log.Info("✅ Card storage resource applied",
		zap.String("target_card_id", targetCardID),
		zap.String("resource_type", string(output.Type)),
		zap.Int("amount", output.Amount),
		zap.Int("new_storage_amount", resourceStorage[targetCardID]))

	return nil
}

// applyGlobalParameterEffects applies global parameter changes (temperature, oxygen, oceans) from a card's behaviors
// choiceIndex is optional and used when the card has choices between different effects
func (cp *CardProcessor) applyGlobalParameterEffects(ctx context.Context, game *Game, p *player.Player, c *card.Card, choiceIndex *int) error {
	log := logger.WithGameContext(p.GameID(), p.ID())

	// Get current game to read current global parameters

	// Track changes for logging
	var temperatureChange, oxygenChange, oceansChange int

	// Process all behaviors to find global parameter effects
	for _, behavior := range c.Behaviors {
		// Only process auto triggers WITHOUT conditions (immediate effects when card is played)
		if len(behavior.Triggers) > 0 && behavior.Triggers[0].Type == card.ResourceTriggerAuto && behavior.Triggers[0].Condition == nil {
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
				case types.ResourceTemperature:
					temperatureChange += output.Amount
				case types.ResourceOxygen:
					oxygenChange += output.Amount
				case types.ResourceOceans:
					oceansChange += output.Amount
				}
			}
		}
	}

	// Apply temperature changes
	if temperatureChange != 0 {
		oldTemperature := game.GlobalParameters.Temperature
		newTemperature := oldTemperature + temperatureChange
		// Clamp to valid range
		if newTemperature > types.MaxTemperature {
			newTemperature = types.MaxTemperature
		}
		if newTemperature < types.MinTemperature {
			newTemperature = types.MinTemperature
		}

		// Calculate actual change after clamping (each step is 2°C)
		actualChange := newTemperature - oldTemperature
		if actualChange == 0 {
			return nil // No actual change after clamping
		}

		game.GlobalParameters.Temperature = newTemperature

		// Increase TR for each step of temperature increase (each step is 2°C)
		stepsChanged := actualChange / 2
		if stepsChanged > 0 {

			newTR := p.TerraformRating() + stepsChanged
			if err := p.SetTerraformRating(ctx, newTR); err != nil {
				log.Error("Failed to update TR after temperature change", zap.Error(err))
				return fmt.Errorf("failed to update terraform rating: %w", err)
			}

			log.Info("🌡️ Temperature changed (TR granted)",
				zap.String("card_name", c.Name),
				zap.Int("temperature_change", actualChange),
				zap.Int("new_temperature", newTemperature),
				zap.Int("tr_change", stepsChanged),
				zap.Int("new_tr", newTR))
		} else {
			log.Info("🌡️ Temperature changed",
				zap.String("card_name", c.Name),
				zap.Int("change", actualChange),
				zap.Int("new_temperature", newTemperature))
		}
	}

	// Apply oxygen changes
	if oxygenChange != 0 {
		oldOxygen := game.GlobalParameters.Oxygen
		newOxygen := oldOxygen + oxygenChange
		// Clamp to valid range
		if newOxygen > types.MaxOxygen {
			newOxygen = types.MaxOxygen
		}
		if newOxygen < types.MinOxygen {
			newOxygen = types.MinOxygen
		}

		// Calculate actual change after clamping (each step is 1%)
		oxygenChange = newOxygen - oldOxygen
		if oxygenChange == 0 {
			return nil // No actual change after clamping
		}

		game.GlobalParameters.Oxygen = newOxygen

		// Increase TR for each step of oxygen increase (each step is 1%)
		if oxygenChange > 0 {

			newTR := p.TerraformRating() + oxygenChange
			if err := p.SetTerraformRating(ctx, newTR); err != nil {
				log.Error("Failed to update TR after oxygen change", zap.Error(err))
				return fmt.Errorf("failed to update terraform rating: %w", err)
			}

			log.Info("💨 Oxygen changed (TR granted)",
				zap.String("card_name", c.Name),
				zap.Int("oxygen_change", oxygenChange),
				zap.Int("new_oxygen", newOxygen),
				zap.Int("tr_change", oxygenChange),
				zap.Int("new_tr", newTR))
		} else {
			log.Info("💨 Oxygen changed",
				zap.String("card_name", c.Name),
				zap.Int("change", oxygenChange),
				zap.Int("new_oxygen", newOxygen))
		}
	}

	// Apply oceans changes
	if oceansChange != 0 {
		oldOceans := game.GlobalParameters.Oceans
		newOceans := oldOceans + oceansChange
		// Clamp to valid range
		if newOceans > types.MaxOceans {
			newOceans = types.MaxOceans
		}
		if newOceans < types.MinOceans {
			newOceans = types.MinOceans
		}

		// Calculate actual change after clamping
		oceansChange = newOceans - oldOceans
		if oceansChange == 0 {
			return nil // No actual change after clamping
		}

		game.GlobalParameters.Oceans = newOceans

		// Increase TR for each ocean tile placed
		if oceansChange > 0 {

			newTR := p.TerraformRating() + oceansChange
			if err := p.SetTerraformRating(ctx, newTR); err != nil {
				log.Error("Failed to update TR after ocean placement", zap.Error(err))
				return fmt.Errorf("failed to update terraform rating: %w", err)
			}

			log.Info("🌊 Oceans changed (TR granted)",
				zap.String("card_name", c.Name),
				zap.Int("oceans_change", oceansChange),
				zap.Int("new_oceans", newOceans),
				zap.Int("tr_change", oceansChange),
				zap.Int("new_tr", newTR))
		} else {
			log.Info("🌊 Oceans changed",
				zap.String("card_name", c.Name),
				zap.Int("change", oceansChange),
				zap.Int("new_oceans", newOceans))
		}
	}

	return nil
}
