package card_selection

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"
)

// SelectCardsAction handles card selection routing
// Determines whether to route to sell patents or production card selection based on pending state
type SelectCardsAction struct {
	playerRepo                  player.Repository
	submitSellPatentsAction     *SubmitSellPatentsAction
	selectProductionCardsAction *SelectProductionCardsAction
}

// NewSelectCardsAction creates a new select cards action
func NewSelectCardsAction(
	playerRepo player.Repository,
	submitSellPatentsAction *SubmitSellPatentsAction,
	selectProductionCardsAction *SelectProductionCardsAction,
) *SelectCardsAction {
	return &SelectCardsAction{
		playerRepo:                  playerRepo,
		submitSellPatentsAction:     submitSellPatentsAction,
		selectProductionCardsAction: selectProductionCardsAction,
	}
}

// Execute processes card selection by routing to the appropriate action
func (a *SelectCardsAction) Execute(ctx context.Context, gameID, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Debug("🃏 Executing select cards action",
		zap.Strings("card_ids", cardIDs),
		zap.Int("count", len(cardIDs)))

	// Check if player has a pending card selection (e.g., sell patents)
	pendingCardSelection, err := a.playerRepo.GetPendingCardSelection(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to check pending card selection", zap.Error(err))
		return fmt.Errorf("failed to check pending card selection: %w", err)
	}

	// Route to the appropriate action based on pending state
	if pendingCardSelection != nil {
		log.Info("Routing to SubmitSellPatentsAction",
			zap.String("source", pendingCardSelection.Source),
			zap.Int("cards_selected", len(cardIDs)))

		if err := a.submitSellPatentsAction.Execute(ctx, gameID, playerID, cardIDs); err != nil {
			return err
		}

		log.Info("✅ Pending card selection completed",
			zap.String("source", pendingCardSelection.Source))
		return nil
	}

	// Otherwise, handle as production card selection
	log.Debug("Routing to SelectProductionCardsAction")
	if err := a.selectProductionCardsAction.Execute(ctx, gameID, playerID, cardIDs); err != nil {
		log.Error("Failed to select production cards", zap.Error(err))
		return fmt.Errorf("card selection failed: %w", err)
	}

	log.Info("✅ Production card selection completed",
		zap.Strings("selected_cards", cardIDs))

	return nil
}
