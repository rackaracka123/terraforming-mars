package game_test

import (
	"slices"
	"testing"

	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/board"
	"terraforming-mars-backend/internal/game/shared"
	"terraforming-mars-backend/test/testutil"
)

// buildSeededGame reproduces a game from a seed via the replay primitive (which uses
// the production deck-construction path) and returns the shuffled project-card order.
func buildSeededGame(t *testing.T, seed uint64) []string {
	t.Helper()
	g, _, _ := testutil.NewSeededGame(t, seed, shared.GameSettings{})
	return g.Deck().ProjectCards()
}

func TestDeckShuffle_SameSeedIsReproducible(t *testing.T) {
	const seed uint64 = 0xC0FFEE

	first := buildSeededGame(t, seed)
	second := buildSeededGame(t, seed)

	if len(first) == 0 {
		t.Fatal("expected a non-empty project deck")
	}
	if !slices.Equal(first, second) {
		t.Fatalf("same seed produced different deck orders:\n first[:5]=%v\nsecond[:5]=%v",
			first[:min(5, len(first))], second[:min(5, len(second))])
	}
}

func TestDeckShuffle_DifferentSeedsDiffer(t *testing.T) {
	a := buildSeededGame(t, 1)
	b := buildSeededGame(t, 2)

	if slices.Equal(a, b) {
		t.Fatal("different seeds produced identical deck orders (RNG not seed-dependent)")
	}
}

// TestSetupRand_SameSeedIsReproducible guards the setup RNG stream (turn order,
// milestone/award/colony selection) used by StartGameAction.
func TestSetupRand_SameSeedIsReproducible(t *testing.T) {
	repo := testutil.NewTestGameRepository(t)
	settings := shared.GameSettings{MaxPlayers: 4}
	mk := func() []int {
		g := game.NewGame(repo.DataStore(), "setup-test", "", settings, board.GenerateMarsBoard(false))
		g.SetSeed(42)
		return g.SetupRand().Perm(10)
	}
	if !slices.Equal(mk(), mk()) {
		t.Fatal("SetupRand with the same seed produced different sequences")
	}
}
