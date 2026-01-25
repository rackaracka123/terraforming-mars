package action

import (
	"context"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// SubscribePassiveEffectToEvents subscribes passive effects to relevant domain events
// This function is called when cards with passive effects are played or corporations are selected
func SubscribePassiveEffectToEvents(
	ctx context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	log *zap.Logger,
) {
	for _, trigger := range effect.Behavior.Triggers {
		// Only handle auto triggers with conditions (passive effects)
		if trigger.Type != "auto" || trigger.Condition == nil {
			continue
		}

		// Handle placement-bonus-gained trigger
		if trigger.Condition.Type == "placement-bonus-gained" {
			subscribePlacementBonusEffect(ctx, g, p, effect, trigger, log)
		}

		// Handle city-placed trigger
		if trigger.Condition.Type == "city-placed" {
			subscribeCityPlacedEffect(ctx, g, p, effect, trigger, log)
		}

		// Future: Add more trigger types here (e.g., "temperature-changed", "oxygen-changed")
	}
}

// subscribePlacementBonusEffect subscribes to PlacementBonusGainedEvent
func subscribePlacementBonusEffect(
	ctx context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	trigger shared.Trigger,
	log *zap.Logger,
) {
	events.Subscribe(g.EventBus(), func(event events.PlacementBonusGainedEvent) {
		// Only process if event is for this game and player
		if event.GameID != g.ID() {
			return
		}

		// Check target condition (self-player, any-player, etc.)
		target := "self-player" // Default
		if trigger.Condition.Target != nil {
			target = *trigger.Condition.Target
		}

		if target == "self-player" && event.PlayerID != p.ID() {
			return // Effect only applies to self
		}

		// Check if affected resources match the condition
		if len(trigger.Condition.AffectedResources) > 0 {
			matchFound := false
			for _, affectedResource := range trigger.Condition.AffectedResources {
				if _, exists := event.Resources[affectedResource]; exists {
					matchFound = true
					break
				}
			}
			if !matchFound {
				return // No matching resources in the bonus
			}
		}

		// Condition matched! Apply the effect outputs using BehaviorApplier
		log.Info("🎴 Passive effect triggered",
			zap.String("card_name", effect.CardName),
			zap.String("trigger_type", trigger.Condition.Type),
			zap.Any("resources_gained", event.Resources))

		applier := gamecards.NewBehaviorApplier(p, g, effect.CardName, log).
			WithSourceCardID(effect.CardID)
		if err := applier.ApplyOutputs(context.Background(), effect.Behavior.Outputs); err != nil {
			log.Error("Failed to apply passive effect outputs",
				zap.String("card_name", effect.CardName),
				zap.Error(err))
		}
	})

	log.Debug("📬 Subscribed passive effect to PlacementBonusGainedEvent",
		zap.String("card_name", effect.CardName))
}

// subscribeCityPlacedEffect subscribes to TilePlacedEvent for city placements
func subscribeCityPlacedEffect(
	ctx context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	trigger shared.Trigger,
	log *zap.Logger,
) {
	events.Subscribe(g.EventBus(), func(event events.TilePlacedEvent) {
		// Only process if event is for this game
		if event.GameID != g.ID() {
			return
		}

		// Only process city tile placements
		// TileType is ResourceCityTile constant value: "city-tile"
		if event.TileType != string(shared.ResourceCityTile) {
			return
		}

		// Check target condition (self-player, any-player, etc.)
		target := "self-player" // Default
		if trigger.Condition.Target != nil {
			target = *trigger.Condition.Target
		}

		if target == "self-player" && event.PlayerID != p.ID() {
			return // Effect only applies to self
		}

		// Check location condition
		location := "anywhere" // Default
		if trigger.Condition.Location != nil {
			location = *trigger.Condition.Location
		}

		// For now, we treat all tile placements as "mars" or "anywhere"
		// Future: implement Phobos/colony distinction if needed
		if location != "anywhere" && location != "mars" {
			return // Location doesn't match
		}

		// Condition matched! Apply the effect outputs using BehaviorApplier
		log.Info("🎴 Passive effect triggered (city placement)",
			zap.String("card_name", effect.CardName),
			zap.String("player_id", p.ID()),
			zap.String("placed_by", event.PlayerID),
			zap.String("tile_type", event.TileType))

		applier := gamecards.NewBehaviorApplier(p, g, effect.CardName, log).
			WithSourceCardID(effect.CardID)
		if err := applier.ApplyOutputs(context.Background(), effect.Behavior.Outputs); err != nil {
			log.Error("Failed to apply passive effect outputs",
				zap.String("card_name", effect.CardName),
				zap.Error(err))
		}
	})

	log.Debug("📬 Subscribed passive effect to TilePlacedEvent (city)",
		zap.String("card_name", effect.CardName))
}
