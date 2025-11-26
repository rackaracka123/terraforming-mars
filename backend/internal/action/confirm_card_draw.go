package action

import (
	"context"
	"fmt"
	"slices"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"
	playerevents "terraforming-mars-backend/internal/session/game/player"

	"go.uber.org/zap"
)

// ConfirmCardDrawAction handles the business logic for confirming card draw selection
type ConfirmCardDrawAction struct {
	BaseAction
	gameRepo game.Repository
	eventBus *events.EventBusImpl
}

// NewConfirmCardDrawAction creates a new confirm card draw action
func NewConfirmCardDrawAction(
	gameRepo game.Repository,
	sessionMgrFactory session.SessionManagerFactory,
	eventBus *events.EventBusImpl,
) *ConfirmCardDrawAction {
	return &ConfirmCardDrawAction{
		BaseAction: NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
		eventBus:   eventBus,
	}
}

// Execute performs the confirm card draw action
func (a *ConfirmCardDrawAction) Execute(ctx context.Context, sess *session.Session, playerID string, cardsToTake []string, cardsToBuy []string) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID)
	log.Info("🃏 Confirming card draw selection",
		zap.Int("cards_to_take", len(cardsToTake)),
		zap.Int("cards_to_buy", len(cardsToBuy)))

	// 1. Validate game is active
	_, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Get session and player
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. Validate pending card draw selection exists (card selection phase state on Player)
	selection := player.Selection().GetPendingCardDrawSelection()
	if selection == nil {
		log.Warn("No pending card draw selection found")
		return fmt.Errorf("no pending card draw selection found")
	}

	// 4. Validate total cards selected
	totalSelected := len(cardsToTake) + len(cardsToBuy)
	maxAllowed := selection.FreeTakeCount + selection.MaxBuyCount

	if totalSelected > maxAllowed {
		log.Warn("Too many cards selected",
			zap.Int("selected", totalSelected),
			zap.Int("max_allowed", maxAllowed))
		return fmt.Errorf("too many cards selected: selected %d, max allowed %d", totalSelected, maxAllowed)
	}

	// 5. Validate free take count
	if len(cardsToTake) > selection.FreeTakeCount {
		log.Warn("Too many free cards selected",
			zap.Int("selected", len(cardsToTake)),
			zap.Int("max", selection.FreeTakeCount))
		return fmt.Errorf("too many free cards selected: selected %d, max %d", len(cardsToTake), selection.FreeTakeCount)
	}

	// 6. For pure card-draw scenarios (all cards must be taken, no choice), require player to take all
	isPureCardDraw := selection.MaxBuyCount == 0 && selection.FreeTakeCount == len(selection.AvailableCards)
	if isPureCardDraw && len(cardsToTake) != selection.FreeTakeCount {
		log.Warn("Must take all cards for pure card-draw effect",
			zap.Int("required", selection.FreeTakeCount),
			zap.Int("selected", len(cardsToTake)))
		return fmt.Errorf("must take all %d cards for card-draw effect", selection.FreeTakeCount)
	}

	// 7. Validate buy count
	if len(cardsToBuy) > selection.MaxBuyCount {
		log.Warn("Too many cards to buy",
			zap.Int("selected", len(cardsToBuy)),
			zap.Int("max", selection.MaxBuyCount))
		return fmt.Errorf("too many cards to buy: selected %d, max %d", len(cardsToBuy), selection.MaxBuyCount)
	}

	// 8. Validate all selected cards are in available cards
	allSelectedCards := append(cardsToTake, cardsToBuy...)
	for _, cardID := range allSelectedCards {
		if !slices.Contains(selection.AvailableCards, cardID) {
			log.Warn("Card not in available cards", zap.String("card_id", cardID))
			return fmt.Errorf("card %s not in available cards", cardID)
		}
	}

	// 9. Calculate total cost for bought cards
	totalCost := len(cardsToBuy) * selection.CardBuyCost

	// 10. Validate player can afford bought cards and deduct credits
	if totalCost > 0 {
		resources := player.Resources().Get()
		if resources.Credits < totalCost {
			log.Warn("Insufficient credits to buy cards",
				zap.Int("needed", totalCost),
				zap.Int("available", resources.Credits))
			return fmt.Errorf("insufficient credits to buy cards: need %d, have %d", totalCost, resources.Credits)
		}

		// Deduct credits for bought cards
		resources.Credits -= totalCost
		player.Resources().Set(resources)

		log.Info("💰 Paid for bought cards",
			zap.Int("cards_bought", len(cardsToBuy)),
			zap.Int("cost", totalCost),
			zap.Int("remaining_credits", resources.Credits))
	}

	// 11. Add all selected cards to player's hand
	for _, cardID := range allSelectedCards {
		player.Hand().AddCard(cardID)
	}

	log.Info("🃏 Added selected cards to hand",
		zap.Int("cards_taken", len(cardsToTake)),
		zap.Int("cards_bought", len(cardsToBuy)),
		zap.Int("total_cards", len(allSelectedCards)))

	// 12. Log discarded cards (they were already popped from deck, so we just don't add them to hand)
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

	// 13. Clear pending card draw selection (card selection phase state on Player)
	player.Selection().SetPendingCardDrawSelection(nil)

	// 14. Publish event - ForcedActionManager will handle if this was a forced action
	events.Publish(a.eventBus, playerevents.CardDrawConfirmedEvent{
		GameID:   gameID,
		PlayerID: playerID,
		Source:   selection.Source,
		Cards:    allSelectedCards,
	})

	log.Debug("📢 Published CardDrawConfirmedEvent",
		zap.String("source", selection.Source),
		zap.Int("card_count", len(allSelectedCards)))

	// 15. Broadcast game state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Card draw confirmation completed",
		zap.String("source", selection.Source),
		zap.Int("cards_taken", len(cardsToTake)),
		zap.Int("cards_bought", len(cardsToBuy)),
		zap.Int("total_cost", totalCost))

	return nil
}
