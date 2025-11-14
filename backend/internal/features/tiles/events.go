package tiles

import "time"

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
