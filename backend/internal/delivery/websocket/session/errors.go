package session

import "fmt"

// SessionNotFoundError represents a session-related not found error
type SessionNotFoundError struct {
	Resource string
	ID       string
}

func (e *SessionNotFoundError) Error() string {
	return fmt.Sprintf("%s %s not found", e.Resource, e.ID)
}
