package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// ConfirmSellPatentsAction handles the business logic for confirming sell patents card selection
// This is Phase 2: processes the selected cards and awards credits
type ConfirmSellPatentsAction struct {
	BaseAction
}

// NewConfirmSellPatentsAction creates a new confirm sell patents action
func NewConfirmSellPatentsAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *ConfirmSellPatentsAction {
	return &ConfirmSellPatentsAction{
		BaseAction: NewBaseAction(gameRepo, playerRepo, sessionMgrFactory),
	}
}

// Execute performs the confirm sell patents action (Phase 2: process card selection)
func (a *ConfirmSellPatentsAction) Execute(ctx context.Context, gameID, playerID string, selectedCardIDs []string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🏛️ Confirming sell patents card selection", zap.Int("cards_selected", len(selectedCardIDs)))

	// 1. Validate game is active
	g, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate it's the player's turn
	if err := ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 3. Validate player exists
	p, err := ValidatePlayer(ctx, a.playerRepo, gameID, playerID, log)
	if err != nil {
		return err
	}

	// 4. Validate pending card selection exists
	if p.PendingCardSelection == nil {
		log.Warn("No pending card selection found")
		return fmt.Errorf("no pending card selection found")
	}

	if p.PendingCardSelection.Source != "sell-patents" {
		log.Warn("Pending card selection is not for sell patents",
			zap.String("source", p.PendingCardSelection.Source))
		return fmt.Errorf("pending card selection is not for sell patents")
	}

	// 5. Validate selection count
	if len(selectedCardIDs) < p.PendingCardSelection.MinCards {
		log.Warn("Too few cards selected",
			zap.Int("selected", len(selectedCardIDs)),
			zap.Int("min_required", p.PendingCardSelection.MinCards))
		return fmt.Errorf("must select at least %d cards", p.PendingCardSelection.MinCards)
	}

	if len(selectedCardIDs) > p.PendingCardSelection.MaxCards {
		log.Warn("Too many cards selected",
			zap.Int("selected", len(selectedCardIDs)),
			zap.Int("max_allowed", p.PendingCardSelection.MaxCards))
		return fmt.Errorf("cannot select more than %d cards", p.PendingCardSelection.MaxCards)
	}

	// 6. Validate all selected cards are in available cards
	availableCardsMap := make(map[string]bool)
	for _, cardID := range p.PendingCardSelection.AvailableCards {
		availableCardsMap[cardID] = true
	}

	for _, cardID := range selectedCardIDs {
		if !availableCardsMap[cardID] {
			log.Warn("Selected card not available", zap.String("card_id", cardID))
			return fmt.Errorf("card %s is not available for selection", cardID)
		}
	}

	// 7. Calculate total reward (1 M€ per card)
	totalReward := 0
	for _, cardID := range selectedCardIDs {
		totalReward += p.PendingCardSelection.CardRewards[cardID]
	}

	// 8. Award credits
	if totalReward > 0 {
		newResources := p.Resources
		newResources.Credits += totalReward
		err = a.playerRepo.UpdateResources(ctx, gameID, playerID, newResources)
		if err != nil {
			log.Error("Failed to award credits", zap.Error(err))
			return fmt.Errorf("failed to award credits: %w", err)
		}

		log.Info("💰 Awarded credits for sold cards",
			zap.Int("cards_sold", len(selectedCardIDs)),
			zap.Int("credits_earned", totalReward),
			zap.Int("new_credits", newResources.Credits))
	}

	// 9. Remove sold cards from hand
	for _, cardID := range selectedCardIDs {
		err = a.playerRepo.RemoveCardFromHand(ctx, gameID, playerID, cardID)
		if err != nil {
			log.Error("Failed to remove card from hand",
				zap.Error(err),
				zap.String("card_id", cardID))
			return fmt.Errorf("failed to remove card from hand: %w", err)
		}
	}

	log.Info("🗑️ Removed sold cards from hand", zap.Int("cards_removed", len(selectedCardIDs)))

	// 10. Clear pending card selection
	err = a.playerRepo.ClearPendingCardSelection(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to clear pending card selection", zap.Error(err))
		return fmt.Errorf("failed to clear pending card selection: %w", err)
	}

	// 11. Consume action (only if player actually sold cards and not unlimited actions)
	// Refresh player data after clearing pending selection
	p, err = ValidatePlayer(ctx, a.playerRepo, gameID, playerID, log)
	if err != nil {
		return err
	}

	if len(selectedCardIDs) > 0 && p.AvailableActions > 0 {
		newActions := p.AvailableActions - 1
		err = a.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, newActions)
		if err != nil {
			log.Error("Failed to consume action", zap.Error(err))
			return fmt.Errorf("failed to consume action: %w", err)
		}
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", newActions))
	}

	// 12. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Sell patents completed successfully",
		zap.Int("cards_sold", len(selectedCardIDs)),
		zap.Int("credits_earned", totalReward))
	return nil
}
