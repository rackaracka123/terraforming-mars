package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	// Test server configuration
	testPort = 3002 // Use different port to avoid conflicts
	
	// Test timeouts
	connectionTimeout = 5 * time.Second
	messageTimeout    = 3 * time.Second
)

var (
	testServer *TestServer
	testServerHTTP string
	testServerWS   string
	setupOnce sync.Once
)

// TestMain sets up and tears down the test server
func TestMain(m *testing.M) {
	// Create test server
	var err error
	testServer, err = NewTestServer(testPort)
	if err != nil {
		fmt.Printf("Failed to create test server: %v\n", err)
		os.Exit(1)
	}

	// Start test server
	err = testServer.Start()
	if err != nil {
		fmt.Printf("Failed to start test server: %v\n", err)
		os.Exit(1)
	}

	// Set test server URLs
	testServerHTTP = testServer.GetBaseURL()
	testServerWS = testServer.GetWebSocketURL()

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
	var setupErr error
	setupOnce.Do(func() {
		// Create test server
		var err error
		testServer, err = NewTestServer(testPort)
		if err != nil {
			setupErr = fmt.Errorf("failed to create test server: %w", err)
			return
		}

		// Start test server
		err = testServer.Start()
		if err != nil {
			setupErr = fmt.Errorf("failed to start test server: %w", err)
			return
		}

		// Set test server URLs
		testServerHTTP = testServer.GetBaseURL()
		testServerWS = testServer.GetWebSocketURL()
	})
	return setupErr
}

// TestClient represents a test client for integration tests
type TestClient struct {
	conn     *websocket.Conn
	playerID string
	gameID   string
	messages chan dto.WebSocketMessage
	done     chan struct{}
	t        *testing.T
}

// NewTestClient creates a new test client
func NewTestClient(t *testing.T) *TestClient {
	return &TestClient{
		t:        t,
		messages: make(chan dto.WebSocketMessage, 10),
		done:     make(chan struct{}),
	}
}

// Connect establishes WebSocket connection to the test server
func (c *TestClient) Connect() error {
	// Ensure test server is running
	if err := setupTestServer(); err != nil {
		return fmt.Errorf("failed to setup test server: %w", err)
	}
	
	u, err := url.Parse(testServerWS)
	if err != nil {
		return fmt.Errorf("failed to parse WebSocket URL %s: %w", testServerWS, err)
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = connectionTimeout

	c.conn, _, err = dialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	// Start message reader
	go c.readMessages()

	return nil
}

// readMessages reads messages from WebSocket and forwards to channel
func (c *TestClient) readMessages() {
	defer close(c.messages)
	
	for {
		select {
		case <-c.done:
			return
		default:
			var message dto.WebSocketMessage
			if err := c.conn.ReadJSON(&message); err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.t.Logf("WebSocket read error: %v", err)
				}
				return
			}
			
			c.t.Logf("Received message: %s", message.Type)
			c.messages <- message
		}
	}
}

// Close closes the WebSocket connection and stops the client
func (c *TestClient) Close() {
	if c.conn != nil {
		close(c.done)
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
	}
}

// CreateGameViaHTTP creates a game using the HTTP API
func (c *TestClient) CreateGameViaHTTP() (string, error) {
	// Ensure test server is running
	if err := setupTestServer(); err != nil {
		return "", fmt.Errorf("failed to setup test server: %w", err)
	}
	
	// Create request payload
	requestBody := dto.CreateGameRequest{
		MaxPlayers: 4,
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

	return c.conn.WriteJSON(message)
}

// SendAction sends a game action via WebSocket using proper DTOs
func (c *TestClient) SendAction(action interface{}) error {
	message := dto.WebSocketMessage{
		Type:   dto.MessageTypePlayAction,
		GameID: c.gameID,
		Payload: dto.PlayActionPayload{
			ActionRequest: action,
		},
	}

	return c.conn.WriteJSON(message)
}

// StartGame sends a start-game action (host only)
func (c *TestClient) StartGame() error {
	// Use the same format as the CLI client
	action := map[string]interface{}{
		"type": string(dto.ActionTypeStartGame),
	}
	return c.SendAction(action)
}

// SelectStartingCards sends a select-starting-card action
func (c *TestClient) SelectStartingCards(cardIDs []string) error {
	action := map[string]interface{}{
		"type":    string(dto.ActionTypeSelectStartingCard),
		"cardIds": cardIDs,
	}
	return c.SendAction(action)
}

// LaunchAsteroid sends a launch-asteroid action (standard project)
func (c *TestClient) LaunchAsteroid() error {
	action := map[string]interface{}{
		"type": string(dto.ActionTypeLaunchAsteroid),
	}
	return c.SendAction(action)
}

// PlayCard sends a play-card action
func (c *TestClient) PlayCard(cardID string) error {
	action := map[string]interface{}{
		"type":   string(dto.ActionTypePlayCard),
		"cardId": cardID,
	}
	return c.SendAction(action)
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
	
	// Wait for player connected confirmation
	message, err := client.WaitForMessage(dto.MessageTypePlayerConnected)
	require.NoError(t, err, "Failed to receive player connected message")
	require.NotNil(t, message, "Player connected message should not be nil")
	
	// Extract player ID
	payload, ok := message.Payload.(map[string]interface{})
	require.True(t, ok, "Player connected payload should be a map")
	playerID, ok := payload["playerId"].(string)
	require.True(t, ok, "Player ID should be present in payload")
	require.NotEmpty(t, playerID, "Player ID should not be empty")
	client.playerID = playerID
	
	return client, gameID
}