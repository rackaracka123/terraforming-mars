package query

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	sessionCard "terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// GetCorporationsAction handles the query for getting all corporations
type GetCorporationsAction struct {
	action.BaseAction
	cardRepo sessionCard.Repository
}

// NewGetCorporationsAction creates a new get corporations query action
func NewGetCorporationsAction(
	cardRepo sessionCard.Repository,
	sessionFactory session.SessionFactory,
) *GetCorporationsAction {
	return &GetCorporationsAction{
		BaseAction: action.NewBaseAction(sessionFactory, nil),
		cardRepo:   cardRepo,
	}
}

// Execute performs the get corporations query
func (a *GetCorporationsAction) Execute(ctx context.Context) ([]types.Card, error) {
	log := a.GetLogger()
	log.Info("🔍 Querying corporations")

	// Get all corporation cards
	corporations, err := a.cardRepo.GetCorporations(ctx)
	if err != nil {
		log.Error("Failed to get corporations", zap.Error(err))
		return nil, err
	}

	log.Info("✅ Corporations query completed",
		zap.Int("count", len(corporations)))

	return corporations, nil
}
