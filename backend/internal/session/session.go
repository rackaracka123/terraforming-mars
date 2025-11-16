package session

import (
	"time"

	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"
)

// Session is a runtime aggregate that wraps Player, Game, and feature services for a single active game instance.
// While Game and Player are domain entities representing game state, Session is the operational container
// that manages the lifecycle and coordination of an active game.
//
// Session vs Game/Player:
// - Game and Player are domain entities (persistent game state, stored in repositories)
// - Session is a runtime aggregate (created when game starts, destroyed when game ends, never persisted)
//
// Key Responsibilities:
// - Unified game access (single object contains all game state and services)
// - Clean lifecycle management (created on start, destroyed on end)
// - Scoped feature services (each session has its own feature service instances)
// - Simplified actions (no need to fetch Game, Players, and services separately)
type Session struct {
	// Core aggregates
	GameID  string                    `json:"gameId"`
	Game    *game.Game                `json:"game"`
	Players map[string]*player.Player `json:"players"` // playerID -> Player

	// Feature service instances (scoped to this game)
	ParametersService parameters.Service    `json:"-"` // Global parameters (temperature, oxygen, oceans)
	BoardService      tiles.BoardService    `json:"-"` // Board with tile placement
	CardService       interface{}           `json:"-"` // Card operations (stub interface for now)
	TurnOrderService  turn.TurnOrderService `json:"-"` // Current turn and player rotation

	// Game rule subscribers (lifecycle-managed, created on game start, cleaned up on game end)
	GreeneryRuleSubscriber interface{} `json:"-"` // Subscriber for greenery → O2 → TR rule

	// Session metadata
	CreatedAt    time.Time `json:"createdAt"`
	LastActivity time.Time `json:"lastActivity"`
	HostPlayerID string    `json:"hostPlayerId"`
}

// NewSession creates a new session for an active game
// Feature services must be provided during creation
func NewSession(
	gameID string,
	game *game.Game,
	players map[string]*player.Player,
	parametersService parameters.Service,
	boardService tiles.BoardService,
	cardService interface{},
	turnOrderService turn.TurnOrderService,
	greeneryRuleSubscriber interface{},
	hostPlayerID string,
) *Session {
	now := time.Now()

	return &Session{
		GameID:                 gameID,
		Game:                   game,
		Players:                players,
		ParametersService:      parametersService,
		BoardService:           boardService,
		CardService:            cardService,
		TurnOrderService:       turnOrderService,
		GreeneryRuleSubscriber: greeneryRuleSubscriber,
		CreatedAt:              now,
		LastActivity:           now,
		HostPlayerID:           hostPlayerID,
	}
}

// UpdateActivity updates the last activity timestamp
func (s *Session) UpdateActivity() {
	s.LastActivity = time.Now()
}

// GetPlayer retrieves a player by ID from the session
func (s *Session) GetPlayer(playerID string) (*player.Player, bool) {
	p, exists := s.Players[playerID]
	return p, exists
}

// PlayerIDs returns a slice of all player IDs in the session
func (s *Session) PlayerIDs() []string {
	ids := make([]string, 0, len(s.Players))
	for id := range s.Players {
		ids = append(ids, id)
	}
	return ids
}

// GetTileQueueService returns the tile queue service for a specific player
// This provides access to player-scoped tile queue without player having service dependency
func (s *Session) GetTileQueueService(playerID string) (interface{}, error) {
	// This will be implemented when we integrate with the actual tile queue repositories
	// For now, return nil to allow compilation
	return nil, nil
}

// GetPlayerTurnService returns the player turn service for a specific player
// This provides access to player-scoped turn state without player having service dependency
func (s *Session) GetPlayerTurnService(playerID string) (interface{}, error) {
	// This will be implemented when we integrate with the actual player turn repositories
	// For now, return nil to allow compilation
	return nil, nil
}
