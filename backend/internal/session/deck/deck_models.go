package deck

import "terraforming-mars-backend/internal/session/types"

// GameDeck represents the deck state for a specific game
type GameDeck struct {
	GameID         string   `json:"gameId"`
	ProjectCards   []string `json:"projectCards"`   // Available project card IDs (draw pile)
	Corporations   []string `json:"corporations"`   // Available corporation card IDs
	DiscardPile    []string `json:"discardPile"`    // Discarded card IDs
	RemovedCards   []string `json:"removedCards"`   // Cards removed from game permanently
	PreludeCards   []string `json:"preludeCards"`   // Available prelude card IDs
	DrawnCardCount int      `json:"drawnCardCount"` // Total cards drawn (for statistics)
	ShuffleCount   int      `json:"shuffleCount"`   // Number of times deck was shuffled
}

// CardDefinitions stores all loaded card definitions
type CardDefinitions struct {
	AllCards         map[string]types.Card // All cards indexed by ID
	ProjectCards     []types.Card          // Project cards only
	CorporationCards []types.Card          // Corporation cards only
	PreludeCards     []types.Card          // Prelude cards only
	StartingCards    []types.Card          // Starting cards (subset of project cards)
}

// NewGameDeck creates a new game deck with all cards available
func NewGameDeck(gameID string, projectCardIDs, corpIDs, preludeIDs []string) *GameDeck {
	return &GameDeck{
		GameID:         gameID,
		ProjectCards:   projectCardIDs,
		Corporations:   corpIDs,
		PreludeCards:   preludeIDs,
		DiscardPile:    make([]string, 0),
		RemovedCards:   make([]string, 0),
		DrawnCardCount: 0,
		ShuffleCount:   0,
	}
}
