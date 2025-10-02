package service

import (
	"context"
	"fmt"
	"slices"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// CardService handles card-related operations
type CardService interface {
	// Player actions for card selection and play
	OnSelectStartingCards(ctx context.Context, gameID, playerID string, cardIDs []string) error

	// Player action for production card selection
	OnSelectProductionCards(ctx context.Context, gameID, playerID string, cardIDs []string) error

	// Get starting cards for selection
	GetStartingCards(ctx context.Context) ([]model.Card, error)

	// Get card by ID
	GetCardByID(ctx context.Context, cardID string) (*model.Card, error)

	// Player actions for playing cards
	OnPlayCard(ctx context.Context, gameID, playerID, cardID string, choiceIndex *int) error

	// Play a card action from player's action list
	OnPlayCardAction(ctx context.Context, gameID, playerID, cardID string, behaviorIndex int, choiceIndex *int) error

	// List cards with pagination
	ListCardsPaginated(ctx context.Context, offset, limit int) ([]model.Card, int, error)
}

// CardServiceImpl implements CardService interface using specialized card managers
type CardServiceImpl struct {
	// Core repositories
	gameRepo       repository.GameRepository
	playerRepo     repository.PlayerRepository
	cardRepo       repository.CardRepository
	cardDeckRepo   repository.CardDeckRepository
	sessionManager session.SessionManager

	// Specialized managers from cards package
	selectionManager      *cards.SelectionManager
	requirementsValidator *cards.RequirementsValidator
	effectProcessor       *cards.CardProcessor
	cardManager           cards.CardManager

	// Service dependencies
	tileService TileService
}

// NewCardService creates a new CardService instance
func NewCardService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, cardRepo repository.CardRepository, cardDeckRepo repository.CardDeckRepository, sessionManager session.SessionManager, tileService TileService) CardService {
	return &CardServiceImpl{
		gameRepo:              gameRepo,
		playerRepo:            playerRepo,
		cardRepo:              cardRepo,
		cardDeckRepo:          cardDeckRepo,
		sessionManager:        sessionManager,
		selectionManager:      cards.NewSelectionManager(gameRepo, playerRepo, cardRepo, cardDeckRepo),
		requirementsValidator: cards.NewRequirementsValidator(cardRepo),
		effectProcessor:       cards.NewCardProcessor(gameRepo, playerRepo),
		cardManager:           cards.NewCardManager(gameRepo, playerRepo, cardRepo),
		tileService:           tileService,
	}
}

// Delegation methods - all operations are handled by the specialized cards service

func (s *CardServiceImpl) OnSelectStartingCards(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)

	err := s.selectionManager.SelectStartingCards(ctx, gameID, playerID, cardIDs)
	if err != nil {
		return err
	}

	log.Debug("🃏 Player completed starting card selection", zap.Strings("card_ids", cardIDs))

	// Check if all players have completed their starting card selection
	if s.isAllPlayersCardSelectionComplete(ctx, gameID) {
		log.Info("✅ All players completed starting card selection, advancing to action phase")

		// Get current game state to validate phase transition
		game, err := s.gameRepo.GetByID(ctx, gameID)
		if err != nil {
			log.Error("Failed to get game for phase advancement", zap.Error(err))
			return fmt.Errorf("failed to get game: %w", err)
		}

		// Validate current phase before transition
		if game.CurrentPhase != model.GamePhaseStartingCardSelection {
			log.Warn("Game is not in starting card selection phase, skipping phase transition",
				zap.String("current_phase", string(game.CurrentPhase)))
		} else if game.Status != model.GameStatusActive {
			log.Warn("Game is not active, skipping phase transition",
				zap.String("current_status", string(game.Status)))
		} else {
			// Advance to action phase
			if err := s.gameRepo.UpdatePhase(ctx, gameID, model.GamePhaseAction); err != nil {
				log.Error("Failed to update game phase", zap.Error(err))
				return fmt.Errorf("failed to update game phase: %w", err)
			}

			// Clear temporary card selection data
			s.selectionManager.ClearGameSelectionData(gameID)

			log.Info("🎯 Game phase advanced successfully",
				zap.String("previous_phase", string(model.GamePhaseStartingCardSelection)),
				zap.String("new_phase", string(model.GamePhaseAction)))
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
	return s.selectionManager.SelectProductionCards(ctx, gameID, playerID, cardIDs)
}

// validateStartingCardSelection validates a player's starting card selection (internal use only)
func (s *CardServiceImpl) validateStartingCardSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	return s.selectionManager.ValidateStartingCardSelection(ctx, gameID, playerID, cardIDs)
}

// isAllPlayersCardSelectionComplete checks if all players in the game have completed card selection (internal use only)
func (s *CardServiceImpl) isAllPlayersCardSelectionComplete(ctx context.Context, gameID string) bool {
	return s.selectionManager.IsAllPlayersCardSelectionComplete(ctx, gameID)
}

func (s *CardServiceImpl) ClearGameSelectionData(gameID string) {
	s.selectionManager.ClearGameSelectionData(gameID)
}

func (s *CardServiceImpl) GetStartingCards(ctx context.Context) ([]model.Card, error) {
	return s.cardRepo.GetStartingCardPool(ctx)
}

func (s *CardServiceImpl) GetCardByID(ctx context.Context, cardID string) (*model.Card, error) {
	return s.cardRepo.GetCardByID(ctx, cardID)
}

func (s *CardServiceImpl) OnPlayCard(ctx context.Context, gameID, playerID, cardID string, choiceIndex *int) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🎯 Playing card using simplified interface", zap.String("card_id", cardID))

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

	// -1 Available actions means we have infinite (solo game)
	if player.AvailableActions <= 0 && player.AvailableActions != -1 {
		return fmt.Errorf("no actions available: player has %d actions", player.AvailableActions)
	}

	if !slices.Contains(player.Cards, cardID) {
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

	// STEP 2: Use CardManager for card-specific validation (including choice-based costs)
	if err := s.cardManager.CanPlay(ctx, gameID, playerID, cardID, choiceIndex); err != nil {
		return fmt.Errorf("card cannot be played: %w", err)
	}

	// STEP 3: Use CardManager to play the card with choice index
	if err := s.cardManager.PlayCard(ctx, gameID, playerID, cardID, choiceIndex); err != nil {
		return fmt.Errorf("failed to play card: %w", err)
	}

	// STEP 4: Process any tile queue created by the card
	if err := s.tileService.ProcessTileQueue(ctx, gameID, playerID); err != nil {
		log.Error("Failed to process tile queue", zap.Error(err))
		return fmt.Errorf("card played but failed to process tile queue: %w", err)
	}
	log.Debug("🎯 Tile queue processed (if any existed)")

	// STEP 5: Service-level post-play actions (consume action, broadcast)
	if player.AvailableActions != -1 {
		newActions := player.AvailableActions - 1
		if err := s.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, newActions); err != nil {
			return fmt.Errorf("card played but failed to consume action: %w", err)
		}
		log.Debug("🎯 Action consumed", zap.Int("remaining_actions", newActions))
	}

	// STEP 5: Broadcast game state update
	if err := s.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after card play", zap.Error(err))
		// Don't fail the card play operation, just log the error
	}

	log.Info("✅ Card played successfully", zap.String("card_id", cardID))
	return nil
}

func (s *CardServiceImpl) ListCardsPaginated(ctx context.Context, offset, limit int) ([]model.Card, int, error) {
	// Get all cards from repository
	allCards, err := s.cardRepo.GetAllCards(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get cards: %w", err)
	}

	totalCount := len(allCards)

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

	paginatedCards := allCards[start:end]
	return paginatedCards, totalCount, nil
}

// OnPlayCardAction plays a card action from the player's action list
func (s *CardServiceImpl) OnPlayCardAction(ctx context.Context, gameID, playerID, cardID string, behaviorIndex int, choiceIndex *int) error {
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

	// Check if player has available actions
	if player.AvailableActions <= 0 && player.AvailableActions != -1 {
		return fmt.Errorf("no available actions remaining")
	}

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
	if err := s.applyActionOutputs(ctx, gameID, playerID, targetAction, choiceIndex); err != nil {
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
		log.Debug("🎯 Action consumed", zap.Int("remaining_actions", newActions))
	} else {
		log.Debug("🎯 Action consumed (unlimited actions)", zap.Int("available_actions", -1))
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

	// Apply each input by deducting resources
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

	log.Debug("✅ Action inputs applied successfully")
	return nil
}

// applyActionOutputs applies the action outputs by giving resources/production/etc. to the player
// choiceIndex is optional and used when the action has choices between different effects
func (s *CardServiceImpl) applyActionOutputs(ctx context.Context, gameID, playerID string, action *model.PlayerAction, choiceIndex *int) error {
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

		// Terraform rating
		case model.ResourceTR:
			if err := s.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, player.TerraformRating+output.Amount); err != nil {
				log.Error("Failed to update terraform rating", zap.Error(err))
				return fmt.Errorf("failed to update terraform rating: %w", err)
			}
			trChanged = true

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

	log.Debug("✅ Action outputs applied successfully",
		zap.Bool("resources_changed", resourcesChanged),
		zap.Bool("production_changed", productionChanged),
		zap.Bool("tr_changed", trChanged))
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

	// Get current player to check for pending tile queues
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player for tile queue processing: %w", err)
	}

	// Check if player has any pending tile selection queue
	if player.PendingTileSelectionQueue == nil || len(player.PendingTileSelectionQueue.Items) == 0 {
		log.Debug("🏗️ No pending tile queues to process")
		return nil // No tile queues to process
	}

	log.Info("🏗️ Processing pending tile queues",
		zap.Int("queue_length", len(player.PendingTileSelectionQueue.Items)),
		zap.String("source", player.PendingTileSelectionQueue.Source))

	// Delegate to TileService which handles validation, board service integration, etc.
	if err := s.tileService.ProcessTileQueue(ctx, gameID, playerID); err != nil {
		log.Error("Failed to process tile queue", zap.Error(err))
		return fmt.Errorf("failed to process tile queue: %w", err)
	}

	log.Debug("✅ Successfully processed pending tile queue")
	return nil
}
