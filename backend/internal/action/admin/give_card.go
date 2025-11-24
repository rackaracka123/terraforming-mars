package admin

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/card"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// GiveCardAction handles the admin action to give a card to a player
type GiveCardAction struct {
	action.BaseAction
	cardRepo card.Repository
}

// NewGiveCardAction creates a new give card admin action
func NewGiveCardAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	cardRepo card.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *GiveCardAction {
	return &GiveCardAction{
		BaseAction: action.NewBaseAction(gameRepo, playerRepo, sessionMgrFactory),
		cardRepo:   cardRepo,
	}
}

// Execute performs the give card admin action
func (a *GiveCardAction) Execute(ctx context.Context, gameID, playerID, cardID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🎴 Admin: Giving card to player",
		zap.String("card_id", cardID))

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.GetGameRepo(), gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate player exists
	_, err = action.ValidatePlayer(ctx, a.GetPlayerRepo(), gameID, playerID, log)
	if err != nil {
		return err
	}

	// 3. Validate card exists
	_, err = a.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		log.Error("Card not found", zap.Error(err))
		return fmt.Errorf("card not found: %w", err)
	}

	// 4. Add card to player's hand
	err = a.GetPlayerRepo().AddCard(ctx, gameID, playerID, cardID)
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
