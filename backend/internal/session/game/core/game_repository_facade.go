package core

import (
	"context"
	"fmt"
)

// This file contains all delegation methods for the GameRepository facade
// DEPRECATED: This facade pattern is being removed. Sub-repositories are now game-scoped.
// Use repositories through Session.Game instead:
//   - sess.Game.Core
//   - sess.Game.GlobalParams
//   - sess.Game.Turn

// ============================================================================
// CORE OPERATIONS (delegated to core repository)
// ============================================================================

// Create creates a new game
// DEPRECATED: Sub-repositories are now game-scoped and don't support this operation
func (r *RepositoryImpl) Create(ctx context.Context, game *Game) error {
	return fmt.Errorf("DEPRECATED: Create not supported on facade. Sub-repositories are game-scoped")
}

// GetByID retrieves a game by ID
// DEPRECATED: Sub-repositories are now game-scoped and don't support this operation
func (r *RepositoryImpl) GetByID(ctx context.Context, gameID string) (*Game, error) {
	return nil, fmt.Errorf("DEPRECATED: GetByID not supported on facade. Sub-repositories are game-scoped")
}

// List retrieves all games, optionally filtered by status
// DEPRECATED: Sub-repositories are now game-scoped and don't support this operation
func (r *RepositoryImpl) List(ctx context.Context, status string) ([]*Game, error) {
	return nil, fmt.Errorf("DEPRECATED: List not supported on facade. Sub-repositories are game-scoped")
}

// UpdateStatus updates game status and publishes event
// DEPRECATED: Sub-repositories are now game-scoped. gameID parameter ignored - uses bound gameID
func (r *RepositoryImpl) UpdateStatus(ctx context.Context, gameID string, status GameStatus) error {
	return r.core.UpdateStatus(ctx, status)
}

// UpdatePhase updates game phase and publishes event
// DEPRECATED: Sub-repositories are now game-scoped. gameID parameter ignored - uses bound gameID
func (r *RepositoryImpl) UpdatePhase(ctx context.Context, gameID string, phase GamePhase) error {
	return r.core.UpdatePhase(ctx, phase)
}

// ============================================================================
// TURN OPERATIONS (delegated to turn repository)
// ============================================================================

// SetHostPlayer sets the host player for a game
// DEPRECATED: Sub-repositories are now game-scoped. gameID parameter ignored - uses bound gameID
func (r *RepositoryImpl) SetHostPlayer(ctx context.Context, gameID string, playerID string) error {
	return r.turn.SetHostPlayer(ctx, playerID)
}

// SetCurrentTurn sets the current turn player
// DEPRECATED: Sub-repositories are now game-scoped. gameID parameter ignored - uses bound gameID
func (r *RepositoryImpl) SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	return r.turn.SetCurrentTurn(ctx, playerID)
}

// AddPlayer adds a player to a game
// DEPRECATED: Sub-repositories are now game-scoped. gameID parameter ignored - uses bound gameID
func (r *RepositoryImpl) AddPlayer(ctx context.Context, gameID string, playerID string) error {
	return r.turn.AddPlayer(ctx, playerID)
}

// ============================================================================
// GLOBAL PARAMETERS OPERATIONS (delegated to global parameters repository)
// ============================================================================

// UpdateTemperature updates the game temperature
// DEPRECATED: Sub-repositories are now game-scoped. gameID parameter ignored - uses bound gameID
func (r *RepositoryImpl) UpdateTemperature(ctx context.Context, gameID string, temperature int) error {
	return r.globalParams.UpdateTemperature(ctx, temperature)
}

// UpdateOxygen updates the game oxygen level
// DEPRECATED: Sub-repositories are now game-scoped. gameID parameter ignored - uses bound gameID
func (r *RepositoryImpl) UpdateOxygen(ctx context.Context, gameID string, oxygen int) error {
	return r.globalParams.UpdateOxygen(ctx, oxygen)
}

// UpdateOceans updates the game ocean count
// DEPRECATED: Sub-repositories are now game-scoped. gameID parameter ignored - uses bound gameID
func (r *RepositoryImpl) UpdateOceans(ctx context.Context, gameID string, oceans int) error {
	return r.globalParams.UpdateOceans(ctx, oceans)
}

// UpdateGeneration updates the game generation counter
// DEPRECATED: Sub-repositories are now game-scoped. gameID parameter ignored - uses bound gameID
func (r *RepositoryImpl) UpdateGeneration(ctx context.Context, gameID string, generation int) error {
	return r.globalParams.UpdateGeneration(ctx, generation)
}
