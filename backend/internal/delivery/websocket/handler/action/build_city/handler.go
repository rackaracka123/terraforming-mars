package build_city

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/service"
)

// Handler handles build city standard project action requests
type Handler struct {
	standardProjectService service.StandardProjectService
	parser                 *utils.MessageParser
}

// NewHandler creates a new build city handler
func NewHandler(standardProjectService service.StandardProjectService, parser *utils.MessageParser) *Handler {
	return &Handler{
		standardProjectService: standardProjectService,
		parser:                 parser,
	}
}

// Handle processes the build city action
func (h *Handler) Handle(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	request, err := h.parseRequest(actionRequest)
	if err != nil {
		return err
	}

	hexPosition := dto.ConvertHexPosition(request.HexPosition)
	return h.standardProjectService.BuildCity(ctx, gameID, playerID, hexPosition)
}

// parseRequest parses and validates the action request
func (h *Handler) parseRequest(actionRequest interface{}) (dto.ActionBuildCityRequest, error) {
	var request dto.ActionBuildCityRequest
	if err := h.parser.ParsePayload(actionRequest, &request); err != nil {
		return request, fmt.Errorf("invalid build city request: %w", err)
	}
	return request, nil
}
