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

// CanIncreaseTemperature checks if temperature can be increased by the given steps
func (g *GlobalParameters) CanIncreaseTemperature(steps int) bool {
	return steps >= 0 && (g.Temperature < MaxTemperature || steps == 0)
}

// IncreaseTemperature increases temperature by the given steps, capped at maximum
func (g *GlobalParameters) IncreaseTemperature(steps int) {
	newTemp := g.Temperature + steps*2 // Each step = 2°C
	if newTemp > MaxTemperature {
		g.Temperature = MaxTemperature
	} else {
		g.Temperature = newTemp
	}
}

// CanIncreaseOxygen checks if oxygen can be increased by the given percentage
func (g *GlobalParameters) CanIncreaseOxygen(percent int) bool {
	return percent >= 0 && (g.Oxygen < MaxOxygen || percent == 0)
}

// IncreaseOxygen increases oxygen by the given percentage, capped at maximum
func (g *GlobalParameters) IncreaseOxygen(percent int) {
	newOxygen := g.Oxygen + percent
	if newOxygen > MaxOxygen {
		g.Oxygen = MaxOxygen
	} else {
		g.Oxygen = newOxygen
	}
}

// CanPlaceOcean checks if oceans can be placed
func (g *GlobalParameters) CanPlaceOcean(count int) bool {
	return count >= 0 && (g.Oceans < MaxOceans || count == 0)
}

// PlaceOcean places the given number of oceans, capped at maximum
func (g *GlobalParameters) PlaceOcean(count int) {
	newOceans := g.Oceans + count
	if newOceans > MaxOceans {
		g.Oceans = MaxOceans
	} else {
		g.Oceans = newOceans
	}
}

// IsFullyTerraformed checks if all terraforming parameters are at maximum
func (g *GlobalParameters) IsFullyTerraformed() bool {
	return g.Temperature == MaxTemperature && g.Oxygen == MaxOxygen && g.Oceans == MaxOceans
}

// GetTerraformingProgress returns the overall terraforming progress as a percentage
func (g *GlobalParameters) GetTerraformingProgress() float64 {
	tempProgress := float64(g.Temperature-MinTemperature) / float64(MaxTemperature-MinTemperature)
	oxygenProgress := float64(g.Oxygen-MinOxygen) / float64(MaxOxygen-MinOxygen)
	oceanProgress := float64(g.Oceans-MinOceans) / float64(MaxOceans-MinOceans)

	return (tempProgress + oxygenProgress + oceanProgress) / 3.0 * 100.0
}
