package card

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/session/deck"
	sessionGame "terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// CardManager provides a simplified interface for card validation and playing in session-scoped architecture
type CardManager interface {
	// CanPlay checks if a player can play a specific card (card validation only)
	// payment is the proposed payment method (credits, steel, titanium) for the card cost
	// choiceIndex is optional and used when the card has choices between different effects
	// cardStorageTarget is optional and used when outputs target "any-card" storage
	CanPlay(ctx context.Context, gameID, playerID, cardID string, payment *model.CardPayment, choiceIndex *int, cardStorageTarget *string) error

	// PlayCard plays a card (assumes CanPlay validation has passed)
	// payment is the payment method to use for the card cost
	// choiceIndex is optional and used when the card has choices between different effects
	// cardStorageTarget is optional and used when outputs target "any-card" storage
	PlayCard(ctx context.Context, gameID, playerID, cardID string, payment *model.CardPayment, choiceIndex *int, cardStorageTarget *string) error
}

// CardManagerImpl implements the simplified card management interface with session-scoped repositories
type CardManagerImpl struct {
	gameRepo              sessionGame.Repository
	playerRepo            player.Repository
	cardRepo              Repository
	requirementsValidator *RequirementsValidator
	effectProcessor       *CardProcessor
	effectSubscriber      CardEffectSubscriber
}

// NewCardManager creates a new simplified card manager with session-scoped repositories
func NewCardManager(
	gameRepo sessionGame.Repository,
	playerRepo player.Repository,
	cardRepo Repository,
	deckRepo deck.Repository,
	effectSubscriber CardEffectSubscriber,
) CardManager {
	return &CardManagerImpl{
		gameRepo:              gameRepo,
		playerRepo:            playerRepo,
		cardRepo:              cardRepo,
		requirementsValidator: NewRequirementsValidator(cardRepo),
		effectProcessor:       NewCardProcessor(gameRepo, playerRepo, deckRepo),
		effectSubscriber:      effectSubscriber,
	}
}

// CanPlay validates if a player can play a specific card (card-specific validation only)
// payment is the proposed payment method (credits, steel, titanium) for the card cost
// choiceIndex is optional and used when the card has choices between different effects
// cardStorageTarget is optional and used when outputs target "any-card" storage
func (cm *CardManagerImpl) CanPlay(ctx context.Context, gameID, playerID, cardID string, payment *model.CardPayment, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🔍 Validating card requirements and affordability", zap.String("card_id", cardID))

	// Get game and player data for card validation
	gameModel, err := cm.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	playerModel, err := cm.playerRepo.GetByID(ctx, gameID, playerID)
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
		if err := cm.requirementsValidator.ValidateCardRequirements(ctx, gameID, playerID, card, gameModel, playerModel); err != nil {
			return fmt.Errorf("card requirements not met: %w", err)
		}
	}

	// Validate complete affordability (cost + behavioral resource deductions including choice inputs and payment)
	if err := cm.requirementsValidator.ValidateCardAffordability(ctx, gameID, playerID, card, playerModel, payment, choiceIndex); err != nil {
		return fmt.Errorf("cannot afford to play card: %w", err)
	}

	log.Debug("✅ Card validation passed", zap.String("card_name", card.Name))
	return nil
}

// PlayCard plays a card (assumes CanPlay validation has already passed)
// payment is the payment method to use for the card cost
// choiceIndex is optional and used when the card has choices between different effects
// cardStorageTarget is optional and used when outputs target "any-card" storage
func (cm *CardManagerImpl) PlayCard(ctx context.Context, gameID, playerID, cardID string, payment *model.CardPayment, choiceIndex *int, cardStorageTarget *string) error {
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
	playerModel, err := cm.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// STEP 1: Apply card cost payment (using credits, steel, titanium, and/or payment substitutes)
	if card.Cost > 0 {
		updatedResources := playerModel.Resources
		updatedResources.Credits -= payment.Credits
		updatedResources.Steel -= payment.Steel
		updatedResources.Titanium -= payment.Titanium

		// Deduct payment substitutes (e.g., heat for Helion)
		if payment.Substitutes != nil {
			for resourceType, amount := range payment.Substitutes {
				switch resourceType {
				case model.ResourceHeat:
					updatedResources.Heat -= amount
				case model.ResourceEnergy:
					updatedResources.Energy -= amount
				case model.ResourcePlants:
					updatedResources.Plants -= amount
				// Add other resource types as needed
				default:
					log.Warn("Unknown payment substitute resource type", zap.String("resource_type", string(resourceType)))
				}
			}
		}

		if err := cm.playerRepo.UpdateResources(ctx, gameID, playerID, updatedResources); err != nil {
			return fmt.Errorf("failed to update player resources: %w", err)
		}

		// Log payment details
		logFields := []zap.Field{
			zap.Int("cost", card.Cost),
			zap.Int("credits_spent", payment.Credits),
			zap.Int("steel_spent", payment.Steel),
			zap.Int("titanium_spent", payment.Titanium),
		}
		if payment.Substitutes != nil && len(payment.Substitutes) > 0 {
			for resourceType, amount := range payment.Substitutes {
				logFields = append(logFields, zap.Int(string(resourceType)+"_spent", amount))
			}
		}
		log.Debug("💰 Card cost paid", logFields...)
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
		playerModel, err := cm.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for resource storage init: %w", err)
		}

		// Initialize the map if it's nil
		if playerModel.ResourceStorage == nil {
			playerModel.ResourceStorage = make(map[string]int)
		}

		// Set the starting amount for this card's resource storage
		playerModel.ResourceStorage[cardID] = card.ResourceStorage.Starting

		// Update the player's resource storage
		if err := cm.playerRepo.UpdateResourceStorage(ctx, gameID, playerID, playerModel.ResourceStorage); err != nil {
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
		if err := cm.effectSubscriber.SubscribeCardEffects(ctx, gameID, playerID, cardID, card.ToModelCard()); err != nil {
			return fmt.Errorf("failed to subscribe card effects: %w", err)
		}
		log.Debug("🎆 Passive effects subscribed to event bus")
	}

	log.Info("✅ Card played successfully", zap.String("card_name", card.Name))
	return nil
}
