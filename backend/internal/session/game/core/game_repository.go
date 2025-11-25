package core

import (
	"context"

	"terraforming-mars-backend/internal/events"
)

// Repository manages game data with event-driven updates
type Repository interface {
	// Create creates a new game
	Create(ctx context.Context, game *Game) error

	// GetByID retrieves a game by ID
	GetByID(ctx context.Context, gameID string) (*Game, error)

	// List retrieves all games, optionally filtered by status
	List(ctx context.Context, status string) ([]*Game, error)

	// AddPlayer adds a player to a game (event-driven)
	AddPlayer(ctx context.Context, gameID string, playerID string) error

	// UpdateStatus updates game status (event-driven)
	UpdateStatus(ctx context.Context, gameID string, status GameStatus) error

	// UpdatePhase updates game phase (event-driven)
	UpdatePhase(ctx context.Context, gameID string, phase GamePhase) error

	// SetHostPlayer sets the host player for a game
	SetHostPlayer(ctx context.Context, gameID string, playerID string) error

	// SetCurrentTurn sets the current turn player
	SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error

	// UpdateTemperature updates the game temperature
	UpdateTemperature(ctx context.Context, gameID string, temperature int) error

	// UpdateOxygen updates the game oxygen level
	UpdateOxygen(ctx context.Context, gameID string, oxygen int) error

	// UpdateOceans updates the game ocean count
	UpdateOceans(ctx context.Context, gameID string, oceans int) error

	// UpdateGeneration updates the game generation counter
	UpdateGeneration(ctx context.Context, gameID string, generation int) error
}

// RepositoryImpl implements the Repository interface as a facade
// It delegates operations to specialized sub-repositories
type RepositoryImpl struct {
	core         *GameCoreRepository
	turn         *GameTurnRepository
	globalParams *GameGlobalParametersRepository
}

// NewRepository creates a new game repository with all sub-repositories
// DEPRECATED: This creates repositories without a game ID binding. Use game-scoped repositories via Session instead.
// Temporarily creates repositories with empty gameID for backwards compatibility.
func NewRepository(eventBus *events.EventBusImpl) Repository {
	storage := NewGameStorage()

	// HACK: Using empty gameID for backwards compatibility - this facade will be removed
	coreRepo := NewGameCoreRepository("", storage, eventBus)
	turnRepo := NewGameTurnRepository("", storage, eventBus)
	globalParamsRepo := NewGameGlobalParametersRepository("", storage, eventBus)

	return &RepositoryImpl{
		core:         coreRepo,
		turn:         turnRepo,
		globalParams: globalParamsRepo,
	}
}

// GetCore returns the core repository for direct access
func (r *RepositoryImpl) GetCore() *GameCoreRepository {
	return r.core
}

// GetTurn returns the turn repository for direct access
func (r *RepositoryImpl) GetTurn() *GameTurnRepository {
	return r.turn
}

// GetGlobalParams returns the global parameters repository for direct access
func (r *RepositoryImpl) GetGlobalParams() *GameGlobalParametersRepository {
	return r.globalParams
}
