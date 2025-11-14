package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
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
	playerRepo          player.Repository
	gameRepo            game.Repository
	tilesMech           tiles.Service
	parametersMech      parameters.Service
	forcedActionManager service.ForcedActionManager
	sessionManager      session.SessionManager
}

// NewSelectTileAction creates a new select tile action
func NewSelectTileAction(
	playerRepo player.Repository,
	gameRepo game.Repository,
	tilesMech tiles.Service,
	parametersMech parameters.Service,
	forcedActionManager service.ForcedActionManager,
	sessionManager session.SessionManager,
) *SelectTileAction {
	return &SelectTileAction{
		playerRepo:          playerRepo,
		gameRepo:            gameRepo,
		tilesMech:           tilesMech,
		parametersMech:      parametersMech,
		forcedActionManager: forcedActionManager,
		sessionManager:      sessionManager,
	}
}

// Execute performs the select tile action
// Steps:
// 1. Get and validate pending tile selection exists
// 2. Validate selected coordinate is in available hexes
// 3. Place tile via tiles mechanic
// 4. Check if forced action and mark complete
// 5. Handle plant conversion (raise oxygen, award TR if applicable)
// 6. Clear pending tile selection
// 7. Process next tile in queue
// 8. Broadcast state
func (a *SelectTileAction) Execute(ctx context.Context, gameID string, playerID string, coordinate types.HexPosition) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🎯 Executing select tile action",
		zap.Int("q", coordinate.Q),
		zap.Int("r", coordinate.R),
		zap.Int("s", coordinate.S))

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

	// Check if this is a plant conversion (special handling for raising oxygen and TR)
	isPlantConversion := pendingSelection.Source == "convert-plants-to-greenery"

	// Convert types.HexPosition to types.HexPosition for tiles mechanic
	tilesCoordinate := types.HexPosition{
		Q: coordinate.Q,
		R: coordinate.R,
		S: coordinate.S,
	}

	// Place the tile using tiles mechanic
	if err := a.tilesMech.PlaceTile(ctx, gameID, playerID, pendingSelection.TileType, tilesCoordinate); err != nil {
		log.Error("Failed to place tile", zap.Error(err))
		return fmt.Errorf("failed to place tile: %w", err)
	}

	log.Info("🎯 Tile placed successfully", zap.String("tile_type", pendingSelection.TileType))

	// Check if this tile placement was triggered by a forced action
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for forced action check", zap.Error(err))
	} else {
		isForcedAction := player.ForcedFirstAction != nil &&
			player.ForcedFirstAction.CorporationID == pendingSelection.Source

		if isForcedAction {
			if err := a.forcedActionManager.MarkComplete(ctx, gameID, playerID); err != nil {
				log.Error("Failed to mark forced action complete", zap.Error(err))
				// Don't fail the operation, just log the error
			} else {
				log.Info("🎯 Forced action marked as complete", zap.String("source", pendingSelection.Source))
			}
		}
	}

	// Handle plant conversion completion (raise oxygen and TR)
	if isPlantConversion {
		log.Info("🌱 Completing plant conversion - raising oxygen")

		// Get game for oxygen check
		game, err := a.gameRepo.GetByID(ctx, gameID)
		if err != nil {
			log.Error("Failed to get game for oxygen update", zap.Error(err))
			return fmt.Errorf("failed to get game: %w", err)
		}

		// Raise oxygen if not already maxed
		if game.GlobalParameters.Oxygen < model.MaxOxygen {
			// Use parameters mechanic to raise oxygen (it also awards TR automatically)
			stepsRaised, err := a.parametersMech.RaiseOxygen(ctx, gameID, playerID, 1)
			if err != nil {
				log.Error("Failed to raise oxygen", zap.Error(err))
				return fmt.Errorf("failed to raise oxygen: %w", err)
			}

			if stepsRaised > 0 {
				log.Info("🌍 Oxygen raised and TR awarded via parameters mechanic", zap.Int("steps", stepsRaised))
			} else {
				log.Info("🌍 Oxygen at maximum, no change")
			}
		} else {
			log.Info("🌍 Oxygen already at maximum, no change")
		}

		log.Info("✅ Plant conversion completed successfully")
	}

	// Clear the current pending tile selection
	if err := a.playerRepo.ClearPendingTileSelection(ctx, gameID, playerID); err != nil {
		log.Error("Failed to clear pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to clear pending tile selection: %w", err)
	}

	// Process the next tile in the queue
	if err := a.tilesMech.ProcessTileQueue(ctx, gameID, playerID); err != nil {
		log.Error("Failed to process next tile in queue", zap.Error(err))
		return fmt.Errorf("failed to process next tile in queue: %w", err)
	}

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
