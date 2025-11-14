package turn

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

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
