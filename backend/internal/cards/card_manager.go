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
	// choiceIndex is optional and used when the card has choices between different effects
	// cardStorageTarget is optional and used when outputs target "any-card" storage
	CanPlay(ctx context.Context, gameID, playerID, cardID string, choiceIndex *int, cardStorageTarget *string) error

	// PlayCard plays a card (assumes CanPlay validation has passed)
	// choiceIndex is optional and used when the card has choices between different effects
	// cardStorageTarget is optional and used when outputs target "any-card" storage
	PlayCard(ctx context.Context, gameID, playerID, cardID string, choiceIndex *int, cardStorageTarget *string) error
}

// CardManagerImpl implements the simplified card management interface
type CardManagerImpl struct {
	gameRepo              repository.GameRepository
	playerRepo            repository.PlayerRepository
	cardRepo              repository.CardRepository
	requirementsValidator *RequirementsValidator
	effectProcessor       *CardProcessor
	effectSubscriber      CardEffectSubscriber
}

// NewCardManager creates a new simplified card manager
func NewCardManager(
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardRepo repository.CardRepository,
	effectSubscriber CardEffectSubscriber,
) CardManager {
	return &CardManagerImpl{
		gameRepo:              gameRepo,
		playerRepo:            playerRepo,
		cardRepo:              cardRepo,
		requirementsValidator: NewRequirementsValidator(cardRepo),
		effectProcessor:       NewCardProcessor(gameRepo, playerRepo),
		effectSubscriber:      effectSubscriber,
	}
}

// CanPlay validates if a player can play a specific card (card-specific validation only)
// choiceIndex is optional and used when the card has choices between different effects
// cardStorageTarget is optional and used when outputs target "any-card" storage
func (cm *CardManagerImpl) CanPlay(ctx context.Context, gameID, playerID, cardID string, choiceIndex *int, cardStorageTarget *string) error {
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

	// Validate complete affordability (cost + behavioral resource deductions including choice inputs)
	if err := cm.requirementsValidator.ValidateCardAffordability(ctx, gameID, playerID, card, &player, choiceIndex); err != nil {
		return fmt.Errorf("cannot afford to play card: %w", err)
	}

	log.Debug("✅ Card validation passed", zap.String("card_name", card.Name))
	return nil
}

// PlayCard plays a card (assumes CanPlay validation has already passed)
// choiceIndex is optional and used when the card has choices between different effects
// cardStorageTarget is optional and used when outputs target "any-card" storage
func (cm *CardManagerImpl) PlayCard(ctx context.Context, gameID, playerID, cardID string, choiceIndex *int, cardStorageTarget *string) error {
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

	// STEP 3: Initialize resource storage if the card has storage capability
	if card.ResourceStorage != nil {
		// Get current player to access resource storage map
		player, err := cm.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for resource storage init: %w", err)
		}

		// Initialize the map if it's nil
		if player.ResourceStorage == nil {
			player.ResourceStorage = make(map[string]int)
		}

		// Set the starting amount for this card's resource storage
		player.ResourceStorage[cardID] = card.ResourceStorage.Starting

		// Update the player's resource storage
		if err := cm.playerRepo.UpdateResourceStorage(ctx, gameID, playerID, player.ResourceStorage); err != nil {
			return fmt.Errorf("failed to initialize card resource storage: %w", err)
		}

		log.Debug("💾 Card resource storage initialized",
			zap.String("resource_type", string(card.ResourceStorage.Type)),
			zap.Int("starting_amount", card.ResourceStorage.Starting))
	}

	// STEP 4: Apply card effects with choice index and card storage target
	if err := cm.effectProcessor.ApplyCardEffects(ctx, gameID, playerID, card, choiceIndex, cardStorageTarget); err != nil {
		return fmt.Errorf("failed to apply card effects: %w", err)
	}
	log.Debug("✨ Card effects applied")

	// STEP 5: Subscribe passive effects to event bus
	if cm.effectSubscriber != nil {
		if err := cm.effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, cardID, card); err != nil {
			return fmt.Errorf("failed to subscribe card effects: %w", err)
		}
		log.Debug("🎆 Passive effects subscribed to event bus")
	}

	log.Info("✅ Card played successfully", zap.String("card_name", card.Name))
	return nil
}
