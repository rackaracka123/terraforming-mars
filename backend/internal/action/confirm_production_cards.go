package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// ConfirmProductionCardsAction handles the business logic for confirming production card selection
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type ConfirmProductionCardsAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewConfirmProductionCardsAction creates a new confirm production cards action
func NewConfirmProductionCardsAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *ConfirmProductionCardsAction {
	return &ConfirmProductionCardsAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the confirm production cards action
func (a *ConfirmProductionCardsAction) Execute(ctx context.Context, gameID string, playerID string, selectedCardIDs []string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "confirm_production_cards"),
		zap.Strings("selected_card_ids", selectedCardIDs),
	)
	log.Info("🃏 Player confirming production card selection")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. BUSINESS LOGIC: Validate game phase
	if g.CurrentPhase() != game.GamePhaseProductionAndCardDraw {
		log.Warn("Game is not in production phase",
			zap.String("current_phase", string(g.CurrentPhase())),
			zap.String("expected_phase", string(game.GamePhaseProductionAndCardDraw)))
		return fmt.Errorf("game is not in production phase")
	}

	// 3. Get player from game
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. BUSINESS LOGIC: Validate production phase exists (phase state managed by Game)
	productionPhase := g.GetProductionPhase(playerID)
	if productionPhase == nil {
		log.Error("Player not in production phase")
		return fmt.Errorf("player not in production phase")
	}

	// 5. BUSINESS LOGIC: Check if player already confirmed selection
	if productionPhase.SelectionComplete {
		log.Error("Production selection already complete")
		return fmt.Errorf("production selection already complete")
	}

	// 6. BUSINESS LOGIC: Validate selected cards are in available cards
	availableSet := make(map[string]bool)
	for _, id := range productionPhase.AvailableCards {
		availableSet[id] = true
	}

	for _, cardID := range selectedCardIDs {
		if !availableSet[cardID] {
			log.Error("Selected card not available", zap.String("card_id", cardID))
			return fmt.Errorf("card %s not available for selection", cardID)
		}
	}

	// 7. BUSINESS LOGIC: Calculate cost (3 MC per card)
	cost := len(selectedCardIDs) * 3

	// 8. BUSINESS LOGIC: Validate player has enough credits
	resources := player.Resources().Get()
	if resources.Credits < cost {
		log.Error("Insufficient credits",
			zap.Int("cost", cost),
			zap.Int("available", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", cost, resources.Credits)
	}

	// 9. BUSINESS LOGIC: Deduct card selection cost
	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: -cost,
	})

	resources = player.Resources().Get() // Refresh after update
	log.Info("✅ Resources updated",
		zap.Int("cost", cost),
		zap.Int("remaining_credits", resources.Credits))

	// 10. BUSINESS LOGIC: Add selected cards to player's hand
	log.Debug("🃏 Adding cards to player hand",
		zap.Strings("card_ids", selectedCardIDs),
		zap.Int("count", len(selectedCardIDs)))

	for _, cardID := range selectedCardIDs {
		player.Hand().AddCard(cardID)
	}

	log.Info("✅ Cards added to hand",
		zap.Strings("card_ids_added", selectedCardIDs),
		zap.Int("card_count", len(selectedCardIDs)))

	// 11. BUSINESS LOGIC: Mark production selection as complete (phase state managed by Game)
	productionPhase.SelectionComplete = true
	if err := g.SetProductionPhase(ctx, playerID, productionPhase); err != nil {
		log.Error("Failed to update production phase", zap.Error(err))
		return fmt.Errorf("failed to update production phase: %w", err)
	}

	log.Info("✅ Production selection marked complete")

	// 12. BUSINESS LOGIC: Check if all players completed selection
	allPlayers := g.GetAllPlayers()
	allComplete := true
	for _, p := range allPlayers {
		pPhase := g.GetProductionPhase(p.ID())
		if pPhase == nil || !pPhase.SelectionComplete {
			allComplete = false
			break
		}
	}

	if allComplete {
		log.Info("🎉 All players completed production selection, advancing to action phase")

		// Advance game phase to Action
		if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
			log.Error("Failed to transition game phase", zap.Error(err))
			return fmt.Errorf("failed to transition game phase: %w", err)
		}

		// Set current turn to first player
		if len(allPlayers) > 0 {
			firstPlayerID := allPlayers[0].ID()
			if err := g.SetCurrentTurn(ctx, firstPlayerID, []game.ActionType{}); err != nil {
				log.Error("Failed to set current turn", zap.Error(err))
				return fmt.Errorf("failed to set current turn: %w", err)
			}
		}

		// Clear production phase data for all players (triggers frontend modal to close)
		for _, p := range allPlayers {
			if err := g.SetProductionPhase(ctx, p.ID(), nil); err != nil {
				log.Warn("Failed to clear production phase",
					zap.String("player_id", p.ID()),
					zap.Error(err))
			}
		}
	}

	// 13. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - player.Resources().Add() publishes ResourcesChangedEvent
	//    - player.Hand().AddCard() publishes CardHandUpdatedEvent
	//    - g.SetProductionPhase() publishes BroadcastEvent
	//    - g.UpdatePhase() publishes GamePhaseChangedEvent + BroadcastEvent
	//    - g.SetCurrentTurn() publishes BroadcastEvent
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("🎉 Production card selection completed successfully")
	return nil
}
