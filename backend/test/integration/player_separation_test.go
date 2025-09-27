package integration

import (
	"testing"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"

	"github.com/stretchr/testify/require"
)

// hasMultiplePlayers checks if a game-updated message contains multiple players
func hasMultiplePlayers(msg *dto.WebSocketMessage) bool {
	if msg == nil || msg.Type != dto.MessageTypeGameUpdated {
		return false
	}

	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return false
	}

	game, ok := payload["game"].(map[string]interface{})
	if !ok {
		return false
	}

	otherPlayers, ok := game["otherPlayers"].([]interface{})
	if !ok {
		return false
	}

	// Should have at least one other player (meaning 2 total)
	return len(otherPlayers) >= 1
}

// TestPlayerSeparationTwoPlayers tests that player data is properly separated
// when two players join a game - each should see their own full data and limited data for others
func TestPlayerSeparationTwoPlayers(t *testing.T) {
	// Create two test clients
	client1 := NewTestClient(t)
	client2 := NewTestClient(t)
	defer client1.Close()
	defer client2.Close()

	// Connect both clients
	err := client1.Connect()
	require.NoError(t, err, "Client 1 failed to connect")

	err = client2.Connect()
	require.NoError(t, err, "Client 2 failed to connect")

	// Create game with client 1
	gameID, err := client1.CreateGameViaHTTP()
	require.NoError(t, err, "Failed to create game")
	t.Logf("✅ Game created: %s", gameID)

	// Player 1 joins the game
	player1Name := "Player1"
	err = client1.JoinGameViaWebSocket(gameID, player1Name)
	require.NoError(t, err, "Player 1 failed to join")

	// Wait for player 1 initial game state
	gameUpdate1, err := client1.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Client1 should receive initial game-updated message")
	t.Log("✅ Player 1 connected")

	// Player 2 joins the game
	player2Name := "Player2"
	err = client2.JoinGameViaWebSocket(gameID, player2Name)
	require.NoError(t, err, "Player 2 failed to join")

	// Both clients should receive personalized game-updated messages with both players
	// Player 1 gets an update when Player 2 joins
	gameUpdate1, err = client1.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Client 1 should receive game-updated when Player 2 joins")

	// Player 2 gets their initial game state
	gameUpdate2, err := client2.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Client 2 should receive game-updated with both players")

	// Extract game data from both perspectives
	payload1, ok := gameUpdate1.Payload.(map[string]interface{})
	require.True(t, ok, "Game update payload 1 should be a map")
	game1, ok := payload1["game"].(map[string]interface{})
	require.True(t, ok, "Game data 1 should be present")

	payload2, ok := gameUpdate2.Payload.(map[string]interface{})
	require.True(t, ok, "Game update payload 2 should be a map")
	game2, ok := payload2["game"].(map[string]interface{})
	require.True(t, ok, "Game data 2 should be present")

	// Test Player 1's perspective
	t.Run("Player1Perspective", func(t *testing.T) {
		// Player 1 should have currentPlayer data (full data including hand cards)
		currentPlayer1, ok := game1["currentPlayer"].(map[string]interface{})
		require.True(t, ok, "Player 1 should have currentPlayer data")
		require.Equal(t, player1Name, currentPlayer1["name"], "Current player should be Player 1")

		// CARD INTEGRITY CHECK: Player 1 should see their own cards (hand cards array)
		cards1, ok := currentPlayer1["cards"].([]interface{})
		require.True(t, ok, "Player 1 should have cards array for their own hand")
		t.Logf("Player 1 has %d cards in hand (private to Player 1)", len(cards1))

		// CRITICAL: Verify that cards array contains actual card identifiers (not empty/null)
		for i, card := range cards1 {
			cardStr, isString := card.(string)
			require.True(t, isString, "Card %d should be a string identifier", i)
			require.NotEmpty(t, cardStr, "Card %d should not be empty", i)
			t.Logf("Player 1's card %d: %s", i, cardStr)
		}

		// Player 1 should have otherPlayers data (limited data)
		otherPlayers1, ok := game1["otherPlayers"].([]interface{})
		require.True(t, ok, "Player 1 should have otherPlayers data")
		require.Len(t, otherPlayers1, 1, "Player 1 should see 1 other player")

		// Check other player data (should be limited)
		otherPlayer := otherPlayers1[0].(map[string]interface{})
		require.Equal(t, player2Name, otherPlayer["name"], "Other player should be Player 2")

		// CARD INTEGRITY CHECK: Other player should NOT have access to hand cards
		_, hasCards := otherPlayer["cards"]
		require.False(t, hasCards, "Player 1 should NOT see Player 2's hand cards")

		// CARD INTEGRITY CHECK: Other player should have handCardCount instead of cards array
		handCardCount, ok := otherPlayer["handCardCount"].(float64)
		require.True(t, ok, "Other player should have handCardCount")
		require.GreaterOrEqual(t, int(handCardCount), 0, "Hand card count should be non-negative")
		t.Logf("Player 2 has %d cards (count only visible to Player 1)", int(handCardCount))

		// CARD INTEGRITY CHECK: Other player should have played cards visible (public information)
		playedCards, ok := otherPlayer["playedCards"].([]interface{})
		require.True(t, ok, "Other player should have playedCards visible (public info)")
		require.Empty(t, playedCards, "Played cards should be empty at game start")
		t.Logf("Player 2 has %d played cards (visible to Player 1)", len(playedCards))
	})

	// Test Player 2's perspective
	t.Run("Player2Perspective", func(t *testing.T) {
		// Player 2 should have currentPlayer data (full data including hand cards)
		currentPlayer2, ok := game2["currentPlayer"].(map[string]interface{})
		require.True(t, ok, "Player 2 should have currentPlayer data")
		require.Equal(t, player2Name, currentPlayer2["name"], "Current player should be Player 2")

		// CARD INTEGRITY CHECK: Player 2 should see their own cards (hand cards array)
		cards2, ok := currentPlayer2["cards"].([]interface{})
		require.True(t, ok, "Player 2 should have cards array for their own hand")
		t.Logf("Player 2 has %d cards in hand (private to Player 2)", len(cards2))

		// CRITICAL: Verify that cards array contains actual card identifiers (not empty/null)
		for i, card := range cards2 {
			cardStr, isString := card.(string)
			require.True(t, isString, "Card %d should be a string identifier", i)
			require.NotEmpty(t, cardStr, "Card %d should not be empty", i)
			t.Logf("Player 2's card %d: %s", i, cardStr)
		}

		// CARD INTEGRITY VERIFICATION: Player 1 and Player 2 should have different hand cards
		currentPlayer1, _ := game1["currentPlayer"].(map[string]interface{})
		cards1, _ := currentPlayer1["cards"].([]interface{})

		// Ensure players don't have identical hands (privacy violation check)
		if len(cards1) > 0 && len(cards2) > 0 {
			allSame := true
			minLen := len(cards1)
			if len(cards2) < minLen {
				minLen = len(cards2)
			}

			for i := 0; i < minLen; i++ {
				if cards1[i] != cards2[i] {
					allSame = false
					break
				}
			}
			// It's extremely unlikely that both players get identical hands, so this helps verify separation
			require.False(t, allSame && len(cards1) == len(cards2), "Players should not have identical hands (indicates privacy violation)")
		}

		// Player 2 should have otherPlayers data (limited data)
		otherPlayers2, ok := game2["otherPlayers"].([]interface{})
		require.True(t, ok, "Player 2 should have otherPlayers data")
		require.Len(t, otherPlayers2, 1, "Player 2 should see 1 other player")

		// Check other player data (should be limited)
		otherPlayer := otherPlayers2[0].(map[string]interface{})
		require.Equal(t, player1Name, otherPlayer["name"], "Other player should be Player 1")

		// CARD INTEGRITY CHECK: Other player should NOT have access to hand cards
		_, hasCards := otherPlayer["cards"]
		require.False(t, hasCards, "Player 2 should NOT see Player 1's hand cards")

		// CARD INTEGRITY CHECK: Other player should have handCardCount instead of cards array
		handCardCount, ok := otherPlayer["handCardCount"].(float64)
		require.True(t, ok, "Other player should have handCardCount")
		require.GreaterOrEqual(t, int(handCardCount), 0, "Hand card count should be non-negative")
		t.Logf("Player 1 has %d cards (count only visible to Player 2)", int(handCardCount))

		// CARD INTEGRITY CHECK: Other player should have played cards visible (public information)
		playedCards, ok := otherPlayer["playedCards"].([]interface{})
		require.True(t, ok, "Other player should have playedCards visible (public info)")
		require.Empty(t, playedCards, "Played cards should be empty at game start")
		t.Logf("Player 1 has %d played cards (visible to Player 2)", len(playedCards))
	})

	// Test data consistency
	t.Run("DataConsistency", func(t *testing.T) {
		// Both players should see the same game metadata
		require.Equal(t, game1["id"], game2["id"], "Game ID should be the same")
		require.Equal(t, game1["status"], game2["status"], "Game status should be the same")
		require.Equal(t, game1["currentPhase"], game2["currentPhase"], "Game phase should be the same")
		require.Equal(t, game1["hostPlayerId"], game2["hostPlayerId"], "Host player ID should be the same")

		// Both players should see the same global parameters
		globalParams1 := game1["globalParameters"].(map[string]interface{})
		globalParams2 := game2["globalParameters"].(map[string]interface{})
		require.Equal(t, globalParams1["temperature"], globalParams2["temperature"], "Temperature should match")
		require.Equal(t, globalParams1["oxygen"], globalParams2["oxygen"], "Oxygen should match")
		require.Equal(t, globalParams1["oceans"], globalParams2["oceans"], "Oceans should match")
	})

	// Critical card integrity verification test
	t.Run("CardIntegrityVerification", func(t *testing.T) {
		// Extract all player data for comprehensive verification
		currentPlayer1 := game1["currentPlayer"].(map[string]interface{})
		otherPlayers1, ok := game1["otherPlayers"].([]interface{})
		require.True(t, ok, "Player 1 should have otherPlayers array")
		require.NotEmpty(t, otherPlayers1, "Player 1 should have at least one other player")
		otherPlayer1 := otherPlayers1[0].(map[string]interface{})

		currentPlayer2 := game2["currentPlayer"].(map[string]interface{})
		otherPlayers2, ok := game2["otherPlayers"].([]interface{})
		require.True(t, ok, "Player 2 should have otherPlayers array")
		require.NotEmpty(t, otherPlayers2, "Player 2 should have at least one other player")
		otherPlayer2 := otherPlayers2[0].(map[string]interface{})

		// CRITICAL INTEGRITY CHECK 1: Current player should have cards array, others should not
		_, player1HasCards := currentPlayer1["cards"]
		_, player2HasCards := currentPlayer2["cards"]
		_, other1HasCards := otherPlayer1["cards"]
		_, other2HasCards := otherPlayer2["cards"]

		require.True(t, player1HasCards, "Player 1 should see their own cards")
		require.True(t, player2HasCards, "Player 2 should see their own cards")
		require.False(t, other1HasCards, "Player 1 should NOT see other player's cards")
		require.False(t, other2HasCards, "Player 2 should NOT see other player's cards")

		// CRITICAL INTEGRITY CHECK 2: Other players should have handCardCount, current players should not need it
		_, other1HasCount := otherPlayer1["handCardCount"]
		_, other2HasCount := otherPlayer2["handCardCount"]

		require.True(t, other1HasCount, "Other players should have handCardCount visible")
		require.True(t, other2HasCount, "Other players should have handCardCount visible")

		// CRITICAL INTEGRITY CHECK 3: Played cards should be visible to everyone
		player1PlayedCards, p1HasPlayed := currentPlayer1["playedCards"]
		player2PlayedCards, p2HasPlayed := currentPlayer2["playedCards"]
		other1PlayedCards, o1HasPlayed := otherPlayer1["playedCards"]
		other2PlayedCards, o2HasPlayed := otherPlayer2["playedCards"]

		require.True(t, p1HasPlayed, "Player 1 should see their own played cards")
		require.True(t, p2HasPlayed, "Player 2 should see their own played cards")
		require.True(t, o1HasPlayed, "Player 1 should see other player's played cards")
		require.True(t, o2HasPlayed, "Player 2 should see other player's played cards")

		// Verify played cards are arrays (even if empty)
		require.IsType(t, []interface{}{}, player1PlayedCards, "Played cards should be arrays")
		require.IsType(t, []interface{}{}, player2PlayedCards, "Played cards should be arrays")
		require.IsType(t, []interface{}{}, other1PlayedCards, "Played cards should be arrays")
		require.IsType(t, []interface{}{}, other2PlayedCards, "Played cards should be arrays")

		// CRITICAL INTEGRITY CHECK 4: HandCardCount should match actual card count
		player1Cards := currentPlayer1["cards"].([]interface{})
		player2Cards := currentPlayer2["cards"].([]interface{})
		other1Count := int(otherPlayer1["handCardCount"].(float64))
		other2Count := int(otherPlayer2["handCardCount"].(float64))

		require.Equal(t, len(player2Cards), other1Count, "Player 2's actual card count should match what Player 1 sees")
		require.Equal(t, len(player1Cards), other2Count, "Player 1's actual card count should match what Player 2 sees")

		t.Logf("✅ CARD INTEGRITY VERIFIED:")
		t.Logf("   - Player 1 has %d private cards, Player 2 sees count only", len(player1Cards))
		t.Logf("   - Player 2 has %d private cards, Player 1 sees count only", len(player2Cards))
		t.Logf("   - All played cards are visible to all players")
		t.Logf("   - Hand card counts match actual card counts")
	})

	t.Log("✅ Player separation test completed successfully")
}

// TestPlayerSeparationThreePlayers tests player separation with three players
func TestPlayerSeparationThreePlayers(t *testing.T) {
	// Create three test clients
	client1 := NewTestClient(t)
	client2 := NewTestClient(t)
	client3 := NewTestClient(t)
	defer client1.Close()
	defer client2.Close()
	defer client3.Close()

	// Connect all clients
	err := client1.Connect()
	require.NoError(t, err, "Client 1 failed to connect")

	err = client2.Connect()
	require.NoError(t, err, "Client 2 failed to connect")

	err = client3.Connect()
	require.NoError(t, err, "Client 3 failed to connect")

	// Create game and join all three players
	gameID, err := client1.CreateGameViaHTTP()
	require.NoError(t, err, "Failed to create game")

	// Join players sequentially
	playerNames := []string{"Player1", "Player2", "Player3"}

	// Player 1 joins and gets initial state
	err = client1.JoinGameViaWebSocket(gameID, playerNames[0])
	require.NoError(t, err, "Player 1 failed to join")
	gameUpdate1, err := client1.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive player 1 initial game state")

	// Player 2 joins - Player 1 gets an update, Player 2 gets initial state
	err = client2.JoinGameViaWebSocket(gameID, playerNames[1])
	require.NoError(t, err, "Player 2 failed to join")
	gameUpdate1, err = client1.WaitForMessage(dto.MessageTypeGameUpdated) // Player 1 update
	require.NoError(t, err, "Player 1 should receive update when Player 2 joins")
	gameUpdate2, err := client2.WaitForMessage(dto.MessageTypeGameUpdated) // Player 2 initial
	require.NoError(t, err, "Failed to receive player 2 initial game state")

	// Player 3 joins - All players get updates
	err = client3.JoinGameViaWebSocket(gameID, playerNames[2])
	require.NoError(t, err, "Player 3 failed to join")
	gameUpdate1, err = client1.WaitForMessage(dto.MessageTypeGameUpdated) // Player 1 final update
	require.NoError(t, err, "Player 1 should receive update when Player 3 joins")
	gameUpdate2, err = client2.WaitForMessage(dto.MessageTypeGameUpdated) // Player 2 final update
	require.NoError(t, err, "Player 2 should receive update when Player 3 joins")
	gameUpdate3, err := client3.WaitForMessage(dto.MessageTypeGameUpdated) // Player 3 initial
	require.NoError(t, err, "Failed to receive player 3 initial game state")

	// Collect all final game states
	gameUpdates := []*dto.WebSocketMessage{gameUpdate1, gameUpdate2, gameUpdate3}

	// Test that each player sees exactly 1 current player and 2 other players
	for i, gameUpdate := range gameUpdates {
		payload, ok := gameUpdate.Payload.(map[string]interface{})
		require.True(t, ok, "Game update payload %d should be a map", i+1)
		game, ok := payload["game"].(map[string]interface{})
		require.True(t, ok, "Game data %d should be present", i+1)

		// Each player should have their own currentPlayer data
		currentPlayer, ok := game["currentPlayer"].(map[string]interface{})
		require.True(t, ok, "Player %d should have currentPlayer data", i+1)
		require.Equal(t, playerNames[i], currentPlayer["name"], "Current player should match for client %d", i+1)

		// Each player should see exactly 2 other players
		otherPlayers, ok := game["otherPlayers"].([]interface{})
		require.True(t, ok, "Player %d should have otherPlayers data", i+1)
		require.Len(t, otherPlayers, 2, "Player %d should see exactly 2 other players", i+1)

		// Verify other player names are correct (excluding current player)
		var otherPlayerNames []string
		for _, op := range otherPlayers {
			otherPlayer := op.(map[string]interface{})
			name := otherPlayer["name"].(string)
			otherPlayerNames = append(otherPlayerNames, name)

			// Verify other players have handCardCount instead of cards
			_, hasCards := otherPlayer["cards"]
			require.False(t, hasCards, "Other players should not have 'cards' field")

			_, hasHandCardCount := otherPlayer["handCardCount"]
			require.True(t, hasHandCardCount, "Other players should have 'handCardCount' field")
		}

		// Verify the other player names are the expected ones (all players except current)
		expectedOthers := make([]string, 0)
		for j, name := range playerNames {
			if j != i { // Exclude current player
				expectedOthers = append(expectedOthers, name)
			}
		}
		require.ElementsMatch(t, expectedOthers, otherPlayerNames, "Other player names should match expected for client %d", i+1)
	}

	t.Log("✅ Three-player separation test completed successfully")
}

// TestPlayerSeparationCardVisibility tests that hand cards are private but played cards are public
func TestPlayerSeparationCardVisibility(t *testing.T) {
	// This test verifies the card visibility rules by reusing the core test logic
	// - Hand cards (cards field) should only be visible to the player who owns them
	// - Played cards (playedCards field) should be visible to all players
	// - Other players should only see handCardCount, not the actual cards

	// Use the same setup as TestPlayerSeparationTwoPlayers but focus on card visibility
	client1 := NewTestClient(t)
	client2 := NewTestClient(t)
	defer client1.Close()
	defer client2.Close()

	// Setup: Create game with two players (same as main test)
	err := client1.Connect()
	require.NoError(t, err)
	err = client2.Connect()
	require.NoError(t, err)

	gameID, err := client1.CreateGameViaHTTP()
	require.NoError(t, err)

	// Join players
	// Connect Player1 and ensure they're fully registered before Player2 joins
	err = client1.JoinGameViaWebSocket(gameID, "Player1")
	require.NoError(t, err)

	// Wait for Player1's initial game state
	_, err = client1.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Client1 should receive initial game-updated message")

	// Give a moment for Player1 to be fully registered before Player2 joins
	time.Sleep(50 * time.Millisecond)

	// Connect Player2 - this will trigger game state events that should reach both players
	err = client2.JoinGameViaWebSocket(gameID, "Player2")
	require.NoError(t, err)

	// Both players should eventually receive game-updated messages with both players present
	// We'll read multiple messages until we get ones with both players
	var gameUpdate1, gameUpdate2 *dto.WebSocketMessage

	// Try to get game states with both players for both clients
	for attempts := 0; attempts < 3; attempts++ {
		if gameUpdate1 == nil {
			msg, err := client1.WaitForMessage(dto.MessageTypeGameUpdated)
			if err == nil && hasMultiplePlayers(msg) {
				gameUpdate1 = msg
			} else if err != nil && attempts == 2 {
				require.NoError(t, err, "Client1 should receive game-updated with both players")
			}
		}

		if gameUpdate2 == nil {
			msg, err := client2.WaitForMessage(dto.MessageTypeGameUpdated)
			if err == nil && hasMultiplePlayers(msg) {
				gameUpdate2 = msg
			} else if err != nil && attempts == 2 {
				require.NoError(t, err, "Client2 should receive game-updated with both players")
			}
		}

		if gameUpdate1 != nil && gameUpdate2 != nil {
			break
		}
	}

	// Extract and verify card visibility rules
	payload1 := gameUpdate1.Payload.(map[string]interface{})
	game1 := payload1["game"].(map[string]interface{})
	payload2 := gameUpdate2.Payload.(map[string]interface{})
	game2 := payload2["game"].(map[string]interface{})

	// DEBUG: Print the actual received data structure
	t.Logf("DEBUG: Game1 CurrentPlayer: %+v", game1["currentPlayer"])
	t.Logf("DEBUG: Game1 OtherPlayers: %+v", game1["otherPlayers"])
	t.Logf("DEBUG: Game2 CurrentPlayer: %+v", game2["currentPlayer"])
	t.Logf("DEBUG: Game2 OtherPlayers: %+v", game2["otherPlayers"])

	// CARD VISIBILITY VERIFICATION
	currentPlayer1 := game1["currentPlayer"].(map[string]interface{})
	currentPlayer2 := game2["currentPlayer"].(map[string]interface{})
	otherPlayers1 := game1["otherPlayers"].([]interface{})
	otherPlayers2 := game2["otherPlayers"].([]interface{})

	require.Len(t, otherPlayers1, 1, "Player 1 should see 1 other player")
	require.Len(t, otherPlayers2, 1, "Player 2 should see 1 other player")

	otherPlayer1 := otherPlayers1[0].(map[string]interface{})
	otherPlayer2 := otherPlayers2[0].(map[string]interface{})

	// Test card visibility rules
	t.Run("CardVisibilityRules", func(t *testing.T) {
		// Rule 1: Each player should see their own cards
		_, p1HasOwnCards := currentPlayer1["cards"]
		_, p2HasOwnCards := currentPlayer2["cards"]
		require.True(t, p1HasOwnCards, "Player 1 should see their own cards")
		require.True(t, p2HasOwnCards, "Player 2 should see their own cards")

		// Rule 2: Players should NOT see other players' hand cards
		_, o1HasCards := otherPlayer1["cards"]
		_, o2HasCards := otherPlayer2["cards"]
		require.False(t, o1HasCards, "Player 1 should NOT see Player 2's hand cards")
		require.False(t, o2HasCards, "Player 2 should NOT see Player 1's hand cards")

		// Rule 3: Players should see other players' hand card counts
		_, o1HasCount := otherPlayer1["handCardCount"]
		_, o2HasCount := otherPlayer2["handCardCount"]
		require.True(t, o1HasCount, "Player 1 should see Player 2's hand card count")
		require.True(t, o2HasCount, "Player 2 should see Player 1's hand card count")

		// Rule 4: All players should see played cards (public information)
		_, p1HasPlayedCards := currentPlayer1["playedCards"]
		_, p2HasPlayedCards := currentPlayer2["playedCards"]
		_, o1HasPlayedCards := otherPlayer1["playedCards"]
		_, o2HasPlayedCards := otherPlayer2["playedCards"]
		require.True(t, p1HasPlayedCards, "Player 1 should see their own played cards")
		require.True(t, p2HasPlayedCards, "Player 2 should see their own played cards")
		require.True(t, o1HasPlayedCards, "Player 1 should see Player 2's played cards")
		require.True(t, o2HasPlayedCards, "Player 2 should see Player 1's played cards")

		t.Log("✅ All card visibility rules verified")
	})

	t.Log("✅ Card visibility test completed successfully")
}
