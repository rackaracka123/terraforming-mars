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

	// Step 4: Wait for player connected confirmation
	message, err := client.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err, "Failed to receive player connected message")
	require.NotNil(t, message, "Player connected message should not be nil")
	t.Log("✅ Received player connected confirmation")

	// Extract player ID for later use
	payload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok, "Player connected payload should be a map")
	playerID, ok := payload["playerId"].(string)
	require.True(t, ok, "Player ID should be present in payload")
	client.playerID = playerID
	t.Logf("✅ Player ID assigned: %s", playerID)

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
