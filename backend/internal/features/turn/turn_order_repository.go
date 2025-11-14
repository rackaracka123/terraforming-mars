package turn

import (
	"context"
	"fmt"
	"sync"
)

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
