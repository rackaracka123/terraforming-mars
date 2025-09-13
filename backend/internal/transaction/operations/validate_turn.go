package operations

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// ValidateTurnOperation validates that a player can take a turn
type ValidateTurnOperation struct {
	gameRepo   repository.GameRepository
	playerRepo repository.PlayerRepository
	gameID     string
	playerID   string
}

// NewValidateTurnOperation creates a new validate turn operation
func NewValidateTurnOperation(gameRepo repository.GameRepository, playerRepo repository.PlayerRepository, gameID, playerID string) *ValidateTurnOperation {
	return &ValidateTurnOperation{
		gameRepo:   gameRepo,
		playerRepo: playerRepo,
		gameID:     gameID,
		playerID:   playerID,
	}
}

func (op *ValidateTurnOperation) Execute(ctx context.Context) error {
	// Get game state
	game, err := op.gameRepo.GetByID(ctx, op.gameID)
	if err != nil {
		return fmt.Errorf("failed to get game state: %w", err)
	}

	// Check if game is in action phase
	if game.CurrentPhase != model.GamePhaseAction {
		return fmt.Errorf("actions not allowed in phase %s", game.CurrentPhase)
	}

	// Check if it's the player's turn
	if game.CurrentTurn == nil || *game.CurrentTurn != op.playerID {
		currentPlayer := "none"
		if game.CurrentTurn != nil {
			currentPlayer = *game.CurrentTurn
		}
		return fmt.Errorf("not your turn: current turn player is %s", currentPlayer)
	}

	// Get player state
	player, err := op.playerRepo.GetByID(ctx, op.gameID, op.playerID)
	if err != nil {
		return fmt.Errorf("failed to get player state: %w", err)
	}

	// Check if player has actions remaining
	if player.AvailableActions <= 0 {
		return fmt.Errorf("no actions remaining for this turn")
	}

	return nil
}

func (op *ValidateTurnOperation) Rollback(ctx context.Context) error {
	// Turn validation doesn't modify state, so no rollback needed
	return nil
}

func (op *ValidateTurnOperation) String() string {
	return fmt.Sprintf("ValidateTurn(gameID=%s, playerID=%s)", op.gameID, op.playerID)
}
