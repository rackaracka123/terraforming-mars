package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	gamePackage "terraforming-mars-backend/internal/session/game"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

const (
	// BaseHeatForTemperature is the base cost in heat to raise temperature (before card discounts)
	BaseHeatForTemperature = 8
)

// ConvertHeatToTemperatureAction handles the business logic for converting heat to raise temperature
type ConvertHeatToTemperatureAction struct {
	BaseAction
	gameRepo game.Repository
}

// NewConvertHeatToTemperatureAction creates a new convert heat to temperature action
func NewConvertHeatToTemperatureAction(
	gameRepo game.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *ConvertHeatToTemperatureAction {
	return &ConvertHeatToTemperatureAction{
		BaseAction: NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the convert heat to temperature action
func (a *ConvertHeatToTemperatureAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID)
	log.Info("🔥 Converting heat to temperature")

	// 1. Validate game is active
	g, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate it's the player's turn
	if err := ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 3. Get session and player
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Calculate required heat (with card discount effects)
	requiredHeat := gamePackage.CalculateResourceConversionCost(player, types.StandardProjectConvertHeatToTemperature, BaseHeatForTemperature)
	log.Debug("💰 Calculated heat cost",
		zap.Int("base_cost", BaseHeatForTemperature),
		zap.Int("final_cost", requiredHeat))

	// 5. Validate player has enough heat
	resources := player.Resources().Get()
	if resources.Heat < requiredHeat {
		log.Warn("Player cannot afford heat conversion",
			zap.Int("required", requiredHeat),
			zap.Int("available", resources.Heat))
		return fmt.Errorf("insufficient heat: need %d, have %d", requiredHeat, resources.Heat)
	}

	// 6. Deduct heat
	resources.Heat -= requiredHeat
	player.Resources().Set(resources)

	log.Info("🔥 Deducted heat",
		zap.Int("heat_spent", requiredHeat),
		zap.Int("remaining_heat", resources.Heat))

	// 7. Raise temperature by 1 step (+2°C) if not already maxed
	temperatureRaised := false
	if g.GlobalParameters.Temperature < types.MaxTemperature {
		newTemperature := g.GlobalParameters.Temperature + 2 // Each step is 2°C
		if newTemperature > types.MaxTemperature {
			newTemperature = types.MaxTemperature
		}

		err = a.gameRepo.UpdateTemperature(ctx, gameID, newTemperature)
		if err != nil {
			log.Error("Failed to raise temperature", zap.Error(err))
			return fmt.Errorf("failed to raise temperature: %w", err)
		}

		temperatureRaised = true
		log.Info("🌡️ Temperature raised",
			zap.Int("old_temperature", g.GlobalParameters.Temperature),
			zap.Int("new_temperature", newTemperature))
	} else {
		log.Info("🌡️ Temperature already at maximum, no TR awarded")
	}

	// 8. Award TR if temperature was raised
	if temperatureRaised {
		oldTR := player.Resources().TerraformRating()
		newTR := oldTR + 1
		player.Resources().SetTerraformRating(newTR)

		log.Info("🏆 Increased terraform rating",
			zap.Int("old_tr", oldTR),
			zap.Int("new_tr", newTR))
	}

	// 9. Consume action (only if not unlimited actions)
	availableActions := player.Turn().AvailableActions()
	if availableActions > 0 {
		player.Turn().SetAvailableActions(availableActions - 1)
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions-1))
	}

	// 10. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Heat converted successfully",
		zap.Int("heat_spent", requiredHeat),
		zap.Bool("temperature_raised", temperatureRaised))
	return nil
}
