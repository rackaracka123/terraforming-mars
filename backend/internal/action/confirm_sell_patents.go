package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

// ConfirmSellPatentsAction handles the business logic for confirming sell patents card selection
// This is Phase 2: processes the selected cards and awards credits
type ConfirmSellPatentsAction struct {
	BaseAction
	gameRepo game.Repository // Still needed for validation
}

// NewConfirmSellPatentsAction creates a new confirm sell patents action
func NewConfirmSellPatentsAction(
	gameRepo game.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *ConfirmSellPatentsAction {
	return &ConfirmSellPatentsAction{
		BaseAction: NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the confirm sell patents action (Phase 2: process card selection)
func (a *ConfirmSellPatentsAction) Execute(ctx context.Context, sess *session.Session, playerID string, selectedCardIDs []string) error {
	gameID := sess.GetGameID()
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

	// 3. Get session and player
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Validate pending card selection exists (card selection phase state on Player)
	pendingCardSelection := player.Selection().GetPendingCardSelection()
	if pendingCardSelection == nil {
		log.Warn("No pending card selection found")
		return fmt.Errorf("no pending card selection found")
	}

	if pendingCardSelection.Source != "sell-patents" {
		log.Warn("Pending card selection is not for sell patents",
			zap.String("source", pendingCardSelection.Source))
		return fmt.Errorf("pending card selection is not for sell patents")
	}

	// 5. Validate selection count
	if len(selectedCardIDs) < pendingCardSelection.MinCards {
		log.Warn("Too few cards selected",
			zap.Int("selected", len(selectedCardIDs)),
			zap.Int("min_required", pendingCardSelection.MinCards))
		return fmt.Errorf("must select at least %d cards", pendingCardSelection.MinCards)
	}

	if len(selectedCardIDs) > pendingCardSelection.MaxCards {
		log.Warn("Too many cards selected",
			zap.Int("selected", len(selectedCardIDs)),
			zap.Int("max_allowed", pendingCardSelection.MaxCards))
		return fmt.Errorf("cannot select more than %d cards", pendingCardSelection.MaxCards)
	}

	// 6. Validate all selected cards are in available cards
	availableCardsMap := make(map[string]bool)
	for _, cardID := range pendingCardSelection.AvailableCards {
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
		totalReward += pendingCardSelection.CardRewards[cardID]
	}

	// 8. Award credits
	if totalReward > 0 {
		resources := player.Resources().Get()
		resources.Credits += totalReward
		player.Resources().Set(resources)

		log.Info("💰 Awarded credits for sold cards",
			zap.Int("cards_sold", len(selectedCardIDs)),
			zap.Int("credits_earned", totalReward),
			zap.Int("new_credits", resources.Credits))
	}

	// 9. Remove sold cards from hand
	for _, cardID := range selectedCardIDs {
		removed := player.Hand().RemoveCard(cardID)
		if !removed {
			log.Warn("Failed to remove card from hand", zap.String("card_id", cardID))
		}
	}

	log.Info("🗑️ Removed sold cards from hand", zap.Int("cards_removed", len(selectedCardIDs)))

	// 10. Clear pending card selection (card selection phase state on Player)
	player.Selection().SetPendingCardSelection(nil)

	// 11. Consume action (only if player actually sold cards and not unlimited actions)
	availableActions := player.Turn().AvailableActions()
	if len(selectedCardIDs) > 0 && availableActions > 0 {
		player.Turn().SetAvailableActions(availableActions - 1)
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions-1))
	}

	// 12. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Sell patents completed successfully",
		zap.Int("cards_sold", len(selectedCardIDs)),
		zap.Int("credits_earned", totalReward))
	return nil
}
