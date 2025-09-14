package service

import (
	"context"
	"fmt"

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
}

// CardServiceImpl implements CardService interface
type CardServiceImpl struct {
	gameRepo     repository.GameRepository
	playerRepo   repository.PlayerRepository
	cardRepo     repository.CardRepository
	eventBus     events.EventBus
	cardDeckRepo repository.CardDeckRepository
}

// NewCardService creates a new CardService instance
func NewCardService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, cardRepo repository.CardRepository, eventBus events.EventBus, cardDeckRepo repository.CardDeckRepository) CardService {
	return &CardServiceImpl{
		gameRepo:     gameRepo,
		playerRepo:   playerRepo,
		cardRepo:     cardRepo,
		eventBus:     eventBus,
		cardDeckRepo: cardDeckRepo,
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
	if player.StartingSelection == nil || len(player.StartingSelection) == 0 {
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
	for _, option := range playerOptions {
		optionsMap[option.ID] = true
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
	if game.CurrentPhase == model.GamePhaseStartingCardSelection {
		return s.checkStartingCardSelectionComplete(players, log)
	} else if game.CurrentPhase == model.GamePhaseProductionAndCardDraw {
		return s.checkProductionCardSelectionComplete(players, log)
	}

	// For other phases, assume no card selection is needed
	return true
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
		} else if player.StartingSelection != nil && len(player.StartingSelection) > 0 {
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
func (s *CardServiceImpl) checkProductionCardSelectionComplete(players []model.Player, log *zap.Logger) bool {
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
