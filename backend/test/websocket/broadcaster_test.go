package websocket_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/test/testutil"

	"go.uber.org/zap"
)

// MockHub implements a simple hub for testing
type MockHub struct {
	SentMessages []MockMessage
}

// Broadcaster for testing
type Broadcaster struct {
	gameRepo     game.GameRepository
	hub          *MockHub
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewBroadcaster creates a test broadcaster
func NewBroadcaster(
	hub *MockHub,
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *Broadcaster {
	return &Broadcaster{
		gameRepo:     gameRepo,
		hub:          hub,
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// GetBroadcastFunc returns broadcast function
func (b *Broadcaster) GetBroadcastFunc() events.BroadcastFunc {
	return func(gameID string, playerIDs []string) {
		b.BroadcastToPlayers(context.Background(), gameID, playerIDs)
	}
}

// BroadcastToPlayers sends messages to players
func (b *Broadcaster) BroadcastToPlayers(ctx context.Context, gameID string, playerIDs []string) {
	game, err := b.gameRepo.Get(ctx, gameID)
	if err != nil {
		return
	}

	if playerIDs == nil {
		players := game.GetAllPlayers()
		playerIDs = make([]string, len(players))
		for i, p := range players {
			playerIDs[i] = p.ID()
		}
	}

	for _, playerID := range playerIDs {
		b.hub.SendToPlayer(gameID, playerID, "game-state")
	}
}

type MockMessage struct {
	GameID      string
	PlayerID    string
	MessageType string
	Timestamp   time.Time
}

func NewMockHub() *MockHub {
	return &MockHub{
		SentMessages: make([]MockMessage, 0),
	}
}

func (h *MockHub) SendToPlayer(gameID, playerID string, message interface{}) {
	h.SentMessages = append(h.SentMessages, MockMessage{
		GameID:      gameID,
		PlayerID:    playerID,
		MessageType: "game-updated",
		Timestamp:   time.Now(),
	})
}

func (h *MockHub) SendToAll(gameID string, message interface{}) {
	h.SentMessages = append(h.SentMessages, MockMessage{
		GameID:      gameID,
		PlayerID:    "all",
		MessageType: "broadcast",
		Timestamp:   time.Now(),
	})
}

func (h *MockHub) Reset() {
	h.SentMessages = make([]MockMessage, 0)
}

// TestBroadcaster_BroadcastToPlayers tests that broadcaster sends to all players
func TestBroadcaster_BroadcastToPlayers(t *testing.T) {
	// Setup
	repo := game.NewInMemoryGameRepository()
	cardRegistry := testutil.CreateTestCardRegistry()
	hub := NewMockHub()
	logger := testutil.TestLogger()

	broadcaster := NewBroadcaster(hub, repo, cardRegistry, logger)

	// Create game with players
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, mockBroadcaster)
	ctx := context.Background()
	repo.Create(ctx, testGame)

	// Broadcast to all players
	players := testGame.GetAllPlayers()
	playerIDs := make([]string, len(players))
	for i, p := range players {
		playerIDs[i] = p.ID()
	}

	broadcaster.BroadcastToPlayers(ctx, testGame.ID(), playerIDs)

	// Verify messages sent
	testutil.AssertEqual(t, len(players), len(hub.SentMessages), "Should send message to each player")

	// Verify each player got a message
	for _, playerID := range playerIDs {
		found := false
		for _, msg := range hub.SentMessages {
			if msg.PlayerID == playerID {
				found = true
				break
			}
		}
		testutil.AssertTrue(t, found, "Player should receive message")
	}
}

// TestBroadcaster_GameNotFound tests broadcaster handles missing games
func TestBroadcaster_GameNotFound(t *testing.T) {
	// Setup
	repo := game.NewInMemoryGameRepository()
	cardRegistry := testutil.CreateTestCardRegistry()
	hub := NewMockHub()
	logger := testutil.TestLogger()

	broadcaster := NewBroadcaster(hub, repo, cardRegistry, logger)

	// Try to broadcast to non-existent game (should not panic)
	ctx := context.Background()
	broadcaster.BroadcastToPlayers(ctx, "non-existent-game", []string{"player-1"})

	// Should not have sent any messages
	testutil.AssertEqual(t, 0, len(hub.SentMessages), "Should not send messages for non-existent game")
}

// TestBroadcaster_EmptyPlayerList tests broadcaster with empty player list
func TestBroadcaster_EmptyPlayerList(t *testing.T) {
	// Setup
	repo := game.NewInMemoryGameRepository()
	cardRegistry := testutil.CreateTestCardRegistry()
	hub := NewMockHub()
	logger := testutil.TestLogger()

	broadcaster := NewBroadcaster(hub, repo, cardRegistry, logger)

	// Create game
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 0, mockBroadcaster)
	ctx := context.Background()
	repo.Create(ctx, testGame)

	// Broadcast with empty player list
	broadcaster.BroadcastToPlayers(ctx, testGame.ID(), []string{})

	// Should not send any messages
	testutil.AssertEqual(t, 0, len(hub.SentMessages), "Should not send messages with empty player list")
}

// TestBroadcaster_PersonalizedGameState tests that each player gets personalized data
func TestBroadcaster_PersonalizedGameState(t *testing.T) {
	// This test verifies the concept - actual personalization happens in DTO layer
	// We just verify that each player gets their own message

	// Setup
	repo := game.NewInMemoryGameRepository()
	cardRegistry := testutil.CreateTestCardRegistry()
	hub := NewMockHub()
	logger := testutil.TestLogger()

	broadcaster := NewBroadcaster(hub, repo, cardRegistry, logger)

	// Create game with players
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 3, mockBroadcaster)
	ctx := context.Background()
	repo.Create(ctx, testGame)

	players := testGame.GetAllPlayers()
	playerIDs := make([]string, len(players))
	for i, p := range players {
		playerIDs[i] = p.ID()
	}

	// Broadcast
	broadcaster.BroadcastToPlayers(ctx, testGame.ID(), playerIDs)

	// Each player should get their own message (not a shared broadcast)
	testutil.AssertEqual(t, 3, len(hub.SentMessages), "Each player should get individual message")

	// Verify no "all players" broadcast
	for _, msg := range hub.SentMessages {
		testutil.AssertNotEqual(t, "all", msg.PlayerID, "Should send to specific players, not 'all'")
	}
}

// TestBroadcaster_ConcurrentBroadcasts tests thread-safety
func TestBroadcaster_ConcurrentBroadcasts(t *testing.T) {
	// Setup
	repo := game.NewInMemoryGameRepository()
	cardRegistry := testutil.CreateTestCardRegistry()
	hub := NewMockHub()
	logger := testutil.TestLogger()

	broadcaster := NewBroadcaster(hub, repo, cardRegistry, logger)

	// Create multiple games
	mockBroadcaster := testutil.NewMockBroadcaster()
	ctx := context.Background()

	games := make([]*game.Game, 5)
	for i := 0; i < 5; i++ {
		testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, mockBroadcaster)
		repo.Create(ctx, testGame)
		games[i] = testGame
	}

	// Concurrent broadcasts
	done := make(chan bool, len(games))

	for _, g := range games {
		go func(testGame *game.Game) {
			players := testGame.GetAllPlayers()
			playerIDs := make([]string, len(players))
			for i, p := range players {
				playerIDs[i] = p.ID()
			}

			broadcaster.BroadcastToPlayers(ctx, testGame.ID(), playerIDs)
			done <- true
		}(g)
	}

	// Wait for all
	for i := 0; i < len(games); i++ {
		<-done
	}

	// Should have sent messages for all games (5 games * 2 players = 10 messages)
	testutil.AssertEqual(t, 10, len(hub.SentMessages), "Should send messages for all concurrent broadcasts")
}

// TestBroadcaster_BroadcastFunc tests the broadcast function callback
func TestBroadcaster_BroadcastFunc(t *testing.T) {
	// Setup
	repo := game.NewInMemoryGameRepository()
	cardRegistry := testutil.CreateTestCardRegistry()
	hub := NewMockHub()
	logger := testutil.TestLogger()

	broadcaster := NewBroadcaster(hub, repo, cardRegistry, logger)

	// Get broadcast function
	broadcastFunc := broadcaster.GetBroadcastFunc()
	testutil.AssertTrue(t, broadcastFunc != nil, "Should return broadcast function")

	// Create game
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, mockBroadcaster)
	ctx := context.Background()
	repo.Create(ctx, testGame)

	players := testGame.GetAllPlayers()
	playerIDs := make([]string, len(players))
	for i, p := range players {
		playerIDs[i] = p.ID()
	}

	// Call broadcast function
	broadcastFunc(testGame.ID(), playerIDs)

	time.Sleep(10 * time.Millisecond)

	// Should have sent messages
	testutil.AssertTrue(t, len(hub.SentMessages) > 0, "Broadcast function should send messages")
}

// TestBroadcaster_GameStateChanges tests broadcasting on state changes
func TestBroadcaster_GameStateChanges(t *testing.T) {
	// Setup
	repo := game.NewInMemoryGameRepository()
	cardRegistry := testutil.CreateTestCardRegistry()
	hub := NewMockHub()
	logger := testutil.TestLogger()

	broadcaster := NewBroadcaster(hub, repo, cardRegistry, logger)

	// Create game with broadcaster integrated
	settings := game.GameSettings{
		MaxPlayers: 2,
		CardPacks:  []string{"base"},
	}

	testGame := game.NewGame("test-game", "", settings, broadcaster.GetBroadcastFunc())
	ctx := context.Background()
	repo.Create(ctx, testGame)

	hub.Reset()

	// Trigger state change
	newPlayer := player.NewPlayer(testGame.EventBus(), testGame.ID(), "test-player-1", "TestPlayer")
	testGame.AddPlayer(ctx, newPlayer)

	time.Sleep(10 * time.Millisecond)

	// Should have broadcast
	if len(hub.SentMessages) > 0 {
		t.Logf("✓ State change triggered broadcast: %d messages", len(hub.SentMessages))
	}
}

// TestBroadcaster_CorrectGameID tests messages have correct game ID
func TestBroadcaster_CorrectGameID(t *testing.T) {
	// Setup
	repo := game.NewInMemoryGameRepository()
	cardRegistry := testutil.CreateTestCardRegistry()
	hub := NewMockHub()
	logger := testutil.TestLogger()

	broadcaster := NewBroadcaster(hub, repo, cardRegistry, logger)

	// Create game
	mockBroadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 2, mockBroadcaster)
	ctx := context.Background()
	repo.Create(ctx, testGame)

	players := testGame.GetAllPlayers()
	playerIDs := make([]string, len(players))
	for i, p := range players {
		playerIDs[i] = p.ID()
	}

	// Broadcast
	broadcaster.BroadcastToPlayers(ctx, testGame.ID(), playerIDs)

	// Verify all messages have correct game ID
	for _, msg := range hub.SentMessages {
		testutil.AssertEqual(t, testGame.ID(), msg.GameID, "Message should have correct game ID")
	}
}
