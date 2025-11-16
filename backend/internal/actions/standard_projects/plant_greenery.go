package standard_projects

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// PlantGreeneryAction handles the plant greenery standard project.
// This action orchestrates:
// 1. Validate player can afford 23 credits
// 2. Deduct 23 credits
// 3. Create pending tile selection for greenery placement
// 4. Broadcast updated game state
type PlantGreeneryAction struct {
	playerRepo       player.Repository
	gameRepo         game.Repository
	placementService tiles.PlacementService
	sessionManager   session.SessionManager
}

// NewPlantGreeneryAction creates a new plant greenery action
func NewPlantGreeneryAction(
	playerRepo player.Repository,
	gameRepo game.Repository,
	placementService tiles.PlacementService,
	sessionManager session.SessionManager,
) *PlantGreeneryAction {
	return &PlantGreeneryAction{
		playerRepo:       playerRepo,
		gameRepo:         gameRepo,
		placementService: placementService,
		sessionManager:   sessionManager,
	}
}

// Execute performs the plant greenery action
func (a *PlantGreeneryAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🌳 Executing plant greenery action")

	// Get cost from domain
	cost := domain.StandardProjectCosts.Greenery

	// Validate player can afford the cost
	canAfford, err := a.playerRepo.CanAfford(ctx, gameID, playerID, cost)
	if err != nil {
		log.Error("Failed to check affordability", zap.Error(err))
		return fmt.Errorf("failed to check affordability: %w", err)
	}

	if !canAfford {
		log.Warn("Player cannot afford greenery",
			zap.Int("cost", cost.Credits))
		return fmt.Errorf("insufficient credits: need %d", cost.Credits)
	}

	// Deduct credits
	if err := a.playerRepo.DeductResources(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct credits", zap.Error(err))
		return fmt.Errorf("failed to deduct credits: %w", err)
	}

	log.Info("💰 Credits deducted", zap.Int("amount", cost.Credits))

	// Calculate available hexes for greenery placement (player-specific)
	availableHexes, err := a.placementService.CalculateAvailablePositionsForPlayer(ctx, playerID, "greenery")
	if err != nil {
		log.Error("Failed to calculate available hexes", zap.Error(err))
		return fmt.Errorf("failed to calculate available hexes: %w", err)
	}

	// Convert HexPosition slice to string slice
	availableHexStrings := make([]string, len(availableHexes))
	for i, hex := range availableHexes {
		availableHexStrings[i] = hex.String()
	}

	// Create pending tile selection
	pendingSelection := &tiles.PendingTileSelection{
		TileType:       "greenery",
		AvailableHexes: availableHexStrings,
		Source:         "standard-project-greenery",
	}

	if err := a.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, pendingSelection); err != nil {
		log.Error("Failed to create pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to create pending tile selection: %w", err)
	}

	log.Info("🎯 Pending tile selection created", zap.Int("available_hexes", len(availableHexes)))
	log.Info("✅ Plant greenery action completed successfully")

	// Broadcast updated game state
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Don't fail the operation, just log
	}

	return nil
}
