package build_aquifer

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/actions/standard_projects"
)

// Handler handles build aquifer standard project action requests
type Handler struct {
	buildAquiferAction *standard_projects.BuildAquiferAction
	baseHandler        *utils.StandardProjectHandler
}

// NewHandler creates a new build aquifer handler
func NewHandler(buildAquiferAction *standard_projects.BuildAquiferAction, parser *utils.MessageParser) *Handler {
	return &Handler{
		buildAquiferAction: buildAquiferAction,
		baseHandler:        utils.NewStandardProjectHandler(parser),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	h.baseHandler.HandleStandardProject(ctx, connection, "build aquifer", "🌊", h.buildAquiferAction.Execute)
}
