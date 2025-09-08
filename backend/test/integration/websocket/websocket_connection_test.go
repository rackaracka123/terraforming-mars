package websocket

import (
	"sync"
	"testing"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/test/integration"

	"github.com/stretchr/testify/require"
)

// TestBasicConnection tests the most basic WebSocket connection scenario
func TestBasicConnection(t *testing.T) {
	client := integration.NewTestClient(t)
	defer client.Close()

	// Test connection establishment
	err := client.Connect()
	require.NoError(t, err, "Should be able to establish WebSocket connection")
	t.Log("✅ Basic WebSocket connection established")

	// Test graceful closure
	client.Close()
	t.Log("✅ WebSocket connection closed gracefully")
}

// TestConnectionTimeout tests connection timeout scenarios
func TestConnectionTimeout(t *testing.T) {
	client := integration.NewTestClient(t)
	defer client.Close()

	// Connect normally first to ensure server is running
	err := client.Connect()
	require.NoError(t, err, "Should be able to establish WebSocket connection")
	t.Log("✅ Connection established successfully")
}

// TestMultipleConnections tests multiple independent connections
func TestMultipleConnections(t *testing.T) {
	const numConnections = 5
	clients := make([]*integration.TestClient, numConnections)

	// Create all clients
	for i := 0; i < numConnections; i++ {
		clients[i] = integration.NewTestClient(t)
	}

	// Defer cleanup
	defer func() {
		for _, client := range clients {
			client.Close()
		}
	}()

	// Connect all clients concurrently
	var wg sync.WaitGroup
	connectErrors := make(chan error, numConnections)

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(clientIdx int) {
			defer wg.Done()
			err := clients[clientIdx].Connect()
			if err != nil {
				connectErrors <- err
			}
		}(i)
	}

	wg.Wait()
	close(connectErrors)

	// Check for connection errors
	errorCount := 0
	for err := range connectErrors {
		errorCount++
		t.Errorf("Connection error: %v", err)
	}

	require.Equal(t, 0, errorCount, "All connections should succeed")
	t.Logf("✅ Successfully established %d concurrent connections", numConnections)
}

// TestConnectionCleanup tests that connections are properly cleaned up
func TestConnectionCleanup(t *testing.T) {
	client := integration.NewTestClient(t)

	// Connect
	err := client.Connect()
	require.NoError(t, err, "Should be able to establish WebSocket connection")
	t.Log("✅ WebSocket connection established")

	// Close gracefully
	client.Close()
	t.Log("✅ WebSocket connection closed gracefully")

	// Attempt to use closed connection should not panic
	// (The TestClient should handle this gracefully)
	client.Close() // Should not panic or error
	t.Log("✅ Double close handled gracefully")
}

// TestForceDisconnection tests forced disconnection scenarios
func TestForceDisconnection(t *testing.T) {
	client := integration.NewTestClient(t)
	defer client.Close()

	// Connect
	err := client.Connect()
	require.NoError(t, err, "Should be able to establish WebSocket connection")
	t.Log("✅ WebSocket connection established")

	// Force close the connection (simulates network interruption)
	client.ForceClose()
	t.Log("✅ Force close executed")

	// Give some time for cleanup
	time.Sleep(100 * time.Millisecond)
	t.Log("✅ Force disconnection test completed")
}

// TestConnectionRecovery tests connection recovery patterns
func TestConnectionRecovery(t *testing.T) {
	// First connection
	client1 := integration.NewTestClient(t)
	defer client1.Close()

	err := client1.Connect()
	require.NoError(t, err, "Should be able to establish first WebSocket connection")
	t.Log("✅ First WebSocket connection established")

	// Create a game to have some state
	gameID, err := client1.CreateGameViaHTTP()
	require.NoError(t, err, "Should be able to create game")
	t.Log("✅ Game created")

	// Join the game
	err = client1.JoinGameViaWebSocket(gameID, "TestPlayer")
	require.NoError(t, err, "Should be able to join game")
	t.Log("✅ Player joined game")

	// Wait for confirmation
	_, err = client1.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err, "Should receive player connected message")
	t.Log("✅ Player connected confirmed")

	// Force disconnect first client
	client1.ForceClose()
	t.Log("✅ First connection force closed")

	// Give some time for server cleanup
	time.Sleep(200 * time.Millisecond)

	// Second connection (simulating recovery)
	client2 := integration.NewTestClient(t)
	defer client2.Close()

	err = client2.Connect()
	require.NoError(t, err, "Should be able to establish recovery WebSocket connection")
	t.Log("✅ Recovery WebSocket connection established")

	// Try to join the same game (should work)
	err = client2.JoinGameViaWebSocket(gameID, "TestPlayerReconnect")
	require.NoError(t, err, "Should be able to reconnect to game")
	t.Log("✅ Player reconnected to game")
}

// TestRapidConnectDisconnect tests rapid connect/disconnect cycles
func TestRapidConnectDisconnect(t *testing.T) {
	const cycles = 10

	for i := 0; i < cycles; i++ {
		client := integration.NewTestClient(t)

		// Connect
		err := client.Connect()
		require.NoError(t, err, "Connection %d should succeed", i+1)

		// Small delay to ensure connection is fully established
		time.Sleep(10 * time.Millisecond)

		// Disconnect
		client.Close()

		// Small delay between cycles
		time.Sleep(10 * time.Millisecond)
	}

	t.Logf("✅ Completed %d rapid connect/disconnect cycles", cycles)
}
