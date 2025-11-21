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

// ListCardsAction handles the query for listing cards with pagination
type ListCardsAction struct {
	action.BaseAction
	cardRepo repository.CardRepository
}

// NewListCardsAction creates a new list cards query action
func NewListCardsAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	cardRepo repository.CardRepository,
	sessionMgr session.SessionManager,
) *ListCardsAction {
	return &ListCardsAction{
		BaseAction: action.NewBaseAction(gameRepo, playerRepo, sessionMgr),
		cardRepo:   cardRepo,
	}
}

// Execute performs the list cards query
func (a *ListCardsAction) Execute(ctx context.Context, offset, limit int) ([]model.Card, int, error) {
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
