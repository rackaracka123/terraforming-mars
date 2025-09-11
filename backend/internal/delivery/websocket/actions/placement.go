package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"
)

// PlacementActions handles tile and city placement actions
type PlacementActions struct {
	standardProjectService service.StandardProjectService
	parser                 *utils.MessageParser
}

// NewPlacementActions creates a new placement actions handler
func NewPlacementActions(standardProjectService service.StandardProjectService, parser *utils.MessageParser) *PlacementActions {
	return &PlacementActions{
		standardProjectService: standardProjectService,
		parser:                 parser,
	}
}

// BuildAquifer handles build aquifer standard project
func (pa *PlacementActions) BuildAquifer(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionBuildAquiferRequest
	if err := pa.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid build aquifer request: %w", err)
	}

	hexPosition := model.HexPosition{
		Q: request.HexPosition.Q,
		R: request.HexPosition.R,
		S: request.HexPosition.S,
	}

	// Convert DTO payment to model payment
	payment := &model.Payment{
		Credits:  request.Payment.Credits,
		Steel:    request.Payment.Steel,
		Titanium: request.Payment.Titanium,
	}

	return pa.standardProjectService.BuildAquifer(ctx, gameID, playerID, hexPosition, payment)
}

// PlantGreenery handles plant greenery standard project
func (pa *PlacementActions) PlantGreenery(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionPlantGreeneryRequest
	if err := pa.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid plant greenery request: %w", err)
	}

	hexPosition := model.HexPosition{
		Q: request.HexPosition.Q,
		R: request.HexPosition.R,
		S: request.HexPosition.S,
	}

	// Convert DTO payment to model payment
	payment := &model.Payment{
		Credits:  request.Payment.Credits,
		Steel:    request.Payment.Steel,
		Titanium: request.Payment.Titanium,
	}

	return pa.standardProjectService.PlantGreenery(ctx, gameID, playerID, hexPosition, payment)
}

// BuildCity handles build city standard project
func (pa *PlacementActions) BuildCity(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionBuildCityRequest
	if err := pa.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid build city request: %w", err)
	}

	hexPosition := model.HexPosition{
		Q: request.HexPosition.Q,
		R: request.HexPosition.R,
		S: request.HexPosition.S,
	}

	// Convert DTO payment to model payment
	payment := &model.Payment{
		Credits:  request.Payment.Credits,
		Steel:    request.Payment.Steel,
		Titanium: request.Payment.Titanium,
	}

	return pa.standardProjectService.BuildCity(ctx, gameID, playerID, hexPosition, payment)
}
