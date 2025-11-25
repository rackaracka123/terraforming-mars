package card

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/game/deck"
	"terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// CardManager provides a simplified interface for card validation and playing in session-scoped architecture
type CardManager interface {
	// CanPlay checks if a player can play a specific card (card validation only)
	// payment is the proposed payment method (credits, steel, titanium) for the card cost
	// choiceIndex is optional and used when the card has choices between different effects
	// cardStorageTarget is optional and used when outputs target "any-card" storage
	CanPlay(ctx context.Context, game *types.Game, p *player.Player, cardID string, payment *types.CardPayment, choiceIndex *int, cardStorageTarget *string) error

	// PlayCard plays a card (assumes CanPlay validation has passed)
	// payment is the payment method to use for the card cost
	// choiceIndex is optional and used when the card has choices between different effects
	// cardStorageTarget is optional and used when outputs target "any-card" storage
	PlayCard(ctx context.Context, game *types.Game, p *player.Player, cardID string, payment *types.CardPayment, choiceIndex *int, cardStorageTarget *string) error
}

// CardManagerImpl implements the simplified card management interface with session-scoped repositories
type CardManagerImpl struct {
	cardRepo              Repository
	requirementsValidator *RequirementsValidator
	effectProcessor       *CardProcessor
	effectSubscriber      CardEffectSubscriber
}

// NewCardManager creates a new simplified card manager with session-scoped repositories
func NewCardManager(
	cardRepo Repository,
	deckRepo deck.Repository,
	effectSubscriber CardEffectSubscriber,
) CardManager {
	return &CardManagerImpl{
		cardRepo:              cardRepo,
		requirementsValidator: NewRequirementsValidator(cardRepo),
		effectProcessor:       NewCardProcessor(deckRepo),
		effectSubscriber:      effectSubscriber,
	}
}

// CanPlay validates if a player can play a specific card (card-specific validation only)
// payment is the proposed payment method (credits, steel, titanium) for the card cost
// choiceIndex is optional and used when the card has choices between different effects
// cardStorageTarget is optional and used when outputs target "any-card" storage
func (cm *CardManagerImpl) CanPlay(ctx context.Context, game *types.Game, p *player.Player, cardID string, payment *types.CardPayment, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(p.GameID, p.ID)
	log.Debug("🔍 Validating card requirements and affordability", zap.String("card_id", cardID))

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
		if err := cm.requirementsValidator.ValidateCardRequirements(ctx, game, p, card); err != nil {
			return fmt.Errorf("card requirements not met: %w", err)
		}
	}

	// Validate complete affordability (cost + behavioral resource deductions including choice inputs and payment)
	if err := cm.requirementsValidator.ValidateCardAffordability(ctx, p, card, payment, choiceIndex); err != nil {
		return fmt.Errorf("cannot afford to play card: %w", err)
	}

	log.Debug("✅ Card validation passed", zap.String("card_name", card.Name))
	return nil
}

// PlayCard plays a card (assumes CanPlay validation has already passed)
// payment is the payment method to use for the card cost
// choiceIndex is optional and used when the card has choices between different effects
// cardStorageTarget is optional and used when outputs target "any-card" storage
func (cm *CardManagerImpl) PlayCard(ctx context.Context, game *types.Game, p *player.Player, cardID string, payment *types.CardPayment, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(p.GameID, p.ID)
	log.Debug("🎮 Playing card", zap.String("card_id", cardID))

	// Get card data (we need this for cost and effects)
	card, err := cm.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		return fmt.Errorf("failed to get card data: %w", err)
	}

	if card == nil {
		return fmt.Errorf("card %s not found", cardID)
	}

	// STEP 1: Apply card cost payment (using credits, steel, titanium, and/or payment substitutes)
	if card.Cost > 0 {
		// Get current resources (thread-safe)
		updatedResources, err := p.Resources.Get(ctx)
		if err != nil {
			return fmt.Errorf("failed to get player resources: %w", err)
		}
		updatedResources.Credits -= payment.Credits
		updatedResources.Steel -= payment.Steel
		updatedResources.Titanium -= payment.Titanium

		// Deduct payment substitutes (e.g., heat for Helion)
		if payment.Substitutes != nil {
			for resourceType, amount := range payment.Substitutes {
				switch resourceType {
				case types.ResourceHeat:
					updatedResources.Heat -= amount
				case types.ResourceEnergy:
					updatedResources.Energy -= amount
				case types.ResourcePlants:
					updatedResources.Plants -= amount
				// Add other resource types as needed
				default:
					log.Warn("Unknown payment substitute resource type", zap.String("resource_type", string(resourceType)))
				}
			}
		}

		if err := p.Resources.Update(ctx, updatedResources); err != nil {
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
	err = p.Hand.RemoveCard(ctx, cardID)
	if err != nil {
		return fmt.Errorf("failed to play card: %w", err)
	}
	log.Debug("🃏 Card moved to played cards")

	// STEP 3: Initialize resource storage if the card has storage capability
	if card.ResourceStorage != nil {
		// Initialize the map if it's nil
		if p.ResourceStorage == nil {
			p.ResourceStorage = make(map[string]int)
		}

		// Set the starting amount for this card's resource storage
		p.ResourceStorage[cardID] = card.ResourceStorage.Starting

		// Update the player's resource storage
		if err := p.Resources.UpdateStorage(ctx, p.ResourceStorage); err != nil {
			return fmt.Errorf("failed to initialize card resource storage: %w", err)
		}

		log.Debug("💾 Card resource storage initialized",
			zap.String("resource_type", string(card.ResourceStorage.Type)),
			zap.Int("starting_amount", card.ResourceStorage.Starting))
	}

	// STEP 4: Apply card effects with choice index and card storage target
	if err := cm.effectProcessor.ApplyCardEffects(ctx, game, p, card, choiceIndex, cardStorageTarget); err != nil {
		return fmt.Errorf("failed to apply card effects: %w", err)
	}
	log.Debug("✨ Card effects applied")

	// STEP 5: Subscribe passive effects to event bus
	if cm.effectSubscriber != nil {
		if err := cm.effectSubscriber.SubscribeCardEffects(ctx, p, cardID, card); err != nil {
			return fmt.Errorf("failed to subscribe card effects: %w", err)
		}
		log.Debug("🎆 Passive effects subscribed to event bus")
	}

	log.Info("✅ Card played successfully", zap.String("card_name", card.Name))
	return nil
}
