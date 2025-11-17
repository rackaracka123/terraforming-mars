package card

import (
	"context"
	"math/rand"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// DeckRepository interface for deck operations (to avoid circular dependency)
type DeckRepository interface {
	GetCardByID(ctx context.Context, cardID string) (*model.Card, error)
	GetAllCards(ctx context.Context) ([]model.Card, error)
	GetProjectCards(ctx context.Context) ([]model.Card, error)
	GetStartingCardPool(ctx context.Context) ([]model.Card, error)
	ListCardsByIdMap(ctx context.Context, ids map[string]struct{}) (map[string]model.Card, error)
}

// Repository manages card data
// Wraps the deck repository for card operations
type Repository interface {
	// DrawProjectCards draws N random project cards
	DrawProjectCards(ctx context.Context, count int) ([]Card, error)

	// DrawCorporations draws N random corporation cards
	DrawCorporations(ctx context.Context, count int) ([]Card, error)

	// GetCardByID retrieves a specific card by ID
	GetCardByID(ctx context.Context, cardID string) (*Card, error)

	// ListCardsByIdMap retrieves multiple cards by their IDs
	ListCardsByIdMap(ctx context.Context, ids map[string]struct{}) (map[string]Card, error)

	// GetStartingCardPool retrieves all starting cards
	GetStartingCardPool(ctx context.Context) ([]Card, error)

	// GetAllCards retrieves all cards
	GetAllCards(ctx context.Context) ([]Card, error)
}

// RepositoryImpl implements the Repository interface by wrapping repositories
type RepositoryImpl struct {
	deckRepo    DeckRepository            // NEW: Primary source for card definitions
	oldCardRepo repository.CardRepository // OLD: Backup card repository
}

// NewRepository creates a new card repository with deck repository
func NewRepository(deckRepo DeckRepository, oldCardRepo repository.CardRepository) Repository {
	return &RepositoryImpl{
		deckRepo:    deckRepo,
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
	// Use NEW deck repository
	mc, err := r.deckRepo.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, err
	}

	card := FromModelCard(*mc)
	return &card, nil
}

// ListCardsByIdMap retrieves multiple cards by their IDs
func (r *RepositoryImpl) ListCardsByIdMap(ctx context.Context, ids map[string]struct{}) (map[string]Card, error) {
	// Use NEW deck repository
	modelCards, err := r.deckRepo.ListCardsByIdMap(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Convert model cards to new card type
	result := make(map[string]Card, len(modelCards))
	for id, mc := range modelCards {
		result[id] = FromModelCard(mc)
	}

	return result, nil
}

// GetStartingCardPool retrieves all starting cards
func (r *RepositoryImpl) GetStartingCardPool(ctx context.Context) ([]Card, error) {
	// Use NEW deck repository
	modelCards, err := r.deckRepo.GetStartingCardPool(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to session card type
	return FromModelCards(modelCards), nil
}

// GetAllCards retrieves all cards
func (r *RepositoryImpl) GetAllCards(ctx context.Context) ([]Card, error) {
	// Use NEW deck repository
	modelCards, err := r.deckRepo.GetAllCards(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to session card type
	return FromModelCards(modelCards), nil
}
