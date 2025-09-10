package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"
)

// StandardProjects handles standard project actions
type StandardProjects struct {
	standardProjectService service.StandardProjectService
	parser                 *utils.MessageParser
}

// NewStandardProjects creates a new standard projects handler
func NewStandardProjects(standardProjectService service.StandardProjectService, parser *utils.MessageParser) *StandardProjects {
	return &StandardProjects{
		standardProjectService: standardProjectService,
		parser:                 parser,
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
func (sp *StandardProjects) BuildPowerPlant(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionBuildPowerPlantRequest
	if err := sp.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid build power plant request: %w", err)
	}

	// Convert DTO payment to model payment
	payment := &model.Payment{
		Credits:  request.Payment.Credits,
		Steel:    request.Payment.Steel,
		Titanium: request.Payment.Titanium,
	}

	return sp.standardProjectService.BuildPowerPlant(ctx, gameID, playerID, payment)
}

// LaunchAsteroid handles launch asteroid standard project
func (sp *StandardProjects) LaunchAsteroid(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionLaunchAsteroidRequest
	if err := sp.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid launch asteroid request: %w", err)
	}

	// Convert DTO payment to model payment
	payment := &model.Payment{
		Credits:  request.Payment.Credits,
		Steel:    request.Payment.Steel,
		Titanium: request.Payment.Titanium,
	}

	return sp.standardProjectService.LaunchAsteroid(ctx, gameID, playerID, payment)
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

	// Convert DTO payment to model payment
	payment := &model.Payment{
		Credits:  request.Payment.Credits,
		Steel:    request.Payment.Steel,
		Titanium: request.Payment.Titanium,
	}

	return sp.standardProjectService.BuildAquifer(ctx, gameID, playerID, hexPosition, payment)
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

	// Convert DTO payment to model payment
	payment := &model.Payment{
		Credits:  request.Payment.Credits,
		Steel:    request.Payment.Steel,
		Titanium: request.Payment.Titanium,
	}

	return sp.standardProjectService.PlantGreenery(ctx, gameID, playerID, hexPosition, payment)
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

	// Convert DTO payment to model payment
	payment := &model.Payment{
		Credits:  request.Payment.Credits,
		Steel:    request.Payment.Steel,
		Titanium: request.Payment.Titanium,
	}

	return sp.standardProjectService.BuildCity(ctx, gameID, playerID, hexPosition, payment)
}
