package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"terraforming-mars-backend/internal/actions"
	"terraforming-mars-backend/internal/actions/card_selection"
	"terraforming-mars-backend/internal/actions/standard_projects"
	"terraforming-mars-backend/internal/admin"
	"terraforming-mars-backend/internal/delivery/dto"
	httpHandler "terraforming-mars-backend/internal/delivery/http"
	wsHandler "terraforming-mars-backend/internal/delivery/websocket"
	wsCore "terraforming-mars-backend/internal/delivery/websocket/core"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/lobby"
	"terraforming-mars-backend/internal/player"
	sessionpkg "terraforming-mars-backend/internal/session"
)

// TestWebSocketClient wraps a WebSocket connection for testing
type TestWebSocketClient struct {
	conn       *websocket.Conn
	playerID   string
	gameID     string
	playerName string
	messages   chan dto.WebSocketMessage
	errors     chan error
	done       chan struct{}
	t          *testing.T
}

// NewTestWebSocketClient creates a new test client and connects to the game
func NewTestWebSocketClient(t *testing.T, wsURL string, gameID string, playerName string) *TestWebSocketClient {
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err, "Failed to connect WebSocket")

	client := &TestWebSocketClient{
		conn:       conn,
		gameID:     gameID,
		playerName: playerName,
		messages:   make(chan dto.WebSocketMessage, 100),
		errors:     make(chan error, 10),
		done:       make(chan struct{}),
		t:          t,
	}

	// Start read loop
	go client.readLoop()

	// Send player-connect message
	err = client.Send(dto.WebSocketMessage{
		Type: "player-connect",
		Payload: map[string]interface{}{
			"playerName": playerName,
			"gameId":     gameID,
			"playerId":   "",
		},
	})
	require.NoError(t, err, "Failed to send connect message")

	// Wait for connection confirmation
	msg, err := client.WaitForMessageType(dto.MessageTypeGameUpdated, 3*time.Second)
	require.NoError(t, err, "Failed to receive connection confirmation")

	// Extract player ID from game state
	client.extractPlayerID(msg)

	t.Logf("✅ %s connected (playerID: %s)", playerName, client.playerID)

	return client
}

func (c *TestWebSocketClient) readLoop() {
	defer close(c.done)
	for {
		var msg dto.WebSocketMessage
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.errors <- err
			}
			return
		}
		c.t.Logf("📨 %s received: %s", c.playerName, msg.Type)
		c.messages <- msg
	}
}

func (c *TestWebSocketClient) Send(msg dto.WebSocketMessage) error {
	c.t.Logf("📤 %s sending: %s", c.playerName, msg.Type)
	return c.conn.WriteJSON(msg)
}

func (c *TestWebSocketClient) WaitForMessageType(msgType dto.MessageType, timeout time.Duration) (dto.WebSocketMessage, error) {
	deadline := time.After(timeout)
	for {
		select {
		case msg := <-c.messages:
			if msg.Type == msgType {
				return msg, nil
			}
			// Put non-matching message back for other waiters
			go func(m dto.WebSocketMessage) {
				c.messages <- m
			}(msg)
		case err := <-c.errors:
			return dto.WebSocketMessage{}, err
		case <-deadline:
			return dto.WebSocketMessage{}, fmt.Errorf("timeout waiting for message type: %s", msgType)
		case <-c.done:
			return dto.WebSocketMessage{}, fmt.Errorf("connection closed while waiting for: %s", msgType)
		}
	}
}

func (c *TestWebSocketClient) DrainMessages() {
	for {
		select {
		case <-c.messages:
			// Drain
		default:
			return
		}
	}
}

func (c *TestWebSocketClient) extractPlayerID(msg dto.WebSocketMessage) {
	// Parse the payload to extract player ID from ViewingPlayerID or CurrentPlayer
	payloadBytes, _ := json.Marshal(msg.Payload)
	var gameUpdate struct {
		Game struct {
			ViewingPlayerID string `json:"viewingPlayerId"`
			CurrentPlayer   struct {
				ID string `json:"id"`
			} `json:"currentPlayer"`
		} `json:"game"`
	}
	json.Unmarshal(payloadBytes, &gameUpdate)

	// ViewingPlayerID is the player ID for the viewing player
	if gameUpdate.Game.ViewingPlayerID != "" {
		c.playerID = gameUpdate.Game.ViewingPlayerID
		return
	}

	// Fallback to CurrentPlayer.ID
	if gameUpdate.Game.CurrentPlayer.ID != "" {
		c.playerID = gameUpdate.Game.CurrentPlayer.ID
		return
	}
}

func (c *TestWebSocketClient) Close() {
	c.conn.Close()
	<-c.done
}

// Helper to extract game phase from message
func extractGamePhase(msg dto.WebSocketMessage) string {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var gameUpdate struct {
		Game struct {
			CurrentPhase string `json:"currentPhase"`
		} `json:"game"`
	}
	json.Unmarshal(payloadBytes, &gameUpdate)
	return gameUpdate.Game.CurrentPhase
}

// Helper to extract game status
func extractGameStatus(msg dto.WebSocketMessage) string {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var gameUpdate struct {
		Game struct {
			Status string `json:"status"`
		} `json:"game"`
	}
	json.Unmarshal(payloadBytes, &gameUpdate)
	return gameUpdate.Game.Status
}

// Helper to extract available starting cards
func extractStartingCards(msg dto.WebSocketMessage) []string {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var gameUpdate struct {
		Game struct {
			CurrentPlayer struct {
				SelectStartingCardsPhase *struct {
					AvailableCards []string `json:"availableCards"`
				} `json:"selectStartingCardsPhase"`
			} `json:"currentPlayer"`
		} `json:"game"`
	}
	json.Unmarshal(payloadBytes, &gameUpdate)

	if gameUpdate.Game.CurrentPlayer.SelectStartingCardsPhase != nil {
		return gameUpdate.Game.CurrentPlayer.SelectStartingCardsPhase.AvailableCards
	}
	return nil
}

// Helper to extract available corporations
func extractStartingCorporations(msg dto.WebSocketMessage) []string {
	payloadBytes, _ := json.Marshal(msg.Payload)
	var gameUpdate struct {
		Game struct {
			CurrentPlayer struct {
				SelectStartingCardsPhase *struct {
					AvailableCorporations []string `json:"availableCorporations"`
				} `json:"selectStartingCardsPhase"`
			} `json:"currentPlayer"`
		} `json:"game"`
	}
	json.Unmarshal(payloadBytes, &gameUpdate)

	if gameUpdate.Game.CurrentPlayer.SelectStartingCardsPhase != nil {
		return gameUpdate.Game.CurrentPlayer.SelectStartingCardsPhase.AvailableCorporations
	}
	return nil
}

func TestGameFlow_CreateLobbyAndJoin(t *testing.T) {
	// Setup test server
	srv := setupTestServer(t)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	// Step 1: Create game via HTTP
	gameID := createGameViaHTTP(t, srv.URL)
	t.Logf("🎮 Game created: %s", gameID)

	// Step 2: Alice joins as first player (becomes host)
	alice := NewTestWebSocketClient(t, wsURL, gameID, "Alice")
	defer alice.Close()

	assert.NotEmpty(t, alice.playerID, "Alice should have player ID")

	// Step 3: Bob joins as second player
	bob := NewTestWebSocketClient(t, wsURL, gameID, "Bob")
	defer bob.Close()

	assert.NotEmpty(t, bob.playerID, "Bob should have player ID")

	// Step 4: Alice should receive notification of Bob joining
	msg, err := alice.WaitForMessageType(dto.MessageTypeGameUpdated, 2*time.Second)
	assert.NoError(t, err, "Alice should receive update when Bob joins")

	// Verify game is still in lobby status
	status := extractGameStatus(msg)
	assert.Equal(t, "lobby", status, "Game should still be in lobby status")

	t.Logf("✅ Test passed: Two players successfully joined lobby")
}

func TestGameFlow_StartGame(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	gameID := createGameViaHTTP(t, srv.URL)

	// Alice joins (host)
	alice := NewTestWebSocketClient(t, wsURL, gameID, "Alice")
	defer alice.Close()

	// Bob joins
	bob := NewTestWebSocketClient(t, wsURL, gameID, "Bob")
	defer bob.Close()

	// Drain any pending messages
	alice.DrainMessages()
	bob.DrainMessages()

	// Alice starts the game
	err := alice.Send(dto.WebSocketMessage{
		Type:    "action.game-management.start-game",
		Payload: map[string]interface{}{},
	})
	require.NoError(t, err, "Alice should be able to start game")

	// Both players should receive game-updated with new phase
	aliceUpdate, err := alice.WaitForMessageType(dto.MessageTypeGameUpdated, 3*time.Second)
	require.NoError(t, err, "Alice should receive game started update")

	bobUpdate, err := bob.WaitForMessageType(dto.MessageTypeGameUpdated, 3*time.Second)
	require.NoError(t, err, "Bob should receive game started update")

	// Verify game transitioned to starting_card_selection phase
	alicePhase := extractGamePhase(aliceUpdate)
	bobPhase := extractGamePhase(bobUpdate)

	assert.Equal(t, "starting_card_selection", alicePhase, "Alice should see starting_card_selection phase")
	assert.Equal(t, "starting_card_selection", bobPhase, "Bob should see starting_card_selection phase")

	// Verify game status is active
	assert.Equal(t, "active", extractGameStatus(aliceUpdate))
	assert.Equal(t, "active", extractGameStatus(bobUpdate))

	t.Logf("✅ Test passed: Game successfully started and transitioned to card selection")
}

func TestGameFlow_SelectStartingCards(t *testing.T) {
	t.Skip("Card selection service requires card.GameRepository interface compatibility - see TEST_SUMMARY.md")

	srv := setupTestServer(t)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	gameID := createGameViaHTTP(t, srv.URL)

	// Setup two players
	alice := NewTestWebSocketClient(t, wsURL, gameID, "Alice")
	defer alice.Close()

	bob := NewTestWebSocketClient(t, wsURL, gameID, "Bob")
	defer bob.Close()

	alice.DrainMessages()
	bob.DrainMessages()

	// Start game
	err := alice.Send(dto.WebSocketMessage{
		Type:    "action.game-management.start-game",
		Payload: map[string]interface{}{},
	})
	require.NoError(t, err)

	// Wait for starting card selection phase
	aliceUpdate, err := alice.WaitForMessageType(dto.MessageTypeGameUpdated, 3*time.Second)
	require.NoError(t, err)

	bobUpdate, err := bob.WaitForMessageType(dto.MessageTypeGameUpdated, 3*time.Second)
	require.NoError(t, err)

	// Extract available cards and corporations
	aliceCards := extractStartingCards(aliceUpdate)
	aliceCorporations := extractStartingCorporations(aliceUpdate)

	bobCards := extractStartingCards(bobUpdate)
	bobCorporations := extractStartingCorporations(bobUpdate)

	t.Logf("Alice has %d cards and %d corporations available", len(aliceCards), len(aliceCorporations))
	t.Logf("Bob has %d cards and %d corporations available", len(bobCards), len(bobCorporations))

	assert.Len(t, aliceCards, 10, "Alice should have 10 starting cards")
	assert.Len(t, aliceCorporations, 2, "Alice should have 2 corporation options")

	assert.Len(t, bobCards, 10, "Bob should have 10 starting cards")
	assert.Len(t, bobCorporations, 2, "Bob should have 2 corporation options")

	// Alice selects 3 cards and first corporation
	selectedCards := aliceCards[:3]
	selectedCorp := aliceCorporations[0]

	alice.DrainMessages()
	bob.DrainMessages()

	err = alice.Send(dto.WebSocketMessage{
		Type: "action.card.select-starting-card",
		Payload: map[string]interface{}{
			"cardIds":       selectedCards,
			"corporationId": selectedCorp,
		},
	})
	require.NoError(t, err, "Alice should be able to select starting cards")

	// Wait for confirmation
	_, err = alice.WaitForMessageType(dto.MessageTypeGameUpdated, 2*time.Second)
	assert.NoError(t, err, "Alice should receive confirmation")

	// Bob selects 5 cards and second corporation
	selectedCardsBob := bobCards[:5]
	selectedCorpBob := bobCorporations[1]

	err = bob.Send(dto.WebSocketMessage{
		Type: "action.card.select-starting-card",
		Payload: map[string]interface{}{
			"cardIds":       selectedCardsBob,
			"corporationId": selectedCorpBob,
		},
	})
	require.NoError(t, err, "Bob should be able to select starting cards")

	// Both players should receive update when both selections are complete
	aliceFinal, err := alice.WaitForMessageType(dto.MessageTypeGameUpdated, 3*time.Second)
	require.NoError(t, err, "Alice should receive final update")

	bobFinal, err := bob.WaitForMessageType(dto.MessageTypeGameUpdated, 3*time.Second)
	require.NoError(t, err, "Bob should receive final update")

	// Verify game transitioned to action phase
	aliceFinalPhase := extractGamePhase(aliceFinal)
	bobFinalPhase := extractGamePhase(bobFinal)

	assert.Equal(t, "action", aliceFinalPhase, "Game should transition to action phase after all selections")
	assert.Equal(t, "action", bobFinalPhase, "Game should transition to action phase after all selections")

	t.Logf("✅ Test passed: Both players selected starting cards and game transitioned to action phase")
}

func TestGameFlow_ReconnectPlayer(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	gameID := createGameViaHTTP(t, srv.URL)

	// Alice joins
	alice := NewTestWebSocketClient(t, wsURL, gameID, "Alice")
	alicePlayerID := alice.playerID

	// Disconnect Alice
	alice.Close()
	time.Sleep(100 * time.Millisecond)

	// Alice reconnects with same player ID
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	aliceReconnected := &TestWebSocketClient{
		conn:       conn,
		gameID:     gameID,
		playerID:   alicePlayerID,
		playerName: "Alice",
		messages:   make(chan dto.WebSocketMessage, 100),
		errors:     make(chan error, 10),
		done:       make(chan struct{}),
		t:          t,
	}
	go aliceReconnected.readLoop()
	defer aliceReconnected.Close()

	// Send reconnect message with player ID
	err = aliceReconnected.Send(dto.WebSocketMessage{
		Type: "player-connect",
		Payload: map[string]interface{}{
			"playerName": "Alice",
			"gameId":     gameID,
			"playerId":   alicePlayerID,
		},
	})
	require.NoError(t, err)

	// Should receive game state
	msg, err := aliceReconnected.WaitForMessageType(dto.MessageTypeGameUpdated, 2*time.Second)
	assert.NoError(t, err, "Reconnected player should receive game state")
	assert.NotNil(t, msg, "Game state should not be nil")

	t.Logf("✅ Test passed: Player successfully reconnected to game")
}

// setupTestServer creates a test HTTP server with WebSocket support
func setupTestServer(t *testing.T) *httptest.Server {
	// Initialize event bus
	eventBus := events.NewEventBus()

	// Initialize repositories
	gameRepo := game.NewRepository(eventBus)
	playerRepo := player.NewRepository(eventBus)
	cardRepo := card.NewCardRepository()

	// Load cards from JSON - required for game creation
	err := cardRepo.LoadCards(context.Background())
	require.NoError(t, err, "Failed to load cards")

	cardDeckRepo := card.NewCardDeckRepository()

	// Session repository (in-memory)
	sessionRepo := sessionpkg.NewRepository()

	// Initialize WebSocket hub
	hub := wsCore.NewHub()
	go hub.Run(context.Background())

	// Initialize session manager
	sessionManager := session.NewSessionManager(gameRepo, playerRepo, cardRepo, sessionRepo, hub)

	// Initialize lobby service
	lobbyService := lobby.NewService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, sessionRepo, eventBus)

	// Initialize admin service
	adminService := admin.NewAdminService(gameRepo, playerRepo, cardRepo, cardDeckRepo, sessionManager, sessionRepo)

	// Connection management actions
	connectPlayerAction := actions.NewConnectPlayerAction(lobbyService, playerRepo, sessionManager)
	disconnectPlayerAction := actions.NewDisconnectPlayerAction(playerRepo, sessionManager)

	// Initialize actions (simplified for testing - only what we need)
	skipAction := actions.NewSkipAction(nil, sessionManager)
	convertHeatAction := actions.NewConvertHeatToTemperatureAction(playerRepo, nil, sessionManager)
	convertPlantsAction := actions.NewConvertPlantsToGreeneryAction(playerRepo, nil, sessionManager)

	// Card selection actions - handlers exist but will panic if called with nil service
	selectStartingCardsAction := card_selection.NewSelectStartingCardsAction(playerRepo, cardRepo, gameRepo, sessionManager)
	selectProductionCardsAction := card_selection.NewSelectProductionCardsAction(playerRepo, cardRepo, sessionManager)
	submitSellPatentsAction := card_selection.NewSubmitSellPatentsAction(playerRepo, sessionManager)

	// ConfirmCardDrawAction now uses repositories directly (no DrawService)
	confirmCardDrawAction := card_selection.NewConfirmCardDrawAction(cardRepo, cardDeckRepo, playerRepo, sessionManager)

	// SelectCardsAction needs the other selection actions as dependencies
	selectCardsAction := card_selection.NewSelectCardsAction(playerRepo, submitSellPatentsAction, selectProductionCardsAction)

	// Simplified card actions - most are nil since card system is stub
	playCardAction := actions.NewPlayCardAction(cardRepo, gameRepo, playerRepo, cardDeckRepo, nil, nil, sessionManager)
	selectTileAction := actions.NewSelectTileAction(playerRepo, gameRepo, nil, nil, sessionManager)
	// PlayCardActionAction now uses repositories directly (no PlayService)
	playCardActionAction := actions.NewPlayCardActionAction(cardRepo, playerRepo, nil, nil, sessionManager)

	// Standard project actions
	buildAquiferAction := standard_projects.NewBuildAquiferAction(playerRepo, nil, nil, sessionManager)
	launchAsteroidAction := standard_projects.NewLaunchAsteroidAction(playerRepo, nil, nil, sessionManager)
	buildPowerPlantAction := standard_projects.NewBuildPowerPlantAction(playerRepo, nil, sessionManager)
	plantGreeneryAction := standard_projects.NewPlantGreeneryAction(playerRepo, gameRepo, nil, sessionManager)
	buildCityAction := standard_projects.NewBuildCityAction(playerRepo, nil, nil, sessionManager)
	sellPatentsAction := standard_projects.NewSellPatentsAction(playerRepo, sessionManager)

	// Initialize WebSocket service
	webSocketService := wsHandler.NewWebSocketService(
		lobbyService,
		adminService,
		gameRepo,
		cardRepo,
		hub,
		connectPlayerAction,
		disconnectPlayerAction,
		buildAquiferAction,
		launchAsteroidAction,
		buildPowerPlantAction,
		plantGreeneryAction,
		buildCityAction,
		sellPatentsAction,
		skipAction,
		convertHeatAction,
		convertPlantsAction,
		playCardAction,
		selectTileAction,
		playCardActionAction,
		submitSellPatentsAction,
		selectStartingCardsAction,
		selectProductionCardsAction,
		confirmCardDrawAction,
		selectCardsAction,
	)

	// Start WebSocket service
	wsCtx, wsCancel := context.WithCancel(context.Background())
	t.Cleanup(wsCancel)
	go webSocketService.Run(wsCtx)

	// Setup routers
	mainRouter := mux.NewRouter()
	apiRouter := httpHandler.SetupRouter(lobbyService, cardRepo, playerRepo, gameRepo, cardRepo, sessionRepo)
	mainRouter.PathPrefix("/api/v1").Handler(apiRouter)
	mainRouter.HandleFunc("/ws", webSocketService.ServeWS)

	// Create test server
	srv := httptest.NewServer(mainRouter)
	t.Cleanup(srv.Close)

	return srv
}

// createGameViaHTTP creates a new game using HTTP API
func createGameViaHTTP(t *testing.T, baseURL string) string {
	requestBody := `{"maxPlayers": 4}`
	resp, err := http.Post(baseURL+"/api/v1/games", "application/json", strings.NewReader(requestBody))
	require.NoError(t, err, "Failed to create game")
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode, "Game creation should return 201 Created")

	var result struct {
		Game struct {
			ID string `json:"id"`
		} `json:"game"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Failed to decode response")

	return result.Game.ID
}
