package websocket

import (
	"testing"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/test/integration"

	"github.com/stretchr/testify/require"
)

// TestHostLobbyReconnectionBug tests the specific bug where host reconnection
// causes other players to disappear from host's view
func TestHostLobbyReconnectionBug(t *testing.T) {
	t.Log("🎯 Testing host lobby reconnection bug fix")

	// === SETUP: Create host and player2 in lobby ===
	hostClient := integration.NewTestClient(t)
	player2Client := integration.NewTestClient(t)
	defer hostClient.Close()
	defer player2Client.Close()

	// Connect both clients
	err := hostClient.Connect()
	require.NoError(t, err, "Host should connect")
	err = player2Client.Connect()
	require.NoError(t, err, "Player 2 should connect")

	// Create game and join players
	gameID, err := hostClient.CreateGameViaHTTP()
	require.NoError(t, err, "Should create game")
	t.Logf("✅ Game created: %s", gameID)

	// Host joins first (becomes host)
	err = hostClient.JoinGameViaWebSocket(gameID, "Host")
	require.NoError(t, err, "Host should join")

	hostConnectedMsg, err := hostClient.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err, "Host should receive connection confirmation")

	// Extract host player ID
	hostPayload, ok := hostConnectedMsg.Payload.(map[string]interface{})
	require.True(t, ok, "Host payload should be map")
	hostPlayerID, ok := hostPayload["playerId"].(string)
	require.True(t, ok, "Host player ID should be present")
	hostClient.SetPlayerID(hostPlayerID)
	t.Logf("✅ Host connected with ID: %s", hostPlayerID)

	// Player 2 joins
	err = player2Client.JoinGameViaWebSocket(gameID, "Player2")
	require.NoError(t, err, "Player 2 should join")

	player2ConnectedMsg, err := player2Client.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err, "Player 2 should receive connection confirmation")

	// Extract player 2 ID
	player2Payload, ok := player2ConnectedMsg.Payload.(map[string]interface{})
	require.True(t, ok, "Player 2 payload should be map")
	player2PlayerID, ok := player2Payload["playerId"].(string)
	require.True(t, ok, "Player 2 ID should be present")
	player2Client.SetPlayerID(player2PlayerID)
	t.Logf("✅ Player 2 connected with ID: %s", player2PlayerID)

	// === VERIFY INITIAL STATE: Both players see each other ===
	// Wait for all connection messages to be processed
	time.Sleep(100 * time.Millisecond)

	// Get the latest game-updated message for each client (should be the last message received)
	var initialHostGameState map[string]interface{}
	var initialPlayer2GameState map[string]interface{}

	// Process all remaining messages and keep the latest game-updated
	for {
		msg, err := hostClient.WaitForAnyMessageWithTimeout(50 * time.Millisecond)
		if err != nil {
			break
		}
		if msg.Type == dto.MessageTypeGameUpdated {
			if payload, ok := msg.Payload.(map[string]interface{}); ok {
				if gameData, ok := payload["game"].(map[string]interface{}); ok {
					initialHostGameState = gameData
				}
			}
		}
	}

	for {
		msg, err := player2Client.WaitForAnyMessageWithTimeout(50 * time.Millisecond)
		if err != nil {
			break
		}
		if msg.Type == dto.MessageTypeGameUpdated {
			if payload, ok := msg.Payload.(map[string]interface{}); ok {
				if gameData, ok := payload["game"].(map[string]interface{}); ok {
					initialPlayer2GameState = gameData
				}
			}
		}
	}

	require.NotNil(t, initialHostGameState, "Host should have game state")
	require.NotNil(t, initialPlayer2GameState, "Player 2 should have game state")

	hostPlayerCount := CountPlayersInGameState(t, initialHostGameState)
	player2PlayerCount := CountPlayersInGameState(t, initialPlayer2GameState)

	require.Equal(t, 2, hostPlayerCount, "Host should see 2 players initially")
	require.Equal(t, 2, player2PlayerCount, "Player 2 should see 2 players initially")
	t.Logf("✅ Both clients see 2 players initially")

	// === SIMULATE HOST REFRESH/RECONNECTION ===
	t.Log("🔄 Simulating host refresh by disconnecting and reconnecting...")

	// Disconnect host (simulating browser refresh)
	hostClient.ForceClose()

	// Create new client for host reconnection
	reconnectHostClient := integration.NewTestClient(t)
	defer reconnectHostClient.Close()

	err = reconnectHostClient.Connect()
	require.NoError(t, err, "Host reconnect client should connect")

	// Host reconnects using player-reconnect message (as frontend does on refresh)
	err = reconnectHostClient.ReconnectToGame(gameID, "Host")
	require.NoError(t, err, "Host should reconnect to game")

	// Wait for reconnection confirmation
	reconnectedMsg, err := reconnectHostClient.WaitForMessage(dto.MessageTypePlayerReconnected)
	require.NoError(t, err, "Host should receive reconnection confirmation")
	t.Log("✅ Host reconnected successfully")

	// Extract reconnected host player ID and verify it matches
	reconnectPayload, ok := reconnectedMsg.Payload.(map[string]interface{})
	require.True(t, ok, "Reconnect payload should be map")
	reconnectedPlayerID, ok := reconnectPayload["playerId"].(string)
	require.True(t, ok, "Reconnected player ID should be present")
	reconnectHostClient.SetPlayerID(reconnectedPlayerID)
	require.Equal(t, hostPlayerID, reconnectedPlayerID, "Host player ID should be preserved")

	// === VERIFY BUG FIX: Both players still see each other after host reconnection ===

	// Get current state from both clients
	reconnectedHostGameState, ok := reconnectPayload["game"].(map[string]interface{})
	require.True(t, ok, "Reconnected game state should be present")

	// Get fresh state from player 2 - wait for game-updated message after host reconnection
	var currentPlayer2State map[string]interface{}
	player2GameUpdateMsg, err := player2Client.WaitForMessage(dto.MessageTypeGameUpdated)
	if err != nil {
		// If no game-updated message, use the initial state (lobby hasn't changed)
		t.Log("No game-updated for player2, using initial state")
		currentPlayer2State = initialPlayer2GameState
	} else {
		player2UpdatePayload, ok := player2GameUpdateMsg.Payload.(map[string]interface{})
		require.True(t, ok, "Player2 update payload should be map")
		currentPlayer2State, ok = player2UpdatePayload["game"].(map[string]interface{})
		require.True(t, ok, "Player2 should have updated game state")
	}
	require.NotNil(t, currentPlayer2State, "Player 2 should have current state")

	// Both clients should see 2 players (this is the main bug fix verification)
	reconnectedHostPlayerCount := CountPlayersInGameState(t, reconnectedHostGameState)
	currentPlayer2PlayerCount := CountPlayersInGameState(t, currentPlayer2State)

	require.Equal(t, 2, reconnectedHostPlayerCount,
		"BUG FIX: Host should still see 2 players after reconnection (was seeing fewer)")
	require.Equal(t, 2, currentPlayer2PlayerCount,
		"Player 2 should still see 2 players after host reconnection")

	t.Logf("✅ BUG FIX VERIFIED: Host sees %d players, Player2 sees %d players",
		reconnectedHostPlayerCount, currentPlayer2PlayerCount)

	// === VERIFY STATE CONSISTENCY ===

	// Both clients should see the same game status
	reconnectedStatus := ExtractGameStatus(t, reconnectedHostGameState)
	player2Status := ExtractGameStatus(t, currentPlayer2State)
	require.Equal(t, "lobby", reconnectedStatus, "Game should still be in lobby")
	require.Equal(t, reconnectedStatus, player2Status, "Both clients should see same status")

	// Both clients should see the same set of players
	hostPlayerIDs := ExtractPlayerIDs(t, reconnectedHostGameState)
	player2PlayerIDs := ExtractPlayerIDs(t, currentPlayer2State)
	require.ElementsMatch(t, hostPlayerIDs, player2PlayerIDs,
		"Both clients should see identical player IDs")

	// Host should still be the host
	hostGameHostID, ok := reconnectedHostGameState["hostPlayerId"].(string)
	require.True(t, ok, "Host player ID should be present in game state")
	require.Equal(t, hostPlayerID, hostGameHostID, "Host player should still be the host")

	// Quality verification
	VerifyGameStateQuality(t, reconnectedHostGameState, "ReconnectedHost", 2)
	VerifyGameStateQuality(t, currentPlayer2State, "Player2", 2)

	// === SUCCESS ===
	t.Log("🎉 Host lobby reconnection bug fix test completed successfully!")
	t.Log("✅ Host reconnection maintains visibility of all players")
	t.Log("✅ State consistency preserved across all clients")
	t.Log("✅ Host privileges maintained after reconnection")
}

// TestPlayer2LobbyReconnectionAsymmetry tests that player2 reconnection works correctly
// and doesn't have the same bug (this should have always worked)
func TestPlayer2LobbyReconnectionAsymmetry(t *testing.T) {
	t.Log("🎯 Testing player2 lobby reconnection (should work correctly)")

	// === SETUP: Create host and player2 in lobby ===
	hostClient := integration.NewTestClient(t)
	player2Client := integration.NewTestClient(t)
	defer hostClient.Close()
	defer player2Client.Close()

	// Connect both clients
	err := hostClient.Connect()
	require.NoError(t, err, "Host should connect")
	err = player2Client.Connect()
	require.NoError(t, err, "Player 2 should connect")

	// Create game and join players
	gameID, err := hostClient.CreateGameViaHTTP()
	require.NoError(t, err, "Should create game")

	// Host joins first
	err = hostClient.JoinGameViaWebSocket(gameID, "Host")
	require.NoError(t, err, "Host should join")
	hostConnectedMsg, err := hostClient.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err, "Host should receive connection confirmation")
	hostPayload, _ := hostConnectedMsg.Payload.(map[string]interface{})
	hostPlayerID, _ := hostPayload["playerId"].(string)
	hostClient.SetPlayerID(hostPlayerID)

	// Player 2 joins
	err = player2Client.JoinGameViaWebSocket(gameID, "Player2")
	require.NoError(t, err, "Player 2 should join")
	player2ConnectedMsg, err := player2Client.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err, "Player 2 should receive connection confirmation")
	player2Payload, _ := player2ConnectedMsg.Payload.(map[string]interface{})
	player2PlayerID, _ := player2Payload["playerId"].(string)
	player2Client.SetPlayerID(player2PlayerID)

	// Clear additional messages
	_, _ = hostClient.WaitForAnyMessageWithTimeout(100 * time.Millisecond)
	_, _ = player2Client.WaitForAnyMessageWithTimeout(100 * time.Millisecond)

	// === SIMULATE PLAYER2 REFRESH/RECONNECTION ===
	t.Log("🔄 Simulating player2 refresh...")

	// Disconnect player2
	player2Client.ForceClose()

	// Create new client for player2 reconnection
	reconnectPlayer2Client := integration.NewTestClient(t)
	defer reconnectPlayer2Client.Close()

	err = reconnectPlayer2Client.Connect()
	require.NoError(t, err, "Player2 reconnect client should connect")

	err = reconnectPlayer2Client.ReconnectToGame(gameID, "Player2")
	require.NoError(t, err, "Player2 should reconnect to game")

	reconnectedMsg, err := reconnectPlayer2Client.WaitForMessage(dto.MessageTypePlayerReconnected)
	require.NoError(t, err, "Player2 should receive reconnection confirmation")

	// === VERIFY: Both clients see each other after player2 reconnection ===
	reconnectPayload, _ := reconnectedMsg.Payload.(map[string]interface{})
	reconnectedPlayer2GameState, _ := reconnectPayload["game"].(map[string]interface{})
	reconnectPlayer2Client.SetPlayerID(player2PlayerID)

	currentHostState := GetGameStateFromClient(t, hostClient)

	reconnectedPlayer2PlayerCount := CountPlayersInGameState(t, reconnectedPlayer2GameState)
	currentHostPlayerCount := CountPlayersInGameState(t, currentHostState)

	require.Equal(t, 2, reconnectedPlayer2PlayerCount, "Player2 should see 2 players after reconnection")
	require.Equal(t, 2, currentHostPlayerCount, "Host should still see 2 players after player2 reconnection")

	t.Logf("✅ Player2 reconnection works correctly: Player2 sees %d players, Host sees %d players",
		reconnectedPlayer2PlayerCount, currentHostPlayerCount)

	t.Log("🎉 Player2 reconnection test completed - no asymmetry issues detected")
}
