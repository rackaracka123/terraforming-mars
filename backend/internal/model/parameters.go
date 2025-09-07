package model

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseSetup                 GamePhase = "setup"
	GamePhaseStartingCardSelection GamePhase = "starting_card_selection"
	GamePhaseCorporationSelection  GamePhase = "corporation_selection"
	GamePhaseAction                GamePhase = "action"
	GamePhaseProduction            GamePhase = "production"
	GamePhaseComplete              GamePhase = "complete"
)

// GameStatus represents the current status of the game
type GameStatus string

const (
	GameStatusLobby     GameStatus = "lobby"
	GameStatusActive    GameStatus = "active"
	GameStatusCompleted GameStatus = "completed"
)

// GameSettings contains configurable game parameters (all optional)
type GameSettings struct {
	MaxPlayers  int  `json:"maxPlayers,omitempty" ts:"number"`              // Default: 5
	Temperature *int `json:"temperature,omitempty" ts:"number | undefined"` // Default: -30°C
	Oxygen      *int `json:"oxygen,omitempty" ts:"number | undefined"`      // Default: 0%
	Oceans      *int `json:"oceans,omitempty" ts:"number | undefined"`      // Default: 0
}

// Default values for game settings
const (
	DefaultMaxPlayers  = 5
	DefaultTemperature = MinTemperature // -30°C
	DefaultOxygen      = MinOxygen      // 0%
	DefaultOceans      = MinOceans      // 0
)

// GlobalParameters represents the terraforming progress
type GlobalParameters struct {
	Temperature int `json:"temperature" ts:"number"` // Range: -30 to +8°C
	Oxygen      int `json:"oxygen" ts:"number"`      // Range: 0-14%
	Oceans      int `json:"oceans" ts:"number"`      // Range: 0-9
}

// Constants for terraforming limits
const (
	MinTemperature = -30
	MaxTemperature = 8
	MinOxygen      = 0
	MaxOxygen      = 14
	MinOceans      = 0
	MaxOceans      = 9
)

// DeepCopy creates a deep copy of the GameSettings
func (gs *GameSettings) DeepCopy() *GameSettings {
	if gs == nil {
		return nil
	}

	var tempCopy, oxygenCopy, oceansCopy *int
	if gs.Temperature != nil {
		temp := *gs.Temperature
		tempCopy = &temp
	}
	if gs.Oxygen != nil {
		oxygen := *gs.Oxygen
		oxygenCopy = &oxygen
	}
	if gs.Oceans != nil {
		oceans := *gs.Oceans
		oceansCopy = &oceans
	}

	return &GameSettings{
		MaxPlayers:  gs.MaxPlayers,
		Temperature: tempCopy,
		Oxygen:      oxygenCopy,
		Oceans:      oceansCopy,
	}
}

// DeepCopy creates a deep copy of the GlobalParameters
func (g *GlobalParameters) DeepCopy() *GlobalParameters {
	if g == nil {
		return nil
	}

	return &GlobalParameters{
		Temperature: g.Temperature,
		Oxygen:      g.Oxygen,
		Oceans:      g.Oceans,
	}
}
