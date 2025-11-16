package actions

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/domain"
	"terraforming-mars-backend/internal/features/tiles"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"
)

// ConvertPlantsToGreeneryAction handles converting plants to place a greenery tile
// This action orchestrates:
// 1. Validation and plant deduction
// 2. Available hex calculation
// 3. Pending tile selection creation
type ConvertPlantsToGreeneryAction struct {
	playerRepo       player.Repository
	placementService tiles.PlacementService
	sessionManager   session.SessionManager
}

// NewConvertPlantsToGreeneryAction creates a new plant conversion action
func NewConvertPlantsToGreeneryAction(
	playerRepo player.Repository,
	placementService tiles.PlacementService,
	sessionManager session.SessionManager,
) *ConvertPlantsToGreeneryAction {
	return &ConvertPlantsToGreeneryAction{
		playerRepo:       playerRepo,
		placementService: placementService,
		sessionManager:   sessionManager,
	}
}

// Execute performs the plant to greenery conversion
// Steps:
// 1. Validate player has enough plants
// 2. Deduct plants from player
// 3. Calculate available hexes for greenery placement
// 4. Create pending tile selection
// 5. Broadcast game state
func (a *ConvertPlantsToGreeneryAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🌱 Initiating plant conversion")

	// Use the standard project cost from domain
	cost := domain.StandardProjectCosts.ConvertPlantsToGreenery

	// Validate player can afford the cost
	canAfford, err := a.playerRepo.CanAfford(ctx, gameID, playerID, cost)
	if err != nil {
		log.Error("Failed to check affordability", zap.Error(err))
		return fmt.Errorf("failed to check affordability: %w", err)
	}

	if !canAfford {
		log.Warn("Player cannot afford plants conversion",
			zap.Int("required_plants", cost.Plants))
		return fmt.Errorf("insufficient plants: need %d", cost.Plants)
	}

	// Deduct plants
	if err := a.playerRepo.DeductResources(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct plants", zap.Error(err))
		return fmt.Errorf("failed to deduct plants: %w", err)
	}

	// Calculate available hexes using placement service
	var availableHexes []tiles.HexPosition
	if a.placementService != nil {
		var err error
		availableHexes, err = a.placementService.CalculateAvailablePositionsForPlayer(ctx, playerID, "greenery")
		if err != nil {
			log.Error("Failed to calculate available hexes", zap.Error(err))
			return fmt.Errorf("failed to calculate available hexes: %w", err)
		}
	} else {
		// Fallback: return empty list (will cause an error when player tries to place)
		log.Warn("PlacementService not available, using empty available hexes")
		availableHexes = []tiles.HexPosition{}
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
		Source:         "convert-plants-to-greenery",
	}

	if err := a.playerRepo.UpdatePendingTileSelection(ctx, gameID, playerID, pendingSelection); err != nil {
		log.Error("Failed to create pending tile selection", zap.Error(err))
		return fmt.Errorf("failed to create pending tile selection: %w", err)
	}

	// Broadcast updated game state
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Don't fail the conversion, just log
	}

	log.Info("✅ Plant conversion initiated, waiting for tile selection",
		zap.Int("plants_spent", cost.Plants),
		zap.Int("available_hexes", len(availableHexes)))

	return nil
}
