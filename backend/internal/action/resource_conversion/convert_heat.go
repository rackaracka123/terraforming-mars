package resource_conversion

import (
	"context"
	"fmt"
	baseaction "terraforming-mars-backend/internal/action"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

const (
	// BaseHeatForTemperature is the base cost in heat to raise temperature (before card discounts)
	BaseHeatForTemperature = 8
)

// ConvertHeatToTemperatureAction handles converting heat to raise temperature
// New architecture: Uses only GameRepository + logger, events handle broadcasting
type ConvertHeatToTemperatureAction struct {
	baseaction.BaseAction
	cardRegistry cards.CardRegistry
}

// NewConvertHeatToTemperatureAction creates a new convert heat action
func NewConvertHeatToTemperatureAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *ConvertHeatToTemperatureAction {
	return &ConvertHeatToTemperatureAction{
		BaseAction:   baseaction.NewBaseAction(gameRepo, nil),
		cardRegistry: cardRegistry,
	}
}

// Execute performs the convert heat to temperature action
func (a *ConvertHeatToTemperatureAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🔥 Converting heat to temperature")

	// 1. Fetch game from repository and validate it's active
	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate it's the player's turn
	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 3. Get player from game
	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	// 5. Calculate required heat (with card discount effects)
	calculator := gamecards.NewRequirementModifierCalculator(a.cardRegistry)
	discounts := calculator.CalculateStandardProjectDiscounts(player, shared.StandardProjectConvertHeatToTemperature)
	heatDiscount := discounts[shared.ResourceHeat]
	requiredHeat := BaseHeatForTemperature - heatDiscount
	if requiredHeat < 1 {
		requiredHeat = 1 // Minimum cost is 1
	}
	log.Debug("💰 Calculated heat cost",
		zap.Int("base_cost", BaseHeatForTemperature),
		zap.Int("discount", heatDiscount),
		zap.Int("final_cost", requiredHeat))

	// 6. Validate player has enough heat
	resources := player.Resources().Get()
	if resources.Heat < requiredHeat {
		log.Warn("Player cannot afford heat conversion",
			zap.Int("required", requiredHeat),
			zap.Int("available", resources.Heat))
		return fmt.Errorf("insufficient heat: need %d, have %d", requiredHeat, resources.Heat)
	}

	// 7. Deduct heat (updates player resources, publishes ResourcesChangedEvent)
	resources.Heat -= requiredHeat
	player.Resources().Set(resources)

	log.Info("🔥 Deducted heat",
		zap.Int("heat_spent", requiredHeat),
		zap.Int("remaining_heat", resources.Heat))

	// 8. Raise temperature using encapsulated method (publishes TemperatureChangedEvent)
	currentTemp := g.GlobalParameters().Temperature()
	if currentTemp < global_parameters.MaxTemperature {
		stepsRaised, err := g.GlobalParameters().IncreaseTemperature(ctx, 1)
		if err != nil {
			log.Error("Failed to raise temperature", zap.Error(err))
			return fmt.Errorf("failed to raise temperature: %w", err)
		}

		if stepsRaised > 0 {
			newTemp := g.GlobalParameters().Temperature()
			log.Info("🌡️ Temperature raised",
				zap.Int("old_temperature", currentTemp),
				zap.Int("new_temperature", newTemp),
				zap.Int("steps_raised", stepsRaised))

			// 9. Award TR if temperature was raised (publishes TerraformRatingChangedEvent)
			oldTR := player.Resources().TerraformRating()
			player.Resources().UpdateTerraformRating(1)
			newTR := player.Resources().TerraformRating()

			log.Info("🏆 Increased terraform rating",
				zap.Int("old_tr", oldTR),
				zap.Int("new_tr", newTR))
		}
	} else {
		log.Info("🌡️ Temperature already at maximum, no TR awarded")
	}

	// 10. Consume action (only if not unlimited actions)
	a.ConsumePlayerAction(g, log)

	log.Info("✅ Heat converted successfully",
		zap.Int("heat_spent", requiredHeat))
	return nil
}
