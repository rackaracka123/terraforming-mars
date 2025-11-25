package session

import (
	"sync"

	"terraforming-mars-backend/internal/events"
)

// SessionFactory manages creation and retrieval of game sessions
type SessionFactory interface {
	Get(gameID string) *Session
	GetOrCreate(gameID string) *Session
	Remove(gameID string)
}

// SessionFactoryImpl implements SessionFactory
type SessionFactoryImpl struct {
	sessions map[string]*Session
	eventBus *events.EventBusImpl
	mu       sync.RWMutex
}

// NewSessionFactory creates a new session factory
func NewSessionFactory(eventBus *events.EventBusImpl) SessionFactory {
	return &SessionFactoryImpl{
		sessions: make(map[string]*Session),
		eventBus: eventBus,
		mu:       sync.RWMutex{},
	}
}

// Get retrieves an existing session by game ID
// Returns nil if session doesn't exist
func (f *SessionFactoryImpl) Get(gameID string) *Session {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.sessions[gameID]
}

// GetOrCreate retrieves an existing session or creates a new one if it doesn't exist
func (f *SessionFactoryImpl) GetOrCreate(gameID string) *Session {
	// Try to get existing session first with read lock
	f.mu.RLock()
	if session, exists := f.sessions[gameID]; exists {
		f.mu.RUnlock()
		return session
	}
	f.mu.RUnlock()

	// Create new session with write lock
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have created it)
	if session, exists := f.sessions[gameID]; exists {
		return session
	}

	session := NewSession(gameID, f.eventBus)
	f.sessions[gameID] = session
	return session
}

// Remove deletes a session from the factory
func (f *SessionFactoryImpl) Remove(gameID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.sessions, gameID)
}
