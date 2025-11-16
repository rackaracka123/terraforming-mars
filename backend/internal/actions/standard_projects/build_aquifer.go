package standard_projects

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/features/parameters"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// BuildAquiferAction handles building an aquifer (ocean tile)
// This action orchestrates:
// - Deduct credits
// - Raise ocean count and award TR
// - Process tile queue for placement
type BuildAquiferAction struct {
	playerRepo        player.Repository
	parametersService parameters.Service
	placementService  tiles.PlacementService
	sessionManager    session.SessionManager
}

// NewBuildAquiferAction creates a new build aquifer action
func NewBuildAquiferAction(
	playerRepo player.Repository,
	parametersService parameters.Service,
	placementService tiles.PlacementService,
	sessionManager session.SessionManager,
) *BuildAquiferAction {
	return &BuildAquiferAction{
		playerRepo:        playerRepo,
		parametersService: parametersService,
		placementService:  placementService,
		sessionManager:    sessionManager,
	}
}

// Execute performs the build aquifer action
// Steps:
// 1. Validate player can afford (18 credits)
// 2. Deduct credits
// 3. Raise ocean count (if not maxed)
// 4. Award TR (if ocean was raised)
// 5. Calculate available hexes and create pending tile selection
// 6. Broadcast state
func (a *BuildAquiferAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🌊 Executing build aquifer action")

	// Get cost from domain
	cost := domain.StandardProjectCosts.Aquifer

	// Validate player can afford the cost
	canAfford, err := a.playerRepo.CanAfford(ctx, gameID, playerID, cost)
	if err != nil {
		log.Error("Failed to check affordability", zap.Error(err))
		return fmt.Errorf("failed to check affordability: %w", err)
	}

	if !canAfford {
		log.Warn("Player cannot afford aquifer",
			zap.Int("cost", cost.Credits))
		return fmt.Errorf("insufficient credits: need %d", cost.Credits)
	}

	// Deduct credits
	if err := a.playerRepo.DeductResources(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct credits", zap.Error(err))
		return fmt.Errorf("failed to deduct credits: %w", err)
	}

	log.Info("💰 Credits deducted", zap.Int("amount", cost.Credits))

	// Check if ocean can be raised
	globalParams, err := a.parametersService.GetGlobalParameters(ctx)
	if err != nil {
		log.Error("Failed to get global parameters", zap.Error(err))
		return fmt.Errorf("failed to get global parameters: %w", err)
	}

	oceanRaised := false
	if globalParams.Oceans < parameters.MaxOceans {
		// Place ocean via parameters service (also awards TR)
		if err := a.parametersService.PlaceOcean(ctx); err != nil {
			log.Error("Failed to place ocean", zap.Error(err))
			return fmt.Errorf("failed to place ocean: %w", err)
		}

		oceanRaised = true
		log.Info("🌊 Ocean placed and TR awarded")
	} else {
		log.Info("🌊 Oceans already at maximum, no TR awarded")

		// Still need to award TR for placing ocean tile even if ocean count maxed
		// (per game rules - you get TR for the tile placement itself)
		currentPlayer, err := a.playerRepo.GetByID(ctx, gameID, playerID)
		if err != nil {
			log.Error("Failed to get player for TR update", zap.Error(err))
			return fmt.Errorf("failed to get player: %w", err)
		}

		newTR := currentPlayer.TerraformRating + 1
		if err := a.playerRepo.UpdateTerraformRating(ctx, gameID, playerID, newTR); err != nil {
			log.Error("Failed to update terraform rating", zap.Error(err))
			return fmt.Errorf("failed to update terraform rating: %w", err)
		}

		log.Info("⭐ Terraform rating increased (ocean maxed)", zap.Int("new_tr", newTR))
	}

	// Calculate available hexes for ocean placement
	availableHexes, err := a.placementService.CalculateAvailablePositions(ctx, "ocean")
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
		TileType:       "ocean",
		AvailableHexes: availableHexStrings,
		Source:         "standard-project-aquifer",
	}

	if err := a.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, pendingSelection); err != nil {
		log.Error("Failed to create pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to create pending tile selection: %w", err)
	}

	log.Info("🎯 Pending tile selection created", zap.Int("available_hexes", len(availableHexes)))

	// Broadcast updated game state (includes pendingTileSelection)
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Don't fail the action, just log
	}

	if oceanRaised {
		log.Info("✅ Build aquifer action completed successfully - ocean raised")
	} else {
		log.Info("✅ Build aquifer action completed successfully - ocean maxed")
	}

	return nil
}
