package tile_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/board"
	"terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/game/tile"
)

// TestTileQueueCreatedEvent_TriggersAutomaticProcessing tests that creating a tile queue
// automatically triggers the TileProcessor via events
func TestTileQueueCreatedEvent_TriggersAutomaticProcessing(t *testing.T) {
	ctx := context.Background()

	// Initialize event bus
	eventBus := events.NewEventBus()

	// Initialize repositories
	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	boardRepo := board.NewRepository(eventBus)

	// Initialize board processor and tile processor
	boardProcessor := board.NewBoardProcessor()
	tileProcessor := tile.NewProcessor(gameRepo, playerRepo, boardRepo, boardProcessor, eventBus)

	// Subscribe to events BEFORE any operations
	tileProcessor.SubscribeToEvents()

	// Create a test game
	testGame := &game.Game{
		ID:         "test-game",
		Status:     "active",
		Phase:      "action",
		PlayerIDs:  []string{"player-1"},
		Generation: 1,
	}
	err := gameRepo.Create(ctx, testGame)
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}

	// Create test board
	testBoard := &board.Board{
		Tiles: make(map[string]*board.Tile),
	}
	// Initialize some tiles for testing
	testBoard.Tiles["0,0,0"] = &board.Tile{
		Coordinate: board.HexCoordinate{Q: 0, R: 0, S: 0},
		Type:       board.TileTypeLand,
	}
	testBoard.Tiles["1,-1,0"] = &board.Tile{
		Coordinate: board.HexCoordinate{Q: 1, R: -1, S: 0},
		Type:       board.TileTypeLand,
	}
	err = boardRepo.Create(ctx, testGame.ID, testBoard)
	if err != nil {
		t.Fatalf("Failed to create test board: %v", err)
	}

	// Create a test player
	testPlayer := &player.Player{
		ID:    "player-1",
		Name:  "Test Player",
		Color: "red",
		Resources: model.Resources{
			Credits: 50,
		},
	}
	err = playerRepo.Create(ctx, testGame.ID, testPlayer)
	if err != nil {
		t.Fatalf("Failed to create test player: %v", err)
	}

	// Create a tile queue (this should trigger the event)
	err = playerRepo.CreateTileQueue(ctx, testGame.ID, testPlayer.ID, "test-card", []string{"city"})
	if err != nil {
		t.Fatalf("Failed to create tile queue: %v", err)
	}

	// Give the event system a moment to process (since it's async)
	time.Sleep(100 * time.Millisecond)

	// Verify that the pending tile selection was set automatically
	updatedPlayer, err := playerRepo.GetByID(ctx, testGame.ID, testPlayer.ID)
	if err != nil {
		t.Fatalf("Failed to get updated player: %v", err)
	}

	if updatedPlayer.PendingTileSelection == nil {
		t.Fatal("Expected PendingTileSelection to be set automatically, but it was nil")
	}

	if updatedPlayer.PendingTileSelection.TileType != "city" {
		t.Errorf("Expected tile type 'city', got %s", updatedPlayer.PendingTileSelection.TileType)
	}

	if len(updatedPlayer.PendingTileSelection.AvailableHexes) == 0 {
		t.Error("Expected available hexes to be calculated, but got empty list")
	}

	t.Logf("✅ Tile queue automatically processed via event. Pending selection: %+v", updatedPlayer.PendingTileSelection)
}

// TestTileQueueCreatedEvent_MultipleTilesInQueue tests that a queue with multiple tiles
// processes only the first tile initially
func TestTileQueueCreatedEvent_MultipleTilesInQueue(t *testing.T) {
	ctx := context.Background()

	// Initialize event bus
	eventBus := events.NewEventBus()

	// Initialize repositories
	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	boardRepo := board.NewRepository(eventBus)

	// Initialize board processor and tile processor
	boardProcessor := board.NewBoardProcessor()
	tileProcessor := tile.NewProcessor(gameRepo, playerRepo, boardRepo, boardProcessor, eventBus)

	// Subscribe to events
	tileProcessor.SubscribeToEvents()

	// Create test game
	testGame := &game.Game{
		ID:         "test-game",
		Status:     "active",
		Phase:      "action",
		PlayerIDs:  []string{"player-1"},
		Generation: 1,
	}
	err := gameRepo.Create(ctx, testGame)
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}

	// Create test board with multiple available tiles
	testBoard := &board.Board{
		Tiles: make(map[string]*board.Tile),
	}
	for q := -2; q <= 2; q++ {
		for r := -2; r <= 2; r++ {
			s := -q - r
			if abs(q) <= 2 && abs(r) <= 2 && abs(s) <= 2 {
				coord := board.HexCoordinate{Q: q, R: r, S: s}
				coordKey := coord.String()
				testBoard.Tiles[coordKey] = &board.Tile{
					Coordinate: coord,
					Type:       board.TileTypeLand,
				}
			}
		}
	}
	err = boardRepo.Create(ctx, testGame.ID, testBoard)
	if err != nil {
		t.Fatalf("Failed to create test board: %v", err)
	}

	// Create test player
	testPlayer := &player.Player{
		ID:    "player-1",
		Name:  "Test Player",
		Color: "red",
	}
	err = playerRepo.Create(ctx, testGame.ID, testPlayer)
	if err != nil {
		t.Fatalf("Failed to create test player: %v", err)
	}

	// Create a queue with multiple tiles: city, city, ocean
	err = playerRepo.CreateTileQueue(ctx, testGame.ID, testPlayer.ID, "test-card", []string{"city", "city", "ocean"})
	if err != nil {
		t.Fatalf("Failed to create tile queue: %v", err)
	}

	// Give event system time to process
	time.Sleep(100 * time.Millisecond)

	// Verify that only the first tile is set up for selection
	updatedPlayer, err := playerRepo.GetByID(ctx, testGame.ID, testPlayer.ID)
	if err != nil {
		t.Fatalf("Failed to get updated player: %v", err)
	}

	if updatedPlayer.PendingTileSelection == nil {
		t.Fatal("Expected PendingTileSelection to be set for first tile")
	}

	if updatedPlayer.PendingTileSelection.TileType != "city" {
		t.Errorf("Expected first tile type 'city', got %s", updatedPlayer.PendingTileSelection.TileType)
	}

	// Verify that the queue still has remaining tiles
	if updatedPlayer.PendingTileSelectionQueue == nil {
		t.Fatal("Expected queue to still exist with remaining tiles")
	}

	// The queue should have 2 items left (the first city was popped, city and ocean remain)
	if len(updatedPlayer.PendingTileSelectionQueue.Items) != 2 {
		t.Errorf("Expected 2 tiles remaining in queue, got %d", len(updatedPlayer.PendingTileSelectionQueue.Items))
	}

	t.Logf("✅ First tile processed. Remaining queue: %v", updatedPlayer.PendingTileSelectionQueue.Items)
}

// TestTileQueueCreatedEvent_OceanValidation tests that ocean tiles are validated
// and skipped if the maximum of 9 oceans has been reached
func TestTileQueueCreatedEvent_OceanValidation(t *testing.T) {
	ctx := context.Background()

	// Initialize event bus
	eventBus := events.NewEventBus()

	// Initialize repositories
	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	boardRepo := board.NewRepository(eventBus)

	// Initialize board processor and tile processor
	boardProcessor := board.NewBoardProcessor()
	tileProcessor := tile.NewProcessor(gameRepo, playerRepo, boardRepo, boardProcessor, eventBus)

	// Subscribe to events
	tileProcessor.SubscribeToEvents()

	// Create test game
	testGame := &game.Game{
		ID:         "test-game",
		Status:     "active",
		Phase:      "action",
		PlayerIDs:  []string{"player-1"},
		Generation: 1,
	}
	err := gameRepo.Create(ctx, testGame)
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}

	// Create test board with 9 oceans already placed (maximum)
	testBoard := &board.Board{
		Tiles: make(map[string]*board.Tile),
	}
	oceanCount := 0
	for q := -2; q <= 2 && oceanCount < 9; q++ {
		for r := -2; r <= 2 && oceanCount < 9; r++ {
			s := -q - r
			if abs(q) <= 2 && abs(r) <= 2 && abs(s) <= 2 {
				coord := board.HexCoordinate{Q: q, R: r, S: s}
				coordKey := coord.String()

				if oceanCount < 9 {
					// Place ocean
					testBoard.Tiles[coordKey] = &board.Tile{
						Coordinate: coord,
						Type:       board.TileTypeOcean,
						OccupiedBy: &board.TileOccupancy{
							Type:     board.ResourceOceanTile,
							PlayerID: "player-1",
						},
					}
					oceanCount++
				} else {
					// Empty land tile
					testBoard.Tiles[coordKey] = &board.Tile{
						Coordinate: coord,
						Type:       board.TileTypeLand,
					}
				}
			}
		}
	}
	err = boardRepo.Create(ctx, testGame.ID, testBoard)
	if err != nil {
		t.Fatalf("Failed to create test board: %v", err)
	}

	// Create test player
	testPlayer := &player.Player{
		ID:    "player-1",
		Name:  "Test Player",
		Color: "red",
	}
	err = playerRepo.Create(ctx, testGame.ID, testPlayer)
	if err != nil {
		t.Fatalf("Failed to create test player: %v", err)
	}

	// Create a queue with ocean and city (ocean should be skipped, city should be processed)
	err = playerRepo.CreateTileQueue(ctx, testGame.ID, testPlayer.ID, "test-card", []string{"ocean", "city"})
	if err != nil {
		t.Fatalf("Failed to create tile queue: %v", err)
	}

	// Give event system time to process
	time.Sleep(100 * time.Millisecond)

	// Verify that ocean was skipped and city is set for selection
	updatedPlayer, err := playerRepo.GetByID(ctx, testGame.ID, testPlayer.ID)
	if err != nil {
		t.Fatalf("Failed to get updated player: %v", err)
	}

	if updatedPlayer.PendingTileSelection == nil {
		t.Fatal("Expected PendingTileSelection to be set for city (ocean should be skipped)")
	}

	if updatedPlayer.PendingTileSelection.TileType != "city" {
		t.Errorf("Expected tile type 'city' (ocean should be skipped), got %s", updatedPlayer.PendingTileSelection.TileType)
	}

	// Queue should be empty since city was the last valid tile
	if updatedPlayer.PendingTileSelectionQueue != nil && len(updatedPlayer.PendingTileSelectionQueue.Items) > 0 {
		t.Errorf("Expected queue to be empty, got %d items", len(updatedPlayer.PendingTileSelectionQueue.Items))
	}

	t.Logf("✅ Ocean tile skipped due to validation, city tile processed")
}

// Helper function for absolute value
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
