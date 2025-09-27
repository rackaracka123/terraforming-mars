package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"terraforming-mars-backend/internal/delivery/dto"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

const (
	// Test timeouts - increased for stability after deadlock fixes
	connectionTimeout = 5 * time.Second
	messageTimeout    = 5 * time.Second
)

var (
	testServer     *TestServer
	testServerHTTP string
	testServerWS   string
	setupMu        sync.Mutex
	serverStarted  bool
)

// getFreePort returns an available port number
func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// TestMain sets up and tears down the test server
func TestMain(m *testing.M) {
	// Setup test server using shared function
	if err := setupTestServer(); err != nil {
		fmt.Printf("Failed to setup test server: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Test server started at %s\n", testServerHTTP)

	// Run tests
	code := m.Run()

	// Stop test server
	if testServer != nil {
		testServer.Stop()
	}

	os.Exit(code)
}

// setupTestServer ensures the test server is running (thread-safe)
func setupTestServer() error {
	setupMu.Lock()
	defer setupMu.Unlock()

	// If server is already running, return
	if serverStarted && testServer != nil {
		return nil
	}

	// Get a free port
	port, err := getFreePort()
	if err != nil {
		return fmt.Errorf("failed to get free port: %w", err)
	}

	// Create test server on dynamic port
	testServer, err = NewTestServer(port)
	if err != nil {
		return fmt.Errorf("failed to create test server on port %d: %w", port, err)
	}

	// Start test server
	err = testServer.Start()
	if err != nil {
		return fmt.Errorf("failed to start test server on port %d: %w", port, err)
	}

	// Set test server URLs
	testServerHTTP = testServer.GetBaseURL()
	testServerWS = testServer.GetWebSocketURL()
	serverStarted = true

	return nil
}

// TestClient represents a test client for integration tests
type TestClient struct {
	conn     *websocket.Conn
	playerID string
	gameID   string
	messages chan dto.WebSocketMessage
	done     chan struct{}
	t        *testing.T
	writeMu  sync.Mutex // Protect WebSocket writes from concurrent access
	closed   bool       // Track if client has been closed
	mu       sync.Mutex // Protect client state
}

// NewTestClient creates a new test client
func NewTestClient(t *testing.T) *TestClient {
	return &TestClient{
		t:        t,
		messages: make(chan dto.WebSocketMessage, 10),
		done:     make(chan struct{}),
	}
}

// Connect establishes WebSocket connection to the test server with retry logic
func (c *TestClient) Connect() error {
	// Ensure test server is running
	if err := setupTestServer(); err != nil {
		return fmt.Errorf("failed to setup test server: %w", err)
	}

	u, err := url.Parse(testServerWS)
	if err != nil {
		return fmt.Errorf("failed to parse WebSocket URL %s: %w", testServerWS, err)
	}

	dialer := &websocket.Dialer{
		HandshakeTimeout: connectionTimeout,
	}

	// Retry connection with exponential backoff
	maxRetries := 3
	baseDelay := 100 * time.Millisecond
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		c.conn, _, err = dialer.Dial(u.String(), nil)
		if err == nil {
			// Connection successful
			go c.readMessages()
			return nil
		}

		lastErr = err
		if attempt < maxRetries {
			// Calculate delay with exponential backoff
			delay := time.Duration(1<<uint(attempt)) * baseDelay
			if delay > 2*time.Second {
				delay = 2 * time.Second
			}
			c.t.Logf("WebSocket connection attempt %d/%d failed: %v. Retrying in %v...",
				attempt+1, maxRetries+1, err, delay)
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("failed to connect to WebSocket after %d attempts: %w", maxRetries+1, lastErr)
}

// readMessages reads messages from WebSocket and forwards to channel
func (c *TestClient) readMessages() {
	defer close(c.messages)

	for {
		select {
		case <-c.done:
			return
		default:
			// Check if connection is closed before trying to read
			c.mu.Lock()
			closed := c.closed
			conn := c.conn
			c.mu.Unlock()

			if closed || conn == nil {
				return
			}

			var message dto.WebSocketMessage
			if err := conn.ReadJSON(&message); err != nil {
				// Check if client is closed before logging
				c.mu.Lock()
				closed = c.closed
				c.mu.Unlock()

				if !closed && websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNormalClosure) {
					c.t.Logf("WebSocket read error: %v", err)
				}
				return
			}

			// Check if client is closed before processing
			c.mu.Lock()
			closed = c.closed
			c.mu.Unlock()

			if closed {
				return
			}

			c.t.Logf("Received message: %s", message.Type)

			// Try to send message with timeout
			select {
			case c.messages <- message:
				// Message sent successfully
			case <-c.done:
				return
			case <-time.After(50 * time.Millisecond):
				// Channel might be full, try to drop oldest and retry once
				select {
				case <-c.messages:
					// Dropped one message
				default:
				}
				select {
				case c.messages <- message:
					// Sent after making space
				case <-c.done:
					return
				default:
					c.t.Logf("Warning: Dropping message due to full channel: %s", message.Type)
				}
			}
		}
	}
}

// Close closes the WebSocket connection and stops the client
func (c *TestClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed || c.conn == nil {
		return
	}

	c.closed = true

	// Close done channel to signal goroutines to stop
	close(c.done)

	// Send close message
	c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))

	// Close the connection
	c.conn.Close()
	c.conn = nil
}

// CreateGameViaHTTP creates a game using the HTTP API
func (c *TestClient) CreateGameViaHTTP() (string, error) {
	// Ensure test server is running
	if err := setupTestServer(); err != nil {
		return "", fmt.Errorf("failed to setup test server: %w", err)
	}

	// Create request payload
	requestBody := dto.CreateGameRequest{
		MaxPlayers:      4,
		DevelopmentMode: false, // Default to false for tests
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP POST request
	resp, err := http.Post(testServerHTTP+"/api/v1/games", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	// Parse response
	var response dto.CreateGameResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Game.ID, nil
}

// JoinGameViaWebSocket joins a game via WebSocket
func (c *TestClient) JoinGameViaWebSocket(gameID, playerName string) error {
	c.gameID = gameID

	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayerConnect,
		GameID: gameID,
		Payload: dto.PlayerConnectPayload{
			PlayerName: playerName,
			GameID:     gameID,
		},
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteJSON(message)
}

// ReconnectToGame sends a player-connect message via WebSocket (unified connection flow)
func (c *TestClient) ReconnectToGame(gameID, playerName string) error {
	c.gameID = gameID

	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayerConnect,
		GameID: gameID,
		Payload: dto.PlayerConnectPayload{
			PlayerName: playerName,
			GameID:     gameID,
		},
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteJSON(message)
}

// StartGame sends a start-game action (host only)
func (c *TestClient) StartGame() error {
	message := dto.WebSocketMessage{
		Type:    dto.MessageTypeActionStartGame,
		GameID:  c.gameID,
		Payload: map[string]interface{}{},
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.conn.WriteJSON(message)
}

// SelectStartingCards sends a select-starting-card action
func (c *TestClient) SelectStartingCards(cardIDs []string) error {
	message := dto.WebSocketMessage{
		Type:    dto.MessageTypeActionSelectStartingCard,
		GameID:  c.gameID,
		Payload: map[string]interface{}{"cardIds": cardIDs},
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.conn.WriteJSON(message)
}

// LaunchAsteroid sends a launch-asteroid action (standard project)
func (c *TestClient) LaunchAsteroid() error {
	message := dto.WebSocketMessage{
		Type:    dto.MessageTypeActionLaunchAsteroid,
		GameID:  c.gameID,
		Payload: map[string]interface{}{},
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.conn.WriteJSON(message)
}

// PlayCard sends a play-card action
func (c *TestClient) PlayCard(cardID string) error {
	message := dto.WebSocketMessage{
		Type:    dto.MessageTypeActionPlayCard,
		GameID:  c.gameID,
		Payload: map[string]interface{}{"cardId": cardID},
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.conn.WriteJSON(message)
}

// WaitForMessage waits for a specific message type with timeout
func (c *TestClient) WaitForMessage(messageType dto.MessageType) (*dto.WebSocketMessage, error) {
	timeout := time.After(messageTimeout)

	for {
		select {
		case message := <-c.messages:
			c.t.Logf("Looking for %s, got %s", messageType, message.Type)
			if message.Type == messageType {
				return &message, nil
			}
			// Continue waiting for the expected message type
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for message type: %s", messageType)
		case <-c.done:
			return nil, fmt.Errorf("client closed while waiting for message")
		}
	}
}

// WaitForMessageTypes waits for any of the specified message types with timeout
func (c *TestClient) WaitForMessageTypes(messageTypes ...dto.MessageType) (*dto.WebSocketMessage, error) {
	timeout := time.After(messageTimeout)

	// Create a map for quick lookup
	typeMap := make(map[dto.MessageType]bool)
	for _, msgType := range messageTypes {
		typeMap[msgType] = true
	}

	for {
		select {
		case message := <-c.messages:
			c.t.Logf("Looking for %v, got %s", messageTypes, message.Type)
			if typeMap[message.Type] {
				return &message, nil
			}
			// Continue waiting for one of the expected message types
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for message types: %v", messageTypes)
		case <-c.done:
			return nil, fmt.Errorf("client closed while waiting for message")
		}
	}
}

// WaitForBothMessages waits for both player-connected and game-updated messages
func (c *TestClient) WaitForBothMessages() (playerConnected, gameUpdated *dto.WebSocketMessage, err error) {
	timeout := time.After(messageTimeout * 2) // Double timeout since we need 2 messages
	messagesNeeded := 2

	for messagesNeeded > 0 {
		select {
		case message := <-c.messages:
			c.t.Logf("Received message: %s", message.Type)
			switch message.Type {
			case dto.MessageTypePlayerConnected:
				if playerConnected == nil {
					playerConnected = &message
					messagesNeeded--
				}
			case dto.MessageTypeGameUpdated:
				if gameUpdated == nil {
					gameUpdated = &message
					messagesNeeded--
				}
			}
		case <-timeout:
			return nil, nil, fmt.Errorf("timeout waiting for both player-connected and game-updated messages")
		case <-c.done:
			return nil, nil, fmt.Errorf("client closed while waiting for messages")
		}
	}

	return playerConnected, gameUpdated, nil
}

// WaitForAnyMessage waits for any message with timeout
func (c *TestClient) WaitForAnyMessage() (*dto.WebSocketMessage, error) {
	timeout := time.After(messageTimeout)

	select {
	case message := <-c.messages:
		return &message, nil
	case <-timeout:
		return nil, fmt.Errorf("timeout waiting for any message")
	case <-c.done:
		return nil, fmt.Errorf("client closed while waiting for message")
	}
}

// WaitForAnyMessageWithTimeout waits for any message with a custom timeout
func (c *TestClient) WaitForAnyMessageWithTimeout(timeout time.Duration) (dto.WebSocketMessage, error) {
	timeoutChan := time.After(timeout)

	select {
	case message := <-c.messages:
		return message, nil
	case <-timeoutChan:
		return dto.WebSocketMessage{}, fmt.Errorf("timeout waiting for any message")
	case <-c.done:
		return dto.WebSocketMessage{}, fmt.Errorf("client closed while waiting for message")
	}
}

// WaitForMessageWithTimeout waits for a specific message type with custom timeout
func (c *TestClient) WaitForMessageWithTimeout(messageType dto.MessageType, timeout time.Duration) (*dto.WebSocketMessage, error) {
	timeoutChan := time.After(timeout)

	for {
		select {
		case message := <-c.messages:
			c.t.Logf("Looking for %s, got %s", messageType, message.Type)
			if message.Type == messageType {
				return &message, nil
			}
			// Continue waiting for the expected message type
		case <-timeoutChan:
			return nil, fmt.Errorf("timeout waiting for message type: %s", messageType)
		case <-c.done:
			return nil, fmt.Errorf("client closed while waiting for message")
		}
	}
}

// ClearMessageQueue drains all pending messages from the message channel
func (c *TestClient) ClearMessageQueue() {
	for {
		select {
		case msg := <-c.messages:
			c.t.Logf("Cleared message from queue: %s", msg.Type)
		default:
			// No more messages in queue
			return
		}
	}
}

// SetupBasicGameFlow sets up a basic game flow: connect, create, join
// Returns the client with established connection and joined game
func SetupBasicGameFlow(t *testing.T, playerName string) (*TestClient, string) {
	client := NewTestClient(t)

	// Connect to WebSocket
	err := client.Connect()
	require.NoError(t, err, "Failed to connect to WebSocket server")

	// Create game via HTTP API
	gameID, err := client.CreateGameViaHTTP()
	require.NoError(t, err, "Failed to create game via HTTP")
	require.NotEmpty(t, gameID, "Game ID should not be empty")

	// Join the created game via WebSocket
	err = client.JoinGameViaWebSocket(gameID, playerName)
	require.NoError(t, err, "Failed to join game via WebSocket")

	// Wait for game state confirmation
	message, err := client.WaitForMessage(dto.MessageTypeGameUpdated)
	require.NoError(t, err, "Failed to receive game state message")
	require.NotNil(t, message, "Game state message should not be nil")

	// Extract player ID from game state
	payload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok, "Game updated payload should be a map")
	game, ok := payload["game"].(map[string]interface{})
	require.True(t, ok, "Game should be present in payload")
	currentPlayer, ok := game["currentPlayer"].(map[string]interface{})
	require.True(t, ok, "Current player should be present in game")
	playerID, ok := currentPlayer["id"].(string)
	require.True(t, ok, "Player ID should be present in current player")
	require.NotEmpty(t, playerID, "Player ID should not be empty")
	client.playerID = playerID

	return client, gameID
}

// ForceClose forces connection closure (for testing network interruptions)
func (c *TestClient) ForceClose() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed || c.conn == nil {
		return
	}

	c.closed = true

	// Send abnormal close message
	c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseAbnormalClosure, ""))
	c.conn.Close()
	c.conn = nil
	close(c.done)
}

// VerifyGameStatus extracts and verifies game status from a WebSocket message
func VerifyGameStatus(t *testing.T, message dto.WebSocketMessage, expectedStatus string) {
	payload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok, "Message payload should be a map")
	gameData, ok := payload["game"].(map[string]interface{})
	require.True(t, ok, "Game data should be present in payload")
	status, ok := gameData["status"].(string)
	require.True(t, ok, "Game status should be present")
	require.Equal(t, expectedStatus, status, "Game status should be %s", expectedStatus)
}

// WaitForGameStatusChange waits for a game update message and verifies the status
func (c *TestClient) WaitForGameStatusChange(expectedStatus string) error {
	// Try multiple times to account for race conditions with async events
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		msg, err := c.WaitForMessage(dto.MessageTypeGameUpdated)
		if err != nil {
			return fmt.Errorf("failed to receive game update (attempt %d): %w", i+1, err)
		}

		payload, ok := msg.Payload.(map[string]interface{})
		if !ok {
			return fmt.Errorf("game update payload should be a map")
		}

		gameData, ok := payload["game"].(map[string]interface{})
		if !ok {
			return fmt.Errorf("game data should be present in payload")
		}

		status, ok := gameData["status"].(string)
		if !ok {
			return fmt.Errorf("game status should be present")
		}

		if status == expectedStatus {
			return nil // Success!
		}

		// If this is the last attempt, return error
		if i == maxRetries-1 {
			return fmt.Errorf("expected status %s, got %s after %d attempts", expectedStatus, status, maxRetries)
		}

		// Small delay before trying again
		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("failed to verify status change after %d attempts", maxRetries)
}

// WaitForStartGameComplete waits for StartGame action to complete and verifies game becomes active
func (c *TestClient) WaitForStartGameComplete() error {
	// Wait for game update message
	err := c.WaitForGameStatusChange("active")
	if err != nil {
		return fmt.Errorf("failed to verify game became active: %w", err)
	}

	// Allow extended time for all async operations to complete
	// StartGame triggers multiple async events that need to finish
	time.Sleep(500 * time.Millisecond)

	return nil
}

// SetPlayerID sets the player ID for this client
func (c *TestClient) SetPlayerID(playerID string) {
	c.playerID = playerID
}

// IsHost checks if this client is the host of the current game
func (c *TestClient) IsHost() (bool, error) {
	if c.gameID == "" || c.playerID == "" {
		return false, fmt.Errorf("client not connected to a game")
	}

	// Get game state via HTTP API
	resp, err := http.Get(fmt.Sprintf("%s/api/v1/games/%s", testServerHTTP, c.gameID))
	if err != nil {
		return false, fmt.Errorf("failed to get game state: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to get game state: status %d", resp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return false, fmt.Errorf("failed to decode game response: %w", err)
	}

	// The response is wrapped in a "game" field
	gameData, ok := response["game"].(map[string]interface{})
	if !ok {
		return false, fmt.Errorf("game field not found in response")
	}

	hostPlayerID, ok := gameData["hostPlayerId"].(string)
	if !ok {
		return false, fmt.Errorf("hostPlayerId not found in game data")
	}

	return hostPlayerID == c.playerID, nil
}

// SendRawMessage sends a raw message for protocol testing (bypasses specific action routing)
func (c *TestClient) SendRawMessage(messageType dto.MessageType, payload interface{}) error {
	message := dto.WebSocketMessage{
		Type:    messageType,
		GameID:  c.gameID,
		Payload: payload,
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.conn.WriteJSON(message)
}
