package operations

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// DeductResourcesOperation deducts resources from a player
type DeductResourcesOperation struct {
	playerRepo    repository.PlayerRepository
	gameID        string
	playerID      string
	cost          model.Resources
	originalState *model.Resources
}

// NewDeductResourcesOperation creates a new deduct resources operation
func NewDeductResourcesOperation(playerRepo repository.PlayerRepository, gameID, playerID string, cost model.Resources) *DeductResourcesOperation {
	return &DeductResourcesOperation{
		playerRepo:    playerRepo,
		gameID:        gameID,
		playerID:      playerID,
		cost:          cost,
		originalState: nil, // Will be populated during execution
	}
}

func (op *DeductResourcesOperation) Execute(ctx context.Context) error {
	// Get current player state
	player, err := op.playerRepo.GetByID(ctx, op.gameID, op.playerID)
	if err != nil {
		return fmt.Errorf("failed to get player state: %w", err)
	}

	// Store original state for rollback
	originalResources := player.Resources.DeepCopy()
	op.originalState = &originalResources

	// Validate sufficient resources
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

	// Calculate new resource values
	newResources := model.Resources{
		Credits:  player.Resources.Credits - op.cost.Credits,
		Steel:    player.Resources.Steel - op.cost.Steel,
		Titanium: player.Resources.Titanium - op.cost.Titanium,
		Plants:   player.Resources.Plants - op.cost.Plants,
		Energy:   player.Resources.Energy - op.cost.Energy,
		Heat:     player.Resources.Heat - op.cost.Heat,
	}

	// Update player resources
	return op.playerRepo.UpdateResources(ctx, op.gameID, op.playerID, newResources)
}

func (op *DeductResourcesOperation) Rollback(ctx context.Context) error {
	if op.originalState == nil {
		return nil // Nothing to rollback
	}

	// Restore original resource state
	return op.playerRepo.UpdateResources(ctx, op.gameID, op.playerID, *op.originalState)
}

func (op *DeductResourcesOperation) String() string {
	return fmt.Sprintf("DeductResources(gameID=%s, playerID=%s, cost=%+v)", op.gameID, op.playerID, op.cost)
}
