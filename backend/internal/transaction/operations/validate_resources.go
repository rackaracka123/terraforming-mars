package operations

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// ValidateResourcesOperation validates resources without deducting them
type ValidateResourcesOperation struct {
	playerRepo repository.PlayerRepository
	gameID     string
	playerID   string
	cost       model.Resources
}

// NewValidateResourcesOperation creates a new validate resources operation
func NewValidateResourcesOperation(playerRepo repository.PlayerRepository, gameID, playerID string, cost model.Resources) *ValidateResourcesOperation {
	return &ValidateResourcesOperation{
		playerRepo: playerRepo,
		gameID:     gameID,
		playerID:   playerID,
		cost:       cost,
	}
}

func (op *ValidateResourcesOperation) Execute(ctx context.Context) error {
	// Get current player state
	player, err := op.playerRepo.GetByID(ctx, op.gameID, op.playerID)
	if err != nil {
		return fmt.Errorf("failed to get player state: %w", err)
	}

	// Validate sufficient resources (no deduction)
	if player.Resources.Credits < op.cost.Credits {
		return fmt.Errorf("insufficient credits: need %d, have %d", op.cost.Credits, player.Resources.Credits)
	}
	if player.Resources.Steel < op.cost.Steel {
		return fmt.Errorf("insufficient steel: need %d, have %d", op.cost.Steel, player.Resources.Steel)
	}
	if player.Resources.Titanium < op.cost.Titanium {
		return fmt.Errorf("insufficient titanium: need %d, have %d", op.cost.Titanium, player.Resources.Titanium)
	}
	if player.Resources.Plants < op.cost.Plants {
		return fmt.Errorf("insufficient plants: need %d, have %d", op.cost.Plants, player.Resources.Plants)
	}
	if player.Resources.Energy < op.cost.Energy {
		return fmt.Errorf("insufficient energy: need %d, have %d", op.cost.Energy, player.Resources.Energy)
	}
	if player.Resources.Heat < op.cost.Heat {
		return fmt.Errorf("insufficient heat: need %d, have %d", op.cost.Heat, player.Resources.Heat)
	}

	return nil
}

func (op *ValidateResourcesOperation) Rollback(ctx context.Context) error {
	// Resource validation doesn't modify state, so no rollback needed
	return nil
}

func (op *ValidateResourcesOperation) String() string {
	return fmt.Sprintf("ValidateResources(gameID=%s, playerID=%s, cost=%+v)", op.gameID, op.playerID, op.cost)
}
