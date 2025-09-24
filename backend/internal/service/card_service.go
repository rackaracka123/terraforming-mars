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
	// Select starting cards for a player (immediately commits to hand)
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

	// List cards with pagination
	ListCardsPaginated(ctx context.Context, offset, limit int) ([]model.Card, int, error)
}

// CardServiceImpl implements CardService interface using specialized card managers
type CardServiceImpl struct {
	// Core repositories
	gameRepo     repository.GameRepository
	playerRepo   repository.PlayerRepository
	cardRepo     repository.CardRepository
	eventBus     events.EventBus
	cardDeckRepo repository.CardDeckRepository

	// Specialized managers from cards package
	selectionManager      *cards.SelectionManager
	requirementsValidator *cards.RequirementsValidator
	effectProcessor       *cards.EffectProcessor
	tagManager            *cards.TagManager
}

// NewCardService creates a new CardService instance
func NewCardService(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, cardRepo repository.CardRepository, eventBus events.EventBus, cardDeckRepo repository.CardDeckRepository) CardService {
	return &CardServiceImpl{
		gameRepo:              gameRepo,
		playerRepo:            playerRepo,
		cardRepo:              cardRepo,
		eventBus:              eventBus,
		cardDeckRepo:          cardDeckRepo,
		selectionManager:      cards.NewSelectionManager(gameRepo, playerRepo, cardRepo, eventBus, cardDeckRepo),
		requirementsValidator: cards.NewRequirementsValidator(cardRepo),
		effectProcessor:       cards.NewEffectProcessor(gameRepo, playerRepo),
		tagManager:            cards.NewTagManager(cardRepo),
	}
}

// Delegation methods - all operations are handled by the specialized cards service

func (s *CardServiceImpl) SelectStartingCards(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	return s.selectionManager.SelectStartingCards(ctx, gameID, playerID, cardIDs)
}

func (s *CardServiceImpl) SelectProductionCards(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	return s.selectionManager.SelectProductionCards(ctx, gameID, playerID, cardIDs)
}

func (s *CardServiceImpl) ValidateStartingCardSelection(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	return s.selectionManager.ValidateStartingCardSelection(ctx, gameID, playerID, cardIDs)
}

// IsAllPlayersCardSelectionComplete checks if all players in the game have completed card selection
func (s *CardServiceImpl) IsAllPlayersCardSelectionComplete(ctx context.Context, gameID string) bool {
	return s.selectionManager.IsAllPlayersCardSelectionComplete(ctx, gameID)
}

func (s *CardServiceImpl) ClearGameSelectionData(gameID string) {
	s.selectionManager.ClearGameSelectionData(gameID)
}

func (s *CardServiceImpl) GetStartingCards(ctx context.Context) ([]model.Card, error) {
	return s.cardRepo.GetStartingCardPool(ctx)
}

func (s *CardServiceImpl) GenerateStartingCardOptions(ctx context.Context, gameID, playerID string) ([]model.Card, error) {
	return s.selectionManager.GenerateStartingCardOptions(ctx, gameID, playerID)
}

func (s *CardServiceImpl) GetCardByID(ctx context.Context, cardID string) (*model.Card, error) {
	return s.cardRepo.GetCardByID(ctx, cardID)
}

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

	// Check if the player has the card in their hand
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

	// Validate card requirements using enhanced validator
	if s.requirementsValidator.HasRequirements(card) {
		log.Debug("🚨 Validating card requirements - card has requirements to check")
		game, err := s.gameRepo.GetByID(ctx, gameID)
		if err != nil {
			log.Error("Failed to get game for validation", zap.Error(err))
			return fmt.Errorf("failed to get game: %w", err)
		}
		if err := s.requirementsValidator.ValidateCardRequirements(ctx, gameID, playerID, card, &game, &player); err != nil {
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

	// Apply card effects using the effect processor
	if err := s.effectProcessor.ApplyCardEffects(ctx, gameID, playerID, card); err != nil {
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
