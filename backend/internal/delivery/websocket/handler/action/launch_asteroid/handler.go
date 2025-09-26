package launch_asteroid

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/usecase/common"
	"terraforming-mars-backend/internal/usecase/standard_project"

	"go.uber.org/zap"
)

// Handler handles launch asteroid standard project action requests
type Handler struct {
	asteroidUseCase common.UseCase[standard_project.AsteroidRequest, standard_project.AsteroidResponse]
	errorHandler    *utils.ErrorHandler
	logger          *zap.Logger
}

// NewHandler creates a new launch asteroid handler
func NewHandler(asteroidUseCase common.UseCase[standard_project.AsteroidRequest, standard_project.AsteroidResponse]) *Handler {
	return &Handler{
		asteroidUseCase: asteroidUseCase,
		errorHandler:    utils.NewErrorHandler(),
		logger:          logger.Get(),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	playerID, gameID := connection.GetPlayer()
	if playerID == "" || gameID == "" {
		h.logger.Warn("Launch asteroid action received from unassigned connection",
			zap.String("connection_id", connection.ID))
		h.errorHandler.SendError(connection, utils.ErrMustConnectFirst)
		return
	}

	h.logger.Debug("🚀 Processing launch asteroid action",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID))

	// Create the asteroid request
	request := standard_project.AsteroidRequest{
		ActionRequest: common.ActionRequest{
			GameID:   gameID,
			PlayerID: playerID,
		},
	}

	// Execute the asteroid use case
	response, err := h.asteroidUseCase.Execute(ctx, request)
	if err != nil {
		h.logger.Error("Failed to execute asteroid use case",
			zap.Error(err),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, utils.ErrActionFailed+": "+err.Error())
		return
	}

	if !response.Success {
		h.logger.Warn("Asteroid use case returned failure",
			zap.String("message", response.Message),
			zap.String("player_id", playerID),
			zap.String("game_id", gameID))
		h.errorHandler.SendError(connection, response.Message)
		return
	}

	h.logger.Info("✅ Launch asteroid action completed successfully",
		zap.String("connection_id", connection.ID),
		zap.String("player_id", playerID),
		zap.String("game_id", gameID),
		zap.Int("new_temperature", response.NewTemperature),
		zap.Int("new_terraform_rating", response.NewTerraformRating))
}
