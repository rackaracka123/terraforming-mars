package cards

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// CardEffectSubscriber manages subscriptions for card passive effects to domain events
type CardEffectSubscriber interface {
	// SubscribeCardEffects subscribes all passive effects for a card when it's played
	SubscribeCardEffects(ctx context.Context, gameID, playerID, cardID string, card *model.Card) error

	// UnsubscribeCardEffects unsubscribes all effects for a card (cleanup on card removal)
	UnsubscribeCardEffects(cardID string) error
}

// CardEffectSubscriberImpl implements CardEffectSubscriber
type CardEffectSubscriberImpl struct {
	eventBus   *events.EventBusImpl
	playerRepo repository.PlayerRepository
	gameRepo   repository.GameRepository

	// Track subscription IDs for cleanup
	subscriptions map[string][]events.SubscriptionID // cardID -> list of subscription IDs
}

// NewCardEffectSubscriber creates a new card effect subscriber
func NewCardEffectSubscriber(
	eventBus *events.EventBusImpl,
	playerRepo repository.PlayerRepository,
	gameRepo repository.GameRepository,
) CardEffectSubscriber {
	return &CardEffectSubscriberImpl{
		eventBus:      eventBus,
		playerRepo:    playerRepo,
		gameRepo:      gameRepo,
		subscriptions: make(map[string][]events.SubscriptionID),
	}
}

// SubscribeCardEffects subscribes passive effects based on card behavior
func (ces *CardEffectSubscriberImpl) SubscribeCardEffects(ctx context.Context, gameID, playerID, cardID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)

	// Check if card has passive effects
	if card.Behavior == nil || len(card.Behavior.PassiveEffects) == 0 {
		log.Debug("Card has no passive effects to subscribe",
			zap.String("card_id", cardID),
			zap.String("card_name", card.Name))
		return nil
	}

	log.Info("🎆 Subscribing card passive effects",
		zap.String("card_id", cardID),
		zap.String("card_name", card.Name),
		zap.Int("effect_count", len(card.Behavior.PassiveEffects)))

	// Subscribe each passive effect
	var subIDs []events.SubscriptionID

	for i, effect := range card.Behavior.PassiveEffects {
		if len(effect.Triggers) == 0 {
			log.Debug("Passive effect has no triggers, skipping",
				zap.String("card_name", card.Name),
				zap.Int("effect_index", i))
			continue
		}

		trigger := effect.Triggers[0] // Get first trigger

		// Only subscribe auto-triggers with conditions
		if trigger.Type != model.ResourceTriggerAuto || trigger.Condition == nil {
			log.Debug("Passive effect trigger is not auto or has no condition, skipping",
				zap.String("card_name", card.Name),
				zap.String("trigger_type", string(trigger.Type)))
			continue
		}

		// Subscribe based on trigger condition type
		subID, err := ces.subscribeEffectByTriggerType(gameID, playerID, cardID, card.Name, trigger.Condition.Type, effect)
		if err != nil {
			return fmt.Errorf("failed to subscribe effect for card %s: %w", cardID, err)
		}

		if subID != "" {
			subIDs = append(subIDs, subID)
			log.Debug("✅ Effect subscribed",
				zap.String("card_name", card.Name),
				zap.String("trigger_type", string(trigger.Condition.Type)),
				zap.String("subscription_id", string(subID)))
		}
	}

	// Store subscription IDs for cleanup
	if len(subIDs) > 0 {
		ces.subscriptions[cardID] = subIDs
		log.Info("🎉 Card effects subscribed successfully",
			zap.String("card_name", card.Name),
			zap.Int("subscription_count", len(subIDs)))
	}

	return nil
}

// subscribeEffectByTriggerType subscribes an effect based on its trigger condition type
func (ces *CardEffectSubscriberImpl) subscribeEffectByTriggerType(
	gameID, playerID, cardID, cardName string,
	triggerType model.TriggerType,
	effect model.PassiveEffect,
) (events.SubscriptionID, error) {
	log := logger.WithGameContext(gameID, playerID)

	switch triggerType {
	case model.TriggerTypeTemperatureIncrease:
		// Subscribe to TemperatureChangedEvent
		subID := events.Subscribe(ces.eventBus, func(event repository.TemperatureChangedEvent) {
			// Only trigger if temperature increased and it's this player's game
			if event.GameID == gameID && event.NewValue > event.OldValue {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, effect, event)
			}
		})
		return subID, nil

	case model.TriggerTypeOxygenIncrease:
		// Subscribe to OxygenChangedEvent
		subID := events.Subscribe(ces.eventBus, func(event repository.OxygenChangedEvent) {
			// Only trigger if oxygen increased and it's this player's game
			if event.GameID == gameID && event.NewValue > event.OldValue {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, effect, event)
			}
		})
		return subID, nil

	default:
		log.Debug("Trigger type not yet supported for event subscription",
			zap.String("trigger_type", string(triggerType)),
			zap.String("card_name", cardName))
		return "", nil
	}
}

// executePassiveEffect executes a passive effect when its trigger event fires
func (ces *CardEffectSubscriberImpl) executePassiveEffect(
	gameID, playerID, cardID, cardName string,
	effect model.PassiveEffect,
	event interface{},
) {
	log := logger.WithGameContext(gameID, playerID)

	log.Info("🌟 Passive effect triggered",
		zap.String("card_id", cardID),
		zap.String("card_name", cardName),
		zap.Any("event", event))

	ctx := context.Background()

	// Execute effect outputs
	for _, output := range effect.Outputs {
		if err := ces.applyEffectOutput(ctx, gameID, playerID, cardName, output); err != nil {
			log.Error("Failed to apply passive effect output",
				zap.String("card_name", cardName),
				zap.Error(err))
		}
	}
}

// applyEffectOutput applies a single output from a passive effect
func (ces *CardEffectSubscriberImpl) applyEffectOutput(
	ctx context.Context,
	gameID, playerID, cardName string,
	output model.ResourceOutput,
) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player resources
	player, err := ces.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	resources := player.Resources

	// Apply resource change based on output type
	switch output.Type {
	case model.ResourceTypeCredits:
		resources.Credits += output.Amount
	case model.ResourceTypeSteel:
		resources.Steel += output.Amount
	case model.ResourceTypeTitanium:
		resources.Titanium += output.Amount
	case model.ResourceTypePlants:
		resources.Plants += output.Amount
	case model.ResourceTypeEnergy:
		resources.Energy += output.Amount
	case model.ResourceTypeHeat:
		resources.Heat += output.Amount
	default:
		log.Warn("Unsupported resource type in passive effect",
			zap.String("resource_type", string(output.Type)))
		return nil
	}

	// Update player resources
	if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, resources); err != nil {
		return fmt.Errorf("failed to update resources: %w", err)
	}

	log.Info("✨ Passive effect applied",
		zap.String("card_name", cardName),
		zap.String("resource_type", string(output.Type)),
		zap.Int("amount", output.Amount))

	return nil
}

// UnsubscribeCardEffects unsubscribes all effects for a card
func (ces *CardEffectSubscriberImpl) UnsubscribeCardEffects(cardID string) error {
	subIDs, exists := ces.subscriptions[cardID]
	if !exists {
		return nil // No subscriptions for this card
	}

	log := logger.Get()
	log.Info("🗑️ Unsubscribing card effects",
		zap.String("card_id", cardID),
		zap.Int("subscription_count", len(subIDs)))

	// Unsubscribe all
	for _, subID := range subIDs {
		ces.eventBus.Unsubscribe(subID)
	}

	// Remove from tracking
	delete(ces.subscriptions, cardID)

	return nil
}
