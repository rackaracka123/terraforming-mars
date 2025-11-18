package card

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	gameRepo "terraforming-mars-backend/internal/session/game"
	playerRepo "terraforming-mars-backend/internal/session/game/player"
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
	playerRepo playerRepo.Repository
	gameRepo   gameRepo.Repository
	cardRepo   Repository

	// Track subscription IDs for cleanup
	subscriptions map[string][]events.SubscriptionID // cardID -> list of subscription IDs
}

// NewCardEffectSubscriber creates a new card effect subscriber
func NewCardEffectSubscriber(
	eventBus *events.EventBusImpl,
	playerRepository playerRepo.Repository,
	gameRepository gameRepo.Repository,
	cardRepository Repository,
) CardEffectSubscriber {
	return &CardEffectSubscriberImpl{
		eventBus:      eventBus,
		playerRepo:    playerRepository,
		gameRepo:      gameRepository,
		cardRepo:      cardRepository,
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
	var playerEffects []model.PlayerEffect
	needsInitialTrigger := false // Track if we need to trigger CardHandUpdated after subscription

	for i, behavior := range card.Behaviors {
		if len(behavior.Triggers) == 0 {
			log.Debug("Behavior has no triggers, skipping",
				zap.String("card_name", card.Name),
				zap.Int("behavior_index", i))
			continue
		}

		trigger := behavior.Triggers[0] // Get first trigger

		// Skip non-auto triggers (manual actions and corporation starting bonuses)
		if trigger.Type == model.ResourceTriggerManual {
			log.Debug("Behavior trigger is manual, skipping",
				zap.String("card_name", card.Name))
			continue
		}

		// Skip corporation starting bonuses (not an effect)
		if trigger.Type == model.ResourceTriggerAutoCorporationStart {
			log.Debug("⏭️ Skipping corporation starting bonus (not an effect)",
				zap.String("card_name", card.Name))
			continue
		}

		// Skip corporation forced first actions (not an ongoing effect)
		if trigger.Type == model.ResourceTriggerAutoCorporationFirstAction {
			log.Debug("⏭️ Skipping corporation forced first action (not an effect)",
				zap.String("card_name", card.Name))
			continue
		}

		// Only process auto triggers (immediate effects and event-driven passive effects)
		if trigger.Type != model.ResourceTriggerAuto {
			log.Debug("Behavior trigger type not supported for effect subscription, skipping",
				zap.String("card_name", card.Name),
				zap.String("trigger_type", string(trigger.Type)))
			continue
		}

		// Handle auto triggers with conditions (event-driven passive effects)
		if trigger.Condition != nil {
			// Subscribe based on trigger condition type
			subID, err := ces.subscribeEffectByTriggerType(gameID, playerID, cardID, card.Name, trigger.Condition.Type, behavior)
			if err != nil {
				return fmt.Errorf("failed to subscribe effect for card %s: %w", cardID, err)
			}

			if subID != "" {
				subIDs = append(subIDs, subID)
				// Add effect to player's effects list for frontend display
				playerEffects = append(playerEffects, model.PlayerEffect{
					CardID:        cardID,
					CardName:      card.Name,
					BehaviorIndex: i,
					Behavior:      behavior,
				})
				log.Debug("✅ Event-driven effect subscribed",
					zap.String("card_name", card.Name),
					zap.String("trigger_type", string(trigger.Condition.Type)),
					zap.String("subscription_id", string(subID)))
			}
		} else {
			// Auto trigger without condition - check if it's a static passive effect or immediate effect
			// Only add to player's effects list if it's truly a passive effect (discounts, value modifiers, etc.)
			if isPassiveEffect(behavior) {
				playerEffects = append(playerEffects, model.PlayerEffect{
					CardID:        cardID,
					CardName:      card.Name,
					BehaviorIndex: i,
					Behavior:      behavior,
				})
				log.Debug("✅ Static passive effect registered",
					zap.String("card_name", card.Name),
					zap.Int("behavior_index", i))

				// Auto-subscribe discounts/lenience to CardHandUpdated if they filter cards
				// This ensures modifiers are recalculated when cards are added/removed from hand
				if needsCardHandSubscription(behavior) {
					subID, err := ces.subscribeEffectByTriggerType(
						gameID, playerID, cardID, card.Name,
						model.TriggerCardHandUpdated, // Implicit subscription
						behavior)
					if err != nil {
						return fmt.Errorf("failed to auto-subscribe effect to card hand updates for card %s: %w", cardID, err)
					}

					if subID != "" {
						subIDs = append(subIDs, subID)
						needsInitialTrigger = true // Mark that we need to trigger recalculation for existing cards
						log.Debug("✅ Auto-subscribed static passive effect to card hand updates",
							zap.String("card_name", card.Name),
							zap.String("subscription_id", string(subID)))
					}
				}
			} else {
				log.Debug("⏭️ Skipping immediate effect (not a passive effect)",
					zap.String("card_name", card.Name),
					zap.Int("behavior_index", i))
			}
		}
	}

	// Store subscription IDs for cleanup
	if len(subIDs) > 0 {
		ces.subscriptions[cardID] = subIDs
		log.Info("🎉 Card effects subscribed successfully",
			zap.String("card_name", card.Name),
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

	// If we auto-subscribed discount/lenience effects, trigger immediate recalculation
	// for existing cards in hand by publishing CardHandUpdatedEvent
	if needsInitialTrigger {
		log.Debug("🔄 Publishing CardHandUpdatedEvent to recalculate modifiers for existing cards",
			zap.String("card_name", card.Name))
		events.Publish(ces.eventBus, repository.CardHandUpdatedEvent{
			GameID:   gameID,
			PlayerID: playerID,
		})
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
		// Note: TilePlacedEvent.TileType uses ResourceType constants like "city-tile", not "city"
		subID := events.Subscribe(ces.eventBus, func(event repository.TilePlacedEvent) {
			// Only trigger if it's a city tile and it's this player's game
			if event.GameID == gameID && (event.TileType == string(model.ResourceCityTile) || event.TileType == model.TileTypeCity) {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
			}
		})
		return subID, nil

	case model.TriggerGreeneryPlaced:
		// Subscribe to TilePlacedEvent for greenery tiles
		// Note: TilePlacedEvent.TileType uses ResourceType constants like "greenery-tile", not "greenery"
		subID := events.Subscribe(ces.eventBus, func(event repository.TilePlacedEvent) {
			// Only trigger if it's a greenery tile and it's this player's game
			if event.GameID == gameID && (event.TileType == string(model.ResourceGreeneryTile) || event.TileType == model.TileTypeGreenery) {
				ces.executePassiveEffect(gameID, playerID, cardID, cardName, behavior, event)
			}
		})
		return subID, nil

	case model.TriggerPlacementBonusGained:
		// Subscribe to PlacementBonusGainedEvent for tile placement bonuses
		subID := events.Subscribe(ces.eventBus, func(event repository.PlacementBonusGainedEvent) {
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

	case model.TriggerCardPlayed:
		// Subscribe to CardPlayedEvent for card-played triggers
		subID := events.Subscribe(ces.eventBus, func(event repository.CardPlayedEvent) {
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

	case model.TriggerCardHandUpdated:
		// Subscribe to CardHandUpdatedEvent for requirement modifier recalculation
		subID := events.Subscribe(ces.eventBus, func(event repository.CardHandUpdatedEvent) {
			// Only trigger if it's this player's card hand
			if event.GameID == gameID && event.PlayerID == playerID {
				ctx := context.Background()
				if err := ces.recalculateRequirementModifiers(ctx, gameID, playerID); err != nil {
					log.Error("Failed to recalculate requirement modifiers on card hand update",
						zap.String("card_name", cardName),
						zap.Error(err))
				}
			}
		})
		return subID, nil

	case model.TriggerPlayerEffectsChanged:
		// Subscribe to PlayerEffectsChangedEvent for requirement modifier recalculation
		subID := events.Subscribe(ces.eventBus, func(event repository.PlayerEffectsChangedEvent) {
			// Only trigger if it's this player's effects
			if event.GameID == gameID && event.PlayerID == playerID {
				ctx := context.Background()
				if err := ces.recalculateRequirementModifiers(ctx, gameID, playerID); err != nil {
					log.Error("Failed to recalculate requirement modifiers on effects update",
						zap.String("card_name", cardName),
						zap.Error(err))
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
	behavior model.CardBehavior,
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
	case repository.TilePlacedEvent:
		eventPlayerID = e.PlayerID
	case repository.PlacementBonusGainedEvent:
		eventPlayerID = e.PlayerID
	case repository.CardPlayedEvent:
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
		if output.Target == model.TargetSelfPlayer {
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
	output model.ResourceCondition,
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
	case model.ResourceCredits:
		resources := player.Resources
		resources.Credits += output.Amount
		if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, resources); err != nil {
			return fmt.Errorf("failed to update resources: %w", err)
		}
	case model.ResourceSteel:
		resources := player.Resources
		resources.Steel += output.Amount
		if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, resources); err != nil {
			return fmt.Errorf("failed to update resources: %w", err)
		}
	case model.ResourceTitanium:
		resources := player.Resources
		resources.Titanium += output.Amount
		if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, resources); err != nil {
			return fmt.Errorf("failed to update resources: %w", err)
		}
	case model.ResourcePlants:
		resources := player.Resources
		resources.Plants += output.Amount
		if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, resources); err != nil {
			return fmt.Errorf("failed to update resources: %w", err)
		}
	case model.ResourceEnergy:
		resources := player.Resources
		resources.Energy += output.Amount
		if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, resources); err != nil {
			return fmt.Errorf("failed to update resources: %w", err)
		}
	case model.ResourceHeat:
		resources := player.Resources
		resources.Heat += output.Amount
		if err := ces.playerRepo.UpdateResources(ctx, gameID, playerID, resources); err != nil {
			return fmt.Errorf("failed to update resources: %w", err)
		}

	// Production resources
	case model.ResourceCreditsProduction:
		production := player.Production
		production.Credits += output.Amount
		if err := ces.playerRepo.UpdateProduction(ctx, gameID, playerID, production); err != nil {
			return fmt.Errorf("failed to update production: %w", err)
		}
	case model.ResourceSteelProduction:
		production := player.Production
		production.Steel += output.Amount
		if err := ces.playerRepo.UpdateProduction(ctx, gameID, playerID, production); err != nil {
			return fmt.Errorf("failed to update production: %w", err)
		}
	case model.ResourceTitaniumProduction:
		production := player.Production
		production.Titanium += output.Amount
		if err := ces.playerRepo.UpdateProduction(ctx, gameID, playerID, production); err != nil {
			return fmt.Errorf("failed to update production: %w", err)
		}
	case model.ResourcePlantsProduction:
		production := player.Production
		production.Plants += output.Amount
		if err := ces.playerRepo.UpdateProduction(ctx, gameID, playerID, production); err != nil {
			return fmt.Errorf("failed to update production: %w", err)
		}
	case model.ResourceEnergyProduction:
		production := player.Production
		production.Energy += output.Amount
		if err := ces.playerRepo.UpdateProduction(ctx, gameID, playerID, production); err != nil {
			return fmt.Errorf("failed to update production: %w", err)
		}
	case model.ResourceHeatProduction:
		production := player.Production
		production.Heat += output.Amount
		if err := ces.playerRepo.UpdateProduction(ctx, gameID, playerID, production); err != nil {
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

// recalculateRequirementModifiers recalculates all requirement modifiers for a player
// based on their current card hand and active effects
func (ces *CardEffectSubscriberImpl) recalculateRequirementModifiers(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Get player's current state
	player, err := ces.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	log.Debug("🔄 Recalculating requirement modifiers",
		zap.Int("cards_in_hand", len(player.Cards)),
		zap.Int("active_effects", len(player.Effects)))

	// Map to accumulate modifiers by target
	// Key format: "card:{cardID}" or "standardProject:{projectName}" or "global" for unfiltered
	modifierMap := make(map[string]*model.RequirementModifier)

	// Iterate through player's active effects
	for _, effect := range player.Effects {
		for _, output := range effect.Behavior.Outputs {
			// Only process discount and global-parameter-lenience outputs
			if output.Type != model.ResourceDiscount && output.Type != model.ResourceGlobalParameterLenience {
				continue
			}

			// Determine affected resources for this modifier
			var affectedResources []model.ResourceType
			if len(output.AffectedResources) > 0 {
				// Convert []string to []ResourceType
				affectedResources = make([]model.ResourceType, len(output.AffectedResources))
				for i, res := range output.AffectedResources {
					affectedResources[i] = model.ResourceType(res)
				}
			} else {
				// If no specific resources, use the output type itself
				// For discount, default to credits; for lenience, default to global-parameter
				if output.Type == model.ResourceDiscount {
					affectedResources = []model.ResourceType{model.ResourceCredits}
				} else {
					affectedResources = []model.ResourceType{model.ResourceGlobalParameter}
				}
			}

			// Check if this effect has AffectedStandardProjects
			if len(output.AffectedStandardProjects) > 0 {
				for _, standardProject := range output.AffectedStandardProjects {
					key := fmt.Sprintf("standardProject:%s", standardProject)
					ces.accumulateModifier(modifierMap, key, output.Amount, affectedResources, nil, &standardProject)
				}
				continue
			}

			// Special case: ResourceGlobalParameterLenience should check card requirements
			// and create per-card modifiers (e.g., Inventrix)
			if output.Type == model.ResourceGlobalParameterLenience {
				// Iterate through cards in hand and check for global parameter requirements
				for _, cardID := range player.Cards {
					card, err := ces.cardRepo.GetCardByID(ctx, cardID)
					if err != nil {
						log.Warn("Failed to get card for requirement modifier calculation",
							zap.String("card_id", cardID),
							zap.Error(err))
						continue
					}

					// Check if card has any global parameter requirements (temperature, oxygen, oceans)
					hasGlobalParamReq := false
					for _, req := range card.Requirements {
						if req.Type == model.RequirementTemperature ||
							req.Type == model.RequirementOxygen ||
							req.Type == model.RequirementOceans {
							hasGlobalParamReq = true
							break
						}
					}

					// If card has global parameter requirements, create modifier for it
					if hasGlobalParamReq {
						key := fmt.Sprintf("card:%s", cardID)
						ces.accumulateModifier(modifierMap, key, output.Amount, affectedResources, &cardID, nil)
					}
				}
				continue
			}

			// Check if this effect has AffectedTags or AffectedCardTypes filters
			hasTagFilter := len(output.AffectedTags) > 0
			hasTypeFilter := len(output.AffectedCardTypes) > 0

			if hasTagFilter || hasTypeFilter {
				// This effect applies to specific cards in hand - check each card
				for _, cardID := range player.Cards {
					card, err := ces.cardRepo.GetCardByID(ctx, cardID)
					if err != nil {
						log.Warn("Failed to get card for requirement modifier calculation",
							zap.String("card_id", cardID),
							zap.Error(err))
						continue
					}

					matchesTags := !hasTagFilter   // If no tag filter, matches by default
					matchesTypes := !hasTypeFilter // If no type filter, matches by default

					// Check tag matching
					if hasTagFilter {
						for _, cardTag := range card.Tags {
							for _, affectedTag := range output.AffectedTags {
								if cardTag == affectedTag {
									matchesTags = true
									break
								}
							}
							if matchesTags {
								break
							}
						}
					}

					// Check type matching
					if hasTypeFilter {
						for _, affectedType := range output.AffectedCardTypes {
							if card.Type == string(affectedType) {
								matchesTypes = true
								break
							}
						}
					}

					// If card matches filters, add modifier for this card
					if matchesTags && matchesTypes {
						key := fmt.Sprintf("card:%s", cardID)
						ces.accumulateModifier(modifierMap, key, output.Amount, affectedResources, &cardID, nil)
					}
				}
			} else {
				// No filters and NOT global-parameter-lenience - this is a generic global modifier
				key := "global"
				ces.accumulateModifier(modifierMap, key, output.Amount, affectedResources, nil, nil)
			}
		}
	}

	// Convert map to slice
	modifiers := make([]model.RequirementModifier, 0, len(modifierMap))
	for _, modifier := range modifierMap {
		modifiers = append(modifiers, *modifier)
	}

	// Update player's RequirementModifiers via repository
	if err := ces.playerRepo.UpdateRequirementModifiers(ctx, gameID, playerID, modifiers); err != nil {
		return fmt.Errorf("failed to update requirement modifiers: %w", err)
	}

	log.Info("✨ Requirement modifiers recalculated",
		zap.Int("total_modifiers", len(modifiers)))

	return nil
}

// accumulateModifier adds or merges a modifier into the modifier map
func (ces *CardEffectSubscriberImpl) accumulateModifier(
	modifierMap map[string]*model.RequirementModifier,
	key string,
	amount int,
	affectedResources []model.ResourceType,
	cardTarget *string,
	standardProjectTarget *model.StandardProject,
) {
	existing, exists := modifierMap[key]
	if exists {
		// Merge with existing modifier (accumulate amount)
		existing.Amount += amount
	} else {
		// Create new modifier
		modifierMap[key] = &model.RequirementModifier{
			Amount:                amount,
			AffectedResources:     affectedResources,
			CardTarget:            cardTarget,
			StandardProjectTarget: standardProjectTarget,
		}
	}
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

// isPassiveEffect checks if a behavior contains passive effect outputs (not immediate effects)
// Passive effects include discounts, value modifiers, payment substitutes, etc.
// Immediate effects include resources, production, global parameters, tile placements, etc.
func isPassiveEffect(behavior model.CardBehavior) bool {
	// Check if any output is a passive effect type
	for _, output := range behavior.Outputs {
		switch output.Type {
		// Passive effect types (ongoing modifiers)
		case model.ResourceDiscount,
			model.ResourceValueModifier,
			model.ResourcePaymentSubstitute,
			model.ResourceOceanAdjacencyBonus,
			model.ResourceDefense,
			model.ResourceGlobalParameterLenience:
			return true
		}
	}
	return false
}

// needsCardHandSubscription determines if a static passive effect (without explicit trigger condition)
// should automatically subscribe to CardHandUpdated events.
// This applies to effects that create per-card modifiers based on cards in hand (discounts, lenience).
func needsCardHandSubscription(behavior model.CardBehavior) bool {
	for _, output := range behavior.Outputs {
		// ResourceDiscount with tag/type filters needs to react to card hand changes
		// (e.g., Shuttles giving -2 MC discount for space-tagged cards)
		if output.Type == model.ResourceDiscount &&
			(len(output.AffectedTags) > 0 || len(output.AffectedCardTypes) > 0) {
			return true
		}

		// ResourceGlobalParameterLenience always needs to react to card hand changes
		// (e.g., Inventrix giving +2 lenience for cards with temperature/oxygen/oceans requirements)
		if output.Type == model.ResourceGlobalParameterLenience {
			return true
		}
	}
	return false
}
