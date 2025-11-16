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

// TurnOrderRepository manages game-level turn order state
type TurnOrderRepository interface {
	// Get turn state
	GetCurrentTurn(ctx context.Context) (*string, error)
	GetPlayerOrder(ctx context.Context) ([]string, error)

	// Granular operations
	SetCurrentTurn(ctx context.Context, playerID *string) error
	AdvanceTurn(ctx context.Context) (*string, error) // Returns next player ID
	SetPlayerOrder(ctx context.Context, playerIDs []string) error
}

// TurnOrderRepositoryImpl implements independent in-memory storage for turn order
type TurnOrderRepositoryImpl struct {
	mu          sync.RWMutex
	playerIDs   []string
	currentTurn *string
}

// NewTurnOrderRepository creates a new independent turn order repository with initial state
func NewTurnOrderRepository(playerIDs []string, currentTurn *string) TurnOrderRepository {
	return &TurnOrderRepositoryImpl{
		playerIDs:   append([]string{}, playerIDs...),
		currentTurn: currentTurn,
	}
}

// GetCurrentTurn retrieves the current player's turn
func (r *TurnOrderRepositoryImpl) GetCurrentTurn(ctx context.Context) (*string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.currentTurn, nil
}

// GetPlayerOrder retrieves the order of players
func (r *TurnOrderRepositoryImpl) GetPlayerOrder(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy
	playersCopy := append([]string{}, r.playerIDs...)
	return playersCopy, nil
}

// SetCurrentTurn sets the current player's turn
func (r *TurnOrderRepositoryImpl) SetCurrentTurn(ctx context.Context, playerID *string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.currentTurn = playerID
	return nil
}

// AdvanceTurn moves to the next player in order
// Returns the new current player ID
func (r *TurnOrderRepositoryImpl) AdvanceTurn(ctx context.Context) (*string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.playerIDs) == 0 {
		return nil, fmt.Errorf("no players in turn order")
	}

	// Find current player index
	currentIndex := -1
	if r.currentTurn != nil {
		for i, pid := range r.playerIDs {
			if pid == *r.currentTurn {
				currentIndex = i
				break
			}
		}
	}

	// Move to next player (wrap around)
	nextIndex := (currentIndex + 1) % len(r.playerIDs)
	nextPlayerID := r.playerIDs[nextIndex]
	r.currentTurn = &nextPlayerID

	return r.currentTurn, nil
}

// SetPlayerOrder sets the order of players for turn rotation
func (r *TurnOrderRepositoryImpl) SetPlayerOrder(ctx context.Context, playerIDs []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.playerIDs = append([]string{}, playerIDs...)
	return nil
}

// ============================================================================
// SERVICE
// ============================================================================

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
