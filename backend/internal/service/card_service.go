package service

import (
	"context"
	"fmt"
	"slices"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/card"
	"terraforming-mars-backend/internal/session/deck"
	sessionGame "terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/tile"
)

// CardService handles card-related operations
type CardService interface {
	// Player actions for card selection and play (includes corporation selection)
	OnSelectStartingCards(ctx context.Context, gameID, playerID string, cardIDs []string, corporationID string) error

	// Player action for production card selection
	OnSelectProductionCards(ctx context.Context, gameID, playerID string, cardIDs []string) error

	// Player action for confirming card draw/peek selection
	OnConfirmCardDraw(ctx context.Context, gameID, playerID string, cardsToTake []string, cardsToBuy []string) error

	// Get starting cards for selection
	GetStartingCards(ctx context.Context) ([]model.Card, error)

	// Get card by ID
	GetCardByID(ctx context.Context, cardID string) (*model.Card, error)

	// Player actions for playing cards
	OnPlayCard(ctx context.Context, gameID, playerID, cardID string, payment *model.CardPayment, choiceIndex *int, cardStorageTarget *string) error

	// Play a card action from player's action list
	OnPlayCardAction(ctx context.Context, gameID, playerID, cardID string, behaviorIndex int, choiceIndex *int, cardStorageTarget *string) error

	// List cards with pagination
	ListCardsPaginated(ctx context.Context, offset, limit int) ([]model.Card, int, error)

	// Get all corporations
	GetCorporations(ctx context.Context) ([]model.Card, error)
}

// CardServiceImpl implements CardService interface using specialized card managers
type CardServiceImpl struct {
	// Core repositories (session-based)
	gameRepo       sessionGame.Repository
	playerRepo     player.Repository
	cardRepo       card.Repository
	deckRepo       deck.Repository
	sessionManager session.SessionManager

	// Specialized managers from session-scoped card package
	requirementsValidator *card.RequirementsValidator
	effectProcessor       *card.CardProcessor
	cardManager           card.CardManager
	forcedActionManager   card.ForcedActionManager

	// NEW session-based tile processor
	tileProcessor *tile.Processor
}

// NewCardService creates a new CardService instance with session-based repositories
func NewCardService(gameRepo sessionGame.Repository, playerRepo player.Repository, cardRepo card.Repository, deckRepo deck.Repository, sessionManager session.SessionManager, tileProcessor *tile.Processor, effectSubscriber card.CardEffectSubscriber, forcedActionManager card.ForcedActionManager) CardService {
	return &CardServiceImpl{
		gameRepo:              gameRepo,
		playerRepo:            playerRepo,
		cardRepo:              cardRepo,
		deckRepo:              deckRepo,
		sessionManager:        sessionManager,
		requirementsValidator: card.NewRequirementsValidator(cardRepo),
		effectProcessor:       card.NewCardProcessor(gameRepo, playerRepo, deckRepo),
		cardManager:           card.NewCardManager(gameRepo, playerRepo, cardRepo, deckRepo, effectSubscriber),
		tileProcessor:         tileProcessor,
		forcedActionManager:   forcedActionManager,
	}
}

// Delegation methods - all operations are handled by the specialized cards service

func (s *CardServiceImpl) OnSelectStartingCards(ctx context.Context, gameID, playerID string, cardIDs []string, corporationID string) error {
	log := logger.WithGameContext(gameID, playerID)

	// DISABLED during migration - selectionManager not available
	return fmt.Errorf("OnSelectStartingCards not implemented during migration")

	// err := s.selectionManager.SelectStartingCards(ctx, gameID, playerID, cardIDs, corporationID)
	// if err != nil {
	// 	return err
	// }

	log.Debug("🃏 Player completed starting card selection", zap.Strings("card_ids", cardIDs))

	// Check if all players have completed their starting card selection
	if s.isAllPlayersCardSelectionComplete(ctx, gameID) {
		log.Info("✅ All players completed starting card selection, advancing to action phase")

		// Get current game state to validate phase transition
		g, err := s.gameRepo.GetByID(ctx, gameID)
		if err != nil {
			log.Error("Failed to get game for phase advancement", zap.Error(err))
			return fmt.Errorf("failed to get game: %w", err)
		}

		// Validate current phase before transition - use NEW game types
		if g.CurrentPhase != sessionGame.GamePhaseStartingCardSelection {
			log.Warn("Game is not in starting card selection phase, skipping phase transition",
				zap.String("current_phase", string(g.CurrentPhase)))
		} else if g.Status != sessionGame.GameStatusActive {
			log.Warn("Game is not active, skipping phase transition",
				zap.String("current_status", string(g.Status)))
		} else {
			// Advance to action phase - use NEW game types
			if err := s.gameRepo.UpdatePhase(ctx, gameID, sessionGame.GamePhaseAction); err != nil {
				log.Error("Failed to update game phase", zap.Error(err))
				return fmt.Errorf("failed to update game phase: %w", err)
			}

			// TODO: Clear temporary card selection data (disabled during migration)
			// s.selectionManager.ClearGameSelectionData(gameID)

			log.Info("🎯 Game phase advanced successfully",
				zap.String("previous_phase", string(model.GamePhaseStartingCardSelection)),
				zap.String("new_phase", string(model.GamePhaseAction)))

			// Note: Forced actions are now triggered via event system (GamePhaseChangedEvent)
		}
	}

	// Broadcast updated game state to all players after successful card selection (and potential phase change)
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		logger.Get().Error("Failed to broadcast game state after starting card selection",
			zap.Error(err),
			zap.String("game_id", gameID),
			zap.String("player_id", playerID))
		// Don't fail the card selection operation, just log the error
	}

	return nil
}

func (s *CardServiceImpl) OnSelectProductionCards(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	// TODO: Implement during migration
	return fmt.Errorf("OnSelectProductionCards not implemented during migration")
}

// OnConfirmCardDraw handles player confirmation of card draw/peek/take/buy selection
func (s *CardServiceImpl) OnConfirmCardDraw(ctx context.Context, gameID, playerID string, cardsToTake []string, cardsToBuy []string) error {
	// TODO: Implement during migration - PendingCardDrawSelection not yet in NEW player type
	return fmt.Errorf("OnConfirmCardDraw not implemented during migration")

	/* DISABLED during migration - PendingCardDrawSelection not in NEW player type
	log := logger.WithGameContext(gameID, playerID)

	// Get player's pending card draw selection
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	if player.PendingCardDrawSelection == nil {
		return fmt.Errorf("no pending card draw selection found")
	}

	selection := player.PendingCardDrawSelection

	// Validate total cards selected
	totalSelected := len(cardsToTake) + len(cardsToBuy)
	maxAllowed := selection.FreeTakeCount + selection.MaxBuyCount

	if totalSelected > maxAllowed {
		return fmt.Errorf("too many cards selected: selected %d, max allowed %d", totalSelected, maxAllowed)
	}

	// Validate free take count
	if len(cardsToTake) > selection.FreeTakeCount {
		return fmt.Errorf("too many free cards selected: selected %d, max %d", len(cardsToTake), selection.FreeTakeCount)
	}

	// For pure card-draw scenarios (all cards must be taken, no choice), require player to take all
	// This only applies when MaxBuyCount = 0 AND FreeTakeCount = total available cards
	isPureCardDraw := selection.MaxBuyCount == 0 && selection.FreeTakeCount == len(selection.AvailableCards)
	if isPureCardDraw && len(cardsToTake) != selection.FreeTakeCount {
		return fmt.Errorf("must take all %d cards for card-draw effect", selection.FreeTakeCount)
	}

	// Validate buy count
	if len(cardsToBuy) > selection.MaxBuyCount {
		return fmt.Errorf("too many cards to buy: selected %d, max %d", len(cardsToBuy), selection.MaxBuyCount)
	}

	// Validate all selected cards are in available cards
	allSelectedCards := append(cardsToTake, cardsToBuy...)
	for _, cardID := range allSelectedCards {
		if !slices.Contains(selection.AvailableCards, cardID) {
			return fmt.Errorf("card %s not in available cards", cardID)
		}
	}

	// Calculate total cost for bought cards
	totalCost := len(cardsToBuy) * selection.CardBuyCost

	// Validate player can afford bought cards
	if totalCost > 0 {
		if player.Resources.Credits < totalCost {
			return fmt.Errorf("insufficient credits to buy cards: need %d, have %d", totalCost, player.Resources.Credits)
		}

		// Deduct credits for bought cards
		newResources := player.Resources
		newResources.Credits -= totalCost
		if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
			return fmt.Errorf("failed to deduct credits for bought cards: %w", err)
		}

		log.Debug("💰 Paid for bought cards",
			zap.Int("cards_bought", len(cardsToBuy)),
			zap.Int("cost", totalCost),
			zap.Int("remaining_credits", newResources.Credits))
	}

	// Add all selected cards to player's hand
	for _, cardID := range allSelectedCards {
		if err := s.playerRepo.AddCard(ctx, gameID, playerID, cardID); err != nil {
			return fmt.Errorf("failed to add card %s to hand: %w", cardID, err)
		}
	}

	log.Debug("🃏 Added selected cards to hand",
		zap.Int("cards_taken", len(cardsToTake)),
		zap.Int("cards_bought", len(cardsToBuy)),
		zap.Int("total_cards", len(allSelectedCards)))

	// Discard unselected cards (they were already popped from deck, so we just don't add them to hand)
	unselectedCards := []string{}
	for _, cardID := range selection.AvailableCards {
		if !slices.Contains(allSelectedCards, cardID) {
			unselectedCards = append(unselectedCards, cardID)
		}
	}

	if len(unselectedCards) > 0 {
		log.Debug("🗑️ Discarded unselected cards",
			zap.Int("count", len(unselectedCards)),
			zap.Strings("card_ids", unselectedCards))
	}

	// Check if this card draw was triggered by a forced action
	isForcedAction := false
	if player.ForcedFirstAction != nil && player.ForcedFirstAction.CorporationID == selection.Source {
		isForcedAction = true
	}

	// Clear pending card draw selection
	if err := s.playerRepo.ClearPendingCardDrawSelection(ctx, gameID, playerID); err != nil {
		return fmt.Errorf("failed to clear pending card draw selection: %w", err)
	}

	// If this was a forced action, mark it as complete
	if isForcedAction {
		if err := s.forcedActionManager.MarkComplete(ctx, gameID, playerID); err != nil {
			log.Error("Failed to mark forced action complete", zap.Error(err))
			// Don't fail the operation, just log the error
		} else {
			log.Info("🎯 Forced action marked as complete", zap.String("corporation_id", selection.Source))
		}
	}

	log.Info("✅ Card draw confirmation completed",
		zap.String("source", selection.Source),
		zap.Int("cards_taken", len(cardsToTake)),
		zap.Int("cards_bought", len(cardsToBuy)),
		zap.Int("total_cost", totalCost))

	// Broadcast game state update
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after card draw confirmation", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	return nil
	*/
}

// validateStartingCardSelection validates a player's starting card selection (internal use only)
// DISABLED during migration - selectionManager not available
/* func (s *CardServiceImpl) validateStartingCardSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	return s.selectionManager.ValidateStartingCardSelection(ctx, gameID, playerID, cardIDs)
} */

// isAllPlayersCardSelectionComplete checks if all players in the game have completed card selection (internal use only)
// DISABLED during migration - selectionManager not available
func (s *CardServiceImpl) isAllPlayersCardSelectionComplete(ctx context.Context, gameID string) bool {
	return false // TODO: Implement during migration
	// return s.selectionManager.IsAllPlayersCardSelectionComplete(ctx, gameID)
}

// DISABLED during migration - selectionManager not available
/* func (s *CardServiceImpl) ClearGameSelectionData(gameID string) {
	s.selectionManager.ClearGameSelectionData(gameID)
} */

func (s *CardServiceImpl) GetStartingCards(ctx context.Context) ([]model.Card, error) {
	// Get cards from NEW card repository
	newCards, err := s.cardRepo.GetStartingCardPool(ctx)
	if err != nil {
		return nil, err
	}

	// Convert NEW card types to OLD model types
	modelCards := make([]model.Card, len(newCards))
	for i, c := range newCards {
		modelCards[i] = cardToModel(c)
	}

	return modelCards, nil
}

func (s *CardServiceImpl) GetCardByID(ctx context.Context, cardID string) (*model.Card, error) {
	// Get card from NEW card repository
	newCard, err := s.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	// Convert NEW card type to OLD model type
	modelCard := cardToModel(*newCard)
	return &modelCard, nil
}

func (s *CardServiceImpl) OnPlayCard(ctx context.Context, gameID, playerID, cardID string, payment *model.CardPayment, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🎯 Playing card using simplified interface", zap.String("card_id", cardID))

	// Validate payment is provided
	if payment == nil {
		return fmt.Errorf("payment is required")
	}

	// STEP 1: Service-level validations (turn, actions, ownership)
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	if game.CurrentTurn == nil {
		return fmt.Errorf("no current player turn set")
	}

	if *game.CurrentTurn != playerID {
		return fmt.Errorf("not current player's turn: current turn is %s", *game.CurrentTurn)
	}

	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	log.Debug("🔍 Validating card in hand",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("requested_card", cardID),
		zap.Strings("player_cards", player.Cards),
		zap.Int("card_count", len(player.Cards)),
		zap.Bool("has_card", slices.Contains(player.Cards, cardID)))

	// TODO: Re-enable action count validation during migration
	// -1 Available actions means we have infinite (solo game)
	// if player.AvailableActions <= 0 && player.AvailableActions != -1 {
	// 	return fmt.Errorf("no actions available: player has %d actions", player.AvailableActions)
	// }

	if !slices.Contains(player.Cards, cardID) {
		log.Warn("❌ Card NOT found in player's hand",
			zap.String("game_id", gameID),
			zap.String("player_id", playerID),
			zap.String("requested_card", cardID),
			zap.Strings("player_cards", player.Cards))
		return fmt.Errorf("player does not have card %s", cardID)
	}

	// STEP 1.5: Validate choice selection for cards with AUTO-triggered choices
	// Manual-triggered behaviors (actions) will have their choices resolved when the action is played
	card, err := s.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("failed to get card: %w", err)
	}

	// Check if any AUTO-triggered behavior has choices
	hasAutoChoices := false
	for _, behavior := range card.Behaviors {
		// Only check behaviors with auto triggers
		hasAutoTrigger := false
		for _, trigger := range behavior.Triggers {
			if trigger.Type == model.ResourceTriggerAuto {
				hasAutoTrigger = true
				break
			}
		}

		// If this is an auto-triggered behavior with choices, validate choiceIndex
		if hasAutoTrigger && len(behavior.Choices) > 0 {
			hasAutoChoices = true
			// Validate that choiceIndex is provided and within valid range
			if choiceIndex == nil {
				return fmt.Errorf("card has auto-triggered choices but no choiceIndex provided")
			}
			if *choiceIndex < 0 || *choiceIndex >= len(behavior.Choices) {
				return fmt.Errorf("invalid choiceIndex %d: must be between 0 and %d", *choiceIndex, len(behavior.Choices)-1)
			}
			break
		}
	}

	if hasAutoChoices {
		log.Debug("🎯 Card has auto-triggered choices, using choiceIndex", zap.Int("choice_index", *choiceIndex))
	}

	// STEP 2: Use CardManager for card-specific validation (including payment and choice-based costs)
	if err := s.cardManager.CanPlay(ctx, gameID, playerID, cardID, payment, choiceIndex, cardStorageTarget); err != nil {
		return fmt.Errorf("card cannot be played: %w", err)
	}

	// STEP 3: Use CardManager to play the card with payment, choice index, and card storage target
	if err := s.cardManager.PlayCard(ctx, gameID, playerID, cardID, payment, choiceIndex, cardStorageTarget); err != nil {
		return fmt.Errorf("failed to play card: %w", err)
	}

	// STEP 4: Tile queue processing (now automatic via TileQueueCreatedEvent)
	// No manual call needed - TileProcessor subscribes to events and processes automatically

	// TODO: Re-enable action consumption during migration
	// STEP 5: Service-level post-play actions (consume action, broadcast)
	// if player.AvailableActions != -1 {
	// 	newActions := player.AvailableActions - 1
	// 	if err := s.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, newActions); err != nil {
	// 		return fmt.Errorf("card played but failed to consume action: %w", err)
	// 	}
	// 	log.Debug("🎯 Action consumed", zap.Int("remaining_actions", newActions))
	// }

	// STEP 5: Broadcast game state update
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after card play", zap.Error(err))
		// Don't fail the card play operation, just log the error
	}

	log.Info("✅ Card played successfully", zap.String("card_id", cardID))
	return nil
}

func (s *CardServiceImpl) ListCardsPaginated(ctx context.Context, offset, limit int) ([]model.Card, int, error) {
	// Get all cards from NEW repository
	newCards, err := s.cardRepo.GetAllCards(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get cards: %w", err)
	}

	totalCount := len(newCards)

	// Apply pagination
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 50 // Default limit
	}

	start := offset
	end := offset + limit

	if start >= totalCount {
		return []model.Card{}, totalCount, nil
	}

	if end > totalCount {
		end = totalCount
	}

	paginatedNewCards := newCards[start:end]

	// Convert NEW card types to OLD model types
	paginatedCards := make([]model.Card, len(paginatedNewCards))
	for i, c := range paginatedNewCards {
		paginatedCards[i] = cardToModel(c)
	}

	return paginatedCards, totalCount, nil
}

func (s *CardServiceImpl) GetCorporations(ctx context.Context) ([]model.Card, error) {
	// Get corporations from NEW card repository
	newCorporations, err := s.cardRepo.GetCorporations(ctx)
	if err != nil {
		return nil, err
	}

	// Convert NEW card types to OLD model types
	modelCorporations := make([]model.Card, len(newCorporations))
	for i, c := range newCorporations {
		modelCorporations[i] = cardToModel(c)
	}

	return modelCorporations, nil
}

// OnPlayCardAction plays a card action from the player's action list
func (s *CardServiceImpl) OnPlayCardAction(ctx context.Context, gameID, playerID, cardID string, behaviorIndex int, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🎯 Starting card action play",
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex))

	game, err := s.gameRepo.GetByID(ctx, gameID)

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

	// Get the player to validate they exist and check their actions
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for card action play", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// TODO: Re-enable action count validation during migration
	// Check if player has available actions
	// if player.AvailableActions <= 0 && player.AvailableActions != -1 {
	// 	return fmt.Errorf("no available actions remaining")
	// }

	// Find the specific action in the player's action list
	var targetAction *model.PlayerAction
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

	// Validate that the action hasn't been played this generation (playCount must be 0)
	if targetAction.PlayCount > 0 {
		return fmt.Errorf("action has already been played this generation: current play count %d", targetAction.PlayCount)
	}

	// Validate choice selection for actions with choices
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

	// Validate that the player can afford the action inputs (including choice-specific inputs)
	if err := s.validateActionInputs(ctx, gameID, playerID, targetAction, choiceIndex); err != nil {
		return fmt.Errorf("action input validation failed: %w", err)
	}

	// Apply the action inputs (deduct resources, including choice-specific inputs)
	if err := s.applyActionInputs(ctx, gameID, playerID, targetAction, choiceIndex); err != nil {
		return fmt.Errorf("failed to apply action inputs: %w", err)
	}

	// Apply the action outputs (give resources/production/etc., including choice-specific outputs)
	if err := s.applyActionOutputs(ctx, gameID, playerID, targetAction, choiceIndex, cardStorageTarget); err != nil {
		return fmt.Errorf("failed to apply action outputs: %w", err)
	}

	// Increment the play count for this action
	if err := s.incrementActionPlayCount(ctx, gameID, playerID, cardID, behaviorIndex); err != nil {
		return fmt.Errorf("failed to increment action play count: %w", err)
	}

	// Consume one action now that all steps have succeeded
	if player.AvailableActions > 0 {
		newActions := player.AvailableActions - 1
		if err := s.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, newActions); err != nil {
			log.Error("Failed to consume player action", zap.Error(err))
			// Note: Action has already been applied, but we couldn't consume the action
			// This is a critical error but we don't rollback the entire action
			return fmt.Errorf("action applied but failed to consume available action: %w", err)
		}
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", newActions))
	} else {
		log.Debug("✅ Action consumed (unlimited actions)", zap.Int("available_actions", -1))
	}

	// Broadcast game state update
	if err := s.sessionManager.Broadcast(gameID); err != nil {
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
func (s *CardServiceImpl) validateActionInputs(ctx context.Context, gameID, playerID string, action *model.PlayerAction, choiceIndex *int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to check resources
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
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
		case model.ResourceCredits:
			if player.Resources.Credits < input.Amount {
				return fmt.Errorf("insufficient credits: need %d, have %d", input.Amount, player.Resources.Credits)
			}
		case model.ResourceSteel:
			if player.Resources.Steel < input.Amount {
				return fmt.Errorf("insufficient steel: need %d, have %d", input.Amount, player.Resources.Steel)
			}
		case model.ResourceTitanium:
			if player.Resources.Titanium < input.Amount {
				return fmt.Errorf("insufficient titanium: need %d, have %d", input.Amount, player.Resources.Titanium)
			}
		case model.ResourcePlants:
			if player.Resources.Plants < input.Amount {
				return fmt.Errorf("insufficient plants: need %d, have %d", input.Amount, player.Resources.Plants)
			}
		case model.ResourceEnergy:
			if player.Resources.Energy < input.Amount {
				return fmt.Errorf("insufficient energy: need %d, have %d", input.Amount, player.Resources.Energy)
			}
		case model.ResourceHeat:
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
func (s *CardServiceImpl) applyActionInputs(ctx context.Context, gameID, playerID string, action *model.PlayerAction, choiceIndex *int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player resources
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
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
		case model.ResourceCredits:
			if newResources.Credits < input.Amount {
				return fmt.Errorf("insufficient credits: need %d, have %d", input.Amount, newResources.Credits)
			}
		case model.ResourceSteel:
			if newResources.Steel < input.Amount {
				return fmt.Errorf("insufficient steel: need %d, have %d", input.Amount, newResources.Steel)
			}
		case model.ResourceTitanium:
			if newResources.Titanium < input.Amount {
				return fmt.Errorf("insufficient titanium: need %d, have %d", input.Amount, newResources.Titanium)
			}
		case model.ResourcePlants:
			if newResources.Plants < input.Amount {
				return fmt.Errorf("insufficient plants: need %d, have %d", input.Amount, newResources.Plants)
			}
		case model.ResourceEnergy:
			if newResources.Energy < input.Amount {
				return fmt.Errorf("insufficient energy: need %d, have %d", input.Amount, newResources.Energy)
			}
		case model.ResourceHeat:
			if newResources.Heat < input.Amount {
				return fmt.Errorf("insufficient heat: need %d, have %d", input.Amount, newResources.Heat)
			}

			// TODO: Re-enable card storage resource validation during migration
			// Card storage resources (animals, microbes, floaters, science, asteroid)
			// case model.ResourceAnimals, model.ResourceMicrobes, model.ResourceFloaters, model.ResourceScience, model.ResourceAsteroid:
			// 	// Validate card storage resource inputs
			// 	if input.Target == model.TargetSelfCard {
			// 		// Initialize resource storage map if nil (for checking)
			// 		if player.ResourceStorage == nil {
			// 			player.ResourceStorage = make(map[string]int)
			// 		}

			// 		currentStorage := player.ResourceStorage[action.CardID]
			// 		if currentStorage < input.Amount {
			// 			return fmt.Errorf("insufficient %s storage on card %s: need %d, have %d",
			// 				input.Type, action.CardID, input.Amount, currentStorage)
			// 		}
			// 	}
		}
	}

	// TODO: Re-enable resource storage tracking during migration
	// Track if resource storage was modified
	// resourceStorageModified := false

	// APPLICATION PHASE: Apply each input by deducting resources
	for _, input := range allInputs {
		switch input.Type {
		case model.ResourceCredits:
			newResources.Credits -= input.Amount
		case model.ResourceSteel:
			newResources.Steel -= input.Amount
		case model.ResourceTitanium:
			newResources.Titanium -= input.Amount
		case model.ResourcePlants:
			newResources.Plants -= input.Amount
		case model.ResourceEnergy:
			newResources.Energy -= input.Amount
		case model.ResourceHeat:
			newResources.Heat -= input.Amount

		// TODO: Re-enable card storage resource application during migration
		// Card storage resources (animals, microbes, floaters, science, asteroid)
		// case model.ResourceAnimals, model.ResourceMicrobes, model.ResourceFloaters, model.ResourceScience, model.ResourceAsteroid:
		// 	// Handle card storage resource inputs
		// 	if input.Target == model.TargetSelfCard {
		// 		// Initialize resource storage map if nil
		// 		if player.ResourceStorage == nil {
		// 			player.ResourceStorage = make(map[string]int)
		// 		}

		// 		// Deduct from card storage
		// 		currentStorage := player.ResourceStorage[action.CardID]
		// 		player.ResourceStorage[action.CardID] = currentStorage - input.Amount
		// 		resourceStorageModified = true

		// 		log.Debug("📉 Deducted card storage resource",
		// 			zap.String("card_id", action.CardID),
		// 			zap.String("resource_type", string(input.Type)),
		// 			zap.Int("amount", input.Amount),
		// 			zap.Int("previous_storage", currentStorage),
		// 			zap.Int("new_storage", player.ResourceStorage[action.CardID]))
		// 	} else {
		// 		log.Warn("Card storage input with non-self-card target not supported",
		// 			zap.String("type", string(input.Type)),
		// 			zap.String("target", string(input.Target)))
		// 	}

		default:
			log.Warn("Unknown input resource type during application", zap.String("type", string(input.Type)))
		}

		log.Debug("💰 Applied input",
			zap.String("resource_type", string(input.Type)),
			zap.Int("amount", input.Amount))
	}

	// Update player resources
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
		log.Error("Failed to update player resources for action inputs", zap.Error(err))
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	// TODO: Re-enable resource storage update during migration
	// Update resource storage if modified
	// if resourceStorageModified {
	// 	if err := s.playerRepo.UpdateResourceStorage(ctx, gameID, playerID, player.ResourceStorage); err != nil {
	// 		log.Error("Failed to update resource storage for action inputs", zap.Error(err))
	// 		return fmt.Errorf("failed to update resource storage: %w", err)
	// 	}
	// }

	log.Debug("✅ Action inputs applied successfully")
	return nil
}

// applyActionOutputs applies the action outputs by giving resources/production/etc. to the player
// choiceIndex is optional and used when the action has choices between different effects
func (s *CardServiceImpl) applyActionOutputs(ctx context.Context, gameID, playerID string, action *model.PlayerAction, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to read current resources and production
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
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
		case model.ResourceCredits:
			newResources.Credits += output.Amount
			resourcesChanged = true
		case model.ResourceSteel:
			newResources.Steel += output.Amount
			resourcesChanged = true
		case model.ResourceTitanium:
			newResources.Titanium += output.Amount
			resourcesChanged = true
		case model.ResourcePlants:
			newResources.Plants += output.Amount
			resourcesChanged = true
		case model.ResourceEnergy:
			newResources.Energy += output.Amount
			resourcesChanged = true
		case model.ResourceHeat:
			newResources.Heat += output.Amount
			resourcesChanged = true

		// Production increases
		case model.ResourceCreditsProduction:
			newProduction.Credits += output.Amount
			// Ensure production doesn't go below 0
			if newProduction.Credits < 0 {
				newProduction.Credits = 0
			}
			productionChanged = true
		case model.ResourceSteelProduction:
			newProduction.Steel += output.Amount
			if newProduction.Steel < 0 {
				newProduction.Steel = 0
			}
			productionChanged = true
		case model.ResourceTitaniumProduction:
			newProduction.Titanium += output.Amount
			if newProduction.Titanium < 0 {
				newProduction.Titanium = 0
			}
			productionChanged = true
		case model.ResourcePlantsProduction:
			newProduction.Plants += output.Amount
			if newProduction.Plants < 0 {
				newProduction.Plants = 0
			}
			productionChanged = true
		case model.ResourceEnergyProduction:
			newProduction.Energy += output.Amount
			if newProduction.Energy < 0 {
				newProduction.Energy = 0
			}
			productionChanged = true
		case model.ResourceHeatProduction:
			newProduction.Heat += output.Amount
			if newProduction.Heat < 0 {
				newProduction.Heat = 0
			}
			productionChanged = true

		// TODO: Re-enable terraform rating update during migration
		// Terraform rating
		// case model.ResourceTR:
		// 	if err := s.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, player.TerraformRating+output.Amount); err != nil {
		// 		log.Error("Failed to update terraform rating", zap.Error(err))
		// 		return fmt.Errorf("failed to update terraform rating: %w", err)
		// 	}
		// 	trChanged = true

		// Card storage resources (animals, microbes, floaters, science, asteroid)
		case model.ResourceAnimals, model.ResourceMicrobes, model.ResourceFloaters, model.ResourceScience, model.ResourceAsteroid:
			// Use the CardProcessor's applyCardStorageResource method
			// For actions, the "played card" is the card that has this action
			if err := s.effectProcessor.ApplyCardStorageResource(ctx, gameID, playerID, action.CardID, output, cardStorageTarget, log); err != nil {
				return fmt.Errorf("failed to apply card storage resource for action: %w", err)
			}

		// Card draw/peek/take/buy effects
		case model.ResourceCardDraw:
			cardDrawAmount += output.Amount
		case model.ResourceCardPeek:
			cardPeekAmount += output.Amount
		case model.ResourceCardTake:
			cardTakeAmount += output.Amount
		case model.ResourceCardBuy:
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
		if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
			log.Error("Failed to update player resources for action outputs", zap.Error(err))
			return fmt.Errorf("failed to update player resources: %w", err)
		}
	}

	// Update production if it changed
	if productionChanged {
		if err := s.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
			log.Error("Failed to update player production for action outputs", zap.Error(err))
			return fmt.Errorf("failed to update player production: %w", err)
		}
	}

	// Process card draw/peek/take/buy effects if any were found
	if cardDrawAmount > 0 || cardPeekAmount > 0 || cardTakeAmount > 0 || cardBuyAmount > 0 {
		if err := s.applyActionCardDrawPeekEffects(ctx, gameID, playerID, action.CardID, cardDrawAmount, cardPeekAmount, cardTakeAmount, cardBuyAmount); err != nil {
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
func (s *CardServiceImpl) applyActionCardDrawPeekEffects(ctx context.Context, gameID, playerID, sourceCardID string, cardDrawAmount, cardPeekAmount, cardTakeAmount, cardBuyAmount int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Determine the scenario and create appropriate PendingCardDrawSelection
	var cardsToShow []string
	var freeTakeCount, maxBuyCount int
	var cardBuyCost int = 3 // Default cost for buying cards in Terraforming Mars

	if cardDrawAmount > 0 && cardPeekAmount == 0 && cardTakeAmount == 0 && cardBuyAmount == 0 {
		// Scenario 1: Simple card-draw (e.g., "Draw 2 cards")
		// Draw cards from deck and auto-select all
		drawnCards, err := s.deckRepo.DrawProjectCards(ctx, gameID, cardDrawAmount)
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
		peekedCards, err := s.deckRepo.DrawProjectCards(ctx, gameID, cardPeekAmount)
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

	// TODO: Re-enable pending card draw selection during migration
	// Create PendingCardDrawSelection
	// selection := &model.PendingCardDrawSelection{
	// 	AvailableCards: cardsToShow,
	// 	FreeTakeCount:  freeTakeCount,
	// 	MaxBuyCount:    maxBuyCount,
	// 	CardBuyCost:    cardBuyCost,
	// 	Source:         sourceCardID,
	// }

	// Store in player repository
	// if err := s.playerRepo.UpdatePendingCardDrawSelection(ctx, gameID, playerID, selection); err != nil {
	// 	log.Error("Failed to create pending card draw selection", zap.Error(err))
	// 	return fmt.Errorf("failed to create pending card draw selection: %w", err)
	// }

	log.Info("✅ Pending card draw selection created (from action)",
		zap.String("source_card_id", sourceCardID),
		zap.Int("available_cards", len(cardsToShow)),
		zap.Int("free_take_count", freeTakeCount),
		zap.Int("max_buy_count", maxBuyCount),
		zap.Int("card_buy_cost", cardBuyCost))

	return nil
}

// incrementActionPlayCount increments the play count for a specific action
func (s *CardServiceImpl) incrementActionPlayCount(ctx context.Context, gameID, playerID, cardID string, behaviorIndex int) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to read current actions
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for play count update: %w", err)
	}

	// Find and update the specific action
	updatedActions := make([]model.PlayerAction, len(player.Actions))
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
	if err := s.playerRepo.UpdatePlayerActions(ctx, gameID, playerID, updatedActions); err != nil {
		log.Error("Failed to update player actions for play count", zap.Error(err))
		return fmt.Errorf("failed to update player actions: %w", err)
	}

	return nil
}

// processPendingTileQueues checks for and processes any pending tile queues created by card effects
// This function delegates to TileService.ProcessTileQueue which handles validation and hex calculation
func (s *CardServiceImpl) processPendingTileQueues(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	// TODO: Re-enable pending tile selection queue during migration
	// Get current player to check for pending tile queues
	// player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	// if err != nil {
	// 	return fmt.Errorf("failed to get player for tile queue processing: %w", err)
	// }

	// Check if player has any pending tile selection queue
	// if player.PendingTileSelectionQueue == nil || len(player.PendingTileSelectionQueue.Items) == 0 {
	// 	log.Debug("🏗️ No pending tile queues to process")
	// 	return nil // No tile queues to process
	// }

	// log.Info("🏗️ Processing pending tile queues",
	// 	zap.Int("queue_length", len(player.PendingTileSelectionQueue.Items)),
	// 	zap.String("source", player.PendingTileSelectionQueue.Source))

	// For now, always skip tile queue processing during migration
	log.Debug("🏗️ Tile queue processing disabled during migration")
	return nil

	// Delegate to tile processor which handles tile queue management
	if err := s.tileProcessor.ProcessTileQueue(ctx, gameID, playerID); err != nil {
		log.Error("Failed to process tile queue", zap.Error(err))
		return fmt.Errorf("failed to process tile queue: %w", err)
	}

	log.Debug("✅ Successfully processed pending tile queue")
	return nil
}

// ========== Type Converters: NEW session types → OLD model types ==========
// These converters bridge the gap between NEW session-scoped card types and OLD model types

// cardToModel converts a NEW card.Card to an OLD model.Card
func cardToModel(c card.Card) model.Card {
	return model.Card{
		ID:                 c.ID,
		Name:               c.Name,
		Type:               model.CardType(c.Type),
		Cost:               c.Cost,
		Description:        c.Description,
		Pack:               c.Pack,
		Tags:               c.Tags,
		Requirements:       c.Requirements,
		Behaviors:          c.Behaviors,
		ResourceStorage:    c.ResourceStorage,
		VPConditions:       c.VPConditions,
		StartingResources:  c.StartingResources,
		StartingProduction: c.StartingProduction,
	}
}
