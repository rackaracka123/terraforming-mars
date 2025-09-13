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
func NewConsumeActionOperation(playerRepo repository.PlayerRepository, gameID, playerID string) *ConsumeActionOperation {
	return &ConsumeActionOperation{
		playerRepo:      playerRepo,
		gameID:          gameID,
		playerID:        playerID,
		originalActions: 0, // Will be populated during execution
	}
}

func (op *ConsumeActionOperation) Execute(ctx context.Context) error {
	// Get current player state
	player, err := op.playerRepo.GetByID(ctx, op.gameID, op.playerID)
	if err != nil {
		return fmt.Errorf("failed to get player state: %w", err)
	}

	// Store original actions for rollback
	op.originalActions = player.AvailableActions

	// Validate player has actions available
	if player.AvailableActions <= 0 {
		return fmt.Errorf("no actions remaining to consume")
	}

	// Consume one action
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
