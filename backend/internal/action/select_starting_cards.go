package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game/card"
	game "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

// SelectStartingCardsAction handles the business logic for selecting starting cards and corporation
type SelectStartingCardsAction struct {
	BaseAction
	gameRepo game.Repository
	cardRepo card.Repository
}

// NewSelectStartingCardsAction creates a new select starting cards action
func NewSelectStartingCardsAction(
	gameRepo game.Repository,
	cardRepo card.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *SelectStartingCardsAction {
	return &SelectStartingCardsAction{
		BaseAction: NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
		cardRepo:   cardRepo,
	}
}

// Execute performs the select starting cards action
func (a *SelectStartingCardsAction) Execute(ctx context.Context, sess *session.Session, playerID string, cardIDs []string, corporationID string) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID).With(
		zap.Strings("card_ids", cardIDs),
		zap.String("corporation_id", corporationID),
	)
	log.Info("🃏 Player selecting starting cards and corporation")

	// 1. Get session and player
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 2. Validate selection phase exists
	if player.SelectStartingCardsPhase == nil {
		log.Error("Player not in starting card selection phase")
		return fmt.Errorf("not in starting card selection phase")
	}

	// Check if player already has a corporation (selection already complete)
	if player.CorporationID != "" {
		log.Error("Starting selection already complete")
		return fmt.Errorf("starting selection already complete")
	}

	// 3. Validate selected cards are in available cards
	availableSet := make(map[string]bool)
	for _, id := range player.SelectStartingCardsPhase.AvailableCards {
		availableSet[id] = true
	}

	for _, cardID := range cardIDs {
		if !availableSet[cardID] {
			log.Error("Selected card not available", zap.String("card_id", cardID))
			return fmt.Errorf("card %s not available for selection", cardID)
		}
	}

	// 4. Validate corporation is in available corporations
	corpAvailable := false
	for _, corpID := range player.SelectStartingCardsPhase.AvailableCorporations {
		if corpID == corporationID {
			corpAvailable = true
			break
		}
	}

	if !corpAvailable {
		log.Error("Selected corporation not available")
		return fmt.Errorf("corporation %s not available", corporationID)
	}

	// 5. Calculate cost (3 MC per card)
	cost := len(cardIDs) * 3

	// 6. Get corporation to apply starting effects
	corp, err := a.cardRepo.GetCardByID(ctx, corporationID)
	if err != nil {
		log.Error("Failed to get corporation", zap.Error(err))
		return fmt.Errorf("failed to get corporation: %w", err)
	}

	// 7. Apply corporation starting resources and production (simplified)
	// In a full implementation, we'd parse corporation effects here
	// For now, just set corporation and give default starting resources
	err = player.Corporation.Set(ctx, corporationID)
	if err != nil {
		log.Error("Failed to set corporation", zap.Error(err))
		return fmt.Errorf("failed to set corporation: %w", err)
	}

	log.Info("✅ Corporation selected", zap.String("corporation_name", corp.Name))

	// 8. Apply default starting resources (typically from corporation)
	// For simplicity, give all players 42 MC to start
	currentResources, err := player.Resources.Get(ctx)
	if err != nil {
		log.Error("Failed to get player resources", zap.Error(err))
		return fmt.Errorf("failed to get resources: %w", err)
	}

	startingResources := currentResources
	startingResources.Credits = 42

	// 9. Deduct card selection cost
	if startingResources.Credits < cost {
		log.Error("Insufficient credits",
			zap.Int("cost", cost),
			zap.Int("available", startingResources.Credits))
		return fmt.Errorf("insufficient credits: need %d, have %d", cost, startingResources.Credits)
	}

	startingResources.Credits -= cost

	err = player.Resources.Update(ctx, startingResources)
	if err != nil {
		log.Error("Failed to update resources", zap.Error(err))
		return fmt.Errorf("failed to update resources: %w", err)
	}

	log.Info("✅ Resources updated",
		zap.Int("cost", cost),
		zap.Int("remaining_credits", startingResources.Credits))

	// 10. Add selected cards to player's hand
	log.Debug("🃏 Adding cards to player hand",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Strings("card_ids", cardIDs),
		zap.Int("count", len(cardIDs)))

	for _, cardID := range cardIDs {
		err = player.Hand.AddCard(ctx, cardID)
		if err != nil {
			log.Error("Failed to add card", zap.String("card_id", cardID), zap.Error(err))
			return fmt.Errorf("failed to add card %s: %w", cardID, err)
		}
	}

	log.Info("✅ Cards added to hand",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.Strings("card_ids_added", cardIDs),
		zap.Int("card_count", len(cardIDs)))

	// 11. Mark selection as complete
	err = player.Selection.CompleteStartingSelection(ctx)
	if err != nil {
		log.Error("Failed to complete selection", zap.Error(err))
		return fmt.Errorf("failed to complete selection: %w", err)
	}

	log.Info("✅ Starting selection marked complete")

	// 12. Check if all players completed selection
	allComplete := true
	for _, p := range sess.GetAllPlayers() {
		if p.CorporationID == "" {
			allComplete = false
			break
		}
	}

	if allComplete {
		log.Info("🎉 All players completed starting selection, advancing to action phase")

		// Advance game phase to Action
		if err := TransitionGamePhase(ctx, a.gameRepo, gameID, game.GamePhaseAction, log); err != nil {
			log.Error("Failed to transition game phase", zap.Error(err))
			// Non-fatal
		}
	}

	// 13. Broadcast state
	a.BroadcastGameState(gameID, log)

	log.Info("🎉 Starting card selection completed successfully")
	return nil
}
