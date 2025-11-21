package types

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

// NotFoundError represents a generic resource not found error
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}

// SessionNotFoundError represents a session-related not found error
type SessionNotFoundError struct {
	Resource string
	ID       string
}

func (e *SessionNotFoundError) Error() string {
	return fmt.Sprintf("%s %s not found", e.Resource, e.ID)
}
