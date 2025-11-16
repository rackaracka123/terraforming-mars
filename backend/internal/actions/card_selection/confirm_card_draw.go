package card_selection

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// ConfirmCardDrawAction handles player confirmation of card draw/peek/take/buy selection
// This action orchestrates:
// - Card draw confirmation (inline, following ARCHITECTURE_FLOW.md)
// - Forced action completion check
// - State broadcasting
type ConfirmCardDrawAction struct {
	cardRepo       card.CardRepository
	cardDeckRepo   card.CardDeckRepository
	playerRepo     player.Repository
	sessionManager session.SessionManager
}

// NewConfirmCardDrawAction creates a new confirm card draw action
func NewConfirmCardDrawAction(
	cardRepo card.CardRepository,
	cardDeckRepo card.CardDeckRepository,
	playerRepo player.Repository,
	sessionManager session.SessionManager,
) *ConfirmCardDrawAction {
	return &ConfirmCardDrawAction{
		cardRepo:       cardRepo,
		cardDeckRepo:   cardDeckRepo,
		playerRepo:     playerRepo,
		sessionManager: sessionManager,
	}
}

// Execute performs the confirm card draw action
// Steps:
// 1. Combine cards to take and cards to buy into a single selection list
// 2. Delegate to DrawService.ConfirmCardDraw for validation and processing
// 3. Check if this was a forced action and mark complete
// 4. Broadcast state
func (a *ConfirmCardDrawAction) Execute(ctx context.Context, gameID string, playerID string, cardsToTake []string, cardsToBuy []string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🃏 Executing confirm card draw action",
		zap.Int("cards_to_take", len(cardsToTake)),
		zap.Int("cards_to_buy", len(cardsToBuy)))

	// Get player to check for forced action before DrawService clears the selection
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if this card draw was triggered by a forced action
	var isForcedAction bool
	var forcedActionSource string
	if player.PendingCardDrawSelection != nil {
		if player.ForcedFirstAction != nil && player.ForcedFirstAction.CorporationID == player.PendingCardDrawSelection.Source {
			isForcedAction = true
			forcedActionSource = player.PendingCardDrawSelection.Source
		}
	}

	// Combine all selected cards (both free and bought)
	allSelectedCards := append(cardsToTake, cardsToBuy...)

	// Process card draw confirmation inline
	if err := a.confirmCardDrawInline(ctx, gameID, playerID, allSelectedCards, log); err != nil {
		log.Error("Failed to confirm card draw", zap.Error(err))
		return fmt.Errorf("failed to confirm card draw: %w", err)
	}

	log.Info("🃏 Card draw confirmed successfully",
		zap.Int("cards_taken", len(cardsToTake)),
		zap.Int("cards_bought", len(cardsToBuy)),
		zap.Int("total_cards", len(allSelectedCards)))

	// If this was a forced action, mark it as complete
	if isForcedAction {
		if err := a.playerRepo.MarkForcedFirstActionComplete(ctx, gameID, playerID); err != nil {
			log.Error("Failed to mark forced action complete", zap.Error(err))
			// Don't fail the operation, just log the error
		} else {
			log.Info("🎯 Forced action marked as complete", zap.String("corporation_id", forcedActionSource))
		}
	}

	// Broadcast game state update
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after card draw confirmation", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	log.Info("✅ Confirm card draw action completed successfully")

	return nil
}

// confirmCardDrawInline processes card draw confirmation inline
// This implements the logic that was previously in DrawService.ConfirmCardDraw stub
func (a *ConfirmCardDrawAction) confirmCardDrawInline(ctx context.Context, gameID, playerID string, selectedCardIDs []string, log *zap.Logger) error {
	// Get player's PendingCardDrawSelection
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		return fmt.Errorf("failed to get player: %w", err)
	}

	if player.PendingCardDrawSelection == nil {
		return fmt.Errorf("no pending card draw selection found")
	}

	pending := player.PendingCardDrawSelection

	// Validate selectedCardIDs are subset of AvailableCards
	for _, selectedID := range selectedCardIDs {
		found := false
		for _, availableID := range pending.AvailableCards {
			if selectedID == availableID {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("selected card %s is not in available cards", selectedID)
		}
	}

	// Validate count: len(selectedCardIDs) <= FreeTakeCount + MaxBuyCount
	maxAllowed := pending.FreeTakeCount + pending.MaxBuyCount
	if len(selectedCardIDs) > maxAllowed {
		return fmt.Errorf("too many cards selected: %d selected, max allowed %d (free: %d, buy: %d)",
			len(selectedCardIDs), maxAllowed, pending.FreeTakeCount, pending.MaxBuyCount)
	}

	// Calculate cost
	freeCards := pending.FreeTakeCount
	if len(selectedCardIDs) < freeCards {
		freeCards = len(selectedCardIDs)
	}
	buyCards := len(selectedCardIDs) - freeCards
	cost := buyCards * pending.CardBuyCost

	log.Debug("💰 Card draw cost calculated",
		zap.Int("free_cards", freeCards),
		zap.Int("buy_cards", buyCards),
		zap.Int("total_cost", cost))

	// Deduct payment if there are cards to buy
	if cost > 0 {
		if player.Resources.Credits < cost {
			return fmt.Errorf("insufficient credits: need %d, have %d", cost, player.Resources.Credits)
		}

		newResources := player.Resources
		newResources.Credits -= cost
		if err := a.playerRepo.UpdateResources(ctx, gameID, playerID, newResources); err != nil {
			return fmt.Errorf("failed to deduct credits: %w", err)
		}

		log.Info("💸 Credits deducted for cards",
			zap.Int("cost", cost),
			zap.Int("remaining_credits", newResources.Credits))
	}

	// Add cards to hand
	for _, cardID := range selectedCardIDs {
		cardObj, err := a.cardRepo.GetCardByID(ctx, cardID)
		if err != nil {
			return fmt.Errorf("failed to get card %s: %w", cardID, err)
		}

		if err := a.playerRepo.AddCard(ctx, gameID, playerID, *cardObj); err != nil {
			return fmt.Errorf("failed to add card %s to hand: %w", cardID, err)
		}

		log.Debug("🃏 Card added to hand", zap.String("card_id", cardID), zap.String("card_name", cardObj.Name))
	}

	// Note: Unselected cards remain in the available pool and are not added to anyone's hand
	// They are implicitly "returned" by not being selected
	unselectedCount := len(pending.AvailableCards) - len(selectedCardIDs)
	if unselectedCount > 0 {
		log.Debug("🔄 Unselected cards remain available", zap.Int("count", unselectedCount))
	}

	// Clear PendingCardDrawSelection
	if err := a.playerRepo.UpdatePendingCardDrawSelection(ctx, gameID, playerID, nil); err != nil {
		return fmt.Errorf("failed to clear pending card draw selection: %w", err)
	}

	log.Debug("✅ Card draw confirmed and selection cleared")
	return nil
}
