package player

import "time"

// Player-related domain events published by repositories
// These events represent granular changes to player state

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
