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

// SetCorporationAction handles the admin action to set a player's corporation
type SetCorporationAction struct {
	action.BaseAction
	gameRepo       game.Repository
	cardRepo       card.Repository
	sessionFactory session.SessionFactory
}

// NewSetCorporationAction creates a new set corporation admin action
func NewSetCorporationAction(
	gameRepo game.Repository,
	cardRepo card.Repository,
	sessionMgrFactory session.SessionManagerFactory,
	sessionFactory session.SessionFactory,
) *SetCorporationAction {
	return &SetCorporationAction{
		BaseAction:     action.NewBaseAction(sessionMgrFactory),
		gameRepo:       gameRepo,
		cardRepo:       cardRepo,
		sessionFactory: sessionFactory,
	}
}

// Execute performs the set corporation admin action
func (a *SetCorporationAction) Execute(ctx context.Context, gameID, playerID, corporationID string) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🏢 Admin: Setting player corporation",
		zap.String("corporation_id", corporationID))

	// 1. Validate game exists
	_, err := action.ValidateGameExists(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Get session and player
	sess := a.sessionFactory.Get(gameID)
	if sess == nil {
		log.Error("Session not found")
		return fmt.Errorf("game session not found: %s", gameID)
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 3. Get corporation card
	corp, err := a.cardRepo.GetCardByID(ctx, corporationID)
	if err != nil {
		log.Error("Corporation card not found", zap.Error(err))
		return fmt.Errorf("corporation not found: %w", err)
	}

	// 4. Update player corporation
	player.Corp().SetCard(*corp)

	log.Info("✅ Player corporation updated")

	// 5. Broadcast updated game state
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Admin set corporation completed")
	return nil
}
