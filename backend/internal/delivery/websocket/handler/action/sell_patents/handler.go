package sell_patents

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/service"
)

// Handler handles sell patents standard project action requests
type Handler struct {
	standardProjectService service.StandardProjectService
	parser                 *utils.MessageParser
}

// NewHandler creates a new sell patents handler
func NewHandler(standardProjectService service.StandardProjectService, parser *utils.MessageParser) *Handler {
	return &Handler{
		standardProjectService: standardProjectService,
		parser:                 parser,
	}
}

// Handle processes the sell patents action
func (h *Handler) Handle(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	request, err := h.parseRequest(actionRequest)
	if err != nil {
		return err
	}

	return h.standardProjectService.SellPatents(ctx, gameID, playerID, request.CardCount)
}

// parseRequest parses and validates the action request
func (h *Handler) parseRequest(actionRequest interface{}) (dto.ActionSellPatentsRequest, error) {
	var request dto.ActionSellPatentsRequest
	if err := h.parser.ParsePayload(actionRequest, &request); err != nil {
		return request, fmt.Errorf("invalid sell patents request: %w", err)
	}
	return request, nil
}
