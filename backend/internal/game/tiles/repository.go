package tiles

import (
	"context"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// Repository defines the data access interface for the Tiles mechanic
type Repository interface {
	// Game operations
	GetGame(ctx context.Context, gameID string) (Game, error)
	UpdateTileOccupancy(ctx context.Context, gameID string, coordinate HexPosition, occupant *TileOccupant, ownerID *string) error

	// Player operations
	GetPlayer(ctx context.Context, gameID, playerID string) (Player, error)
	UpdateResources(ctx context.Context, gameID, playerID string, resources Resources) error
	ProcessNextTileInQueue(ctx context.Context, gameID, playerID string) (string, error)
	GetPendingTileSelectionQueue(ctx context.Context, gameID, playerID string) (*PendingTileSelectionQueue, error)
	UpdatePendingTileSelection(ctx context.Context, gameID, playerID string, selection *PendingTileSelection) error
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

func (r *RepositoryImpl) UpdateTileOccupancy(ctx context.Context, gameID string, coordinate HexPosition, occupant *TileOccupant, ownerID *string) error {
	modelCoord := toModelHexPosition(coordinate)
	modelOccupant := toModelTileOccupant(occupant)
	return r.gameRepo.UpdateTileOccupancy(ctx, gameID, modelCoord, modelOccupant, ownerID)
}

func (r *RepositoryImpl) GetPlayer(ctx context.Context, gameID, playerID string) (Player, error) {
	player, err := r.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return Player{}, err
	}
	return toPlayerModel(player), nil
}

func (r *RepositoryImpl) UpdateResources(ctx context.Context, gameID, playerID string, resources Resources) error {
	return r.playerRepo.UpdateResources(ctx, gameID, playerID, toModelResources(resources))
}

func (r *RepositoryImpl) ProcessNextTileInQueue(ctx context.Context, gameID, playerID string) (string, error) {
	return r.playerRepo.ProcessNextTileInQueue(ctx, gameID, playerID)
}

func (r *RepositoryImpl) GetPendingTileSelectionQueue(ctx context.Context, gameID, playerID string) (*PendingTileSelectionQueue, error) {
	queue, err := r.playerRepo.GetPendingTileSelectionQueue(ctx, gameID, playerID)
	if err != nil {
		return nil, err
	}
	if queue == nil {
		return nil, nil
	}
	result := toPendingTileSelectionQueueModel(*queue)
	return &result, nil
}

func (r *RepositoryImpl) UpdatePendingTileSelection(ctx context.Context, gameID, playerID string, selection *PendingTileSelection) error {
	modelSelection := toModelPendingTileSelection(selection)
	return r.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, modelSelection)
}

// Conversion functions between mechanic types and central model types

func toGameModel(mg model.Game) Game {
	return Game{
		ID:    mg.ID,
		Board: toBoardModel(mg.Board),
	}
}

func toBoardModel(mb model.Board) Board {
	tiles := make([]Tile, len(mb.Tiles))
	for i, t := range mb.Tiles {
		tiles[i] = toTileModel(t)
	}
	return Board{
		Tiles: tiles,
	}
}

func toTileModel(mt model.Tile) Tile {
	var occupiedBy *TileOccupant
	if mt.OccupiedBy != nil {
		occ := toTileOccupantModel(*mt.OccupiedBy)
		occupiedBy = &occ
	}

	bonuses := make([]TileBonus, len(mt.Bonuses))
	for i, b := range mt.Bonuses {
		bonuses[i] = toTileBonusModel(b)
	}

	return Tile{
		Coordinates: toHexPositionModel(mt.Coordinates),
		OccupiedBy:  occupiedBy,
		OwnerID:     mt.OwnerID,
		Bonuses:     bonuses,
	}
}

func toHexPositionModel(mh model.HexPosition) HexPosition {
	return HexPosition{
		Q: mh.Q,
		R: mh.R,
		S: mh.S,
	}
}

func toTileOccupantModel(mo model.TileOccupant) TileOccupant {
	return TileOccupant{
		Type: ResourceType(mo.Type),
		Tags: mo.Tags,
	}
}

func toTileBonusModel(mb model.TileBonus) TileBonus {
	return TileBonus{
		Type:   ResourceType(mb.Type),
		Amount: mb.Amount,
	}
}

func toPlayerModel(mp model.Player) Player {
	return Player{
		ID:          mp.ID,
		Resources:   toResourcesModel(mp.Resources),
		PlayedCards: mp.PlayedCards,
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

func toPendingTileSelectionQueueModel(mq model.PendingTileSelectionQueue) PendingTileSelectionQueue {
	return PendingTileSelectionQueue{
		Items:  mq.Items,
		Source: mq.Source,
	}
}

// Conversion functions from mechanic types to model types

func toModelHexPosition(h HexPosition) model.HexPosition {
	return model.HexPosition{
		Q: h.Q,
		R: h.R,
		S: h.S,
	}
}

func toModelTileOccupant(o *TileOccupant) *model.TileOccupant {
	if o == nil {
		return nil
	}
	return &model.TileOccupant{
		Type: model.ResourceType(o.Type),
		Tags: o.Tags,
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

func toModelPendingTileSelection(s *PendingTileSelection) *model.PendingTileSelection {
	if s == nil {
		return nil
	}
	return &model.PendingTileSelection{
		TileType:       s.TileType,
		Source:         s.Source,
		AvailableHexes: s.AvailableHexes,
	}
}
