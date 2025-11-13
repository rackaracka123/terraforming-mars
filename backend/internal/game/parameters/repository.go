package parameters

import (
	"context"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// Repository defines the data access interface for the Parameters mechanic
// This abstracts away the underlying repository implementations
type Repository interface {
	// Global parameters operations
	GetGlobalParameters(ctx context.Context, gameID string) (GlobalParameters, error)
	UpdateGlobalParameters(ctx context.Context, gameID string, params GlobalParameters) error

	// Player terraform rating operations
	GetPlayerTR(ctx context.Context, gameID, playerID string) (int, error)
	UpdatePlayerTR(ctx context.Context, gameID, playerID string, tr int) error
}

// RepositoryImpl implements the Repository interface
// It wraps the central repositories but only exposes parameters operations
type RepositoryImpl struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
}

// NewRepository creates a new parameters repository
func NewRepository(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) Repository {
	return &RepositoryImpl{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
	}
}

// GetGlobalParameters retrieves global parameters
func (r *RepositoryImpl) GetGlobalParameters(ctx context.Context, gameID string) (GlobalParameters, error) {
	game, err := r.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return GlobalParameters{}, err
	}
	return toGlobalParametersModel(game.GlobalParameters), nil
}

// UpdateGlobalParameters updates global parameters
func (r *RepositoryImpl) UpdateGlobalParameters(ctx context.Context, gameID string, params GlobalParameters) error {
	modelParams := toModelGlobalParameters(params)
	return r.gameRepo.UpdateGlobalParameters(ctx, gameID, modelParams)
}

// GetPlayerTR retrieves player terraform rating
func (r *RepositoryImpl) GetPlayerTR(ctx context.Context, gameID, playerID string) (int, error) {
	player, err := r.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return 0, err
	}
	return player.TerraformRating, nil
}

// UpdatePlayerTR updates player terraform rating
func (r *RepositoryImpl) UpdatePlayerTR(ctx context.Context, gameID, playerID string, tr int) error {
	return r.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, tr)
}

// Conversion functions between mechanic types and central model types

func toGlobalParametersModel(mp model.GlobalParameters) GlobalParameters {
	return GlobalParameters{
		Temperature: mp.Temperature,
		Oxygen:      mp.Oxygen,
		Oceans:      mp.Oceans,
	}
}

func toModelGlobalParameters(p GlobalParameters) model.GlobalParameters {
	return model.GlobalParameters{
		Temperature: p.Temperature,
		Oxygen:      p.Oxygen,
		Oceans:      p.Oceans,
	}
}
