package action

import (
	"context"

	"terraforming-mars-backend/internal/service"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// ExecuteCardActionAction handles the business logic for executing card actions
// NOTE: Currently delegates to CardService for complex card action logic
// TODO: Migrate card action logic directly into this action in Phase 7
type ExecuteCardActionAction struct {
	BaseAction
	cardService service.CardService // Temporary delegation to existing service
}

// NewExecuteCardActionAction creates a new execute card action action
func NewExecuteCardActionAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgr session.SessionManager,
	cardService service.CardService,
) *ExecuteCardActionAction {
	return &ExecuteCardActionAction{
		BaseAction:  NewBaseAction(gameRepo, playerRepo, sessionMgr),
		cardService: cardService,
	}
}

// Execute performs the execute card action
func (a *ExecuteCardActionAction) Execute(
	ctx context.Context,
	gameID, playerID, cardID string,
	behaviorIndex int,
	choiceIndex *int,
	cardStorageTarget *string,
) error {
	log := a.InitLogger(gameID, playerID)
	log.Info("🎯 Executing card action",
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex))

	// Delegate to CardService for now (complex card action logic)
	// This will be migrated directly into this action in Phase 7
	err := a.cardService.OnPlayCardAction(ctx, gameID, playerID, cardID, behaviorIndex, choiceIndex, cardStorageTarget)
	if err != nil {
		log.Error("Card action execution failed", zap.Error(err))
		return err
	}

	log.Info("✅ Card action executed successfully")
	return nil
}
