package session

import (
	"sync"

	"github.com/google/uuid"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session/game/player"
	"terraforming-mars-backend/internal/session/types"
)

// Session represents a game session with all its players and game state.
// Each session is bound to a specific gameID and manages all state for that game.
// Session now owns a Game which in turn owns all players and repositories.
type Session struct {
	gameID   string
	Game     *types.Game // Game instance owns repositories, infrastructure, and players
	eventBus *events.EventBusImpl
	mu       sync.RWMutex // Session-level mutex for thread safety
}

// NewSession creates a new session for a specific game
func NewSession(gameID string, eventBus *events.EventBusImpl) *Session {
	return &Session{
		gameID:   gameID,
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

	// Initialize Game.Players map if needed and cast to proper type
	if s.Game.Players == nil {
		s.Game.Players = make(map[string]*types.Player)
	}

	// Store as interface{} and cast to player.Player map
	playersMap, ok := s.Game.Players.(map[string]*player.Player)
	if !ok {
		// First time: convert from types.Player map to player.Player map
		playersMap = make(map[string]*player.Player)
		s.Game.Players = playersMap
	}
	playersMap[p.ID] = playerWithRepos
}

// CreateAndAddPlayer creates a new player and adds it to the session
// playerID is optional - if empty, a new UUID will be generated
// Returns the player with wired repositories
func (s *Session) CreateAndAddPlayer(playerName string, playerID ...string) *player.Player {
	// Use provided playerID or generate new one
	var pid string
	if len(playerID) > 0 && playerID[0] != "" {
		pid = playerID[0]
	} else {
		pid = uuid.New().String()
	}

	// Create player with initial state
	p := &types.Player{
		ID:     pid,
		Name:   playerName,
		GameID: s.gameID,
		Resources: types.Resources{
			Credits:  0,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		Production: types.Production{
			Credits:  0,
			Steel:    0,
			Titanium: 0,
			Plants:   0,
			Energy:   0,
			Heat:     0,
		},
		TerraformRating:  20,
		Cards:            []string{},
		PlayedCards:      []string{},
		Passed:           false,
		AvailableActions: 0,
		VictoryPoints:    0,
		IsConnected:      true,
		Effects:          []types.PlayerEffect{},
		Actions:          []types.PlayerAction{},
		ResourceStorage:  make(map[string]int),
	}

	// Add to session (wraps with repositories)
	s.AddPlayer(p)

	// Return the wrapped player
	wrappedPlayer, _ := s.GetPlayer(pid)
	return wrappedPlayer
}

// RemovePlayer removes a player from the session
func (s *Session) RemovePlayer(playerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if playersMap, ok := s.Game.Players.(map[string]*player.Player); ok {
		delete(playersMap, playerID)
	}
}

// GetPlayer retrieves a player by ID with read lock
func (s *Session) GetPlayer(playerID string) (*player.Player, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if playersMap, ok := s.Game.Players.(map[string]*player.Player); ok {
		p, exists := playersMap[playerID]
		return p, exists
	}
	return nil, false
}

// GetAllPlayers returns all players in the session
func (s *Session) GetAllPlayers() []*player.Player {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if playersMap, ok := s.Game.Players.(map[string]*player.Player); ok {
		players := make([]*player.Player, 0, len(playersMap))
		for _, p := range playersMap {
			players = append(players, p)
		}
		return players
	}
	return []*player.Player{}
}

// SetGame sets the game instance for this session
func (s *Session) SetGame(game *types.Game) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Game = game
}
