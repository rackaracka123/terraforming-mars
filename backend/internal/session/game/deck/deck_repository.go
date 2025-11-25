package deck

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// Repository manages card decks and definitions
// Deck operations are scoped to a specific game, while card definitions are global
type Repository interface {
	// Game-scoped deck operations (no gameID parameter needed)
	CreateDeck(ctx context.Context, settings types.GameSettings) error
	DrawProjectCards(ctx context.Context, count int) ([]string, error)
	DrawCorporations(ctx context.Context, count int) ([]string, error)
	DiscardCards(ctx context.Context, cardIDs []string) error
	GetAvailableCardCount(ctx context.Context) (int, error)

	// Global card definition queries (not game-scoped)
	GetCardByID(ctx context.Context, cardID string) (*types.Card, error)
	GetAllCards(ctx context.Context) ([]types.Card, error)
	GetProjectCards(ctx context.Context) ([]types.Card, error)
	GetCorporationCards(ctx context.Context) ([]types.Card, error)
	GetStartingCardPool(ctx context.Context) ([]types.Card, error)
	ListCardsByIdMap(ctx context.Context, ids map[string]struct{}) (map[string]types.Card, error)
}

// RepositoryImpl implements the Repository interface
type RepositoryImpl struct {
	gameID      string // Bound to specific game (empty for global instance)
	mu          sync.RWMutex
	decks       map[string]*GameDeck // gameID -> GameDeck (shared storage)
	definitions *CardDefinitions     // All card definitions
}

// NewRepository creates a global deck repository with loaded card definitions
// For game-scoped operations, use NewGameScopedRepository
func NewRepository(ctx context.Context) (Repository, error) {
	// Load card definitions from JSON
	defs, err := LoadCardsFromJSON(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load card definitions: %w", err)
	}

	return &RepositoryImpl{
		gameID:      "", // Empty means global instance
		decks:       make(map[string]*GameDeck),
		definitions: defs,
	}, nil
}

// NewGameScopedRepository creates a game-scoped deck repository
func NewGameScopedRepository(gameID string, decks map[string]*GameDeck, definitions *CardDefinitions) Repository {
	return &RepositoryImpl{
		gameID:      gameID,
		decks:       decks,
		definitions: definitions,
	}
}

// GetDefinitions returns the card definitions (for creating game-scoped instances)
func (r *RepositoryImpl) GetDefinitions() *CardDefinitions {
	return r.definitions
}

// CreateDeck initializes a new game deck
func (r *RepositoryImpl) CreateDeck(ctx context.Context, settings types.GameSettings) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	log := logger.WithGameContext(r.gameID, "")

	// Get all available card IDs
	projectCardIDs := extractCardIDs(r.definitions.ProjectCards)
	corpIDs := extractCardIDs(r.definitions.CorporationCards)
	preludeIDs := extractCardIDs(r.definitions.PreludeCards)

	// Filter cards based on game settings (card packs)
	// TODO: Implement pack filtering when settings include card pack selection
	// Game settings should specify which expansions/card packs are enabled (Base, Corporate Era, Venus Next, etc.)
	// Cards should be filtered to only include those from enabled packs before shuffling

	// Shuffle the decks
	projectCardIDs = shuffleStrings(projectCardIDs)
	corpIDs = shuffleStrings(corpIDs)
	preludeIDs = shuffleStrings(preludeIDs)

	// Create game deck
	deck := NewGameDeck(r.gameID, projectCardIDs, corpIDs, preludeIDs)
	r.decks[r.gameID] = deck

	log.Info("🎴 Game deck created",
		zap.Int("project_cards", len(deck.ProjectCards)),
		zap.Int("corporations", len(deck.Corporations)),
		zap.Int("prelude_cards", len(deck.PreludeCards)))

	return nil
}

// DrawProjectCards draws N random project cards from the deck
func (r *RepositoryImpl) DrawProjectCards(ctx context.Context, count int) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	deck, err := r.getDeck(r.gameID)
	if err != nil {
		return nil, err
	}

	log := logger.WithGameContext(r.gameID, "")

	// Check if enough cards available
	available := len(deck.ProjectCards)
	if count > available {
		log.Warn("Not enough project cards in deck, drawing all remaining",
			zap.Int("requested", count),
			zap.Int("available", available))
		count = available
	}

	// Draw cards from top of deck
	drawnCards := deck.ProjectCards[:count]
	deck.ProjectCards = deck.ProjectCards[count:]
	deck.DrawnCardCount += count

	log.Debug("🎴 Drew project cards",
		zap.Int("count", count),
		zap.Int("remaining", len(deck.ProjectCards)))

	return drawnCards, nil
}

// DrawCorporations draws N random corporation cards
func (r *RepositoryImpl) DrawCorporations(ctx context.Context, count int) ([]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	deck, err := r.getDeck(r.gameID)
	if err != nil {
		return nil, err
	}

	log := logger.WithGameContext(r.gameID, "")

	// Check if enough corporations available
	available := len(deck.Corporations)
	if count > available {
		log.Warn("Not enough corporations in deck, drawing all remaining",
			zap.Int("requested", count),
			zap.Int("available", available))
		count = available
	}

	// Draw corporations from top of deck
	drawnCorps := deck.Corporations[:count]
	deck.Corporations = deck.Corporations[count:]

	log.Debug("🏢 Drew corporations",
		zap.Int("count", count),
		zap.Int("remaining", len(deck.Corporations)))

	return drawnCorps, nil
}

// DiscardCards adds cards to the discard pile
func (r *RepositoryImpl) DiscardCards(ctx context.Context, cardIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	deck, err := r.getDeck(r.gameID)
	if err != nil {
		return err
	}

	deck.DiscardPile = append(deck.DiscardPile, cardIDs...)

	logger.WithGameContext(r.gameID, "").Debug("🗑️ Cards discarded",
		zap.Int("count", len(cardIDs)),
		zap.Int("discard_pile_size", len(deck.DiscardPile)))

	return nil
}

// GetAvailableCardCount returns the number of cards remaining in draw pile
func (r *RepositoryImpl) GetAvailableCardCount(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	deck, err := r.getDeck(r.gameID)
	if err != nil {
		return 0, err
	}

	return len(deck.ProjectCards), nil
}

// GetCardByID retrieves a specific card by ID
func (r *RepositoryImpl) GetCardByID(ctx context.Context, cardID string) (*types.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	card, exists := r.definitions.AllCards[cardID]
	if !exists {
		return nil, fmt.Errorf("card not found: %s", cardID)
	}

	return &card, nil
}

// GetAllCards retrieves all card definitions
func (r *RepositoryImpl) GetAllCards(ctx context.Context) ([]types.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cards := make([]types.Card, 0, len(r.definitions.AllCards))
	for _, card := range r.definitions.AllCards {
		cards = append(cards, card)
	}

	return cards, nil
}

// GetProjectCards retrieves all project card definitions
func (r *RepositoryImpl) GetProjectCards(ctx context.Context) ([]types.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.definitions.ProjectCards, nil
}

// GetCorporationCards retrieves all corporation card definitions
func (r *RepositoryImpl) GetCorporationCards(ctx context.Context) ([]types.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.definitions.CorporationCards, nil
}

// GetStartingCardPool retrieves all starting card definitions
func (r *RepositoryImpl) GetStartingCardPool(ctx context.Context) ([]types.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.definitions.StartingCards, nil
}

// ListCardsByIdMap retrieves multiple cards by their IDs
func (r *RepositoryImpl) ListCardsByIdMap(ctx context.Context, ids map[string]struct{}) (map[string]types.Card, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]types.Card, len(ids))
	for id := range ids {
		if card, exists := r.definitions.AllCards[id]; exists {
			result[id] = card
		}
	}

	return result, nil
}

// getDeck retrieves a game deck (must be called with lock held)
func (r *RepositoryImpl) getDeck(gameID string) (*GameDeck, error) {
	deck, exists := r.decks[gameID]
	if !exists {
		return nil, fmt.Errorf("deck not found for game: %s", gameID)
	}
	return deck, nil
}

// shuffleStrings shuffles a slice of strings using Fisher-Yates algorithm
func shuffleStrings(slice []string) []string {
	result := make([]string, len(slice))
	copy(result, slice)

	for i := len(result) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		result[i], result[j] = result[j], result[i]
	}

	return result
}
