package build_power_plant

import (
	"context"

	"terraforming-mars-backend/internal/actions/standard_projects"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
)

// Handler handles build power plant standard project action requests
type Handler struct {
	buildPowerPlantAction *standard_projects.BuildPowerPlantAction
	baseHandler           *utils.StandardProjectHandler
}

// NewHandler creates a new build power plant handler
func NewHandler(buildPowerPlantAction *standard_projects.BuildPowerPlantAction) *Handler {
	return &Handler{
		buildPowerPlantAction: buildPowerPlantAction,
		baseHandler:           utils.NewStandardProjectHandler(utils.NewMessageParser()),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	h.baseHandler.HandleStandardProject(ctx, connection, "build power plant", "⚡", h.buildPowerPlantAction.Execute)
}
