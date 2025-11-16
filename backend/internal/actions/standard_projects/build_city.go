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

// BuildCityAction handles the build city standard project.
// This action orchestrates:
// 1. Validate player can afford 25 credits
// 2. Deduct 25 credits
// 3. Increase credit production by 1
// 4. Create pending tile selection for city placement
// 5. Broadcast updated game state
type BuildCityAction struct {
	playerRepo       player.Repository
	gameRepo         game.Repository
	placementService tiles.PlacementService
	sessionManager   session.SessionManager
}

// NewBuildCityAction creates a new build city action
func NewBuildCityAction(
	playerRepo player.Repository,
	gameRepo game.Repository,
	placementService tiles.PlacementService,
	sessionManager session.SessionManager,
) *BuildCityAction {
	return &BuildCityAction{
		playerRepo:       playerRepo,
		gameRepo:         gameRepo,
		placementService: placementService,
		sessionManager:   sessionManager,
	}
}

// Execute performs the build city action
func (a *BuildCityAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🏙️ Executing build city action")

	// Get cost from domain
	cost := domain.StandardProjectCosts.City

	// Validate player can afford the cost
	canAfford, err := a.playerRepo.CanAfford(ctx, gameID, playerID, cost)
	if err != nil {
		log.Error("Failed to check affordability", zap.Error(err))
		return fmt.Errorf("failed to check affordability: %w", err)
	}

	if !canAfford {
		log.Warn("Player cannot afford city",
			zap.Int("cost", cost.Credits))
		return fmt.Errorf("insufficient credits: need %d", cost.Credits)
	}

	// Deduct credits
	if err := a.playerRepo.DeductResources(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct credits", zap.Error(err))
		return fmt.Errorf("failed to deduct credits: %w", err)
	}

	log.Info("💰 Credits deducted", zap.Int("amount", cost.Credits))

	// Increase credit production by 1
	production := domain.ResourceSet{
		Credits: 1,
	}

	if err := a.playerRepo.AddProduction(ctx, gameID, playerID, production); err != nil {
		log.Error("Failed to increase credit production", zap.Error(err))
		return fmt.Errorf("failed to increase credit production: %w", err)
	}

	log.Info("💰 Credit production increased", zap.Int("amount", 1))

	// Calculate available hexes for city placement
	availableHexes, err := a.placementService.CalculateAvailablePositions(ctx, "city")
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
		TileType:       "city",
		AvailableHexes: availableHexStrings,
		Source:         "standard-project-city",
	}

	if err := a.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, pendingSelection); err != nil {
		log.Error("Failed to create pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to create pending tile selection: %w", err)
	}

	log.Info("🎯 Pending tile selection created", zap.Int("available_hexes", len(availableHexes)))
	log.Info("✅ Build city action completed successfully")

	// Broadcast updated game state
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Don't fail the operation, just log
	}

	return nil
}
