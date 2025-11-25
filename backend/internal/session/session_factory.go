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

	session := NewSession(gameID, f.eventBus)

	// Fetch game from storage and initialize session
	g, err := f.gameStorage.Get(gameID)
	if err == nil && g != nil {
		// Convert to types.Game
		typesGame := &types.Game{
			ID:               g.ID,
			CreatedAt:        g.CreatedAt,
			UpdatedAt:        g.UpdatedAt,
			Status:           types.GameStatus(g.Status),
			Settings:         g.Settings,
			PlayerIDs:        g.PlayerIDs,
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

		// Set on session
		session.SetGame(typesGame)
	}

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
func (f *SessionFactoryImpl) WireGameRepositories(g *types.Game) {
	if g == nil {
		return
	}

	gameID := g.ID

	// Create game-scoped repository instances
	g.Core = game.NewGameCoreRepository(gameID, f.gameStorage, f.eventBus)
	g.GlobalParams = game.NewGameGlobalParametersRepository(gameID, f.gameStorage, f.eventBus)
	g.Turn = game.NewGameTurnRepository(gameID, f.gameStorage, f.eventBus)
	g.Board = board.NewRepository(gameID, f.boards, f.eventBus)
	g.Deck = deck.NewGameScopedRepository(gameID, f.decks, f.deckDefinitions)

	// Wire up global repositories
	g.Cards = f.cardRepo

	// Wire up infrastructure components with game-scoped repositories
	g.CardManager = card.NewCardManager(f.cardRepo, g.Deck.(deck.Repository), f.effectSubscriber)
	// Note: TileProcessor and BonusCalculator will need updating since we removed facade
	// For now, commenting out - we'll fix when we get to actions
	// g.TileProcessor = board.NewProcessor(?, g.Board.(board.Repository), f.boardProcessor)
	// g.BonusCalculator = board.NewBonusCalculator(?, g.Board.(board.Repository), g.Deck.(deck.Repository))
}
