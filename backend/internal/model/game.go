package model

import "time"

// Game represents a unified game entity containing both metadata and state
type Game struct {
	ID               string           `json:"id" ts:"string"`
	CreatedAt        time.Time        `json:"createdAt" ts:"string"`
	UpdatedAt        time.Time        `json:"updatedAt" ts:"string"`
	Status           GameStatus       `json:"status" ts:"GameStatus"`
	Settings         GameSettings     `json:"settings" ts:"GameSettings"`
	Players          []Player         `json:"players" ts:"Player[]"`
	HostPlayerID     string           `json:"hostPlayerId" ts:"string"`
	CurrentPhase     GamePhase        `json:"currentPhase" ts:"GamePhase"`
	GlobalParameters GlobalParameters `json:"globalParameters" ts:"GlobalParameters"`
	CurrentPlayerID  string           `json:"currentPlayerId" ts:"string"`
	Generation       int              `json:"generation" ts:"number"`
	RemainingActions int              `json:"remainingActions" ts:"number"`
}


// NewGame creates a new game with the given settings
func NewGame(id string, settings GameSettings) *Game {
	now := time.Now()

	return &Game{
		ID:           id,
		CreatedAt:    now,
		UpdatedAt:    now,
		Status:       GameStatusLobby,
		Settings:     settings,
		Players:      make([]Player, 0),
		CurrentPhase: GamePhaseSetup,
		GlobalParameters: GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
		Generation:       1,
		RemainingActions: 0,
	}
}

// AddPlayer adds a player to the game
func (g *Game) AddPlayer(player Player) bool {
	if len(g.Players) >= g.Settings.MaxPlayers {
		return false
	}

	g.Players = append(g.Players, player)
	g.UpdatedAt = time.Now()

	return true
}

// GetPlayer returns a player by ID
func (g *Game) GetPlayer(playerID string) (*Player, bool) {
	for i := range g.Players {
		if g.Players[i].ID == playerID {
			return &g.Players[i], true
		}
	}
	return nil, false
}

// IsGameFull returns true if the game has reached maximum players
func (g *Game) IsGameFull() bool {
	return len(g.Players) >= g.Settings.MaxPlayers
}

// IsHost returns true if the given player ID is the host of the game
func (g *Game) IsHost(playerID string) bool {
	return g.HostPlayerID == playerID
}
