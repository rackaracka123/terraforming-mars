package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

// ConfirmProductionCardsAction handles the business logic for confirming production card selection
type ConfirmProductionCardsAction struct {
	BaseAction
	gameRepo game.Repository
}

// NewConfirmProductionCardsAction creates a new confirm production cards action
func NewConfirmProductionCardsAction(
	gameRepo game.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *ConfirmProductionCardsAction {
	return &ConfirmProductionCardsAction{
		BaseAction: NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the confirm production cards action
func (a *ConfirmProductionCardsAction) Execute(ctx context.Context, sess *session.Session, playerID string, selectedCardIDs []string) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID).With(zap.Strings("selected_card_ids", selectedCardIDs))
	log.Info("🃏 Player confirming production card selection")

	// 1. Get game and validate phase
	g, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Game not found", zap.Error(err))
		return fmt.Errorf("game not found: %w", err)
	}

	if err := ValidateGamePhase(g, game.GamePhaseProductionAndCardDraw, log); err != nil {
		return err
	}

	// 2. Get session and player
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Validate production phase exists (card selection phase state on Player)
	productionPhase := sess.Game().GetProductionPhase(playerID)
	if productionPhase == nil {
		log.Error("Player not in production phase")
		return fmt.Errorf("player not in production phase")
	}

	// 5. Check if player already confirmed selection
	if productionPhase.SelectionComplete {
		log.Error("Production selection already complete")
		return fmt.Errorf("production selection already complete")
	}

	// 6. Validate selected cards are in available cards
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

	// 7. Calculate cost (3 MC per card)
	cost := len(selectedCardIDs) * 3

	// 8. Validate player has enough credits
	resources := player.Resources().Get()
	if resources.Credits < cost {
		log.Error("Insufficient credits",
			zap.Int("cost", cost),
			zap.Int("available", resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", cost, resources.Credits)
	}

	// 9. Deduct card selection cost
	resources.Credits -= cost
	player.Resources().Set(resources)

	log.Info("✅ Resources updated",
		zap.Int("cost", cost),
		zap.Int("remaining_credits", resources.Credits))

	// 10. Add selected cards to player's hand
	log.Debug("🃏 Adding cards to player hand",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Strings("card_ids", selectedCardIDs),
		zap.Int("count", len(selectedCardIDs)))

	for _, cardID := range selectedCardIDs {
		player.Hand().AddCard(cardID)
	}

	log.Info("✅ Cards added to hand",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Strings("card_ids_added", selectedCardIDs),
		zap.Int("card_count", len(selectedCardIDs)))

	// 11. Mark production selection as complete (card selection phase state on Player)
	// Refresh production phase to get updated state
	productionPhase = sess.Game().GetProductionPhase(playerID)
	if productionPhase != nil {
		productionPhase.SelectionComplete = true
		sess.Game().SetProductionPhase(ctx, playerID, productionPhase)
	}

	log.Info("✅ Production selection marked complete")

	// 12. Check if all players completed selection (phase state managed by Game)
	allComplete := true
	for _, p := range sess.GetAllPlayers() {
		pPhase := sess.Game().GetProductionPhase(p.ID())
		if pPhase == nil || !pPhase.SelectionComplete {
			allComplete = false
			break
		}
	}

	if allComplete {
		log.Info("🎉 All players completed production selection, advancing to action phase")

		// Advance game phase to Action
		if err := TransitionGamePhase(ctx, a.gameRepo, gameID, game.GamePhaseAction, log); err != nil {
			log.Error("Failed to transition game phase", zap.Error(err))
			// Non-fatal
		} else {
			// Set current turn to first player
			if len(g.Players) > 0 {
				// Get first player ID from map (note: order not guaranteed - TODO: proper turn order)
				var firstPlayerID string
				for id := range g.Players {
					firstPlayerID = id
					break
				}
				if firstPlayerID != "" {
					if err := SetCurrentTurn(ctx, a.gameRepo, gameID, &firstPlayerID, log); err != nil {
						log.Error("Failed to set current turn", zap.Error(err))
						// Non-fatal
					}
				}
			}

			// Clear production phase data for all players (triggers frontend modal to close)
			for _, p := range sess.GetAllPlayers() {
				sess.Game().SetProductionPhase(ctx, p.ID(), nil)
				log.Debug("✅ Cleared production phase",
					zap.String("player_id", p.ID()))
			}
		}
	}

	// 13. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("🎉 Production card selection completed successfully")
	return nil
}
