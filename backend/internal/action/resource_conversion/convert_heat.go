package resource_conversion

import (
	"context"
	"fmt"
	"log/slog"
	baseaction "terraforming-mars-backend/internal/action"

	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/global_parameters"
	"terraforming-mars-backend/internal/game/shared"
)

const (
	// BaseHeatForTemperature is the base cost in heat to raise temperature (before card discounts)
	BaseHeatForTemperature = 8
)

// ConvertHeatToTemperatureAction handles converting heat to raise temperature
// New architecture: Uses only GameRepository + logger, events handle broadcasting
type ConvertHeatToTemperatureAction struct {
	baseaction.BaseAction
	cardRegistry gamecards.CardRegistry
}

// NewConvertHeatToTemperatureAction creates a new convert heat action
func NewConvertHeatToTemperatureAction(
	gameRepo game.GameRepository,
	cardRegistry gamecards.CardRegistry,
	stateRepo game.GameStateRepository,
	logger *slog.Logger,
) *ConvertHeatToTemperatureAction {
	return &ConvertHeatToTemperatureAction{
		BaseAction:   baseaction.NewBaseActionWithStateRepo(gameRepo, nil, stateRepo),
		cardRegistry: cardRegistry,
	}
}

// Execute performs the convert heat to temperature action
func (a *ConvertHeatToTemperatureAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
	storageSubstitutes map[string]int,
) error {
	log := a.InitLogger(gameID, playerID)
	log.Debug("Converting heat to temperature")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	if err := baseaction.ValidateGamePhase(g, shared.GamePhaseAction, log); err != nil {
		return err
	}

	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	if err := baseaction.ValidateActionsRemaining(g, playerID, log); err != nil {
		return err
	}

	if err := baseaction.ValidateNoPendingSelections(g, playerID, log); err != nil {
		return err
	}

	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	calculator := gamecards.NewRequirementModifierCalculator(a.cardRegistry)
	discounts := calculator.CalculateStandardProjectDiscounts(player, shared.StandardProjectConvertHeatToTemperature)
	heatDiscount := discounts[shared.ResourceHeat]
	requiredHeat := BaseHeatForTemperature - heatDiscount
	if requiredHeat < 1 {
		requiredHeat = 1
	}
	log.Debug("Calculated heat cost",
		slog.Int("base_cost", BaseHeatForTemperature),
		slog.Int("discount", heatDiscount),
		slog.Int("final_cost", requiredHeat))

	storageValue, err := ValidateAndDeductStorageSubstitutes(player, storageSubstitutes, shared.ResourceHeat, log)
	if err != nil {
		return fmt.Errorf("storage substitute error: %w", err)
	}

	remainingCost := requiredHeat - storageValue
	if remainingCost < 0 {
		remainingCost = 0
	}

	resources := player.Resources().Get()
	if resources.Heat < remainingCost {
		log.Warn("Player cannot afford heat conversion",
			slog.Int("required", requiredHeat),
			slog.Int("storage_value", storageValue),
			slog.Int("remaining_cost", remainingCost),
			slog.Int("available_heat", resources.Heat))
		return fmt.Errorf("insufficient heat: need %d (after %d from storage), have %d", remainingCost, storageValue, resources.Heat)
	}

	resources.Heat -= remainingCost
	player.Resources().Set(resources)

	log.Debug("Deducted heat",
		slog.Int("heat_spent", remainingCost),
		slog.Int("storage_value", storageValue),
		slog.Int("remaining_heat", resources.Heat))

	var stepsRaised int
	currentTemp := g.GlobalParameters().Temperature()
	if currentTemp < global_parameters.MaxTemperature {
		var err error
		stepsRaised, err = g.GlobalParameters().IncreaseTemperature(ctx, 1, playerID)
		if err != nil {
			log.Error("Failed to raise temperature", slog.Any("error", err))
			return fmt.Errorf("failed to raise temperature: %w", err)
		}

		if stepsRaised > 0 {
			newTemp := g.GlobalParameters().Temperature()
			log.Debug("Temperature raised",
				slog.Int("old_temperature", currentTemp),
				slog.Int("new_temperature", newTemp),
				slog.Int("steps_raised", stepsRaised))

			oldTR := player.Resources().TerraformRating()
			player.Resources().UpdateTerraformRating(1)
			newTR := player.Resources().TerraformRating()

			log.Debug("Increased terraform rating",
				slog.Int("old_tr", oldTR),
				slog.Int("new_tr", newTR))
		}
	} else {
		log.Debug("Temperature already at maximum, no TR awarded")
	}

	a.ConsumePlayerAction(g, log)

	calculatedOutputs := []shared.CalculatedOutput{
		{ResourceType: string(shared.ResourceTemperature), Amount: stepsRaised, IsScaled: false},
	}
	if stepsRaised > 0 {
		calculatedOutputs = append(calculatedOutputs, shared.CalculatedOutput{
			ResourceType: string(shared.ResourceTR), Amount: 1, IsScaled: false,
		})
	}

	g.AddTriggeredEffect(shared.TriggeredEffect{
		CardName:          "Convert Heat",
		PlayerID:          playerID,
		SourceType:        shared.SourceTypeResourceConvert,
		CalculatedOutputs: calculatedOutputs,
	})

	displayData := baseaction.GetStandardProjectDisplayData("Convert Heat")
	a.WriteStateLogFull(ctx, g, "Convert Heat", shared.SourceTypeResourceConvert, playerID, "Converted heat to raise temperature", nil, calculatedOutputs, displayData)

	log.Info("Heat converted",
		slog.Int("heat_spent", requiredHeat))
	return nil
}
