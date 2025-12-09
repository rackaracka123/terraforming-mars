package query

import (
	"context"

	"terraforming-mars-backend/internal/cards"
	gamecards "terraforming-mars-backend/internal/game/cards"

	"go.uber.org/zap"
)

// ListCardsResult represents the result of listing cards
type ListCardsResult struct {
	Cards      []gamecards.Card
	TotalCount int
	Offset     int
	Limit      int
}

// ListCardsAction handles querying all cards with pagination
type ListCardsAction struct {
	cardRegistry cards.CardRegistry
	logger       *zap.Logger
}

// NewListCardsAction creates a new list cards query action
func NewListCardsAction(
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *ListCardsAction {
	return &ListCardsAction{
		cardRegistry: cardRegistry,
		logger:       logger,
	}
}

// Execute retrieves cards with pagination
func (a *ListCardsAction) Execute(ctx context.Context, offset, limit int) (*ListCardsResult, error) {
	log := a.logger.With(
		zap.Int("offset", offset),
		zap.Int("limit", limit),
	)
	log.Info("🔍 Querying cards")

	// Get all cards from registry
	allCards := a.cardRegistry.GetAll()
	totalCount := len(allCards)

	// Apply pagination
	start := offset
	if start > totalCount {
		start = totalCount
	}

	end := start + limit
	if end > totalCount {
		end = totalCount
	}

	paginatedCards := allCards[start:end]

	log.Info("✅ Cards query completed",
		zap.Int("total_count", totalCount),
		zap.Int("returned_count", len(paginatedCards)),
	)

	return &ListCardsResult{
		Cards:      paginatedCards,
		TotalCount: totalCount,
		Offset:     offset,
		Limit:      limit,
	}, nil
}
