package operations

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/repository"
)

// ConsumeActionOperation consumes one player action
type ConsumeActionOperation struct {
	playerRepo      repository.PlayerRepository
	gameID          string
	playerID        string
	originalActions int
}

// NewConsumeActionOperation creates a new consume action operation
func NewConsumeActionOperation(ctx context.Context, playerRepo repository.PlayerRepository, gameID, playerID string) (*ConsumeActionOperation, error) {
	// Fetch current player state to store original actions for rollback
	player, err := playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get player state: %w", err)
	}

	return &ConsumeActionOperation{
		playerRepo:      playerRepo,
		gameID:          gameID,
		playerID:        playerID,
		originalActions: player.AvailableActions,
	}, nil
}

func (op *ConsumeActionOperation) Execute(ctx context.Context) error {
	// Get current player state for validation and calculation
	player, err := op.playerRepo.GetByID(ctx, op.gameID, op.playerID)
	if err != nil {
		return fmt.Errorf("failed to get current player state: %w", err)
	}

	// Validate player has actions available using current state
	if player.AvailableActions <= 0 {
		return fmt.Errorf("no actions remaining to consume")
	}

	// Consume one action from current state
	newActions := player.AvailableActions - 1
	return op.playerRepo.UpdateAvailableActions(ctx, op.gameID, op.playerID, newActions)
}

func (op *ConsumeActionOperation) Rollback(ctx context.Context) error {
	// Restore original action count
	return op.playerRepo.UpdateAvailableActions(ctx, op.gameID, op.playerID, op.originalActions)
}

func (op *ConsumeActionOperation) String() string {
	return fmt.Sprintf("ConsumeAction(gameID=%s, playerID=%s)", op.gameID, op.playerID)
}
