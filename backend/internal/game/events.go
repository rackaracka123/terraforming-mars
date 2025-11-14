package game

import "time"

// Game-related domain events published by game repository
// These events represent changes to game-level state (not feature-level state)

// GamePhaseChangedEvent is published when the game phase transitions
type GamePhaseChangedEvent struct {
	GameID    string
	OldPhase  string
	NewPhase  string
	Timestamp time.Time
}

// GameStatusChangedEvent is published when the game status changes (lobby, active, completed)
type GameStatusChangedEvent struct {
	GameID    string
	OldStatus string
	NewStatus string
	Timestamp time.Time
}

// GenerationAdvancedEvent is published when the game generation advances
type GenerationAdvancedEvent struct {
	GameID        string
	OldGeneration int
	NewGeneration int
	Timestamp     time.Time
}

// PlacementBonusGainedEvent is published when a player gains resources from tile placement bonuses
type PlacementBonusGainedEvent struct {
	GameID    string
	PlayerID  string
	Resources map[string]int // Map of resource type to amount (e.g., {"steel": 2, "titanium": 1})
	Q         int            // Hex coordinate Q
	R         int            // Hex coordinate R
	S         int            // Hex coordinate S
	Timestamp time.Time
}
