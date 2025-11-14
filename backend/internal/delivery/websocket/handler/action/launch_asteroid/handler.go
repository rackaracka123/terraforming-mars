package launch_asteroid

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/actions/standard_projects"
)

// Handler handles launch asteroid standard project action requests
type Handler struct {
	launchAsteroidAction *standard_projects.LaunchAsteroidAction
	baseHandler          *utils.StandardProjectHandler
}

// NewHandler creates a new launch asteroid handler
func NewHandler(launchAsteroidAction *standard_projects.LaunchAsteroidAction, parser *utils.MessageParser) *Handler {
	return &Handler{
		launchAsteroidAction: launchAsteroidAction,
		baseHandler:          utils.NewStandardProjectHandler(parser),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	h.baseHandler.HandleStandardProject(ctx, connection, "launch asteroid", "🚀", h.launchAsteroidAction.Execute)
}
