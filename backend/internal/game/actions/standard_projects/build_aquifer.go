package standard_projects

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/game/parameters"
	"terraforming-mars-backend/internal/game/resources"
	"terraforming-mars-backend/internal/game/tiles"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

const (
	// AquiferCost is the credit cost to build an aquifer
	AquiferCost = 18
)

// BuildAquiferAction handles building an aquifer (ocean tile)
// This action orchestrates:
// - Resources mechanic (deduct credits)
// - Parameters mechanic (raise ocean count, award TR)
// - Tiles mechanic (process tile queue for placement)
type BuildAquiferAction struct {
	playerRepo     repository.PlayerRepository
	gameRepo       repository.GameRepository
	resourcesMech  resources.Service
	parametersMech parameters.Service
	tilesMech      tiles.Service
	sessionManager session.SessionManager
}

// NewBuildAquiferAction creates a new build aquifer action
func NewBuildAquiferAction(
	playerRepo repository.PlayerRepository,
	gameRepo repository.GameRepository,
	resourcesMech resources.Service,
	parametersMech parameters.Service,
	tilesMech tiles.Service,
	sessionManager session.SessionManager,
) *BuildAquiferAction {
	return &BuildAquiferAction{
		playerRepo:     playerRepo,
		gameRepo:       gameRepo,
		resourcesMech:  resourcesMech,
		parametersMech: parametersMech,
		tilesMech:      tilesMech,
		sessionManager: sessionManager,
	}
}

// Execute performs the build aquifer action
// Steps:
// 1. Validate player can afford (18 credits)
// 2. Deduct credits via resources mechanic
// 3. Raise ocean count via parameters mechanic (if not maxed)
// 4. Award TR via parameters mechanic (if ocean was raised)
// 5. Create tile queue for ocean placement
// 6. Process tile queue to prepare tile selection
// 7. Broadcast state
func (a *BuildAquiferAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🌊 Executing build aquifer action")

	// Validate player can afford
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	if player.Resources.Credits < AquiferCost {
		log.Warn("Player cannot afford aquifer",
			zap.Int("cost", AquiferCost),
			zap.Int("available", player.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", AquiferCost, player.Resources.Credits)
	}

	// Deduct credits via resources mechanic
	cost := resources.ResourceSet{
		Credits: AquiferCost,
	}

	if err := a.resourcesMech.PayResourceCost(ctx, gameID, playerID, cost); err != nil {
		log.Error("Failed to deduct credits", zap.Error(err))
		return fmt.Errorf("failed to deduct credits: %w", err)
	}

	log.Info("💰 Credits deducted", zap.Int("amount", AquiferCost))

	// Check if ocean can be raised
	game, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	oceanRaised := false
	if game.GlobalParameters.Oceans < model.MaxOceans {
		// Place ocean via parameters mechanic (also awards TR)
		if err := a.parametersMech.PlaceOcean(ctx, gameID, playerID); err != nil {
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

	// Create tile queue for ocean placement
	queueSource := "standard-project-ocean"
	if err := a.playerRepo.CreateTileQueue(ctx, gameID, playerID, queueSource, []string{"ocean"}); err != nil {
		log.Error("Failed to create tile queue", zap.Error(err))
		return fmt.Errorf("failed to create tile queue: %w", err)
	}

	// Process tile queue to set pendingTileSelection
	if err := a.tilesMech.ProcessTileQueue(ctx, gameID, playerID); err != nil {
		log.Error("Failed to process tile queue", zap.Error(err))
		return fmt.Errorf("failed to process tile queue: %w", err)
	}

	log.Info("🎯 Tile queue processed, awaiting player selection")

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
