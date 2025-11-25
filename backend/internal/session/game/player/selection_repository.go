package player

import (
	"context"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/types"
)

var logSelection = logger.Get()

// SelectionRepository handles card selection phases and pending selections for a specific player
// Auto-saves changes after every operation
type SelectionRepository struct {
	player *Player // Reference to parent player
}

// NewSelectionRepository creates a new selection repository for a specific player
func NewSelectionRepository(player *Player) *SelectionRepository {
	return &SelectionRepository{
		player: player,
	}
}

// SetStartingCardsSelection sets the starting cards selection phase
// Auto-saves changes to the player
func (r *SelectionRepository) SetStartingCardsSelection(ctx context.Context, cardIDs []string, corpIDs []string) error {
	r.player.SelectStartingCardsPhase = &types.SelectStartingCardsPhase{
		AvailableCards:        cardIDs,
		AvailableCorporations: corpIDs,
	}
	return nil
}

// CompleteStartingSelection marks the starting selection as complete and clears the phase
// Auto-saves changes to the player
func (r *SelectionRepository) CompleteStartingSelection(ctx context.Context) error {
	// Clear the phase entirely - selection is complete and modal should close
	r.player.SelectStartingCardsPhase = nil
	return nil
}

// CompleteProductionSelection marks the production selection as complete
// Auto-saves changes to the player
func (r *SelectionRepository) CompleteProductionSelection(ctx context.Context) error {
	if r.player.ProductionPhase != nil {
		r.player.ProductionPhase.SelectionComplete = true
	}
	return nil
}

// UpdateStartingCardsPhase updates the starting cards selection phase
// Auto-saves changes to the player
func (r *SelectionRepository) UpdateStartingCardsPhase(ctx context.Context, phase *types.SelectStartingCardsPhase) error {
	r.player.SelectStartingCardsPhase = phase
	return nil
}

// UpdateProductionPhase updates the production phase
// Auto-saves changes to the player
func (r *SelectionRepository) UpdateProductionPhase(ctx context.Context, phase *types.ProductionPhase) error {
	r.player.ProductionPhase = phase
	return nil
}

// UpdatePendingCardSelection updates player pending card selection
// Auto-saves changes to the player
func (r *SelectionRepository) UpdatePendingCardSelection(ctx context.Context, selection *types.PendingCardSelection) error {
	r.player.PendingCardSelection = selection

	logSelection.Debug("✅ Player pending card selection updated",
		zap.String("game_id", r.player.GameID),
		zap.String("player_id", r.player.ID),
		zap.String("source", selection.Source),
		zap.Int("available_cards", len(selection.AvailableCards)))

	return nil
}

// ClearPendingCardSelection clears the pending card selection
// Auto-saves changes to the player
func (r *SelectionRepository) ClearPendingCardSelection(ctx context.Context) error {
	r.player.PendingCardSelection = nil

	logSelection.Debug("✅ Player pending card selection cleared",
		zap.String("game_id", r.player.GameID),
		zap.String("player_id", r.player.ID))

	return nil
}

// UpdatePendingCardDrawSelection updates player pending card draw selection
// Auto-saves changes to the player
func (r *SelectionRepository) UpdatePendingCardDrawSelection(ctx context.Context, selection *types.PendingCardDrawSelection) error {
	r.player.PendingCardDrawSelection = selection
	return nil
}
