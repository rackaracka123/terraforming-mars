package errors

import "fmt"

// NotFoundError represents a resource not found error
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