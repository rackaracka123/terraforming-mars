package resources

import (
	"context"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// Repository defines the data access interface for the Resources mechanic
// This abstracts away the underlying player repository implementation
type Repository interface {
	// Resource operations
	GetPlayerResources(ctx context.Context, gameID, playerID string) (Resources, error)
	UpdatePlayerResources(ctx context.Context, gameID, playerID string, resources Resources) error

	// Production operations
	GetPlayerProduction(ctx context.Context, gameID, playerID string) (Production, error)
	UpdatePlayerProduction(ctx context.Context, gameID, playerID string, production Production) error
}

// RepositoryImpl implements the Repository interface
// It wraps the central PlayerRepository but only exposes resources/production operations
type RepositoryImpl struct {
	playerRepo repository.PlayerRepository
}

// NewRepository creates a new resources repository
func NewRepository(playerRepo repository.PlayerRepository) Repository {
	return &RepositoryImpl{
		playerRepo: playerRepo,
	}
}

// GetPlayerResources retrieves player resources
func (r *RepositoryImpl) GetPlayerResources(ctx context.Context, gameID, playerID string) (Resources, error) {
	player, err := r.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return Resources{}, err
	}
	return toResourcesModel(player.Resources), nil
}

// UpdatePlayerResources updates player resources
func (r *RepositoryImpl) UpdatePlayerResources(ctx context.Context, gameID, playerID string, resources Resources) error {
	modelResources := toModelResources(resources)
	return r.playerRepo.UpdateResources(ctx, gameID, playerID, modelResources)
}

// GetPlayerProduction retrieves player production
func (r *RepositoryImpl) GetPlayerProduction(ctx context.Context, gameID, playerID string) (Production, error) {
	player, err := r.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return Production{}, err
	}
	return toProductionModel(player.Production), nil
}

// UpdatePlayerProduction updates player production
func (r *RepositoryImpl) UpdatePlayerProduction(ctx context.Context, gameID, playerID string, production Production) error {
	modelProduction := toModelProduction(production)
	return r.playerRepo.UpdateProduction(ctx, gameID, playerID, modelProduction)
}

// Conversion functions between mechanic types and central model types

func toResourcesModel(mr model.Resources) Resources {
	return Resources{
		Credits:  mr.Credits,
		Steel:    mr.Steel,
		Titanium: mr.Titanium,
		Plants:   mr.Plants,
		Energy:   mr.Energy,
		Heat:     mr.Heat,
	}
}

func toModelResources(r Resources) model.Resources {
	return model.Resources{
		Credits:  r.Credits,
		Steel:    r.Steel,
		Titanium: r.Titanium,
		Plants:   r.Plants,
		Energy:   r.Energy,
		Heat:     r.Heat,
	}
}

func toProductionModel(mp model.Production) Production {
	return Production{
		Credits:  mp.Credits,
		Steel:    mp.Steel,
		Titanium: mp.Titanium,
		Plants:   mp.Plants,
		Energy:   mp.Energy,
		Heat:     mp.Heat,
	}
}

func toModelProduction(p Production) model.Production {
	return model.Production{
		Credits:  p.Credits,
		Steel:    p.Steel,
		Titanium: p.Titanium,
		Plants:   p.Plants,
		Energy:   p.Energy,
		Heat:     p.Heat,
	}
}
