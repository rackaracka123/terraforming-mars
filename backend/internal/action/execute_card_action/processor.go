package execute_card_action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/game/deck"
	"terraforming-mars-backend/internal/session/game/player/actions"
	"terraforming-mars-backend/internal/session/game/player/selection"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// Processor handles the application logic for card action execution
type Processor struct {
	sessionFactory session.SessionFactory
	cardProcessor  *game.CardProcessor
	deckRepo       deck.Repository
}

// NewProcessor creates a new Processor instance
func NewProcessor(sessionFactory session.SessionFactory, cardProcessor *game.CardProcessor, deckRepo deck.Repository) *Processor {
	return &Processor{
		sessionFactory: sessionFactory,
		cardProcessor:  cardProcessor,
		deckRepo:       deckRepo,
	}
}

// ApplyActionInputs applies the action inputs by deducting resources from the player
// choiceIndex is optional and used when the action has choices between different effects
func (p *Processor) ApplyActionInputs(ctx context.Context, gameID, playerID string, action *actions.PlayerAction, choiceIndex *int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get session and player
	sess := p.sessionFactory.Get(gameID)
	if sess == nil {
		return fmt.Errorf("game session not found: %s", gameID)
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		return fmt.Errorf("player not found: %s", playerID)
	}

	// Get current resources
	currentResources := player.Resources().Get()

	// Calculate new resource values after applying inputs
	newResources := currentResources

	// Aggregate all inputs: behavior.Inputs + choice[choiceIndex].Inputs
	allInputs := action.Behavior.Inputs

	// If choiceIndex is provided and this action has choices, add choice inputs
	if choiceIndex != nil && len(action.Behavior.Choices) > 0 && *choiceIndex < len(action.Behavior.Choices) {
		selectedChoice := action.Behavior.Choices[*choiceIndex]
		allInputs = append(allInputs, selectedChoice.Inputs...)
		log.Debug("🎯 Applying choice inputs",
			zap.Int("choice_index", *choiceIndex),
			zap.Int("choice_inputs_count", len(selectedChoice.Inputs)))
	}

	// VALIDATION PHASE: Check if all inputs can be afforded before making any changes
	for _, input := range allInputs {
		switch input.Type {
		case types.ResourceCredits, types.ResourceSteel, types.ResourceTitanium,
			types.ResourcePlants, types.ResourceEnergy, types.ResourceHeat:
			// Validate standard resources directly
			var currentAmount int
			switch input.Type {
			case types.ResourceCredits:
				currentAmount = newResources.Credits
			case types.ResourceSteel:
				currentAmount = newResources.Steel
			case types.ResourceTitanium:
				currentAmount = newResources.Titanium
			case types.ResourcePlants:
				currentAmount = newResources.Plants
			case types.ResourceEnergy:
				currentAmount = newResources.Energy
			case types.ResourceHeat:
				currentAmount = newResources.Heat
			}
			if currentAmount < input.Amount {
				return fmt.Errorf("insufficient %s: need %d, have %d", input.Type, input.Amount, currentAmount)
			}

		// Card storage resources (animals, microbes, floaters, science, asteroid)
		case types.ResourceAnimals, types.ResourceMicrobes, types.ResourceFloaters, types.ResourceScience, types.ResourceAsteroid:
			// Validate card storage resource inputs
			if input.Target == card.TargetSelfCard {
				resourceStorage := player.Resources().Storage()
				currentStorage := resourceStorage[action.CardID]
				if currentStorage < input.Amount {
					return fmt.Errorf("insufficient %s storage on card %s: need %d, have %d",
						input.Type, action.CardID, input.Amount, currentStorage)
				}
			}
		}
	}

	// APPLICATION PHASE: Apply each input by deducting resources
	for _, input := range allInputs {
		switch input.Type {
		case types.ResourceCredits, types.ResourceSteel, types.ResourceTitanium,
			types.ResourcePlants, types.ResourceEnergy, types.ResourceHeat:
			// Apply resource change (negative amount)
			switch input.Type {
			case types.ResourceCredits:
				newResources.Credits -= input.Amount
			case types.ResourceSteel:
				newResources.Steel -= input.Amount
			case types.ResourceTitanium:
				newResources.Titanium -= input.Amount
			case types.ResourcePlants:
				newResources.Plants -= input.Amount
			case types.ResourceEnergy:
				newResources.Energy -= input.Amount
			case types.ResourceHeat:
				newResources.Heat -= input.Amount
			}

		// Card storage resources (animals, microbes, floaters, science, asteroid)
		case types.ResourceAnimals, types.ResourceMicrobes, types.ResourceFloaters, types.ResourceScience, types.ResourceAsteroid:
			// Handle card storage resource inputs
			if input.Target == card.TargetSelfCard {
				resourceStorage := player.Resources().Storage()
				// Deduct from card storage
				currentStorage := resourceStorage[action.CardID]
				newStorage := make(map[string]int)
				for k, v := range resourceStorage {
					newStorage[k] = v
				}
				newStorage[action.CardID] = currentStorage - input.Amount

				player.Resources().SetStorage(newStorage)

				log.Debug("📉 Deducted card storage resource",
					zap.String("card_id", action.CardID),
					zap.String("resource_type", string(input.Type)),
					zap.Int("amount", input.Amount),
					zap.Int("previous_storage", currentStorage),
					zap.Int("new_storage", newStorage[action.CardID]))
			} else {
				log.Warn("Card storage input with non-self-card target not supported",
					zap.String("type", string(input.Type)),
					zap.String("target", string(input.Target)))
			}

		default:
			log.Warn("Unknown input resource type during application", zap.String("type", string(input.Type)))
		}

		log.Debug("💰 Applied input",
			zap.String("resource_type", string(input.Type)),
			zap.Int("amount", input.Amount))
	}

	// Update player resources
	player.Resources().Set(newResources)

	log.Debug("✅ Action inputs applied successfully")
	return nil
}

// ApplyActionOutputs applies the action outputs by giving resources/production/etc. to the player
// choiceIndex is optional and used when the action has choices between different effects
func (p *Processor) ApplyActionOutputs(ctx context.Context, gameID, playerID string, action *actions.PlayerAction, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get session and player
	sess := p.sessionFactory.Get(gameID)
	if sess == nil {
		return fmt.Errorf("game session not found: %s", gameID)
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		return fmt.Errorf("player not found: %s", playerID)
	}

	// Get current resources and production
	currentResources := player.Resources().Get()
	currentProduction := player.Resources().Production()
	currentTR := player.Resources().TerraformRating()

	// Track what needs to be updated
	var resourcesChanged bool
	var productionChanged bool
	var trChanged bool
	newResources := currentResources
	newProduction := currentProduction

	// Track card draw/peek effects
	var cardDrawAmount, cardPeekAmount, cardTakeAmount, cardBuyAmount int

	// Aggregate all outputs: behavior.Outputs + choice[choiceIndex].Outputs
	allOutputs := action.Behavior.Outputs

	// If choiceIndex is provided and this action has choices, add choice outputs
	if choiceIndex != nil && len(action.Behavior.Choices) > 0 && *choiceIndex < len(action.Behavior.Choices) {
		selectedChoice := action.Behavior.Choices[*choiceIndex]
		allOutputs = append(allOutputs, selectedChoice.Outputs...)
		log.Debug("🎯 Applying choice outputs",
			zap.Int("choice_index", *choiceIndex),
			zap.Int("choice_outputs_count", len(selectedChoice.Outputs)))
	}

	// Apply each output
	for _, output := range allOutputs {
		switch output.Type {
		// Immediate resource gains
		case types.ResourceCredits:
			newResources.Credits += output.Amount
			resourcesChanged = true
		case types.ResourceSteel:
			newResources.Steel += output.Amount
			resourcesChanged = true
		case types.ResourceTitanium:
			newResources.Titanium += output.Amount
			resourcesChanged = true
		case types.ResourcePlants:
			newResources.Plants += output.Amount
			resourcesChanged = true
		case types.ResourceEnergy:
			newResources.Energy += output.Amount
			resourcesChanged = true
		case types.ResourceHeat:
			newResources.Heat += output.Amount
			resourcesChanged = true

		// Production increases
		case types.ResourceCreditsProduction:
			newProduction.Credits += output.Amount
			productionChanged = true
		case types.ResourceSteelProduction:
			newProduction.Steel += output.Amount
			productionChanged = true
		case types.ResourceTitaniumProduction:
			newProduction.Titanium += output.Amount
			productionChanged = true
		case types.ResourcePlantsProduction:
			newProduction.Plants += output.Amount
			productionChanged = true
		case types.ResourceEnergyProduction:
			newProduction.Energy += output.Amount
			productionChanged = true
		case types.ResourceHeatProduction:
			newProduction.Heat += output.Amount
			productionChanged = true

		// Terraform rating
		case types.ResourceTR:
			newTR := currentTR + output.Amount
			player.Resources().SetTerraformRating(newTR)
			trChanged = true

		// Card storage resources (animals, microbes, floaters, science, asteroid)
		case types.ResourceAnimals, types.ResourceMicrobes, types.ResourceFloaters, types.ResourceScience, types.ResourceAsteroid:
			// Use the CardProcessor's applyCardStorageResource method
			// For actions, the "played card" is the card that has this action
			if err := p.cardProcessor.ApplyCardStorageResource(ctx, player, action.CardID, output, cardStorageTarget, log); err != nil {
				return fmt.Errorf("failed to apply card storage resource for action: %w", err)
			}

		// Card draw/peek/take/buy effects
		case types.ResourceCardDraw:
			cardDrawAmount += output.Amount
		case types.ResourceCardPeek:
			cardPeekAmount += output.Amount
		case types.ResourceCardTake:
			cardTakeAmount += output.Amount
		case types.ResourceCardBuy:
			cardBuyAmount += output.Amount

		default:
			log.Warn("Unknown output resource type", zap.String("type", string(output.Type)))
		}

		log.Debug("📈 Applied output",
			zap.String("resource_type", string(output.Type)),
			zap.Int("amount", output.Amount))
	}

	// Update resources if they changed
	if resourcesChanged {
		player.Resources().Set(newResources)
	}

	// Update production if it changed
	if productionChanged {
		player.Resources().SetProduction(newProduction)
	}

	// Process card draw/peek/take/buy effects if any were found
	if cardDrawAmount > 0 || cardPeekAmount > 0 || cardTakeAmount > 0 || cardBuyAmount > 0 {
		if err := p.ApplyCardDrawPeekEffects(ctx, gameID, playerID, action.CardID, cardDrawAmount, cardPeekAmount, cardTakeAmount, cardBuyAmount); err != nil {
			return fmt.Errorf("failed to apply card draw/peek effects: %w", err)
		}
	}

	log.Debug("✅ Action outputs applied successfully",
		zap.Bool("resources_changed", resourcesChanged),
		zap.Bool("production_changed", productionChanged),
		zap.Bool("tr_changed", trChanged))
	return nil
}

// ApplyCardDrawPeekEffects handles card draw/peek/take/buy effects from action outputs
func (p *Processor) ApplyCardDrawPeekEffects(ctx context.Context, gameID, playerID, sourceCardID string, cardDrawAmount, cardPeekAmount, cardTakeAmount, cardBuyAmount int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get session and player
	sess := p.sessionFactory.Get(gameID)
	if sess == nil {
		return fmt.Errorf("game session not found: %s", gameID)
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		return fmt.Errorf("player not found: %s", playerID)
	}

	// Determine the scenario and create appropriate PendingCardDrawSelection
	var cardsToShow []string
	var freeTakeCount, maxBuyCount int
	var cardBuyCost int = 3 // Default cost for buying cards in Terraforming Mars

	if cardDrawAmount > 0 && cardPeekAmount == 0 && cardTakeAmount == 0 && cardBuyAmount == 0 {
		// Scenario 1: Simple card-draw (e.g., "Draw 2 cards")
		// Draw cards from deck and auto-select all
		drawnCards, err := p.deckRepo.DrawProjectCards(ctx, cardDrawAmount)
		if err != nil {
			log.Error("Failed to draw cards from deck", zap.Error(err))
			return fmt.Errorf("failed to draw card: %w", err)
		}
		cardsToShow = drawnCards

		// For card-draw, player must take all cards (freeTakeCount = number of cards)
		freeTakeCount = len(drawnCards)
		maxBuyCount = 0

		log.Info("🃏 Card draw effect detected (from action)",
			zap.String("source_card_id", sourceCardID),
			zap.Int("cards_to_draw", len(drawnCards)))

	} else if cardPeekAmount > 0 {
		// Scenario 2/3/4: Peek-based scenarios (card-peek + card-take/card-buy)
		// Draw cards from deck to peek at them (they won't be returned)
		peekedCards, err := p.deckRepo.DrawProjectCards(ctx, cardPeekAmount)
		if err != nil {
			log.Error("Failed to draw cards from deck for peek", zap.Error(err))
			return fmt.Errorf("failed to peek card: %w", err)
		}
		cardsToShow = peekedCards

		// If card-draw is combined with card-peek, the draw amount becomes mandatory takes
		// card-take adds optional takes on top
		freeTakeCount = cardDrawAmount + cardTakeAmount
		maxBuyCount = cardBuyAmount

		log.Info("🃏 Card peek effect detected (from action)",
			zap.String("source_card_id", sourceCardID),
			zap.Int("cards_to_peek", len(peekedCards)),
			zap.Int("card_draw_amount", cardDrawAmount),
			zap.Int("card_take_amount", cardTakeAmount),
			zap.Int("free_take_count", freeTakeCount),
			zap.Int("max_buy_count", cardBuyAmount))
	} else {
		// Invalid combination (e.g., card-take without card-peek, or card-buy without card-peek)
		log.Warn("⚠️ Invalid card effect combination (from action)",
			zap.String("source_card_id", sourceCardID),
			zap.Int("card_draw", cardDrawAmount),
			zap.Int("card_peek", cardPeekAmount),
			zap.Int("card_take", cardTakeAmount),
			zap.Int("card_buy", cardBuyAmount))
		return fmt.Errorf("invalid card effect combination: must have either card-draw or card-peek")
	}

	// Create PendingCardDrawSelection
	selection := &selection.PendingCardDrawSelection{
		AvailableCards: cardsToShow,
		FreeTakeCount:  freeTakeCount,
		MaxBuyCount:    maxBuyCount,
		CardBuyCost:    cardBuyCost,
		Source:         sourceCardID,
	}

	// Store on Player (card selection phase state on Player)
	player.Selection().SetPendingCardDrawSelection(selection)

	log.Info("✅ Pending card draw selection created (from action)",
		zap.String("source_card_id", sourceCardID),
		zap.Int("available_cards", len(cardsToShow)),
		zap.Int("free_take_count", freeTakeCount),
		zap.Int("max_buy_count", maxBuyCount),
		zap.Int("card_buy_cost", cardBuyCost))

	return nil
}

// IncrementActionPlayCount increments the play count for a specific action
func (p *Processor) IncrementActionPlayCount(ctx context.Context, gameID, playerID, cardID string, behaviorIndex int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get session and player
	sess := p.sessionFactory.Get(gameID)
	if sess == nil {
		return fmt.Errorf("game session not found: %s", gameID)
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		return fmt.Errorf("player not found: %s", playerID)
	}

	// Get current actions and update the specific action
	currentActions := player.Actions().List()
	updatedActions := make([]actions.PlayerAction, len(currentActions))
	copy(updatedActions, currentActions)

	for i := range updatedActions {
		if updatedActions[i].CardID == cardID && updatedActions[i].BehaviorIndex == behaviorIndex {
			updatedActions[i].PlayCount++
			log.Debug("🎯 Incremented play count",
				zap.String("card_id", cardID),
				zap.Int("behavior_index", behaviorIndex),
				zap.Int("new_play_count", updatedActions[i].PlayCount))
			break
		}
	}

	// Update player actions
	player.Actions().SetActions(updatedActions)

	return nil
}
