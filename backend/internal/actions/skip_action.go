package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/production"
	"terraforming-mars-backend/internal/features/turn"
	"terraforming-mars-backend/internal/logger"

	"go.uber.org/zap"
)

// SkipAction handles the player skip turn action
// This action orchestrates the turn mechanic and production mechanic
type SkipAction struct {
	turnFeature       turn.Service
	productionFeature production.Service
	sessionManager    session.SessionManager
}

// NewSkipAction creates a new skip action orchestrator
func NewSkipAction(
	turnFeature turn.Service,
	productionFeature production.Service,
	sessionManager session.SessionManager,
) *SkipAction {
	return &SkipAction{
		turnFeature:       turnFeature,
		productionFeature: productionFeature,
		sessionManager:    sessionManager,
	}
}

// Execute performs the skip turn action
// This orchestrates:
// 1. Turn mechanic to advance turn/check generation end
// 2. Production mechanic if generation ended
// 3. Session broadcasting to notify all players
func (a *SkipAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("⏭️ Executing skip turn action")

	// Delegate to turn mechanic
	generationEnded, err := a.turnFeature.SkipTurn(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to skip turn", zap.Error(err))
		return fmt.Errorf("failed to skip turn: %w", err)
	}

	// If generation ended, execute production phase
	if generationEnded {
		log.Info("🏭 Generation ended - executing production phase")

		if err := a.productionFeature.ExecuteProductionPhase(ctx, gameID); err != nil {
			log.Error("Failed to execute production phase", zap.Error(err))
			return fmt.Errorf("failed to execute production phase: %w", err)
		}
	}

	// Broadcast updated game state to all players
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after skip turn", zap.Error(err))
		// Don't fail the skip operation, just log the error
	}

	log.Info("✅ Skip turn action completed successfully")
	return nil
}
