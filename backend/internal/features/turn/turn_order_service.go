package turn

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// TurnOrderService handles game-level turn order
//
// Scope: Isolated turn order management for a game
//   - Current turn tracking
//   - Turn advancement
//   - Player order management
type TurnOrderService interface {
	GetCurrentTurn(ctx context.Context) (*string, error)
	GetPlayerOrder(ctx context.Context) ([]string, error)

	SetCurrentTurn(ctx context.Context, playerID *string) error
	AdvanceTurn(ctx context.Context) (*string, error)
	SetPlayerOrder(ctx context.Context, playerIDs []string) error

}

// TurnOrderServiceImpl implements the turn order service
type TurnOrderServiceImpl struct {
	repo TurnOrderRepository
}

// NewTurnOrderService creates a new turn order service
func NewTurnOrderService(repo TurnOrderRepository) TurnOrderService {
	return &TurnOrderServiceImpl{
		repo: repo,
	}
}

// GetCurrentTurn retrieves the current player's turn
func (s *TurnOrderServiceImpl) GetCurrentTurn(ctx context.Context) (*string, error) {
	return s.repo.GetCurrentTurn(ctx)
}

// GetPlayerOrder retrieves the order of players
func (s *TurnOrderServiceImpl) GetPlayerOrder(ctx context.Context) ([]string, error) {
	return s.repo.GetPlayerOrder(ctx)
}

// SetCurrentTurn sets the current player's turn
func (s *TurnOrderServiceImpl) SetCurrentTurn(ctx context.Context, playerID *string) error {
	if err := s.repo.SetCurrentTurn(ctx, playerID); err != nil {
		return fmt.Errorf("failed to set current turn: %w", err)
	}

	if playerID != nil {
		logger.Get().Debug("Current turn set", zap.String("player_id", *playerID))
	}

	return nil
}

// AdvanceTurn moves to the next player in turn order
func (s *TurnOrderServiceImpl) AdvanceTurn(ctx context.Context) (*string, error) {
	nextPlayerID, err := s.repo.AdvanceTurn(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to advance turn: %w", err)
	}

	if nextPlayerID != nil {
		logger.Get().Info("🔄 Turn advanced", zap.String("next_player_id", *nextPlayerID))
	}

	// TODO Phase 6: Publish TurnAdvancedEvent

	return nextPlayerID, nil
}

// SetPlayerOrder sets the order of players for turn rotation
func (s *TurnOrderServiceImpl) SetPlayerOrder(ctx context.Context, playerIDs []string) error {
	if err := s.repo.SetPlayerOrder(ctx, playerIDs); err != nil {
		return fmt.Errorf("failed to set player order: %w", err)
	}

	logger.Get().Debug("Player order set", zap.Strings("player_ids", playerIDs))

	return nil
}
