package interfaces

import (
	"context"
	"terraforming-mars-backend/internal/model"
)

// PlayerService defines the interface for player operations
type PlayerService interface {
	// Resource operations
	PayResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error
	AddResources(ctx context.Context, gameID, playerID string, resources model.ResourceSet) error
	ValidateResourceCost(ctx context.Context, gameID, playerID string, cost model.ResourceSet) error
	
	// Production operations
	AddProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error
	RemoveProduction(ctx context.Context, gameID, playerID string, production model.ResourceSet) error
	ValidateProductionRequirement(ctx context.Context, gameID, playerID string, requirement model.ResourceSet) error
}