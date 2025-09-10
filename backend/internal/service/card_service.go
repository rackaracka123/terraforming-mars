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
	GetStartingCards() []model.Card

	// Generate starting card options for a player
	GenerateStartingCardOptions(gameID, playerID string) []string

	// Get the underlying card data service
	GetCardDataService() CardDataService

	// Play a card with multi-currency payment
	PlayCard(ctx context.Context, gameID, playerID, cardID string, payment *model.Payment) error

	// Get payment cost for a specific card
	GetCardPaymentCost(cardID string) (*model.PaymentCost, error)

	// Validate if a payment is valid for a card
	ValidateCardPayment(cardID string, payment *model.Payment, playerResources *model.Resources) error
}

// CardServiceImpl implements CardService interface
type CardServiceImpl struct {
	gameRepo        repository.GameRepository
	playerRepo      repository.PlayerRepository
	cardDataService CardDataService
	paymentService  PaymentService
	// Store starting card options temporarily during selection phase
	playerCardOptions map[string]map[string][]string // gameID -> playerID -> cardOptions
	selectionStatus   map[string]map[string]bool     // gameID -> playerID -> hasSelected
}

// NewCardService creates a new CardService instance
func NewCardService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, cardDataService CardDataService, paymentService PaymentService) CardService {
	return &CardServiceImpl{
		gameRepo:          gameRepo,
		playerRepo:        playerRepo,
		cardDataService:   cardDataService,
		paymentService:    paymentService,
		playerCardOptions: make(map[string]map[string][]string),
		selectionStatus:   make(map[string]map[string]bool),
	}
}

// StorePlayerCardOptions stores the starting card options for a player (called during game start)
func (s *CardServiceImpl) StorePlayerCardOptions(gameID, playerID string, cardOptions []string) {
	if s.playerCardOptions[gameID] == nil {
		s.playerCardOptions[gameID] = make(map[string][]string)
	}
	if s.selectionStatus[gameID] == nil {
		s.selectionStatus[gameID] = make(map[string]bool)
	}

	s.playerCardOptions[gameID][playerID] = cardOptions
	s.selectionStatus[gameID][playerID] = false
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

	// Calculate cost (first card free, 3 MC per additional card)
	cost := 0
	if len(cardIDs) > 0 {
		cost = (len(cardIDs) - 1) * 3
	}

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
	if s.selectionStatus[gameID] != nil {
		s.selectionStatus[gameID][playerID] = true
	}

	// Create and log the starting card selected event
	event := events.NewCardSelectedEvent(gameID, playerID, cardIDs, cost)
	log.Debug("Starting card selected event created",
		zap.String("event_type", event.GetType()),
		zap.Int("cost", cost))

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
	gameOptions, exists := s.playerCardOptions[gameID]
	if !exists {
		return fmt.Errorf("no card options found for game")
	}

	playerOptions, exists := gameOptions[playerID]
	if !exists {
		return fmt.Errorf("no card options found for player")
	}

	// Check if player already selected cards
	if gameStatus, exists := s.selectionStatus[gameID]; exists {
		if hasSelected, exists := gameStatus[playerID]; exists && hasSelected {
			return fmt.Errorf("player has already selected starting cards")
		}
	}

	// Check maximum cards (4 is the maximum starting cards dealt)
	if len(cardIDs) > 4 {
		return fmt.Errorf("cannot select more than 4 cards, got %d", len(cardIDs))
	}

	// Validate card IDs exist in the starting cards pool first
	allStartingCards := s.cardDataService.GetStartingCardPool()
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
	// Get players for checking selection completion
	players, err := s.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		logger.WithGameContext(gameID, "").Error("Failed to get players for selection completion check", zap.Error(err))
		return false
	}

	// Check if we have selection status for this game
	gameStatus, exists := s.selectionStatus[gameID]
	if !exists {
		return false
	}

	// Check if all players have completed selection
	for _, player := range players {
		hasSelected, exists := gameStatus[player.ID]
		if !exists || !hasSelected {
			return false
		}
	}

	return true
}

// ClearGameSelectionData clears temporary selection data for a game (called after selection phase completes)
func (s *CardServiceImpl) ClearGameSelectionData(gameID string) {
	delete(s.playerCardOptions, gameID)
	delete(s.selectionStatus, gameID)

	logger.WithGameContext(gameID, "").Debug("Cleared game selection data")
}

// GetStartingCards returns cards available for starting selection
func (s *CardServiceImpl) GetStartingCards() []model.Card {
	return s.cardDataService.GetStartingCardPool()
}

// GenerateStartingCardOptions generates 4 random starting card options for a player
func (s *CardServiceImpl) GenerateStartingCardOptions(gameID, playerID string) []string {
	startingCards := s.cardDataService.GetStartingCardPool()

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

	return options
}

// GetCardDataService returns the underlying card data service
func (s *CardServiceImpl) GetCardDataService() CardDataService {
	return s.cardDataService
}

// PlayCard handles playing a card with multi-currency payment
func (s *CardServiceImpl) PlayCard(ctx context.Context, gameID, playerID, cardID string, payment *model.Payment) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("Processing card play", zap.String("card_id", cardID))

	// Get the card details
	card, err := s.cardDataService.GetCardByID(cardID)
	if err != nil {
		log.Error("Failed to get card details", zap.String("card_id", cardID), zap.Error(err))
		return fmt.Errorf("failed to get card details: %w", err)
	}

	// Get current player
	player, err := s.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Validate the payment
	if err := s.ValidateCardPayment(cardID, payment, &player.Resources); err != nil {
		log.Error("Payment validation failed", zap.Error(err))
		return fmt.Errorf("invalid payment: %w", err)
	}

	// Process the payment
	newResources, err := s.paymentService.ProcessPayment(payment, &player.Resources)
	if err != nil {
		log.Error("Failed to process payment", zap.Error(err))
		return fmt.Errorf("failed to process payment: %w", err)
	}

	// Update player resources
	if err := s.playerRepo.UpdateResources(ctx, gameID, playerID, *newResources); err != nil {
		log.Error("Failed to update player resources", zap.Error(err))
		return fmt.Errorf("failed to update player resources: %w", err)
	}

	// Play the card (moves from hand to played cards)
	if err := s.playerRepo.PlayCard(ctx, gameID, playerID, cardID); err != nil {
		log.Error("Failed to play card", zap.String("card_id", cardID), zap.Error(err))
		return fmt.Errorf("failed to play card: %w", err)
	}

	log.Info("Player played card successfully",
		zap.String("card_name", card.Name),
		zap.String("card_id", cardID),
		zap.Int("credits_paid", payment.Credits),
		zap.Int("steel_paid", payment.Steel),
		zap.Int("titanium_paid", payment.Titanium))

	return nil
}

// GetCardPaymentCost returns the payment cost structure for a specific card
func (s *CardServiceImpl) GetCardPaymentCost(cardID string) (*model.PaymentCost, error) {
	card, err := s.cardDataService.GetCardByID(cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card details: %w", err)
	}

	return s.paymentService.GetCardPaymentCost(card), nil
}

// ValidateCardPayment validates if a payment is valid for a card and player resources
func (s *CardServiceImpl) ValidateCardPayment(cardID string, payment *model.Payment, playerResources *model.Resources) error {
	// Get the card payment cost
	paymentCost, err := s.GetCardPaymentCost(cardID)
	if err != nil {
		return fmt.Errorf("failed to get payment cost: %w", err)
	}

	// Check if player can afford the payment
	if !s.paymentService.CanAfford(payment, playerResources) {
		return fmt.Errorf("insufficient resources for payment")
	}

	// Check if the payment is valid for this card
	if !s.paymentService.IsValidPayment(payment, paymentCost) {
		return fmt.Errorf("invalid payment method for this card")
	}

	return nil
}
