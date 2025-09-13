package middleware_test

import (
	"context"
	"errors"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/delivery/websocket/core"
	websocketmiddleware "terraforming-mars-backend/internal/middleware/websocket"
	"terraforming-mars-backend/test/integration"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestActionMiddlewareIntegration tests that action middleware works with real WebSocket connections
func TestActionMiddlewareIntegration(t *testing.T) {
	// Setup basic game flow using existing test utilities
	client, gameID := integration.SetupBasicGameFlow(t, "TestPlayer")
	defer client.Close()

	// Start the game to transition to action phase
	err := client.StartGame()
	require.NoError(t, err, "Failed to start game")

	// Wait for game to become active
	err = client.WaitForStartGameComplete()
	require.NoError(t, err, "Failed to wait for game start completion")

	// Test: Perform an action that requires middleware validation
	// Use LaunchAsteroid which should be registered with middleware
	err = client.LaunchAsteroid()
	require.NoError(t, err, "Failed to send launch asteroid action")

	// Wait for action to process successfully
	gameUpdateMsg, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game update after action")

	// Verify the action was processed (game state should be updated)
	payload := gameUpdateMsg.Payload.(map[string]interface{})
	gameData := payload["game"].(map[string]interface{})

	// Check that the game state has been updated in some way
	assert.NotNil(t, gameData, "Game data should be present")
	assert.Equal(t, gameID, gameData["id"].(string), "Game ID should match")

	t.Logf("✅ Action middleware successfully processed action in integration test")
}

// TestTurnValidationIntegration tests turn validation with multiple players
func TestTurnValidationIntegration(t *testing.T) {
	// Create first client (host)
	client1, gameID := integration.SetupBasicGameFlow(t, "Alice")
	defer client1.Close()

	// Create second client
	client2 := integration.NewTestClient(t)
	defer client2.Close()

	// Connect second client
	err := client2.Connect()
	require.NoError(t, err, "Failed to connect second client")

	// Join the same game
	err = client2.JoinGameViaWebSocket(gameID, "Bob")
	require.NoError(t, err, "Failed to join game with second client")

	// Wait for both player connections
	_, err = client2.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err, "Second client should receive connection confirmation")

	// Start game (client1 is host)
	err = client1.StartGame()
	require.NoError(t, err, "Failed to start game")

	// Both clients wait for game start
	err = client1.WaitForStartGameComplete()
	require.NoError(t, err, "Client1 failed to see game start")

	// Clear any pending messages from client2
	client2.ClearMessageQueue()

	// Test: Second player tries to act when it's not their turn
	// This should be rejected by turn validation middleware
	err = client2.LaunchAsteroid()
	require.NoError(t, err, "Should be able to send action (rejection happens server-side)")

	// Verify no game update occurs (action should be rejected)
	// Wait a short time to ensure any messages would have arrived
	_, err = client2.WaitForMessageWithTimeout(dto.MessageTypeGameUpdated, 500*time.Millisecond)
	assert.Error(t, err, "Should timeout waiting for game update (action should be rejected)")

	t.Logf("✅ Turn validation middleware successfully prevented wrong player action")
}

// TestMiddlewareUtilities tests the middleware utility functions
func TestMiddlewareUtilities(t *testing.T) {
	var callOrder []string

	// Create test middleware functions
	middleware1 := websocketmiddleware.MiddlewareFunc(func(ctx context.Context, gameID, playerID string, actionRequest interface{}, next core.ActionHandler) error {
		callOrder = append(callOrder, "middleware1-before")
		err := next.Handle(ctx, gameID, playerID, actionRequest)
		callOrder = append(callOrder, "middleware1-after")
		return err
	})

	middleware2 := websocketmiddleware.MiddlewareFunc(func(ctx context.Context, gameID, playerID string, actionRequest interface{}, next core.ActionHandler) error {
		callOrder = append(callOrder, "middleware2-before")
		err := next.Handle(ctx, gameID, playerID, actionRequest)
		callOrder = append(callOrder, "middleware2-after")
		return err
	})

	// Create mock handler
	mockHandler := &MockActionHandler{shouldFail: false}
	mockHandler.onCall = func() {
		callOrder = append(callOrder, "final-handler")
	}

	// Test middleware chaining
	chained := websocketmiddleware.ChainMiddleware(middleware1, middleware2)

	err := chained(context.Background(), "test-game", "test-player", "test-data", mockHandler)

	require.NoError(t, err)
	assert.True(t, mockHandler.called, "Final handler should have been called")

	// Verify middleware execution order
	expectedOrder := []string{
		"middleware1-before",
		"middleware2-before",
		"final-handler",
		"middleware2-after",
		"middleware1-after",
	}

	assert.Equal(t, expectedOrder, callOrder, "Middleware should execute in correct order")
}

// TestMiddlewareErrorPropagation tests error handling through middleware chain
func TestMiddlewareErrorPropagation(t *testing.T) {
	testError := errors.New("test error")

	middleware1 := websocketmiddleware.MiddlewareFunc(func(ctx context.Context, gameID, playerID string, actionRequest interface{}, next core.ActionHandler) error {
		return next.Handle(ctx, gameID, playerID, actionRequest)
	})

	// Create failing handler
	failingHandler := &MockActionHandler{shouldFail: true, errorToReturn: testError}

	err := middleware1(context.Background(), "test-game", "test-player", "test-data", failingHandler)

	require.Error(t, err)
	assert.Equal(t, testError, err, "Error should propagate through middleware")
	assert.True(t, failingHandler.called, "Handler should have been called")
}

// TestWrapWithMiddleware tests wrapping a handler with middleware
func TestWrapWithMiddleware(t *testing.T) {
	middlewareCalled := false

	testMiddleware := websocketmiddleware.MiddlewareFunc(func(ctx context.Context, gameID, playerID string, actionRequest interface{}, next core.ActionHandler) error {
		middlewareCalled = true
		assert.Equal(t, "test-game", gameID)
		assert.Equal(t, "test-player", playerID)
		return next.Handle(ctx, gameID, playerID, actionRequest)
	})

	originalHandler := &MockActionHandler{shouldFail: false}
	wrappedHandler := websocketmiddleware.WrapWithMiddleware(originalHandler, testMiddleware)

	err := wrappedHandler.Handle(context.Background(), "test-game", "test-player", "test-data")

	assert.NoError(t, err)
	assert.True(t, middlewareCalled, "Middleware should have been called")
	assert.True(t, originalHandler.called, "Original handler should have been called")
}

// MockActionHandler for testing
type MockActionHandler struct {
	called        bool
	shouldFail    bool
	errorToReturn error
	onCall        func()
}

func (h *MockActionHandler) Handle(ctx context.Context, gameID, playerID string, actionRequest interface{}) error {
	h.called = true

	if h.onCall != nil {
		h.onCall()
	}

	if h.shouldFail {
		if h.errorToReturn != nil {
			return h.errorToReturn
		}
		return errors.New("mock handler failed")
	}

	return nil
}
