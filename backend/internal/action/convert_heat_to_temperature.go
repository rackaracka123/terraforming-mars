package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/player"

	"go.uber.org/zap"
)

const (
	// BaseHeatForTemperature is the base cost in heat to raise temperature (before card discounts)
	BaseHeatForTemperature = 8
)

// ConvertHeatToTemperatureAction handles the business logic for converting heat to raise temperature
type ConvertHeatToTemperatureAction struct {
	BaseAction
}

// NewConvertHeatToTemperatureAction creates a new convert heat to temperature action
func NewConvertHeatToTemperatureAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
) *ConvertHeatToTemperatureAction {
	return &ConvertHeatToTemperatureAction{
		BaseAction: NewBaseAction(gameRepo, playerRepo, sessionMgr),
	}
}

// Execute performs the convert heat to temperature action
func (a *ConvertHeatToTemperatureAction) Execute(ctx context.Context, gameID, playerID string) error {
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

	// 3. Validate player exists
	p, err := ValidatePlayer(ctx, a.playerRepo, gameID, playerID, log)
	if err != nil {
		return err
	}

	// 4. Calculate required heat
	// TODO: Implement card discount calculation when player model conversion is available
	requiredHeat := BaseHeatForTemperature

	// 5. Validate player has enough heat
	if p.Resources.Heat < requiredHeat {
		log.Warn("Player cannot afford heat conversion",
			zap.Int("required", requiredHeat),
			zap.Int("available", p.Resources.Heat))
		return fmt.Errorf("insufficient heat: need %d, have %d", requiredHeat, p.Resources.Heat)
	}

	// 6. Deduct heat
	newResources := p.Resources
	newResources.Heat -= requiredHeat
	err = a.playerRepo.UpdateResources(ctx, gameID, playerID, newResources)
	if err != nil {
		log.Error("Failed to deduct heat", zap.Error(err))
		return fmt.Errorf("failed to update resources: %w", err)
	}

	log.Info("🔥 Deducted heat",
		zap.Int("heat_spent", requiredHeat),
		zap.Int("remaining_heat", newResources.Heat))

	// 7. Raise temperature by 1 step (+2°C) if not already maxed
	temperatureRaised := false
	if g.GlobalParameters.Temperature < model.MaxTemperature {
		newTemperature := g.GlobalParameters.Temperature + 2 // Each step is 2°C
		if newTemperature > model.MaxTemperature {
			newTemperature = model.MaxTemperature
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
		// Refresh player data
		p, err = ValidatePlayer(ctx, a.playerRepo, gameID, playerID, log)
		if err != nil {
			return err
		}

		newTR := p.TerraformRating + 1
		err = a.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR)
		if err != nil {
			log.Error("Failed to update terraform rating", zap.Error(err))
			return fmt.Errorf("failed to update terraform rating: %w", err)
		}

		log.Info("🏆 Increased terraform rating",
			zap.Int("old_tr", p.TerraformRating),
			zap.Int("new_tr", newTR))
	}

	// 9. Consume action (only if not unlimited actions)
	// Refresh player data
	p, err = ValidatePlayer(ctx, a.playerRepo, gameID, playerID, log)
	if err != nil {
		return err
	}

	if p.AvailableActions > 0 {
		newActions := p.AvailableActions - 1
		err = a.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, newActions)
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
