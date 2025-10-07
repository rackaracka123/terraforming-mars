package service

import (
	"context"
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

// setupCardTest creates a basic game setup for card testing
func setupCardTest(t *testing.T) (context.Context, string, string, service.CardService, repository.PlayerRepository, repository.GameRepository) {
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
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber)

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
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, nil, nil)
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
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, nil, nil)
	assert.Error(t, err, "Should fail when oxygen requirement not met")
	assert.Contains(t, err.Error(), "requirements not met")

	// Verify card still in hand, credits unchanged
	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Contains(t, playerAfter.Cards, cardID, "Card still in hand")
	assert.Equal(t, 50, playerAfter.Resources.Credits, "Credits unchanged")

	t.Log("✅ Colonizer Training Camp requirement test passed")
}

func TestRoverConstruction_PassiveEffectExtraction(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, _ := setupCardTest(t)

	// Set player resources to afford Rover Construction (8 MC)
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 50})

	// Add Rover Construction to player's hand (card 038)
	cardID := "038"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	// Get initial state
	playerBefore, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Empty(t, playerBefore.Effects, "Should start with no passive effects")
	assert.Equal(t, 50, playerBefore.Resources.Credits, "Should start with 50 credits")

	// Play Rover Construction
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, nil, nil)
	require.NoError(t, err, "Should successfully play Rover Construction")

	// Verify passive effect was extracted and added to player
	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 42, playerAfter.Resources.Credits, "Should have 50-8=42 credits after cost")
	assert.Contains(t, playerAfter.PlayedCards, cardID, "Card should be played")
	assert.NotEmpty(t, playerAfter.Effects, "Should have passive effects from the card")

	// Verify the effect was added to player's effect list
	hasRoverConstructionEffect := false
	for _, effect := range playerAfter.Effects {
		if effect.CardID == cardID {
			hasRoverConstructionEffect = true
			assert.Equal(t, "Rover Construction", effect.CardName)
			assert.Equal(t, 0, effect.BehaviorIndex, "Should be the first behavior")

			// Verify effect trigger condition
			assert.Len(t, effect.Behavior.Triggers, 1, "Should have 1 trigger")
			assert.Equal(t, model.ResourceTriggerAuto, effect.Behavior.Triggers[0].Type, "Should be auto trigger")
			assert.NotNil(t, effect.Behavior.Triggers[0].Condition, "Should have a condition")
			assert.Equal(t, model.TriggerCityPlaced, effect.Behavior.Triggers[0].Condition.Type, "Should trigger on city placement")

			// Verify effect outputs
			assert.Len(t, effect.Behavior.Outputs, 1, "Should have 1 output")
			assert.Equal(t, model.ResourceCredits, effect.Behavior.Outputs[0].Type, "Should give credits")
			assert.Equal(t, 2, effect.Behavior.Outputs[0].Amount, "Should give 2 credits")
			assert.Equal(t, model.TargetSelfPlayer, effect.Behavior.Outputs[0].Target, "Should target self")
			break
		}
	}
	assert.True(t, hasRoverConstructionEffect, "Should have Rover Construction passive effect")

	t.Log("✅ Rover Construction passive effect extraction test passed")
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
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, nil, nil)
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
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, nil, nil)
	require.NoError(t, err)

	// Get state after playing card
	playerBefore, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Equal(t, 3, playerBefore.Resources.Steel, "Should have 3 steel")
	assert.Equal(t, 23, playerBefore.Resources.Credits, "Should have 23 credits after card cost")

	// Use the Space Elevator action (spend 1 steel → gain 5 credits)
	behaviorIndex := 1 // The manual action is the second behavior
	err = cardService.OnPlayCardAction(ctx, gameID, playerID, cardID, behaviorIndex, nil, nil)
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

// TestRoverConstruction_PassiveEffectTriggering tests the full event system integration:
// 1. Play Rover Construction card (registers passive effect)
// 2. Simulate a city tile placement event
// 3. Verify passive effect triggers and awards 2 credits
func TestRoverConstruction_PassiveEffectTriggering(t *testing.T) {
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
	cardService := service.NewCardService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, tileService, effectSubscriber)

	// Load cards
	require.NoError(t, cardRepo.LoadCards(ctx))

	// Create test game in action phase
	game, err := gameRepo.Create(ctx, model.GameSettings{MaxPlayers: 4})
	require.NoError(t, err)
	gameRepo.UpdateStatus(ctx, game.ID, model.GameStatusActive)
	gameRepo.UpdatePhase(ctx, game.ID, model.GamePhaseAction)

	// Create test player with enough credits
	player := model.Player{
		ID:               "player1",
		Name:             "Test Player",
		Resources:        model.Resources{Credits: 50},
		Production:       model.Production{Credits: 1},
		TerraformRating:  20,
		IsConnected:      true,
		AvailableActions: 2,
	}

	playerRepo.Create(ctx, game.ID, player)
	gameRepo.AddPlayerID(ctx, game.ID, player.ID)
	gameRepo.UpdateCurrentTurn(ctx, game.ID, &player.ID)

	// Step 1: Play Rover Construction to register passive effect
	roverConstructionCardID := "038"
	playerRepo.AddCard(ctx, game.ID, player.ID, roverConstructionCardID)

	// Get initial credits (should be 50)
	playerBefore, _ := playerRepo.GetByID(ctx, game.ID, player.ID)
	initialCredits := playerBefore.Resources.Credits
	t.Logf("Initial credits: %d", initialCredits)

	// Play Rover Construction
	err = cardService.OnPlayCard(ctx, game.ID, player.ID, roverConstructionCardID, nil, nil)
	require.NoError(t, err, "Should successfully play Rover Construction")

	// Verify card was played and effect registered
	playerAfterCard, _ := playerRepo.GetByID(ctx, game.ID, player.ID)
	creditsAfterCard := playerAfterCard.Resources.Credits
	t.Logf("Credits after playing Rover Construction: %d", creditsAfterCard)
	assert.Equal(t, initialCredits-8, creditsAfterCard, "Should spend 8 MC for Rover Construction")
	assert.Len(t, playerAfterCard.Effects, 1, "Should have 1 passive effect registered")
	assert.Equal(t, model.TriggerCityPlaced, playerAfterCard.Effects[0].Behavior.Triggers[0].Condition.Type, "Effect should trigger on city placement")

	// Step 2: Simulate a city placement event by calling EffectProcessor directly
	coordinate := model.HexPosition{Q: 0, R: 0, S: 0}
	effectContext := model.EffectContext{
		TriggeringPlayerID: player.ID,
		TileCoordinate:     &coordinate,
		TileType:           nil, // Optional field, not needed for this test
	}
	err = effectProcessor.TriggerEffects(ctx, game.ID, model.TriggerCityPlaced, effectContext)
	require.NoError(t, err, "Should successfully trigger passive effects")

	// Step 3: Verify passive effect triggered and awarded 2 credits
	playerAfterEffect, _ := playerRepo.GetByID(ctx, game.ID, player.ID)
	creditsAfterEffect := playerAfterEffect.Resources.Credits
	t.Logf("Credits after triggering city-placed event: %d", creditsAfterEffect)

	// Calculate expected credits:
	// - Start: 50 MC
	// - Play Rover Construction: -8 MC = 42 MC
	// - Rover Construction passive effect: +2 MC = 44 MC
	expectedCredits := 44
	assert.Equal(t, expectedCredits, creditsAfterEffect, "Should gain 2 credits from Rover Construction passive effect")

	t.Log("✅ Rover Construction passive effect triggering integration test passed")
}

func TestLavaFlows_TemperatureIncrease(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, gameRepo := setupCardTest(t)

	// Set initial temperature to -20°C
	gameRepo.UpdateGlobalParameters(ctx, gameID, model.GlobalParameters{
		Temperature: -20, Oxygen: 0, Oceans: 0,
	})

	// Give player enough credits to play Lava Flows (cost: 18)
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 20})

	// Add Lava Flows to player's hand
	cardID := "140"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	// Get state before playing card
	gameBefore, _ := gameRepo.GetByID(ctx, gameID)
	assert.Equal(t, -20, gameBefore.GlobalParameters.Temperature, "Initial temperature should be -20°C")

	playerBefore, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.Contains(t, playerBefore.Cards, cardID, "Card should be in hand")
	assert.Equal(t, 20, playerBefore.Resources.Credits, "Should have 20 credits")

	// Play Lava Flows
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, nil, nil)
	require.NoError(t, err, "Should successfully play Lava Flows")

	// Verify effects
	gameAfter, _ := gameRepo.GetByID(ctx, gameID)
	assert.Equal(t, -18, gameAfter.GlobalParameters.Temperature, "Temperature should increase by 2 steps (-20 + 2 = -18)")

	playerAfter, _ := playerRepo.GetByID(ctx, gameID, playerID)
	assert.NotContains(t, playerAfter.Cards, cardID, "Card removed from hand")
	assert.Contains(t, playerAfter.PlayedCards, cardID, "Card added to played cards")
	assert.Equal(t, 2, playerAfter.Resources.Credits, "Should spend 18 MC (20 - 18 = 2)")

	t.Log("✅ Lava Flows temperature increase test passed")
}

func TestLavaFlows_TemperatureClampedAtMax(t *testing.T) {
	ctx, gameID, playerID, cardService, playerRepo, gameRepo := setupCardTest(t)

	// Set temperature to 7°C (1 step below max of 8°C)
	gameRepo.UpdateGlobalParameters(ctx, gameID, model.GlobalParameters{
		Temperature: 7, Oxygen: 0, Oceans: 0,
	})

	// Give player enough credits
	playerRepo.UpdateResources(ctx, gameID, playerID, model.Resources{Credits: 20})

	// Add Lava Flows to player's hand
	cardID := "140"
	playerRepo.AddCard(ctx, gameID, playerID, cardID)

	// Get state before
	gameBefore, _ := gameRepo.GetByID(ctx, gameID)
	assert.Equal(t, 7, gameBefore.GlobalParameters.Temperature, "Initial temperature should be 7°C")

	// Play Lava Flows (would increase by 2, but max is 8)
	err := cardService.OnPlayCard(ctx, gameID, playerID, cardID, nil, nil)
	require.NoError(t, err, "Should successfully play Lava Flows")

	// Verify temperature is clamped at max
	gameAfter, _ := gameRepo.GetByID(ctx, gameID)
	assert.Equal(t, 8, gameAfter.GlobalParameters.Temperature, "Temperature should be clamped at max (8°C)")

	t.Log("✅ Lava Flows temperature clamping test passed")
}
