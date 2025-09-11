package launch_asteroid

import (
	"context"

	"terraforming-mars-backend/internal/service"
)

// Handler handles launch asteroid standard project action requests
type Handler struct {
	standardProjectService service.StandardProjectService
}

// NewHandler creates a new launch asteroid handler
func NewHandler(standardProjectService service.StandardProjectService) *Handler {
	return &Handler{
		standardProjectService: standardProjectService,
	}
}

// Handle processes the launch asteroid action
func (h *Handler) Handle(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	return h.standardProjectService.LaunchAsteroid(ctx, gameID, playerID)
}
