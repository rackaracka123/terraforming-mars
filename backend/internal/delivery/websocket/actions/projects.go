package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"
)

// StandardProjects handles standard project actions
type StandardProjects struct {
	standardProjectService service.StandardProjectService
	parser                 *MessageParser
}

// NewStandardProjects creates a new standard projects handler
func NewStandardProjects(standardProjectService service.StandardProjectService) *StandardProjects {
	return &StandardProjects{
		standardProjectService: standardProjectService,
		parser:                 NewMessageParser(),
	}
}

// SellPatents handles sell patents standard project
func (sp *StandardProjects) SellPatents(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionSellPatentsRequest
	if err := sp.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid sell patents request: %w", err)
	}

	return sp.standardProjectService.SellPatents(ctx, gameID, playerID, request.CardCount)
}

// BuildPowerPlant handles build power plant standard project
func (sp *StandardProjects) BuildPowerPlant(ctx context.Context, gameID, playerID string) error {
	return sp.standardProjectService.BuildPowerPlant(ctx, gameID, playerID)
}

// LaunchAsteroid handles launch asteroid standard project
func (sp *StandardProjects) LaunchAsteroid(ctx context.Context, gameID, playerID string) error {
	return sp.standardProjectService.LaunchAsteroid(ctx, gameID, playerID)
}

// BuildAquifer handles build aquifer standard project
func (sp *StandardProjects) BuildAquifer(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionBuildAquiferRequest
	if err := sp.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid build aquifer request: %w", err)
	}

	hexPosition := model.HexPosition{
		Q: request.HexPosition.Q,
		R: request.HexPosition.R,
		S: request.HexPosition.S,
	}

	return sp.standardProjectService.BuildAquifer(ctx, gameID, playerID, hexPosition)
}

// PlantGreenery handles plant greenery standard project
func (sp *StandardProjects) PlantGreenery(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionPlantGreeneryRequest
	if err := sp.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid plant greenery request: %w", err)
	}

	hexPosition := model.HexPosition{
		Q: request.HexPosition.Q,
		R: request.HexPosition.R,
		S: request.HexPosition.S,
	}

	return sp.standardProjectService.PlantGreenery(ctx, gameID, playerID, hexPosition)
}

// BuildCity handles build city standard project
func (sp *StandardProjects) BuildCity(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionBuildCityRequest
	if err := sp.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid build city request: %w", err)
	}

	hexPosition := model.HexPosition{
		Q: request.HexPosition.Q,
		R: request.HexPosition.R,
		S: request.HexPosition.S,
	}

	return sp.standardProjectService.BuildCity(ctx, gameID, playerID, hexPosition)
}
