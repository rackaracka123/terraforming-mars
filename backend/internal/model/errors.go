package model

import "fmt"

// ErrPlayerNotFound represents a player not found error
type ErrPlayerNotFound struct {
	PlayerID string
}

func (e *ErrPlayerNotFound) Error() string {
	return fmt.Sprintf("player not found: %s", e.PlayerID)
}

// ErrGameNotFound represents a game not found error
type ErrGameNotFound struct {
	GameID string
}

func (e *ErrGameNotFound) Error() string {
	return fmt.Sprintf("game not found: %s", e.GameID)
}
