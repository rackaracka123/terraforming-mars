package turn

import (
	"context"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// Repository defines the data access interface for the Turn mechanic
type Repository interface {
	// Game operations
	GetGame(ctx context.Context, gameID string) (Game, error)
	UpdateCurrentTurn(ctx context.Context, gameID string, playerID *string) error

	// Player operations
	GetPlayer(ctx context.Context, gameID, playerID string) (Player, error)
	ListPlayers(ctx context.Context, gameID string) ([]Player, error)
	UpdatePassed(ctx context.Context, gameID, playerID string, passed bool) error
	UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error
}

// RepositoryImpl wraps the central repositories
type RepositoryImpl struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
}

func NewRepository(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) Repository {
	return &RepositoryImpl{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
	}
}

func (r *RepositoryImpl) GetGame(ctx context.Context, gameID string) (Game, error) {
	game, err := r.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return Game{}, err
	}
	return toGameModel(game), nil
}

func (r *RepositoryImpl) UpdateCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	return r.gameRepo.UpdateCurrentTurn(ctx, gameID, playerID)
}

func (r *RepositoryImpl) GetPlayer(ctx context.Context, gameID, playerID string) (Player, error) {
	player, err := r.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return Player{}, err
	}
	return toPlayerModel(player), nil
}

func (r *RepositoryImpl) ListPlayers(ctx context.Context, gameID string) ([]Player, error) {
	players, err := r.playerRepo.ListByGameID(ctx, gameID)
	if err != nil {
		return nil, err
	}
	result := make([]Player, len(players))
	for i, p := range players {
		result[i] = toPlayerModel(p)
	}
	return result, nil
}

func (r *RepositoryImpl) UpdatePassed(ctx context.Context, gameID, playerID string, passed bool) error {
	return r.playerRepo.UpdatePassed(ctx, gameID, playerID, passed)
}

func (r *RepositoryImpl) UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error {
	return r.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, actions)
}

// Conversion functions between mechanic types and central model types

func toGameModel(mg model.Game) Game {
	return Game{
		CurrentTurn: mg.CurrentTurn,
		Status:      GameStatus(mg.Status),
		PlayerIDs:   mg.PlayerIDs,
	}
}

func toPlayerModel(mp model.Player) Player {
	return Player{
		ID:               mp.ID,
		Passed:           mp.Passed,
		AvailableActions: mp.AvailableActions,
	}
}
