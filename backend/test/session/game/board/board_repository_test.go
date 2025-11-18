package board_test

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/session/game/board"
)

func TestBoardRepository_GenerateBoard(t *testing.T) {
	// Create repository
	repo := board.NewRepository()
	ctx := context.Background()
	gameID := "test-game-1"

	// Generate board
	err := repo.GenerateBoard(ctx, gameID)
	if err != nil {
		t.Fatalf("GenerateBoard failed: %v", err)
	}

	// Get board
	b, err := repo.GetByGameID(ctx, gameID)
	if err != nil {
		t.Fatalf("GetByGameID failed: %v", err)
	}

	// Verify board has 61 tiles (pattern: 5-6-7-8-9-8-7-6-5)
	if len(b.Tiles) != 61 {
		t.Errorf("Expected 61 tiles, got %d", len(b.Tiles))
	}

	// Verify some tiles are ocean spaces
	oceanCount := 0
	for _, tile := range b.Tiles {
		if tile.Type == board.ResourceOceanTile {
			oceanCount++
		}
	}

	if oceanCount != 9 {
		t.Errorf("Expected 9 ocean tiles, got %d", oceanCount)
	}
}

func TestBoardRepository_GetByGameID_NotFound(t *testing.T) {
	repo := board.NewRepository()
	ctx := context.Background()

	_, err := repo.GetByGameID(ctx, "nonexistent-game")
	if err == nil {
		t.Error("Expected error for nonexistent game, got nil")
	}
}

func TestBoardRepository_UpdateTileOccupancy(t *testing.T) {
	repo := board.NewRepository()
	ctx := context.Background()
	gameID := "test-game-2"

	// Generate board
	err := repo.GenerateBoard(ctx, gameID)
	if err != nil {
		t.Fatalf("GenerateBoard failed: %v", err)
	}

	// Get a tile coordinate
	b, _ := repo.GetByGameID(ctx, gameID)
	coord := b.Tiles[0].Coordinates

	// Update tile occupancy
	occupant := &board.TileOccupant{
		Type: board.ResourceCityTile,
		Tags: []string{"capital"},
	}
	ownerID := "player-1"

	err = repo.UpdateTileOccupancy(ctx, gameID, coord, occupant, &ownerID)
	if err != nil {
		t.Fatalf("UpdateTileOccupancy failed: %v", err)
	}

	// Verify tile was updated
	tile, err := repo.GetTile(ctx, gameID, coord)
	if err != nil {
		t.Fatalf("GetTile failed: %v", err)
	}

	if tile.OccupiedBy == nil {
		t.Fatal("Expected tile to be occupied")
	}

	if tile.OccupiedBy.Type != board.ResourceCityTile {
		t.Errorf("Expected city tile, got %s", tile.OccupiedBy.Type)
	}

	if *tile.OwnerID != ownerID {
		t.Errorf("Expected owner %s, got %s", ownerID, *tile.OwnerID)
	}
}

func TestBoardProcessor_GenerateTiles(t *testing.T) {
	processor := board.NewBoardProcessor()
	tiles := processor.GenerateTiles()

	// Verify tile count (pattern: 5-6-7-8-9-8-7-6-5 = 61 tiles)
	if len(tiles) != 61 {
		t.Errorf("Expected 61 tiles, got %d", len(tiles))
	}

	// Verify row pattern (5-6-7-8-9-8-7-6-5)
	// This is implicitly verified by the tile count

	// Verify ocean positions
	oceanCount := 0
	for _, tile := range tiles {
		if tile.Type == board.ResourceOceanTile {
			oceanCount++
		}
	}

	if oceanCount != 9 {
		t.Errorf("Expected 9 ocean tiles, got %d", oceanCount)
	}

	// Verify all tiles have valid cube coordinates (q + r + s = 0)
	for _, tile := range tiles {
		sum := tile.Coordinates.Q + tile.Coordinates.R + tile.Coordinates.S
		if sum != 0 {
			t.Errorf("Invalid cube coordinates for tile: q=%d, r=%d, s=%d (sum=%d)",
				tile.Coordinates.Q, tile.Coordinates.R, tile.Coordinates.S, sum)
		}
	}
}

func TestBoardProcessor_CalculateAvailableHexes(t *testing.T) {
	processor := board.NewBoardProcessor()
	tiles := processor.GenerateTiles()
	testBoard := &board.Board{Tiles: tiles}

	// Test ocean hexes
	oceanHexes := processor.CalculateAvailableHexesForTileType(testBoard, "ocean", nil)
	if len(oceanHexes) != 9 {
		t.Errorf("Expected 9 available ocean hexes, got %d", len(oceanHexes))
	}

	// Test city hexes (all non-ocean tiles initially)
	cityHexes := processor.CalculateAvailableHexesForTileType(testBoard, "city", nil)
	expectedCityHexes := 61 - 9 // 61 total - 9 ocean
	if len(cityHexes) != expectedCityHexes {
		t.Errorf("Expected %d available city hexes, got %d", expectedCityHexes, len(cityHexes))
	}

	// Test greenery hexes (same as city initially)
	greeneryHexes := processor.CalculateAvailableHexesForTileType(testBoard, "greenery", nil)
	if len(greeneryHexes) != expectedCityHexes {
		t.Errorf("Expected %d available greenery hexes, got %d", expectedCityHexes, len(greeneryHexes))
	}
}
