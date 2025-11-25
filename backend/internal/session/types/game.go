package types

import (
	"sync"
	"time"
)

// Game represents a unified game entity containing both metadata and state
type Game struct {
	// Serialized game data (sent to frontend)
	ID               string           `json:"id" ts:"string"`
	CreatedAt        time.Time        `json:"createdAt" ts:"string"`
	UpdatedAt        time.Time        `json:"updatedAt" ts:"string"`
	Status           GameStatus       `json:"status" ts:"GameStatus"`
	Settings         GameSettings     `json:"settings" ts:"GameSettings"`
	PlayerIDs        []string         `json:"playerIds" ts:"string[]"` // Player IDs in this game
	HostPlayerID     string           `json:"hostPlayerId" ts:"string"`
	CurrentPhase     GamePhase        `json:"currentPhase" ts:"GamePhase"`
	GlobalParameters GlobalParameters `json:"globalParameters" ts:"GlobalParameters"`
	ViewingPlayerID  string           `json:"viewingPlayerId" ts:"string"`  // The player viewing this game state
	CurrentTurn      *string          `json:"currentTurn" ts:"string|null"` // Whose turn it is (nullable)
	Generation       int              `json:"generation" ts:"number"`
	board            Board            `json:"board" ts:"Board"` // Game board with tiles and occupancy state (lowercase to avoid conflict with Board repository)

	// Non-serialized runtime state (repositories, infrastructure, players)
	// These fields are excluded from JSON serialization
	mu sync.RWMutex `json:"-"`

	// Players managed by this game (stored as interface{} to avoid circular dependency)
	// In session package, this will be map[string]*player.Player
	// In types package, player.Player is not available due to circular imports
	Players interface{} `json:"-"`

	// Game sub-repositories (grouped operations)
	// Using interface{} to avoid circular imports - will be properly typed in session package
	Core         interface{} `json:"-"` // *GameCoreRepository
	GlobalParams interface{} `json:"-"` // *GameGlobalParametersRepository
	Turn         interface{} `json:"-"` // *GameTurnRepository

	// Domain repositories (no "Repo" suffix as per naming convention)
	Board interface{} `json:"-"` // board.Repository
	Cards interface{} `json:"-"` // card.Repository
	Deck  interface{} `json:"-"` // deck.Repository

	// Infrastructure components (exported for access from session package)
	CardManager     interface{} `json:"-"` // card.CardManager
	TileProcessor   interface{} `json:"-"` // *board.Processor
	BonusCalculator interface{} `json:"-"` // *board.BonusCalculator
}

// NewGame creates a new game with the given settings
func NewGame(id string, settings GameSettings) *Game {
	now := time.Now()

	return &Game{
		ID:           id,
		CreatedAt:    now,
		UpdatedAt:    now,
		Status:       GameStatusLobby,
		Settings:     settings,
		PlayerIDs:    make([]string, 0),
		CurrentPhase: GamePhaseWaitingForGameStart,
		GlobalParameters: GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
		Generation: 1,
		board:      Board{Tiles: []Tile{}}, // Initialize with empty board, service will populate
		Players:    nil,                    // Will be initialized by session package
		mu:         sync.RWMutex{},
	}
}

// GetBoard returns the board data for serialization
// This provides access to the private board field
func (g *Game) GetBoard() Board {
	return g.board
}

// SetBoard sets the board data
// This allows updating the private board field
func (g *Game) SetBoard(b Board) {
	g.board = b
}

// Next returns the next player ID in turn order based on current turn
// Returns nil if CurrentTurn is nil or no players exist
func (g *Game) Next() *string {
	if g.CurrentTurn == nil || len(g.PlayerIDs) == 0 {
		return nil
	}

	// Find current player index
	currentIndex := -1
	for i, playerID := range g.PlayerIDs {
		if playerID == *g.CurrentTurn {
			currentIndex = i
			break
		}
	}

	if currentIndex == -1 {
		// Current turn player not found, return first player
		return &g.PlayerIDs[0]
	}

	// Calculate next player index (wrap around)
	nextIndex := (currentIndex + 1) % len(g.PlayerIDs)
	return &g.PlayerIDs[nextIndex]
}
