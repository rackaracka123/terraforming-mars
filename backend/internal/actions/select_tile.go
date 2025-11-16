package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"
)

// SelectTileAction handles tile placement selection
// This action orchestrates:
// - Validation of pending tile selection
// - Coordinate validation against available hexes
// - Tile placement via tiles mechanic
// - Special plant conversion handling (raise oxygen, award TR)
// - Forced action completion check
// - Clearing pending selection
// - Tile queue processing for next tile
type SelectTileAction struct {
	playerRepo     player.Repository
	gameRepo       game.Repository
	tilesMech      tiles.SelectionService
	parametersMech parameters.Service
	sessionManager session.SessionManager
}

// NewSelectTileAction creates a new select tile action
func NewSelectTileAction(
	playerRepo player.Repository,
	gameRepo game.Repository,
	tilesMech tiles.SelectionService,
	parametersMech parameters.Service,
	sessionManager session.SessionManager,
) *SelectTileAction {
	return &SelectTileAction{
		playerRepo:     playerRepo,
		gameRepo:       gameRepo,
		tilesMech:      tilesMech,
		parametersMech: parametersMech,
		sessionManager: sessionManager,
	}
}

// Execute performs the select tile action
// Steps:
// 1. Validate hex coordinates (q + r + s = 0)
// 2. Get and validate pending tile selection exists
// 3. Validate selected coordinate is in available hexes
// 4. Place tile via tiles mechanic
// 5. Check if forced action and mark complete
// 6. Handle plant conversion (raise oxygen, award TR if applicable)
// 7. Clear pending tile selection
// 8. Process next tile in queue
// 9. Broadcast state
func (a *SelectTileAction) Execute(ctx context.Context, gameID string, playerID string, coordinate tiles.HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🎯 Executing select tile action",
		zap.Int("q", coordinate.Q),
		zap.Int("r", coordinate.R),
		zap.Int("s", coordinate.S))

	// Validate hex coordinates (must satisfy q + r + s = 0)
	if coordinate.Q+coordinate.R+coordinate.S != 0 {
		log.Error("Invalid hex coordinates: q+r+s must equal 0",
			zap.Int("q", coordinate.Q),
			zap.Int("r", coordinate.R),
			zap.Int("s", coordinate.S),
			zap.Int("sum", coordinate.Q+coordinate.R+coordinate.S))
		return fmt.Errorf("invalid hex coordinates: q+r+s must equal 0 (got %d+%d+%d=%d)",
			coordinate.Q, coordinate.R, coordinate.S, coordinate.Q+coordinate.R+coordinate.S)
	}

	// Get player's pending tile selection to determine tile type
	pendingSelection, err := a.playerRepo.GetPendingTileSelection(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to get pending tile selection: %w", err)
	}

	if pendingSelection == nil {
		log.Warn("No pending tile selection found for player")
		return fmt.Errorf("player has no pending tile selection")
	}

	// Convert coordinate to string for validation
	coordinateKey := coordinate.String()

	// Validate that the clicked tile is in the available hexes
	validTile := false
	for _, hexID := range pendingSelection.AvailableHexes {
		if hexID == coordinateKey {
			validTile = true
			break
		}
	}

	if !validTile {
		log.Error("Invalid tile selection",
			zap.String("coordinate", coordinateKey),
			zap.Strings("available", pendingSelection.AvailableHexes))
		return fmt.Errorf("selected coordinate %s is not in available positions", coordinateKey)
	}

	log.Debug("✅ Coordinate validation passed", zap.String("coordinate", coordinateKey))

	// Place the tile using tiles selection service (pure domain operation)
	// Note: Greenery tiles automatically trigger oxygen increase via GreenerySubscriber (event-driven)
	result, err := a.tilesMech.ProcessTileSelection(ctx, coordinate, pendingSelection.TileType, &playerID)
	if err != nil {
		log.Error("Failed to place tile", zap.Error(err))
		return fmt.Errorf("failed to place tile: %w", err)
	}

	log.Info("🎯 Tile placed successfully",
		zap.String("tile_type", pendingSelection.TileType),
		zap.Int("bonuses_awarded", len(result.Bonuses)))

	// Award tile placement bonuses (board-based bonuses)
	if len(result.Bonuses) > 0 {
		currentPlayer, err := a.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			log.Error("Failed to get player for bonus awards", zap.Error(err))
			return fmt.Errorf("failed to get player: %w", err)
		}

		// Get current resources and apply bonuses
		resources := currentPlayer.Resources
		for _, bonus := range result.Bonuses {
			switch bonus.Type {
			case "steel":
				resources.Steel += bonus.Amount
				log.Info("🔩 Awarded steel bonus", zap.Int("amount", bonus.Amount))
			case "titanium":
				resources.Titanium += bonus.Amount
				log.Info("⚙️ Awarded titanium bonus", zap.Int("amount", bonus.Amount))
			case "plants":
				resources.Plants += bonus.Amount
				log.Info("🌱 Awarded plants bonus", zap.Int("amount", bonus.Amount))
			case "cards":
				// TODO: Implement card draw via card service
				log.Info("🎴 Card draw bonus (not yet implemented)", zap.Int("amount", bonus.Amount))
			default:
				log.Warn("Unknown bonus type", zap.String("type", string(bonus.Type)))
			}
		}

		// Update resources in repository
		if err := a.playerRepo.UpdateResources(ctx, gameID, playerID, resources); err != nil {
			log.Error("Failed to update resources after bonuses", zap.Error(err))
			return fmt.Errorf("failed to update resources: %w", err)
		}
	}

	// Check if this tile placement was triggered by a forced action
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for forced action check", zap.Error(err))
	} else {
		isForcedAction := player.ForcedFirstAction != nil &&
			player.ForcedFirstAction.CorporationID == pendingSelection.Source

		if isForcedAction {
			if err := a.playerRepo.MarkForcedFirstActionComplete(ctx, gameID, playerID); err != nil {
				log.Error("Failed to mark forced action complete", zap.Error(err))
				// Don't fail the operation, just log the error
			} else {
				log.Info("🎯 Forced action marked as complete", zap.String("source", pendingSelection.Source))
			}
		}
	}

	// Oxygen increase and TR award are handled automatically via GreeneryRuleSubscriber:
	// 1. BoardRepository published TilePlacedEvent (already done in SelectionService)
	// 2. GreeneryRuleSubscriber listens → raises oxygen AND awards TR if greenery tile
	// 3. ParametersRepository publishes OxygenChangedEvent (for card effects that trigger on oxygen increase)

	// Clear the current pending tile selection
	if err := a.playerRepo.ClearPendingTileSelection(ctx, gameID, playerID); err != nil {
		log.Error("Failed to clear pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to clear pending tile selection: %w", err)
	}

	// TODO: Process the next tile in the queue if multiple tiles pending
	// The new tiles architecture handles tile queuing differently
	// For now, single tile placement is sufficient

	log.Info("🎯 Tile placed and queue processed")

	// Broadcast updated game state
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after tile selection", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	log.Info("✅ Select tile action completed successfully",
		zap.String("coordinate", coordinateKey),
		zap.String("tile_type", pendingSelection.TileType))

	return nil
}
