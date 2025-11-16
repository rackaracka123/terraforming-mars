package session

import (
	"context"
	"fmt"
	"sync"
)

// Repository provides storage and retrieval operations for active game sessions
// Sessions are runtime-only and never persisted to disk
type Repository interface {
	// Add registers a new session
	Add(ctx context.Context, session *Session) error

	// GetByID retrieves a session by game ID
	GetByID(ctx context.Context, gameID string) (*Session, error)

	// Remove deletes a session (called when game ends)
	Remove(ctx context.Context, gameID string) error

	// ListActive returns all active sessions
	ListActive(ctx context.Context) ([]*Session, error)

	// Exists checks if a session exists for the given game ID
	Exists(ctx context.Context, gameID string) bool
}

// RepositoryImpl implements Repository with in-memory storage
type RepositoryImpl struct {
	sessions map[string]*Session // gameID -> Session
	mu       sync.RWMutex
}

// NewRepository creates a new session repository
func NewRepository() Repository {
	return &RepositoryImpl{
		sessions: make(map[string]*Session),
	}
}

// Add registers a new session
func (r *RepositoryImpl) Add(ctx context.Context, session *Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	if session.GameID == "" {
		return fmt.Errorf("session gameID cannot be empty")
	}

	// Check if session already exists
	if _, exists := r.sessions[session.GameID]; exists {
		return fmt.Errorf("session already exists for game %s", session.GameID)
	}

	r.sessions[session.GameID] = session
	return nil
}

// GetByID retrieves a session by game ID
func (r *RepositoryImpl) GetByID(ctx context.Context, gameID string) (*Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[gameID]
	if !exists {
		return nil, fmt.Errorf("session not found for game %s", gameID)
	}

	// Update last activity
	session.UpdateActivity()

	return session, nil
}

// Remove deletes a session
func (r *RepositoryImpl) Remove(ctx context.Context, gameID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sessions[gameID]; !exists {
		return fmt.Errorf("session not found for game %s", gameID)
	}

	delete(r.sessions, gameID)
	return nil
}

// ListActive returns all active sessions
func (r *RepositoryImpl) ListActive(ctx context.Context) ([]*Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*Session, 0, len(r.sessions))
	for _, session := range r.sessions {
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// Exists checks if a session exists for the given game ID
func (r *RepositoryImpl) Exists(ctx context.Context, gameID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.sessions[gameID]
	return exists
}
