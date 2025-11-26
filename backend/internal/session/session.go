package session

import (
	"sync"

	"github.com/google/uuid"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/player"
)

// Session represents a game session with all its players and game state.
// Session owns the EventBus for all events in the game.
// Game object contains the game ID and owns all players.
type Session struct {
	game     *game.Game           // Private game instance owns all players and game state
	eventBus *events.EventBusImpl // EventBus for all events in this game
	mu       sync.RWMutex         // Session-level mutex for thread safety
}

// NewSession creates a new session with a game instance
func NewSession(g *game.Game, eventBus *events.EventBusImpl) *Session {
	return &Session{
		game:     g,
		eventBus: eventBus,
		mu:       sync.RWMutex{},
	}
}

// Game returns the game instance (read-only access)
func (s *Session) Game() *game.Game {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.game
}

// setGame sets the game instance for this session (private setter)
func (s *Session) setGame(g *game.Game) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.game = g
}

// GetGameID returns the game ID from the Game object
func (s *Session) GetGameID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.game == nil {
		return ""
	}
	return s.game.ID
}

// AddPlayer adds a player to the session
func (s *Session) AddPlayer(p *player.Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.game.AddPlayer(p)
}

// CreateAndAddPlayer creates a new player and adds it to the session
// playerID is optional - if empty, a new UUID will be generated
// Returns the created player
// DEPRECATED: Consider using player.Factory directly instead
func (s *Session) CreateAndAddPlayer(playerName string, playerID ...string) *player.Player {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Use provided playerID or generate new one
	var pid string
	if len(playerID) > 0 && playerID[0] != "" {
		pid = playerID[0]
	} else {
		pid = uuid.New().String()
	}

	// Create player factory and create player
	factory := player.NewFactory(s.eventBus)
	p := factory.CreatePlayer(s.game.ID, pid, playerName)

	// Set connection status to true for new players
	_ = p.SetConnectionStatus(nil, true)

	// Add to game
	_ = s.game.AddPlayer(p)

	// Return the player
	player, _ := s.game.GetPlayer(pid)
	return player
}

// RemovePlayer removes a player from the session
func (s *Session) RemovePlayer(playerID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.game.RemovePlayer(playerID)
}

// GetPlayer retrieves a player by ID with read lock
func (s *Session) GetPlayer(playerID string) (*player.Player, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, err := s.game.GetPlayer(playerID)
	if err != nil {
		return nil, false
	}
	return p, true
}

// GetAllPlayers returns all players in the session
func (s *Session) GetAllPlayers() []*player.Player {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.game.GetAllPlayers()
}
