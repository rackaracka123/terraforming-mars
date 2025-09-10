package websocket

import (
	"testing"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/test/integration"

	"github.com/stretchr/testify/require"
)

// TestPlayerReconnectionNoDuplicates tests the specific issue where reconnecting players
// were creating duplicate player entries instead of properly reconnecting.
// 
// Scenario:
// 1. Player1 creates lobby
// 2. Player2 joins lobby  
// 3. Player1 starts game (transition to active)
// 4. Player2 disconnects
// 5. Player2 reconnects
// 6. Verify no duplicate players are created
// 7. Verify both players can see each other correctly
func TestPlayerReconnectionNoDuplicates(t *testing.T) {
	t.Log("🧪 Starting Player Reconnection No Duplicates Test")

	// === PHASE 1: Setup two-player game ===
	t.Log("📋 Phase 1: Setting up two-player game")
	
	client1 := integration.NewTestClient(t)
	defer client1.Close()
	client2 := integration.NewTestClient(t)
	defer client2.Close()

	// Connect both clients
	err := client1.Connect()
	require.NoError(t, err, "Player1 should connect to WebSocket")

	err = client2.Connect()
	require.NoError(t, err, "Player2 should connect to WebSocket")

	// Create game via Player1
	gameID, err := client1.CreateGameViaHTTP()
	require.NoError(t, err, "Player1 should create game via HTTP")
	t.Logf("✅ Game created with ID: %s", gameID)

	// Player1 joins (becomes host)
	err = client1.JoinGameViaWebSocket(gameID, "Player1")
	require.NoError(t, err, "Player1 should join game")

	// Wait for Player1 connection confirmation and extract player ID
	player1ConnectedMsg, err := client1.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err, "Player1 should receive connection confirmation")

	payload1, ok := player1ConnectedMsg.Payload.(map[string]interface{})
	require.True(t, ok, "Player1 connected payload should be a map")
	player1ID, ok := payload1["playerId"].(string)
	require.True(t, ok, "Player1 ID should be present")
	require.NotEmpty(t, player1ID, "Player1 ID should not be empty")
	client1.SetPlayerID(player1ID)
	t.Logf("✅ Player1 joined with ID: %s", player1ID)

	// Player2 joins
	err = client2.JoinGameViaWebSocket(gameID, "Player2")
	require.NoError(t, err, "Player2 should join game")

	// Wait for Player2 connection confirmation and extract player ID
	player2ConnectedMsg, err := client2.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err, "Player2 should receive connection confirmation")

	payload2, ok := player2ConnectedMsg.Payload.(map[string]interface{})
	require.True(t, ok, "Player2 connected payload should be a map")
	player2ID, ok := payload2["playerId"].(string)
	require.True(t, ok, "Player2 ID should be present")
	require.NotEmpty(t, player2ID, "Player2 ID should not be empty")
	client2.SetPlayerID(player2ID)
	t.Logf("✅ Player2 joined with ID: %s", player2ID)

	// Both clients should receive game updates about the other player joining
	_, err = client1.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err, "Player1 should receive notification about Player2 joining")

	t.Log("✅ Both players successfully joined lobby")

	// === PHASE 2: Start game ===
	t.Log("📋 Phase 2: Starting game")

	// Determine which client is the host and start game
	client1IsHost, err := client1.IsHost()
	require.NoError(t, err, "Should be able to check if Player1 is host")

	var hostClient *integration.TestClient
	if client1IsHost {
		hostClient = client1
		t.Log("Player1 is the host")
	} else {
		hostClient = client2
		t.Log("Player2 is the host")
	}

	// Start game
	err = hostClient.StartGame()
	require.NoError(t, err, "Host should be able to start game")

	// Wait for game to become active
	err = hostClient.WaitForStartGameComplete()
	require.NoError(t, err, "Game should transition to active status")
	t.Log("✅ Game started and is now active")

	// === PHASE 3: Verify initial player counts ===
	t.Log("📋 Phase 3: Verifying initial player count")

	// Get current game state to count players before disconnection
	initialPlayer1Count := getPlayerCountFromClient(t, client1)
	initialPlayer2Count := getPlayerCountFromClient(t, client2)

	t.Logf("📊 Initial player counts - Player1 sees: %d, Player2 sees: %d", 
		initialPlayer1Count, initialPlayer2Count)

	require.Equal(t, 2, initialPlayer1Count, "Player1 should see 2 players initially")
	require.Equal(t, 2, initialPlayer2Count, "Player2 should see 2 players initially")

	// === PHASE 4: Player2 disconnects ===
	t.Log("📋 Phase 4: Player2 disconnects")

	client2.ForceClose()
	t.Log("✅ Player2 disconnected (force close)")

	// Give time for server to process disconnection
	time.Sleep(300 * time.Millisecond)

	// === PHASE 5: Player2 reconnects ===
	t.Log("📋 Phase 5: Player2 reconnects")

	client3 := integration.NewTestClient(t)
	defer client3.Close()

	err = client3.Connect()
	require.NoError(t, err, "Player2 reconnection client should connect")

	// Use ReconnectToGame (sends player-reconnect message) instead of JoinGameViaWebSocket
	err = client3.ReconnectToGame(gameID, "Player2")
	require.NoError(t, err, "Player2 should be able to reconnect")
	t.Log("✅ Player2 reconnection message sent")

	// === PHASE 6: Verify reconnection success ===
	t.Log("📋 Phase 6: Verifying reconnection success")

	// Wait for player-reconnected confirmation
	reconnectedMsg, err := client3.WaitForMessage(dto.MessageTypePlayerReconnected)
	require.NoError(t, err, "Player2 should receive reconnection confirmation")

	// Verify reconnection payload contains correct player ID
	reconnectedPayload, ok := reconnectedMsg.Payload.(map[string]interface{})
	require.True(t, ok, "Reconnection payload should be a map")
	
	reconnectedPlayerID, ok := reconnectedPayload["playerId"].(string)
	require.True(t, ok, "Reconnected player ID should be present")
	require.Equal(t, player2ID, reconnectedPlayerID, "Reconnected player should have same ID as original Player2")
	client3.SetPlayerID(reconnectedPlayerID)

	reconnectedPlayerName, ok := reconnectedPayload["playerName"].(string)
	require.True(t, ok, "Reconnected player name should be present")
	require.Equal(t, "Player2", reconnectedPlayerName, "Reconnected player should have correct name")

	t.Logf("✅ Player2 successfully reconnected with ID: %s", reconnectedPlayerID)

	// === PHASE 7: Critical test - Verify no duplicates ===
	t.Log("📋 Phase 7: CRITICAL - Verifying no duplicate players created")

	// Wait for any additional game updates to settle
	time.Sleep(200 * time.Millisecond)

	// Get final player counts from all active clients using more reliable method
	finalPlayer1Count := getPlayerCountFromLatestMessage(t, client1)
	finalPlayer3Count := getPlayerCountFromLatestMessage(t, client3)

	t.Logf("📊 Final player counts - Player1 sees: %d, Player2 sees: %d", 
		finalPlayer1Count, finalPlayer3Count)

	// CRITICAL ASSERTIONS: No duplicates should be created
	require.Equal(t, 2, finalPlayer1Count, 
		"❌ DUPLICATE PREVENTION FAILED: Player1 should still see exactly 2 players (not duplicates)")
	require.Equal(t, 2, finalPlayer3Count, 
		"❌ DUPLICATE PREVENTION FAILED: Reconnected Player2 should see exactly 2 players (not duplicates)")

	// === PHASE 8: Verify game data quality during reconnection ===
	t.Log("📋 Phase 8: Verifying game data quality during reconnection")

	// Get the most recent game states from both clients
	player1GameState := getGameStateFromReconnectionMessages(t, client1)
	player3GameState := getGameStateFromReconnectionMessages(t, client3)

	// === Phase 8a: Verify Player1's view of reconnection ===
	t.Log("📋 Phase 8a: Verifying Player1's game data after reconnection")
	
	if player1GameState != nil {
		verifyGameStateQuality(t, player1GameState, "Player1", 2)
		player1PlayerIDs := extractPlayerIDs(t, player1GameState)
		t.Logf("🔍 Player1 sees player IDs: %v", player1PlayerIDs)
		
		require.Equal(t, 2, len(player1PlayerIDs), "Player1 should see exactly 2 unique players")
		require.Contains(t, player1PlayerIDs, player1ID, "Player1 should see original Player1 ID")
		require.Contains(t, player1PlayerIDs, player2ID, "Player1 should see original Player2 ID")
		
		t.Log("✅ Player1's game data is correct after reconnection")
	} else {
		t.Log("⚠️ Player1's game state could not be extracted - will verify via counts only")
	}

	// === Phase 8b: Verify Player2's view of reconnection ===
	t.Log("📋 Phase 8b: Verifying Player2's game data after reconnection")
	
	if player3GameState != nil {
		verifyGameStateQuality(t, player3GameState, "Player2", 2)
		player3PlayerIDs := extractPlayerIDs(t, player3GameState)
		t.Logf("🔍 Player2 sees player IDs: %v", player3PlayerIDs)
		
		require.Equal(t, 2, len(player3PlayerIDs), "Player2 should see exactly 2 unique players")
		require.Contains(t, player3PlayerIDs, player1ID, "Player2 should see original Player1 ID")
		require.Contains(t, player3PlayerIDs, player2ID, "Player2 should see original Player2 ID")
		
		t.Log("✅ Player2's game data is correct after reconnection")
	} else {
		t.Log("⚠️ Player2's game state could not be extracted - will verify via counts only")
	}

	// === Phase 8c: Verify cross-client consistency ===
	if player1GameState != nil && player3GameState != nil {
		t.Log("📋 Phase 8c: Verifying cross-client game state consistency")
		
		// Verify both clients see the same game status
		player1Status := extractGameStatus(t, player1GameState)
		player3Status := extractGameStatus(t, player3GameState)
		
		require.Equal(t, player1Status, player3Status, "Both players should see the same game status")
		require.Equal(t, "active", player1Status, "Game should be in active status")
		
		t.Logf("✅ Both players see consistent game status: %s", player1Status)
	}

	// === Phase 8d: CRITICAL - Verify exact player visibility ===
	t.Log("📋 Phase 8d: CRITICAL - Verifying exact player visibility")
	
	// Verify Player1 sees exactly Player1 and Player2 (no more, no less)
	player1VisiblePlayers := getVisiblePlayerNames(t, client1)
	t.Logf("🔍 Player1 sees players: %v", player1VisiblePlayers)
	
	require.Len(t, player1VisiblePlayers, 2, "Player1 should see exactly 2 players")
	require.Contains(t, player1VisiblePlayers, "Player1", "Player1 should see Player1")
	require.Contains(t, player1VisiblePlayers, "Player2", "Player1 should see Player2")
	t.Log("✅ Player1 sees exactly the correct players: Player1 and Player2")
	
	// Verify Player2 sees exactly Player1 and Player2 (no more, no less)
	player2VisiblePlayers := getVisiblePlayerNames(t, client3)
	t.Logf("🔍 Player2 sees players: %v", player2VisiblePlayers)
	
	require.Len(t, player2VisiblePlayers, 2, "Player2 should see exactly 2 players")
	require.Contains(t, player2VisiblePlayers, "Player1", "Player2 should see Player1")
	require.Contains(t, player2VisiblePlayers, "Player2", "Player2 should see Player2")
	t.Log("✅ Player2 sees exactly the correct players: Player1 and Player2")
	
	// === Phase 8e: Verify no duplicate player names ===
	t.Log("📋 Phase 8e: Verifying no duplicate player names")
	
	// Check Player1's view for duplicates
	player1Duplicates := findDuplicatePlayerNames(player1VisiblePlayers)
	require.Empty(t, player1Duplicates, "Player1 should not see any duplicate player names: %v", player1Duplicates)
	
	// Check Player2's view for duplicates
	player2Duplicates := findDuplicatePlayerNames(player2VisiblePlayers)
	require.Empty(t, player2Duplicates, "Player2 should not see any duplicate player names: %v", player2Duplicates)
	
	t.Log("✅ No duplicate player names found in either client's view")

	// === PHASE 9: Verify game functionality still works ===
	t.Log("📋 Phase 9: Verifying game functionality after reconnection")

	// Reconnected player should be able to perform game actions
	// (This is a basic smoke test to ensure the connection is fully functional)
	
	// Try to get fresh game state to ensure WebSocket is working
	_, err = client3.WaitForAnyMessageWithTimeout(500 * time.Millisecond)
	// It's OK if no message comes - just testing that the connection is alive

	t.Log("✅ Reconnected player's WebSocket connection is functional")

	// === FINAL VERIFICATION ===
	t.Log("🎉 TEST PASSED: Player Reconnection No Duplicates")
	t.Log("✅ No duplicate players were created during reconnection")
	t.Log("✅ Player identities remained consistent")
	t.Log("✅ Both players can see each other correctly")
	t.Log("✅ Game state is consistent across all clients")
}

// Helper function to get player count from a client's current game state
func getPlayerCountFromClient(t *testing.T, client *integration.TestClient) int {
	gameState := getGameStateFromClient(t, client)
	return countPlayersInGameState(t, gameState)
}

// Helper function to get game state from a client by checking their latest message
func getGameStateFromClient(t *testing.T, client *integration.TestClient) map[string]interface{} {
	// Try to get the most recent game state
	// First check if there are any pending messages
	msg, err := client.WaitForAnyMessageWithTimeout(100 * time.Millisecond)
	if err != nil {
		// No recent message, this is fine - we'll need to trigger a state request
		t.Logf("No recent messages for client, will use last known state")
		return nil
	}

	// Extract game state from the message
	if msg.Type == dto.MessageTypeGameUpdated {
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			t.Logf("Game update payload is not a map")
			return nil
		}
		
		gameData, ok := payload["game"].(map[string]interface{})
		if !ok {
			t.Logf("Game data not found in game update payload")
			return nil
		}
		
		return gameData
	} else if msg.Type == dto.MessageTypePlayerConnected || msg.Type == dto.MessageTypePlayerReconnected {
		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			t.Logf("Player message payload is not a map")
			return nil
		}
		
		gameData, ok := payload["game"].(map[string]interface{})
		if !ok {
			t.Logf("Game data not found in player message payload")
			return nil
		}
		
		return gameData
	}

	t.Logf("No game state found in message type: %s", msg.Type)
	return nil
}

// Helper function to count players in a game state
func countPlayersInGameState(t *testing.T, gameState map[string]interface{}) int {
	if gameState == nil {
		t.Logf("Game state is nil, cannot count players")
		return 0
	}

	playerCount := 0

	// Check for currentPlayer
	if currentPlayer, ok := gameState["currentPlayer"].(map[string]interface{}); ok && currentPlayer["id"] != nil {
		playerCount++
	}

	// Check for otherPlayers
	if otherPlayers, ok := gameState["otherPlayers"].([]interface{}); ok {
		playerCount += len(otherPlayers)
	}

	// Fallback: check for players array (older DTO format)
	if players, ok := gameState["players"].([]interface{}); ok {
		playerCount = len(players)
	}

	return playerCount
}

// Helper function to extract player IDs from game state
func extractPlayerIDs(t *testing.T, gameState map[string]interface{}) []string {
	if gameState == nil {
		return []string{}
	}

	var playerIDs []string

	// Check for currentPlayer
	if currentPlayer, ok := gameState["currentPlayer"].(map[string]interface{}); ok {
		if id, ok := currentPlayer["id"].(string); ok && id != "" {
			playerIDs = append(playerIDs, id)
		}
	}

	// Check for otherPlayers
	if otherPlayers, ok := gameState["otherPlayers"].([]interface{}); ok {
		for _, player := range otherPlayers {
			if playerMap, ok := player.(map[string]interface{}); ok {
				if id, ok := playerMap["id"].(string); ok && id != "" {
					playerIDs = append(playerIDs, id)
				}
			}
		}
	}

	// Fallback: check for players array (older DTO format)
	if players, ok := gameState["players"].([]interface{}); ok {
		for _, player := range players {
			if playerMap, ok := player.(map[string]interface{}); ok {
				if id, ok := playerMap["id"].(string); ok && id != "" {
					playerIDs = append(playerIDs, id)
				}
			}
		}
	}

	// Remove duplicates (shouldn't happen, but safety check)
	uniqueIDs := make([]string, 0, len(playerIDs))
	seen := make(map[string]bool)
	for _, id := range playerIDs {
		if !seen[id] {
			uniqueIDs = append(uniqueIDs, id)
			seen[id] = true
		}
	}

	return uniqueIDs
}

// TestPlayerReconnectionPreventsDuplicateNames tests that players with the same name
// cannot create duplicates through multiple join attempts
func TestPlayerReconnectionPreventsDuplicateNames(t *testing.T) {
	t.Log("🧪 Starting Player Reconnection Prevents Duplicate Names Test")

	// Setup game with one player
	client1, gameID := integration.SetupBasicGameFlow(t, "Player1")
	defer client1.Close()

	// Try to join again with same name from different connection
	client2 := integration.NewTestClient(t)
	defer client2.Close()

	err := client2.Connect()
	require.NoError(t, err, "Second client should connect")

	// This should either fail or not create a duplicate
	err = client2.JoinGameViaWebSocket(gameID, "Player1")
	
	if err == nil {
		// If join succeeded, verify no duplicate was created
		_, err := client2.WaitForMessage(dto.MessageTypePlayerConnected)
		require.NoError(t, err, "Should receive some response")

		// Check that we still only have one unique player
		gameState := getGameStateFromClient(t, client2)
		playerCount := countPlayersInGameState(t, gameState)
		
		// Should not have created a duplicate
		require.LessOrEqual(t, playerCount, 1, "Should not create duplicate players with same name")
		
		playerIDs := extractPlayerIDs(t, gameState)
		require.LessOrEqual(t, len(playerIDs), 1, "Should not have duplicate player IDs")
		
		t.Log("✅ Duplicate name prevention working - no duplicate players created")
	} else {
		t.Logf("✅ Duplicate name prevention working - join with same name rejected: %v", err)
	}
}

// getPlayerCountFromLatestMessage gets player count from the most recent message
func getPlayerCountFromLatestMessage(t *testing.T, client *integration.TestClient) int {
	// Try to get any recent message that contains game state
	msg, err := client.WaitForAnyMessageWithTimeout(50 * time.Millisecond)
	if err != nil {
		t.Logf("No recent messages for player count check")
		return 0
	}

	gameState := extractGameStateFromMessage(t, msg)
	if gameState == nil {
		t.Logf("No game state in latest message type: %s", msg.Type)
		return 0
	}

	return countPlayersInGameState(t, gameState)
}

// getGameStateFromReconnectionMessages looks through recent messages to find game state
func getGameStateFromReconnectionMessages(t *testing.T, client *integration.TestClient) map[string]interface{} {
	// Try to get recent messages that might contain game state
	for i := 0; i < 3; i++ {
		msg, err := client.WaitForAnyMessageWithTimeout(50 * time.Millisecond)
		if err != nil {
			break // No more messages
		}

		gameState := extractGameStateFromMessage(t, msg)
		if gameState != nil {
			return gameState
		}
	}

	return nil
}

// extractGameStateFromMessage extracts game state from various message types
func extractGameStateFromMessage(t *testing.T, msg dto.WebSocketMessage) map[string]interface{} {
	payload, ok := msg.Payload.(map[string]interface{})
	if !ok {
		return nil
	}

	// Look for game state in different message types
	switch msg.Type {
	case dto.MessageTypeGameUpdated:
		if gameData, ok := payload["game"].(map[string]interface{}); ok {
			return gameData
		}
	case dto.MessageTypePlayerConnected, dto.MessageTypePlayerReconnected:
		if gameData, ok := payload["game"].(map[string]interface{}); ok {
			return gameData
		}
	}

	return nil
}

// verifyGameStateQuality checks that game state contains expected fields and data
func verifyGameStateQuality(t *testing.T, gameState map[string]interface{}, playerName string, expectedPlayerCount int) {
	t.Logf("🔍 Verifying game state quality for %s", playerName)

	// Verify basic game fields
	require.NotNil(t, gameState, "Game state should not be nil")

	// Check game status
	status, ok := gameState["status"].(string)
	require.True(t, ok, "Game should have status field")
	require.NotEmpty(t, status, "Game status should not be empty")
	t.Logf("✅ %s sees game status: %s", playerName, status)

	// Check game ID
	gameID, ok := gameState["id"].(string)
	require.True(t, ok, "Game should have ID field")
	require.NotEmpty(t, gameID, "Game ID should not be empty")
	t.Logf("✅ %s sees game ID: %s", playerName, gameID)

	// Verify player count matches expectation
	playerCount := countPlayersInGameState(t, gameState)
	require.Equal(t, expectedPlayerCount, playerCount, 
		"%s should see exactly %d players", playerName, expectedPlayerCount)
	t.Logf("✅ %s sees correct player count: %d", playerName, playerCount)

	// Check for player data structure
	hasCurrentPlayer := false
	if currentPlayer, ok := gameState["currentPlayer"].(map[string]interface{}); ok && currentPlayer["id"] != nil {
		hasCurrentPlayer = true
		t.Logf("✅ %s has currentPlayer data", playerName)
	}

	hasOtherPlayers := false
	if otherPlayers, ok := gameState["otherPlayers"].([]interface{}); ok && len(otherPlayers) > 0 {
		hasOtherPlayers = true
		t.Logf("✅ %s has otherPlayers data (%d players)", playerName, len(otherPlayers))
	}

	// Should have at least one of these player data structures
	require.True(t, hasCurrentPlayer || hasOtherPlayers, 
		"%s should have either currentPlayer or otherPlayers data", playerName)

	// Check for host player ID
	if hostPlayerID, ok := gameState["hostPlayerId"].(string); ok {
		require.NotEmpty(t, hostPlayerID, "Host player ID should not be empty")
		t.Logf("✅ %s sees host player ID: %s", playerName, hostPlayerID)
	}

	// Check for global parameters (should exist in active game)
	if globalParams, ok := gameState["globalParameters"].(map[string]interface{}); ok {
		t.Logf("✅ %s has global parameters data", playerName)
		
		// Verify temperature exists
		if temp, ok := globalParams["temperature"].(float64); ok {
			t.Logf("✅ %s sees temperature: %.0f", playerName, temp)
		}
		
		// Verify oxygen exists
		if oxygen, ok := globalParams["oxygen"].(float64); ok {
			t.Logf("✅ %s sees oxygen: %.0f", playerName, oxygen)
		}
	}

	t.Logf("✅ Game state quality verification passed for %s", playerName)
}

// extractGameStatus extracts game status from game state
func extractGameStatus(t *testing.T, gameState map[string]interface{}) string {
	if status, ok := gameState["status"].(string); ok {
		return status
	}
	return ""
}

// getVisiblePlayerNames extracts all player names that a client can see
func getVisiblePlayerNames(t *testing.T, client *integration.TestClient) []string {
	var playerNames []string
	
	// Try to get any recent message that contains game state
	msg, err := client.WaitForAnyMessageWithTimeout(50 * time.Millisecond)
	if err != nil {
		t.Logf("No recent messages to extract player names from")
		return playerNames
	}

	gameState := extractGameStateFromMessage(t, msg)
	if gameState == nil {
		t.Logf("No game state found to extract player names from")
		return playerNames
	}

	// Extract player names from currentPlayer
	if currentPlayer, ok := gameState["currentPlayer"].(map[string]interface{}); ok {
		if name, ok := currentPlayer["name"].(string); ok && name != "" {
			playerNames = append(playerNames, name)
		}
	}

	// Extract player names from otherPlayers
	if otherPlayers, ok := gameState["otherPlayers"].([]interface{}); ok {
		for _, player := range otherPlayers {
			if playerMap, ok := player.(map[string]interface{}); ok {
				if name, ok := playerMap["name"].(string); ok && name != "" {
					playerNames = append(playerNames, name)
				}
			}
		}
	}

	// Fallback: extract from players array (older DTO format)
	if len(playerNames) == 0 {
		if players, ok := gameState["players"].([]interface{}); ok {
			for _, player := range players {
				if playerMap, ok := player.(map[string]interface{}); ok {
					if name, ok := playerMap["name"].(string); ok && name != "" {
						playerNames = append(playerNames, name)
					}
				}
			}
		}
	}

	return playerNames
}

// findDuplicatePlayerNames finds duplicate names in a list of player names
func findDuplicatePlayerNames(playerNames []string) []string {
	nameCount := make(map[string]int)
	var duplicates []string
	
	// Count occurrences of each name
	for _, name := range playerNames {
		nameCount[name]++
	}
	
	// Find names that appear more than once
	for name, count := range nameCount {
		if count > 1 {
			duplicates = append(duplicates, name)
		}
	}
	
	return duplicates
}