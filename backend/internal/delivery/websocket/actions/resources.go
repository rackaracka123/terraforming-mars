package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/utils"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/service"
)

// ResourceActions handles resource-related actions (patents, power plants, asteroids)
type ResourceActions struct {
	standardProjectService service.StandardProjectService
	parser                 *utils.MessageParser
}

// NewResourceActions creates a new resource actions handler
func NewResourceActions(standardProjectService service.StandardProjectService, parser *utils.MessageParser) *ResourceActions {
	return &ResourceActions{
		standardProjectService: standardProjectService,
		parser:                 parser,
	}
}

// SellPatents handles sell patents standard project
func (ra *ResourceActions) SellPatents(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionSellPatentsRequest
	if err := ra.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid sell patents request: %w", err)
	}

	return ra.standardProjectService.SellPatents(ctx, gameID, playerID, request.CardCount)
}

// BuildPowerPlant handles build power plant standard project
func (ra *ResourceActions) BuildPowerPlant(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionBuildPowerPlantRequest
	if err := ra.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid build power plant request: %w", err)
	}

	// Convert DTO payment to model payment
	payment := &model.Payment{
		Credits:  request.Payment.Credits,
		Steel:    request.Payment.Steel,
		Titanium: request.Payment.Titanium,
	}

	return ra.standardProjectService.BuildPowerPlant(ctx, gameID, playerID, payment)
}

// LaunchAsteroid handles launch asteroid standard project
func (ra *ResourceActions) LaunchAsteroid(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	var request dto.ActionLaunchAsteroidRequest
	if err := ra.parser.ParsePayload(actionRequest, &request); err != nil {
		return fmt.Errorf("invalid launch asteroid request: %w", err)
	}

	// Convert DTO payment to model payment
	payment := &model.Payment{
		Credits:  request.Payment.Credits,
		Steel:    request.Payment.Steel,
		Titanium: request.Payment.Titanium,
	}

	return ra.standardProjectService.LaunchAsteroid(ctx, gameID, playerID, payment)
}