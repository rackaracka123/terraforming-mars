package production

import (
	"context"
	"fmt"
)

// Service handles production phase card selection state
//
// Scope: Isolated production phase UI state for a single player
//   - Track available cards for selection
//   - Track whether selection is complete
type Service interface {
	Get(ctx context.Context) (*ProductionPhaseState, error)
	IsSelectionComplete(ctx context.Context) (bool, error)
	GetAvailableCards(ctx context.Context) ([]string, error)

	SetAvailableCards(ctx context.Context, cardIDs []string) error
	MarkSelectionComplete(ctx context.Context) error
	ClearState(ctx context.Context) error
}

// ServiceImpl implements the production phase state service
type ServiceImpl struct {
	repo Repository
}

// NewService creates a new production phase state service
func NewService(repo Repository) Service {
	return &ServiceImpl{
		repo: repo,
	}
}

// Get retrieves current production phase state
func (s *ServiceImpl) Get(ctx context.Context) (*ProductionPhaseState, error) {
	return s.repo.Get(ctx)
}

// IsSelectionComplete checks if selection is complete
func (s *ServiceImpl) IsSelectionComplete(ctx context.Context) (bool, error) {
	return s.repo.IsSelectionComplete(ctx)
}

// GetAvailableCards retrieves available cards
func (s *ServiceImpl) GetAvailableCards(ctx context.Context) ([]string, error) {
	return s.repo.GetAvailableCards(ctx)
}

// SetAvailableCards sets cards available for selection
func (s *ServiceImpl) SetAvailableCards(ctx context.Context, cardIDs []string) error {
	if err := s.repo.SetAvailableCards(ctx, cardIDs); err != nil {
		return fmt.Errorf("failed to set available cards: %w", err)
	}
	return nil
}

// MarkSelectionComplete marks selection as complete
func (s *ServiceImpl) MarkSelectionComplete(ctx context.Context) error {
	if err := s.repo.MarkSelectionComplete(ctx); err != nil {
		return fmt.Errorf("failed to mark selection complete: %w", err)
	}

	// TODO Phase 6: Publish ProductionPhaseSelectionCompleteEvent

	return nil
}

// ClearState clears production phase state
func (s *ServiceImpl) ClearState(ctx context.Context) error {
	if err := s.repo.ClearState(ctx); err != nil {
		return fmt.Errorf("failed to clear state: %w", err)
	}
	return nil
}
