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

	// Validate starting card selection
	ValidateStartingCardSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error

	// Check if all players have completed card selection
	IsAllPlayersCardSelectionComplete(ctx context.Context, gameID string) bool
}

// CardServiceImpl implements CardService interface
type CardServiceImpl struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
	// Store starting card options temporarily during selection phase
	playerCardOptions map[string]map[string][]string // gameID -> playerID -> cardOptions
	selectionStatus   map[string]map[string]bool     // gameID -> playerID -> hasSelected
}

// NewCardService creates a new CardService instance
func NewCardService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) CardService {
	return &CardServiceImpl{
		gameRepo:          gameRepo,
		playerRepo:        playerRepo,
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
	allStartingCards := model.GetStartingCards()
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
