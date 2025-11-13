package card_selection

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// SelectStartingCardsAction handles selection of starting cards and corporation
// This action orchestrates:
// - Card and corporation selection via cards.SelectionManager
// - Checking if all players are ready
// - Game phase advancement to action phase when all players ready
// - Clearing selection data
type SelectStartingCardsAction struct {
	selectionManager *cards.SelectionManager
	gameRepo         repository.GameRepository
	sessionManager   session.SessionManager
}

// NewSelectStartingCardsAction creates a new select starting cards action
func NewSelectStartingCardsAction(
	selectionManager *cards.SelectionManager,
	gameRepo repository.GameRepository,
	sessionManager session.SessionManager,
) *SelectStartingCardsAction {
	return &SelectStartingCardsAction{
		selectionManager: selectionManager,
		gameRepo:         gameRepo,
		sessionManager:   sessionManager,
	}
}

// Execute performs the select starting cards action
// Steps:
// 1. Process card and corporation selection via SelectionManager
// 2. Check if all players have completed selection
// 3. If all ready:
//    - Validate game phase
//    - Advance to action phase
//    - Clear selection data
// 4. Broadcast state
func (a *SelectStartingCardsAction) Execute(ctx context.Context, gameID string, playerID string, cardIDs []string, corporationID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🃏 Executing select starting cards action",
		zap.Int("card_count", len(cardIDs)),
		zap.String("corporation_id", corporationID))

	// Process selection via SelectionManager
	err := a.selectionManager.SelectStartingCards(ctx, gameID, playerID, cardIDs, corporationID)
	if err != nil {
		log.Error("Failed to select starting cards", zap.Error(err))
		return fmt.Errorf("failed to select starting cards: %w", err)
	}

	log.Debug("🃏 Player completed starting card selection", zap.Strings("card_ids", cardIDs))

	// Check if all players have completed their starting card selection
	if a.selectionManager.IsAllPlayersCardSelectionComplete(ctx, gameID) {
		log.Info("✅ All players completed starting card selection, advancing to action phase")

		// Get current game state to validate phase transition
		game, err := a.gameRepo.GetByID(ctx, gameID)
		if err != nil {
			log.Error("Failed to get game for phase advancement", zap.Error(err))
			return fmt.Errorf("failed to get game: %w", err)
		}

		// Validate current phase before transition
		if game.CurrentPhase != model.GamePhaseStartingCardSelection {
			log.Warn("Game is not in starting card selection phase, skipping phase transition",
				zap.String("current_phase", string(game.CurrentPhase)))
		} else if game.Status != model.GameStatusActive {
			log.Warn("Game is not active, skipping phase transition",
				zap.String("current_status", string(game.Status)))
		} else {
			// Advance to action phase
			if err := a.gameRepo.UpdatePhase(ctx, gameID, model.GamePhaseAction); err != nil {
				log.Error("Failed to update game phase", zap.Error(err))
				return fmt.Errorf("failed to update game phase: %w", err)
			}

			// Clear temporary card selection data
			a.selectionManager.ClearGameSelectionData(gameID)

			log.Info("🎯 Game phase advanced successfully",
				zap.String("previous_phase", string(model.GamePhaseStartingCardSelection)),
				zap.String("new_phase", string(model.GamePhaseAction)))

			// Note: Forced actions are now triggered via event system (GamePhaseChangedEvent)
		}
	}

	// Broadcast updated game state to all players after successful card selection (and potential phase change)
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after starting card selection", zap.Error(err))
		// Don't fail the card selection operation, just log the error
	}

	log.Info("✅ Select starting cards action completed successfully")

	return nil
}
