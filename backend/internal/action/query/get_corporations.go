package query

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// GetCorporationsAction handles the query for getting all corporations
type GetCorporationsAction struct {
	action.BaseAction
	cardRepo repository.CardRepository
}

// NewGetCorporationsAction creates a new get corporations query action
func NewGetCorporationsAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	cardRepo repository.CardRepository,
	sessionMgr session.SessionManager,
) *GetCorporationsAction {
	return &GetCorporationsAction{
		BaseAction: action.NewBaseAction(gameRepo, playerRepo, sessionMgr),
		cardRepo:   cardRepo,
	}
}

// Execute performs the get corporations query
func (a *GetCorporationsAction) Execute(ctx context.Context) ([]model.Card, error) {
	log := a.GetLogger()
	log.Info("🔍 Querying corporations")

	// Get all corporation cards
	corporations, err := a.cardRepo.GetCorporationCards(ctx)
	if err != nil {
		log.Error("Failed to get corporations", zap.Error(err))
		return nil, err
	}

	log.Info("✅ Corporations query completed",
		zap.Int("count", len(corporations)))

	return corporations, nil
}
