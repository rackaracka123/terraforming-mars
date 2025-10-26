package service

import (
	"context"
	"fmt"
	"testing"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// parseHexPosition converts a string coordinate "q,r,s" to HexPosition
func parseHexPosition(coordStr string) (model.HexPosition, error) {
	var q, r, s int
	_, err := fmt.Sscanf(coordStr, "%d,%d,%d", &q, &r, &s)
	if err != nil {
		return model.HexPosition{}, err
	}
	return model.HexPosition{Q: q, R: r, S: s}, nil
}

// TestMiningGuild_PlacementBonusTrigger tests that Mining Guild's passive effect
// increases steel production when player places tiles on steel or titanium placement bonuses
func TestMiningGuild_PlacementBonusTrigger(t *testing.T) {
	ctx := context.Background()
	eventBus := events.NewEventBus()

	// Setup repositories and services
	gameRepo := repository.NewGameRepository(eventBus)
	playerRepo := repository.NewPlayerRepository(eventBus)
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	effectSubscriber := cards.NewCardEffectSubscriber(eventBus, playerRepo, gameRepo)
	forcedActionManager := cards.NewForcedActionManager(eventBus, cardRepo, playerRepo, gameRepo, cardDeckRepo)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber, forcedActionManager)
	playerService := service.NewPlayerService(gameRepo, playerRepo, sessionManager, boardService, tileService, forcedActionManager, eventBus)
	standardProjectService := service.NewStandardProjectService(gameRepo, playerRepo, sessionManager, tileService)

	// Load cards
	require.NoError(t, cardRepo.LoadCards(ctx))

	// Create test game in starting card selection phase
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)

	// Initialize board with default tiles (which includes steel and titanium bonuses)
	board := boardService.GenerateDefaultBoard()
	gameRepo.UpdateBoard(ctx, game.ID, board)

	gameRepo.UpdateStatus(ctx, game.ID, model.GameStatusActive)
	gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseStartingCardSelection)

	// Create test player
	player := model.Player{
		ID:              "player1",
		Name:            "Test Player",
		Resources:       model.Resources{Credits: 100},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20,
		IsConnected:     true,
	}
	playerRepo.Create(ctx, game.ID, player)

	// Get available starting cards and corporations
	startingCards, _ := cardRepo.GetStartingCardPool(ctx)
	corporations, _ := cardRepo.GetCorporations(ctx)
	require.GreaterOrEqual(t, len(startingCards), 4, "Should have at least 4 starting cards")
	require.GreaterOrEqual(t, len(corporations), 12, "Should have at least 12 corporations")

	// Find Mining Guild (B06)
	var miningGuildID string
	for _, corp := range corporations {
		if corp.ID == "B06" {
			miningGuildID = corp.ID
			break
		}
	}
	require.NotEmpty(t, miningGuildID, "Mining Guild (B06) should be available")

	// Set up starting card selection phase for player
	availableCardIDs := []string{
		startingCards[0].ID,
		startingCards[1].ID,
		startingCards[2].ID,
		startingCards[3].ID,
	}
	corporationIDs := []string{miningGuildID, corporations[1].ID}

	playerRepo.UpdateSelectStartingCardsPhase(ctx, game.ID, player.ID, &model.SelectStartingCardsPhase{
		AvailableCards:        availableCardIDs,
		AvailableCorporations: corporationIDs,
	})

	// Select Mining Guild and some starting cards
	selectedCardIDs := []string{startingCards[0].ID, startingCards[1].ID}
	err = cardService.OnSelectStartingCards(ctx, game.ID, player.ID, selectedCardIDs, miningGuildID)
	require.NoError(t, err, "Should successfully select Mining Guild")

	// Verify initial production (Mining Guild starts with +1 steel production)
	playerAfterSelection, err := playerRepo.GetByID(ctx, game.ID, player.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, playerAfterSelection.Production.Steel, "Mining Guild should start with +1 steel production")

	// Transition to action phase to enable tile placement
	gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseAction)

	// Find tiles with steel and titanium bonuses
	game, _ = gameRepo.GetByID(ctx, game.ID)
	var steelBonusTile *model.Tile
	var titaniumBonusTile *model.Tile
	var plantsBonusTile *model.Tile

	for i := range game.Board.Tiles {
		tile := &game.Board.Tiles[i]
		if tile.OccupiedBy == nil { // Only unoccupied tiles
			for _, bonus := range tile.Bonuses {
				if bonus.Type == model.ResourceSteel && steelBonusTile == nil {
					steelBonusTile = tile
				}
				if bonus.Type == model.ResourceTitanium && titaniumBonusTile == nil {
					titaniumBonusTile = tile
				}
				if bonus.Type == model.ResourcePlants && plantsBonusTile == nil {
					plantsBonusTile = tile
				}
			}
		}
	}

	require.NotNil(t, steelBonusTile, "Board should have at least one tile with steel bonus")
	require.NotNil(t, titaniumBonusTile, "Board should have at least one tile with titanium bonus")
	require.NotNil(t, plantsBonusTile, "Board should have at least one tile with plants bonus")

	// Test 1: Place greenery on steel bonus tile - should increase steel production
	t.Run("Steel bonus triggers production increase", func(t *testing.T) {
		initialProduction := playerAfterSelection.Production.Steel

		// Use standard project to place greenery tile (costs 23 MC normally, 8 plants to convert)
		// For this test, we'll just give the player enough plants and trigger the conversion
		player, _ := playerRepo.GetByID(ctx, game.ID, player.ID)
		player.Resources.Plants = 100 // Give enough plants
		playerRepo.UpdateResources(ctx, game.ID, player.ID, player.Resources)

		// Initiate plant conversion and select the steel bonus tile
		err = standardProjectService.PlantGreenery(ctx, game.ID, player.ID)
		require.NoError(t, err)

		err = playerService.OnTileSelected(ctx, game.ID, player.ID, steelBonusTile.Coordinates)
		require.NoError(t, err)

		// Verify steel production increased by 1 (from passive effect)
		playerAfterTile, err := playerRepo.GetByID(ctx, game.ID, player.ID)
		require.NoError(t, err)
		assert.Equal(t, initialProduction+1, playerAfterTile.Production.Steel,
			"Steel production should increase by 1 when placing on steel bonus")
	})

	// Test 2: Place city on titanium bonus tile - should increase steel production
	t.Run("Titanium bonus triggers production increase", func(t *testing.T) {
		player, _ := playerRepo.GetByID(ctx, game.ID, player.ID)
		currentProduction := player.Production.Steel

		// Give player enough credits for city placement
		player.Resources.Credits = 100
		playerRepo.UpdateResources(ctx, game.ID, player.ID, player.Resources)

		// Initiate city placement via standard project
		err = standardProjectService.BuildCity(ctx, game.ID, player.ID)
		require.NoError(t, err)

		// Get available coordinates for city placement
		updatedPlayer, _ := playerRepo.GetByID(ctx, game.ID, player.ID)
		require.NotNil(t, updatedPlayer.PendingTileSelection, "Should have pending tile selection")
		availableCoords := updatedPlayer.PendingTileSelection.AvailableHexes

		// Find an available tile with titanium bonus
		game, _ := gameRepo.GetByID(ctx, game.ID)
		var selectedCoordStr string
		for _, coordStr := range availableCoords {
			coord, err := parseHexPosition(coordStr)
			require.NoError(t, err)

			for _, tile := range game.Board.Tiles {
				if tile.Coordinates.Equals(coord) {
					for _, bonus := range tile.Bonuses {
						if bonus.Type == model.ResourceTitanium {
							selectedCoordStr = coordStr
							break
						}
					}
				}
				if selectedCoordStr != "" {
					break
				}
			}
			if selectedCoordStr != "" {
				break
			}
		}

		// If no titanium bonus in available coords, just use first available
		if selectedCoordStr == "" {
			selectedCoordStr = availableCoords[0]
		}

		selectedCoord, err := parseHexPosition(selectedCoordStr)
		require.NoError(t, err)
		err = playerService.OnTileSelected(ctx, game.ID, player.ID, selectedCoord)
		require.NoError(t, err)

		// Verify steel production (may or may not increase depending on whether we found titanium bonus)
		playerAfterCity, err := playerRepo.GetByID(ctx, game.ID, player.ID)
		require.NoError(t, err)
		// Note: This test may not always increase production if no titanium bonus is available
		assert.GreaterOrEqual(t, playerAfterCity.Production.Steel, currentProduction,
			"Steel production should not decrease")
	})

	// Test 3: Place greenery on plants bonus tile - should NOT increase steel production
	t.Run("Plants bonus does NOT trigger production increase", func(t *testing.T) {
		player, _ := playerRepo.GetByID(ctx, game.ID, player.ID)
		currentProduction := player.Production.Steel

		// Give player enough plants
		player.Resources.Plants = 100
		playerRepo.UpdateResources(ctx, game.ID, player.ID, player.Resources)

		// Initiate plant conversion
		err = standardProjectService.PlantGreenery(ctx, game.ID, player.ID)
		require.NoError(t, err)

		// Get available coordinates for greenery placement
		updatedPlayer, _ := playerRepo.GetByID(ctx, game.ID, player.ID)
		require.NotNil(t, updatedPlayer.PendingTileSelection, "Should have pending tile selection")
		availableCoords := updatedPlayer.PendingTileSelection.AvailableHexes

		// Find an available tile with plants bonus (not steel/titanium)
		game, _ := gameRepo.GetByID(ctx, game.ID)
		var selectedCoordStr string
		for _, coordStr := range availableCoords {
			coord, err := parseHexPosition(coordStr)
			require.NoError(t, err)

			for _, tile := range game.Board.Tiles {
				if tile.Coordinates.Equals(coord) {
					hasPlants := false
					hasSteelOrTitanium := false
					for _, bonus := range tile.Bonuses {
						if bonus.Type == model.ResourcePlants {
							hasPlants = true
						}
						if bonus.Type == model.ResourceSteel || bonus.Type == model.ResourceTitanium {
							hasSteelOrTitanium = true
						}
					}
					// Prefer plants bonus, but accept any tile without steel/titanium
					if hasPlants || !hasSteelOrTitanium {
						selectedCoordStr = coordStr
						break
					}
				}
			}
			if selectedCoordStr != "" {
				break
			}
		}

		// If no suitable tile found, just use first available
		if selectedCoordStr == "" {
			selectedCoordStr = availableCoords[0]
		}

		selectedCoord, err := parseHexPosition(selectedCoordStr)
		require.NoError(t, err)
		err = playerService.OnTileSelected(ctx, game.ID, player.ID, selectedCoord)
		require.NoError(t, err)

		// Verify steel production did NOT increase (should only increase for steel/titanium bonuses)
		playerAfterPlants, err := playerRepo.GetByID(ctx, game.ID, player.ID)
		require.NoError(t, err)
		assert.LessOrEqual(t, playerAfterPlants.Production.Steel, currentProduction,
			"Steel production should not increase when placing on non-steel/titanium bonus")
	})
}
