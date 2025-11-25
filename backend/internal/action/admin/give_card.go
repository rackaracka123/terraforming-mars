package admin

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game/card"
	game "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

// GiveCardAction handles the admin action to give a card to a player
type GiveCardAction struct {
	action.BaseAction
	gameRepo game.Repository
	cardRepo card.Repository
}

// NewGiveCardAction creates a new give card admin action
func NewGiveCardAction(
	gameRepo game.Repository,
	cardRepo card.Repository,
	sessionFactory session.SessionFactory,
	sessionMgrFactory session.SessionManagerFactory,
) *GiveCardAction {
	return &GiveCardAction{
		BaseAction: action.NewBaseAction(sessionFactory, sessionMgrFactory),
		gameRepo:   gameRepo,
		cardRepo:   cardRepo,
	}
}

// Execute performs the give card admin action
func (a *GiveCardAction) Execute(ctx context.Context, gameID, playerID, cardID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🎴 Admin: Giving card to player",
		zap.String("card_id", cardID))

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Get session and player
	sess := a.GetSessionFactory().Get(gameID)
	if sess == nil {
		log.Error("Game session not found")
		return fmt.Errorf("game not found: %s", gameID)
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. Validate card exists
	_, err = a.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		log.Error("Card not found", zap.Error(err))
		return fmt.Errorf("card not found: %w", err)
	}

	// 4. Add card to player's hand
	err = player.Hand.AddCard(ctx, cardID)
	if err != nil {
		log.Error("Failed to add card to hand", zap.Error(err))
		return fmt.Errorf("failed to add card: %w", err)
	}

	log.Info("✅ Card added to player's hand")

	// 5. Broadcast updated game state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Admin give card completed")
	return nil
}
