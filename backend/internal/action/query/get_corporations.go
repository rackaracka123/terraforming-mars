package query

import (
	"context"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session/game/card"

	"go.uber.org/zap"
)

// GetCorporationsAction handles the query for getting all corporations
type GetCorporationsAction struct {
	action.BaseAction
	cardRepo card.Repository
}

// NewGetCorporationsAction creates a new get corporations query action
func NewGetCorporationsAction(
	cardRepo card.Repository,
) *GetCorporationsAction {
	return &GetCorporationsAction{
		BaseAction: action.NewBaseAction(nil),
		cardRepo:   cardRepo,
	}
}

// Execute performs the get corporations query
func (a *GetCorporationsAction) Execute(ctx context.Context) ([]card.Card, error) {
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
