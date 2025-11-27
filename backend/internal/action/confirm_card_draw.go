package action

import (
	"context"
	"fmt"
	"slices"

	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"

	"go.uber.org/zap"
)

// ConfirmCardDrawAction handles the business logic for confirming card draw selection
// MIGRATION: Uses new architecture (GameRepository + EventBus for ForcedActionManager integration)
// NOTE: EventBus dependency required to publish CardDrawConfirmedEvent for ForcedActionManager
type ConfirmCardDrawAction struct {
	gameRepo game.GameRepository
	eventBus *events.EventBusImpl
	logger   *zap.Logger
}

// NewConfirmCardDrawAction creates a new confirm card draw action
func NewConfirmCardDrawAction(
	gameRepo game.GameRepository,
	eventBus *events.EventBusImpl,
	logger *zap.Logger,
) *ConfirmCardDrawAction {
	return &ConfirmCardDrawAction{
		gameRepo: gameRepo,
		eventBus: eventBus,
		logger:   logger,
	}
}

// Execute performs the confirm card draw action
func (a *ConfirmCardDrawAction) Execute(ctx context.Context, gameID string, playerID string, cardsToTake []string, cardsToBuy []string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "confirm_card_draw"),
		zap.Int("cards_to_take", len(cardsToTake)),
		zap.Int("cards_to_buy", len(cardsToBuy)),
	)
	log.Info("🃏 Confirming card draw selection")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. BUSINESS LOGIC: Validate game is active
	if g.Status() != game.GameStatusActive {
		log.Warn("Game is not active", zap.String("status", string(g.Status())))
		return fmt.Errorf("game is not active: %s", g.Status())
	}

	// 3. Get player from game
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. BUSINESS LOGIC: Validate pending card draw selection exists (card selection state on Player)
	selection := player.Selection().GetPendingCardDrawSelection()
	if selection == nil {
		log.Warn("No pending card draw selection found")
		return fmt.Errorf("no pending card draw selection found")
	}

	// 5. BUSINESS LOGIC: Validate total cards selected
	totalSelected := len(cardsToTake) + len(cardsToBuy)
	maxAllowed := selection.FreeTakeCount + selection.MaxBuyCount

	if totalSelected > maxAllowed {
		log.Warn("Too many cards selected",
			zap.Int("selected", totalSelected),
			zap.Int("max_allowed", maxAllowed))
		return fmt.Errorf("too many cards selected: selected %d, max allowed %d", totalSelected, maxAllowed)
	}

	// 6. BUSINESS LOGIC: Validate free take count
	if len(cardsToTake) > selection.FreeTakeCount {
		log.Warn("Too many free cards selected",
			zap.Int("selected", len(cardsToTake)),
			zap.Int("max", selection.FreeTakeCount))
		return fmt.Errorf("too many free cards selected: selected %d, max %d", len(cardsToTake), selection.FreeTakeCount)
	}

	// 7. BUSINESS LOGIC: For pure card-draw scenarios (all cards must be taken, no choice), require player to take all
	isPureCardDraw := selection.MaxBuyCount == 0 && selection.FreeTakeCount == len(selection.AvailableCards)
	if isPureCardDraw && len(cardsToTake) != selection.FreeTakeCount {
		log.Warn("Must take all cards for pure card-draw effect",
			zap.Int("required", selection.FreeTakeCount),
			zap.Int("selected", len(cardsToTake)))
		return fmt.Errorf("must take all %d cards for card-draw effect", selection.FreeTakeCount)
	}

	// 8. BUSINESS LOGIC: Validate buy count
	if len(cardsToBuy) > selection.MaxBuyCount {
		log.Warn("Too many cards to buy",
			zap.Int("selected", len(cardsToBuy)),
			zap.Int("max", selection.MaxBuyCount))
		return fmt.Errorf("too many cards to buy: selected %d, max %d", len(cardsToBuy), selection.MaxBuyCount)
	}

	// 9. BUSINESS LOGIC: Validate all selected cards are in available cards
	allSelectedCards := append(cardsToTake, cardsToBuy...)
	for _, cardID := range allSelectedCards {
		if !slices.Contains(selection.AvailableCards, cardID) {
			log.Warn("Card not in available cards", zap.String("card_id", cardID))
			return fmt.Errorf("card %s not in available cards", cardID)
		}
	}

	// 10. BUSINESS LOGIC: Calculate total cost for bought cards
	totalCost := len(cardsToBuy) * selection.CardBuyCost

	// 11. BUSINESS LOGIC: Validate player can afford bought cards and deduct credits
	if totalCost > 0 {
		resources := player.Resources().Get()
		if resources.Credits < totalCost {
			log.Warn("Insufficient credits to buy cards",
				zap.Int("needed", totalCost),
				zap.Int("available", resources.Credits))
			return fmt.Errorf("insufficient credits to buy cards: need %d, have %d", totalCost, resources.Credits)
		}

		// Deduct credits for bought cards
		player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredits: -totalCost,
		})

		newResources := player.Resources().Get()
		log.Info("💰 Paid for bought cards",
			zap.Int("cards_bought", len(cardsToBuy)),
			zap.Int("cost", totalCost),
			zap.Int("remaining_credits", newResources.Credits))
	}

	// 12. BUSINESS LOGIC: Add all selected cards to player's hand
	for _, cardID := range allSelectedCards {
		player.Hand().AddCard(cardID)
	}

	log.Info("🃏 Added selected cards to hand",
		zap.Int("cards_taken", len(cardsToTake)),
		zap.Int("cards_bought", len(cardsToBuy)),
		zap.Int("total_cards", len(allSelectedCards)))

	// 13. Log discarded cards (they were already popped from deck, so we just don't add them to hand)
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

	// 14. Clear pending card draw selection (card selection state on Player)
	player.Selection().SetPendingCardDrawSelection(nil)

	// 15. Event publishing - domain events automatically published by repository updates
	// NOTE: CardDrawConfirmedEvent not yet migrated to new architecture
	// BroadcastEvent will be triggered by repository updates above
	_ = a.eventBus // Keep eventBus for future event integration

	// 16. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - player.Resources().Add() publishes ResourcesChangedEvent
	//    - player.Hand().AddCard() publishes CardHandUpdatedEvent
	//    - player.Selection().SetPendingCardDrawSelection() publishes events
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("✅ Card draw confirmation completed",
		zap.String("source", selection.Source),
		zap.Int("cards_taken", len(cardsToTake)),
		zap.Int("cards_bought", len(cardsToBuy)),
		zap.Int("total_cost", totalCost))

	return nil
}
