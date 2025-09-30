package cards

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// CardManager provides a simplified interface for card validation and playing
type CardManager interface {
	// CanPlay checks if a player can play a specific card (card validation only)
	CanPlay(ctx context.Context, gameID, playerID, cardID string) error

	// PlayCard plays a card (assumes CanPlay validation has passed)
	PlayCard(ctx context.Context, gameID, playerID, cardID string) error
}

// CardManagerImpl implements the simplified card management interface
type CardManagerImpl struct {
	gameRepo              repository.GameRepository
	playerRepo            repository.PlayerRepository
	cardRepo              repository.CardRepository
	requirementsValidator *RequirementsValidator
	effectProcessor       *CardProcessor
}

// NewCardManager creates a new simplified card manager
func NewCardManager(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, cardRepo repository.CardRepository) CardManager {
	return &CardManagerImpl{
		gameRepo:              gameRepo,
		playerRepo:            playerRepo,
		cardRepo:              cardRepo,
		requirementsValidator: NewRequirementsValidator(cardRepo),
		effectProcessor:       NewCardProcessor(gameRepo, playerRepo),
	}
}

// CanPlay validates if a player can play a specific card (card-specific validation only)
func (cm *CardManagerImpl) CanPlay(ctx context.Context, gameID, playerID, cardID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🔍 Validating card requirements and affordability", zap.String("card_id", cardID))

	// Get game and player data for card validation
	game, err := cm.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	player, err := cm.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Get and validate card data
	card, err := cm.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("failed to get card data: %w", err)
	}

	if card == nil {
		return fmt.Errorf("card %s not found", cardID)
	}

	// Validate card requirements (global parameters, tags, production, etc.)
	if cm.requirementsValidator.HasRequirements(card) {
		if err := cm.requirementsValidator.ValidateCardRequirements(ctx, gameID, playerID, card, &game, &player); err != nil {
			return fmt.Errorf("card requirements not met: %w", err)
		}
	}

	// Validate complete affordability (cost + behavioral resource deductions)
	if err := cm.requirementsValidator.ValidateCardAffordability(ctx, gameID, playerID, card, &player); err != nil {
		return fmt.Errorf("cannot afford to play card: %w", err)
	}

	log.Debug("✅ Card validation passed", zap.String("card_name", card.Name))
	return nil
}

// PlayCard plays a card (assumes CanPlay validation has already passed)
func (cm *CardManagerImpl) PlayCard(ctx context.Context, gameID, playerID, cardID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🎮 Playing card", zap.String("card_id", cardID))

	// Get card data (we need this for cost and effects)
	card, err := cm.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("failed to get card data: %w", err)
	}

	if card == nil {
		return fmt.Errorf("card %s not found", cardID)
	}

	// Get player data for resource updates
	player, err := cm.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// STEP 1: Apply card cost payment
	if card.Cost > 0 {
		updatedResources := player.Resources
		updatedResources.Credits -= card.Cost
		if err := cm.playerRepo.UpdateResources(ctx, gameID, playerID, updatedResources); err != nil {
			return fmt.Errorf("failed to update player resources: %w", err)
		}
		log.Debug("💰 Card cost paid", zap.Int("cost", card.Cost))
	}

	// STEP 2: Move card from hand to played cards
	err = cm.playerRepo.RemoveCardFromHand(ctx, gameID, playerID, cardID)
	if err != nil {
		return fmt.Errorf("failed to play card: %w", err)
	}
	log.Debug("🃏 Card moved to played cards")

	// STEP 3: Apply card effects
	if err := cm.effectProcessor.ApplyCardEffects(ctx, gameID, playerID, card); err != nil {
		return fmt.Errorf("failed to apply card effects: %w", err)
	}
	log.Debug("✨ Card effects applied")

	log.Info("✅ Card played successfully", zap.String("card_name", card.Name))
	return nil
}
