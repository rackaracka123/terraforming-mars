package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/game/player"

	"go.uber.org/zap"
)

// ConfirmProductionCardsAction handles the business logic for confirming production card selection
type ConfirmProductionCardsAction struct {
	BaseAction
	cardRepo card.Repository
}

// NewConfirmProductionCardsAction creates a new confirm production cards action
func NewConfirmProductionCardsAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	cardRepo card.Repository,
	sessionMgr session.SessionManager,
) *ConfirmProductionCardsAction {
	return &ConfirmProductionCardsAction{
		BaseAction: NewBaseAction(gameRepo, playerRepo, sessionMgr),
		cardRepo:   cardRepo,
	}
}

// Execute performs the confirm production cards action
func (a *ConfirmProductionCardsAction) Execute(ctx context.Context, gameID string, playerID string, selectedCardIDs []string) error {
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

	// 2. Validate player exists
	p, err := ValidatePlayer(ctx, a.playerRepo, gameID, playerID, log)
	if err != nil {
		return err
	}

	// 4. Validate production phase exists
	if p.ProductionPhase == nil {
		log.Error("Player not in production phase")
		return fmt.Errorf("player not in production phase")
	}

	// 5. Check if player already confirmed selection
	if p.ProductionPhase.SelectionComplete {
		log.Error("Production selection already complete")
		return fmt.Errorf("production selection already complete")
	}

	// 6. Validate selected cards are in available cards
	availableSet := make(map[string]bool)
	for _, id := range p.ProductionPhase.AvailableCards {
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
	if p.Resources.Credits < cost {
		log.Error("Insufficient credits",
			zap.Int("cost", cost),
			zap.Int("available", p.Resources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", cost, p.Resources.Credits)
	}

	// 9. Deduct card selection cost
	updatedResources := p.Resources
	updatedResources.Credits -= cost

	err = a.playerRepo.UpdateResources(ctx, gameID, playerID, updatedResources)
	if err != nil {
		log.Error("Failed to update resources", zap.Error(err))
		return fmt.Errorf("failed to update resources: %w", err)
	}

	log.Info("✅ Resources updated",
		zap.Int("cost", cost),
		zap.Int("remaining_credits", updatedResources.Credits))

	// 10. Add selected cards to player's hand
	log.Debug("🃏 Adding cards to player hand",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Strings("card_ids", selectedCardIDs),
		zap.Int("count", len(selectedCardIDs)))

	for _, cardID := range selectedCardIDs {
		err = a.playerRepo.AddCard(ctx, gameID, playerID, cardID)
		if err != nil {
			log.Error("Failed to add card", zap.String("card_id", cardID), zap.Error(err))
			return fmt.Errorf("failed to add card %s: %w", cardID, err)
		}
	}

	log.Info("✅ Cards added to hand",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Strings("card_ids_added", selectedCardIDs),
		zap.Int("card_count", len(selectedCardIDs)))

	// 11. Mark production selection as complete
	err = a.playerRepo.CompleteProductionSelection(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to complete production selection", zap.Error(err))
		return fmt.Errorf("failed to complete production selection: %w", err)
	}

	log.Info("✅ Production selection marked complete")

	// 12. Check if all players completed selection
	allComplete, err := CheckAllPlayersComplete(ctx, a.playerRepo, gameID, func(p *player.Player) bool {
		return p.ProductionPhase != nil && p.ProductionPhase.SelectionComplete
	})
	if err != nil {
		log.Error("Failed to check completion status", zap.Error(err))
		// Non-fatal, continue
	} else if allComplete {
		log.Info("🎉 All players completed production selection, advancing to action phase")

		// Advance game phase to Action
		if err := TransitionGamePhase(ctx, a.gameRepo, gameID, game.GamePhaseAction, log); err != nil {
			log.Error("Failed to transition game phase", zap.Error(err))
			// Non-fatal
		} else {
			// Set current turn to first player
			if len(g.PlayerIDs) > 0 {
				firstPlayer := g.PlayerIDs[0]
				if err := SetCurrentTurn(ctx, a.gameRepo, gameID, &firstPlayer, log); err != nil {
					log.Error("Failed to set current turn", zap.Error(err))
					// Non-fatal
				}
			}

			// Clear production phase data for all players (triggers frontend modal to close)
			players, err := GetAllPlayers(ctx, a.playerRepo, gameID, log)
			if err == nil {
				for _, p := range players {
					err = a.playerRepo.UpdateProductionPhase(ctx, gameID, p.ID, nil)
					if err != nil {
						log.Error("Failed to clear production phase",
							zap.String("player_id", p.ID),
							zap.Error(err))
					} else {
						log.Debug("✅ Cleared production phase",
							zap.String("player_id", p.ID))
					}
				}
			} else {
				log.Error("Failed to list players for production phase cleanup", zap.Error(err))
			}
		}
	}

	// 13. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("🎉 Production card selection completed successfully")
	return nil
}
