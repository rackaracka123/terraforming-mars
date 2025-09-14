package service

import (
	"context"
	"fmt"
	"slices"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// CardService handles card-related operations
type CardService interface {
	// Select starting cards for a player
	SelectStartingCards(ctx context.Context, gameID, playerID string, cardIDs []string) error

	// SelectProductionCards stores the starting card options for a player (called during game start)
	SelectProductionCards(ctx context.Context, gameID, playerID string, cardIDs []string) error

	// Validate starting card selection
	ValidateStartingCardSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error

	// Check if all players have completed card selection
	IsAllPlayersCardSelectionComplete(ctx context.Context, gameID string) bool

	// Get starting cards for selection
	GetStartingCards(ctx context.Context) ([]model.Card, error)

	// Generate starting card options for a player
	GenerateStartingCardOptions(ctx context.Context, gameID, playerID string) ([]model.Card, error)

	// Get card by ID
	GetCardByID(ctx context.Context, cardID string) (*model.Card, error)

	// Play a card from player's hand
	PlayCard(ctx context.Context, gameID, playerID, cardID string) error
}

// CardServiceImpl implements CardService interface
type CardServiceImpl struct {
	gameRepo     repository.GameRepository
	playerRepo   repository.PlayerRepository
	cardRepo     repository.CardRepository
	eventBus     events.EventBus
	cardDeckRepo repository.CardDeckRepository
	cardRegistry *cards.Registry
}

// NewCardService creates a new CardService instance
func NewCardService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, cardRepo repository.CardRepository, eventBus events.EventBus, cardDeckRepo repository.CardDeckRepository) CardService {
	return &CardServiceImpl{
		gameRepo:     gameRepo,
		playerRepo:   playerRepo,
		cardRepo:     cardRepo,
		eventBus:     eventBus,
		cardDeckRepo: cardDeckRepo,
		cardRegistry: cards.NewRegistry(cardRepo, gameRepo, playerRepo),
	}
}

// StorePlayerCardOptions stores the starting card options for a player (called during game start)
func (s *CardServiceImpl) StorePlayerCardOptions(gameID, playerID string, cardOptions []model.Card) {
	ctx := context.Background()
	productionPhase := &model.ProductionPhase{
		AvailableCards:    cardOptions,
		SelectionComplete: false,
	}
	s.playerRepo.SetCardSelection(ctx, gameID, playerID, productionPhase)
}

// SelectStartingCards handles the starting card selection for a player
func (s *CardServiceImpl) SelectStartingCards(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Processing starting card selection", zap.Strings("card_ids", cardIDs))

	// Validate the selection
	if err := s.ValidateStartingCardSelection(ctx, gameID, playerID, cardIDs); err != nil {
		log.Error("Starting card selection validation failed", zap.Error(err))
		return fmt.Errorf("invalid card selection: %w", err)
	}

	// Get current player
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Calculate cost (3 MC per card)
	cost := len(cardIDs) * 3

	// Check if player can afford the selection
	if player.Resources.Credits < cost {
		return fmt.Errorf("insufficient credits: need %d, have %d", cost, player.Resources.Credits)
	}

	// Update player resources with granular update
	updatedResources := player.Resources
	updatedResources.Credits -= cost
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, updatedResources); err != nil {
		log.Error("Failed to update player resources", zap.Error(err))
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	// Add selected cards to player's hand using granular updates
	for _, cardID := range cardIDs {
		if err := s.playerRepo.AddCard(ctx, gameID, playerID, cardID); err != nil {
			log.Error("Failed to add card to player hand", zap.String("card_id", cardID), zap.Error(err))
			return fmt.Errorf("failed to add card %s: %w", cardID, err)
		}
	}

	// Clear the starting selection to mark completion (prevents indefinite prompting)
	if err := s.playerRepo.SetStartingSelection(ctx, gameID, playerID, nil); err != nil {
		log.Error("Failed to clear starting selection", zap.Error(err))
		return fmt.Errorf("failed to clear starting selection: %w", err)
	}

	// Create and publish the starting card selected event
	event := events.NewCardSelectedEvent(gameID, playerID, cardIDs, cost)
	log.Debug("Starting card selected event created",
		zap.String("event_type", event.GetType()),
		zap.Int("cost", cost))

	// Publish the event through the event bus
	if s.eventBus != nil {
		if err := s.eventBus.Publish(ctx, event); err != nil {
			log.Warn("Failed to publish card selected event", zap.Error(err))
		}
	}

	log.Info("Player completed starting card selection",
		zap.Strings("selected_cards", cardIDs),
		zap.Int("cost_paid", cost))

	return nil
}

// SelectProductionCards handles the card selection during production phase (stub implementation)
func (s *CardServiceImpl) SelectProductionCards(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	// For simplicity, assume any selection is valid during production phase
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Processing production card selection", zap.Strings("card_ids", cardIDs))

	// Get current player
	_, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Add selected cards to player's hand using granular updates
	for _, cardID := range cardIDs {
		if err := s.playerRepo.AddCard(ctx, gameID, playerID, cardID); err != nil {
			log.Error("Failed to add card to player hand", zap.String("card_id", cardID), zap.Error(err))
			return fmt.Errorf("failed to add card %s: %w", cardID, err)
		}
	}

	log.Info("Player completed production card selection",
		zap.Strings("selected_cards", cardIDs))

	return nil
}

// ValidateStartingCardSelection validates a player's starting card selection
func (s *CardServiceImpl) ValidateStartingCardSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player to check starting card selection state
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if player has starting cards available for selection
	if len(player.StartingSelection) == 0 {
		return fmt.Errorf("player has no starting cards available for selection")
	}

	// Check if player already has cards in hand (indicating they already completed selection)
	if len(player.Cards) > 0 {
		log.Debug("Player has cards in hand, selection already completed", zap.Int("cards_in_hand", len(player.Cards)))
		return fmt.Errorf("player has already completed card selection")
	}

	playerOptions := player.StartingSelection

	// Check maximum cards (10 is the maximum starting cards dealt)
	if len(cardIDs) > 10 {
		return fmt.Errorf("cannot select more than 10 cards, got %d", len(cardIDs))
	}

	// Validate card IDs exist in the starting cards pool first
	allStartingCards, err := s.cardRepo.GetStartingCardPool(ctx)
	if err != nil {
		return fmt.Errorf("failed to get starting card pool: %w", err)
	}
	cardMap := make(map[string]bool)
	for _, card := range allStartingCards {
		cardMap[card.ID] = true
	}

	for _, cardID := range cardIDs {
		if !cardMap[cardID] {
			return fmt.Errorf("invalid card ID: %s", cardID)
		}
	}

	// Then validate selected cards are in player's options
	optionsMap := make(map[string]bool)
	for _, optionID := range playerOptions {
		optionsMap[optionID] = true
	}

	for _, cardID := range cardIDs {
		if !optionsMap[cardID] {
			return fmt.Errorf("card %s is not in player's available options", cardID)
		}
	}

	return nil
}

// IsAllPlayersCardSelectionComplete checks if all players in the game have completed card selection
func (s *CardServiceImpl) IsAllPlayersCardSelectionComplete(ctx context.Context, gameID string) bool {
	log := logger.WithGameContext(gameID, "")

	// Get current game state to determine which selection type to check
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for selection completion check", zap.Error(err))
		return false
	}

	// Get all players in the game
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get players for selection completion check", zap.Error(err))
		return false
	}

	// If no players exist, selection is not complete
	if len(players) == 0 {
		return false
	}

	// Check completion based on current game phase
	switch game.CurrentPhase {
	case model.GamePhaseStartingCardSelection:
		return s.checkStartingCardSelectionComplete(players, log)
	case model.GamePhaseProductionAndCardDraw:
		return s.checkProductionCardSelectionComplete(players)
	default:
		// For other phases, assume no card selection is needed
		return true
	}
}

// checkStartingCardSelectionComplete checks if all players have starting cards (indicating selection is complete)
func (s *CardServiceImpl) checkStartingCardSelectionComplete(players []model.Player, log *zap.Logger) bool {
	playersWithStartingCards := 0

	// For starting card selection, we check if players have cards in their hand
	// (cards are added to hand after selection is confirmed)
	for _, player := range players {
		if len(player.Cards) > 0 {
			playersWithStartingCards++
			log.Debug("Player has completed starting card selection", zap.String("player_id", player.ID), zap.Int("cards_count", len(player.Cards)))
		} else if len(player.StartingSelection) > 0 {
			log.Debug("Player has starting selection but hasn't confirmed", zap.String("player_id", player.ID))
		} else {
			log.Debug("Player has no starting cards or selection", zap.String("player_id", player.ID))
		}
	}

	// All players must have selected and confirmed their starting cards
	allComplete := playersWithStartingCards == len(players)
	log.Debug("Starting card selection completion check",
		zap.Int("total_players", len(players)),
		zap.Int("players_with_cards", playersWithStartingCards),
		zap.Bool("all_complete", allComplete))

	return allComplete
}

// checkProductionCardSelectionComplete checks production card selection completion (original logic)
func (s *CardServiceImpl) checkProductionCardSelectionComplete(players []model.Player) bool {
	// Track selection states
	playersWithIncompleteSelection := 0
	playersWithSelectionData := 0

	// Check each player's selection status
	for _, player := range players {
		if player.ProductionSelection != nil {
			playersWithSelectionData++
			// If any player has selection data but hasn't completed, selection is not done
			if !player.ProductionSelection.SelectionComplete {
				playersWithIncompleteSelection++
			}
		}
	}

	// If any player has incomplete selection, overall selection is not complete
	if playersWithIncompleteSelection > 0 {
		return false
	}

	// If no players have selection data at all, selection is not complete
	// This handles the case where the selection phase hasn't been initiated yet
	if playersWithSelectionData == 0 {
		return false
	}

	// If we reach here, all players with selection data have completed selection
	return true
}

// ClearGameSelectionData clears temporary selection data for a game (called after selection phase completes)
func (s *CardServiceImpl) ClearGameSelectionData(gameID string) {
	ctx := context.Background()

	// Get all players in the game and clear their selection data
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		logger.WithGameContext(gameID, "").Error("Failed to get players for clearing selection data", zap.Error(err))
		return
	}

	for _, player := range players {
		s.playerRepo.ClearCardSelection(ctx, gameID, player.ID)
	}

	logger.WithGameContext(gameID, "").Debug("Cleared game selection data")
}

// GetStartingCards returns cards available for starting selection
func (s *CardServiceImpl) GetStartingCards(ctx context.Context) ([]model.Card, error) {
	return s.cardRepo.GetStartingCardPool(ctx)
}

// GenerateStartingCardOptions generates 4 random starting card options for a player
func (s *CardServiceImpl) GenerateStartingCardOptions(ctx context.Context, gameID, playerID string) ([]model.Card, error) {
	startingCards, err := s.cardRepo.GetStartingCardPool(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get starting card pool: %w", err)
	}

	// For now, return first 4 cards as options
	// In a full implementation, you'd randomize this selection
	var options []model.Card
	for i, card := range startingCards {
		if i >= 4 {
			break
		}
		options = append(options, card)
	}

	// Store the options for validation
	s.StorePlayerCardOptions(gameID, playerID, options)

	return options, nil
}

// GetCardByID returns a card by its ID
func (s *CardServiceImpl) GetCardByID(ctx context.Context, cardID string) (*model.Card, error) {
	return s.cardRepo.GetCardByID(ctx, cardID)
}

// PlayCard implements playing a card from a player's hand
func (s *CardServiceImpl) PlayCard(ctx context.Context, gameID, playerID, cardID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🎯 Playing card", zap.String("card_id", cardID))

	// Get the player to verify they have the card and available actions
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for card play", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Validate player has available actions
	if player.AvailableActions <= 0 {
		log.Warn("Player has no available actions", zap.Int("available_actions", player.AvailableActions))
		return fmt.Errorf("no actions available: player has %d actions", player.AvailableActions)
	}

	// Check if the player has the card in their hand using slices.Contains
	if !slices.Contains(player.Cards, cardID) {
		log.Warn("Player attempted to play card they don't have", zap.String("card_id", cardID))
		return fmt.Errorf("player does not have card %s", cardID)
	}

	// Get the card data to validate it exists
	card, err := s.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		log.Error("Failed to get card data", zap.String("card_id", cardID), zap.Error(err))
		return fmt.Errorf("failed to get card data: %w", err)
	}

	if card == nil {
		return fmt.Errorf("card %s not found", cardID)
	}

	// Debug card requirements and validate
	log.Debug("🃏 Card details",
		zap.String("card_name", card.Name),
		zap.Int("card_cost", card.Cost),
		zap.Any("requirements", card.Requirements))

	// Check if card has any requirements at all
	hasRequirements := card.Requirements.MinTemperature != nil ||
		card.Requirements.MaxTemperature != nil ||
		card.Requirements.MinOxygen != nil ||
		card.Requirements.MaxOxygen != nil ||
		card.Requirements.MinOceans != nil ||
		card.Requirements.MaxOceans != nil ||
		len(card.Requirements.RequiredTags) > 0 ||
		card.Requirements.RequiredProduction != nil

	log.Debug("🔍 Requirements check",
		zap.Bool("has_requirements", hasRequirements),
		zap.Bool("min_oceans_set", card.Requirements.MinOceans != nil))

	if hasRequirements {
		log.Debug("🚨 Validating card requirements - card has requirements to check")
		// Validate card requirements using the card registry
		game, err := s.gameRepo.GetByID(ctx, gameID)
		if err != nil {
			log.Error("Failed to get game for validation", zap.Error(err))
			return fmt.Errorf("failed to get game: %w", err)
		}
		if err := s.cardRegistry.ValidateCardPlay(card, &game, &player); err != nil {
			log.Warn("❌ Card requirements not met", zap.String("card_id", cardID), zap.Error(err))
			return fmt.Errorf("card requirements not met: %w", err)
		}
		log.Debug("✅ Card requirements validation passed")
	} else {
		log.Debug("⏭️ Skipping validation - card has no requirements")
	}

	// Handle card cost payment
	if card.Cost > 0 {
		if player.Resources.Credits < card.Cost {
			return fmt.Errorf("insufficient credits: need %d, have %d", card.Cost, player.Resources.Credits)
		}

		// Deduct card cost from player's resources
		updatedResources := player.Resources
		updatedResources.Credits -= card.Cost
		if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, updatedResources); err != nil {
			log.Error("Failed to update player resources for card cost", zap.Error(err))
			return fmt.Errorf("failed to update player resources: %w", err)
		}
		log.Debug("💰 Card cost paid", zap.Int("cost", card.Cost), zap.Int("remaining_credits", updatedResources.Credits))
	}

	// Remove the card from player's hand and add to played cards
	err = s.playerRepo.PlayCard(ctx, gameID, playerID, cardID)
	if err != nil {
		log.Error("Failed to play card", zap.String("card_id", cardID), zap.Error(err))
		return fmt.Errorf("failed to play card: %w", err)
	}

	// Apply card effects using the card registry
	if err := s.cardRegistry.ApplyCardEffects(ctx, gameID, playerID, card); err != nil {
		log.Error("Failed to apply card effects", zap.String("card_id", cardID), zap.Error(err))
		return fmt.Errorf("failed to apply card effects: %w", err)
	}

	// Publish card played event
	cardPlayedEvent := events.NewCardPlayedEvent(gameID, playerID, cardID)
	if err := s.eventBus.Publish(ctx, cardPlayedEvent); err != nil {
		log.Warn("Failed to publish card played event", zap.Error(err))
	}

	// Also publish game updated event to trigger client refreshes
	gameUpdatedEvent := events.NewGameUpdatedEvent(gameID)
	if err := s.eventBus.Publish(ctx, gameUpdatedEvent); err != nil {
		log.Warn("Failed to publish game updated event", zap.Error(err))
	}

	// Consume one action now that all card playing steps have succeeded
	newActions := player.AvailableActions - 1
	if err := s.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, newActions); err != nil {
		log.Error("Failed to consume player action", zap.Error(err))
		// Note: Card has already been played and effects applied, but we couldn't consume the action
		// This is a critical error but we don't rollback the entire card play
		return fmt.Errorf("card played but failed to consume action: %w", err)
	}
	log.Debug("🎯 Action consumed", zap.Int("remaining_actions", newActions))

	log.Info("✅ Card played successfully", zap.String("card_id", cardID), zap.String("card_name", card.Name))
	return nil
}

// validateCardRequirements validates that a card's requirements are met
func (s *CardServiceImpl) validateCardRequirements(ctx context.Context, gameID, playerID string, card *model.Card) error {
	// Get current game state for global parameter checks
	game, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	// Get current player state for tag and production checks
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player state: %w", err)
	}

	// Validate global parameter requirements
	if err := s.validateGlobalParameterRequirements(card.Requirements, game.GlobalParameters); err != nil {
		return err
	}

	// Validate tag requirements
	if len(card.Requirements.RequiredTags) > 0 {
		if err := s.validateTagRequirements(card.Requirements.RequiredTags, player); err != nil {
			return err
		}
	}

	// Validate production requirements
	if card.Requirements.RequiredProduction != nil {
		if err := s.validateProductionRequirements(*card.Requirements.RequiredProduction, player.Production); err != nil {
			return err
		}
	}

	return nil
}

// validateGlobalParameterRequirements checks temperature, oxygen, and ocean requirements
func (s *CardServiceImpl) validateGlobalParameterRequirements(requirements model.CardRequirements, globalParams model.GlobalParameters) error {
	log := logger.Get()

	// Log current global parameters and requirements for debugging
	log.Debug("🌍 Validating global parameter requirements",
		zap.Int("current_temperature", globalParams.Temperature),
		zap.Int("current_oxygen", globalParams.Oxygen),
		zap.Int("current_oceans", globalParams.Oceans))

	if requirements.MinTemperature != nil {
		log.Debug("❄️ Checking min temperature", zap.Int("required", *requirements.MinTemperature), zap.Int("current", globalParams.Temperature))
	}
	if requirements.MinOxygen != nil {
		log.Debug("💨 Checking min oxygen", zap.Int("required", *requirements.MinOxygen), zap.Int("current", globalParams.Oxygen))
	}
	if requirements.MinOceans != nil {
		log.Debug("🌊 Checking min oceans", zap.Int("required", *requirements.MinOceans), zap.Int("current", globalParams.Oceans))
	}

	// Check temperature requirements
	if requirements.MinTemperature != nil && globalParams.Temperature < *requirements.MinTemperature {
		return fmt.Errorf("minimum temperature requirement not met: need %d°C, current %d°C", *requirements.MinTemperature, globalParams.Temperature)
	}
	if requirements.MaxTemperature != nil && globalParams.Temperature > *requirements.MaxTemperature {
		return fmt.Errorf("maximum temperature requirement exceeded: limit %d°C, current %d°C", *requirements.MaxTemperature, globalParams.Temperature)
	}

	// Check oxygen requirements
	if requirements.MinOxygen != nil && globalParams.Oxygen < *requirements.MinOxygen {
		return fmt.Errorf("minimum oxygen requirement not met: need %d%%, current %d%%", *requirements.MinOxygen, globalParams.Oxygen)
	}
	if requirements.MaxOxygen != nil && globalParams.Oxygen > *requirements.MaxOxygen {
		return fmt.Errorf("maximum oxygen requirement exceeded: limit %d%%, current %d%%", *requirements.MaxOxygen, globalParams.Oxygen)
	}

	// Check ocean requirements
	if requirements.MinOceans != nil && globalParams.Oceans < *requirements.MinOceans {
		return fmt.Errorf("minimum ocean requirement not met: need %d, current %d", *requirements.MinOceans, globalParams.Oceans)
	}
	if requirements.MaxOceans != nil && globalParams.Oceans > *requirements.MaxOceans {
		return fmt.Errorf("maximum ocean requirement exceeded: limit %d, current %d", *requirements.MaxOceans, globalParams.Oceans)
	}

	return nil
}

// validateTagRequirements checks if player has required card tags
func (s *CardServiceImpl) validateTagRequirements(requiredTags []model.CardTag, player model.Player) error {
	playerTagCounts := s.countPlayerTags(player)

	for _, requiredTag := range requiredTags {
		if playerTagCounts[requiredTag] == 0 {
			return fmt.Errorf("required tag not found: %s", requiredTag)
		}
	}

	return nil
}

// validateProductionRequirements checks if player has sufficient production levels
func (s *CardServiceImpl) validateProductionRequirements(requiredProduction model.ResourceSet, playerProduction model.Production) error {
	if requiredProduction.Credits > 0 && playerProduction.Credits < requiredProduction.Credits {
		return fmt.Errorf("insufficient credit production: need %d, have %d", requiredProduction.Credits, playerProduction.Credits)
	}
	if requiredProduction.Steel > 0 && playerProduction.Steel < requiredProduction.Steel {
		return fmt.Errorf("insufficient steel production: need %d, have %d", requiredProduction.Steel, playerProduction.Steel)
	}
	if requiredProduction.Titanium > 0 && playerProduction.Titanium < requiredProduction.Titanium {
		return fmt.Errorf("insufficient titanium production: need %d, have %d", requiredProduction.Titanium, playerProduction.Titanium)
	}
	if requiredProduction.Plants > 0 && playerProduction.Plants < requiredProduction.Plants {
		return fmt.Errorf("insufficient plant production: need %d, have %d", requiredProduction.Plants, playerProduction.Plants)
	}
	if requiredProduction.Energy > 0 && playerProduction.Energy < requiredProduction.Energy {
		return fmt.Errorf("insufficient energy production: need %d, have %d", requiredProduction.Energy, playerProduction.Energy)
	}
	if requiredProduction.Heat > 0 && playerProduction.Heat < requiredProduction.Heat {
		return fmt.Errorf("insufficient heat production: need %d, have %d", requiredProduction.Heat, playerProduction.Heat)
	}

	return nil
}

// countPlayerTags counts the occurrence of each tag in player's played cards
func (s *CardServiceImpl) countPlayerTags(player model.Player) map[model.CardTag]int {
	tagCounts := make(map[model.CardTag]int)

	// Count tags from played cards
	for _, cardID := range player.PlayedCards {
		card, err := s.cardRepo.GetCardByID(context.Background(), cardID)
		if err != nil || card == nil {
			continue // Skip if card not found
		}

		for _, tag := range card.Tags {
			tagCounts[tag]++
		}
	}

	// Add corporation tags if player has a corporation
	if player.Corporation != nil && *player.Corporation != "" {
		corporationCard, err := s.cardRepo.GetCardByID(context.Background(), *player.Corporation)
		if err == nil && corporationCard != nil {
			for _, tag := range corporationCard.Tags {
				tagCounts[tag]++
			}
		}
	}

	return tagCounts
}

// applyCardEffects applies the production and other effects of a played card
func (s *CardServiceImpl) applyCardEffects(ctx context.Context, gameID, playerID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)

	// Apply production effects if the card has them
	if card.ProductionEffects != nil {
		// Get current player to read current production
		player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
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
		if err := s.playerRepo.UpdateProduction(ctx, gameID, playerID, newProduction); err != nil {
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
	}

	return nil
}
