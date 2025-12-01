package action

import (
	"context"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// subscribePassiveEffectToEvents subscribes passive effects to relevant domain events
// This function is called when cards with passive effects are played or corporations are selected
func subscribePassiveEffectToEvents(
	ctx context.Context,
	g *game.Game,
	p *player.Player,
	effect player.CardEffect,
	log *zap.Logger,
) {
	// Check each trigger in the effect's behavior
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
	// Subscribe to PlacementBonusGainedEvent
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

		// Condition matched! Apply the effect outputs
		log.Info("🎴 Passive effect triggered",
			zap.String("card_name", effect.CardName),
			zap.String("trigger_type", trigger.Condition.Type),
			zap.Any("resources_gained", event.Resources))

		// Apply outputs (similar to applyOutputs in use_card_action.go)
		for _, output := range effect.Behavior.Outputs {
			switch output.ResourceType {
			// Production resources
			case shared.ResourceSteelProduction:
				p.Resources().AddProduction(map[shared.ResourceType]int{
					shared.ResourceSteelProduction: output.Amount,
				})
				log.Info("🔩 Passive effect added steel production",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceTitaniumProduction:
				p.Resources().AddProduction(map[shared.ResourceType]int{
					shared.ResourceTitaniumProduction: output.Amount,
				})
				log.Info("⚙️ Passive effect added titanium production",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceCreditsProduction:
				p.Resources().AddProduction(map[shared.ResourceType]int{
					shared.ResourceCreditsProduction: output.Amount,
				})
				log.Info("💰 Passive effect added credits production",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourcePlantsProduction:
				p.Resources().AddProduction(map[shared.ResourceType]int{
					shared.ResourcePlantsProduction: output.Amount,
				})
				log.Info("🌱 Passive effect added plants production",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceEnergyProduction:
				p.Resources().AddProduction(map[shared.ResourceType]int{
					shared.ResourceEnergyProduction: output.Amount,
				})
				log.Info("⚡ Passive effect added energy production",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceHeatProduction:
				p.Resources().AddProduction(map[shared.ResourceType]int{
					shared.ResourceHeatProduction: output.Amount,
				})
				log.Info("🔥 Passive effect added heat production",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			// Basic resources
			case shared.ResourceCredits:
				p.Resources().Add(map[shared.ResourceType]int{
					shared.ResourceCredits: output.Amount,
				})
				log.Info("💰 Passive effect added credits",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceSteel:
				p.Resources().Add(map[shared.ResourceType]int{
					shared.ResourceSteel: output.Amount,
				})
				log.Info("🔩 Passive effect added steel",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceTitanium:
				p.Resources().Add(map[shared.ResourceType]int{
					shared.ResourceTitanium: output.Amount,
				})
				log.Info("⚙️ Passive effect added titanium",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourcePlants:
				p.Resources().Add(map[shared.ResourceType]int{
					shared.ResourcePlants: output.Amount,
				})
				log.Info("🌱 Passive effect added plants",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceEnergy:
				p.Resources().Add(map[shared.ResourceType]int{
					shared.ResourceEnergy: output.Amount,
				})
				log.Info("⚡ Passive effect added energy",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceHeat:
				p.Resources().Add(map[shared.ResourceType]int{
					shared.ResourceHeat: output.Amount,
				})
				log.Info("🔥 Passive effect added heat",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			default:
				log.Warn("⚠️ Unhandled output type in passive effect",
					zap.String("type", string(output.ResourceType)),
					zap.String("card", effect.CardName))
			}
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
	// Subscribe to TilePlacedEvent
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

		// Condition matched! Apply the effect outputs
		log.Info("🎴 Passive effect triggered (city placement)",
			zap.String("card_name", effect.CardName),
			zap.String("player_id", p.ID()),
			zap.String("placed_by", event.PlayerID),
			zap.String("tile_type", event.TileType))

		// Apply outputs (same pattern as placement bonus)
		for _, output := range effect.Behavior.Outputs {
			switch output.ResourceType {
			// Production resources
			case shared.ResourceSteelProduction:
				p.Resources().AddProduction(map[shared.ResourceType]int{
					shared.ResourceSteelProduction: output.Amount,
				})
				log.Info("🔩 Passive effect added steel production",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceTitaniumProduction:
				p.Resources().AddProduction(map[shared.ResourceType]int{
					shared.ResourceTitaniumProduction: output.Amount,
				})
				log.Info("⚙️ Passive effect added titanium production",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceCreditsProduction:
				p.Resources().AddProduction(map[shared.ResourceType]int{
					shared.ResourceCreditsProduction: output.Amount,
				})
				log.Info("💰 Passive effect added credits production",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourcePlantsProduction:
				p.Resources().AddProduction(map[shared.ResourceType]int{
					shared.ResourcePlantsProduction: output.Amount,
				})
				log.Info("🌱 Passive effect added plants production",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceEnergyProduction:
				p.Resources().AddProduction(map[shared.ResourceType]int{
					shared.ResourceEnergyProduction: output.Amount,
				})
				log.Info("⚡ Passive effect added energy production",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceHeatProduction:
				p.Resources().AddProduction(map[shared.ResourceType]int{
					shared.ResourceHeatProduction: output.Amount,
				})
				log.Info("🔥 Passive effect added heat production",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			// Basic resources
			case shared.ResourceCredits:
				p.Resources().Add(map[shared.ResourceType]int{
					shared.ResourceCredits: output.Amount,
				})
				log.Info("💰 Passive effect added credits",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceSteel:
				p.Resources().Add(map[shared.ResourceType]int{
					shared.ResourceSteel: output.Amount,
				})
				log.Info("🔩 Passive effect added steel",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceTitanium:
				p.Resources().Add(map[shared.ResourceType]int{
					shared.ResourceTitanium: output.Amount,
				})
				log.Info("⚙️ Passive effect added titanium",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourcePlants:
				p.Resources().Add(map[shared.ResourceType]int{
					shared.ResourcePlants: output.Amount,
				})
				log.Info("🌱 Passive effect added plants",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceEnergy:
				p.Resources().Add(map[shared.ResourceType]int{
					shared.ResourceEnergy: output.Amount,
				})
				log.Info("⚡ Passive effect added energy",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			case shared.ResourceHeat:
				p.Resources().Add(map[shared.ResourceType]int{
					shared.ResourceHeat: output.Amount,
				})
				log.Info("🔥 Passive effect added heat",
					zap.Int("amount", output.Amount),
					zap.String("card", effect.CardName))

			default:
				log.Warn("⚠️ Unhandled output type in passive effect",
					zap.String("type", string(output.ResourceType)),
					zap.String("card", effect.CardName))
			}
		}
	})

	log.Debug("📬 Subscribed passive effect to TilePlacedEvent (city)",
		zap.String("card_name", effect.CardName))
}
