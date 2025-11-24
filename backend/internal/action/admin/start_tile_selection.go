package admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"
	sessionBoard "terraforming-mars-backend/internal/session/board"
	sessionGame "terraforming-mars-backend/internal/session/game"
	sessionPlayer "terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/types"
)

// StartTileSelectionAction handles admin command to start tile selection for testing
type StartTileSelectionAction struct {
	gameRepo          sessionGame.Repository
	playerRepo        sessionPlayer.Repository
	boardRepo         sessionBoard.Repository
	boardProcessor    *sessionBoard.BoardProcessor
	sessionMgrFactory session.SessionManagerFactory
	logger            *zap.Logger
}

// NewStartTileSelectionAction creates a new start tile selection action
func NewStartTileSelectionAction(
	gameRepo sessionGame.Repository,
	playerRepo sessionPlayer.Repository,
	boardRepo sessionBoard.Repository,
	boardProcessor *sessionBoard.BoardProcessor,
	sessionMgrFactory session.SessionManagerFactory,
) *StartTileSelectionAction {
	return &StartTileSelectionAction{
		gameRepo:          gameRepo,
		playerRepo:        playerRepo,
		boardRepo:         boardRepo,
		boardProcessor:    boardProcessor,
		sessionMgrFactory: sessionMgrFactory,
		logger:            logger.Get(),
	}
}

// Execute starts tile selection for a player (admin/testing feature)
func (a *StartTileSelectionAction) Execute(ctx context.Context, gameID string, playerID string, tileType string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🎯 Admin starting tile selection", zap.String("tile_type", tileType))

	// 1. Validate player exists
	_, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Player not found", zap.Error(err))
		return fmt.Errorf("player not found: %w", err)
	}

	// 2. Get actual board from board repository
	board, err := a.boardRepo.GetByGameID(ctx, gameID)
	if err != nil {
		log.Error("Board not found", zap.Error(err))
		return fmt.Errorf("board not found: %w", err)
	}

	// 3. Calculate available hexes based on tile type using board processor
	availableHexes := a.boardProcessor.CalculateAvailableHexesForTileType(board, tileType, &playerID)

	// 4. Validate available positions
	if len(availableHexes) == 0 {
		log.Warn("No valid positions available for tile type",
			zap.String("tile_type", tileType),
			zap.Int("total_tiles", len(board.Tiles)))
		return fmt.Errorf("no valid positions available for %s placement", tileType)
	}

	log.Info("✅ Found available tile positions",
		zap.Int("available_count", len(availableHexes)),
		zap.Strings("positions", availableHexes[:min(5, len(availableHexes))])) // Log first 5

	// 5. Set pending tile selection
	pendingSelection := &types.PendingTileSelection{
		TileType:       tileType,
		AvailableHexes: availableHexes,
		Source:         "admin_demo",
	}

	if err := a.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, pendingSelection); err != nil {
		log.Error("Failed to set pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to set pending tile selection: %w", err)
	}

	log.Info("✅ Tile selection started successfully",
		zap.String("tile_type", tileType),
		zap.Int("available_positions", len(availableHexes)))

	// 6. Broadcast updated game state
	if err := a.sessionMgrFactory.GetOrCreate(gameID).Broadcast(); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Don't fail the operation, just log
	}

	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
