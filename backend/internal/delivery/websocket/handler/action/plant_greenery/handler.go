package plant_greenery

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/service"
)

// Handler handles plant greenery standard project action requests
type Handler struct {
	standardProjectService service.StandardProjectService
	parser                 *utils.MessageParser
}

// NewHandler creates a new plant greenery handler
func NewHandler(standardProjectService service.StandardProjectService, parser *utils.MessageParser) *Handler {
	return &Handler{
		standardProjectService: standardProjectService,
		parser:                 parser,
	}
}

// Handle processes the plant greenery action
func (h *Handler) Handle(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	request, err := h.parseRequest(actionRequest)
	if err != nil {
		return err
	}

	hexPosition := dto.ConvertHexPosition(request.HexPosition)
	return h.standardProjectService.PlantGreenery(ctx, gameID, playerID, hexPosition)
}

// parseRequest parses and validates the action request
func (h *Handler) parseRequest(actionRequest interface{}) (dto.ActionPlantGreeneryRequest, error) {
	var request dto.ActionPlantGreeneryRequest
	if err := h.parser.ParsePayload(actionRequest, &request); err != nil {
		return request, fmt.Errorf("invalid plant greenery request: %w", err)
	}
	return request, nil
}
