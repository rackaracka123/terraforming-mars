package types

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseWaitingForGameStart   GamePhase = "waiting_for_game_start"
	GamePhaseStartingCardSelection GamePhase = "starting_card_selection"
	GamePhaseStartGameSelection    GamePhase = "start_game_selection"
	GamePhaseAction                GamePhase = "action"
	GamePhaseProductionAndCardDraw GamePhase = "production_and_card_draw"
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
	MaxPlayers      int      // Default: 5
	Temperature     *int     // Default: -30°C
	Oxygen          *int     // Default: 0%
	Oceans          *int     // Default: 0
	DevelopmentMode bool     // Default: false
	CardPacks       []string // Default: ["base-game"]
}

// Card pack constants
const (
	PackBaseGame = "base-game" // Tested simple cards only
	PackFuture   = "future"    // Untested/complex cards for future implementation
)

// Default values for game settings
const (
	DefaultMaxPlayers  = 5
	DefaultTemperature = MinTemperature // -30°C
	DefaultOxygen      = MinOxygen      // 0%
	DefaultOceans      = MinOceans      // 0
)

// DefaultCardPacks returns the default card packs
func DefaultCardPacks() []string {
	return []string{PackBaseGame}
}

// GlobalParameters represents the terraforming progress
type GlobalParameters struct {
	Temperature int // Range: -30 to +8°C
	Oxygen      int // Range: 0-14%
	Oceans      int // Range: 0-9
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

// RaiseTemperature raises the temperature by the specified number of steps
// Returns the actual number of steps raised (may be less if limit reached)
func (gp *GlobalParameters) RaiseTemperature(steps int) int {
	oldValue := gp.Temperature
	newValue := gp.Temperature + (steps * 2) // Each step is 2 degrees
	if newValue > MaxTemperature {
		newValue = MaxTemperature
	}
	gp.Temperature = newValue
	actualSteps := (newValue - oldValue) / 2
	return actualSteps
}

// RaiseOxygen raises the oxygen by the specified number of steps
// Returns the actual number of steps raised (may be less if limit reached)
func (gp *GlobalParameters) RaiseOxygen(steps int) int {
	oldValue := gp.Oxygen
	newValue := gp.Oxygen + steps
	if newValue > MaxOxygen {
		newValue = MaxOxygen
	}
	gp.Oxygen = newValue
	return newValue - oldValue
}

// PlaceOcean places an ocean tile (increments ocean count)
// Returns true if successful, false if limit reached
func (gp *GlobalParameters) PlaceOcean() bool {
	if gp.Oceans >= MaxOceans {
		return false
	}
	gp.Oceans++
	return true
}

// IsMaxed returns true if all global parameters have reached their maximum values
func (gp *GlobalParameters) IsMaxed() bool {
	return gp.Temperature >= MaxTemperature &&
		gp.Oxygen >= MaxOxygen &&
		gp.Oceans >= MaxOceans
}
