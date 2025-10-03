package integration

import (
	"fmt"
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

	// Select starting cards (keep all 10 cards by buying them for 30 MC total)
	// Extract card IDs from starting cards
	cardIDs := make([]string, 0, startingCardCount)
	for _, cardInterface := range startingCardsInterface {
		card, ok := cardInterface.(map[string]interface{})
		if ok {
			cardID, ok := card["id"].(string)
			if ok {
				cardIDs = append(cardIDs, cardID)
			}
		}
	}

	selectStartingCardsPayload := map[string]interface{}{
		"cardIds": cardIDs, // Select all cards
	}
	err = client.SendRawMessage(dto.MessageTypeActionSelectStartingCard, selectStartingCardsPayload)
	require.NoError(t, err, "Failed to send select starting cards")
	t.Log("✅ Selected starting cards")

	// Wait for game update after card selection
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game update after card selection")
	client.ClearMessageQueue()

	// Now the cards should be in the player's hand
	// Extract current state
	payload, ok = message.Payload.(map[string]interface{})
	require.True(t, ok)
	gameData, ok = payload["game"].(map[string]interface{})
	require.True(t, ok)
	currentPlayer, ok = gameData["currentPlayer"].(map[string]interface{})
	require.True(t, ok)

	// Get player's cards in hand
	cardsInHand, ok := currentPlayer["cards"].([]interface{})
	require.True(t, ok, "Cards should be present")
	require.Greater(t, len(cardsInHand), 2, "Player should have at least 3 cards")

	resources, ok := currentPlayer["resources"].(map[string]interface{})
	require.True(t, ok)
	creditsBeforeSell := int(resources["credits"].(float64))

	t.Logf("📊 Player has %d cards in hand and %d credits before sell", len(cardsInHand), creditsBeforeSell)

	// Step 1: Initiate sell patents (no cardCount parameter in new flow)
	sellPatentsPayload := map[string]interface{}{
		"type": dto.ActionTypeSellPatents,
	}
	err = client.SendAction(dto.MessageTypeActionSellPatents, sellPatentsPayload)
	require.NoError(t, err, "Failed to send sell patents action")
	t.Log("✅ Sent sell patents initiation")

	// Step 2: Wait for game update with pending card selection
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game update after sell patents initiation")

	// Extract pending card selection
	payload, ok = message.Payload.(map[string]interface{})
	require.True(t, ok)
	gameData, ok = payload["game"].(map[string]interface{})
	require.True(t, ok)
	currentPlayer, ok = gameData["currentPlayer"].(map[string]interface{})
	require.True(t, ok)

	pendingSelection, ok := currentPlayer["pendingCardSelection"].(map[string]interface{})
	require.True(t, ok, "Pending card selection should be present")
	require.Equal(t, "sell-patents", pendingSelection["source"].(string), "Source should be sell-patents")

	availableCards, ok := pendingSelection["availableCards"].([]interface{})
	require.True(t, ok)
	require.Greater(t, len(availableCards), 0, "Should have available cards")

	t.Logf("📊 Pending card selection created with %d available cards", len(availableCards))

	// Step 3: Select 3 cards to sell
	cardsToSell := 3
	selectedCardIDs := make([]string, 0, cardsToSell)
	for i := 0; i < cardsToSell && i < len(availableCards); i++ {
		cardObj, ok := availableCards[i].(map[string]interface{})
		require.True(t, ok, "Failed to assert card as map. Type: %T", availableCards[i])
		cardID, ok := cardObj["id"].(string)
		require.True(t, ok, "Failed to get card ID. Type: %T", cardObj["id"])
		selectedCardIDs = append(selectedCardIDs, cardID)
	}

	selectCardsPayload := map[string]interface{}{
		"cardIds": selectedCardIDs,
	}
	err = client.SendRawMessage(dto.MessageTypeActionSelectCards, selectCardsPayload)
	require.NoError(t, err, "Failed to send card selection")
	t.Logf("✅ Sent card selection: %d cards", len(selectedCardIDs))

	// Step 4: Wait for final game update
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game update after card selection")

	// Verify final state
	payload, ok = message.Payload.(map[string]interface{})
	require.True(t, ok)
	gameData, ok = payload["game"].(map[string]interface{})
	require.True(t, ok)
	currentPlayer, ok = gameData["currentPlayer"].(map[string]interface{})
	require.True(t, ok)

	// Verify credits increased by number of cards sold (1 MC per card)
	resources, ok = currentPlayer["resources"].(map[string]interface{})
	require.True(t, ok)
	creditsAfterSell := int(resources["credits"].(float64))
	require.Equal(t, creditsBeforeSell+cardsToSell, creditsAfterSell, "Credits should increase by 1 per card sold")

	// Verify cards removed from hand
	cardsAfterSell, ok := currentPlayer["cards"].([]interface{})
	require.True(t, ok)
	require.Equal(t, len(cardsInHand)-cardsToSell, len(cardsAfterSell), "Cards should be removed from hand")

	// Verify pending card selection cleared
	pendingAfter, exists := currentPlayer["pendingCardSelection"]
	require.True(t, !exists || pendingAfter == nil, "Pending card selection should be cleared")

	t.Log("✅ Sell patents test passed!")
	t.Logf("   Sold %d cards, gained %d MC (from %d to %d MC)", cardsToSell, cardsToSell, creditsBeforeSell, creditsAfterSell)
	t.Logf("   Cards in hand: %d → %d", len(cardsInHand), len(cardsAfterSell))
}

// TestSellPatents_SelectZeroCards tests selling zero cards (allowed by min=0)
func TestSellPatents_SelectZeroCards(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game
	err := client.StartGame()
	require.NoError(t, err)
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)
	client.ClearMessageQueue()

	// Select starting cards to have cards in hand
	payload, _ := message.Payload.(map[string]interface{})
	gameData, _ := payload["game"].(map[string]interface{})
	currentPlayer, _ := gameData["currentPlayer"].(map[string]interface{})
	startingCards, _ := currentPlayer["startingCards"].([]interface{})

	cardIDs := make([]string, 0)
	for _, cardInterface := range startingCards {
		card, ok := cardInterface.(map[string]interface{})
		if ok {
			if cardID, ok := card["id"].(string); ok {
				cardIDs = append(cardIDs, cardID)
			}
		}
	}

	err = client.SendRawMessage(dto.MessageTypeActionSelectStartingCard, map[string]interface{}{"cardIds": cardIDs})
	require.NoError(t, err)
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)
	client.ClearMessageQueue()

	// Get initial state
	payload, _ = message.Payload.(map[string]interface{})
	gameData, _ = payload["game"].(map[string]interface{})
	currentPlayer, _ = gameData["currentPlayer"].(map[string]interface{})

	resources, _ := currentPlayer["resources"].(map[string]interface{})
	initialCredits := int(resources["credits"].(float64))
	cardsInHand, _ := currentPlayer["cards"].([]interface{})
	initialCardCount := len(cardsInHand)

	// Initiate sell patents
	err = client.SendAction(dto.MessageTypeActionSellPatents, map[string]interface{}{"type": dto.ActionTypeSellPatents})
	require.NoError(t, err)

	// Wait for pending card selection
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)

	// Select ZERO cards
	err = client.SendRawMessage(dto.MessageTypeActionSelectCards, map[string]interface{}{"cardIds": []string{}})
	require.NoError(t, err)
	t.Log("✅ Sent selection with zero cards")

	// Wait for final state
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)

	// Verify nothing changed
	payload, _ = message.Payload.(map[string]interface{})
	gameData, _ = payload["game"].(map[string]interface{})
	currentPlayer, _ = gameData["currentPlayer"].(map[string]interface{})

	resources, _ = currentPlayer["resources"].(map[string]interface{})
	finalCredits := int(resources["credits"].(float64))
	cardsAfter, _ := currentPlayer["cards"].([]interface{})

	require.Equal(t, initialCredits, finalCredits, "Credits should not change when selling zero cards")
	require.Equal(t, initialCardCount, len(cardsAfter), "Card count should not change")

	// Verify pending selection cleared
	pendingAfter, exists := currentPlayer["pendingCardSelection"]
	require.True(t, !exists || pendingAfter == nil, "Pending card selection should be cleared")

	t.Log("✅ Select zero cards test passed!")
	t.Logf("   Credits unchanged: %d MC", finalCredits)
}

// TestSellPatents_SelectAllCards tests selling all cards in hand
func TestSellPatents_SelectAllCards(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game
	err := client.StartGame()
	require.NoError(t, err)
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)
	client.ClearMessageQueue()

	// Select starting cards
	payload, _ := message.Payload.(map[string]interface{})
	gameData, _ := payload["game"].(map[string]interface{})
	currentPlayer, _ := gameData["currentPlayer"].(map[string]interface{})
	startingCards, _ := currentPlayer["startingCards"].([]interface{})

	cardIDs := make([]string, 0)
	for _, cardInterface := range startingCards {
		card, ok := cardInterface.(map[string]interface{})
		if ok {
			if cardID, ok := card["id"].(string); ok {
				cardIDs = append(cardIDs, cardID)
			}
		}
	}

	err = client.SendRawMessage(dto.MessageTypeActionSelectStartingCard, map[string]interface{}{"cardIds": cardIDs})
	require.NoError(t, err)
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)
	client.ClearMessageQueue()

	// Get initial state
	payload, _ = message.Payload.(map[string]interface{})
	gameData, _ = payload["game"].(map[string]interface{})
	currentPlayer, _ = gameData["currentPlayer"].(map[string]interface{})

	resources, _ := currentPlayer["resources"].(map[string]interface{})
	initialCredits := int(resources["credits"].(float64))
	cardsInHand, _ := currentPlayer["cards"].([]interface{})
	initialCardCount := len(cardsInHand)
	require.Greater(t, initialCardCount, 0, "Player should have cards")

	// Initiate sell patents
	err = client.SendAction(dto.MessageTypeActionSellPatents, map[string]interface{}{"type": dto.ActionTypeSellPatents})
	require.NoError(t, err)

	// Wait for pending card selection
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)

	// Extract available cards
	payload, _ = message.Payload.(map[string]interface{})
	gameData, _ = payload["game"].(map[string]interface{})
	currentPlayer, _ = gameData["currentPlayer"].(map[string]interface{})
	pendingSelection, _ := currentPlayer["pendingCardSelection"].(map[string]interface{})
	availableCards, _ := pendingSelection["availableCards"].([]interface{})

	// Select ALL cards
	allCardIDs := make([]string, 0, len(availableCards))
	for _, card := range availableCards {
		cardObj := card.(map[string]interface{})
		cardID := cardObj["id"].(string)
		allCardIDs = append(allCardIDs, cardID)
	}

	err = client.SendRawMessage(dto.MessageTypeActionSelectCards, map[string]interface{}{"cardIds": allCardIDs})
	require.NoError(t, err)
	t.Logf("✅ Sent selection with all %d cards", len(allCardIDs))

	// Wait for final state
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)

	// Verify all cards sold
	payload, _ = message.Payload.(map[string]interface{})
	gameData, _ = payload["game"].(map[string]interface{})
	currentPlayer, _ = gameData["currentPlayer"].(map[string]interface{})

	resources, _ = currentPlayer["resources"].(map[string]interface{})
	finalCredits := int(resources["credits"].(float64))
	cardsAfter, _ := currentPlayer["cards"].([]interface{})

	require.Equal(t, initialCredits+initialCardCount, finalCredits, "Should gain 1 MC per card")
	require.Equal(t, 0, len(cardsAfter), "Hand should be empty after selling all cards")

	t.Log("✅ Select all cards test passed!")
	t.Logf("   Sold all %d cards, gained %d MC (from %d to %d MC)", initialCardCount, initialCardCount, initialCredits, finalCredits)
}

// TestSellPatents_InvalidSelection tests error handling for invalid card selections
func TestSellPatents_InvalidSelection(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game and select starting cards
	err := client.StartGame()
	require.NoError(t, err)
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)
	client.ClearMessageQueue()

	payload, _ := message.Payload.(map[string]interface{})
	gameData, _ := payload["game"].(map[string]interface{})
	currentPlayer, _ := gameData["currentPlayer"].(map[string]interface{})
	startingCards, _ := currentPlayer["startingCards"].([]interface{})

	cardIDs := make([]string, 0)
	for _, cardInterface := range startingCards {
		card, ok := cardInterface.(map[string]interface{})
		if ok {
			if cardID, ok := card["id"].(string); ok {
				cardIDs = append(cardIDs, cardID)
			}
		}
	}

	err = client.SendRawMessage(dto.MessageTypeActionSelectStartingCard, map[string]interface{}{"cardIds": cardIDs})
	require.NoError(t, err)
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)
	client.ClearMessageQueue()

	// Initiate sell patents
	err = client.SendAction(dto.MessageTypeActionSellPatents, map[string]interface{}{"type": dto.ActionTypeSellPatents})
	require.NoError(t, err)

	// Wait for pending card selection
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)

	// Extract max cards allowed
	payload, _ = message.Payload.(map[string]interface{})
	gameData, _ = payload["game"].(map[string]interface{})
	currentPlayer, _ = gameData["currentPlayer"].(map[string]interface{})
	pendingSelection, _ := currentPlayer["pendingCardSelection"].(map[string]interface{})
	maxCards := int(pendingSelection["maxCards"].(float64))

	// Try to select MORE cards than allowed
	invalidCardIDs := make([]string, maxCards+10)
	for i := range invalidCardIDs {
		invalidCardIDs[i] = fmt.Sprintf("invalid-card-%d", i)
	}

	err = client.SendRawMessage(dto.MessageTypeActionSelectCards, map[string]interface{}{"cardIds": invalidCardIDs})
	require.NoError(t, err)
	t.Logf("✅ Sent invalid selection: %d cards (max: %d)", len(invalidCardIDs), maxCards)

	// Wait for error message
	errorMessage, err := client.WaitForMessageWithTimeout(dto.MessageTypeError, 2*time.Second)
	require.NoError(t, err, "Should receive error message")
	require.NotNil(t, errorMessage, "Error message should not be nil")

	t.Log("✅ Invalid selection test passed!")
	t.Log("   Backend correctly rejected invalid card selection")
}

// TestSellPatents_NoCardsInHand tests error when player has no cards to sell
func TestSellPatents_NoCardsInHand(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game
	err := client.StartGame()
	require.NoError(t, err)
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)
	client.ClearMessageQueue()

	// Select starting cards FIRST (don't buy any, so we keep all starting cards)
	payload, _ := message.Payload.(map[string]interface{})
	gameData, _ := payload["game"].(map[string]interface{})
	currentPlayer, _ := gameData["currentPlayer"].(map[string]interface{})

	// Select ZERO starting cards (decline all)
	err = client.SendRawMessage(dto.MessageTypeActionSelectStartingCard, map[string]interface{}{"cardIds": []string{}})
	require.NoError(t, err)
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)
	client.ClearMessageQueue()

	// Verify player has no cards in hand
	payload, _ = message.Payload.(map[string]interface{})
	gameData, _ = payload["game"].(map[string]interface{})
	currentPlayer, _ = gameData["currentPlayer"].(map[string]interface{})
	cardsInHand, _ := currentPlayer["cards"].([]interface{})
	require.Equal(t, 0, len(cardsInHand), "Player should have no cards")

	// Try to initiate sell patents
	err = client.SendAction(dto.MessageTypeActionSellPatents, map[string]interface{}{"type": dto.ActionTypeSellPatents})
	require.NoError(t, err)
	t.Log("✅ Sent sell patents with no cards in hand")

	// Wait for error message
	errorMessage, err := client.WaitForMessageWithTimeout(dto.MessageTypeError, 2*time.Second)
	require.NoError(t, err, "Should receive error message")
	require.NotNil(t, errorMessage, "Error message should not be nil")

	// Verify error mentions no cards
	errorPayload, ok := errorMessage.Payload.(map[string]interface{})
	require.True(t, ok)

	errorMsg := ""
	if msg, ok := errorPayload["message"].(string); ok {
		errorMsg = msg
	}
	require.Contains(t, errorMsg, "no cards", "Error should mention no cards")

	t.Log("✅ No cards in hand test passed!")
	t.Logf("   Error message: %s", errorMsg)
}

// TestSellPatents_MultipleSelectionPhases tests multiple sell patents in sequence
func TestSellPatents_MultipleSelectionPhases(t *testing.T) {
	client, _ := SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game and select starting cards
	err := client.StartGame()
	require.NoError(t, err)
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)
	client.ClearMessageQueue()

	payload, _ := message.Payload.(map[string]interface{})
	gameData, _ := payload["game"].(map[string]interface{})
	currentPlayer, _ := gameData["currentPlayer"].(map[string]interface{})
	startingCards, _ := currentPlayer["startingCards"].([]interface{})

	cardIDs := make([]string, 0)
	for _, cardInterface := range startingCards {
		card, ok := cardInterface.(map[string]interface{})
		if ok {
			if cardID, ok := card["id"].(string); ok {
				cardIDs = append(cardIDs, cardID)
			}
		}
	}

	err = client.SendRawMessage(dto.MessageTypeActionSelectStartingCard, map[string]interface{}{"cardIds": cardIDs})
	require.NoError(t, err)
	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)
	client.ClearMessageQueue()

	// FIRST sell patents - sell 2 cards
	t.Log("📊 First sell patents session")
	err = client.SendAction(dto.MessageTypeActionSellPatents, map[string]interface{}{"type": dto.ActionTypeSellPatents})
	require.NoError(t, err)

	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)

	payload, _ = message.Payload.(map[string]interface{})
	gameData, _ = payload["game"].(map[string]interface{})
	currentPlayer, _ = gameData["currentPlayer"].(map[string]interface{})
	pendingSelection, _ := currentPlayer["pendingCardSelection"].(map[string]interface{})
	availableCards, _ := pendingSelection["availableCards"].([]interface{})

	// Select 2 cards
	selectedCards := make([]string, 0, 2)
	for i := 0; i < 2 && i < len(availableCards); i++ {
		cardObj := availableCards[i].(map[string]interface{})
		cardID := cardObj["id"].(string)
		selectedCards = append(selectedCards, cardID)
	}

	err = client.SendRawMessage(dto.MessageTypeActionSelectCards, map[string]interface{}{"cardIds": selectedCards})
	require.NoError(t, err)

	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)
	client.ClearMessageQueue()

	// Verify first sell completed
	payload, _ = message.Payload.(map[string]interface{})
	gameData, _ = payload["game"].(map[string]interface{})
	currentPlayer, _ = gameData["currentPlayer"].(map[string]interface{})
	pendingAfterFirst, exists := currentPlayer["pendingCardSelection"]
	require.True(t, !exists || pendingAfterFirst == nil, "Pending selection should be cleared after first sell")
	t.Log("✅ First sell patents completed")

	// SECOND sell patents - sell 3 more cards
	t.Log("📊 Second sell patents session")
	err = client.SendAction(dto.MessageTypeActionSellPatents, map[string]interface{}{"type": dto.ActionTypeSellPatents})
	require.NoError(t, err)

	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)

	payload, _ = message.Payload.(map[string]interface{})
	gameData, _ = payload["game"].(map[string]interface{})
	currentPlayer, _ = gameData["currentPlayer"].(map[string]interface{})
	pendingSelection2, _ := currentPlayer["pendingCardSelection"].(map[string]interface{})
	availableCards2, _ := pendingSelection2["availableCards"].([]interface{})

	// Select 3 cards
	selectedCards2 := make([]string, 0, 3)
	for i := 0; i < 3 && i < len(availableCards2); i++ {
		cardObj := availableCards2[i].(map[string]interface{})
		cardID := cardObj["id"].(string)
		selectedCards2 = append(selectedCards2, cardID)
	}

	err = client.SendRawMessage(dto.MessageTypeActionSelectCards, map[string]interface{}{"cardIds": selectedCards2})
	require.NoError(t, err)

	message, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err)

	// Verify second sell completed
	payload, _ = message.Payload.(map[string]interface{})
	gameData, _ = payload["game"].(map[string]interface{})
	currentPlayer, _ = gameData["currentPlayer"].(map[string]interface{})
	pendingAfterSecond, exists := currentPlayer["pendingCardSelection"]
	require.True(t, !exists || pendingAfterSecond == nil, "Pending selection should be cleared after second sell")

	t.Log("✅ Multiple selection phases test passed!")
	t.Logf("   Successfully completed two separate sell patents sessions")
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

	// Execute build aquifer action (triggers tile placement)
	buildAquiferPayload := map[string]interface{}{
		"type": dto.ActionTypeBuildAquifer,
	}
	err = client.SendAction(dto.MessageTypeActionBuildAquifer, buildAquiferPayload)
	require.NoError(t, err, "Failed to send build aquifer action")
	t.Log("✅ Sent build aquifer action")

	// Wait for game update (tile placement should be queued)
	_, err = client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Should receive game update")
	client.ClearMessageQueue()

	// Now send INVALID hex coordinates for tile placement
	tileSelectedPayload := map[string]interface{}{
		"coordinate": map[string]interface{}{
			"q": 1,
			"r": 1,
			"s": 1, // Invalid: 1 + 1 + 1 = 3 (should be 0)
		},
	}
	err = client.SendRawMessage(dto.MessageType("tile-selected"), tileSelectedPayload)
	require.NoError(t, err, "Failed to send tile selected")
	t.Log("✅ Sent tile selection with invalid hex coordinates")

	// Wait for error message
	message, err := client.WaitForMessageWithTimeout(dto.MessageTypeError, 2*time.Second)
	require.NoError(t, err, "Should receive error message for invalid hex position")
	require.NotNil(t, message, "Error message should not be nil")

	t.Log("✅ Build aquifer invalid hex position test passed!")
	t.Log("   Correctly rejected invalid hex coordinates in tile-selected phase")
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

	// Try both "error" and "message" fields for compatibility
	errorMsg := ""
	if msg, ok := errorPayload["message"].(string); ok {
		errorMsg = msg
	} else if msg, ok := errorPayload["error"].(string); ok {
		errorMsg = msg
	}
	require.NotEmpty(t, errorMsg, "Error message should be present")
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
