package websocket

import (
	"testing"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/test/integration"
)

// CountPlayersInGameState counts the number of players in a game state
func CountPlayersInGameState(t *testing.T, gameState map[string]interface{}) int {
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

// ExtractPlayerIDs extracts player IDs from game state
func ExtractPlayerIDs(t *testing.T, gameState map[string]interface{}) []string {
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

// VerifyGameStateQuality checks that game state contains expected fields and data
func VerifyGameStateQuality(t *testing.T, gameState map[string]interface{}, playerName string, expectedPlayerCount int) {
	t.Logf("🔍 Verifying game state quality for %s", playerName)

	// Verify basic game fields
	if gameState == nil {
		t.Fatalf("Game state should not be nil")
		return
	}

	// Check game status
	status, ok := gameState["status"].(string)
	if !ok || status == "" {
		t.Fatalf("Game should have non-empty status field")
		return
	}
	t.Logf("✅ %s sees game status: %s", playerName, status)

	// Check game ID
	gameID, ok := gameState["id"].(string)
	if !ok || gameID == "" {
		t.Fatalf("Game should have non-empty ID field")
		return
	}
	t.Logf("✅ %s sees game ID: %s", playerName, gameID)

	// Verify player count matches expectation
	playerCount := CountPlayersInGameState(t, gameState)
	if playerCount != expectedPlayerCount {
		t.Fatalf("%s should see exactly %d players, but sees %d", playerName, expectedPlayerCount, playerCount)
		return
	}
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
	if !hasCurrentPlayer && !hasOtherPlayers {
		t.Fatalf("%s should have either currentPlayer or otherPlayers data", playerName)
		return
	}

	// Check for host player ID
	if hostPlayerID, ok := gameState["hostPlayerId"].(string); ok {
		if hostPlayerID == "" {
			t.Fatalf("Host player ID should not be empty")
			return
		}
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

// ExtractGameStatus extracts game status from game state
func ExtractGameStatus(t *testing.T, gameState map[string]interface{}) string {
	if status, ok := gameState["status"].(string); ok {
		return status
	}
	return ""
}

// ExtractPlayerNamesFromGameState extracts player names directly from a game state map
func ExtractPlayerNamesFromGameState(t *testing.T, gameState map[string]interface{}) []string {
	var playerNames []string

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

	// Fallback: check for players array (older DTO format)
	if players, ok := gameState["players"].([]interface{}); ok {
		for _, player := range players {
			if playerMap, ok := player.(map[string]interface{}); ok {
				if name, ok := playerMap["name"].(string); ok && name != "" {
					playerNames = append(playerNames, name)
				}
			}
		}
	}

	return playerNames
}

// FindDuplicatePlayerNames finds duplicate names in a list of player names
func FindDuplicatePlayerNames(playerNames []string) []string {
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

// GetGameStateFromClient gets game state from a client by waiting for game-updated message
func GetGameStateFromClient(t *testing.T, client *integration.TestClient) map[string]interface{} {
	// Wait for a game-updated message specifically
	msg, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	if err != nil {
		t.Logf("No game-updated message available: %v", err)
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
	} else if msg.Type == dto.MessageTypePlayerConnected {
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

// GetPlayerCountFromClient gets player count from a client's current game state
func GetPlayerCountFromClient(t *testing.T, client *integration.TestClient) int {
	gameState := GetGameStateFromClient(t, client)
	return CountPlayersInGameState(t, gameState)
}
