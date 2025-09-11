package build_power_plant

import (
	"context"

	"terraforming-mars-backend/internal/service"
)

// Handler handles build power plant standard project action requests
type Handler struct {
	standardProjectService service.StandardProjectService
}

// NewHandler creates a new build power plant handler
func NewHandler(standardProjectService service.StandardProjectService) *Handler {
	return &Handler{
		standardProjectService: standardProjectService,
	}
}

// Handle processes the build power plant action
func (h *Handler) Handle(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	return h.standardProjectService.BuildPowerPlant(ctx, gameID, playerID)
}
