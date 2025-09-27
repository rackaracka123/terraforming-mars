package usecase

import (
	"context"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/usecase/common"

	"go.uber.org/zap"
)

// LaunchAsteroid handles the asteroid standard project business logic
func (u *UseCase) LaunchAsteroid(ctx context.Context, gameID, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)

	// Define the cost for asteroid (14 M€)
	cost := common.ActionCost{
		Credits: model.StandardProjectCost[model.StandardProjectAsteroid],
	}

	// Validate the action can be performed
	if err := u.actionValidator.ValidatePlayerAction(ctx, gameID, playerID, cost); err != nil {
		log.Warn("Asteroid action validation failed", zap.Error(err))
		return err
	}

	// Get player to update
	player, err := u.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for asteroid action", zap.Error(err))
		return err
	}

	// Deduct cost (14 M€)
	updatedResources := player.Resources
	updatedResources.Credits -= cost.Credits

	// Update player resources
	if err := u.playerRepo.UpdateResources(ctx, gameID, playerID, updatedResources); err != nil {
		log.Error("Failed to update player resources", zap.Error(err))
		return err
	}

	// Get current game state to check temperature
	currentGame, err := u.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get current game state", zap.Error(err))
		return err
	}

	// Increase temperature by 1 step (2°C)
	if err := u.gameService.IncreaseTemperature(ctx, gameID, 1); err != nil {
		log.Error("Failed to increase temperature", zap.Error(err))
		return err
	}

	// Get updated game state to check if temperature actually increased
	updatedGameAfterTemp, err := u.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get updated game state", zap.Error(err))
		return err
	}

	// Only increase terraform rating if temperature actually increased
	if updatedGameAfterTemp.GlobalParameters.Temperature > currentGame.GlobalParameters.Temperature {
		newTerraformRating := player.TerraformRating + 1
		if err := u.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTerraformRating); err != nil {
			log.Error("Failed to update terraform rating", zap.Error(err))
			return err
		}
	}

	// Decrement remaining actions
	if err := u.gameRepo.DecrementRemainingActions(ctx, gameID); err != nil {
		log.Error("Failed to decrement remaining actions", zap.Error(err))
		// Note: We don't return error here as the action was successful, just log the issue
	}

	// Broadcast game state update to all players
	u.sessionManager.Broadcast(gameID)

	log.Info("🚀 Asteroid launched successfully",
		zap.Int("new_temperature", updatedGameAfterTemp.GlobalParameters.Temperature))

	return nil
}
