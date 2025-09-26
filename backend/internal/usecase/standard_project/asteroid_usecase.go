package standard_project

import (
	"context"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/internal/usecase/common"

	"go.uber.org/zap"
)

// AsteroidRequest represents the request to launch an asteroid
type AsteroidRequest struct {
	common.ActionRequest
}

// AsteroidResponse represents the response after launching an asteroid
type AsteroidResponse struct {
	common.ActionResponse
	NewTemperature     int `json:"newTemperature" ts:"number"`
	NewTerraformRating int `json:"newTerraformRating" ts:"number"`
}

// AsteroidUseCase handles the asteroid standard project business logic
type AsteroidUseCase struct {
	actionValidator common.ActionValidator
	gameRepo        repository.GameRepository
	playerRepo      repository.PlayerRepository
	gameService     service.GameService
}

// NewAsteroidUseCase creates a new AsteroidUseCase
func NewAsteroidUseCase(
	actionValidator common.ActionValidator,
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	gameService service.GameService,
) common.UseCase[AsteroidRequest, AsteroidResponse] {
	return &AsteroidUseCase{
		actionValidator: actionValidator,
		gameRepo:        gameRepo,
		playerRepo:      playerRepo,
		gameService:     gameService,
	}
}

// Execute performs the asteroid action
func (u *AsteroidUseCase) Execute(ctx context.Context, request AsteroidRequest) (AsteroidResponse, error) {
	log := logger.WithGameContext(request.GameID, request.PlayerID)

	// Define the cost for asteroid (14 M€)
	cost := common.ActionCost{
		Credits: model.StandardProjectCost[model.StandardProjectAsteroid],
	}

	// Validate the action can be performed
	if err := u.actionValidator.ValidatePlayerAction(ctx, request.GameID, request.PlayerID, cost); err != nil {
		log.Warn("Asteroid action validation failed", zap.Error(err))
		return AsteroidResponse{
			ActionResponse: common.ActionResponse{
				Success: false,
				Message: err.Error(),
			},
		}, err
	}

	// Get player to update
	player, err := u.playerRepo.GetByID(ctx, request.GameID, request.PlayerID)
	if err != nil {
		log.Error("Failed to get player for asteroid action", zap.Error(err))
		return AsteroidResponse{
			ActionResponse: common.ActionResponse{
				Success: false,
				Message: "Failed to get player",
			},
		}, err
	}

	// Deduct cost (14 M€)
	updatedResources := player.Resources
	updatedResources.Credits -= cost.Credits

	// Update player resources
	if err := u.playerRepo.UpdateResources(ctx, request.GameID, request.PlayerID, updatedResources); err != nil {
		log.Error("Failed to update player resources", zap.Error(err))
		return AsteroidResponse{
			ActionResponse: common.ActionResponse{
				Success: false,
				Message: "Failed to update resources",
			},
		}, err
	}

	// Get current game state to check temperature
	currentGame, err := u.gameRepo.GetByID(ctx, request.GameID)
	if err != nil {
		log.Error("Failed to get current game state", zap.Error(err))
		return AsteroidResponse{
			ActionResponse: common.ActionResponse{
				Success: false,
				Message: "Failed to get game state",
			},
		}, err
	}

	// Increase temperature by 1 step (2°C)
	if err := u.gameService.IncreaseTemperature(ctx, request.GameID, 1); err != nil {
		log.Error("Failed to increase temperature", zap.Error(err))
		return AsteroidResponse{
			ActionResponse: common.ActionResponse{
				Success: false,
				Message: "Failed to increase temperature",
			},
		}, err
	}

	// Get updated game state to check if temperature actually increased
	updatedGameAfterTemp, err := u.gameRepo.GetByID(ctx, request.GameID)
	if err != nil {
		log.Error("Failed to get updated game state", zap.Error(err))
		return AsteroidResponse{
			ActionResponse: common.ActionResponse{
				Success: false,
				Message: "Failed to get updated game state",
			},
		}, err
	}

	// Only increase terraform rating if temperature actually increased
	newTerraformRating := player.TerraformRating
	if updatedGameAfterTemp.GlobalParameters.Temperature > currentGame.GlobalParameters.Temperature {
		newTerraformRating = player.TerraformRating + 1
		if err := u.playerRepo.UpdateTerraformRating(ctx, request.GameID, request.PlayerID, newTerraformRating); err != nil {
			log.Error("Failed to update terraform rating", zap.Error(err))
			return AsteroidResponse{
				ActionResponse: common.ActionResponse{
					Success: false,
					Message: "Failed to update terraform rating",
				},
			}, err
		}
	}

	// Decrement remaining actions
	if err := u.gameRepo.DecrementRemainingActions(ctx, request.GameID); err != nil {
		log.Error("Failed to decrement remaining actions", zap.Error(err))
		// Note: We don't return error here as the action was successful, just log the issue
	}

	// Use the already fetched updated game state
	updatedGame := updatedGameAfterTemp

	log.Info("🚀 Asteroid launched successfully",
		zap.Int("new_temperature", updatedGame.GlobalParameters.Temperature))

	return AsteroidResponse{
		ActionResponse: common.ActionResponse{
			Success: true,
			Message: "Asteroid launched successfully",
		},
		NewTemperature:     updatedGame.GlobalParameters.Temperature,
		NewTerraformRating: newTerraformRating,
	}, nil
}
