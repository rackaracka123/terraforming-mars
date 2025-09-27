package integration

import (
	"encoding/json"
	"testing"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"

	"github.com/stretchr/testify/require"
)

// TestStartGameFlow tests the complete flow from lobby to game start
func TestStartGameFlow(t *testing.T) {
	// Create test client
	client := NewTestClient(t)
	defer client.Close()

	// Step 1: Connect to WebSocket
	err := client.Connect()
	require.NoError(t, err, "Failed to connect to WebSocket server")
	t.Log("✅ Connected to WebSocket server")

	// Step 2: Create game via HTTP API
	gameID, err := client.CreateGameViaHTTP()
	require.NoError(t, err, "Failed to create game via HTTP")
	require.NotEmpty(t, gameID, "Game ID should not be empty")
	t.Logf("✅ Game created with ID: %s", gameID)

	// Step 3: Join the created game via WebSocket
	playerName := "TestPlayer"
	err = client.JoinGameViaWebSocket(gameID, playerName)
	require.NoError(t, err, "Failed to join game via WebSocket")
	t.Log("✅ Sent join game message")

	// Step 4: Wait for game state confirmation
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game state message")
	require.NotNil(t, message, "Game state message should not be nil")
	t.Log("✅ Received game state confirmation")

	// Extract player ID from the game state
	payload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok, "Game state payload should be a map")

	// Navigate through the game state to find the current player ID
	if gameData, ok := payload["game"].(map[string]interface{}); ok {
		if currentPlayer, ok := gameData["currentPlayer"].(map[string]interface{}); ok {
			if playerID, ok := currentPlayer["id"].(string); ok {
				client.playerID = playerID
				t.Logf("✅ Player ID assigned: %s", playerID)
			} else {
				t.Fatal("Failed to extract player ID from current player")
			}
		} else {
			t.Fatal("Current player not found in game state")
		}
	} else {
		t.Fatal("Game data not found in payload")
	}

	// Step 5: Start the game (as host)
	err = client.StartGame()
	require.NoError(t, err, "Failed to send start game action")
	t.Log("✅ Sent start game action")

	// Step 6: Give a moment for the server to process the start game action
	time.Sleep(100 * time.Millisecond)

	// Wait for any message and check if it's the correct game-updated message
	// We might receive multiple game-updated messages, need to find the one with active status
	var gameUpdatedMessage *dto.WebSocketMessage
	for attempts := 0; attempts < 5; attempts++ {
		message, err = client.WaitForAnyMessage()
		if err != nil {
			break
		}

		if message.Type == dto.MessageTypeGameUpdated {
			// Check if this message contains the active game state
			if payload, ok := message.Payload.(map[string]interface{}); ok {
				if gameData, ok := payload["game"].(map[string]interface{}); ok {
					if gameStatus, ok := gameData["status"].(string); ok {
						if gameStatus == "active" {
							gameUpdatedMessage = message
							break
						}
					}
				}
			}
		}
	}

	require.NotNil(t, gameUpdatedMessage, "Failed to receive game updated message with active status")
	message = gameUpdatedMessage
	t.Log("✅ Received game-updated message with active status")

	// Step 7: Verify game status changed to active
	payload, ok = message.Payload.(map[string]interface{})
	require.True(t, ok, "Game updated payload should be a map")

	gameData, ok := payload["game"].(map[string]interface{})
	require.True(t, ok, "Game data should be present in payload")

	gameStatus, ok := gameData["status"].(string)
	require.True(t, ok, "Game status should be present")

	// Verify game phase changed
	currentPhase, ok := gameData["currentPhase"].(string)
	require.True(t, ok, "Current phase should be present")

	t.Logf("📊 Game State: Status=%s, Phase=%s", gameStatus, currentPhase)

	if gameStatus != "active" {
		t.Errorf("❌ Game status should be 'active', got '%s'", gameStatus)
		// Let's also log the full game data for debugging
		if gameDataBytes, err := json.Marshal(gameData); err == nil {
			t.Logf("🔍 Full game data: %s", string(gameDataBytes))
		}
		t.FailNow()
	}

	require.Equal(t, "starting_card_selection", currentPhase, "Game should be in card selection phase after start")

	t.Log("✅ Game successfully started and transitioned to active status!")
	t.Logf("   Status: %s, Phase: %s", gameStatus, currentPhase)
}
