package session

import (
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/board"
	"terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/game/deck"
)

// SessionFactory manages creation and retrieval of game sessions
type SessionFactory interface {
	Get(gameID string) *Session
	GetOrCreate(gameID string) *Session
	Remove(gameID string)
	WireGameRepositories(g *game.Game) // Wires repositories and infrastructure to a Game instance
}

// SessionFactoryImpl implements SessionFactory
type SessionFactoryImpl struct {
	sessions map[string]*Session
	eventBus *events.EventBusImpl

	// Shared storage for game-scoped repositories
	boards map[string]*board.Board   // Shared board storage
	decks  map[string]*deck.GameDeck // Shared deck storage

	// Global repositories (not game-scoped)
	cardRepo        card.Repository       // Card definitions (global)
	deckDefinitions *deck.CardDefinitions // Card definitions (global)

	// Infrastructure components
	effectSubscriber card.CardEffectSubscriber
	boardProcessor   *board.BoardProcessor

	mu sync.RWMutex
}

// NewSessionFactory creates a new session factory with repository dependencies
func NewSessionFactory(
	eventBus *events.EventBusImpl,
	cardRepo card.Repository,
	deckRepo deck.Repository,
) SessionFactory {
	// Create shared infrastructure components
	effectSubscriber := card.NewCardEffectSubscriber(eventBus, cardRepo)
	boardProcessor := board.NewBoardProcessor()

	// Extract card definitions from deck repository for creating game-scoped instances
	var deckDefinitions *deck.CardDefinitions
	if deckRepoImpl, ok := deckRepo.(*deck.RepositoryImpl); ok {
		deckDefinitions = deckRepoImpl.GetDefinitions()
	}

	return &SessionFactoryImpl{
		sessions:         make(map[string]*Session),
		eventBus:         eventBus,
		boards:           make(map[string]*board.Board),
		decks:            make(map[string]*deck.GameDeck),
		cardRepo:         cardRepo,
		deckDefinitions:  deckDefinitions,
		effectSubscriber: effectSubscriber,
		boardProcessor:   boardProcessor,
		mu:               sync.RWMutex{},
	}
}

// Get retrieves an existing session by game ID
// Returns nil if session doesn't exist
func (f *SessionFactoryImpl) Get(gameID string) *Session {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.sessions[gameID]
}

// GetOrCreate retrieves an existing session or creates a new one if it doesn't exist
func (f *SessionFactoryImpl) GetOrCreate(gameID string) *Session {
	// Try to get existing session first with read lock
	f.mu.RLock()
	if session, exists := f.sessions[gameID]; exists {
		f.mu.RUnlock()
		return session
	}
	f.mu.RUnlock()

	// Create new session with write lock
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have created it)
	if session, exists := f.sessions[gameID]; exists {
		return session
	}

	// Create session with nil game (will be populated by game creation action)
	// Note: In the new architecture, games are created via actions, not loaded from storage
	session := NewSession(nil, f.eventBus)

	f.sessions[gameID] = session
	return session
}

// Remove deletes a session from the factory
func (f *SessionFactoryImpl) Remove(gameID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.sessions, gameID)
}

// WireGameRepositories wires up a game.Game instance with game-scoped repositories and infrastructure
// DEPRECATED: In the new architecture, Game receives its CardManager in the constructor (NewGame)
// This method is kept for interface compatibility but does nothing
// TODO: Remove this method once all callers are updated
func (f *SessionFactoryImpl) WireGameRepositories(g *game.Game) {
	// No-op: Game now owns its own CardManager, set during construction
	// Infrastructure components are passed to NewGame() instead
}
