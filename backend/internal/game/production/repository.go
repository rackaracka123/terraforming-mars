package production

import (
	"context"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// Repository defines the data access interface for the Production mechanic
type Repository interface {
	// Game operations
	GetGame(ctx context.Context, gameID string) (Game, error)
	UpdateGeneration(ctx context.Context, gameID string, generation int) error
	UpdatePhase(ctx context.Context, gameID string, phase GamePhase) error
	UpdateCurrentTurn(ctx context.Context, gameID string, playerID *string) error
	SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error

	// Player operations
	GetPlayer(ctx context.Context, gameID, playerID string) (Player, error)
	ListPlayers(ctx context.Context, gameID string) ([]Player, error)
	UpdateResources(ctx context.Context, gameID, playerID string, resources Resources) error
	UpdateProductionPhase(ctx context.Context, gameID, playerID string, productionPhase *ProductionPhase) error
	UpdatePassed(ctx context.Context, gameID, playerID string, passed bool) error
	UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error

	// Card deck operations
	PopCard(ctx context.Context, gameID string) (string, error)
}

// RepositoryImpl wraps the central repositories
type RepositoryImpl struct {
	gameRepo     repository.GameRepository
	playerRepo   repository.PlayerRepository
	cardDeckRepo repository.CardDeckRepository
}

func NewRepository(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, cardDeckRepo repository.CardDeckRepository) Repository {
	return &RepositoryImpl{
		gameRepo:     gameRepo,
		playerRepo:   playerRepo,
		cardDeckRepo: cardDeckRepo,
	}
}

func (r *RepositoryImpl) GetGame(ctx context.Context, gameID string) (Game, error) {
	game, err := r.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return Game{}, err
	}
	return toGameModel(game), nil
}

func (r *RepositoryImpl) UpdateGeneration(ctx context.Context, gameID string, generation int) error {
	return r.gameRepo.UpdateGeneration(ctx, gameID, generation)
}

func (r *RepositoryImpl) UpdatePhase(ctx context.Context, gameID string, phase GamePhase) error {
	return r.gameRepo.UpdatePhase(ctx, gameID, model.GamePhase(phase))
}

func (r *RepositoryImpl) UpdateCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	return r.gameRepo.UpdateCurrentTurn(ctx, gameID, playerID)
}

func (r *RepositoryImpl) SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	return r.gameRepo.SetCurrentTurn(ctx, gameID, playerID)
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

func (r *RepositoryImpl) UpdateResources(ctx context.Context, gameID, playerID string, resources Resources) error {
	return r.playerRepo.UpdateResources(ctx, gameID, playerID, toModelResources(resources))
}

func (r *RepositoryImpl) UpdateProductionPhase(ctx context.Context, gameID, playerID string, productionPhase *ProductionPhase) error {
	modelPP := toModelProductionPhase(productionPhase)
	return r.playerRepo.UpdateProductionPhase(ctx, gameID, playerID, modelPP)
}

func (r *RepositoryImpl) UpdatePassed(ctx context.Context, gameID, playerID string, passed bool) error {
	return r.playerRepo.UpdatePassed(ctx, gameID, playerID, passed)
}

func (r *RepositoryImpl) UpdateAvailableActions(ctx context.Context, gameID, playerID string, actions int) error {
	return r.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, actions)
}

func (r *RepositoryImpl) PopCard(ctx context.Context, gameID string) (string, error) {
	return r.cardDeckRepo.Pop(ctx, gameID)
}

// Conversion functions between mechanic types and central model types

func toGameModel(mg model.Game) Game {
	return Game{
		ID:           mg.ID,
		Status:       GameStatus(mg.Status),
		PlayerIDs:    mg.PlayerIDs,
		Generation:   mg.Generation,
		CurrentPhase: GamePhase(mg.CurrentPhase),
	}
}

func toPlayerModel(mp model.Player) Player {
	var productionPhase ProductionPhase
	if mp.ProductionPhase != nil {
		productionPhase = toProductionPhaseModel(*mp.ProductionPhase)
	}
	return Player{
		ID:              mp.ID,
		Resources:       toResourcesModel(mp.Resources),
		Production:      toProductionModel(mp.Production),
		TerraformRating: mp.TerraformRating,
		ProductionPhase: productionPhase,
	}
}

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

func toProductionPhaseModel(mpp model.ProductionPhase) ProductionPhase {
	return ProductionPhase{
		AvailableCards:    mpp.AvailableCards,
		SelectionComplete: mpp.SelectionComplete,
		BeforeResources:   toResourcesModel(mpp.BeforeResources),
		AfterResources:    toResourcesModel(mpp.AfterResources),
		EnergyConverted:   mpp.EnergyConverted,
		CreditsIncome:     mpp.CreditsIncome,
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

func toModelProductionPhase(pp *ProductionPhase) *model.ProductionPhase {
	if pp == nil {
		return nil
	}
	return &model.ProductionPhase{
		AvailableCards:    pp.AvailableCards,
		SelectionComplete: pp.SelectionComplete,
		BeforeResources:   toModelResources(pp.BeforeResources),
		AfterResources:    toModelResources(pp.AfterResources),
		EnergyConverted:   pp.EnergyConverted,
		CreditsIncome:     pp.CreditsIncome,
	}
}
