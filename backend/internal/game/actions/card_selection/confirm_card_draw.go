package card_selection

import (
	"context"
	"fmt"
	"slices"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/game/resources"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// ConfirmCardDrawAction handles player confirmation of card draw/peek/take/buy selection
// This action orchestrates:
// - Validation of pending card draw selection
// - Validation of free vs paid card selections
// - Payment processing for bought cards via resources mechanic
// - Card distribution to player hand
// - Forced action completion check
// - Clearing pending selection
type ConfirmCardDrawAction struct {
	playerRepo          repository.PlayerRepository
	resourcesMech       resources.Service
	forcedActionManager cards.ForcedActionManager
	sessionManager      session.SessionManager
}

// NewConfirmCardDrawAction creates a new confirm card draw action
func NewConfirmCardDrawAction(
	playerRepo repository.PlayerRepository,
	resourcesMech resources.Service,
	forcedActionManager cards.ForcedActionManager,
	sessionManager session.SessionManager,
) *ConfirmCardDrawAction {
	return &ConfirmCardDrawAction{
		playerRepo:          playerRepo,
		resourcesMech:       resourcesMech,
		forcedActionManager: forcedActionManager,
		sessionManager:      sessionManager,
	}
}

// Execute performs the confirm card draw action
// Steps:
// 1. Get and validate pending card draw selection exists
// 2. Validate free take count
// 3. Validate pure card-draw scenario (all must be taken)
// 4. Validate buy count
// 5. Validate all selected cards are available
// 6. Process payment for bought cards via resources mechanic
// 7. Add all selected cards to hand
// 8. Discard unselected cards
// 9. Check if this was a forced action and mark complete
// 10. Clear pending selection
// 11. Broadcast state
func (a *ConfirmCardDrawAction) Execute(ctx context.Context, gameID string, playerID string, cardsToTake []string, cardsToBuy []string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🃏 Executing confirm card draw action",
		zap.Int("cards_to_take", len(cardsToTake)),
		zap.Int("cards_to_buy", len(cardsToBuy)))

	// Get player's pending card draw selection
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	if player.PendingCardDrawSelection == nil {
		log.Warn("No pending card draw selection found")
		return fmt.Errorf("no pending card draw selection found")
	}

	selection := player.PendingCardDrawSelection

	// Validate total cards selected
	totalSelected := len(cardsToTake) + len(cardsToBuy)
	maxAllowed := selection.FreeTakeCount + selection.MaxBuyCount

	if totalSelected > maxAllowed {
		log.Warn("Too many cards selected",
			zap.Int("selected", totalSelected),
			zap.Int("max_allowed", maxAllowed))
		return fmt.Errorf("too many cards selected: selected %d, max allowed %d", totalSelected, maxAllowed)
	}

	// Validate free take count
	if len(cardsToTake) > selection.FreeTakeCount {
		log.Warn("Too many free cards selected",
			zap.Int("selected", len(cardsToTake)),
			zap.Int("max", selection.FreeTakeCount))
		return fmt.Errorf("too many free cards selected: selected %d, max %d", len(cardsToTake), selection.FreeTakeCount)
	}

	// For pure card-draw scenarios (all cards must be taken, no choice), require player to take all
	// This only applies when MaxBuyCount = 0 AND FreeTakeCount = total available cards
	isPureCardDraw := selection.MaxBuyCount == 0 && selection.FreeTakeCount == len(selection.AvailableCards)
	if isPureCardDraw && len(cardsToTake) != selection.FreeTakeCount {
		log.Warn("Must take all cards for pure card-draw effect",
			zap.Int("required", selection.FreeTakeCount),
			zap.Int("selected", len(cardsToTake)))
		return fmt.Errorf("must take all %d cards for card-draw effect", selection.FreeTakeCount)
	}

	// Validate buy count
	if len(cardsToBuy) > selection.MaxBuyCount {
		log.Warn("Too many cards to buy",
			zap.Int("selected", len(cardsToBuy)),
			zap.Int("max", selection.MaxBuyCount))
		return fmt.Errorf("too many cards to buy: selected %d, max %d", len(cardsToBuy), selection.MaxBuyCount)
	}

	// Validate all selected cards are in available cards
	allSelectedCards := append(cardsToTake, cardsToBuy...)
	for _, cardID := range allSelectedCards {
		if !slices.Contains(selection.AvailableCards, cardID) {
			log.Warn("Card not in available cards", zap.String("card_id", cardID))
			return fmt.Errorf("card %s not in available cards", cardID)
		}
	}

	// Calculate total cost for bought cards
	totalCost := len(cardsToBuy) * selection.CardBuyCost

	// Process payment for bought cards via resources mechanic
	if totalCost > 0 {
		// Validate player can afford
		if player.Resources.Credits < totalCost {
			log.Warn("Insufficient credits to buy cards",
				zap.Int("cost", totalCost),
				zap.Int("available", player.Resources.Credits))
			return fmt.Errorf("insufficient credits to buy cards: need %d, have %d", totalCost, player.Resources.Credits)
		}

		// Deduct credits
		cost := resources.ResourceSet{
			Credits: totalCost,
		}

		if err := a.resourcesMech.PayResourceCost(ctx, gameID, playerID, cost); err != nil {
			log.Error("Failed to deduct credits for bought cards", zap.Error(err))
			return fmt.Errorf("failed to deduct credits for bought cards: %w", err)
		}

		log.Info("💰 Paid for bought cards",
			zap.Int("cards_bought", len(cardsToBuy)),
			zap.Int("cost", totalCost))
	}

	// Add all selected cards to player's hand
	for _, cardID := range allSelectedCards {
		if err := a.playerRepo.AddCard(ctx, gameID, playerID, cardID); err != nil {
			log.Error("Failed to add card to hand",
				zap.String("card_id", cardID),
				zap.Error(err))
			return fmt.Errorf("failed to add card %s to hand: %w", cardID, err)
		}
	}

	log.Info("🃏 Added selected cards to hand",
		zap.Int("cards_taken", len(cardsToTake)),
		zap.Int("cards_bought", len(cardsToBuy)),
		zap.Int("total_cards", len(allSelectedCards)))

	// Discard unselected cards (they were already popped from deck, so we just don't add them to hand)
	unselectedCards := []string{}
	for _, cardID := range selection.AvailableCards {
		if !slices.Contains(allSelectedCards, cardID) {
			unselectedCards = append(unselectedCards, cardID)
		}
	}

	if len(unselectedCards) > 0 {
		log.Debug("🗑️ Discarded unselected cards",
			zap.Int("count", len(unselectedCards)),
			zap.Strings("card_ids", unselectedCards))
	}

	// Check if this card draw was triggered by a forced action
	isForcedAction := false
	if player.ForcedFirstAction != nil && player.ForcedFirstAction.CorporationID == selection.Source {
		isForcedAction = true
	}

	// Clear pending card draw selection
	if err := a.playerRepo.ClearPendingCardDrawSelection(ctx, gameID, playerID); err != nil {
		log.Error("Failed to clear pending card draw selection", zap.Error(err))
		return fmt.Errorf("failed to clear pending card draw selection: %w", err)
	}

	// If this was a forced action, mark it as complete
	if isForcedAction {
		if err := a.forcedActionManager.MarkComplete(ctx, gameID, playerID); err != nil {
			log.Error("Failed to mark forced action complete", zap.Error(err))
			// Don't fail the operation, just log the error
		} else {
			log.Info("🎯 Forced action marked as complete", zap.String("corporation_id", selection.Source))
		}
	}

	// Broadcast game state update
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after card draw confirmation", zap.Error(err))
		// Don't fail the operation, just log the error
	}

	log.Info("✅ Confirm card draw action completed successfully",
		zap.String("source", selection.Source),
		zap.Int("cards_taken", len(cardsToTake)),
		zap.Int("cards_bought", len(cardsToBuy)),
		zap.Int("total_cost", totalCost))

	return nil
}
