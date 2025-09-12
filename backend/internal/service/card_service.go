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
	GenerateStartingCardOptions(ctx context.Context, gameID, playerID string) ([]string, error)

	// Get card by ID
	GetCardByID(ctx context.Context, cardID string) (*model.Card, error)
}

// CardServiceImpl implements CardService interface
type CardServiceImpl struct {
	gameRepo          repository.GameRepository
	playerRepo        repository.PlayerRepository
	cardRepo          repository.CardRepository
	eventBus          events.EventBus
	cardDeckRepo      repository.CardDeckRepository
	cardSelectionRepo repository.CardSelectionRepository
}

// NewCardService creates a new CardService instance
func NewCardService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, cardRepo repository.CardRepository, eventBus events.EventBus, cardDeckRepo repository.CardDeckRepository, cardSelectionRepo repository.CardSelectionRepository) CardService {
	return &CardServiceImpl{
		gameRepo:          gameRepo,
		playerRepo:        playerRepo,
		cardRepo:          cardRepo,
		eventBus:          eventBus,
		cardDeckRepo:      cardDeckRepo,
		cardSelectionRepo: cardSelectionRepo,
	}
}

// StorePlayerCardOptions stores the starting card options for a player (called during game start)
func (s *CardServiceImpl) StorePlayerCardOptions(gameID, playerID string, cardOptions []string) {
	ctx := context.Background()
	s.cardSelectionRepo.StorePlayerOptions(ctx, gameID, playerID, cardOptions)
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

	// Mark player as having completed selection
	s.cardSelectionRepo.MarkSelectionComplete(ctx, gameID, playerID)

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
	// Check if player has card options stored
	playerOptions, err := s.cardSelectionRepo.GetPlayerOptions(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player options: %w", err)
	}
	if playerOptions == nil {
		return fmt.Errorf("no card options found for player")
	}

	// Check if player already selected cards
	hasSelected, err := s.cardSelectionRepo.IsSelectionComplete(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to check selection status: %w", err)
	}
	if hasSelected {
		return fmt.Errorf("player has already selected starting cards")
	}

	// Check maximum cards (4 is the maximum starting cards dealt)
	if len(cardIDs) > 4 {
		return fmt.Errorf("cannot select more than 4 cards, got %d", len(cardIDs))
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
		optionsMap[option] = true
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
	// Use the repository to check if all selections are complete
	allComplete, err := s.cardSelectionRepo.IsAllSelectionsComplete(ctx, gameID)
	if err != nil {
		logger.WithGameContext(gameID, "").Error("Failed to check selection completion", zap.Error(err))
		return false
	}

	return allComplete
}

// ClearGameSelectionData clears temporary selection data for a game (called after selection phase completes)
func (s *CardServiceImpl) ClearGameSelectionData(gameID string) {
	ctx := context.Background()
	s.cardSelectionRepo.Clear(ctx, gameID)

	logger.WithGameContext(gameID, "").Debug("Cleared game selection data")
}

// GetStartingCards returns cards available for starting selection
func (s *CardServiceImpl) GetStartingCards(ctx context.Context) ([]model.Card, error) {
	return s.cardRepo.GetStartingCardPool(ctx)
}

// GenerateStartingCardOptions generates 4 random starting card options for a player
func (s *CardServiceImpl) GenerateStartingCardOptions(ctx context.Context, gameID, playerID string) ([]string, error) {
	startingCards, err := s.cardRepo.GetStartingCardPool(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get starting card pool: %w", err)
	}

	// For now, return first 4 cards as options
	// In a full implementation, you'd randomize this selection
	var options []string
	for i, card := range startingCards {
		if i >= 4 {
			break
		}
		options = append(options, card.ID)
	}

	// Store the options for validation
	s.StorePlayerCardOptions(gameID, playerID, options)

	return options, nil
}

// GetCardByID returns a card by its ID
func (s *CardServiceImpl) GetCardByID(ctx context.Context, cardID string) (*model.Card, error) {
	return s.cardRepo.GetCardByID(ctx, cardID)
}
