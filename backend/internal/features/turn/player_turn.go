package turn

import (
	"context"
	"fmt"
	"sync"

	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// ============================================================================
// REPOSITORY
// ============================================================================

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

// ============================================================================
// SERVICE
// ============================================================================

// PlayerTurnService handles player-level turn state
//
// Scope: Isolated turn state management for a player
//   - Passed status
//   - Available actions tracking
type PlayerTurnService interface {
	GetPassed(ctx context.Context) (bool, error)
	GetAvailableActions(ctx context.Context) (int, error)

	SetPassed(ctx context.Context, passed bool) error
	SetAvailableActions(ctx context.Context, actions int) error
	DecrementAvailableActions(ctx context.Context) error
}

// PlayerTurnServiceImpl implements the player turn service
type PlayerTurnServiceImpl struct {
	repo PlayerTurnRepository
}

// NewPlayerTurnService creates a new player turn service
func NewPlayerTurnService(repo PlayerTurnRepository) PlayerTurnService {
	return &PlayerTurnServiceImpl{
		repo: repo,
	}
}

// GetPassed retrieves whether the player has passed
func (s *PlayerTurnServiceImpl) GetPassed(ctx context.Context) (bool, error) {
	return s.repo.GetPassed(ctx)
}

// GetAvailableActions retrieves the number of available actions
func (s *PlayerTurnServiceImpl) GetAvailableActions(ctx context.Context) (int, error) {
	return s.repo.GetAvailableActions(ctx)
}

// SetPassed sets whether the player has passed
func (s *PlayerTurnServiceImpl) SetPassed(ctx context.Context, passed bool) error {
	if err := s.repo.SetPassed(ctx, passed); err != nil {
		return fmt.Errorf("failed to set passed: %w", err)
	}

	logger.Get().Debug("Player passed status set", zap.Bool("passed", passed))

	return nil
}

// SetAvailableActions sets the number of available actions
func (s *PlayerTurnServiceImpl) SetAvailableActions(ctx context.Context, actions int) error {
	if err := s.repo.SetAvailableActions(ctx, actions); err != nil {
		return fmt.Errorf("failed to set available actions: %w", err)
	}

	logger.Get().Debug("Available actions set", zap.Int("actions", actions))

	return nil
}

// DecrementAvailableActions reduces available actions by 1
func (s *PlayerTurnServiceImpl) DecrementAvailableActions(ctx context.Context) error {
	if err := s.repo.DecrementAvailableActions(ctx); err != nil {
		return fmt.Errorf("failed to decrement available actions: %w", err)
	}

	currentActions, _ := s.repo.GetAvailableActions(ctx)
	logger.Get().Debug("Available actions decremented", zap.Int("remaining", currentActions))

	return nil
}
