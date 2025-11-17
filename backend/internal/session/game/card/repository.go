package card

import (
	"context"
	"math/rand"

	"terraforming-mars-backend/internal/repository"
)

// Repository manages card data
// For Phase 2-3, this wraps the old card repository
type Repository interface {
	// DrawProjectCards draws N random project cards
	DrawProjectCards(ctx context.Context, count int) ([]Card, error)

	// DrawCorporations draws N random corporation cards
	DrawCorporations(ctx context.Context, count int) ([]Card, error)

	// GetCardByID retrieves a specific card by ID
	GetCardByID(ctx context.Context, cardID string) (*Card, error)
}

// RepositoryImpl implements the Repository interface by wrapping the old card repository
type RepositoryImpl struct {
	oldCardRepo repository.CardRepository
}

// NewRepository creates a new card repository
func NewRepository(oldCardRepo repository.CardRepository) Repository {
	return &RepositoryImpl{
		oldCardRepo: oldCardRepo,
	}
}

// DrawProjectCards draws N random project cards
func (r *RepositoryImpl) DrawProjectCards(ctx context.Context, count int) ([]Card, error) {
	// Get all project cards from old repository
	projectCards, err := r.oldCardRepo.GetProjectCards(ctx)
	if err != nil {
		return nil, err
	}

	// Shuffle and take N cards
	shuffled := make([]Card, len(projectCards))
	for i, mc := range projectCards {
		shuffled[i] = FromModelCard(mc)
	}

	// Fisher-Yates shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	// Take first N cards
	if count > len(shuffled) {
		count = len(shuffled)
	}

	return shuffled[:count], nil
}

// DrawCorporations draws N random corporation cards
func (r *RepositoryImpl) DrawCorporations(ctx context.Context, count int) ([]Card, error) {
	// Get all corporation cards from old repository
	corpCards, err := r.oldCardRepo.GetCorporationCards(ctx)
	if err != nil {
		return nil, err
	}

	// Shuffle and take N cards
	shuffled := make([]Card, len(corpCards))
	for i, mc := range corpCards {
		shuffled[i] = FromModelCard(mc)
	}

	// Fisher-Yates shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	// Take first N cards
	if count > len(shuffled) {
		count = len(shuffled)
	}

	return shuffled[:count], nil
}

// GetCardByID retrieves a specific card by ID
func (r *RepositoryImpl) GetCardByID(ctx context.Context, cardID string) (*Card, error) {
	mc, err := r.oldCardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	card := FromModelCard(*mc)
	return &card, nil
}
