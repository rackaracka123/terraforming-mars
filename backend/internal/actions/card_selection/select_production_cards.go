package card_selection

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// SelectProductionCardsAction handles card selection during production phase
// This action orchestrates:
// - Card selection via cards.SelectionManager
// - Multi-player production phase coordination via GameService
// - Phase transition when all players ready
type SelectProductionCardsAction struct {
	selectionManager *cards.SelectionManager
	gameService      service.GameService
	sessionManager   session.SessionManager
}

// NewSelectProductionCardsAction creates a new select production cards action
func NewSelectProductionCardsAction(
	selectionManager *cards.SelectionManager,
	gameService service.GameService,
	sessionManager session.SessionManager,
) *SelectProductionCardsAction {
	return &SelectProductionCardsAction{
		selectionManager: selectionManager,
		gameService:      gameService,
		sessionManager:   sessionManager,
	}
}

// Execute performs the select production cards action
// Steps:
// 1. Process card selection via SelectionManager
// 2. Process production phase ready (marks player ready, checks if all ready)
// 3. If all ready: advance phase, reset action counts, clear production data
// 4. Broadcast state (handled by GameService.ProcessProductionPhaseReady)
//
// Note: This action delegates to GameService.ProcessProductionPhaseReady for
// multi-player coordination. This is temporary until production phase coordination
// is refactored into its own action or mechanic.
func (a *SelectProductionCardsAction) Execute(ctx context.Context, gameID string, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🃏 Executing select production cards action", zap.Int("card_count", len(cardIDs)))

	// Process card selection via SelectionManager
	if err := a.selectionManager.SelectProductionCards(ctx, gameID, playerID, cardIDs); err != nil {
		log.Error("Failed to select production cards", zap.Error(err))
		return fmt.Errorf("card selection failed: %w", err)
	}

	log.Debug("🃏 Production cards selected successfully", zap.Strings("card_ids", cardIDs))

	// Process production phase ready (multi-player coordination)
	// This marks player as ready and transitions phase if all players ready
	updatedGame, err := a.gameService.ProcessProductionPhaseReady(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to process production phase ready", zap.Error(err))
		return fmt.Errorf("failed to process production phase ready: %w", err)
	}

	log.Info("✅ Select production cards action completed successfully",
		zap.Strings("selected_cards", cardIDs),
		zap.String("game_phase", string(updatedGame.CurrentPhase)))

	return nil
}
