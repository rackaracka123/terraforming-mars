package deck

import (
	"sync"
	"terraforming-mars-backend/internal/session/game/card"
)

// GameDeck represents the deck state for a specific game with encapsulated state
type GameDeck struct {
	// Private fields
	mu             sync.RWMutex
	gameID         string
	projectCards   []string // Available project card IDs (draw pile)
	corporations   []string // Available corporation card IDs
	discardPile    []string // Discarded card IDs
	removedCards   []string // Cards removed from game permanently
	preludeCards   []string // Available prelude card IDs
	drawnCardCount int      // Total cards drawn (for statistics)
	shuffleCount   int      // Number of times deck was shuffled
}

// CardDefinitions stores all loaded card definitions
type CardDefinitions struct {
	AllCards         map[string]card.Card // All cards indexed by ID
	ProjectCards     []card.Card          // Project cards only
	CorporationCards []card.Card          // Corporation cards only
	PreludeCards     []card.Card          // Prelude cards only
	StartingCards    []card.Card          // Starting cards (subset of project cards)
}

// NewGameDeck creates a new game deck with all cards available
func NewGameDeck(gameID string, projectCardIDs, corpIDs, preludeIDs []string) *GameDeck {
	return &GameDeck{
		gameID:         gameID,
		projectCards:   projectCardIDs,
		corporations:   corpIDs,
		preludeCards:   preludeIDs,
		discardPile:    make([]string, 0),
		removedCards:   make([]string, 0),
		drawnCardCount: 0,
		shuffleCount:   0,
	}
}

// ================== Getters ==================

func (d *GameDeck) GameID() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.gameID
}

func (d *GameDeck) ProjectCards() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	cardsCopy := make([]string, len(d.projectCards))
	copy(cardsCopy, d.projectCards)
	return cardsCopy
}

func (d *GameDeck) Corporations() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	corpsCopy := make([]string, len(d.corporations))
	copy(corpsCopy, d.corporations)
	return corpsCopy
}

func (d *GameDeck) DiscardPile() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	discardCopy := make([]string, len(d.discardPile))
	copy(discardCopy, d.discardPile)
	return discardCopy
}

func (d *GameDeck) RemovedCards() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	removedCopy := make([]string, len(d.removedCards))
	copy(removedCopy, d.removedCards)
	return removedCopy
}

func (d *GameDeck) PreludeCards() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	preludeCopy := make([]string, len(d.preludeCards))
	copy(preludeCopy, d.preludeCards)
	return preludeCopy
}

func (d *GameDeck) DrawnCardCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.drawnCardCount
}

func (d *GameDeck) ShuffleCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.shuffleCount
}

// ================== Operations ==================

// GetAvailableCardCount returns the number of available project cards
func (d *GameDeck) GetAvailableCardCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.projectCards)
}

// Draw draws N project cards from the deck
func (d *GameDeck) Draw(count int) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	available := len(d.projectCards)
	if count > available {
		count = available
	}

	// Draw cards from top of deck
	drawnCards := make([]string, count)
	copy(drawnCards, d.projectCards[:count])
	d.projectCards = d.projectCards[count:]
	d.drawnCardCount += count

	return drawnCards, nil
}

// DrawCorporations draws N corporation cards
func (d *GameDeck) DrawCorporations(count int) ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	available := len(d.corporations)
	if count > available {
		count = available
	}

	// Draw corporations from top
	drawnCorps := make([]string, count)
	copy(drawnCorps, d.corporations[:count])
	d.corporations = d.corporations[count:]

	return drawnCorps, nil
}

// Discard adds cards to the discard pile
func (d *GameDeck) Discard(cardIDs []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.discardPile = append(d.discardPile, cardIDs...)
	return nil
}

// Shuffle reshuffles the discard pile back into the project cards
func (d *GameDeck) Shuffle() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Add discard pile back to project cards
	d.projectCards = append(d.projectCards, d.discardPile...)
	d.discardPile = make([]string, 0)
	d.shuffleCount++

	return nil
}
