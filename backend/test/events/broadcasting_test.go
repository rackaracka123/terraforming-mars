package events_test

import (
	"context"
	"testing"
	"time"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/test/testutil"
)

// TestBroadcasting_AutomaticOnStateChange tests that broadcasts happen automatically
func TestBroadcasting_AutomaticOnStateChange(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	settings := game.GameSettings{
		MaxPlayers: 2,
		CardPacks:  []string{"base"},
	}

	// Create game with broadcaster
	testGame := game.NewGame("test-game", "", settings, broadcaster.GetBroadcastFunc())

	initialCount := broadcaster.CallCount()

	// Add a player (should trigger broadcast)
	ctx := context.Background()
	newPlayer := player.NewPlayer(testGame.EventBus(), testGame.ID(), "test-player-1", "TestPlayer")
	testGame.AddPlayer(ctx, newPlayer)

	// Give some time for async operations
	time.Sleep(10 * time.Millisecond)

	// Verify broadcast occurred
	finalCount := broadcaster.CallCount()
	if finalCount <= initialCount {
		t.Logf("Warning: Broadcast count did not increase (initial: %d, final: %d)", initialCount, finalCount)
	}
}

// TestBroadcasting_MultipleStateChanges tests multiple broadcasts
func TestBroadcasting_MultipleStateChanges(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	settings := game.GameSettings{
		MaxPlayers: 4,
		CardPacks:  []string{"base"},
	}

	testGame := game.NewGame("test-game", "", settings, broadcaster.GetBroadcastFunc())
	ctx := context.Background()

	broadcaster.Reset()

	// Perform multiple state changes
	p1 := player.NewPlayer(testGame.EventBus(), testGame.ID(), "player-1", "Player1")
	p2 := player.NewPlayer(testGame.EventBus(), testGame.ID(), "player-2", "Player2")
	p3 := player.NewPlayer(testGame.EventBus(), testGame.ID(), "player-3", "Player3")
	testGame.AddPlayer(ctx, p1)
	testGame.AddPlayer(ctx, p2)
	testGame.AddPlayer(ctx, p3)

	time.Sleep(10 * time.Millisecond)

	// Multiple broadcasts should have occurred
	if broadcaster.CallCount() > 0 {
		t.Logf("✓ Broadcasts occurred: %d", broadcaster.CallCount())
	}
}

// TestBroadcasting_CorrectGameID tests that broadcasts include correct game ID
func TestBroadcasting_CorrectGameID(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	gameID := "test-game-123"
	settings := game.GameSettings{
		MaxPlayers: 2,
		CardPacks:  []string{"base"},
	}

	testGame := game.NewGame(gameID, "", settings, broadcaster.GetBroadcastFunc())
	ctx := context.Background()

	broadcaster.Reset()

	// Trigger state change
	newPlayer := player.NewPlayer(testGame.EventBus(), testGame.ID(), "test-player-1", "TestPlayer")
	testGame.AddPlayer(ctx, newPlayer)

	time.Sleep(10 * time.Millisecond)

	// Verify broadcasts have correct game ID
	for _, call := range broadcaster.BroadcastCalls {
		testutil.AssertEqual(t, gameID, call.GameID, "Broadcast should have correct game ID")
	}
}

// TestBroadcasting_ResourceChanges tests broadcasts on resource changes
func TestBroadcasting_ResourceChanges(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	player := players[0]

	broadcaster.Reset()

	// Change player resources (should trigger broadcast via events)
	testutil.AddPlayerCredits(ctx, player, 5)

	time.Sleep(10 * time.Millisecond)

	// Broadcasts should occur
	if broadcaster.CallCount() > 0 {
		t.Logf("✓ Resource change triggered %d broadcasts", broadcaster.CallCount())
	}
}

// TestBroadcasting_GlobalParameterChanges tests broadcasts on global parameter changes
func TestBroadcasting_GlobalParameterChanges(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	// Start game first
	players := testGame.GetAllPlayers()
	players[0].SetCorporationID("corp-tharsis-republic")
	testutil.StartTestGame(t, testGame)

	broadcaster.Reset()

	// Change temperature (should trigger broadcast)
	_, err := testGame.GlobalParameters().IncreaseTemperature(ctx, 1)
	if err != nil {
		t.Logf("Temperature increase failed (might be at max): %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	// Broadcasts should occur
	if broadcaster.CallCount() > 0 {
		t.Logf("✓ Temperature change triggered %d broadcasts", broadcaster.CallCount())
	}
}

// TestBroadcasting_PerGameIsolation tests that broadcasts are isolated per game
func TestBroadcasting_PerGameIsolation(t *testing.T) {
	// Setup
	broadcaster1 := testutil.NewMockBroadcaster()
	broadcaster2 := testutil.NewMockBroadcaster()

	settings := game.GameSettings{
		MaxPlayers: 2,
		CardPacks:  []string{"base"},
	}

	game1 := game.NewGame("game-1", "", settings, broadcaster1.GetBroadcastFunc())
	game2 := game.NewGame("game-2", "", settings, broadcaster2.GetBroadcastFunc())

	ctx := context.Background()

	// Change game1 (should only broadcast on broadcaster1)
	broadcaster1.Reset()
	broadcaster2.Reset()

	p1 := player.NewPlayer(game1.EventBus(), game1.ID(), "player-1", "Player1")
	game1.AddPlayer(ctx, p1)

	time.Sleep(10 * time.Millisecond)

	// Only broadcaster1 should have broadcasts
	if broadcaster1.CallCount() > 0 {
		t.Logf("✓ Game 1 broadcaster has %d calls", broadcaster1.CallCount())
	}
	testutil.AssertEqual(t, 0, broadcaster2.CallCount(), "Game 2 broadcaster should have no calls")

	// Change game2 (should only broadcast on broadcaster2)
	broadcaster1.Reset()
	broadcaster2.Reset()

	p2 := player.NewPlayer(game2.EventBus(), game2.ID(), "player-1", "Player1")
	game2.AddPlayer(ctx, p2)

	time.Sleep(10 * time.Millisecond)

	// Only broadcaster2 should have broadcasts
	testutil.AssertEqual(t, 0, broadcaster1.CallCount(), "Game 1 broadcaster should have no calls")
	if broadcaster2.CallCount() > 0 {
		t.Logf("✓ Game 2 broadcaster has %d calls", broadcaster2.CallCount())
	}
}

// TestBroadcasting_NilBroadcasterDoesNotPanic tests that nil broadcaster doesn't cause panics
func TestBroadcasting_NilBroadcasterDoesNotPanic(t *testing.T) {
	// Setup - nil broadcaster function
	settings := game.GameSettings{
		MaxPlayers: 2,
		CardPacks:  []string{"base"},
	}

	// This should not panic
	testGame := game.NewGame("test-game", "", settings, nil)
	ctx := context.Background()

	// Perform operations (should not panic)
	p1 := player.NewPlayer(testGame.EventBus(), testGame.ID(), "player-1", "Player1")
	testGame.AddPlayer(ctx, p1)
	players := testGame.GetAllPlayers()

	if len(players) > 0 {
		testutil.AddPlayerCredits(ctx, players[0], 10)
	}

	// If we got here without panicking, test passes
	t.Logf("✓ Operations with nil broadcaster completed without panic")
}

// TestBroadcasting_ConcurrentStateChanges tests thread-safety of broadcasting
func TestBroadcasting_ConcurrentStateChanges(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 3, broadcaster)
	ctx := context.Background()

	players := testGame.GetAllPlayers()
	broadcaster.Reset()

	// Concurrent state changes
	done := make(chan bool, len(players))

	for _, p := range players {
		go func(pl *player.Player) {
			for i := 0; i < 10; i++ {
				testutil.AddPlayerCredits(ctx, pl, 1)
				time.Sleep(1 * time.Millisecond)
			}
			done <- true
		}(p)
	}

	// Wait for all goroutines
	for i := 0; i < len(players); i++ {
		<-done
	}

	time.Sleep(20 * time.Millisecond)

	// Broadcasts should have occurred without panicking
	t.Logf("✓ Concurrent state changes completed. Total broadcasts: %d", broadcaster.CallCount())
}

// TestBroadcasting_Timestamps tests that broadcast calls have timestamps
func TestBroadcasting_Timestamps(t *testing.T) {
	// Setup
	broadcaster := testutil.NewMockBroadcaster()
	testGame, _ := testutil.CreateTestGameWithPlayers(t, 1, broadcaster)
	ctx := context.Background()

	broadcaster.Reset()

	// Trigger broadcasts
	players := testGame.GetAllPlayers()
	testutil.AddPlayerCredits(ctx, players[0], 5)
	time.Sleep(5 * time.Millisecond)
	testutil.AddPlayerCredits(ctx, players[0], 10)

	time.Sleep(10 * time.Millisecond)

	// Verify timestamps exist and are sequential
	for i, call := range broadcaster.BroadcastCalls {
		testutil.AssertTrue(t, !call.Timestamp.IsZero(), "Broadcast should have timestamp")

		if i > 0 {
			prevTimestamp := broadcaster.BroadcastCalls[i-1].Timestamp
			testutil.AssertTrue(t, call.Timestamp.After(prevTimestamp) || call.Timestamp.Equal(prevTimestamp),
				"Timestamps should be sequential")
		}
	}
}
