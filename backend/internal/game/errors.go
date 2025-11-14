package game

import "fmt"

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
