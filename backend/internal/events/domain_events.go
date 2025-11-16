package events

import "time"

// Domain Events for Terraforming Mars
// All event type definitions are centralized here to avoid circular dependencies
// Features and repositories publish these events, subscribers consume them

// =============================================================================
// PLAYER EVENTS
// =============================================================================

// ResourcesChangedEvent is published when a player's resources change
type ResourcesChangedEvent struct {
	GameID       string
	PlayerID     string
	ResourceType string // "credits", "steel", "titanium", "plants", "energy", "heat"
	OldAmount    int
	NewAmount    int
	Timestamp    time.Time
}

// ProductionChangedEvent is published when a player's production changes
type ProductionChangedEvent struct {
	GameID        string
	PlayerID      string
	ResourceType  string // "credits", "steel", "titanium", "plants", "energy", "heat"
	OldProduction int
	NewProduction int
	Timestamp     time.Time
}

// TerraformRatingChangedEvent is published when a player's terraform rating changes
type TerraformRatingChangedEvent struct {
	GameID    string
	PlayerID  string
	OldRating int
	NewRating int
	Timestamp time.Time
}

// CorporationSelectedEvent is published when a player selects their corporation
type CorporationSelectedEvent struct {
	GameID          string
	PlayerID        string
	CorporationID   string
	CorporationName string
	Timestamp       time.Time
}

// CardPlayedEvent is published when a player plays a card
type CardPlayedEvent struct {
	GameID    string
	PlayerID  string
	CardID    string
	CardName  string
	CardType  string // Type of card played (event, automated, active, corporation, prelude)
	Timestamp time.Time
}

// CardAddedToHandEvent is published when a card is added to a player's hand
type CardAddedToHandEvent struct {
	GameID    string
	PlayerID  string
	CardID    string
	Timestamp time.Time
}

// VictoryPointsChangedEvent is published when a player's victory points change
type VictoryPointsChangedEvent struct {
	GameID    string
	PlayerID  string
	OldPoints int
	NewPoints int
	Source    string // What caused the change (e.g., "card", "milestone", "award")
	Timestamp time.Time
}

// PlayerEffectAddedEvent is published when a passive effect is added to a player
type PlayerEffectAddedEvent struct {
	GameID     string
	PlayerID   string
	CardID     string
	CardName   string
	EffectType string
	Timestamp  time.Time
}

// ResourceStorageChangedEvent is published when resource storage on a card changes
type ResourceStorageChangedEvent struct {
	GameID       string
	PlayerID     string
	CardID       string
	ResourceType string
	OldAmount    int
	NewAmount    int
	Timestamp    time.Time
}

// =============================================================================
// GAME EVENTS
// =============================================================================

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

// =============================================================================
// GLOBAL PARAMETER EVENTS
// =============================================================================

// TemperatureChangedEvent is published when temperature changes
type TemperatureChangedEvent struct {
	GameID    string
	OldValue  int
	NewValue  int
	ChangedBy string
	Timestamp time.Time
}

// OxygenChangedEvent is published when oxygen changes
type OxygenChangedEvent struct {
	GameID    string
	OldValue  int
	NewValue  int
	ChangedBy string
	Timestamp time.Time
}

// OceansChangedEvent is published when ocean count changes
type OceansChangedEvent struct {
	GameID    string
	OldValue  int
	NewValue  int
	ChangedBy string
	Timestamp time.Time
}

// =============================================================================
// TILE EVENTS
// =============================================================================

// TilePlacedEvent is published when a tile is placed
type TilePlacedEvent struct {
	GameID    string
	PlayerID  string
	TileType  string
	Q         int
	R         int
	S         int
	Timestamp time.Time
}
