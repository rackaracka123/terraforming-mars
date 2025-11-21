package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/card"
	"terraforming-mars-backend/internal/session/deck"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// ExecuteCardActionAction handles the business logic for executing card actions
// Fully migrated to session-based architecture
type ExecuteCardActionAction struct {
	BaseAction
	cardProcessor *card.CardProcessor
	deckRepo      deck.Repository
}

// NewExecuteCardActionAction creates a new execute card action action
func NewExecuteCardActionAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
	cardProcessor *card.CardProcessor,
	deckRepo deck.Repository,
) *ExecuteCardActionAction {
	return &ExecuteCardActionAction{
		BaseAction:    NewBaseAction(gameRepo, playerRepo, sessionMgr),
		cardProcessor: cardProcessor,
		deckRepo:      deckRepo,
	}
}

// Execute performs the execute card action
func (a *ExecuteCardActionAction) Execute(
	ctx context.Context,
	gameID, playerID, cardID string,
	behaviorIndex int,
	choiceIndex *int,
	cardStorageTarget *string,
) error {
	log := a.InitLogger(gameID, playerID)
	log.Debug("🎯 Starting card action play",
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex))

	// 1. Get game and validate current turn
	game, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for card action", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	if game.CurrentTurn == nil {
		log.Error("No current player turn set", zap.String("requesting_player", playerID))
		return fmt.Errorf("no current player turn set, requesting player is %s", playerID)
	}

	if *game.CurrentTurn != playerID {
		log.Error("Not current players turn", zap.String("current_turn", *game.CurrentTurn), zap.String("requesting_player", playerID))
		return fmt.Errorf("not current player's turn: current turn is %s, requesting player is %s", *game.CurrentTurn, playerID)
	}

	// 2. Get the player to validate they exist and check their actions
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for card action play", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if player has available actions
	if player.AvailableActions <= 0 && player.AvailableActions != -1 {
		return fmt.Errorf("no available actions remaining")
	}

	// 3. Find the specific action in the player's action list
	var targetAction *types.PlayerAction
	for i := range player.Actions {
		action := &player.Actions[i]
		if action.CardID == cardID && action.BehaviorIndex == behaviorIndex {
			targetAction = action
			break
		}
	}

	if targetAction == nil {
		return fmt.Errorf("card action not found in player's action list: card %s, behavior %d", cardID, behaviorIndex)
	}

	// 4. Validate that the action hasn't been played this generation (playCount must be 0)
	if targetAction.PlayCount > 0 {
		return fmt.Errorf("action has already been played this generation: current play count %d", targetAction.PlayCount)
	}

	// 5. Validate choice selection for actions with choices
	if len(targetAction.Behavior.Choices) > 0 {
		if choiceIndex == nil {
			return fmt.Errorf("action has choices but no choiceIndex provided")
		}
		if *choiceIndex < 0 || *choiceIndex >= len(targetAction.Behavior.Choices) {
			return fmt.Errorf("invalid choiceIndex %d: must be between 0 and %d", *choiceIndex, len(targetAction.Behavior.Choices)-1)
		}
		log.Debug("🎯 Action has choices, using choiceIndex", zap.Int("choice_index", *choiceIndex))
	}

	log.Debug("🎯 Found target action",
		zap.String("card_name", targetAction.CardName),
		zap.Int("play_count", targetAction.PlayCount))

	// 6. Validate that the player can afford the action inputs (including choice-specific inputs)
	if err := a.validateActionInputs(ctx, gameID, playerID, targetAction, choiceIndex); err != nil {
		return fmt.Errorf("action input validation failed: %w", err)
	}

	// 7. Apply the action inputs (deduct resources, including choice-specific inputs)
	if err := a.applyActionInputs(ctx, gameID, playerID, targetAction, choiceIndex); err != nil {
		return fmt.Errorf("failed to apply action inputs: %w", err)
	}

	// 8. Apply the action outputs (give resources/production/etc., including choice-specific outputs)
	if err := a.applyActionOutputs(ctx, gameID, playerID, targetAction, choiceIndex, cardStorageTarget); err != nil {
		return fmt.Errorf("failed to apply action outputs: %w", err)
	}

	// 9. Increment the play count for this action
	if err := a.incrementActionPlayCount(ctx, gameID, playerID, cardID, behaviorIndex); err != nil {
		return fmt.Errorf("failed to increment action play count: %w", err)
	}

	// 10. Consume one action now that all steps have succeeded
	if player.AvailableActions > 0 {
		newActions := player.AvailableActions - 1
		if err := a.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, newActions); err != nil {
			log.Error("Failed to consume player action", zap.Error(err))
			// Note: Action has already been applied, but we couldn't consume the action
			// This is a critical error but we don't rollback the entire action
			return fmt.Errorf("action applied but failed to consume available action: %w", err)
		}
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", newActions))
	} else {
		log.Debug("✅ Action consumed (unlimited actions)", zap.Int("available_actions", -1))
	}

	// 11. Broadcast game state update
	if err := a.sessionMgr.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after card action play",
			zap.Error(err))
		// Don't fail the action, just log the error
	}

	log.Info("✅ Card action played successfully",
		zap.String("card_id", cardID),
		zap.String("card_name", targetAction.CardName),
		zap.Int("behavior_index", behaviorIndex))
	return nil
}

// validateActionInputs validates that the player has sufficient resources for the action inputs
// choiceIndex is optional and used when the action has choices between different effects
func (a *ExecuteCardActionAction) validateActionInputs(ctx context.Context, gameID, playerID string, action *types.PlayerAction, choiceIndex *int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to check resources
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for input validation: %w", err)
	}

	// Aggregate all inputs: behavior.Inputs + choice[choiceIndex].Inputs
	allInputs := action.Behavior.Inputs

	// If choiceIndex is provided and this action has choices, add choice inputs
	if choiceIndex != nil && len(action.Behavior.Choices) > 0 && *choiceIndex < len(action.Behavior.Choices) {
		selectedChoice := action.Behavior.Choices[*choiceIndex]
		allInputs = append(allInputs, selectedChoice.Inputs...)
		log.Debug("🎯 Validating choice inputs",
			zap.Int("choice_index", *choiceIndex),
			zap.Int("choice_inputs_count", len(selectedChoice.Inputs)))
	}

	// Check each input requirement
	for _, input := range allInputs {
		switch input.Type {
		case types.ResourceCredits:
			if player.Resources.Credits < input.Amount {
				return fmt.Errorf("insufficient credits: need %d, have %d", input.Amount, player.Resources.Credits)
			}
		case types.ResourceSteel:
			if player.Resources.Steel < input.Amount {
				return fmt.Errorf("insufficient steel: need %d, have %d", input.Amount, player.Resources.Steel)
			}
		case types.ResourceTitanium:
			if player.Resources.Titanium < input.Amount {
				return fmt.Errorf("insufficient titanium: need %d, have %d", input.Amount, player.Resources.Titanium)
			}
		case types.ResourcePlants:
			if player.Resources.Plants < input.Amount {
				return fmt.Errorf("insufficient plants: need %d, have %d", input.Amount, player.Resources.Plants)
			}
		case types.ResourceEnergy:
			if player.Resources.Energy < input.Amount {
				return fmt.Errorf("insufficient energy: need %d, have %d", input.Amount, player.Resources.Energy)
			}
		case types.ResourceHeat:
			if player.Resources.Heat < input.Amount {
				return fmt.Errorf("insufficient heat: need %d, have %d", input.Amount, player.Resources.Heat)
			}
		default:
			log.Warn("Unknown input resource type", zap.String("type", string(input.Type)))
			// For unknown types, we'll allow the action to proceed
		}
	}

	log.Debug("✅ Action input validation passed")
	return nil
}

// applyActionInputs applies the action inputs by deducting resources from the player
// choiceIndex is optional and used when the action has choices between different effects
func (a *ExecuteCardActionAction) applyActionInputs(ctx context.Context, gameID, playerID string, action *types.PlayerAction, choiceIndex *int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player resources
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for input application: %w", err)
	}

	// Calculate new resource values after applying inputs
	newResources := player.Resources

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
		case types.ResourceCredits:
			if newResources.Credits < input.Amount {
				return fmt.Errorf("insufficient credits: need %d, have %d", input.Amount, newResources.Credits)
			}
		case types.ResourceSteel:
			if newResources.Steel < input.Amount {
				return fmt.Errorf("insufficient steel: need %d, have %d", input.Amount, newResources.Steel)
			}
		case types.ResourceTitanium:
			if newResources.Titanium < input.Amount {
				return fmt.Errorf("insufficient titanium: need %d, have %d", input.Amount, newResources.Titanium)
			}
		case types.ResourcePlants:
			if newResources.Plants < input.Amount {
				return fmt.Errorf("insufficient plants: need %d, have %d", input.Amount, newResources.Plants)
			}
		case types.ResourceEnergy:
			if newResources.Energy < input.Amount {
				return fmt.Errorf("insufficient energy: need %d, have %d", input.Amount, newResources.Energy)
			}
		case types.ResourceHeat:
			if newResources.Heat < input.Amount {
				return fmt.Errorf("insufficient heat: need %d, have %d", input.Amount, newResources.Heat)
			}

		// Card storage resources (animals, microbes, floaters, science, asteroid)
		case types.ResourceAnimals, types.ResourceMicrobes, types.ResourceFloaters, types.ResourceScience, types.ResourceAsteroid:
			// Validate card storage resource inputs
			if input.Target == types.TargetSelfCard {
				// Initialize resource storage map if nil (for checking)
				if player.ResourceStorage == nil {
					player.ResourceStorage = make(map[string]int)
				}

				currentStorage := player.ResourceStorage[action.CardID]
				if currentStorage < input.Amount {
					return fmt.Errorf("insufficient %s storage on card %s: need %d, have %d",
						input.Type, action.CardID, input.Amount, currentStorage)
				}
			}
		}
	}

	// Track if resource storage was modified
	resourceStorageModified := false

	// APPLICATION PHASE: Apply each input by deducting resources
	for _, input := range allInputs {
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

		// Card storage resources (animals, microbes, floaters, science, asteroid)
		case types.ResourceAnimals, types.ResourceMicrobes, types.ResourceFloaters, types.ResourceScience, types.ResourceAsteroid:
			// Handle card storage resource inputs
			if input.Target == types.TargetSelfCard {
				// Initialize resource storage map if nil
				if player.ResourceStorage == nil {
					player.ResourceStorage = make(map[string]int)
				}

				// Deduct from card storage
				currentStorage := player.ResourceStorage[action.CardID]
				player.ResourceStorage[action.CardID] = currentStorage - input.Amount
				resourceStorageModified = true

				log.Debug("📉 Deducted card storage resource",
					zap.String("card_id", action.CardID),
					zap.String("resource_type", string(input.Type)),
					zap.Int("amount", input.Amount),
					zap.Int("previous_storage", currentStorage),
					zap.Int("new_storage", player.ResourceStorage[action.CardID]))
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
	if err := a.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
		log.Error("Failed to update player resources for action inputs", zap.Error(err))
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	// Update resource storage if modified
	if resourceStorageModified {
		if err := a.playerRepo.UpdateResourceStorage(ctx, gameID, playerID, player.ResourceStorage); err != nil {
			log.Error("Failed to update resource storage for action inputs", zap.Error(err))
			return fmt.Errorf("failed to update resource storage: %w", err)
		}
	}

	log.Debug("✅ Action inputs applied successfully")
	return nil
}

// applyActionOutputs applies the action outputs by giving resources/production/etc. to the player
// choiceIndex is optional and used when the action has choices between different effects
func (a *ExecuteCardActionAction) applyActionOutputs(ctx context.Context, gameID, playerID string, action *types.PlayerAction, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to read current resources and production
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for output application: %w", err)
	}

	// Track what needs to be updated
	var resourcesChanged bool
	var productionChanged bool
	var trChanged bool
	newResources := player.Resources
	newProduction := player.Production

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
			// Ensure production doesn't go below 0
			if newProduction.Credits < 0 {
				newProduction.Credits = 0
			}
			productionChanged = true
		case types.ResourceSteelProduction:
			newProduction.Steel += output.Amount
			if newProduction.Steel < 0 {
				newProduction.Steel = 0
			}
			productionChanged = true
		case types.ResourceTitaniumProduction:
			newProduction.Titanium += output.Amount
			if newProduction.Titanium < 0 {
				newProduction.Titanium = 0
			}
			productionChanged = true
		case types.ResourcePlantsProduction:
			newProduction.Plants += output.Amount
			if newProduction.Plants < 0 {
				newProduction.Plants = 0
			}
			productionChanged = true
		case types.ResourceEnergyProduction:
			newProduction.Energy += output.Amount
			if newProduction.Energy < 0 {
				newProduction.Energy = 0
			}
			productionChanged = true
		case types.ResourceHeatProduction:
			newProduction.Heat += output.Amount
			if newProduction.Heat < 0 {
				newProduction.Heat = 0
			}
			productionChanged = true

		// Terraform rating
		case types.ResourceTR:
			if err := a.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, player.TerraformRating+output.Amount); err != nil {
				log.Error("Failed to update terraform rating", zap.Error(err))
				return fmt.Errorf("failed to update terraform rating: %w", err)
			}
			trChanged = true

		// Card storage resources (animals, microbes, floaters, science, asteroid)
		case types.ResourceAnimals, types.ResourceMicrobes, types.ResourceFloaters, types.ResourceScience, types.ResourceAsteroid:
			// Use the CardProcessor's applyCardStorageResource method
			// For actions, the "played card" is the card that has this action
			if err := a.cardProcessor.ApplyCardStorageResource(ctx, gameID, playerID, action.CardID, output, cardStorageTarget, log); err != nil {
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
		if err := a.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
			log.Error("Failed to update player resources for action outputs", zap.Error(err))
			return fmt.Errorf("failed to update player resources: %w", err)
		}
	}

	// Update production if it changed
	if productionChanged {
		if err := a.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
			log.Error("Failed to update player production for action outputs", zap.Error(err))
			return fmt.Errorf("failed to update player production: %w", err)
		}
	}

	// Process card draw/peek/take/buy effects if any were found
	if cardDrawAmount > 0 || cardPeekAmount > 0 || cardTakeAmount > 0 || cardBuyAmount > 0 {
		if err := a.applyActionCardDrawPeekEffects(ctx, gameID, playerID, action.CardID, cardDrawAmount, cardPeekAmount, cardTakeAmount, cardBuyAmount); err != nil {
			return fmt.Errorf("failed to apply card draw/peek effects: %w", err)
		}
	}

	log.Debug("✅ Action outputs applied successfully",
		zap.Bool("resources_changed", resourcesChanged),
		zap.Bool("production_changed", productionChanged),
		zap.Bool("tr_changed", trChanged))
	return nil
}

// applyActionCardDrawPeekEffects handles card draw/peek/take/buy effects from action outputs
func (a *ExecuteCardActionAction) applyActionCardDrawPeekEffects(ctx context.Context, gameID, playerID, sourceCardID string, cardDrawAmount, cardPeekAmount, cardTakeAmount, cardBuyAmount int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Determine the scenario and create appropriate PendingCardDrawSelection
	var cardsToShow []string
	var freeTakeCount, maxBuyCount int
	var cardBuyCost int = 3 // Default cost for buying cards in Terraforming Mars

	if cardDrawAmount > 0 && cardPeekAmount == 0 && cardTakeAmount == 0 && cardBuyAmount == 0 {
		// Scenario 1: Simple card-draw (e.g., "Draw 2 cards")
		// Draw cards from deck and auto-select all
		drawnCards, err := a.deckRepo.DrawProjectCards(ctx, gameID, cardDrawAmount)
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
		peekedCards, err := a.deckRepo.DrawProjectCards(ctx, gameID, cardPeekAmount)
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
	selection := &types.PendingCardDrawSelection{
		AvailableCards: cardsToShow,
		FreeTakeCount:  freeTakeCount,
		MaxBuyCount:    maxBuyCount,
		CardBuyCost:    cardBuyCost,
		Source:         sourceCardID,
	}

	// Store in player repository
	if err := a.playerRepo.UpdatePendingCardDrawSelection(ctx, gameID, playerID, selection); err != nil {
		log.Error("Failed to create pending card draw selection", zap.Error(err))
		return fmt.Errorf("failed to create pending card draw selection: %w", err)
	}

	log.Info("✅ Pending card draw selection created (from action)",
		zap.String("source_card_id", sourceCardID),
		zap.Int("available_cards", len(cardsToShow)),
		zap.Int("free_take_count", freeTakeCount),
		zap.Int("max_buy_count", maxBuyCount),
		zap.Int("card_buy_cost", cardBuyCost))

	return nil
}

// incrementActionPlayCount increments the play count for a specific action
func (a *ExecuteCardActionAction) incrementActionPlayCount(ctx context.Context, gameID, playerID, cardID string, behaviorIndex int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to read current actions
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for play count update: %w", err)
	}

	// Find and update the specific action
	updatedActions := make([]types.PlayerAction, len(player.Actions))
	copy(updatedActions, player.Actions)

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
	if err := a.playerRepo.UpdatePlayerActions(ctx, gameID, playerID, updatedActions); err != nil {
		log.Error("Failed to update player actions for play count", zap.Error(err))
		return fmt.Errorf("failed to update player actions: %w", err)
	}

	return nil
}
