package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/board"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/tile"

	"go.uber.org/zap"
)

// SelectTileAction handles tile selection and placement
type SelectTileAction struct {
	BaseAction
	boardRepo       board.Repository
	tileProcessor   *tile.Processor
	bonusCalculator *tile.BonusCalculator
}

// NewSelectTileAction creates a new select tile action
func NewSelectTileAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	boardRepo board.Repository,
	tileProcessor *tile.Processor,
	bonusCalculator *tile.BonusCalculator,
	sessionMgr session.SessionManager,
) *SelectTileAction {
	return &SelectTileAction{
		BaseAction:      NewBaseAction(gameRepo, playerRepo, sessionMgr),
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

	// Validate player exists
	p, err := ValidatePlayer(ctx, a.playerRepo, gameID, playerID, log)
	if err != nil {
		return err
	}

	// Validate player has pending tile selection
	if p.PendingTileSelection == nil {
		return fmt.Errorf("no pending tile selection for player")
	}

	// Validate coordinate
	coord := board.HexPosition{Q: q, R: r, S: s}
	coordStr := coord.String() // Format: "q,r,s"
	if !a.isValidCoordinateString(coordStr, p.PendingTileSelection.AvailableHexes) {
		return fmt.Errorf("invalid tile coordinate: not in available hexes")
	}

	// Create tile occupant based on tile type
	var occupant *board.TileOccupant
	switch p.PendingTileSelection.TileType {
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
		return fmt.Errorf("unknown tile type: %s", p.PendingTileSelection.TileType)
	}

	// Place tile on board (publishes TilePlacedEvent)
	if err := a.boardRepo.UpdateTileOccupancy(ctx, gameID, coord, occupant, &playerID); err != nil {
		return fmt.Errorf("failed to place tile: %w", err)
	}

	log.Info("✅ Tile placed successfully",
		zap.String("tile_type", string(occupant.Type)))

	// Calculate and award tile bonuses
	if err := a.bonusCalculator.CalculateAndAwardBonuses(ctx, gameID, playerID, coord); err != nil {
		return fmt.Errorf("failed to award bonuses: %w", err)
	}

	// Handle greenery special case: increase oxygen and terraform rating
	if p.PendingTileSelection.TileType == board.TileTypeGreenery {
		if err := a.handleGreeneryPlacement(ctx, gameID, playerID, log); err != nil {
			return fmt.Errorf("failed to handle greenery placement: %w", err)
		}
	}

	// Clear pending tile selection
	if err := a.playerRepo.ClearPendingTileSelection(ctx, gameID, playerID); err != nil {
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
func (a *SelectTileAction) handleGreeneryPlacement(ctx context.Context, gameID, playerID string, log *zap.Logger) error {
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
	p, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	newTR := p.TerraformRating + 1
	if err := a.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR); err != nil {
		return fmt.Errorf("failed to update terraform rating: %w", err)
	}

	log.Info("🏆 Increased terraform rating",
		zap.Int("old_tr", p.TerraformRating),
		zap.Int("new_tr", newTR))

	return nil
}
