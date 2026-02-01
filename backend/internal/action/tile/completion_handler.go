package tile

import (
	"context"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// Callback types for tile completion
const (
	CallbackConvertPlantsToGreenery = "convert-plants-to-greenery"
	CallbackStandardProjectGreenery = "standard-project-greenery"
	CallbackStandardProjectAquifer  = "standard-project-aquifer"
)

// TileCompletionHandlerFunc is the signature for tile completion callbacks
type TileCompletionHandlerFunc func(ctx context.Context, g *game.Game, playerID string, result *TilePlacementResult, callback *player.TileCompletionCallback) error

// TileCompletionRegistry holds registered completion handlers
type TileCompletionRegistry struct {
	handlers  map[string]TileCompletionHandlerFunc
	stateRepo game.GameStateRepository
}

// NewTileCompletionRegistry creates a new registry with default handlers
func NewTileCompletionRegistry(stateRepo game.GameStateRepository) *TileCompletionRegistry {
	r := &TileCompletionRegistry{
		handlers:  make(map[string]TileCompletionHandlerFunc),
		stateRepo: stateRepo,
	}
	r.registerDefaultHandlers()
	return r
}

func (r *TileCompletionRegistry) registerDefaultHandlers() {
	r.handlers[CallbackConvertPlantsToGreenery] = r.handleConvertPlantsToGreenery
	r.handlers[CallbackStandardProjectGreenery] = r.handleStandardProjectGreenery
	r.handlers[CallbackStandardProjectAquifer] = r.handleStandardProjectAquifer
}

// Handle invokes the appropriate handler for the callback type
// If no callback is registered, no log is created - only use cases that explicitly register callbacks get logs
func (r *TileCompletionRegistry) Handle(ctx context.Context, g *game.Game, playerID string, result *TilePlacementResult, callback *player.TileCompletionCallback) error {
	if callback == nil {
		return nil
	}

	handler, exists := r.handlers[callback.Type]
	if !exists {
		return nil
	}

	return handler(ctx, g, playerID, result, callback)
}

func (r *TileCompletionRegistry) handleConvertPlantsToGreenery(ctx context.Context, g *game.Game, playerID string, result *TilePlacementResult, _ *player.TileCompletionCallback) error {
	outputs := []game.CalculatedOutput{
		{ResourceType: string(shared.ResourceGreeneryPlacement), Amount: 1, IsScaled: false},
	}
	if result.OxygenSteps > 0 {
		outputs = append(outputs, game.CalculatedOutput{ResourceType: string(shared.ResourceOxygen), Amount: result.OxygenSteps, IsScaled: false})
	}
	if result.TRGained > 0 {
		outputs = append(outputs, game.CalculatedOutput{ResourceType: string(shared.ResourceTR), Amount: result.TRGained, IsScaled: false})
	}

	_, err := r.stateRepo.WriteWithChoiceAndOutputs(ctx, g.ID(), g, "Convert Plants", game.SourceTypeResourceConvert, playerID, "Converted plants to greenery", nil, outputs)
	return err
}

func (r *TileCompletionRegistry) handleStandardProjectGreenery(ctx context.Context, g *game.Game, playerID string, result *TilePlacementResult, _ *player.TileCompletionCallback) error {
	outputs := []game.CalculatedOutput{
		{ResourceType: string(shared.ResourceGreeneryPlacement), Amount: 1, IsScaled: false},
	}
	if result.OxygenSteps > 0 {
		outputs = append(outputs, game.CalculatedOutput{ResourceType: string(shared.ResourceOxygen), Amount: result.OxygenSteps, IsScaled: false})
	}
	if result.TRGained > 0 {
		outputs = append(outputs, game.CalculatedOutput{ResourceType: string(shared.ResourceTR), Amount: result.TRGained, IsScaled: false})
	}

	_, err := r.stateRepo.WriteWithChoiceAndOutputs(ctx, g.ID(), g, "Standard Project: Greenery", game.SourceTypeStandardProject, playerID, "Planted greenery", nil, outputs)
	return err
}

func (r *TileCompletionRegistry) handleStandardProjectAquifer(ctx context.Context, g *game.Game, playerID string, result *TilePlacementResult, _ *player.TileCompletionCallback) error {
	outputs := []game.CalculatedOutput{
		{ResourceType: string(shared.ResourceOceanPlacement), Amount: 1, IsScaled: false},
	}
	if result.TRGained > 0 {
		outputs = append(outputs, game.CalculatedOutput{ResourceType: string(shared.ResourceTR), Amount: result.TRGained, IsScaled: false})
	}

	_, err := r.stateRepo.WriteWithChoiceAndOutputs(ctx, g.ID(), g, "Standard Project: Aquifer", game.SourceTypeStandardProject, playerID, "Built aquifer", nil, outputs)
	return err
}
