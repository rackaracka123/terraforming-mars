package plant_greenery

import (
	"context"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/service"
)

// Handler handles plant greenery standard project action requests
type Handler struct {
	standardProjectService service.StandardProjectService
	baseHandler            *utils.StandardProjectHandler
}

// NewHandler creates a new plant greenery handler
func NewHandler(standardProjectService service.StandardProjectService, parser *utils.MessageParser) *Handler {
	return &Handler{
		standardProjectService: standardProjectService,
		baseHandler:            utils.NewStandardProjectHandler(parser),
	}
}

// HandleMessage implements the MessageHandler interface
func (h *Handler) HandleMessage(ctx context.Context, connection *core.Connection, message dto.WebSocketMessage) {
	h.baseHandler.HandleStandardProject(ctx, connection, "plant greenery", "🌱", h.standardProjectService.PlantGreenery)
}
