package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// ConfirmSellPatentsAction handles the business logic for confirming sell patents card selection
// This is Phase 2: processes the selected cards and awards credits
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type ConfirmSellPatentsAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewConfirmSellPatentsAction creates a new confirm sell patents action
func NewConfirmSellPatentsAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *ConfirmSellPatentsAction {
	return &ConfirmSellPatentsAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the confirm sell patents action (Phase 2: process card selection)
func (a *ConfirmSellPatentsAction) Execute(ctx context.Context, gameID string, playerID string, selectedCardIDs []string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "confirm_sell_patents"),
		zap.Int("cards_selected", len(selectedCardIDs)),
	)
	log.Info("🏛️ Confirming sell patents card selection")

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

	// 3. BUSINESS LOGIC: Validate it's the player's turn
	currentTurn := g.CurrentTurn()
	if currentTurn == nil || currentTurn.PlayerID() != playerID {
		var turnPlayerID string
		if currentTurn != nil {
			turnPlayerID = currentTurn.PlayerID()
		}
		log.Warn("Not player's turn",
			zap.String("current_turn_player", turnPlayerID),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("not your turn")
	}

	// 4. Get player from game
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 5. BUSINESS LOGIC: Validate pending card selection exists (card selection phase state on Player)
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

	// 6. BUSINESS LOGIC: Validate selection count
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

	// 7. BUSINESS LOGIC: Validate all selected cards are in available cards
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

	// 8. BUSINESS LOGIC: Calculate total reward (1 M€ per card)
	totalReward := 0
	for _, cardID := range selectedCardIDs {
		totalReward += pendingCardSelection.CardRewards[cardID]
	}

	// 9. BUSINESS LOGIC: Award credits
	if totalReward > 0 {
		player.Resources().Add(map[shared.ResourceType]int{
			shared.ResourceCredits: totalReward,
		})

		resources := player.Resources().Get()
		log.Info("💰 Awarded credits for sold cards",
			zap.Int("cards_sold", len(selectedCardIDs)),
			zap.Int("credits_earned", totalReward),
			zap.Int("new_credits", resources.Credits))
	}

	// 10. BUSINESS LOGIC: Remove sold cards from hand
	for _, cardID := range selectedCardIDs {
		removed := player.Hand().RemoveCard(cardID)
		if !removed {
			log.Warn("Failed to remove card from hand", zap.String("card_id", cardID))
		}
	}

	log.Info("🗑️ Removed sold cards from hand", zap.Int("cards_removed", len(selectedCardIDs)))

	// 11. Clear pending card selection (card selection phase state on Player)
	player.Selection().SetPendingCardSelection(nil)

	// 12. BUSINESS LOGIC: Consume action (only if player actually sold cards and not unlimited actions)
	availableActions := player.Turn().AvailableActions()
	if len(selectedCardIDs) > 0 && availableActions > 0 {
		player.Turn().SetAvailableActions(availableActions - 1)
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions-1))
	}

	// 13. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - player.Resources().Add() publishes ResourcesChangedEvent
	//    - player.Hand().RemoveCard() publishes CardHandUpdatedEvent
	//    - player.Selection().SetPendingCardSelection() publishes events
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("✅ Sell patents completed successfully",
		zap.Int("cards_sold", len(selectedCardIDs)),
		zap.Int("credits_earned", totalReward))
	return nil
}
