package player

import "time"

// TileQueueCreatedEvent is published when a tile placement queue is created for a player.
// This typically happens after playing a card with tile placement effects.
type TileQueueCreatedEvent struct {
	GameID    string    // ID of the game
	PlayerID  string    // ID of the player who owns the tile queue
	QueueSize int       // Number of tiles in the queue
	Source    string    // Card ID that created the queue
	Timestamp time.Time // When the queue was created
}

// CardDrawConfirmedEvent is published when a player confirms their card draw selection.
// This happens after drawing cards from various sources (corporations, card effects, etc.).
type CardDrawConfirmedEvent struct {
	GameID   string   // ID of the game
	PlayerID string   // ID of the player who confirmed the draw
	Source   string   // Card/Corporation ID that triggered the draw
	Cards    []string // Card IDs that were selected/confirmed
}
