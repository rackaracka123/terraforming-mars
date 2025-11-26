package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session/game"
	gameCore "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

// ValidateGameExists validates that a game exists (any status)
// Returns the game if valid, or an error if not found
func ValidateGameExists(
	ctx context.Context,
	gameRepo gameCore.Repository,
	gameID string,
	log *zap.Logger,
) (*game.Game, error) {
	g, err := gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return nil, fmt.Errorf("game not found: %w", err)
	}
	return g, nil
}

// ValidateActiveGame validates that a game exists and is in active status
// Returns the game if valid, or an error if not found or wrong status
func ValidateActiveGame(
	ctx context.Context,
	gameRepo gameCore.Repository,
	gameID string,
	log *zap.Logger,
) (*game.Game, error) {
	return ValidateGameStatus(ctx, gameRepo, gameID, gameCore.GameStatusActive, log)
}

// ValidateLobbyGame validates that a game exists and is in lobby status
// Returns the game if valid, or an error if not found or wrong status
func ValidateLobbyGame(
	ctx context.Context,
	gameRepo gameCore.Repository,
	gameID string,
	log *zap.Logger,
) (*game.Game, error) {
	return ValidateGameStatus(ctx, gameRepo, gameID, gameCore.GameStatusLobby, log)
}

// ValidateGameStatus validates that a game exists and has the expected status
// Returns the game if valid, or an error if not found or wrong status
func ValidateGameStatus(
	ctx context.Context,
	gameRepo gameCore.Repository,
	gameID string,
	expectedStatus gameCore.GameStatus,
	log *zap.Logger,
) (*game.Game, error) {
	g, err := gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return nil, fmt.Errorf("game not found: %w", err)
	}

	if g.Status != expectedStatus {
		log.Error("Game not in expected status",
			zap.String("expected", string(expectedStatus)),
			zap.String("actual", string(g.Status)))
		return nil, fmt.Errorf("game not in %s status", expectedStatus)
	}

	return g, nil
}

// ValidateGamePhase validates that a game is in the expected phase
// Returns error if game is not in the expected phase
func ValidateGamePhase(
	g *game.Game,
	expectedPhase gameCore.GamePhase,
	log *zap.Logger,
) error {
	if g.CurrentPhase != expectedPhase {
		log.Error("Game not in expected phase",
			zap.String("expected", string(expectedPhase)),
			zap.String("actual", string(g.CurrentPhase)))
		return fmt.Errorf("game not in %s phase", expectedPhase)
	}
	return nil
}

// ValidatePlayer is deprecated - use session.GetPlayer() directly instead
// This helper function is no longer needed with the new session-based architecture

// ValidateHostPermission validates that the specified player is the game host
// Returns error if player is not the host
func ValidateHostPermission(
	g *game.Game,
	playerID string,
	log *zap.Logger,
) error {
	if g.HostPlayerID != playerID {
		log.Error("Non-host attempted privileged action",
			zap.String("player_id", playerID),
			zap.String("host_id", g.HostPlayerID))
		return fmt.Errorf("only the host can perform this action")
	}
	return nil
}

// ValidateCurrentTurn validates that it's the specified player's turn
// Returns error if it's not their turn or no current turn is set
func ValidateCurrentTurn(
	g *game.Game,
	playerID string,
	log *zap.Logger,
) error {
	if g.CurrentTurn == nil {
		log.Error("No current turn set")
		return fmt.Errorf("no current turn set")
	}

	if *g.CurrentTurn != playerID {
		log.Error("Not player's turn",
			zap.String("player_id", playerID),
			zap.String("current_turn", *g.CurrentTurn))
		return fmt.Errorf("not your turn")
	}

	return nil
}
