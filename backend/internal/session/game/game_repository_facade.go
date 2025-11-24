package game

import (
	"context"
)

// This file contains all delegation methods for the GameRepository facade
// Each method delegates to the appropriate sub-repository

// ============================================================================
// CORE OPERATIONS (delegated to core repository)
// ============================================================================

// Create creates a new game
func (r *RepositoryImpl) Create(ctx context.Context, game *Game) error {
	return r.core.Create(ctx, game)
}

// GetByID retrieves a game by ID
func (r *RepositoryImpl) GetByID(ctx context.Context, gameID string) (*Game, error) {
	return r.core.GetByID(ctx, gameID)
}

// List retrieves all games, optionally filtered by status
func (r *RepositoryImpl) List(ctx context.Context, status string) ([]*Game, error) {
	return r.core.List(ctx, status)
}

// UpdateStatus updates game status and publishes event
func (r *RepositoryImpl) UpdateStatus(ctx context.Context, gameID string, status GameStatus) error {
	return r.core.UpdateStatus(ctx, gameID, status)
}

// UpdatePhase updates game phase and publishes event
func (r *RepositoryImpl) UpdatePhase(ctx context.Context, gameID string, phase GamePhase) error {
	return r.core.UpdatePhase(ctx, gameID, phase)
}

// ============================================================================
// TURN OPERATIONS (delegated to turn repository)
// ============================================================================

// SetHostPlayer sets the host player for a game
func (r *RepositoryImpl) SetHostPlayer(ctx context.Context, gameID string, playerID string) error {
	return r.turn.SetHostPlayer(ctx, gameID, playerID)
}

// SetCurrentTurn sets the current turn player
func (r *RepositoryImpl) SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error {
	return r.turn.SetCurrentTurn(ctx, gameID, playerID)
}

// AddPlayer adds a player to a game
func (r *RepositoryImpl) AddPlayer(ctx context.Context, gameID string, playerID string) error {
	return r.turn.AddPlayer(ctx, gameID, playerID)
}

// ============================================================================
// GLOBAL PARAMETERS OPERATIONS (delegated to global parameters repository)
// ============================================================================

// UpdateTemperature updates the game temperature
func (r *RepositoryImpl) UpdateTemperature(ctx context.Context, gameID string, temperature int) error {
	return r.globalParams.UpdateTemperature(ctx, gameID, temperature)
}

// UpdateOxygen updates the game oxygen level
func (r *RepositoryImpl) UpdateOxygen(ctx context.Context, gameID string, oxygen int) error {
	return r.globalParams.UpdateOxygen(ctx, gameID, oxygen)
}

// UpdateOceans updates the game ocean count
func (r *RepositoryImpl) UpdateOceans(ctx context.Context, gameID string, oceans int) error {
	return r.globalParams.UpdateOceans(ctx, gameID, oceans)
}

// UpdateGeneration updates the game generation counter
func (r *RepositoryImpl) UpdateGeneration(ctx context.Context, gameID string, generation int) error {
	return r.globalParams.UpdateGeneration(ctx, gameID, generation)
}
