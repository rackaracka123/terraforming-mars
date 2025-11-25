package session

import (
	"sync"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/types"
)

// Session represents a game session with all its players and game state.
// Each session is bound to a specific gameID and manages all state for that game.
type Session struct {
	gameID   string
	Players  map[string]*player.Player // playerID -> Player with repositories
	Game     *types.Game               // Game instance with sub-repos
	eventBus *events.EventBusImpl
	mu       sync.RWMutex
}

// NewSession creates a new session for a specific game
func NewSession(gameID string, eventBus *events.EventBusImpl) *Session {
	return &Session{
		gameID:   gameID,
		Players:  make(map[string]*player.Player),
		eventBus: eventBus,
		mu:       sync.RWMutex{},
	}
}

// GetGameID returns the game ID this session is bound to
func (s *Session) GetGameID() string {
	return s.gameID
}

// AddPlayer adds a player to the session with wired repositories
func (s *Session) AddPlayer(p *types.Player) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure player has gameID set
	p.GameID = s.gameID

	// Wrap player with repositories
	playerWithRepos := player.NewPlayer(p, s.eventBus)
	s.Players[p.ID] = playerWithRepos
}

// RemovePlayer removes a player from the session
func (s *Session) RemovePlayer(playerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.Players, playerID)
}

// GetPlayer retrieves a player by ID with read lock
func (s *Session) GetPlayer(playerID string) (*player.Player, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, exists := s.Players[playerID]
	return p, exists
}

// GetAllPlayers returns all players in the session
func (s *Session) GetAllPlayers() []*player.Player {
	s.mu.RLock()
	defer s.mu.RUnlock()

	players := make([]*player.Player, 0, len(s.Players))
	for _, p := range s.Players {
		players = append(players, p)
	}
	return players
}

// SetGame sets the game instance for this session
func (s *Session) SetGame(game *types.Game) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Game = game
}
