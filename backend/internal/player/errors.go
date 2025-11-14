package player

import "fmt"

// ErrPlayerNotFound represents a player not found error
type ErrPlayerNotFound struct {
	PlayerID string
}

func (e *ErrPlayerNotFound) Error() string {
	return fmt.Sprintf("player not found: %s", e.PlayerID)
}
