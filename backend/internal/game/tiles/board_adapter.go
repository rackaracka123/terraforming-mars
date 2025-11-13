package tiles

import (
	"terraforming-mars-backend/internal/model"
)

// ExternalBoardService defines the interface for the external board service
// This avoids import cycles by not importing service package
type ExternalBoardService interface {
	GenerateDefaultBoard() model.Board
	CalculateAvailableHexesForTileType(game model.Game, tileType string) ([]string, error)
	CalculateAvailableHexesForTileTypeWithPlayer(game model.Game, tileType, playerID string) ([]string, error)
}

// BoardServiceAdapter adapts ExternalBoardService to tiles.BoardService
// by converting between model.Game and tiles.Game types
type BoardServiceAdapter struct {
	boardService ExternalBoardService
}

// NewBoardServiceAdapter creates a new adapter
func NewBoardServiceAdapter(boardService ExternalBoardService) BoardService {
	return &BoardServiceAdapter{
		boardService: boardService,
	}
}

// GenerateDefaultBoard generates the default Mars board
func (a *BoardServiceAdapter) GenerateDefaultBoard() Board {
	modelBoard := a.boardService.GenerateDefaultBoard()
	return boardFromModel(modelBoard)
}

// CalculateAvailableHexesForTileType calculates available hexes for tile placement
func (a *BoardServiceAdapter) CalculateAvailableHexesForTileType(game Game, tileType string) ([]string, error) {
	modelGame := gameToModel(game)
	return a.boardService.CalculateAvailableHexesForTileType(modelGame, tileType)
}

// CalculateAvailableHexesForTileTypeWithPlayer calculates available hexes for tile placement with player context
func (a *BoardServiceAdapter) CalculateAvailableHexesForTileTypeWithPlayer(game Game, tileType, playerID string) ([]string, error) {
	modelGame := gameToModel(game)
	return a.boardService.CalculateAvailableHexesForTileTypeWithPlayer(modelGame, tileType, playerID)
}

// gameToModel converts tiles.Game to model.Game
func gameToModel(game Game) model.Game {
	return model.Game{
		ID:    game.ID,
		Board: boardToModel(game.Board),
	}
}

// gameFromModel converts model.Game to tiles.Game
func gameFromModel(game model.Game) Game {
	return Game{
		ID:    game.ID,
		Board: boardFromModel(game.Board),
	}
}

// boardToModel converts tiles.Board to model.Board
func boardToModel(board Board) model.Board {
	tiles := make([]model.Tile, len(board.Tiles))
	for i, tile := range board.Tiles {
		tiles[i] = tileToModel(tile)
	}
	return model.Board{
		Tiles: tiles,
	}
}

// boardFromModel converts model.Board to tiles.Board
func boardFromModel(board model.Board) Board {
	tiles := make([]Tile, len(board.Tiles))
	for i, tile := range board.Tiles {
		tiles[i] = tileFromModel(tile)
	}
	return Board{
		Tiles: tiles,
	}
}

// tileToModel converts tiles.Tile to model.Tile
func tileToModel(tile Tile) model.Tile {
	var occupiedBy *model.TileOccupant
	if tile.OccupiedBy != nil {
		occupiedBy = &model.TileOccupant{
			Type: model.ResourceType(tile.OccupiedBy.Type),
			Tags: tile.OccupiedBy.Tags,
		}
	}

	bonuses := make([]model.TileBonus, len(tile.Bonuses))
	for i, bonus := range tile.Bonuses {
		bonuses[i] = model.TileBonus{
			Type:   model.ResourceType(bonus.Type),
			Amount: bonus.Amount,
		}
	}

	return model.Tile{
		Coordinates: model.HexPosition{
			Q: tile.Coordinates.Q,
			R: tile.Coordinates.R,
			S: tile.Coordinates.S,
		},
		OccupiedBy: occupiedBy,
		OwnerID:    tile.OwnerID,
		Bonuses:    bonuses,
	}
}

// tileFromModel converts model.Tile to tiles.Tile
func tileFromModel(tile model.Tile) Tile {
	var occupiedBy *TileOccupant
	if tile.OccupiedBy != nil {
		occupiedBy = &TileOccupant{
			Type: ResourceType(tile.OccupiedBy.Type),
			Tags: tile.OccupiedBy.Tags,
		}
	}

	bonuses := make([]TileBonus, len(tile.Bonuses))
	for i, bonus := range tile.Bonuses {
		bonuses[i] = TileBonus{
			Type:   ResourceType(bonus.Type),
			Amount: bonus.Amount,
		}
	}

	return Tile{
		Coordinates: HexPosition{
			Q: tile.Coordinates.Q,
			R: tile.Coordinates.R,
			S: tile.Coordinates.S,
		},
		OccupiedBy: occupiedBy,
		OwnerID:    tile.OwnerID,
		Bonuses:    bonuses,
	}
}
