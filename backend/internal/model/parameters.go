package model

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseSetup                    GamePhase = "setup"
	GamePhaseStartingCardSelection    GamePhase = "starting_card_selection"
	GamePhaseCorporationSelection     GamePhase = "corporation_selection"
	GamePhaseAction                   GamePhase = "action"
	GamePhaseProduction               GamePhase = "production"
	GamePhaseComplete                 GamePhase = "complete"
)

// GameStatus represents the current status of the game
type GameStatus string

const (
	GameStatusLobby     GameStatus = "lobby"
	GameStatusActive    GameStatus = "active"
	GameStatusCompleted GameStatus = "completed"
)

// GameSettings contains configurable game parameters
type GameSettings struct {
	MaxPlayers int `json:"maxPlayers" ts:"number"`
}

// GlobalParameters represents the terraforming progress
type GlobalParameters struct {
	Temperature int `json:"temperature" ts:"number"` // Range: -30 to +8°C
	Oxygen      int `json:"oxygen" ts:"number"`      // Range: 0-14%
	Oceans      int `json:"oceans" ts:"number"`      // Range: 0-9
}