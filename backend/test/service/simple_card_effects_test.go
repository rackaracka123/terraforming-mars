package service

import (
	"context"
	"testing"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupCardTest creates a basic game setup for card testing
func setupCardTest(t *testing.T) (context.Context, string, string, service.CardService, repository.PlayerRepository, repository.GameRepository) {
	ctx := context.Background()

	// Setup repositories and services
	gameRepo := repository.NewGameRepository()
	playerRepo := repository.NewPlayerRepository()
	cardRepo := repository.NewCardRepository()
	cardDeckRepo := repository.NewCardDeckRepository()
	sessionManager := test.NewMockSessionManager()
	boardService := service.NewBoardService()
	tileService := service.NewTileService(gameRepo, playerRepo, boardService)
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService)

	// Load cards
	require.NoError(t, cardRepo.LoadCards(ctx))

	// Create test game in action phase
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)
	gameRepo.UpdateStatus(ctx, game.ID, model.GameStatusActive)
	gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseAction)

	// Create test player
	player := model.Player{
		ID: "player1", Name: "Test Player",
		Resources:       model.Resources{Credits: 50},
		Production:      model.Production{Credits: 1},
		TerraformRating: 20, IsConnected: true,
		AvailableActions: 2, // Sufficient actions for testing
	}

	playerRepo.Create(ctx, game.ID, player)
	gameRepo.AddPlayerID(ctx, game.ID, player.ID)
	gameRepo.UpdateCurrentTurn(ctx, game.ID, &player.ID)

	return ctx, game.ID, player.ID, cardService, playerRepo, gameRepo
}

func TestColonizerTrainingCamp(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, gameRepo := setupCardTest(t)

	// Set low oxygen to meet card requirement (≤ 5%)
	gameRepo.UpdateGlobalParameters(ctx, gameID, model.GlobalParameters{
		Temperature: -30, Oxygen: 3, Oceans: 0,
	})

	// Add card to player's hand and verify initial state
	cardID := "001"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	playerBefore, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Contains(t, playerBefore.Cards, cardID)
	assert.Equal(t, 50, playerBefore.Resources.Credits)

	// Play the card
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, nil)
	require.NoError(t, err, "Should successfully play Colonizer Training Camp")

	// Verify effects
	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.NotContains(t, playerAfter.Cards, cardID, "Card removed from hand")
	assert.Contains(t, playerAfter.PlayedCards, cardID, "Card added to played cards")
	assert.Equal(t, 42, playerAfter.Resources.Credits, "8 MC cost deducted")

	t.Log("✅ Colonizer Training Camp test passed")
}

func TestColonizerTrainingCamp_RequirementNotMet(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, gameRepo := setupCardTest(t)

	// Set high oxygen to violate card requirement (> 5%)
	gameRepo.UpdateGlobalParameters(ctx, gameID, model.GlobalParameters{
		Temperature: -30, Oxygen: 8, Oceans: 0, // Above 5% requirement
	})

	cardID := "001"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	// Attempt to play the card - should fail
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, nil)
	assert.Error(t, err, "Should fail when oxygen requirement not met")
	assert.Contains(t, err.Error(), "requirements not met")

	// Verify card still in hand, credits unchanged
	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Contains(t, playerAfter.Cards, cardID, "Card still in hand")
	assert.Equal(t, 50, playerAfter.Resources.Credits, "Credits unchanged")

	t.Log("✅ Colonizer Training Camp requirement test passed")
}

func TestSpaceElevator_ImmediateEffect(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _ := setupCardTest(t)

	// Set player resources to afford Space Elevator (27 MC)
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 50})

	// Add Space Elevator to player's hand
	cardID := "013"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	// Get initial state
	playerBefore, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 0, playerBefore.Production.Titanium, "Should start with 0 titanium production")
	assert.Equal(t, 50, playerBefore.Resources.Credits, "Should start with 50 credits")

	// Play Space Elevator
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, nil)
	require.NoError(t, err, "Should successfully play Space Elevator")

	// Verify immediate effects applied
	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 1, playerAfter.Production.Titanium, "Should gain +1 titanium production")
	assert.Equal(t, 23, playerAfter.Resources.Credits, "Should have 50-27=23 credits after cost")
	assert.Contains(t, playerAfter.PlayedCards, cardID, "Card should be played")
	assert.NotEmpty(t, playerAfter.Actions, "Should have available actions from the card")

	// Verify the action was added to player's action list
	hasSpaceElevatorAction := false
	for _, action := range playerAfter.Actions {
		if action.CardID == cardID {
			hasSpaceElevatorAction = true
			assert.Equal(t, "Space Elevator", action.CardName)
			assert.Equal(t, 0, action.PlayCount, "Action should not be played yet")
			// Verify action inputs/outputs
			assert.Len(t, action.Behavior.Inputs, 1, "Should have 1 input")
			assert.Equal(t, model.ResourceSteel, action.Behavior.Inputs[0].Type)
			assert.Equal(t, 1, action.Behavior.Inputs[0].Amount)
			assert.Len(t, action.Behavior.Outputs, 1, "Should have 1 output")
			assert.Equal(t, model.ResourceCredits, action.Behavior.Outputs[0].Type)
			assert.Equal(t, 5, action.Behavior.Outputs[0].Amount)
			break
		}
	}
	assert.True(t, hasSpaceElevatorAction, "Should have Space Elevator action available")

	t.Log("✅ Space Elevator immediate effect test passed")
}

func TestSpaceElevator_ActionUse(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _ := setupCardTest(t)

	// Set player resources with steel for the action
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 50, Steel: 3})
	cardID := "013"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	// Play Space Elevator first to get the action
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, nil)
	require.NoError(t, err)

	// Get state after playing card
	playerBefore, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 3, playerBefore.Resources.Steel, "Should have 3 steel")
	assert.Equal(t, 23, playerBefore.Resources.Credits, "Should have 23 credits after card cost")

	// Use the Space Elevator action (spend 1 steel → gain 5 credits)
	behaviorIndex := 1 // The manual action is the second behavior
	err = cardService.OnPlayCardAction(ctx, gameID, playerID, cardID, behaviorIndex, nil)
	require.NoError(t, err, "Should successfully use Space Elevator action")

	// Verify action effects
	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 2, playerAfter.Resources.Steel, "Should spend 1 steel (3-1=2)")
	assert.Equal(t, 28, playerAfter.Resources.Credits, "Should gain 5 credits (23+5=28)")

	// Verify action play count incremented
	actionUsed := false
	for _, action := range playerAfter.Actions {
		if action.CardID == cardID && action.BehaviorIndex == behaviorIndex {
			assert.Equal(t, 1, action.PlayCount, "Action should be played once")
			actionUsed = true
			break
		}
	}
	assert.True(t, actionUsed, "Should find used action with play count")

	t.Log("✅ Space Elevator action use test passed")
}
