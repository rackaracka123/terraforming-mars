package fixtures

import (
	"time"

	"github.com/google/uuid"
	"terraforming-mars-backend/internal/game"
)

// GameOption is a function that modifies a test game
type GameOption func(*game.Game)

// NewTestGame creates a new game for testing with optional modifications
func NewTestGame(options ...GameOption) *game.Game {
	g := &game.Game{
		ID:           uuid.New().String(),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Status:       game.GameStatusActive,
		CurrentPhase: game.GamePhaseAction,
		PlayerIDs:    []string{},
		Settings: game.GameSettings{
			DevelopmentMode: false,
			CardPacks:       []string{"base"},
		},
		Generation:   1,
		HostPlayerID: "",
	}

	for _, option := range options {
		option(g)
	}

	return g
}

// WithGameID sets a specific game ID
func WithGameID(id string) GameOption {
	return func(g *game.Game) {
		g.ID = id
	}
}

// WithStatus sets the game status
func WithStatus(status game.GameStatus) GameOption {
	return func(g *game.Game) {
		g.Status = status
	}
}

// WithCurrentPhase sets the game phase
func WithCurrentPhase(phase game.GamePhase) GameOption {
	return func(g *game.Game) {
		g.CurrentPhase = phase
	}
}

// WithPlayerIDs sets the player IDs
func WithPlayerIDs(playerIDs ...string) GameOption {
	return func(g *game.Game) {
		g.PlayerIDs = playerIDs
		if len(playerIDs) > 0 && g.HostPlayerID == "" {
			g.HostPlayerID = playerIDs[0]
		}
	}
}

// WithHostPlayerID sets the host player ID
func WithHostPlayerID(hostID string) GameOption {
	return func(g *game.Game) {
		g.HostPlayerID = hostID
	}
}

// WithGeneration sets the generation number
func WithGeneration(generation int) GameOption {
	return func(g *game.Game) {
		g.Generation = generation
	}
}

// WithDevelopmentMode enables development mode
func WithDevelopmentMode() GameOption {
	return func(g *game.Game) {
		g.Settings.DevelopmentMode = true
	}
}

// WithCardPacks sets the card packs
func WithCardPacks(packs ...string) GameOption {
	return func(g *game.Game) {
		g.Settings.CardPacks = packs
	}
}
