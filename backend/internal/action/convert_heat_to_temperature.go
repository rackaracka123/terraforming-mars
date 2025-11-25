package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game/card"
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
	requiredHeat := card.CalculateResourceConversionCost(player.Player, types.StandardProjectConvertHeatToTemperature, BaseHeatForTemperature)
	log.Debug("💰 Calculated heat cost",
		zap.Int("base_cost", BaseHeatForTemperature),
		zap.Int("final_cost", requiredHeat))

	// 5. Validate player has enough heat
	currentResources, err := player.Resources.Get(ctx)
	if err != nil {
		log.Error("Failed to get player resources", zap.Error(err))
		return fmt.Errorf("failed to get resources: %w", err)
	}

	if currentResources.Heat < requiredHeat {
		log.Warn("Player cannot afford heat conversion",
			zap.Int("required", requiredHeat),
			zap.Int("available", currentResources.Heat))
		return fmt.Errorf("insufficient heat: need %d, have %d", requiredHeat, currentResources.Heat)
	}

	// 6. Deduct heat
	newResources := currentResources
	newResources.Heat -= requiredHeat
	err = player.Resources.Update(ctx, newResources)
	if err != nil {
		log.Error("Failed to deduct heat", zap.Error(err))
		return fmt.Errorf("failed to update resources: %w", err)
	}

	log.Info("🔥 Deducted heat",
		zap.Int("heat_spent", requiredHeat),
		zap.Int("remaining_heat", newResources.Heat))

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
		newTR := player.TerraformRating + 1
		err = player.Resources.UpdateTerraformRating(ctx, newTR)
		if err != nil {
			log.Error("Failed to update terraform rating", zap.Error(err))
			return fmt.Errorf("failed to update terraform rating: %w", err)
		}

		log.Info("🏆 Increased terraform rating",
			zap.Int("old_tr", player.TerraformRating),
			zap.Int("new_tr", newTR))
	}

	// 9. Consume action (only if not unlimited actions)
	if player.AvailableActions > 0 {
		newActions := player.AvailableActions - 1
		err = player.Action.UpdateAvailableActions(ctx, newActions)
		if err != nil {
			log.Error("Failed to consume action", zap.Error(err))
			return fmt.Errorf("failed to consume action: %w", err)
		}
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", newActions))
	}

	// 10. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Heat converted successfully",
		zap.Int("heat_spent", requiredHeat),
		zap.Bool("temperature_raised", temperatureRaised))
	return nil
}
