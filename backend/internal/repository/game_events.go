package repository

import "time"

// Game-related domain events published by GameRepository
// These events represent granular changes to game state

// TemperatureChangedEvent is published when the global temperature parameter changes
type TemperatureChangedEvent struct {
	GameID    string
	OldValue  int
	NewValue  int
	ChangedBy string // PlayerID who triggered the change (empty if system-triggered)
	Timestamp time.Time
}

// OxygenChangedEvent is published when the global oxygen parameter changes
type OxygenChangedEvent struct {
	GameID    string
	OldValue  int
	NewValue  int
	ChangedBy string // PlayerID who triggered the change (empty if system-triggered)
	Timestamp time.Time
}

// OceansChangedEvent is published when the global oceans parameter changes
type OceansChangedEvent struct {
	GameID    string
	OldValue  int
	NewValue  int
	ChangedBy string // PlayerID who triggered the change (empty if system-triggered)
	Timestamp time.Time
}

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

// TilePlacedEvent is published when a tile is placed on the board
type TilePlacedEvent struct {
	GameID    string
	PlayerID  string // Player who placed the tile
	TileType  string
	Q         int // Hex coordinate Q
	R         int // Hex coordinate R
	S         int // Hex coordinate S
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
