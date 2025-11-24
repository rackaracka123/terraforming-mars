package player

import (
	"context"

	"terraforming-mars-backend/internal/session/types"
)

// This file contains all delegation methods for the PlayerRepository facade
// Each method delegates to the appropriate sub-repository

// ============================================================================
// RESOURCE OPERATIONS (delegated to resource repository)
// ============================================================================

// UpdateResources updates player resources and publishes batched event
func (r *RepositoryImpl) UpdateResources(ctx context.Context, gameID string, playerID string, resources types.Resources) error {
	return r.resource.UpdateResources(ctx, gameID, playerID, resources)
}

// UpdateProduction updates player production
func (r *RepositoryImpl) UpdateProduction(ctx context.Context, gameID string, playerID string, production types.Production) error {
	return r.resource.UpdateProduction(ctx, gameID, playerID, production)
}

// UpdateTerraformRating updates player terraform rating
func (r *RepositoryImpl) UpdateTerraformRating(ctx context.Context, gameID string, playerID string, rating int) error {
	return r.resource.UpdateTerraformRating(ctx, gameID, playerID, rating)
}

// UpdateVictoryPoints updates player victory points
func (r *RepositoryImpl) UpdateVictoryPoints(ctx context.Context, gameID string, playerID string, victoryPoints int) error {
	return r.resource.UpdateVictoryPoints(ctx, gameID, playerID, victoryPoints)
}

// UpdateResourceStorage updates player resource storage
func (r *RepositoryImpl) UpdateResourceStorage(ctx context.Context, gameID string, playerID string, storage map[string]int) error {
	return r.resource.UpdateResourceStorage(ctx, gameID, playerID, storage)
}

// UpdatePaymentSubstitutes updates player payment substitutes
func (r *RepositoryImpl) UpdatePaymentSubstitutes(ctx context.Context, gameID string, playerID string, substitutes []types.PaymentSubstitute) error {
	return r.resource.UpdatePaymentSubstitutes(ctx, gameID, playerID, substitutes)
}

// ============================================================================
// HAND OPERATIONS (delegated to hand repository)
// ============================================================================

// AddCard adds a card to player's hand
func (r *RepositoryImpl) AddCard(ctx context.Context, gameID string, playerID string, cardID string) error {
	return r.hand.AddCard(ctx, gameID, playerID, cardID)
}

// RemoveCardFromHand removes a card from the player's hand
func (r *RepositoryImpl) RemoveCardFromHand(ctx context.Context, gameID string, playerID string, cardID string) error {
	return r.hand.RemoveCardFromHand(ctx, gameID, playerID, cardID)
}

// ============================================================================
// SELECTION OPERATIONS (delegated to selection repository)
// ============================================================================

// SetStartingCardsSelection sets the starting cards selection phase for a player
func (r *RepositoryImpl) SetStartingCardsSelection(ctx context.Context, gameID string, playerID string, cardIDs []string, corpIDs []string) error {
	return r.selection.SetStartingCardsSelection(ctx, gameID, playerID, cardIDs, corpIDs)
}

// CompleteStartingSelection marks the starting selection as complete
func (r *RepositoryImpl) CompleteStartingSelection(ctx context.Context, gameID string, playerID string) error {
	return r.selection.CompleteStartingSelection(ctx, gameID, playerID)
}

// CompleteProductionSelection marks the production selection as complete
func (r *RepositoryImpl) CompleteProductionSelection(ctx context.Context, gameID string, playerID string) error {
	return r.selection.CompleteProductionSelection(ctx, gameID, playerID)
}

// UpdateSelectStartingCardsPhase updates the starting cards selection phase
func (r *RepositoryImpl) UpdateSelectStartingCardsPhase(ctx context.Context, gameID string, playerID string, phase *SelectStartingCardsPhase) error {
	return r.selection.UpdateSelectStartingCardsPhase(ctx, gameID, playerID, phase)
}

// UpdateProductionPhase updates the production phase
func (r *RepositoryImpl) UpdateProductionPhase(ctx context.Context, gameID string, playerID string, phase *types.ProductionPhase) error {
	return r.selection.UpdateProductionPhase(ctx, gameID, playerID, phase)
}

// UpdatePendingCardSelection updates player pending card selection
func (r *RepositoryImpl) UpdatePendingCardSelection(ctx context.Context, gameID string, playerID string, selection *PendingCardSelection) error {
	return r.selection.UpdatePendingCardSelection(ctx, gameID, playerID, selection)
}

// ClearPendingCardSelection clears the pending card selection
func (r *RepositoryImpl) ClearPendingCardSelection(ctx context.Context, gameID string, playerID string) error {
	return r.selection.ClearPendingCardSelection(ctx, gameID, playerID)
}

// UpdatePendingCardDrawSelection updates player pending card draw selection
func (r *RepositoryImpl) UpdatePendingCardDrawSelection(ctx context.Context, gameID string, playerID string, selection *types.PendingCardDrawSelection) error {
	return r.selection.UpdatePendingCardDrawSelection(ctx, gameID, playerID, selection)
}

// ============================================================================
// CORPORATION OPERATIONS (delegated to corporation repository)
// ============================================================================

// SetCorporation sets the player's corporation
func (r *RepositoryImpl) SetCorporation(ctx context.Context, gameID string, playerID string, corporationID string) error {
	return r.corporation.SetCorporation(ctx, gameID, playerID, corporationID)
}

// UpdateCorporation updates the player's corporation with full card data
func (r *RepositoryImpl) UpdateCorporation(ctx context.Context, gameID string, playerID string, corporation types.Card) error {
	return r.corporation.UpdateCorporation(ctx, gameID, playerID, corporation)
}

// ============================================================================
// TURN OPERATIONS (delegated to turn repository)
// ============================================================================

// UpdateConnectionStatus updates player connection status
func (r *RepositoryImpl) UpdateConnectionStatus(ctx context.Context, gameID string, playerID string, isConnected bool) error {
	return r.turn.UpdateConnectionStatus(ctx, gameID, playerID, isConnected)
}

// UpdatePassed updates player passed status for generation
func (r *RepositoryImpl) UpdatePassed(ctx context.Context, gameID string, playerID string, passed bool) error {
	return r.turn.UpdatePassed(ctx, gameID, playerID, passed)
}

// UpdateAvailableActions updates player available actions count
func (r *RepositoryImpl) UpdateAvailableActions(ctx context.Context, gameID string, playerID string, actions int) error {
	return r.turn.UpdateAvailableActions(ctx, gameID, playerID, actions)
}

// UpdatePlayerActions updates player actions
func (r *RepositoryImpl) UpdatePlayerActions(ctx context.Context, gameID string, playerID string, actions []types.PlayerAction) error {
	return r.turn.UpdatePlayerActions(ctx, gameID, playerID, actions)
}

// UpdateForcedFirstAction updates player forced first action
func (r *RepositoryImpl) UpdateForcedFirstAction(ctx context.Context, gameID string, playerID string, action *types.ForcedFirstAction) error {
	return r.turn.UpdateForcedFirstAction(ctx, gameID, playerID, action)
}

// ============================================================================
// EFFECT OPERATIONS (delegated to effect repository)
// ============================================================================

// UpdateRequirementModifiers updates player requirement modifiers
func (r *RepositoryImpl) UpdateRequirementModifiers(ctx context.Context, gameID string, playerID string, modifiers []types.RequirementModifier) error {
	return r.effect.UpdateRequirementModifiers(ctx, gameID, playerID, modifiers)
}

// UpdatePlayerEffects updates player active effects
func (r *RepositoryImpl) UpdatePlayerEffects(ctx context.Context, gameID string, playerID string, effects []types.PlayerEffect) error {
	return r.effect.UpdatePlayerEffects(ctx, gameID, playerID, effects)
}

// ============================================================================
// TILE QUEUE OPERATIONS (delegated to tile queue repository)
// ============================================================================

// CreateTileQueue creates a tile placement queue for the player
func (r *RepositoryImpl) CreateTileQueue(ctx context.Context, gameID string, playerID string, cardID string, tileTypes []string) error {
	return r.tileQueue.CreateTileQueue(ctx, gameID, playerID, cardID, tileTypes)
}

// GetPendingTileSelectionQueue retrieves the pending tile selection queue for a player
func (r *RepositoryImpl) GetPendingTileSelectionQueue(ctx context.Context, gameID string, playerID string) (*types.PendingTileSelectionQueue, error) {
	return r.tileQueue.GetPendingTileSelectionQueue(ctx, gameID, playerID)
}

// ProcessNextTileInQueue pops the next tile type from the queue and returns it
func (r *RepositoryImpl) ProcessNextTileInQueue(ctx context.Context, gameID string, playerID string) (string, error) {
	return r.tileQueue.ProcessNextTileInQueue(ctx, gameID, playerID)
}

// UpdatePendingTileSelection updates the pending tile selection for a player
func (r *RepositoryImpl) UpdatePendingTileSelection(ctx context.Context, gameID string, playerID string, selection *types.PendingTileSelection) error {
	return r.tileQueue.UpdatePendingTileSelection(ctx, gameID, playerID, selection)
}

// ClearPendingTileSelection clears the pending tile selection for a player
func (r *RepositoryImpl) ClearPendingTileSelection(ctx context.Context, gameID string, playerID string) error {
	return r.tileQueue.ClearPendingTileSelection(ctx, gameID, playerID)
}

// UpdatePendingTileSelectionQueue updates the pending tile selection queue
func (r *RepositoryImpl) UpdatePendingTileSelectionQueue(ctx context.Context, gameID string, playerID string, queue *types.PendingTileSelectionQueue) error {
	return r.tileQueue.UpdatePendingTileSelectionQueue(ctx, gameID, playerID, queue)
}

// ClearPendingTileSelectionQueue clears the pending tile selection queue
func (r *RepositoryImpl) ClearPendingTileSelectionQueue(ctx context.Context, gameID string, playerID string) error {
	return r.tileQueue.ClearPendingTileSelectionQueue(ctx, gameID, playerID)
}
