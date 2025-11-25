package admin

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// SetGlobalParametersAction handles the admin action to set global parameters
type SetGlobalParametersAction struct {
	action.BaseAction
	gameRepo game.Repository
}

// NewSetGlobalParametersAction creates a new set global parameters admin action
func NewSetGlobalParametersAction(
	gameRepo game.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *SetGlobalParametersAction {
	return &SetGlobalParametersAction{
		BaseAction: action.NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the set global parameters admin action
func (a *SetGlobalParametersAction) Execute(ctx context.Context, gameID string, params types.GlobalParameters) error {
	log := a.InitLogger(gameID, "")
	log.Info("🌍 Admin: Setting global parameters",
		zap.Int("temperature", params.Temperature),
		zap.Int("oxygen", params.Oxygen),
		zap.Int("oceans", params.Oceans))

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Update temperature
	if params.Temperature != 0 {
		err = a.gameRepo.UpdateTemperature(ctx, gameID, params.Temperature)
		if err != nil {
			log.Error("Failed to update temperature", zap.Error(err))
			return err
		}
	}

	// 3. Update oxygen
	if params.Oxygen != 0 {
		err = a.gameRepo.UpdateOxygen(ctx, gameID, params.Oxygen)
		if err != nil {
			log.Error("Failed to update oxygen", zap.Error(err))
			return err
		}
	}

	// 4. Update oceans
	if params.Oceans != 0 {
		err = a.gameRepo.UpdateOceans(ctx, gameID, params.Oceans)
		if err != nil {
			log.Error("Failed to update oceans", zap.Error(err))
			return err
		}
	}

	log.Info("✅ Global parameters updated")

	// 5. Broadcast updated game state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Admin set global parameters completed")
	return nil
}
