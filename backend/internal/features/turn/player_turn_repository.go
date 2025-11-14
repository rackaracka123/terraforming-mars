package turn

import (
	"context"
	"sync"
)

// PlayerTurnRepository manages player-level turn state
type PlayerTurnRepository interface {
	// Get player turn state
	GetPassed(ctx context.Context) (bool, error)
	GetAvailableActions(ctx context.Context) (int, error)

	// Granular operations
	SetPassed(ctx context.Context, passed bool) error
	SetAvailableActions(ctx context.Context, actions int) error
	DecrementAvailableActions(ctx context.Context) error
}

// PlayerTurnRepositoryImpl implements independent in-memory storage for player turn state
type PlayerTurnRepositoryImpl struct {
	mu               sync.RWMutex
	passed           bool
	availableActions int
}

// NewPlayerTurnRepository creates a new independent player turn repository with initial state
func NewPlayerTurnRepository(passed bool, availableActions int) PlayerTurnRepository {
	return &PlayerTurnRepositoryImpl{
		passed:           passed,
		availableActions: availableActions,
	}
}

// GetPassed retrieves whether the player has passed
func (r *PlayerTurnRepositoryImpl) GetPassed(ctx context.Context) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.passed, nil
}

// GetAvailableActions retrieves the number of available actions
func (r *PlayerTurnRepositoryImpl) GetAvailableActions(ctx context.Context) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.availableActions, nil
}

// SetPassed sets whether the player has passed
func (r *PlayerTurnRepositoryImpl) SetPassed(ctx context.Context, passed bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.passed = passed
	return nil
}

// SetAvailableActions sets the number of available actions
func (r *PlayerTurnRepositoryImpl) SetAvailableActions(ctx context.Context, actions int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.availableActions = actions
	return nil
}

// DecrementAvailableActions reduces available actions by 1
func (r *PlayerTurnRepositoryImpl) DecrementAvailableActions(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.availableActions > 0 {
		r.availableActions--
	}

	return nil
}
