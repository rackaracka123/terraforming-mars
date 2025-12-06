package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/deck"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// TestContext provides a reusable test context
func TestContext() context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, "test", true)
}

// TestLogger creates a test logger (no-op or minimal output)
func TestLogger() *zap.Logger {
	return logger.Get()
}

// MockBroadcaster is a placeholder for backward compatibility with test code
// Deprecated: No longer used in tests as broadcasting is now handled explicitly
type MockBroadcaster struct {
	// Deprecated: No longer tracked with automatic broadcasting removed
	BroadcastCalls []BroadcastCall
}

type BroadcastCall struct {
	GameID    string
	PlayerIDs []string
	Timestamp time.Time
}

func NewMockBroadcaster() *MockBroadcaster {
	return &MockBroadcaster{
		BroadcastCalls: make([]BroadcastCall, 0),
	}
}

func (m *MockBroadcaster) CallCount() int {
	return len(m.BroadcastCalls)
}

func (m *MockBroadcaster) Reset() {
	m.BroadcastCalls = make([]BroadcastCall, 0)
}

// CreateTestCardRegistry creates a card registry with test cards
func CreateTestCardRegistry() cards.CardRegistry {
	testCards := []gamecards.Card{
		// Corporations (need at least 8 for 4 players getting 2 each)
		{
			ID:   "corp-tharsis-republic",
			Name: "Tharsis Republic",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-mining-guild",
			Name: "Mining Guild",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-inventrix",
			Name: "Inventrix",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-credicor",
			Name: "Credicor",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-ecoline",
			Name: "Ecoline",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-helion",
			Name: "Helion",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-interplanetary-cinematics",
			Name: "Interplanetary Cinematics",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-phobolog",
			Name: "Phobolog",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-saturn-systems",
			Name: "Saturn Systems",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		{
			ID:   "corp-teractor",
			Name: "Teractor",
			Type: gamecards.CardTypeCorporation,
			Pack: "base",
		},
		// Project cards (need at least 40 for typical games with 4 players getting 10 cards each)
		{
			ID:   "card-power-plant",
			Name: "Power Plant",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 4,
		},
		{
			ID:   "card-asteroid",
			Name: "Asteroid",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 14,
		},
		{
			ID:   "card-water-import",
			Name: "Water Import",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 12,
		},
		{
			ID:   "card-ai-central",
			Name: "AI Central",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 21,
		},
		{
			ID:   "card-aquifer-pumping",
			Name: "Aquifer Pumping",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 18,
		},
		{
			ID:   "card-asteroid-mining",
			Name: "Asteroid Mining",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 30,
		},
		{
			ID:   "card-biomass-combustors",
			Name: "Biomass Combustors",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 4,
		},
		{
			ID:   "card-building-industries",
			Name: "Building Industries",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 6,
		},
		{
			ID:   "card-capital",
			Name: "Capital",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 26,
		},
		{
			ID:   "card-carbonate-processing",
			Name: "Carbonate Processing",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 6,
		},
		{
			ID:   "card-cartel",
			Name: "Cartel",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 8,
		},
		{
			ID:   "card-colonizer-training-camp",
			Name: "Colonizer Training Camp",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 8,
		},
		{
			ID:   "card-comet",
			Name: "Comet",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 21,
		},
		{
			ID:   "card-research",
			Name: "Research",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 11,
		},
		{
			ID:   "card-development-center",
			Name: "Development Center",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 11,
		},
		{
			ID:   "card-dust-seals",
			Name: "Dust Seals",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 2,
		},
		{
			ID:   "card-earth-catapult",
			Name: "Earth Catapult",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 23,
		},
		{
			ID:   "card-earth-office",
			Name: "Earth Office",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 1,
		},
		{
			ID:   "card-energy-tapping",
			Name: "Energy Tapping",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 3,
		},
		{
			ID:   "card-eos-chasma-national-park",
			Name: "Eos Chasma National Park",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 16,
		},
		{
			ID:   "card-extreme-cold-fungus",
			Name: "Extreme-Cold Fungus",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 13,
		},
		{
			ID:   "card-fish",
			Name: "Fish",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 9,
		},
		{
			ID:   "card-food-factory",
			Name: "Food Factory",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 12,
		},
		{
			ID:   "card-fuel-factory",
			Name: "Fuel Factory",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 7,
		},
		{
			ID:   "card-fusion-power",
			Name: "Fusion Power",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 14,
		},
		{
			ID:   "card-ganymede-colony",
			Name: "Ganymede Colony",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 20,
		},
		{
			ID:   "card-gene-repair",
			Name: "Gene Repair",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 12,
		},
		{
			ID:   "card-giant-ice-asteroid",
			Name: "Giant Ice Asteroid",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 36,
		},
		{
			ID:   "card-giant-space-mirror",
			Name: "Giant Space Mirror",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 17,
		},
		{
			ID:   "card-greenhouse",
			Name: "Greenhouse",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 6,
		},
		{
			ID:   "card-heather",
			Name: "Heather",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 6,
		},
		{
			ID:   "card-heat-trappers",
			Name: "Heat Trappers",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 6,
		},
		{
			ID:   "card-herbivores",
			Name: "Herbivores",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 12,
		},
		{
			ID:   "card-hiring-grant",
			Name: "Hiring Grant",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 8,
		},
		{
			ID:   "card-ice-asteroid",
			Name: "Ice Asteroid",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 23,
		},
		{
			ID:   "card-immigration-shuttles",
			Name: "Immigration Shuttles",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 31,
		},
		{
			ID:   "card-imported-ghg",
			Name: "Imported GHG",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 7,
		},
		{
			ID:   "card-imported-hydrogen",
			Name: "Imported Hydrogen",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 16,
		},
		{
			ID:   "card-imported-nitrogen",
			Name: "Imported Nitrogen",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 23,
		},
		{
			ID:   "card-industrial-center",
			Name: "Industrial Center",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 4,
		},
		{
			ID:   "card-insects",
			Name: "Insects",
			Type: gamecards.CardTypeActive,
			Pack: "base",
			Cost: 9,
		},
		{
			ID:   "card-investment-loan",
			Name: "Investment Loan",
			Type: gamecards.CardTypeEvent,
			Pack: "base",
			Cost: 3,
		},
		{
			ID:   "card-kelp-farming",
			Name: "Kelp Farming",
			Type: gamecards.CardTypeAutomated,
			Pack: "base",
			Cost: 17,
		},
		// Preludes
		{
			ID:   "prelude-allied-banks",
			Name: "Allied Banks",
			Type: gamecards.CardTypePrelude,
			Pack: "prelude",
		},
	}

	return cards.NewInMemoryCardRegistry(testCards)
}

// CreateTestGameWithPlayers creates a game with specified number of players
func CreateTestGameWithPlayers(t *testing.T, numPlayers int, broadcaster *MockBroadcaster) (*game.Game, game.GameRepository) {
	t.Helper()

	repo := game.NewInMemoryGameRepository()
	cardRegistry := CreateTestCardRegistry()

	// Create game
	settings := game.GameSettings{
		MaxPlayers: 4,
		CardPacks:  []string{"base"},
	}

	testGame := game.NewGame("test-game-id", "", settings)
	allCards := cardRegistry.GetAll()

	// Separate cards by type
	projectCards := make([]string, 0)
	corpCards := make([]string, 0)
	preludeCards := make([]string, 0)

	for _, card := range allCards {
		switch card.Type {
		case gamecards.CardTypeCorporation:
			corpCards = append(corpCards, card.ID)
		case gamecards.CardTypePrelude:
			preludeCards = append(preludeCards, card.ID)
		default:
			projectCards = append(projectCards, card.ID)
		}
	}

	// Create and set deck
	gameDeck := deck.NewDeck(testGame.ID(), projectCards, corpCards, preludeCards)
	testGame.SetDeck(gameDeck)

	err := repo.Create(context.Background(), testGame)
	if err != nil {
		t.Fatalf("Failed to create test game: %v", err)
	}

	// Add players
	ctx := context.Background()
	for i := 0; i < numPlayers; i++ {
		playerID := fmt.Sprintf("player-%d", i+1)
		playerName := "Player " + string(rune('A'+i))

		// Create player
		newPlayer := player.NewPlayer(testGame.EventBus(), testGame.ID(), playerID, playerName)

		// Add to game
		err := testGame.AddPlayer(ctx, newPlayer)
		if err != nil {
			t.Fatalf("Failed to add player %d: %v", i, err)
		}
	}

	return testGame, repo
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error, message string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error, message string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error, got nil", message)
	}
}

// AssertEqual fails the test if expected != actual
func AssertEqual[T comparable](t *testing.T, expected, actual T, message string) {
	t.Helper()
	if expected != actual {
		t.Fatalf("%s: expected %v, got %v", message, expected, actual)
	}
}

// AssertNotEqual fails the test if expected == actual
func AssertNotEqual[T comparable](t *testing.T, expected, actual T, message string) {
	t.Helper()
	if expected == actual {
		t.Fatalf("%s: expected not equal to %v", message, expected)
	}
}

// AssertTrue fails the test if condition is false
func AssertTrue(t *testing.T, condition bool, message string) {
	t.Helper()
	if !condition {
		t.Fatalf("%s: expected true, got false", message)
	}
}

// AssertFalse fails the test if condition is true
func AssertFalse(t *testing.T, condition bool, message string) {
	t.Helper()
	if condition {
		t.Fatalf("%s: expected false, got true", message)
	}
}
