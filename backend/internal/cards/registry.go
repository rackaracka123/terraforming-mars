package cards

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"
)

// Registry handles card-specific logic and delegates to specialized handlers
type Registry struct {
	validator       *Validator
	effectProcessor *EffectProcessor
	cardRepo        repository.CardRepository
}

// NewRegistry creates a new card registry
func NewRegistry(cardRepo repository.CardRepository, gameRepo repository.GameRepository, playerRepo repository.PlayerRepository) *Registry {
	return &Registry{
		validator:       NewValidator(),
		effectProcessor: NewEffectProcessor(gameRepo, playerRepo),
		cardRepo:        cardRepo,
	}
}

// ValidateCardPlay validates if a card can be played
func (r *Registry) ValidateCardPlay(card *model.Card, game *model.Game, player *model.Player) error {
	return r.validator.ValidateRequirements(card, game, player)
}

// ApplyCardEffects applies the effects of a played card
func (r *Registry) ApplyCardEffects(ctx context.Context, gameID, playerID string, card *model.Card) error {
	return r.effectProcessor.ApplyCardEffects(ctx, gameID, playerID, card)
}

// GetCardByID retrieves a card by its ID
func (r *Registry) GetCardByID(ctx context.Context, cardID string) (*model.Card, error) {
	card, err := r.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card %s: %w", cardID, err)
	}
	if card == nil {
		return nil, fmt.Errorf("card %s not found", cardID)
	}
	return card, nil
}

// GetAllCards retrieves all available cards
func (r *Registry) GetAllCards(ctx context.Context) ([]model.Card, error) {
	return r.cardRepo.GetAllCards(ctx)
}

// Future expansion points:
// - Card-specific handlers (e.g., GreatDamHandler, BirdsHandler)
// - Action card logic
// - Complex card interactions
// - Card synergy calculations
// - Tournament/competitive rule variations
