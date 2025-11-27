package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/game/shared"
)

// SelectStartingCardsAction handles the business logic for selecting starting cards and corporation
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type SelectStartingCardsAction struct {
	gameRepo game.GameRepository
	logger   *zap.Logger
}

// NewSelectStartingCardsAction creates a new select starting cards action
func NewSelectStartingCardsAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SelectStartingCardsAction {
	return &SelectStartingCardsAction{
		gameRepo: gameRepo,
		logger:   logger,
	}
}

// Execute performs the select starting cards action
func (a *SelectStartingCardsAction) Execute(ctx context.Context, gameID string, playerID string, cardIDs []string, corporationID string) error {
	log := a.logger.With(
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("action", "select_starting_cards"),
		zap.Strings("card_ids", cardIDs),
		zap.String("corporation_id", corporationID),
	)
	log.Info("🃏 Player selecting starting cards and corporation")

	// 1. Fetch game from repository
	g, err := a.gameRepo.Get(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("game not found: %s", gameID)
	}

	// 2. Get player from game
	player, err := g.GetPlayer(playerID)
	if err != nil {
		log.Error("Player not found in game", zap.Error(err))
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. BUSINESS LOGIC: Validate selection phase exists (phase state managed by Game)
	selectionPhase := g.GetSelectStartingCardsPhase(playerID)
	if selectionPhase == nil {
		log.Error("Player not in starting card selection phase")
		return fmt.Errorf("not in starting card selection phase")
	}

	// 4. BUSINESS LOGIC: Check if player already has a corporation (selection already complete)
	if player.HasCorporation() {
		log.Error("Starting selection already complete")
		return fmt.Errorf("starting selection already complete")
	}

	// 5. BUSINESS LOGIC: Validate selected cards are in available cards
	availableSet := make(map[string]bool)
	for _, id := range selectionPhase.AvailableCards {
		availableSet[id] = true
	}

	for _, cardID := range cardIDs {
		if !availableSet[cardID] {
			log.Error("Selected card not available", zap.String("card_id", cardID))
			return fmt.Errorf("card %s not available for selection", cardID)
		}
	}

	// 6. BUSINESS LOGIC: Validate corporation is in available corporations
	corpAvailable := false
	for _, corpID := range selectionPhase.AvailableCorporations {
		if corpID == corporationID {
			corpAvailable = true
			break
		}
	}

	if !corpAvailable {
		log.Error("Selected corporation not available")
		return fmt.Errorf("corporation %s not available", corporationID)
	}

	// 7. BUSINESS LOGIC: Calculate cost (3 MC per card)
	cost := len(cardIDs) * 3

	// 8. BUSINESS LOGIC: Apply corporation (assumes 42 MC starting credits for simplicity)
	// NOTE: In full implementation, corporation effects would be parsed from card data
	// For now, we just set the corporation and give default starting credits
	player.SetCorporationID(corporationID)

	log.Info("✅ Corporation selected", zap.String("corporation_id", corporationID))

	// 9. BUSINESS LOGIC: Set starting resources and deduct card selection cost
	startingCredits := 42 // Default starting amount
	if startingCredits < cost {
		log.Error("Insufficient credits",
			zap.Int("cost", cost),
			zap.Int("available", startingCredits))
		return fmt.Errorf("insufficient credits: need %d, have %d", cost, startingCredits)
	}

	player.Resources().Add(map[shared.ResourceType]int{
		shared.ResourceCredits: startingCredits - cost,
	})

	resources := player.Resources().Get()
	log.Info("✅ Resources updated",
		zap.Int("cost", cost),
		zap.Int("remaining_credits", resources.Credits))

	// 10. BUSINESS LOGIC: Add selected cards to player's hand
	log.Debug("🃏 Adding cards to player hand",
		zap.Strings("card_ids", cardIDs),
		zap.Int("count", len(cardIDs)))

	for _, cardID := range cardIDs {
		player.Hand().AddCard(cardID)
	}

	log.Info("✅ Cards added to hand",
		zap.Strings("card_ids_added", cardIDs),
		zap.Int("card_count", len(cardIDs)))

	// 11. BUSINESS LOGIC: Mark selection as complete (phase state managed by Game)
	if err := g.SetSelectStartingCardsPhase(ctx, playerID, nil); err != nil {
		log.Error("Failed to clear starting cards phase", zap.Error(err))
		return fmt.Errorf("failed to clear starting cards phase: %w", err)
	}

	log.Info("✅ Starting selection marked complete")

	// 12. BUSINESS LOGIC: Check if all players completed selection
	allPlayers := g.GetAllPlayers()
	allComplete := true
	for _, p := range allPlayers {
		if !p.HasCorporation() {
			allComplete = false
			break
		}
	}

	if allComplete {
		log.Info("🎉 All players completed starting selection, advancing to action phase")

		// Advance game phase to Action
		if err := g.UpdatePhase(ctx, game.GamePhaseAction); err != nil {
			log.Error("Failed to transition game phase", zap.Error(err))
			return fmt.Errorf("failed to transition game phase: %w", err)
		}

		// Set current turn to first player
		if len(allPlayers) > 0 {
			firstPlayerID := allPlayers[0].ID()
			if err := g.SetCurrentTurn(ctx, firstPlayerID); err != nil {
				log.Error("Failed to set current turn", zap.Error(err))
				return fmt.Errorf("failed to set current turn: %w", err)
			}
		}
	}

	// 13. NO MANUAL BROADCAST - BroadcastEvent automatically triggered by:
	//    - player.SetCorporationID() publishes events
	//    - player.Resources().Add() publishes ResourcesChangedEvent
	//    - player.Hand().AddCard() publishes CardHandUpdatedEvent
	//    - g.SetSelectStartingCardsPhase() publishes BroadcastEvent
	//    - g.UpdatePhase() publishes GamePhaseChangedEvent + BroadcastEvent
	//    - g.SetCurrentTurn() publishes BroadcastEvent
	//    Broadcaster subscribes to BroadcastEvent and handles WebSocket updates

	log.Info("🎉 Starting card selection completed successfully")
	return nil
}
