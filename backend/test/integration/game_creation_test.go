package integration

import (
	"testing"

	"terraforming-mars-backend/internal/delivery/dto"

	"github.com/stretchr/testify/require"
)

// TestCreateAndJoinGame tests creating a game via HTTP and joining via WebSocket
func TestCreateAndJoinGame(t *testing.T) {
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

	// Step 5: Verify the message payload contains correct data
	payload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok, "Game updated payload should be a map")

	// Step 6: Verify game data is present
	gameData, ok := payload["game"].(map[string]interface{})
	require.True(t, ok, "Game data should be present in payload")

	// Check current player data
	currentPlayer, ok := gameData["currentPlayer"].(map[string]interface{})
	require.True(t, ok, "Current player should be present in game data")

	// Check player name
	actualPlayerName, ok := currentPlayer["name"].(string)
	require.True(t, ok, "Player name should be present in current player")
	require.Equal(t, playerName, actualPlayerName, "Player name should match")

	// Check player ID is present
	playerID, ok := currentPlayer["id"].(string)
	require.True(t, ok, "Player ID should be present in current player")
	require.NotEmpty(t, playerID, "Player ID should not be empty")
	client.playerID = playerID
	t.Logf("✅ Player ID assigned: %s", playerID)
	require.True(t, ok, "Game data should be present in payload")

	// Verify game ID matches
	actualGameID, ok := gameData["id"].(string)
	require.True(t, ok, "Game ID should be present in game data")
	require.Equal(t, gameID, actualGameID, "Game ID should match")

	// Verify game status is lobby
	gameStatus, ok := gameData["status"].(string)
	require.True(t, ok, "Game status should be present")
	require.Equal(t, "lobby", gameStatus, "Game should be in lobby status")

	// Verify player is the host (first player)
	hostPlayerID, ok := gameData["hostPlayerId"].(string)
	require.True(t, ok, "Host player ID should be present")
	require.Equal(t, playerID, hostPlayerID, "Player should be the host")

	t.Log("✅ All basic game creation and joining tests passed!")
}
