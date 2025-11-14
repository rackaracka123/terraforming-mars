package service

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/internal/shared/types"

	"go.uber.org/zap"
)

// CardEffectSubscriber manages subscriptions for card passive effects to domain events
type CardEffectSubscriber interface {
	// SubscribeCardEffects subscribes all passive effects for a card when it's played
	SubscribeCardEffects(ctx context.Context, gameID, playerID, cardID string, card *card.Card) error

	// UnsubscribeCardEffects unsubscribes all effects for a card (cleanup on card removal)
	UnsubscribeCardEffects(cardID string) error
}

// CardEffectSubscriberImpl implements CardEffectSubscriber
type CardEffectSubscriberImpl struct {
	eventBus   *events.EventBusImpl
	playerRepo player.Repository
	gameRepo   game.Repository

	// Track subscription IDs for cleanup
	subscriptions map[string][]events.SubscriptionID // cardID -> list of subscription IDs
}

// NewCardEffectSubscriber creates a new card effect subscriber
func NewCardEffectSubscriber(
	eventBus *events.EventBusImpl,
	playerRepo player.Repository,
	gameRepo game.Repository,
) CardEffectSubscriber {
	return &CardEffectSubscriberImpl{
		eventBus:      eventBus,
		playerRepo:    playerRepo,
		gameRepo:      gameRepo,
		subscriptions: make(map[string][]events.SubscriptionID),
	}
}

// SubscribeCardEffects subscribes passive effects based on card behaviors
func (ces *CardEffectSubscriberImpl) SubscribeCardEffects(ctx context.Context, gameID, playerID, cardID string, c *card.Card) error {
	log := logger.WithGameContext(gameID, playerID)

	// Check if card has any behaviors
	if len(c.Behaviors) == 0 {
		log.Debug("Card has no behaviors to subscribe",
			zap.String("card_id", cardID),
			zap.String("card_name", c.Name))
		return nil
	}

	// Subscribe each auto-triggered behavior
	var subIDs []events.SubscriptionID
	var playerEffects []player.PlayerEffect

	for i, behavior := range c.Behaviors {
		if len(behavior.Triggers) == 0 {
			log.Debug("Behavior has no triggers, skipping",
				zap.String("card_name", c.Name),
				zap.Int("behavior_index", i))
			continue
		}

		trigger := behavior.Triggers[0] // Get first trigger

		// Skip non-auto triggers (manual actions)
		if trigger.Type != card.ResourceTriggerAuto {
			log.Debug("Behavior trigger is not auto, skipping",
				zap.String("card_name", c.Name),
				zap.String("trigger_type", string(trigger.Type)))
			continue
		}

		// Handle auto triggers with conditions (event-driven passive effects)
		if trigger.Condition != nil {
			// Subscribe based on trigger condition type
			subID, err := ces.subscribeEffectByTriggerType(gameID, playerID, cardID, c.Name, trigger.Condition.Type, behavior)
			if err != nil {
				return fmt.Errorf("failed to subscribe effect for card %s: %w", cardID, err)
			}

			if subID != "" {
				subIDs = append(subIDs, subID)
				// Add effect to player's effects list for frontend display
				playerEffects = append(playerEffects, player.PlayerEffect{
					CardID:        cardID,
					CardName:      c.Name,
					BehaviorIndex: i,
					Behavior:      behavior,
				})
				log.Debug("✅ Event-driven effect subscribed",
					zap.String("card_name", c.Name),
					zap.String("trigger_type", string(trigger.Condition.Type)),
					zap.String("subscription_id", string(subID)))
			}
		} else {
			// Auto trigger without condition = static passive effect (discounts, value modifiers, etc.)
			// These don't need event subscriptions but must be in player's Effects array
			playerEffects = append(playerEffects, player.PlayerEffect{
				CardID:        cardID,
				CardName:      c.Name,
				BehaviorIndex: i,
				Behavior:      behavior,
			})
			log.Debug("✅ Static passive effect registered",
				zap.String("card_name", c.Name),
				zap.Int("behavior_index", i))
		}
	}

	// Store subscription IDs for cleanup
	if len(subIDs) > 0 {
		ces.subscriptions[cardID] = subIDs
		log.Info("🎉 Card effects subscribed successfully",
			zap.String("card_name", c.Name),
			zap.Int("subscription_count", len(subIDs)))
	}

	// Update player's effects list for frontend display
	if len(playerEffects) > 0 {
		// Get current player to append new effects
		player, err := ces.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			return fmt.Errorf("failed to get player for effects update: %w", err)
		}

		// Append new effects to existing effects
		updatedEffects := append(player.Effects, playerEffects...)
		err = ces.playerRepo.UpdatePlayerEffects(ctx, gameID, playerID, updatedEffects)
		if err != nil {
			return fmt.Errorf("failed to update player effects: %w", err)
		}

		log.Info("✨ Player effects list updated",
			zap.Int("new_effects_added", len(playerEffects)),
			zap.Int("total_effects", len(updatedEffects)))
	}

	return nil
}

// subscribeEffectByTriggerType subscribes a behavior based on its trigger condition type
func (ces *CardEffectSubscriberImpl) subscribeEffectByTriggerType(
	gameID, playerID, cardID, cardName string,
	triggerType card.TriggerType,
	behavior card.CardBehavior,
) (events.SubscriptionID, error) {
	log := logger.WithGameContext(gameID, playerID)

	switch triggerType {
	case card.TriggerTemperatureRaise:
		// Subscribe to TemperatureChangedEvent
		subID := events.Subscribe(ces.eventBus, func(event parameters.TemperatureChangedEvent) {
			// Only trigger if temperature increased and it's this player's game
			if event.GameID == gameID && event.NewValue > event.OldValue {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
			}
		})
		return subID, nil

	case card.TriggerOxygenRaise:
		// Subscribe to OxygenChangedEvent
		subID := events.Subscribe(ces.eventBus, func(event parameters.OxygenChangedEvent) {
			// Only trigger if oxygen increased and it's this player's game
			if event.GameID == gameID && event.NewValue > event.OldValue {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
			}
		})
		return subID, nil

	case card.TriggerOceanPlaced:
		// Subscribe to OceansChangedEvent (oceans parameter increases when ocean placed)
		subID := events.Subscribe(ces.eventBus, func(event parameters.OceansChangedEvent) {
			// Only trigger if oceans increased and it's this player's game
			if event.GameID == gameID && event.NewValue > event.OldValue {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
			}
		})
		return subID, nil

	case card.TriggerCityPlaced:
		// Subscribe to TilePlacedEvent for city tiles
		// Note: TilePlacedEvent.TileType uses ResourceType constants like "city-tile", not "city"
		subID := events.Subscribe(ces.eventBus, func(event tiles.TilePlacedEvent) {
			// Only trigger if it's a city tile and it's this player's game
			if event.GameID == gameID && (event.TileType == string(types.ResourceCityTile) || event.TileType == tiles.TileTypeCity) {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
			}
		})
		return subID, nil

	case card.TriggerGreeneryPlaced:
		// Subscribe to TilePlacedEvent for greenery tiles
		// Note: TilePlacedEvent.TileType uses ResourceType constants like "greenery-tile", not "greenery"
		subID := events.Subscribe(ces.eventBus, func(event tiles.TilePlacedEvent) {
			// Only trigger if it's a greenery tile and it's this player's game
			if event.GameID == gameID && (event.TileType == string(types.ResourceGreeneryTile) || event.TileType == tiles.TileTypeGreenery) {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
			}
		})
		return subID, nil

	case card.TriggerPlacementBonusGained:
		// Subscribe to PlacementBonusGainedEvent for tile placement bonuses
		subID := events.Subscribe(ces.eventBus, func(event game.PlacementBonusGainedEvent) {
			// Only trigger if it's this player's game
			if event.GameID == gameID {
				// Check if any resource type in the event matches AffectedResources filter
				trigger := behavior.Triggers[0]
				if trigger.Condition != nil && trigger.Condition.AffectedResources != nil && len(trigger.Condition.AffectedResources) > 0 {
					shouldTrigger := false
					// Check if any of the gained resources match the affected resources
					for resourceType := range event.Resources {
						for _, affectedResource := range trigger.Condition.AffectedResources {
							if resourceType == affectedResource {
								shouldTrigger = true
								break
							}
						}
						if shouldTrigger {
							break
						}
					}

					if shouldTrigger {
						ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
					}
				}
			}
		})
		return subID, nil

	case card.TriggerCardPlayed:
		// Subscribe to CardPlayedEvent for card-played triggers
		subID := events.Subscribe(ces.eventBus, func(event player.CardPlayedEvent) {
			// Only trigger if it's this player's game
			if event.GameID == gameID {
				// Get trigger condition from behavior
				trigger := behavior.Triggers[0]

				// Check if card type matches affectedCardTypes filter
				if trigger.Condition != nil && trigger.Condition.AffectedCardTypes != nil && len(trigger.Condition.AffectedCardTypes) > 0 {
					shouldTrigger := false
					for _, affectedType := range trigger.Condition.AffectedCardTypes {
						if string(affectedType) == event.CardType {
							shouldTrigger = true
							break
						}
					}

					if shouldTrigger {
						ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
					}
				} else {
					// No filter, trigger on any card played
					ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
				}
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
	behavior card.CardBehavior,
	event interface{},
) {
	log := logger.WithGameContext(gameID, playerID)

	log.Info("🌟 Passive effect triggered",
		zap.String("card_id", cardID),
		zap.String("card_name", cardName),
		zap.Any("event", event))

	// Extract the player who triggered the event (if applicable)
	var eventPlayerID string
	switch e := event.(type) {
	case tiles.TilePlacedEvent:
		eventPlayerID = e.PlayerID
	case game.PlacementBonusGainedEvent:
		eventPlayerID = e.PlayerID
	case player.CardPlayedEvent:
		eventPlayerID = e.PlayerID
	default:
		// For global events (temperature, oxygen, etc.) the event has no specific player
		eventPlayerID = ""
	}

	ctx := context.Background()

	// Execute behavior outputs, filtering by target
	for _, output := range behavior.Outputs {
		// Check if output target matches the event context
		// TargetSelfPlayer: only trigger if the card owner triggered the event
		// Empty target or TargetAnyPlayer: trigger regardless of who triggered the event
		shouldApply := true
		if output.Target == card.TargetSelfPlayer {
			// Self-targeted output: only apply if card owner triggered the event
			if eventPlayerID != "" && eventPlayerID != playerID {
				log.Debug("Skipping output - target is self-player but event triggered by different player",
					zap.String("card_owner", playerID),
					zap.String("event_player", eventPlayerID),
					zap.String("output_type", string(output.Type)))
				shouldApply = false
			}
		}
		// For empty target or TargetAnyPlayer, always apply (shouldApply stays true)

		if shouldApply {
			if err := ces.applyEffectOutput(ctx, gameID, playerID, cardName, output); err != nil {
				log.Error("Failed to apply passive effect output",
					zap.String("card_name", cardName),
					zap.Error(err))
			}
		}
	}
}

// applyEffectOutput applies a single output from a passive effect
func (ces *CardEffectSubscriberImpl) applyEffectOutput(
	ctx context.Context,
	gameID, playerID, cardName string,
	output card.ResourceCondition,
) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get current player state
	player, err := ces.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Apply resource or production change based on output type
	switch output.Type {
	// Regular resources
	case types.ResourceCredits:
		res, err := player.GetResources()
		if err != nil {
			return fmt.Errorf("failed to get resources: %w", err)
		}
		res.Credits += output.Amount
		if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, res); err != nil {
			return fmt.Errorf("failed to update resources: %w", err)
		}
	case types.ResourceSteel:
		res, err := player.GetResources()
		if err != nil {
			return fmt.Errorf("failed to get resources: %w", err)
		}
		res.Steel += output.Amount
		if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, res); err != nil {
			return fmt.Errorf("failed to update resources: %w", err)
		}
	case types.ResourceTitanium:
		res, err := player.GetResources()
		if err != nil {
			return fmt.Errorf("failed to get resources: %w", err)
		}
		res.Titanium += output.Amount
		if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, res); err != nil {
			return fmt.Errorf("failed to update resources: %w", err)
		}
	case types.ResourcePlants:
		res, err := player.GetResources()
		if err != nil {
			return fmt.Errorf("failed to get resources: %w", err)
		}
		res.Plants += output.Amount
		if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, res); err != nil {
			return fmt.Errorf("failed to update resources: %w", err)
		}
	case types.ResourceEnergy:
		res, err := player.GetResources()
		if err != nil {
			return fmt.Errorf("failed to get resources: %w", err)
		}
		res.Energy += output.Amount
		if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, res); err != nil {
			return fmt.Errorf("failed to update resources: %w", err)
		}
	case types.ResourceHeat:
		res, err := player.GetResources()
		if err != nil {
			return fmt.Errorf("failed to get resources: %w", err)
		}
		res.Heat += output.Amount
		if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, res); err != nil {
			return fmt.Errorf("failed to update resources: %w", err)
		}

	// Production resources
	case types.ResourceCreditsProduction:
		prod, err := player.GetProduction()
		if err != nil {
			return fmt.Errorf("failed to get production: %w", err)
		}
		prod.Credits += output.Amount
		if err := ces.playerRepo.UpdateProduction(ctx, gameID, playerID, prod); err != nil {
			return fmt.Errorf("failed to update production: %w", err)
		}
	case types.ResourceSteelProduction:
		prod, err := player.GetProduction()
		if err != nil {
			return fmt.Errorf("failed to get production: %w", err)
		}
		prod.Steel += output.Amount
		if err := ces.playerRepo.UpdateProduction(ctx, gameID, playerID, prod); err != nil {
			return fmt.Errorf("failed to update production: %w", err)
		}
	case types.ResourceTitaniumProduction:
		prod, err := player.GetProduction()
		if err != nil {
			return fmt.Errorf("failed to get production: %w", err)
		}
		prod.Titanium += output.Amount
		if err := ces.playerRepo.UpdateProduction(ctx, gameID, playerID, prod); err != nil {
			return fmt.Errorf("failed to update production: %w", err)
		}
	case types.ResourcePlantsProduction:
		prod, err := player.GetProduction()
		if err != nil {
			return fmt.Errorf("failed to get production: %w", err)
		}
		prod.Plants += output.Amount
		if err := ces.playerRepo.UpdateProduction(ctx, gameID, playerID, prod); err != nil {
			return fmt.Errorf("failed to update production: %w", err)
		}
	case types.ResourceEnergyProduction:
		prod, err := player.GetProduction()
		if err != nil {
			return fmt.Errorf("failed to get production: %w", err)
		}
		prod.Energy += output.Amount
		if err := ces.playerRepo.UpdateProduction(ctx, gameID, playerID, prod); err != nil {
			return fmt.Errorf("failed to update production: %w", err)
		}
	case types.ResourceHeatProduction:
		prod, err := player.GetProduction()
		if err != nil {
			return fmt.Errorf("failed to get production: %w", err)
		}
		prod.Heat += output.Amount
		if err := ces.playerRepo.UpdateProduction(ctx, gameID, playerID, prod); err != nil {
			return fmt.Errorf("failed to update production: %w", err)
		}

	default:
		log.Warn("Unsupported resource type in passive effect",
			zap.String("resource_type", string(output.Type)))
		return nil
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
