package websocket

import (
	"fmt"
	"testing"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/test/integration"

	"github.com/stretchr/testify/require"
)

// TestMultipleClientsJoinSameGame tests multiple clients joining the same game (simplified)
func TestMultipleClientsJoinSameGame(t *testing.T) {
	integration.VerifyTestIsolation(t)

	const numClients = 2 // Reduced for faster execution
	clients := make([]*integration.TestClient, numClients)

	// Create and connect clients
	for i := 0; i < numClients; i++ {
		clients[i] = integration.CreateTestClientWithCleanup(t)

		err := clients[i].Connect()
		require.NoError(t, err, "Client %d should connect", i)
		require.True(t, clients[i].IsConnected(), "Client %d should be connected", i)
	}

	// Ensure cleanup happens
	defer integration.CleanupTestClients(t, clients...)

	// Create game
	gameID, err := clients[0].CreateGameViaHTTP()
	require.NoError(t, err, "Should create game")

	// Join game sequentially for reliability
	for i, client := range clients {
		playerName := fmt.Sprintf("Player%d", i+1)
		err := client.JoinGameViaWebSocket(gameID, playerName)
		require.NoError(t, err, "Client %d should join game", i)

		// Wait for confirmation
		_, err = client.WaitForMessage(dto.MessageTypePlayerConnected)
		require.NoError(t, err, "Client %d should get join confirmation", i)
	}

	t.Logf("✅ All %d clients joined the same game successfully", numClients)
}

// TestConcurrentMessageBroadcasting tests broadcasting messages to multiple clients (simplified)
func TestConcurrentMessageBroadcasting(t *testing.T) {
	integration.VerifyTestIsolation(t)

	const numClients = 2 // Reduced for faster execution
	clients := make([]*integration.TestClient, numClients)

	// Setup clients and game
	for i := 0; i < numClients; i++ {
		clients[i] = integration.CreateTestClientWithCleanup(t)

		err := clients[i].Connect()
		require.NoError(t, err, "Client %d should connect", i)
	}

	// Ensure cleanup happens
	defer integration.CleanupTestClients(t, clients...)

	// Create and join game
	gameID, err := clients[0].CreateGameViaHTTP()
	require.NoError(t, err)

	for i, client := range clients {
		playerName := fmt.Sprintf("Player%d", i+1)
		err := client.JoinGameViaWebSocket(gameID, playerName)
		require.NoError(t, err)

		_, err = client.WaitForMessage(dto.MessageTypePlayerConnected)
		require.NoError(t, err)
	}

	// Test broadcasting by starting game (triggers game-updated broadcast)
	err = clients[0].StartGame()
	require.NoError(t, err, "Host should be able to start game")

	// Verify all clients receive broadcast
	for i, client := range clients {
		_, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
		require.NoError(t, err, "Client %d should receive game update broadcast", i)
	}

	t.Log("✅ Broadcasting to multiple clients works correctly")
}

// TestClientDisconnectionHandling tests how disconnections are handled with multiple clients
func TestClientDisconnectionHandling(t *testing.T) {
	integration.VerifyTestIsolation(t)

	client1 := integration.CreateTestClientWithCleanup(t)
	client2 := integration.CreateTestClientWithCleanup(t)

	// Setup
	err := client1.Connect()
	require.NoError(t, err)
	err = client2.Connect()
	require.NoError(t, err)

	gameID, err := client1.CreateGameViaHTTP()
	require.NoError(t, err)

	// Join game
	err = client1.JoinGameViaWebSocket(gameID, "Player1")
	require.NoError(t, err)
	err = client2.JoinGameViaWebSocket(gameID, "Player2")
	require.NoError(t, err)

	// Get confirmations
	_, err = client1.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err)
	_, err = client2.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err)

	// Disconnect one client
	client1.ForceClose()

	// Other client should be notified
	_, err = client2.WaitForMessage(dto.MessageTypePlayerDisconnected)
	require.NoError(t, err, "Remaining client should be notified of disconnection")

	t.Log("✅ Client disconnection handling works correctly")
}
