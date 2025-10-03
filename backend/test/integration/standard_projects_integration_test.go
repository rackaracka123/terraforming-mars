package integration

import (
	"testing"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"

	"github.com/stretchr/testify/require"
)

// TestSellPatents_Integration tests the complete sell patents flow via WebSocket
func TestSellPatents_Integration(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	t.Log("✅ Test setup complete")

	// Start the game to get to action phase
	err := client.StartGame()
	require.NoError(t, err, "Failed to start game")
	t.Log("✅ Game started")

	// Wait for game to become active
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game update")
	t.Log("✅ Received game update after start")

	// Clear any pending messages
	client.ClearMessageQueue()

	// Get the starting cards dealt to the player (10 cards from game start)
	payload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok, "Game update payload should be a map")

	gameData, ok := payload["game"].(map[string]interface{})
	require.True(t, ok, "Game data should be present in payload")

	currentPlayer, ok := gameData["currentPlayer"].(map[string]interface{})
	require.True(t, ok, "Current player should be present")

	// Get starting cards and credits
	startingCardsInterface, ok := currentPlayer["startingCards"].([]interface{})
	require.True(t, ok, "Starting cards should be present")
	startingCardCount := len(startingCardsInterface)
	require.Greater(t, startingCardCount, 0, "Player should have starting cards")

	initialResources, ok := currentPlayer["resources"].(map[string]interface{})
	require.True(t, ok, "Resources should be present")
	initialCredits := int(initialResources["credits"].(float64))

	t.Logf("📊 Player has %d starting cards and %d credits", startingCardCount, initialCredits)

	// Execute sell patents action - sell 2 starting cards
	cardsToSell := 2
	if startingCardCount < cardsToSell {
		cardsToSell = startingCardCount
	}

	sellPatentsPayload := map[string]interface{}{
		"type":      dto.ActionTypeSellPatents,
		"cardCount": cardsToSell,
	}
	err = client.SendAction(dto.MessageTypeActionSellPatents, sellPatentsPayload)
	require.NoError(t, err, "Failed to send sell patents action")
	t.Log("✅ Sent sell patents action")

	// Wait for game update message
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game update after sell patents")
	t.Log("✅ Received game update after sell patents")

	// Verify the game state
	payload, ok = message.Payload.(map[string]interface{})
	require.True(t, ok, "Game update payload should be a map")

	gameData, ok = payload["game"].(map[string]interface{})
	require.True(t, ok, "Game data should be present in payload")

	currentPlayer, ok = gameData["currentPlayer"].(map[string]interface{})
	require.True(t, ok, "Current player should be present")

	resources, ok := currentPlayer["resources"].(map[string]interface{})
	require.True(t, ok, "Resources should be present")

	// Player should have gained credits (1 per card sold)
	credits := int(resources["credits"].(float64))
	expectedCredits := initialCredits + cardsToSell
	require.Equal(t, expectedCredits, credits, "Credits should increase by number of cards sold")

	// Player should have fewer starting cards
	newStartingCardsInterface, ok := currentPlayer["startingCards"].([]interface{})
	require.True(t, ok, "Starting cards should be present")
	newCardCount := len(newStartingCardsInterface)
	require.Equal(t, startingCardCount-cardsToSell, newCardCount, "Card count should decrease")

	t.Log("✅ Sell patents test passed!")
	t.Logf("   Credits: %d -> %d (+%d), Cards: %d -> %d", initialCredits, credits, cardsToSell, startingCardCount, newCardCount)
}

// TestBuildPowerPlant_Integration tests the complete build power plant flow via WebSocket
func TestBuildPowerPlant_Integration(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game
	err := client.StartGame()
	require.NoError(t, err)
	client.WaitForMessage(dto.MessageTypeGameUpdated)

	// Give player sufficient credits
	setResourcesPayload := map[string]interface{}{
		"playerId": client.playerID,
		"resources": map[string]interface{}{
			"credits":  50,
			"steel":    0,
			"titanium": 0,
			"plants":   0,
			"energy":   0,
			"heat":     0,
		},
	}
	err = client.SendAdminCommand(dto.AdminCommandTypeSetResources, setResourcesPayload)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	client.ClearMessageQueue()

	// Execute build power plant action (costs 11 MC)
	buildPowerPlantPayload := map[string]interface{}{
		"type": dto.ActionTypeBuildPowerPlant,
	}
	err = client.SendAction(dto.MessageTypeActionBuildPowerPlant, buildPowerPlantPayload)
	require.NoError(t, err, "Failed to send build power plant action")
	t.Log("✅ Sent build power plant action")

	// Wait for game update
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game update")

	// Verify game state
	payload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok)

	gameData, ok := payload["game"].(map[string]interface{})
	require.True(t, ok)

	currentPlayer, ok := gameData["currentPlayer"].(map[string]interface{})
	require.True(t, ok)

	resources, ok := currentPlayer["resources"].(map[string]interface{})
	require.True(t, ok)

	production, ok := currentPlayer["production"].(map[string]interface{})
	require.True(t, ok)

	credits := int(resources["credits"].(float64))
	energyProduction := int(production["energy"].(float64))

	// Verify credits deducted (50 - 11 = 39)
	require.Equal(t, 39, credits, "Credits should be deducted by 11")

	// Verify energy production increased by 1
	require.Equal(t, 1, energyProduction, "Energy production should increase by 1")

	t.Log("✅ Build power plant test passed!")
	t.Logf("   Credits: %d (deducted 11), Energy production: %d", credits, energyProduction)
}

// TestLaunchAsteroid_Integration tests the complete launch asteroid flow via WebSocket
func TestLaunchAsteroid_Integration(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game
	err := client.StartGame()
	require.NoError(t, err)
	client.WaitForMessage(dto.MessageTypeGameUpdated)

	// Give player sufficient credits
	setResourcesPayload := map[string]interface{}{
		"playerId": client.playerID,
		"resources": map[string]interface{}{
			"credits":  50,
			"steel":    0,
			"titanium": 0,
			"plants":   0,
			"energy":   0,
			"heat":     0,
		},
	}
	err = client.SendAdminCommand(dto.AdminCommandTypeSetResources, setResourcesPayload)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	client.ClearMessageQueue()

	// Get initial global parameters
	initialTemp := -30 // Default starting temperature

	// Execute launch asteroid action (costs 14 MC, increases temp by 2°C)
	launchAsteroidPayload := map[string]interface{}{
		"type": dto.ActionTypeLaunchAsteroid,
	}
	err = client.SendAction(dto.MessageTypeActionLaunchAsteroid, launchAsteroidPayload)
	require.NoError(t, err, "Failed to send launch asteroid action")
	t.Log("✅ Sent launch asteroid action")

	// Wait for game update
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game update")

	// Verify game state
	payload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok)

	gameData, ok := payload["game"].(map[string]interface{})
	require.True(t, ok)

	currentPlayer, ok := gameData["currentPlayer"].(map[string]interface{})
	require.True(t, ok)

	resources, ok := currentPlayer["resources"].(map[string]interface{})
	require.True(t, ok)

	globalParams, ok := gameData["globalParameters"].(map[string]interface{})
	require.True(t, ok)

	credits := int(resources["credits"].(float64))
	temperature := int(globalParams["temperature"].(float64))
	terraformRating := int(currentPlayer["terraformRating"].(float64))

	// Verify credits deducted (50 - 14 = 36)
	require.Equal(t, 36, credits, "Credits should be deducted by 14")

	// Verify temperature increased by 2°C
	require.Equal(t, initialTemp+2, temperature, "Temperature should increase by 2°C")

	// Verify TR increased by 1
	require.Equal(t, 21, terraformRating, "TR should increase by 1")

	t.Log("✅ Launch asteroid test passed!")
	t.Logf("   Credits: %d (deducted 14), Temperature: %d°C, TR: %d", credits, temperature, terraformRating)
}

// TestBuildAquifer_Integration tests the complete build aquifer flow via WebSocket
func TestBuildAquifer_Integration(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game
	err := client.StartGame()
	require.NoError(t, err)
	client.WaitForMessage(dto.MessageTypeGameUpdated)

	// Give player sufficient credits
	setResourcesPayload := map[string]interface{}{
		"playerId": client.playerID,
		"resources": map[string]interface{}{
			"credits":  50,
			"steel":    0,
			"titanium": 0,
			"plants":   0,
			"energy":   0,
			"heat":     0,
		},
	}
	err = client.SendAdminCommand(dto.AdminCommandTypeSetResources, setResourcesPayload)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	client.ClearMessageQueue()

	// Execute build aquifer action with valid hex position (costs 18 MC)
	buildAquiferPayload := map[string]interface{}{
		"type": dto.ActionTypeBuildAquifer,
		"hexPosition": map[string]interface{}{
			"q": 1,
			"r": -1,
			"s": 0, // Valid: 1 + (-1) + 0 = 0
		},
	}
	err = client.SendAction(dto.MessageTypeActionBuildAquifer, buildAquiferPayload)
	require.NoError(t, err, "Failed to send build aquifer action")
	t.Log("✅ Sent build aquifer action")

	// Wait for game update
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game update")

	// Verify game state
	payload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok)

	gameData, ok := payload["game"].(map[string]interface{})
	require.True(t, ok)

	currentPlayer, ok := gameData["currentPlayer"].(map[string]interface{})
	require.True(t, ok)

	resources, ok := currentPlayer["resources"].(map[string]interface{})
	require.True(t, ok)

	globalParams, ok := gameData["globalParameters"].(map[string]interface{})
	require.True(t, ok)

	credits := int(resources["credits"].(float64))
	oceans := int(globalParams["oceans"].(float64))
	terraformRating := int(currentPlayer["terraformRating"].(float64))

	// Verify credits deducted (50 - 18 = 32)
	require.Equal(t, 32, credits, "Credits should be deducted by 18")

	// Verify ocean count increased by 1
	require.Equal(t, 1, oceans, "Ocean count should increase by 1")

	// Verify TR increased by 1
	require.Equal(t, 21, terraformRating, "TR should increase by 1")

	t.Log("✅ Build aquifer test passed!")
	t.Logf("   Credits: %d (deducted 18), Oceans: %d, TR: %d", credits, oceans, terraformRating)
}

// TestBuildAquifer_InvalidHexPosition tests hex position validation
func TestBuildAquifer_InvalidHexPosition(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game
	err := client.StartGame()
	require.NoError(t, err)
	client.WaitForMessage(dto.MessageTypeGameUpdated)

	// Give player sufficient credits
	setResourcesPayload := map[string]interface{}{
		"playerId": client.playerID,
		"resources": map[string]interface{}{
			"credits":  50,
			"steel":    0,
			"titanium": 0,
			"plants":   0,
			"energy":   0,
			"heat":     0,
		},
	}
	err = client.SendAdminCommand(dto.AdminCommandTypeSetResources, setResourcesPayload)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	client.ClearMessageQueue()

	// Execute build aquifer action with INVALID hex position
	buildAquiferPayload := map[string]interface{}{
		"type": dto.ActionTypeBuildAquifer,
		"hexPosition": map[string]interface{}{
			"q": 1,
			"r": 1,
			"s": 1, // Invalid: 1 + 1 + 1 = 3 (should be 0)
		},
	}
	err = client.SendAction(dto.MessageTypeActionBuildAquifer, buildAquiferPayload)
	require.NoError(t, err, "Failed to send build aquifer action")
	t.Log("✅ Sent build aquifer action with invalid hex position")

	// Wait for error message
	message, err := client.WaitForMessageWithTimeout(dto.MessageTypeError, 2*time.Second)
	require.NoError(t, err, "Should receive error message for invalid hex position")
	require.NotNil(t, message, "Error message should not be nil")

	t.Log("✅ Build aquifer invalid hex position test passed!")
	t.Log("   Correctly rejected invalid hex coordinates")
}

// TestPlantGreenery_Integration tests the complete plant greenery flow via WebSocket
func TestPlantGreenery_Integration(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game
	err := client.StartGame()
	require.NoError(t, err)
	client.WaitForMessage(dto.MessageTypeGameUpdated)

	// Give player sufficient credits
	setResourcesPayload := map[string]interface{}{
		"playerId": client.playerID,
		"resources": map[string]interface{}{
			"credits":  50,
			"steel":    0,
			"titanium": 0,
			"plants":   0,
			"energy":   0,
			"heat":     0,
		},
	}
	err = client.SendAdminCommand(dto.AdminCommandTypeSetResources, setResourcesPayload)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	client.ClearMessageQueue()

	// Execute plant greenery action (costs 23 MC)
	plantGreeneryPayload := map[string]interface{}{
		"type": dto.ActionTypePlantGreenery,
		"hexPosition": map[string]interface{}{
			"q": 2,
			"r": -1,
			"s": -1, // Valid: 2 + (-1) + (-1) = 0
		},
	}
	err = client.SendAction(dto.MessageTypeActionPlantGreenery, plantGreeneryPayload)
	require.NoError(t, err, "Failed to send plant greenery action")
	t.Log("✅ Sent plant greenery action")

	// Wait for game update
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game update")

	// Verify game state
	payload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok)

	gameData, ok := payload["game"].(map[string]interface{})
	require.True(t, ok)

	currentPlayer, ok := gameData["currentPlayer"].(map[string]interface{})
	require.True(t, ok)

	resources, ok := currentPlayer["resources"].(map[string]interface{})
	require.True(t, ok)

	globalParams, ok := gameData["globalParameters"].(map[string]interface{})
	require.True(t, ok)

	credits := int(resources["credits"].(float64))
	oxygen := int(globalParams["oxygen"].(float64))
	terraformRating := int(currentPlayer["terraformRating"].(float64))

	// Verify credits deducted (50 - 23 = 27)
	require.Equal(t, 27, credits, "Credits should be deducted by 23")

	// Verify oxygen increased by 1%
	require.Equal(t, 1, oxygen, "Oxygen should increase by 1%")

	// Verify TR increased by 1
	require.Equal(t, 21, terraformRating, "TR should increase by 1")

	t.Log("✅ Plant greenery test passed!")
	t.Logf("   Credits: %d (deducted 23), Oxygen: %d%%, TR: %d", credits, oxygen, terraformRating)
}

// TestBuildCity_Integration tests the complete build city flow via WebSocket
func TestBuildCity_Integration(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game
	err := client.StartGame()
	require.NoError(t, err)
	client.WaitForMessage(dto.MessageTypeGameUpdated)

	// Give player sufficient credits
	setResourcesPayload := map[string]interface{}{
		"playerId": client.playerID,
		"resources": map[string]interface{}{
			"credits":  50,
			"steel":    0,
			"titanium": 0,
			"plants":   0,
			"energy":   0,
			"heat":     0,
		},
	}
	err = client.SendAdminCommand(dto.AdminCommandTypeSetResources, setResourcesPayload)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	client.ClearMessageQueue()

	// Execute build city action (costs 25 MC)
	buildCityPayload := map[string]interface{}{
		"type": dto.ActionTypeBuildCity,
		"hexPosition": map[string]interface{}{
			"q": -2,
			"r": 1,
			"s": 1, // Valid: -2 + 1 + 1 = 0
		},
	}
	err = client.SendAction(dto.MessageTypeActionBuildCity, buildCityPayload)
	require.NoError(t, err, "Failed to send build city action")
	t.Log("✅ Sent build city action")

	// Wait for game update
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game update")

	// Verify game state
	payload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok)

	gameData, ok := payload["game"].(map[string]interface{})
	require.True(t, ok)

	currentPlayer, ok := gameData["currentPlayer"].(map[string]interface{})
	require.True(t, ok)

	resources, ok := currentPlayer["resources"].(map[string]interface{})
	require.True(t, ok)

	production, ok := currentPlayer["production"].(map[string]interface{})
	require.True(t, ok)

	credits := int(resources["credits"].(float64))
	creditProduction := int(production["credits"].(float64))

	// Verify credits deducted (50 - 25 = 25)
	require.Equal(t, 25, credits, "Credits should be deducted by 25")

	// Verify credit production increased by 1
	require.Equal(t, 2, creditProduction, "Credit production should increase by 1 (base 1 + city 1)")

	t.Log("✅ Build city test passed!")
	t.Logf("   Credits: %d (deducted 25), Credit production: %d", credits, creditProduction)
}

// TestMultiPlayerStandardProjects tests multiple players executing standard projects
func TestMultiPlayerStandardProjects(t *testing.T) {
	// Setup two clients
	client1 := NewTestClient(t)
	defer client1.Close()

	client2 := NewTestClient(t)
	defer client2.Close()

	// Connect both clients
	err := client1.Connect()
	require.NoError(t, err, "Client 1 failed to connect")

	err = client2.Connect()
	require.NoError(t, err, "Client 2 failed to connect")

	// Client 1 creates game
	gameID, err := client1.CreateGameViaHTTP()
	require.NoError(t, err, "Failed to create game")

	// Client 1 joins
	err = client1.JoinGameViaWebSocket(gameID, "Player1")
	require.NoError(t, err, "Client 1 failed to join")

	message1, err := client1.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Client 1 failed to receive game update")

	// Extract player 1 ID
	payload1, _ := message1.Payload.(map[string]interface{})
	gameData1, _ := payload1["game"].(map[string]interface{})
	currentPlayer1, _ := gameData1["currentPlayer"].(map[string]interface{})
	player1ID := currentPlayer1["id"].(string)
	client1.SetPlayerID(player1ID)

	// Client 2 joins
	err = client2.JoinGameViaWebSocket(gameID, "Player2")
	require.NoError(t, err, "Client 2 failed to join")

	// Both clients should receive game updates
	client1.WaitForMessage(dto.MessageTypeGameUpdated)
	message2, err := client2.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Client 2 failed to receive game update")

	// Extract player 2 ID
	payload2, _ := message2.Payload.(map[string]interface{})
	gameData2, _ := payload2["game"].(map[string]interface{})
	currentPlayer2, _ := gameData2["currentPlayer"].(map[string]interface{})
	player2ID := currentPlayer2["id"].(string)
	client2.SetPlayerID(player2ID)

	t.Logf("✅ Two players joined: Player1=%s, Player2=%s", player1ID, player2ID)

	// Start the game (client 1 is host)
	err = client1.StartGame()
	require.NoError(t, err)

	// Both clients receive game updates
	client1.WaitForMessage(dto.MessageTypeGameUpdated)
	client2.WaitForMessage(dto.MessageTypeGameUpdated)

	t.Log("✅ Game started")

	// Give both players resources
	setResourcesPayload1 := map[string]interface{}{
		"playerId": player1ID,
		"resources": map[string]interface{}{
			"credits":  50,
			"steel":    0,
			"titanium": 0,
			"plants":   0,
			"energy":   0,
			"heat":     0,
		},
	}
	client1.SendAdminCommand(dto.AdminCommandTypeSetResources, setResourcesPayload1)

	setResourcesPayload2 := map[string]interface{}{
		"playerId": player2ID,
		"resources": map[string]interface{}{
			"credits":  50,
			"steel":    0,
			"titanium": 0,
			"plants":   0,
			"energy":   0,
			"heat":     0,
		},
	}
	client1.SendAdminCommand(dto.AdminCommandTypeSetResources, setResourcesPayload2)

	time.Sleep(200 * time.Millisecond)
	client1.ClearMessageQueue()
	client2.ClearMessageQueue()

	// Player 1 builds power plant
	buildPowerPlantPayload := map[string]interface{}{
		"type": dto.ActionTypeBuildPowerPlant,
	}
	err = client1.SendAction(dto.MessageTypeActionBuildPowerPlant, buildPowerPlantPayload)
	require.NoError(t, err)
	t.Log("✅ Player 1 executed build power plant")

	// Both clients should receive the update
	msg1, err := client1.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Client 1 should receive broadcast")

	msg2, err := client2.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Client 2 should receive broadcast")

	t.Log("✅ Both clients received broadcast after Player 1 action")

	// Verify both received the same update
	payload1, _ = msg1.Payload.(map[string]interface{})
	payload2, _ = msg2.Payload.(map[string]interface{})

	gameData1, _ = payload1["game"].(map[string]interface{})
	gameData2, _ = payload2["game"].(map[string]interface{})

	// Compare game IDs to verify same game state
	gameID1 := gameData1["id"].(string)
	gameID2 := gameData2["id"].(string)
	require.Equal(t, gameID1, gameID2, "Both clients should see the same game")

	t.Log("✅ Multi-player standard project test passed!")
	t.Log("   Both players received synchronized game state updates")
}

// TestStandardProjectsInsufficientFunds tests error handling for insufficient funds
func TestStandardProjectsInsufficientFunds(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game
	err := client.StartGame()
	require.NoError(t, err)
	client.WaitForMessage(dto.MessageTypeGameUpdated)

	// Give player only 5 credits (not enough for any standard project)
	setResourcesPayload := map[string]interface{}{
		"playerId": client.playerID,
		"resources": map[string]interface{}{
			"credits":  5,
			"steel":    0,
			"titanium": 0,
			"plants":   0,
			"energy":   0,
			"heat":     0,
		},
	}
	err = client.SendAdminCommand(dto.AdminCommandTypeSetResources, setResourcesPayload)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	client.ClearMessageQueue()

	// Try to build power plant (costs 11 MC, player only has 5)
	buildPowerPlantPayload := map[string]interface{}{
		"type": dto.ActionTypeBuildPowerPlant,
	}
	err = client.SendAction(dto.MessageTypeActionBuildPowerPlant, buildPowerPlantPayload)
	require.NoError(t, err, "Failed to send build power plant action")
	t.Log("✅ Sent build power plant action with insufficient funds")

	// Wait for error message
	message, err := client.WaitForMessageWithTimeout(dto.MessageTypeError, 2*time.Second)
	require.NoError(t, err, "Should receive error message for insufficient funds")
	require.NotNil(t, message, "Error message should not be nil")

	// Verify error message mentions insufficient credits
	errorPayload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok, "Error payload should be a map")

	errorMsg, ok := errorPayload["error"].(string)
	require.True(t, ok, "Error message should be present")
	require.Contains(t, errorMsg, "insufficient credits", "Error should mention insufficient credits")

	t.Log("✅ Insufficient funds error handling test passed!")
	t.Logf("   Error message: %s", errorMsg)
}

// TestGlobalParameterLimits tests that global parameters don't exceed maximum values
func TestGlobalParameterLimits(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game
	err := client.StartGame()
	require.NoError(t, err)
	client.WaitForMessage(dto.MessageTypeGameUpdated)

	// Set global parameters to near maximum
	setGlobalParamsPayload := map[string]interface{}{
		"globalParameters": map[string]interface{}{
			"temperature": 6, // Max is 8
			"oxygen":      0,
			"oceans":      0,
		},
	}
	err = client.SendAdminCommand(dto.AdminCommandTypeSetGlobalParams, setGlobalParamsPayload)
	require.NoError(t, err)

	// Give player sufficient credits
	setResourcesPayload := map[string]interface{}{
		"playerId": client.playerID,
		"resources": map[string]interface{}{
			"credits":  50,
			"steel":    0,
			"titanium": 0,
			"plants":   0,
			"energy":   0,
			"heat":     0,
		},
	}
	err = client.SendAdminCommand(dto.AdminCommandTypeSetResources, setResourcesPayload)
	require.NoError(t, err)

	time.Sleep(200 * time.Millisecond)
	client.ClearMessageQueue()

	// Launch asteroid (increases temp by 2, should cap at 8)
	launchAsteroidPayload := map[string]interface{}{
		"type": dto.ActionTypeLaunchAsteroid,
	}
	err = client.SendAction(dto.MessageTypeActionLaunchAsteroid, launchAsteroidPayload)
	require.NoError(t, err)

	// Wait for game update
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)

	// Verify temperature capped at maximum
	payload, _ := message.Payload.(map[string]interface{})
	gameData, _ := payload["game"].(map[string]interface{})
	globalParams, _ := gameData["globalParameters"].(map[string]interface{})
	temperature := int(globalParams["temperature"].(float64))

	// Temperature should be capped at 8 (not exceed it)
	require.LessOrEqual(t, temperature, 8, "Temperature should not exceed maximum of 8")

	t.Log("✅ Global parameter limits test passed!")
	t.Logf("   Temperature capped at: %d°C (max: 8°C)", temperature)
}
