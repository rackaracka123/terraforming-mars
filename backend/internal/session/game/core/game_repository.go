package core

import (
	"context"

	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/types"
)

// Repository defines the interface for game data access operations
// This manages the collection of games (CRUD operations at game level)
// Individual game domain logic is in game.Game methods
//
// DEPRECATED: This interface is in the core package for backward compatibility.
// TODO: Move to parent game package once import cycles are fully resolved.
type Repository interface {
	// Create creates a new game in the repository
	Create(ctx context.Context, g *game.Game) error

	// GetByID retrieves a game by its ID
	GetByID(ctx context.Context, gameID string) (*game.Game, error)

	// List retrieves all games with optional status filter
	// Pass empty string for status to retrieve all games
	List(ctx context.Context, status types.GameStatus) ([]*game.Game, error)

	// UpdateTemperature updates the temperature for a game
	UpdateTemperature(ctx context.Context, gameID string, newTemp int) error

	// UpdateOxygen updates the oxygen level for a game
	UpdateOxygen(ctx context.Context, gameID string, newOxygen int) error

	// UpdateOceans updates the ocean count for a game
	UpdateOceans(ctx context.Context, gameID string, newOceans int) error

	// UpdateGeneration advances the game generation
	UpdateGeneration(ctx context.Context, gameID string, newGeneration int) error

	// UpdatePhase updates the game phase
	UpdatePhase(ctx context.Context, gameID string, newPhase types.GamePhase) error

	// UpdateStatus updates the game status
	UpdateStatus(ctx context.Context, gameID string, newStatus types.GameStatus) error

	// SetCurrentTurn sets which player's turn it is
	SetCurrentTurn(ctx context.Context, gameID string, playerID *string) error

	// AddPlayer adds a player ID reference to the game
	// Note: The actual player object is managed within the game.Game struct
	AddPlayer(ctx context.Context, gameID string, playerID string) error

	// SetHostPlayer sets the host player for the game
	SetHostPlayer(ctx context.Context, gameID string, playerID string) error
}
