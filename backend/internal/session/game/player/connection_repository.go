package player

import (
	"context"
)

// ConnectionRepository handles player connection status
// Auto-saves changes after every operation
type ConnectionRepository struct {
	player *Player // Reference to parent player
}

// NewConnectionRepository creates a new connection repository for a specific player
func NewConnectionRepository(player *Player) *ConnectionRepository {
	return &ConnectionRepository{
		player: player,
	}
}

// UpdateStatus updates player connection status
// Auto-saves changes to the player
func (r *ConnectionRepository) UpdateStatus(ctx context.Context, isConnected bool) error {
	r.player.IsConnected = isConnected
	return nil
}
