package model

import "time"

// Game represents a unified game entity containing both metadata and state
type Game struct {
	ID               string           `json:"id" ts:"string"`
	CreatedAt        time.Time        `json:"createdAt" ts:"string"`
	UpdatedAt        time.Time        `json:"updatedAt" ts:"string"`
	Status           GameStatus       `json:"status" ts:"GameStatus"`
	Settings         GameSettings     `json:"settings" ts:"GameSettings"`
	CurrentPlayer    *Player          `json:"currentPlayer" ts:"Player"`      // Full player data for the viewing player
	OtherPlayers     []OtherPlayer    `json:"otherPlayers" ts:"OtherPlayer[]"` // Limited data for other players
	Players          []Player         `json:"players" ts:"Player[]"`           // Internal - full data for all players
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
		CurrentPhase: GamePhaseWaitingForGameStart,
		GlobalParameters: GlobalParameters{
			Temperature: -30,
			Oxygen:      0,
			Oceans:      0,
		},
		Generation:       1,
		RemainingActions: 0,
	}
}

