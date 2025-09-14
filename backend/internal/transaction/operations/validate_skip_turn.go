package operations

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// ValidateSkipTurnOperation validates that a player can skip their turn
// Unlike ValidateTurnOperation, this allows skip even when AvailableActions <= 0
type ValidateSkipTurnOperation struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
	gameID     string
	playerID   string
}

// NewValidateSkipTurnOperation creates a new validate skip turn operation
func NewValidateSkipTurnOperation(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, gameID, playerID string) *ValidateSkipTurnOperation {
	return &ValidateSkipTurnOperation{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
		gameID:     gameID,
		playerID:   playerID,
	}
}

func (op *ValidateSkipTurnOperation) Execute(ctx context.Context) error {
	// Get game state
	game, err := op.gameRepo.GetByID(ctx, op.gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	// Check if game is in action phase
	if game.CurrentPhase != model.GamePhaseAction {
		return fmt.Errorf("skip action not allowed in phase %s", game.CurrentPhase)
	}

	// Check if it's the player's turn
	if game.CurrentTurn == nil || *game.CurrentTurn != op.playerID {
		currentPlayer := "none"
		if game.CurrentTurn != nil {
			currentPlayer = *game.CurrentTurn
		}
		return fmt.Errorf("not your turn: current turn player is %s", currentPlayer)
	}

	// NOTE: We DON'T check AvailableActions here - players can skip with 0 actions
	// This is the key difference from ValidateTurnOperation

	return nil
}

func (op *ValidateSkipTurnOperation) Rollback(ctx context.Context) error {
	// Skip turn validation doesn't modify state, so no rollback needed
	return nil
}

func (op *ValidateSkipTurnOperation) String() string {
	return fmt.Sprintf("ValidateSkipTurn(gameID=%s, playerID=%s)", op.gameID, op.playerID)
}