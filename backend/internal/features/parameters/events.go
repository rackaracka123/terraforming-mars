package parameters

import "time"

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
