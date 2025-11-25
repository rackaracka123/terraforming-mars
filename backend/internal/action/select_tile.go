package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game/board"
	game "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

// SelectTileAction handles tile selection and placement
type SelectTileAction struct {
	BaseAction
	gameRepo        game.Repository
	boardRepo       board.Repository
	tileProcessor   *board.Processor
	bonusCalculator *board.BonusCalculator
}

// NewSelectTileAction creates a new select tile action
func NewSelectTileAction(
	gameRepo game.Repository,
	sessionFactory session.SessionFactory,
	boardRepo board.Repository,
	tileProcessor *board.Processor,
	bonusCalculator *board.BonusCalculator,
	sessionMgrFactory session.SessionManagerFactory,
) *SelectTileAction {
	return &SelectTileAction{
		BaseAction:      NewBaseAction(sessionFactory, sessionMgrFactory),
		gameRepo:        gameRepo,
		boardRepo:       boardRepo,
		tileProcessor:   tileProcessor,
		bonusCalculator: bonusCalculator,
	}
}

// Execute places a tile at the selected coordinate
func (a *SelectTileAction) Execute(ctx context.Context, gameID, playerID string, q, r, s int) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.Int("q", q),
		zap.Int("r", r),
		zap.Int("s", s),
	)

	log.Info("🎯 Executing select tile action")

	// Get session and player
	sess := a.GetSessionFactory().Get(gameID)
	if sess == nil {
		log.Error("Game session not found")
		return fmt.Errorf("game not found: %s", gameID)
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// Validate player has pending tile selection
	if player.PendingTileSelection == nil {
		return fmt.Errorf("no pending tile selection for player")
	}

	// Validate coordinate
	coord := board.HexPosition{Q: q, R: r, S: s}
	coordStr := coord.String() // Format: "q,r,s"
	if !a.isValidCoordinateString(coordStr, player.PendingTileSelection.AvailableHexes) {
		return fmt.Errorf("invalid tile coordinate: not in available hexes")
	}

	// Create tile occupant based on tile type
	var occupant *board.TileOccupant
	switch player.PendingTileSelection.TileType {
	case board.TileTypeCity:
		occupant = &board.TileOccupant{
			Type: board.ResourceCityTile,
			Tags: []string{},
		}
	case board.TileTypeGreenery:
		occupant = &board.TileOccupant{
			Type: board.ResourceGreeneryTile,
			Tags: []string{},
		}
	case board.TileTypeOcean:
		occupant = &board.TileOccupant{
			Type: board.ResourceOceanTile,
			Tags: []string{},
		}
	default:
		return fmt.Errorf("unknown tile type: %s", player.PendingTileSelection.TileType)
	}

	// Place tile on board (publishes TilePlacedEvent)
	if err := a.boardRepo.UpdateTileOccupancy(ctx, gameID, coord, occupant, &playerID); err != nil {
		return fmt.Errorf("failed to place tile: %w", err)
	}

	log.Info("✅ Tile placed successfully",
		zap.String("tile_type", string(occupant.Type)))

	// Calculate and award tile bonuses
	if err := a.bonusCalculator.CalculateAndAwardBonuses(ctx, player, coord); err != nil {
		return fmt.Errorf("failed to award bonuses: %w", err)
	}

	// Handle greenery special case: increase oxygen and terraform rating
	if player.PendingTileSelection.TileType == board.TileTypeGreenery {
		if err := a.handleGreeneryPlacement(ctx, gameID, playerID, sess, log); err != nil {
			return fmt.Errorf("failed to handle greenery placement: %w", err)
		}
	}

	// Clear pending tile selection
	if err := player.TileQueue.ClearPendingTileSelection(ctx); err != nil {
		return fmt.Errorf("failed to clear pending selection: %w", err)
	}

	log.Info("🧹 Cleared pending tile selection")

	// Next tile processing (now automatic via TileQueueCreatedEvent)
	// When ProcessNextTileInQueue was called during tile placement validation,
	// it published an event if more tiles remain in the queue
	// No manual call needed - TileProcessor will automatically process the next tile

	// Broadcast updated game state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Select tile action completed successfully")
	return nil
}

// isValidCoordinateString checks if the coordinate string is in the available hexes list
func (a *SelectTileAction) isValidCoordinateString(coordStr string, availableHexes []string) bool {
	for _, hexStr := range availableHexes {
		if hexStr == coordStr {
			return true
		}
	}
	return false
}

// handleGreeneryPlacement increases oxygen and terraform rating when placing greenery
func (a *SelectTileAction) handleGreeneryPlacement(ctx context.Context, gameID, playerID string, sess *session.Session, log *zap.Logger) error {
	// Get current game state
	g, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Check if oxygen is already maxed
	if g.GlobalParameters.Oxygen >= 14 {
		log.Info("🌬️  Oxygen already at maximum, no increase")
		return nil
	}

	// Increase oxygen by 1%
	newOxygen := g.GlobalParameters.Oxygen + 1
	if newOxygen > 14 {
		newOxygen = 14
	}

	if err := a.gameRepo.UpdateOxygen(ctx, gameID, newOxygen); err != nil {
		return fmt.Errorf("failed to update oxygen: %w", err)
	}

	log.Info("🌬️  Increased oxygen",
		zap.Int("old_oxygen", g.GlobalParameters.Oxygen),
		zap.Int("new_oxygen", newOxygen))

	// Increase terraform rating by 1
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		return fmt.Errorf("player not found: %s", playerID)
	}

	newTR := player.TerraformRating + 1
	if err := player.Resources.UpdateTerraformRating(ctx, newTR); err != nil {
		return fmt.Errorf("failed to update terraform rating: %w", err)
	}

	log.Info("🏆 Increased terraform rating",
		zap.Int("old_tr", player.TerraformRating),
		zap.Int("new_tr", newTR))

	return nil
}
