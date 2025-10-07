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

// SubscribeCardEffects subscribes passive effects based on card behaviors
func (ces *CardEffectSubscriberImpl) SubscribeCardEffects(ctx context.Context, gameID, playerID, cardID string, card *model.Card) error {
	log := logger.WithGameContext(gameID, playerID)

	// Check if card has any behaviors
	if len(card.Behaviors) == 0 {
		log.Debug("Card has no behaviors to subscribe",
			zap.String("card_id", cardID),
			zap.String("card_name", card.Name))
		return nil
	}

	// Subscribe each auto-triggered behavior
	var subIDs []events.SubscriptionID

	for i, behavior := range card.Behaviors {
		if len(behavior.Triggers) == 0 {
			log.Debug("Behavior has no triggers, skipping",
				zap.String("card_name", card.Name),
				zap.Int("behavior_index", i))
			continue
		}

		trigger := behavior.Triggers[0] // Get first trigger

		// Only subscribe auto-triggers with conditions
		if trigger.Type != model.ResourceTriggerAuto || trigger.Condition == nil {
			log.Debug("Behavior trigger is not auto or has no condition, skipping",
				zap.String("card_name", card.Name),
				zap.String("trigger_type", string(trigger.Type)))
			continue
		}

		// Subscribe based on trigger condition type
		subID, err := ces.subscribeEffectByTriggerType(gameID, playerID, cardID, card.Name, trigger.Condition.Type, behavior)
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

// subscribeEffectByTriggerType subscribes a behavior based on its trigger condition type
func (ces *CardEffectSubscriberImpl) subscribeEffectByTriggerType(
	gameID, playerID, cardID, cardName string,
	triggerType model.TriggerType,
	behavior model.CardBehavior,
) (events.SubscriptionID, error) {
	log := logger.WithGameContext(gameID, playerID)

	switch triggerType {
	case model.TriggerTemperatureRaise:
		// Subscribe to TemperatureChangedEvent
		subID := events.Subscribe(ces.eventBus, func(event repository.TemperatureChangedEvent) {
			// Only trigger if temperature increased and it's this player's game
			if event.GameID == gameID && event.NewValue > event.OldValue {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
			}
		})
		return subID, nil

	case model.TriggerOxygenRaise:
		// Subscribe to OxygenChangedEvent
		subID := events.Subscribe(ces.eventBus, func(event repository.OxygenChangedEvent) {
			// Only trigger if oxygen increased and it's this player's game
			if event.GameID == gameID && event.NewValue > event.OldValue {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
			}
		})
		return subID, nil

	case model.TriggerOceanPlaced:
		// Subscribe to OceansChangedEvent (oceans parameter increases when ocean placed)
		subID := events.Subscribe(ces.eventBus, func(event repository.OceansChangedEvent) {
			// Only trigger if oceans increased and it's this player's game
			if event.GameID == gameID && event.NewValue > event.OldValue {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
			}
		})
		return subID, nil

	case model.TriggerCityPlaced:
		// Subscribe to TilePlacedEvent for city tiles
		subID := events.Subscribe(ces.eventBus, func(event repository.TilePlacedEvent) {
			// Only trigger if it's a city tile and it's this player's game
			if event.GameID == gameID && event.TileType == "city" {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
			}
		})
		return subID, nil

	case model.TriggerGreeneryPlaced:
		// Subscribe to TilePlacedEvent for greenery tiles
		subID := events.Subscribe(ces.eventBus, func(event repository.TilePlacedEvent) {
			// Only trigger if it's a greenery tile and it's this player's game
			if event.GameID == gameID && event.TileType == "greenery" {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
			}
		})
		return subID, nil

	case model.TriggerTilePlaced:
		// Subscribe to TilePlacedEvent for any tile type
		subID := events.Subscribe(ces.eventBus, func(event repository.TilePlacedEvent) {
			// Trigger for any tile placement in this player's game
			if event.GameID == gameID {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
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

// executePassiveEffect executes a behavior's outputs when its trigger event fires
func (ces *CardEffectSubscriberImpl) executePassiveEffect(
	gameID, playerID, cardID, cardName string,
	behavior model.CardBehavior,
	event interface{},
) {
	log := logger.WithGameContext(gameID, playerID)

	log.Info("🌟 Passive effect triggered",
		zap.String("card_id", cardID),
		zap.String("card_name", cardName),
		zap.Any("event", event))

	ctx := context.Background()

	// Execute behavior outputs
	for _, output := range behavior.Outputs {
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
	output model.ResourceCondition,
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
	case model.ResourceCredits:
		resources.Credits += output.Amount
	case model.ResourceSteel:
		resources.Steel += output.Amount
	case model.ResourceTitanium:
		resources.Titanium += output.Amount
	case model.ResourcePlants:
		resources.Plants += output.Amount
	case model.ResourceEnergy:
		resources.Energy += output.Amount
	case model.ResourceHeat:
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
