package cards

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// CorporationProcessor handles applying corporation card effects
type CorporationProcessor struct {
	logger *zap.Logger
}

// NewCorporationProcessor creates a new corporation processor
func NewCorporationProcessor(logger *zap.Logger) *CorporationProcessor {
	return &CorporationProcessor{
		logger: logger,
	}
}

// ApplyStartingEffects processes ONLY auto-corporation-start behaviors
// and applies starting resources/production
func (p *CorporationProcessor) ApplyStartingEffects(
	ctx context.Context,
	card *Card,
	pl *player.Player,
	g *game.Game,
) error {
	log := p.logger.With(
		zap.String("corporation_id", card.ID),
		zap.String("corporation_name", card.Name),
		zap.String("player_id", pl.ID()),
	)

	log.Info("💼 Applying corporation starting effects")

	applier := NewBehaviorApplier(pl, g, card.Name, p.logger)

	// Process ONLY behaviors with auto-corporation-start trigger
	for _, behavior := range card.Behaviors {
		for _, trigger := range behavior.Triggers {
			if trigger.Type == string(ResourceTriggerAutoCorporationStart) {
				log.Info("✨ Found auto-corporation-start behavior",
					zap.Int("outputs", len(behavior.Outputs)))

				if err := applier.ApplyOutputs(ctx, behavior.Outputs); err != nil {
					return fmt.Errorf("failed to apply starting effects: %w", err)
				}
			}
		}
	}

	log.Info("✅ Corporation starting effects applied successfully")
	return nil
}

// ApplyAutoEffects processes auto triggers WITHOUT conditions
// (e.g., payment-substitute for Helion)
func (p *CorporationProcessor) ApplyAutoEffects(
	ctx context.Context,
	card *Card,
	pl *player.Player,
	g *game.Game,
) error {
	log := p.logger.With(
		zap.String("corporation_id", card.ID),
		zap.String("corporation_name", card.Name),
		zap.String("player_id", pl.ID()),
	)

	log.Info("💼 Applying corporation auto effects")

	applier := NewBehaviorApplier(pl, g, card.Name, p.logger)

	// Process behaviors with auto trigger WITHOUT conditions
	for _, behavior := range card.Behaviors {
		for _, trigger := range behavior.Triggers {
			// Handle auto trigger WITHOUT conditions (immediate effects like payment-substitute)
			// Auto triggers WITH conditions are passive effects handled separately
			if trigger.Type == string(ResourceTriggerAuto) && trigger.Condition == nil {
				log.Info("✨ Found auto behavior (no condition)",
					zap.Int("outputs", len(behavior.Outputs)))

				if err := applier.ApplyOutputs(ctx, behavior.Outputs); err != nil {
					return fmt.Errorf("failed to apply auto effects: %w", err)
				}
			}
		}
	}

	log.Info("✅ Corporation auto effects applied successfully")
	return nil
}

// SetupForcedFirstAction processes auto-corporation-first-action behaviors and sets forced actions
func (p *CorporationProcessor) SetupForcedFirstAction(
	ctx context.Context,
	card *Card,
	g *game.Game,
	playerID string,
) error {
	log := p.logger.With(
		zap.String("corporation_id", card.ID),
		zap.String("corporation_name", card.Name),
		zap.String("player_id", playerID),
	)

	log.Info("🎯 Checking for forced first action")

	// Process behaviors with auto-corporation-first-action trigger
	for _, behavior := range card.Behaviors {
		for _, trigger := range behavior.Triggers {
			if trigger.Type == string(ResourceTriggerAutoCorporationFirstAction) {
				log.Info("✨ Found auto-corporation-first-action behavior",
					zap.Int("outputs", len(behavior.Outputs)))

				// Create forced action based on outputs
				for _, output := range behavior.Outputs {
					if err := p.createForcedAction(ctx, output, card, g, playerID, log); err != nil {
						return fmt.Errorf("failed to create forced action: %w", err)
					}
				}
			}
		}
	}

	return nil
}

// GetTriggerEffects returns all trigger effects (conditional triggers) from a corporation card
// These are behaviors with auto triggers that have conditions, for event subscription
// This is a READ-ONLY helper that parses the card behaviors and returns CardEffect structs
// The action layer is responsible for adding these effects to the player
func (p *CorporationProcessor) GetTriggerEffects(card *Card) []player.CardEffect {
	var effects []player.CardEffect

	// Iterate through all behaviors and find conditional triggers
	for behaviorIndex, behavior := range card.Behaviors {
		if HasConditionalTrigger(behavior) {
			effect := player.CardEffect{
				CardID:        card.ID,
				CardName:      card.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
			}
			effects = append(effects, effect)
		}
	}

	return effects
}

// GetManualActions returns all manual actions (manual triggers) from a corporation card
// This is a READ-ONLY helper that parses the card behaviors and returns CardAction structs
// The action layer is responsible for adding these actions to the player
func (p *CorporationProcessor) GetManualActions(card *Card) []player.CardAction {
	var actions []player.CardAction

	// Iterate through all behaviors and find manual triggers
	for behaviorIndex, behavior := range card.Behaviors {
		if HasManualTrigger(behavior) {
			action := player.CardAction{
				CardID:        card.ID,
				CardName:      card.Name,
				BehaviorIndex: behaviorIndex,
				Behavior:      behavior,
				PlayCount:     0,
			}
			actions = append(actions, action)
		}
	}

	return actions
}

// createForcedAction creates a forced first action based on the output
func (p *CorporationProcessor) createForcedAction(
	ctx context.Context,
	output shared.ResourceCondition,
	card *Card,
	g *game.Game,
	playerID string,
	log *zap.Logger,
) error {
	switch output.ResourceType {
	case shared.ResourceCityPlacement:
		action := &player.ForcedFirstAction{
			ActionType:    "city-placement",
			CorporationID: card.ID,
			Source:        "corporation-starting-action",
			Completed:     false,
			Description:   fmt.Sprintf("Place a city tile (%s starting action)", card.Name),
		}
		if err := g.SetForcedFirstAction(ctx, playerID, action); err != nil {
			return fmt.Errorf("failed to set forced city placement action: %w", err)
		}
		log.Info("🏙️ Set forced city placement action",
			zap.String("description", action.Description))

		// Create tile placement queue to trigger actual placement UI
		queue := &player.PendingTileSelectionQueue{
			Items:  []string{"city"},
			Source: "corporation-starting-action",
		}
		if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
			return fmt.Errorf("failed to queue tile placement: %w", err)
		}
		log.Info("🎯 Queued city tile for placement")

		// Subscribe to TilePlacedEvent to handle completion and action consumption
		p.subscribeForcedActionCompletion(ctx, g, playerID, "corporation-starting-action", log)

	case shared.ResourceGreeneryPlacement:
		action := &player.ForcedFirstAction{
			ActionType:    "greenery-placement",
			CorporationID: card.ID,
			Source:        "corporation-starting-action",
			Completed:     false,
			Description:   fmt.Sprintf("Place a greenery tile (%s starting action)", card.Name),
		}
		if err := g.SetForcedFirstAction(ctx, playerID, action); err != nil {
			return fmt.Errorf("failed to set forced greenery placement action: %w", err)
		}
		log.Info("🌳 Set forced greenery placement action",
			zap.String("description", action.Description))

		// Create tile placement queue to trigger actual placement UI
		queue := &player.PendingTileSelectionQueue{
			Items:  []string{"greenery"},
			Source: "corporation-starting-action",
		}
		if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
			return fmt.Errorf("failed to queue tile placement: %w", err)
		}
		log.Info("🎯 Queued greenery tile for placement")

		// Subscribe to TilePlacedEvent to handle completion and action consumption
		p.subscribeForcedActionCompletion(ctx, g, playerID, "corporation-starting-action", log)

	case shared.ResourceOceanPlacement:
		action := &player.ForcedFirstAction{
			ActionType:    "ocean-placement",
			CorporationID: card.ID,
			Source:        "corporation-starting-action",
			Completed:     false,
			Description:   fmt.Sprintf("Place an ocean tile (%s starting action)", card.Name),
		}
		if err := g.SetForcedFirstAction(ctx, playerID, action); err != nil {
			return fmt.Errorf("failed to set forced ocean placement action: %w", err)
		}
		log.Info("🌊 Set forced ocean placement action",
			zap.String("description", action.Description))

		// Create tile placement queue to trigger actual placement UI
		queue := &player.PendingTileSelectionQueue{
			Items:  []string{"ocean"},
			Source: "corporation-starting-action",
		}
		if err := g.SetPendingTileSelectionQueue(ctx, playerID, queue); err != nil {
			return fmt.Errorf("failed to queue tile placement: %w", err)
		}
		log.Info("🎯 Queued ocean tile for placement")

		// Subscribe to TilePlacedEvent to handle completion and action consumption
		p.subscribeForcedActionCompletion(ctx, g, playerID, "corporation-starting-action", log)

	default:
		log.Warn("⚠️ Unhandled forced action type",
			zap.String("type", string(output.ResourceType)))
	}

	return nil
}

// subscribeForcedActionCompletion subscribes to TilePlacedEvent to handle forced action completion
// When the last tile in a forced action is placed, this consumes 1 player action and clears the forced action
func (p *CorporationProcessor) subscribeForcedActionCompletion(
	ctx context.Context,
	g *game.Game,
	playerID string,
	source string,
	log *zap.Logger,
) {
	eventBus := g.EventBus()
	if eventBus == nil {
		log.Warn("⚠️ No event bus available, cannot subscribe to forced action completion")
		return
	}

	// Subscribe to TilePlacedEvent
	events.Subscribe(eventBus, func(event events.TilePlacedEvent) {
		// Only handle events for this player
		if event.PlayerID != playerID {
			return
		}

		log.Debug("📡 Received TilePlacedEvent for forced action check",
			zap.String("player_id", event.PlayerID),
			zap.String("tile_type", event.TileType))

		// Check if there's a forced first action for this player
		forcedAction := g.GetForcedFirstAction(playerID)
		if forcedAction == nil {
			log.Debug("No forced first action, ignoring event")
			return
		}

		// Check if the queue is now empty (last tile was placed)
		queue := g.GetPendingTileSelectionQueue(playerID)
		if queue != nil && len(queue.Items) > 0 {
			log.Debug("🔄 Tile queue still has items, waiting for more tiles",
				zap.Int("remaining_tiles", len(queue.Items)))
			return
		}

		// Queue is empty - forced action is complete!
		log.Info("✅ Forced first action completed, consuming player action",
			zap.String("action_type", forcedAction.ActionType),
			zap.String("corporation_id", forcedAction.CorporationID))

		// Consume player action
		currentTurn := g.CurrentTurn()
		if currentTurn != nil && currentTurn.PlayerID() == playerID {
			consumed := currentTurn.ConsumeAction()
			if consumed {
				log.Info("✅ Action consumed for forced first action completion",
					zap.Int("remaining_actions", currentTurn.ActionsRemaining()))

				// Publish GameStateChangedEvent to trigger broadcast
				events.Publish(eventBus, events.GameStateChangedEvent{
					GameID:    g.ID(),
					Timestamp: time.Now(),
				})
			}
		}

		// Clear forced first action
		if err := g.SetForcedFirstAction(ctx, playerID, nil); err != nil {
			log.Error("Failed to clear forced first action", zap.Error(err))
		}
	})

	log.Info("👂 Subscribed to TilePlacedEvent for forced action completion",
		zap.String("player_id", playerID),
		zap.String("source", source))
}
