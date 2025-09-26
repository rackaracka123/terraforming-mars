package service_test

import (
	"fmt"
	"testing"

	"terraforming-mars-backend/internal/service"
)

func TestBoardService_GenerateDefaultBoard(t *testing.T) {
	srv := service.NewBoardService()
	board := srv.GenerateDefaultBoard()

	// Test that we have the correct number of tiles
	expectedTiles := 5 + 6 + 7 + 8 + 9 + 8 + 7 + 6 + 5 // = 61 tiles
	if len(board.Tiles) != expectedTiles {
		t.Errorf("Expected %d tiles, got %d", expectedTiles, len(board.Tiles))
	}

	// Test row pattern - count tiles per row
	rowPattern := []int{5, 6, 7, 8, 9, 8, 7, 6, 5}
	rowCounts := make(map[int]int)

	for _, tile := range board.Tiles {
		rowCounts[tile.Coordinates.R]++
	}

	// Check each row has the correct number of tiles
	for rowIdx, expectedCount := range rowPattern {
		r := rowIdx - len(rowPattern)/2 // -4 to +4
		actualCount := rowCounts[r]
		if actualCount != expectedCount {
			t.Errorf("Row %d (r=%d): expected %d tiles, got %d", rowIdx, r, expectedCount, actualCount)
		}
	}

	// Test specific coordinates for rows that were problematic (rows 6 and 8, which are r=2 and r=4)
	// Row 6 (r=2) should have 7 tiles with q values from -4 to 2
	row6Tiles := make(map[int]bool)
	for _, tile := range board.Tiles {
		if tile.Coordinates.R == 2 {
			row6Tiles[tile.Coordinates.Q] = true
		}
	}

	expectedRow6Q := []int{-4, -3, -2, -1, 0, 1, 2}
	for _, q := range expectedRow6Q {
		if !row6Tiles[q] {
			t.Errorf("Row 6 (r=2): missing tile with q=%d", q)
		}
	}

	// Row 8 (r=4) should have 5 tiles with q values from -4 to 0
	row8Tiles := make(map[int]bool)
	for _, tile := range board.Tiles {
		if tile.Coordinates.R == 4 {
			row8Tiles[tile.Coordinates.Q] = true
		}
	}

	expectedRow8Q := []int{-4, -3, -2, -1, 0}
	for _, q := range expectedRow8Q {
		if !row8Tiles[q] {
			t.Errorf("Row 8 (r=4): missing tile with q=%d", q)
		}
	}

	// Verify cube coordinate constraint (q + r + s = 0) for all tiles
	for _, tile := range board.Tiles {
		sum := tile.Coordinates.Q + tile.Coordinates.R + tile.Coordinates.S
		if sum != 0 {
			t.Errorf("Invalid cube coordinates for tile at (q=%d, r=%d, s=%d): sum is %d, expected 0",
				tile.Coordinates.Q, tile.Coordinates.R, tile.Coordinates.S, sum)
		}
	}

	// Test that no duplicate coordinates exist
	coordMap := make(map[string]bool)
	for _, tile := range board.Tiles {
		key := fmt.Sprintf("%d,%d,%d", tile.Coordinates.Q, tile.Coordinates.R, tile.Coordinates.S)
		if coordMap[key] {
			t.Errorf("Duplicate coordinate found: (q=%d, r=%d, s=%d)",
				tile.Coordinates.Q, tile.Coordinates.R, tile.Coordinates.S)
		}
		coordMap[key] = true
	}
}

func TestBoardService_TileCoordinateSymmetry(t *testing.T) {
	srv := service.NewBoardService()
	board := srv.GenerateDefaultBoard()

	// The board should be symmetric around the center row (r=0)
	// Check that tiles are properly distributed
	negativeRCount := 0
	zeroRCount := 0
	positiveRCount := 0

	for _, tile := range board.Tiles {
		if tile.Coordinates.R < 0 {
			negativeRCount++
		} else if tile.Coordinates.R == 0 {
			zeroRCount++
		} else {
			positiveRCount++
		}
	}

	// We should have: rows -4,-3,-2,-1 (5+6+7+8=26 tiles) and rows 1,2,3,4 (8+7+6+5=26 tiles)
	// Row 0 has 9 tiles
	if negativeRCount != 26 {
		t.Errorf("Expected 26 tiles with negative R, got %d", negativeRCount)
	}
	if zeroRCount != 9 {
		t.Errorf("Expected 9 tiles with R=0, got %d", zeroRCount)
	}
	if positiveRCount != 26 {
		t.Errorf("Expected 26 tiles with positive R, got %d", positiveRCount)
	}
}
