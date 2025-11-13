package tiles_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game/tiles"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// boardServiceAdapter adapts service.BoardService to tiles.BoardService
type boardServiceAdapter struct {
	boardService service.BoardService
}

func (a *boardServiceAdapter) GenerateDefaultBoard() tiles.Board {
	modelBoard := a.boardService.GenerateDefaultBoard()
	return toTilesBoard(modelBoard)
}

func (a *boardServiceAdapter) CalculateAvailableHexesForTileType(game tiles.Game, tileType string) ([]string, error) {
	modelGame := toModelGame(game)
	return a.boardService.CalculateAvailableHexesForTileType(modelGame, tileType)
}

func (a *boardServiceAdapter) CalculateAvailableHexesForTileTypeWithPlayer(game tiles.Game, tileType, playerID string) ([]string, error) {
	modelGame := toModelGame(game)
	return a.boardService.CalculateAvailableHexesForTileTypeWithPlayer(modelGame, tileType, playerID)
}

// Conversion helpers
func toModelGame(g tiles.Game) model.Game {
	return model.Game{
		ID:    g.ID,
		Board: toModelBoard(g.Board),
	}
}

func toModelBoard(b tiles.Board) model.Board {
	modelTiles := make([]model.Tile, len(b.Tiles))
	for i, t := range b.Tiles {
		modelTiles[i] = toModelTile(t)
	}
	return model.Board{
		Tiles: modelTiles,
	}
}

func toModelTile(t tiles.Tile) model.Tile {
	var occupiedBy *model.TileOccupant
	if t.OccupiedBy != nil {
		occ := model.TileOccupant{
			Type: model.ResourceType(t.OccupiedBy.Type),
			Tags: t.OccupiedBy.Tags,
		}
		occupiedBy = &occ
	}

	bonuses := make([]model.TileBonus, len(t.Bonuses))
	for i, b := range t.Bonuses {
		bonuses[i] = model.TileBonus{
			Type:   model.ResourceType(b.Type),
			Amount: b.Amount,
		}
	}

	return model.Tile{
		Coordinates: model.HexPosition{
			Q: t.Coordinates.Q,
			R: t.Coordinates.R,
			S: t.Coordinates.S,
		},
		OccupiedBy: occupiedBy,
		OwnerID:    t.OwnerID,
		Bonuses:    bonuses,
	}
}

func toTilesBoard(mb model.Board) tiles.Board {
	tilesList := make([]tiles.Tile, len(mb.Tiles))
	for i, t := range mb.Tiles {
		tilesList[i] = toTilesTile(t)
	}
	return tiles.Board{
		Tiles: tilesList,
	}
}

func toTilesTile(mt model.Tile) tiles.Tile {
	var occupiedBy *tiles.TileOccupant
	if mt.OccupiedBy != nil {
		occ := tiles.TileOccupant{
			Type: tiles.ResourceType(mt.OccupiedBy.Type),
			Tags: mt.OccupiedBy.Tags,
		}
		occupiedBy = &occ
	}

	bonuses := make([]tiles.TileBonus, len(mt.Bonuses))
	for i, b := range mt.Bonuses {
		bonuses[i] = tiles.TileBonus{
			Type:   tiles.ResourceType(b.Type),
			Amount: b.Amount,
		}
	}

	return tiles.Tile{
		Coordinates: tiles.HexPosition{
			Q: mt.Coordinates.Q,
			R: mt.Coordinates.R,
			S: mt.Coordinates.S,
		},
		OccupiedBy: occupiedBy,
		OwnerID:    mt.OwnerID,
		Bonuses:    bonuses,
	}
}

func setupTest(t *testing.T) (tiles.Service, repository.GameRepository, repository.PlayerRepository, string, string) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	centralBoardService := service.NewBoardService()
	boardService := &boardServiceAdapter{boardService: centralBoardService}

	tilesRepo := tiles.NewRepository(gameRepo, playerRepo)
	tilesService := tiles.NewService(tilesRepo, boardService, eventBus)

	// Create game
	game, err := gameRepo.Create(ctx, model.GameSettings{
		MaxPlayers: 4,
	})
	require.NoError(t, err)
	gameID := game.ID

	// Initialize board using boardService
	board := boardService.GenerateDefaultBoard()
	modelBoard := toModelBoard(board)
	err = gameRepo.UpdateBoard(ctx, gameID, modelBoard)
	require.NoError(t, err)

	// Create player
	playerID := "test-player-id"
	player := model.Player{
		ID:   playerID,
		Name: "TestPlayer",
		Resources: model.Resources{
			Credits:  50,
			Steel:    10,
			Titanium: 5,
			Plants:   20,
			Energy:   15,
			Heat:     8,
		},
		TerraformRating: 20,
	}
	err = playerRepo.Create(ctx, gameID, player)
	require.NoError(t, err)

	return tilesService, gameRepo, playerRepo, gameID, playerID
}

func TestTilesService_PlaceTile(t *testing.T) {
	tilesService, gameRepo, _, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Find an empty land tile (non-ocean)
	game, err := gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)

	var emptyTileCoord model.HexPosition
	found := false
	for _, tile := range game.Board.Tiles {
		if tile.Type != model.ResourceOceanTile && tile.OccupiedBy == nil {
			emptyTileCoord = tile.Coordinates
			found = true
			break
		}
	}
	require.True(t, found, "Should find an empty land tile")

	// Place a city tile (convert model.HexPosition to tiles.HexPosition)
	err = tilesService.PlaceTile(ctx, gameID, playerID, "city", tiles.HexPosition{
		Q: emptyTileCoord.Q,
		R: emptyTileCoord.R,
		S: emptyTileCoord.S,
	})
	assert.NoError(t, err)

	// Verify tile was placed
	game, err = gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)

	var placedTile *model.Tile
	for i, tile := range game.Board.Tiles {
		if tile.Coordinates.Equals(emptyTileCoord) {
			placedTile = &game.Board.Tiles[i]
			break
		}
	}

	require.NotNil(t, placedTile, "Tile should exist in board")
	require.NotNil(t, placedTile.OccupiedBy, "Tile should be occupied")
	assert.Equal(t, model.ResourceCityTile, placedTile.OccupiedBy.Type)
	require.NotNil(t, placedTile.OwnerID, "Tile should have owner")
	assert.Equal(t, playerID, *placedTile.OwnerID)
}

func TestTilesService_ValidatePlacementOceanLimit(t *testing.T) {
	tilesService, gameRepo, _, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Place maximum oceans (9)
	game, err := gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)

	oceansPlaced := 0
	for _, tile := range game.Board.Tiles {
		if tile.Type == model.ResourceOceanTile && oceansPlaced < model.MaxOceans {
			occupant := &model.TileOccupant{Type: model.ResourceOceanTile}
			err := gameRepo.UpdateTileOccupancy(ctx, gameID, tile.Coordinates, occupant, &playerID)
			require.NoError(t, err)
			oceansPlaced++
		}
	}

	assert.Equal(t, model.MaxOceans, oceansPlaced)

	// Try to validate placing another ocean (should fail)
	err = tilesService.ValidatePlacement(ctx, gameID, "ocean")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum oceans")
}

func TestTilesService_ValidatePlacementNonOcean(t *testing.T) {
	tilesService, _, _, gameID, _ := setupTest(t)
	ctx := context.Background()

	// Non-ocean tiles should always validate (no count limit)
	err := tilesService.ValidatePlacement(ctx, gameID, "city")
	assert.NoError(t, err)

	err = tilesService.ValidatePlacement(ctx, gameID, "greenery")
	assert.NoError(t, err)
}

func TestTilesService_CalculateAvailableHexes(t *testing.T) {
	tilesService, gameRepo, _, gameID, playerID := setupTest(t)
	ctx := context.Background()

	t.Run("City hexes", func(t *testing.T) {
		hexes, err := tilesService.CalculateAvailableHexes(ctx, gameID, playerID, "city")
		assert.NoError(t, err)
		assert.Greater(t, len(hexes), 0, "Should have available city hexes")
	})

	t.Run("Greenery hexes", func(t *testing.T) {
		hexes, err := tilesService.CalculateAvailableHexes(ctx, gameID, playerID, "greenery")
		assert.NoError(t, err)
		assert.Greater(t, len(hexes), 0, "Should have available greenery hexes")
	})

	t.Run("Ocean hexes", func(t *testing.T) {
		hexes, err := tilesService.CalculateAvailableHexes(ctx, gameID, playerID, "ocean")
		assert.NoError(t, err)
		assert.Equal(t, 9, len(hexes), "Should have 9 ocean spaces on default board")
	})

	t.Run("Greenery adjacency rule", func(t *testing.T) {
		// Place a city tile for the player
		game, err := gameRepo.GetByID(ctx, gameID)
		require.NoError(t, err)

		var cityCoord model.HexPosition
		for _, tile := range game.Board.Tiles {
			if tile.Type != model.ResourceOceanTile && tile.OccupiedBy == nil {
				cityCoord = tile.Coordinates
				break
			}
		}

		occupant := &model.TileOccupant{Type: model.ResourceCityTile}
		err = gameRepo.UpdateTileOccupancy(ctx, gameID, cityCoord, occupant, &playerID)
		require.NoError(t, err)

		// Calculate greenery hexes for player - should only return adjacent tiles
		hexes, err := tilesService.CalculateAvailableHexes(ctx, gameID, playerID, "greenery")
		assert.NoError(t, err)
		assert.Greater(t, len(hexes), 0, "Should have greenery hexes adjacent to player's tile")

		// Verify all returned hexes are adjacent to player's city
		neighbors := cityCoord.GetNeighbors()
		neighborSet := make(map[string]bool)
		for _, n := range neighbors {
			neighborSet[n.String()] = true
		}

		for _, hex := range hexes {
			assert.True(t, neighborSet[hex], "Greenery hex should be adjacent to player's tile")
		}
	})
}

func TestTilesService_AwardPlacementBonuses(t *testing.T) {
	tilesService, gameRepo, playerRepo, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Find a tile with bonuses
	game, err := gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)

	var tileWithBonuses *model.Tile
	for i, tile := range game.Board.Tiles {
		if len(tile.Bonuses) > 0 && tile.Type != model.ResourceOceanTile {
			tileWithBonuses = &game.Board.Tiles[i]
			break
		}
	}
	require.NotNil(t, tileWithBonuses, "Should find a tile with bonuses")

	// Get initial resources
	player, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	initialResources := player.Resources

	// Award bonuses (convert model.HexPosition to tiles.HexPosition)
	err = tilesService.AwardPlacementBonuses(ctx, gameID, playerID, tiles.HexPosition{
		Q: tileWithBonuses.Coordinates.Q,
		R: tileWithBonuses.Coordinates.R,
		S: tileWithBonuses.Coordinates.S,
	})
	assert.NoError(t, err)

	// Verify resources increased
	player, err = playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)

	// At least one resource should have increased
	resourcesChanged := player.Resources.Steel != initialResources.Steel ||
		player.Resources.Titanium != initialResources.Titanium ||
		player.Resources.Plants != initialResources.Plants

	assert.True(t, resourcesChanged, "Some resource should have increased from tile bonuses")
}

func TestTilesService_OceanAdjacencyBonus(t *testing.T) {
	tilesService, gameRepo, playerRepo, gameID, playerID := setupTest(t)
	ctx := context.Background()

	// Place an ocean tile
	game, err := gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)

	var oceanCoord model.HexPosition
	for _, tile := range game.Board.Tiles {
		if tile.Type == model.ResourceOceanTile {
			oceanCoord = tile.Coordinates
			break
		}
	}

	occupant := &model.TileOccupant{Type: model.ResourceOceanTile}
	err = gameRepo.UpdateTileOccupancy(ctx, gameID, oceanCoord, occupant, &playerID)
	require.NoError(t, err)

	// Find a non-ocean tile adjacent to the ocean
	game, err = gameRepo.GetByID(ctx, gameID)
	require.NoError(t, err)

	neighbors := oceanCoord.GetNeighbors()
	var adjacentCoord model.HexPosition
	found := false
	for _, neighbor := range neighbors {
		for _, tile := range game.Board.Tiles {
			if tile.Coordinates.Equals(neighbor) && tile.Type != model.ResourceOceanTile {
				adjacentCoord = tile.Coordinates
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	require.True(t, found, "Should find a non-ocean tile adjacent to ocean")

	// Get initial credits
	player, err := playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	initialCredits := player.Resources.Credits

	// Award placement bonuses (includes ocean adjacency) - convert model.HexPosition to tiles.HexPosition
	err = tilesService.AwardPlacementBonuses(ctx, gameID, playerID, tiles.HexPosition{
		Q: adjacentCoord.Q,
		R: adjacentCoord.R,
		S: adjacentCoord.S,
	})
	assert.NoError(t, err)

	// Verify credits increased by at least 2 (base ocean adjacency bonus)
	player, err = playerRepo.GetByID(ctx, gameID, playerID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, player.Resources.Credits, initialCredits+2, "Should get +2 MC from ocean adjacency")
}

func TestTilesService_ProcessTileQueue(t *testing.T) {
	tilesService, gameRepo, playerRepo, gameID, playerID := setupTest(t)
	ctx := context.Background()

	t.Run("Process valid tile from queue", func(t *testing.T) {
		// Create a tile queue with a city tile
		err := playerRepo.CreateTileQueue(ctx, gameID, playerID, "test-card", []string{"city"})
		require.NoError(t, err)

		// Process the queue
		err = tilesService.ProcessTileQueue(ctx, gameID, playerID)
		assert.NoError(t, err)

		// Verify pending selection was set
		pendingSelection, err := playerRepo.GetPendingTileSelection(ctx, gameID, playerID)
		require.NoError(t, err)
		require.NotNil(t, pendingSelection)
		assert.Equal(t, "city", pendingSelection.TileType)
		assert.Greater(t, len(pendingSelection.AvailableHexes), 0)

		// Verify tile was removed from queue
		queue, err := playerRepo.GetPendingTileSelectionQueue(ctx, gameID, playerID)
		require.NoError(t, err)
		// Queue should still exist but be empty after popping one item
		if queue != nil {
			assert.Equal(t, 0, len(queue.Items))
		}
	})

	t.Run("Skip invalid tile and process next", func(t *testing.T) {
		// Place maximum oceans to make ocean placement invalid
		game, err := gameRepo.GetByID(ctx, gameID)
		require.NoError(t, err)

		oceansPlaced := 0
		for _, tile := range game.Board.Tiles {
			if tile.Type == model.ResourceOceanTile && oceansPlaced < model.MaxOceans {
				occupant := &model.TileOccupant{Type: model.ResourceOceanTile}
				err := gameRepo.UpdateTileOccupancy(ctx, gameID, tile.Coordinates, occupant, &playerID)
				require.NoError(t, err)
				oceansPlaced++
			}
		}

		// Create queue with ocean (invalid) and city (valid) tiles
		err = playerRepo.CreateTileQueue(ctx, gameID, playerID, "test-card", []string{"ocean", "city"})
		require.NoError(t, err)

		// Process queue - should skip ocean and process city
		err = tilesService.ProcessTileQueue(ctx, gameID, playerID)
		assert.NoError(t, err)

		// Verify city tile is in pending selection (ocean was skipped)
		pendingSelection, err := playerRepo.GetPendingTileSelection(ctx, gameID, playerID)
		require.NoError(t, err)
		require.NotNil(t, pendingSelection)
		assert.Equal(t, "city", pendingSelection.TileType)
	})

	t.Run("Empty queue does nothing", func(t *testing.T) {
		// Clear queue
		err := playerRepo.ClearPendingTileSelectionQueue(ctx, gameID, playerID)
		require.NoError(t, err)

		// Clear any pending tile selection from previous tests
		err = playerRepo.ClearPendingTileSelection(ctx, gameID, playerID)
		require.NoError(t, err)

		// Process empty queue
		err = tilesService.ProcessTileQueue(ctx, gameID, playerID)
		assert.NoError(t, err)

		// Verify no pending selection was created
		pendingSelection, err := playerRepo.GetPendingTileSelection(ctx, gameID, playerID)
		assert.NoError(t, err)
		assert.Nil(t, pendingSelection, "Should have no pending tile selection when queue is empty")
	})
}
