package admin

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/session"
	sessionBoard "terraforming-mars-backend/internal/session/game/board"
	sessionGame "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/types"
)

// StartTileSelectionAction handles admin command to start tile selection for testing
type StartTileSelectionAction struct {
	gameRepo          sessionGame.Repository
	sessionFactory    session.SessionFactory
	boardRepo         sessionBoard.Repository
	boardProcessor    *sessionBoard.BoardProcessor
	sessionMgrFactory session.SessionManagerFactory
	logger            *zap.Logger
}

// NewStartTileSelectionAction creates a new start tile selection action
func NewStartTileSelectionAction(
	gameRepo sessionGame.Repository,
	sessionFactory session.SessionFactory,
	boardRepo sessionBoard.Repository,
	boardProcessor *sessionBoard.BoardProcessor,
	sessionMgrFactory session.SessionManagerFactory,
) *StartTileSelectionAction {
	return &StartTileSelectionAction{
		gameRepo:          gameRepo,
		sessionFactory:    sessionFactory,
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

	// 1. Get session and validate player exists
	sess := a.sessionFactory.Get(gameID)
	if sess == nil {
		log.Error("Game session not found")
		return fmt.Errorf("game session not found")
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found")
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

	// 5. Set pending tile selection via player's tile queue repository
	pendingSelection := &types.PendingTileSelection{
		TileType:       tileType,
		AvailableHexes: availableHexes,
		Source:         "admin_demo",
	}

	if err := player.TileQueue.UpdatePendingTileSelection(ctx, pendingSelection); err != nil {
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
