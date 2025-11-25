package session

import (
	"context"
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

	// Repository instances (shared across sessions)
	gameRepo  game.Repository
	boardRepo board.Repository
	cardRepo  card.Repository
	deckRepo  deck.Repository

	// Infrastructure component factories
	effectSubscriber card.CardEffectSubscriber
	boardProcessor   *board.BoardProcessor

	mu sync.RWMutex
}

// NewSessionFactory creates a new session factory with repository dependencies
func NewSessionFactory(
	eventBus *events.EventBusImpl,
	gameRepo game.Repository,
	boardRepo board.Repository,
	cardRepo card.Repository,
	deckRepo deck.Repository,
) SessionFactory {
	// Create shared infrastructure components
	effectSubscriber := card.NewCardEffectSubscriber(eventBus, cardRepo)
	boardProcessor := board.NewBoardProcessor()

	return &SessionFactoryImpl{
		sessions:         make(map[string]*Session),
		eventBus:         eventBus,
		gameRepo:         gameRepo,
		boardRepo:        boardRepo,
		cardRepo:         cardRepo,
		deckRepo:         deckRepo,
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

	// Fetch game from repository and initialize session
	g, err := f.gameRepo.GetByID(context.Background(), gameID)
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

// WireGameRepositories wires up a types.Game instance with repositories and infrastructure
// This should be called when setting a Game on a Session to ensure it has access to all dependencies
func (f *SessionFactoryImpl) WireGameRepositories(g *types.Game) {
	if g == nil {
		return
	}

	// Wire up sub-repositories from game repository
	if gameRepoImpl, ok := f.gameRepo.(*game.RepositoryImpl); ok {
		g.Core = gameRepoImpl.GetCore()
		g.GlobalParams = gameRepoImpl.GetGlobalParams()
		g.Turn = gameRepoImpl.GetTurn()
	}

	// Wire up domain repositories
	g.Board = f.boardRepo
	g.Cards = f.cardRepo
	g.Deck = f.deckRepo

	// Wire up infrastructure components
	g.CardManager = card.NewCardManager(f.cardRepo, f.deckRepo, f.effectSubscriber)
	g.TileProcessor = board.NewProcessor(f.gameRepo, f.boardRepo, f.boardProcessor)
	g.BonusCalculator = board.NewBonusCalculator(f.gameRepo, f.boardRepo, f.deckRepo)
}
