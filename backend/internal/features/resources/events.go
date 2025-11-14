package resources

import "time"

// ResourcesChangedEvent is published when a player's resources change
type ResourcesChangedEvent struct {
	GameID       string
	PlayerID     string
	ResourceType string
	OldAmount    int
	NewAmount    int
	Timestamp    time.Time
}

// ProductionChangedEvent is published when a player's production changes
type ProductionChangedEvent struct {
	GameID        string
	PlayerID      string
	ResourceType  string
	OldProduction int
	NewProduction int
	Timestamp     time.Time
}
