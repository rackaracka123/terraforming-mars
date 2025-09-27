package model

import "time"

// Game represents a unified game entity containing both metadata and state
type Game struct {
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
	RemainingActions int              `json:"remainingActions" ts:"number"` // Remaining actions in the current turn
	Board            Board            `json:"board" ts:"Board"`             // Game board with tiles and occupancy state
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
		Board:      Board{Tiles: []Tile{}}, // Initialize with empty board, service will populate
	}
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
