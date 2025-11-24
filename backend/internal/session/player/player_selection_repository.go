package player

import (
	"context"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session/types"
)

var logSelection = logger.Get()

// PlayerSelectionRepository handles player card selection phases and pending selections
type PlayerSelectionRepository struct {
	storage *PlayerStorage
}

// NewPlayerSelectionRepository creates a new player selection repository
func NewPlayerSelectionRepository(storage *PlayerStorage) *PlayerSelectionRepository {
	return &PlayerSelectionRepository{
		storage: storage,
	}
}

// SetStartingCardsSelection sets the starting cards selection phase for a player
func (r *PlayerSelectionRepository) SetStartingCardsSelection(ctx context.Context, gameID string, playerID string, cardIDs []string, corpIDs []string) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.SelectStartingCardsPhase = &SelectStartingCardsPhase{
		AvailableCards:        cardIDs,
		AvailableCorporations: corpIDs,
	}

	return r.storage.Set(gameID, playerID, p)
}

// CompleteStartingSelection marks the starting selection as complete and clears the phase
func (r *PlayerSelectionRepository) CompleteStartingSelection(ctx context.Context, gameID string, playerID string) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	// Clear the phase entirely - selection is complete and modal should close
	p.SelectStartingCardsPhase = nil

	return r.storage.Set(gameID, playerID, p)
}

// CompleteProductionSelection marks the production selection as complete
func (r *PlayerSelectionRepository) CompleteProductionSelection(ctx context.Context, gameID string, playerID string) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	if p.ProductionPhase != nil {
		p.ProductionPhase.SelectionComplete = true
	}

	return r.storage.Set(gameID, playerID, p)
}

// UpdateSelectStartingCardsPhase updates the starting cards selection phase
func (r *PlayerSelectionRepository) UpdateSelectStartingCardsPhase(ctx context.Context, gameID string, playerID string, phase *SelectStartingCardsPhase) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.SelectStartingCardsPhase = phase

	return r.storage.Set(gameID, playerID, p)
}

// UpdateProductionPhase updates the production phase
func (r *PlayerSelectionRepository) UpdateProductionPhase(ctx context.Context, gameID string, playerID string, phase *types.ProductionPhase) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.ProductionPhase = phase

	return r.storage.Set(gameID, playerID, p)
}

// UpdatePendingCardSelection updates player pending card selection
func (r *PlayerSelectionRepository) UpdatePendingCardSelection(ctx context.Context, gameID string, playerID string, selection *PendingCardSelection) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.PendingCardSelection = selection

	logSelection.Debug("✅ Player pending card selection updated",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("source", selection.Source),
		zap.Int("available_cards", len(selection.AvailableCards)))

	return r.storage.Set(gameID, playerID, p)
}

// ClearPendingCardSelection clears the pending card selection
func (r *PlayerSelectionRepository) ClearPendingCardSelection(ctx context.Context, gameID string, playerID string) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.PendingCardSelection = nil

	logSelection.Debug("✅ Player pending card selection cleared",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID))

	return r.storage.Set(gameID, playerID, p)
}

// UpdatePendingCardDrawSelection updates player pending card draw selection
func (r *PlayerSelectionRepository) UpdatePendingCardDrawSelection(ctx context.Context, gameID string, playerID string, selection *types.PendingCardDrawSelection) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	p.PendingCardDrawSelection = selection

	return r.storage.Set(gameID, playerID, p)
}
