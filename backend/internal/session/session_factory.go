package session

import (
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/game/board"
	"terraforming-mars-backend/internal/session/game/card"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/game/deck"
	"terraforming-mars-backend/internal/session/types"
)

// SessionFactory manages creation and retrieval of game sessions
type SessionFactory interface {
	Get(gameID string) *Session
	GetOrCreate(gameID string) *Session
	Remove(gameID string)
	WireGameRepositories(g *types.Game) // Wires repositories and infrastructure to a Game instance
}

// SessionFactoryImpl implements SessionFactory
type SessionFactoryImpl struct {
	sessions map[string]*Session
	eventBus *events.EventBusImpl

	// Shared storage for game-scoped repositories
	gameStorage *game.GameStorage         // Shared game storage
	boards      map[string]*board.Board   // Shared board storage
	decks       map[string]*deck.GameDeck // Shared deck storage

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
		gameStorage:      game.NewGameStorage(),
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

	// Fetch game from storage and initialize session
	g, err := f.gameStorage.Get(gameID)
	var typesGame *types.Game
	if err == nil && g != nil {
		// Convert to types.Game
		typesGame = &types.Game{
			ID:               g.ID,
			CreatedAt:        g.CreatedAt,
			UpdatedAt:        g.UpdatedAt,
			Status:           types.GameStatus(g.Status),
			Settings:         g.Settings,
			HostPlayerID:     g.HostPlayerID,
			CurrentPhase:     types.GamePhase(g.CurrentPhase),
			GlobalParameters: g.GlobalParameters,
			CurrentTurn:      g.CurrentTurn,
			Generation:       g.Generation,
			Board:            g.Board,
			Players:          make(map[string]*types.Player),
		}

		// Wire repositories
		f.WireGameRepositories(typesGame)
	}

	// Create session with game and eventBus
	session := NewSession(typesGame, f.eventBus)

	f.sessions[gameID] = session
	return session
}

// Remove deletes a session from the factory
func (f *SessionFactoryImpl) Remove(gameID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.sessions, gameID)
}

// WireGameRepositories wires up a types.Game instance with game-scoped repositories and infrastructure
// Creates NEW repository instances bound to the specific game ID
// TODO: This will be gradually removed as we migrate to domain methods on Game/Player/Board
func (f *SessionFactoryImpl) WireGameRepositories(g *types.Game) {
	if g == nil {
		return
	}

	gameID := g.ID

	// Wire up infrastructure components
	// Note: During migration, we'll keep minimal infrastructure for card/deck operations
	// that haven't been migrated to domain methods yet
	deckRepo := deck.NewGameScopedRepository(gameID, f.decks, f.deckDefinitions)
	g.CardManager = card.NewCardManager(f.cardRepo, deckRepo, f.effectSubscriber)

	// TODO: TileProcessor and BonusCalculator will be migrated to domain methods
	// For now, keep them as infrastructure until we migrate tile placement logic
}
