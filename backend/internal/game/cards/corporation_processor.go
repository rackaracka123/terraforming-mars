package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"

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

// ApplyStartingEffects processes auto-corporation-start behaviors and applies starting resources/production
func (p *CorporationProcessor) ApplyStartingEffects(
	ctx context.Context,
	card *Card,
	pl *player.Player,
) error {
	log := p.logger.With(
		zap.String("corporation_id", card.ID),
		zap.String("corporation_name", card.Name),
		zap.String("player_id", pl.ID()),
	)

	log.Info("💼 Applying corporation starting effects")

	// Process behaviors with auto-corporation-start trigger
	for _, behavior := range card.Behaviors {
		for _, trigger := range behavior.Triggers {
			if trigger.Type == string(ResourceTriggerAutoCorporationStart) {
				log.Info("✨ Found auto-corporation-start behavior",
					zap.Int("outputs", len(behavior.Outputs)))

				// Apply all outputs
				for _, output := range behavior.Outputs {
					if err := p.applyOutput(ctx, output, pl, log); err != nil {
						return fmt.Errorf("failed to apply output: %w", err)
					}
				}
			}
		}
	}

	log.Info("✅ Corporation starting effects applied successfully")
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

// GetPassiveEffects returns all passive effects (conditional triggers) from a corporation card
// This is a READ-ONLY helper that parses the card behaviors and returns CardEffect structs
// The action layer is responsible for adding these effects to the player
func (p *CorporationProcessor) GetPassiveEffects(card *Card) []player.CardEffect {
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

// applyOutput applies a single output to the player
func (p *CorporationProcessor) applyOutput(
	ctx context.Context,
	output shared.ResourceCondition,
	pl *player.Player,
	log *zap.Logger,
) error {
	switch output.ResourceType {
	// Basic resources
	case shared.ResourceCredits:
		pl.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredits: output.Amount,
		})
		log.Info("💰 Added credits", zap.Int("amount", output.Amount))

	case shared.ResourceSteel:
		pl.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceSteel: output.Amount,
		})
		log.Info("🔩 Added steel", zap.Int("amount", output.Amount))

	case shared.ResourceTitanium:
		pl.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceTitanium: output.Amount,
		})
		log.Info("⚙️ Added titanium", zap.Int("amount", output.Amount))

	case shared.ResourcePlants:
		pl.Resources().Add(map[shared.ResourceType]int{
			shared.ResourcePlants: output.Amount,
		})
		log.Info("🌱 Added plants", zap.Int("amount", output.Amount))

	case shared.ResourceEnergy:
		pl.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceEnergy: output.Amount,
		})
		log.Info("⚡ Added energy", zap.Int("amount", output.Amount))

	case shared.ResourceHeat:
		pl.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceHeat: output.Amount,
		})
		log.Info("🔥 Added heat", zap.Int("amount", output.Amount))

	// Production resources
	case shared.ResourceCreditsProduction:
		pl.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceCredits: output.Amount,
		})
		log.Info("💰 Added credits production", zap.Int("amount", output.Amount))

	case shared.ResourceSteelProduction:
		pl.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceSteel: output.Amount,
		})
		log.Info("🔩 Added steel production", zap.Int("amount", output.Amount))

	case shared.ResourceTitaniumProduction:
		pl.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceTitanium: output.Amount,
		})
		log.Info("⚙️ Added titanium production", zap.Int("amount", output.Amount))

	case shared.ResourcePlantsProduction:
		pl.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourcePlants: output.Amount,
		})
		log.Info("🌱 Added plants production", zap.Int("amount", output.Amount))

	case shared.ResourceEnergyProduction:
		pl.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceEnergy: output.Amount,
		})
		log.Info("⚡ Added energy production", zap.Int("amount", output.Amount))

	case shared.ResourceHeatProduction:
		pl.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceHeat: output.Amount,
		})
		log.Info("🔥 Added heat production", zap.Int("amount", output.Amount))

	case shared.ResourceTR:
		pl.Resources().UpdateTerraformRating(output.Amount)
		log.Info("🌍 Added terraform rating", zap.Int("amount", output.Amount))

	default:
		log.Warn("⚠️ Unhandled output type in corporation starting effects",
			zap.String("type", string(output.ResourceType)))
	}

	return nil
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

	default:
		log.Warn("⚠️ Unhandled forced action type",
			zap.String("type", string(output.ResourceType)))
	}

	return nil
}
