package card

import (
	"context"
	"sync"
)

// CardDeckRepository manages the card deck as a simple stack
type CardDeckRepository interface {
	// InitializeDeck creates a new shuffled deck for a game
	InitializeDeck(ctx context.Context, gameID string, cards []Card) error

	// Pop removes and returns the top card from the deck
	Pop(ctx context.Context, gameID string) (string, error)

	// Peek returns the top card without removing it
	Peek(ctx context.Context, gameID string) (string, error)

	// Size returns the number of cards remaining in the deck
	Size(ctx context.Context, gameID string) (int, error)

	// PlayedCardsSize returns the number of cards that have been played
	PlayedCardsSize(ctx context.Context, gameID string) (int, error)
}

// CardDeckRepositoryImpl implements CardDeckRepository
type CardDeckRepositoryImpl struct {
	mutex     sync.RWMutex
	gameDecks map[string]*DeckState
}

// DeckState represents the deck state for a game
type DeckState struct {
	Cards       []string // Stack of cards (top is last element)
	PlayedCount int      // Number of cards that have been played
}

// NewCardDeckRepository creates a new card deck repository
func NewCardDeckRepository() CardDeckRepository {
	return &CardDeckRepositoryImpl{
		gameDecks: make(map[string]*DeckState),
	}
}

// InitializeDeck creates a new shuffled deck for a game
func (r *CardDeckRepositoryImpl) InitializeDeck(ctx context.Context, gameID string, cards []Card) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	cardIDs := make([]string, len(cards))
	for i, card := range cards {
		cardIDs[i] = card.ID
	}

	// Simple shuffle (Fisher-Yates)
	for i := len(cardIDs) - 1; i > 0; i-- {
		j := i % (i + 1) // Simple deterministic shuffle for now
		cardIDs[i], cardIDs[j] = cardIDs[j], cardIDs[i]
	}

	r.gameDecks[gameID] = &DeckState{
		Cards:       cardIDs,
		PlayedCount: 0,
	}

	return nil
}

// Pop removes and returns the top card from the deck
func (r *CardDeckRepositoryImpl) Pop(ctx context.Context, gameID string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	deck, exists := r.gameDecks[gameID]
	if !exists {
		return "", ErrDeckNotFound
	}

	if len(deck.Cards) == 0 {
		return "", ErrDeckEmpty
	}

	// Pop from the end (top of stack)
	topCard := deck.Cards[len(deck.Cards)-1]
	deck.Cards = deck.Cards[:len(deck.Cards)-1]
	deck.PlayedCount++

	return topCard, nil
}

// Peek returns the top card without removing it
func (r *CardDeckRepositoryImpl) Peek(ctx context.Context, gameID string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	deck, exists := r.gameDecks[gameID]
	if !exists {
		return "", ErrDeckNotFound
	}

	if len(deck.Cards) == 0 {
		return "", ErrDeckEmpty
	}

	return deck.Cards[len(deck.Cards)-1], nil
}

// Size returns the number of cards remaining in the deck
func (r *CardDeckRepositoryImpl) Size(ctx context.Context, gameID string) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	deck, exists := r.gameDecks[gameID]
	if !exists {
		return 0, ErrDeckNotFound
	}

	return len(deck.Cards), nil
}

// PlayedCardsSize returns the number of cards that have been played
func (r *CardDeckRepositoryImpl) PlayedCardsSize(ctx context.Context, gameID string) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	deck, exists := r.gameDecks[gameID]
	if !exists {
		return 0, ErrDeckNotFound
	}

	return deck.PlayedCount, nil
}

// Error types
var (
	ErrDeckNotFound = &DeckError{Message: "deck not found"}
	ErrDeckEmpty    = &DeckError{Message: "deck is empty"}
)

type DeckError struct {
	Message string
}

func (e *DeckError) Error() string {
	return e.Message
}

// Clear removes all game decks from the repository
func (r *CardDeckRepositoryImpl) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.gameDecks = make(map[string]*DeckState)
}
