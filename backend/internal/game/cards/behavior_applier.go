package cards

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// BehaviorApplier handles applying card behavior inputs and outputs
// This is the single source of truth for all input/output application
type BehaviorApplier struct {
	player *player.Player // Player affected by the behavior (may be nil for game-only effects)
	game   *game.Game     // Game context for global params/tiles (may be nil for player-only effects)
	source string         // Source identifier for logging (card name, action name, etc.)
	logger *zap.Logger
}

// NewBehaviorApplier creates a new behavior applier
// player and game can be nil if not needed for the specific operations
func NewBehaviorApplier(
	p *player.Player,
	g *game.Game,
	source string,
	logger *zap.Logger,
) *BehaviorApplier {
	return &BehaviorApplier{
		player: p,
		game:   g,
		source: source,
		logger: logger,
	}
}

// ApplyInputs validates player has required resources and deducts them
// Returns error if player is nil or insufficient resources
func (a *BehaviorApplier) ApplyInputs(
	ctx context.Context,
	inputs []shared.ResourceCondition,
) error {
	if len(inputs) == 0 {
		return nil
	}

	if a.player == nil {
		return fmt.Errorf("cannot apply inputs: no player context")
	}

	log := a.logger.With(
		zap.String("source", a.source),
		zap.Int("input_count", len(inputs)),
	)

	log.Debug("💰 Processing behavior inputs")

	// First validate player has all required resources
	resources := a.player.Resources().Get()

	for _, input := range inputs {
		switch input.ResourceType {
		case shared.ResourceCredits:
			if resources.Credits < input.Amount {
				return fmt.Errorf("insufficient credits: need %d, have %d", input.Amount, resources.Credits)
			}
		case shared.ResourceSteel:
			if resources.Steel < input.Amount {
				return fmt.Errorf("insufficient steel: need %d, have %d", input.Amount, resources.Steel)
			}
		case shared.ResourceTitanium:
			if resources.Titanium < input.Amount {
				return fmt.Errorf("insufficient titanium: need %d, have %d", input.Amount, resources.Titanium)
			}
		case shared.ResourcePlants:
			if resources.Plants < input.Amount {
				return fmt.Errorf("insufficient plants: need %d, have %d", input.Amount, resources.Plants)
			}
		case shared.ResourceEnergy:
			if resources.Energy < input.Amount {
				return fmt.Errorf("insufficient energy: need %d, have %d", input.Amount, resources.Energy)
			}
		case shared.ResourceHeat:
			if resources.Heat < input.Amount {
				return fmt.Errorf("insufficient heat: need %d, have %d", input.Amount, resources.Heat)
			}
		default:
			log.Warn("⚠️ Unhandled input type", zap.String("type", string(input.ResourceType)))
		}
	}

	// All validations passed, now deduct resources
	for _, input := range inputs {
		switch input.ResourceType {
		case shared.ResourceCredits:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceCredits: -input.Amount,
			})
			log.Info("💸 Deducted credits", zap.Int("amount", input.Amount))

		case shared.ResourceSteel:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceSteel: -input.Amount,
			})
			log.Info("🔩 Deducted steel", zap.Int("amount", input.Amount))

		case shared.ResourceTitanium:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceTitanium: -input.Amount,
			})
			log.Info("⚙️ Deducted titanium", zap.Int("amount", input.Amount))

		case shared.ResourcePlants:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourcePlants: -input.Amount,
			})
			log.Info("🌱 Deducted plants", zap.Int("amount", input.Amount))

		case shared.ResourceEnergy:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceEnergy: -input.Amount,
			})
			log.Info("⚡ Deducted energy", zap.Int("amount", input.Amount))

		case shared.ResourceHeat:
			a.player.Resources().Add(map[shared.ResourceType]int{
				shared.ResourceHeat: -input.Amount,
			})
			log.Info("🔥 Deducted heat", zap.Int("amount", input.Amount))
		}
	}

	return nil
}

// ApplyOutputs applies resource gains, production changes, global params, tile placements
// Returns error if required context (player/game) is missing for the operation
func (a *BehaviorApplier) ApplyOutputs(
	ctx context.Context,
	outputs []shared.ResourceCondition,
) error {
	if len(outputs) == 0 {
		return nil
	}

	log := a.logger.With(
		zap.String("source", a.source),
		zap.Int("output_count", len(outputs)),
	)

	log.Debug("✨ Processing behavior outputs")

	for _, output := range outputs {
		if err := a.applyOutput(ctx, output, log); err != nil {
			return err
		}
	}

	return nil
}

// applyOutput applies a single output
func (a *BehaviorApplier) applyOutput(
	ctx context.Context,
	output shared.ResourceCondition,
	log *zap.Logger,
) error {
	switch output.ResourceType {
	// ========== Basic Resources ==========
	case shared.ResourceCredits:
		if a.player == nil {
			return fmt.Errorf("cannot apply credits: no player context")
		}
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredits: output.Amount,
		})
		log.Info("💰 Added credits", zap.Int("amount", output.Amount))

	case shared.ResourceSteel:
		if a.player == nil {
			return fmt.Errorf("cannot apply steel: no player context")
		}
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceSteel: output.Amount,
		})
		log.Info("🔩 Added steel", zap.Int("amount", output.Amount))

	case shared.ResourceTitanium:
		if a.player == nil {
			return fmt.Errorf("cannot apply titanium: no player context")
		}
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceTitanium: output.Amount,
		})
		log.Info("⚙️ Added titanium", zap.Int("amount", output.Amount))

	case shared.ResourcePlants:
		if a.player == nil {
			return fmt.Errorf("cannot apply plants: no player context")
		}
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourcePlants: output.Amount,
		})
		log.Info("🌱 Added plants", zap.Int("amount", output.Amount))

	case shared.ResourceEnergy:
		if a.player == nil {
			return fmt.Errorf("cannot apply energy: no player context")
		}
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceEnergy: output.Amount,
		})
		log.Info("⚡ Added energy", zap.Int("amount", output.Amount))

	case shared.ResourceHeat:
		if a.player == nil {
			return fmt.Errorf("cannot apply heat: no player context")
		}
		a.player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceHeat: output.Amount,
		})
		log.Info("🔥 Added heat", zap.Int("amount", output.Amount))

	// ========== Production Resources ==========
	case shared.ResourceCreditsProduction:
		if a.player == nil {
			return fmt.Errorf("cannot apply credits production: no player context")
		}
		a.player.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceCredits: output.Amount,
		})
		log.Info("💰 Added credits production", zap.Int("amount", output.Amount))

	case shared.ResourceSteelProduction:
		if a.player == nil {
			return fmt.Errorf("cannot apply steel production: no player context")
		}
		a.player.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceSteel: output.Amount,
		})
		log.Info("🔩 Added steel production", zap.Int("amount", output.Amount))

	case shared.ResourceTitaniumProduction:
		if a.player == nil {
			return fmt.Errorf("cannot apply titanium production: no player context")
		}
		a.player.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceTitanium: output.Amount,
		})
		log.Info("⚙️ Added titanium production", zap.Int("amount", output.Amount))

	case shared.ResourcePlantsProduction:
		if a.player == nil {
			return fmt.Errorf("cannot apply plants production: no player context")
		}
		a.player.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourcePlants: output.Amount,
		})
		log.Info("🌱 Added plants production", zap.Int("amount", output.Amount))

	case shared.ResourceEnergyProduction:
		if a.player == nil {
			return fmt.Errorf("cannot apply energy production: no player context")
		}
		a.player.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceEnergy: output.Amount,
		})
		log.Info("⚡ Added energy production", zap.Int("amount", output.Amount))

	case shared.ResourceHeatProduction:
		if a.player == nil {
			return fmt.Errorf("cannot apply heat production: no player context")
		}
		a.player.Resources().AddProduction(map[shared.ResourceType]int{
			shared.ResourceHeat: output.Amount,
		})
		log.Info("🔥 Added heat production", zap.Int("amount", output.Amount))

	// ========== Terraform Rating ==========
	case shared.ResourceTR:
		if a.player == nil {
			return fmt.Errorf("cannot apply terraform rating: no player context")
		}
		a.player.Resources().UpdateTerraformRating(output.Amount)
		log.Info("🌍 Added terraform rating", zap.Int("amount", output.Amount))

	// ========== Global Parameters ==========
	case shared.ResourceOxygen:
		if a.game == nil {
			return fmt.Errorf("cannot apply oxygen: no game context")
		}
		actualSteps, err := a.game.GlobalParameters().IncreaseOxygen(ctx, output.Amount)
		if err != nil {
			return fmt.Errorf("failed to increase oxygen: %w", err)
		}
		log.Info("🌊 Increased oxygen", zap.Int("steps", actualSteps))

	case shared.ResourceTemperature:
		if a.game == nil {
			return fmt.Errorf("cannot apply temperature: no game context")
		}
		actualSteps, err := a.game.GlobalParameters().IncreaseTemperature(ctx, output.Amount)
		if err != nil {
			return fmt.Errorf("failed to increase temperature: %w", err)
		}
		log.Info("🌡️ Increased temperature", zap.Int("steps", actualSteps))

	// ========== Tile Placements ==========
	case shared.ResourceCityPlacement, shared.ResourceGreeneryPlacement, shared.ResourceOceanPlacement:
		if a.game == nil {
			return fmt.Errorf("cannot apply tile placement: no game context")
		}
		if a.player == nil {
			return fmt.Errorf("cannot apply tile placement: no player context")
		}

		// Map resource type to tile type string
		var tileType string
		switch output.ResourceType {
		case shared.ResourceCityPlacement:
			tileType = "city"
		case shared.ResourceGreeneryPlacement:
			tileType = "greenery"
		case shared.ResourceOceanPlacement:
			tileType = "ocean"
		}

		// Build array of tile types to append (for multiple placements)
		tileTypes := make([]string, output.Amount)
		for i := 0; i < output.Amount; i++ {
			tileTypes[i] = tileType
		}

		// Atomically append to queue (thread-safe)
		if err := a.game.AppendToPendingTileSelectionQueue(ctx, a.player.ID(), tileTypes, a.source); err != nil {
			return fmt.Errorf("failed to append to pending tile selection queue: %w", err)
		}

		log.Info("🏗️ Added tile placements to queue",
			zap.String("tile_type", tileType),
			zap.Int("count", output.Amount))

	// ========== Payment Substitutes ==========
	case shared.ResourcePaymentSubstitute:
		if a.player == nil {
			return fmt.Errorf("cannot apply payment substitute: no player context")
		}
		// Extract resource type from affectedResources (e.g., "heat" for Helion)
		if len(output.AffectedResources) > 0 {
			resourceType := shared.ResourceType(output.AffectedResources[0])
			a.player.Resources().AddPaymentSubstitute(resourceType, output.Amount)
			log.Info("💰 Added payment substitute",
				zap.String("resource_type", string(resourceType)),
				zap.Int("conversion_rate", output.Amount))
		} else {
			log.Warn("⚠️ payment-substitute output missing affectedResources")
		}

	default:
		log.Warn("⚠️ Unhandled output type",
			zap.String("type", string(output.ResourceType)))
	}

	return nil
}
