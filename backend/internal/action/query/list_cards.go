package query

import (
	"context"

	"terraforming-mars-backend/internal/action"
	sessionCard "terraforming-mars-backend/internal/session/game/card"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// ListCardsAction handles the query for listing cards with pagination
type ListCardsAction struct {
	action.BaseAction
	cardRepo sessionCard.Repository
}

// NewListCardsAction creates a new list cards query action
func NewListCardsAction(
	cardRepo sessionCard.Repository,
) *ListCardsAction {
	return &ListCardsAction{
		BaseAction: action.NewBaseAction(nil),
		cardRepo:   cardRepo,
	}
}

// Execute performs the list cards query
func (a *ListCardsAction) Execute(ctx context.Context, offset, limit int) ([]types.Card, int, error) {
	log := a.GetLogger()
	log.Info("🔍 Querying cards",
		zap.Int("offset", offset),
		zap.Int("limit", limit))

	// Get all cards
	allCards, err := a.cardRepo.GetAllCards(ctx)
	if err != nil {
		log.Error("Failed to get cards", zap.Error(err))
		return nil, 0, err
	}

	totalCount := len(allCards)

	// Apply pagination
	start := offset
	if start > totalCount {
		start = totalCount
	}

	end := offset + limit
	if end > totalCount {
		end = totalCount
	}

	paginatedCards := allCards[start:end]

	log.Info("✅ Cards query completed",
		zap.Int("total", totalCount),
		zap.Int("returned", len(paginatedCards)))

	return paginatedCards, totalCount, nil
}
