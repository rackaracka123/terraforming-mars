package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"

	"go.uber.org/zap"
)

// CardRepository manages card data as the single source of truth
type CardRepository interface {
	// LoadCards loads all cards from JSON into memory
	LoadCards(ctx context.Context) error

	// GetCardByID finds a card by its ID
	GetCardByID(ctx context.Context, cardID string) (*model.Card, error)

	// ListCardsByIdMap returns cards matching the given IDs
	ListCardsByIdMap(ctx context.Context, ids map[string]struct{}) (map[string]model.Card, error)

	// GetAllCards returns all loaded cards
	GetAllCards(ctx context.Context) ([]model.Card, error)

	// GetProjectCards returns only project cards (automated, active, event)
	GetProjectCards(ctx context.Context) ([]model.Card, error)

	// GetCorporationCards returns only corporation cards
	GetCorporationCards(ctx context.Context) ([]model.Card, error)

	// GetPreludeCards returns only prelude cards
	GetPreludeCards(ctx context.Context) ([]model.Card, error)

	// GetCardsByType returns cards of a specific type
	GetCardsByType(ctx context.Context, cardType model.CardType) ([]model.Card, error)

	// GetCardsByTag returns cards with a specific tag
	GetCardsByTag(ctx context.Context, tag model.CardTag) ([]model.Card, error)

	// GetStartingCardPool returns cards suitable for starting selection
	GetStartingCardPool(ctx context.Context) ([]model.Card, error)

	// GetCardsByCostRange returns cards within a specific cost range
	GetCardsByCostRange(ctx context.Context, minCost, maxCost int) ([]model.Card, error)

	// GetCardsByTags returns cards that have ANY of the specified tags
	GetCardsByTags(ctx context.Context, tags []model.CardTag) ([]model.Card, error)

	// GetCardsByAllTags returns cards that have ALL of the specified tags
	GetCardsByAllTags(ctx context.Context, tags []model.CardTag) ([]model.Card, error)

	// FilterCardsByRequirements filters cards based on current game state requirements
	FilterCardsByRequirements(ctx context.Context, cards []model.Card, gameState interface{}) ([]model.Card, error)

	// GetCorporations converts corporation cards to Corporation structs
	GetCorporations(ctx context.Context) ([]model.Corporation, error)

	// GetCorporationByID returns a specific corporation by ID
	GetCorporationByID(ctx context.Context, id string) (*model.Corporation, error)
}

// CardRepositoryImpl implements CardRepository
type CardRepositoryImpl struct {
	mutex            sync.RWMutex
	allCards         []model.Card
	projectCards     []model.Card
	corporationCards []model.Card
	preludeCards     []model.Card
	cardLookup       map[string]*model.Card
	loaded           bool
}

// No need for separate JSONCard struct - use model.Card directly

// NewCardRepository creates a new card repository
func NewCardRepository() CardRepository {
	return &CardRepositoryImpl{
		cardLookup: make(map[string]*model.Card),
	}
}

// LoadCards loads all cards from the JSON file into memory
func (r *CardRepositoryImpl) LoadCards(ctx context.Context) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.loaded {
		return nil // Already loaded
	}

	// Get the path to the JSON file - look in multiple possible locations
	possiblePaths := []string{
		filepath.Join("assets", "terraforming_mars_cards.json"),
		filepath.Join("..", "..", "assets", "terraforming_mars_cards.json"),
		filepath.Join("backend", "assets", "terraforming_mars_cards.json"),
		filepath.Join("..", "backend", "assets", "terraforming_mars_cards.json"),
		filepath.Join("..", "..", "..", "assets", "terraforming_mars_cards.json"), // For integration tests
	}

	var data []byte
	var err error

	// Try each path until we find the file
	for _, path := range possiblePaths {
		data, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		return fmt.Errorf("failed to read card data file from any location: %w", err)
	}

	// Parse JSON directly into model.Card array since JSON matches the model exactly
	var allLoadedCards []model.Card
	if err := json.Unmarshal(data, &allLoadedCards); err != nil {
		return fmt.Errorf("failed to parse card data: %w", err)
	}

	log := logger.Get()
	log.Info("📦 Loading cards from JSON",
		zap.Int("total_loaded", len(allLoadedCards)))

	// Initialize slices for all cards
	r.allCards = make([]model.Card, 0)
	r.projectCards = make([]model.Card, 0)
	r.corporationCards = make([]model.Card, 0)
	r.preludeCards = make([]model.Card, 0)

	// Process all loaded cards
	for _, card := range allLoadedCards {
		// Categorize by card type
		switch card.Type {
		case model.CardTypeCorporation:
			r.corporationCards = append(r.corporationCards, card)
		case model.CardTypePrelude:
			r.preludeCards = append(r.preludeCards, card)
		default:
			// All other types are project cards (automated, active, event)
			r.projectCards = append(r.projectCards, card)
		}

		r.allCards = append(r.allCards, card)
		r.cardLookup[card.ID] = &card
	}

	// Log final card counts by type
	log.Info("✅ All cards loaded successfully",
		zap.Int("total_cards", len(r.allCards)),
		zap.Int("project_cards", len(r.projectCards)),
		zap.Int("corporation_cards", len(r.corporationCards)),
		zap.Int("prelude_cards", len(r.preludeCards)))

	r.loaded = true
	return nil
}

// No conversion needed - JSON matches model.Card exactly

// All parsing functions removed - JSON matches model.Card exactly

// GetCardByID finds a card by its ID
func (r *CardRepositoryImpl) GetCardByID(ctx context.Context, cardID string) (*model.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	if card, exists := r.cardLookup[cardID]; exists {
		// Return a copy to prevent external mutation
		cardCopy := *card
		return &cardCopy, nil
	}
	return nil, fmt.Errorf("card not found: %s", cardID)
}

// ListCardsByIdMap returns cards matching the given IDs
func (r *CardRepositoryImpl) ListCardsByIdMap(ctx context.Context, ids map[string]struct{}) (map[string]model.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	result := make(map[string]model.Card)
	for id := range ids {
		if card, exists := r.cardLookup[id]; exists {
			// Return a copy to prevent external mutation
			cardCopy := *card
			result[id] = cardCopy
		}
	}
	return result, nil
}

// GetAllCards returns all loaded cards
func (r *CardRepositoryImpl) GetAllCards(ctx context.Context) ([]model.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	// Return a copy to prevent external mutation
	return r.copyCards(r.allCards), nil
}

// GetProjectCards returns only project cards (automated, active, event)
func (r *CardRepositoryImpl) GetProjectCards(ctx context.Context) ([]model.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	return r.copyCards(r.projectCards), nil
}

// GetCorporationCards returns only corporation cards
func (r *CardRepositoryImpl) GetCorporationCards(ctx context.Context) ([]model.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	return r.copyCards(r.corporationCards), nil
}

// GetPreludeCards returns only prelude cards
func (r *CardRepositoryImpl) GetPreludeCards(ctx context.Context) ([]model.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	return r.copyCards(r.preludeCards), nil
}

// GetCardsByType returns cards of a specific type
func (r *CardRepositoryImpl) GetCardsByType(ctx context.Context, cardType model.CardType) ([]model.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	var cards []model.Card
	for _, card := range r.allCards {
		if card.Type == cardType {
			cards = append(cards, card)
		}
	}
	return cards, nil
}

// GetCardsByTag returns cards with a specific tag
func (r *CardRepositoryImpl) GetCardsByTag(ctx context.Context, tag model.CardTag) ([]model.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	var cards []model.Card
	for _, card := range r.allCards {
		for _, cardTag := range card.Tags {
			if cardTag == tag {
				cards = append(cards, card)
				break
			}
		}
	}
	return cards, nil
}

// GetStartingCardPool returns cards suitable for starting selection
// This includes lower-cost cards that are good for game start
func (r *CardRepositoryImpl) GetStartingCardPool(ctx context.Context) ([]model.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	var startingCards []model.Card

	// Include automated, active, and event cards with reasonable cost (up to 25 MC)
	// This ensures we have enough cards for starting selection
	for _, card := range r.projectCards {
		if card.ID != "" && card.Cost <= 25 &&
			(card.Type == model.CardTypeAutomated || card.Type == model.CardTypeActive || card.Type == model.CardTypeEvent) {
			startingCards = append(startingCards, card)
		}
	}

	return startingCards, nil
}

// GetCardsByCostRange returns cards within a specific cost range
func (r *CardRepositoryImpl) GetCardsByCostRange(ctx context.Context, minCost, maxCost int) ([]model.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	var cards []model.Card
	for _, card := range r.allCards {
		if card.Cost >= minCost && card.Cost <= maxCost {
			cards = append(cards, card)
		}
	}
	return cards, nil
}

// GetCardsByTags returns cards that have ANY of the specified tags
func (r *CardRepositoryImpl) GetCardsByTags(ctx context.Context, tags []model.CardTag) ([]model.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	if len(tags) == 0 {
		return []model.Card{}, nil
	}

	var cards []model.Card
	tagSet := make(map[model.CardTag]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	for _, card := range r.allCards {
		for _, cardTag := range card.Tags {
			if tagSet[cardTag] {
				cards = append(cards, card)
				break // Found at least one matching tag
			}
		}
	}

	return cards, nil
}

// GetCardsByAllTags returns cards that have ALL of the specified tags
func (r *CardRepositoryImpl) GetCardsByAllTags(ctx context.Context, tags []model.CardTag) ([]model.Card, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	if len(tags) == 0 {
		return r.copyCards(r.allCards), nil // If no tags specified, return all cards
	}

	var cards []model.Card

	for _, card := range r.allCards {
		hasAllTags := true
		for _, requiredTag := range tags {
			hasTag := false
			for _, cardTag := range card.Tags {
				if cardTag == requiredTag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				hasAllTags = false
				break
			}
		}

		if hasAllTags {
			cards = append(cards, card)
		}
	}

	return cards, nil
}

// FilterCardsByRequirements filters cards based on current game state requirements
// For now, this is a simplified implementation that just returns all cards
// In a full implementation, you would check temperature, oxygen, oceans, etc.
func (r *CardRepositoryImpl) FilterCardsByRequirements(ctx context.Context, cards []model.Card, gameState interface{}) ([]model.Card, error) {
	// Simplified implementation - just return all cards for now
	// In a full implementation, you would:
	// 1. Cast gameState to the appropriate type
	// 2. Check each card's requirements against current game parameters
	// 3. Check player's production requirements
	// 4. Filter out cards that cannot be played

	var playableCards []model.Card
	for _, card := range cards {
		// For now, include all cards except those with complex requirements
		if len(card.Requirements) == 0 {
			playableCards = append(playableCards, card)
		}
	}

	return playableCards, nil
}

// GetCorporations converts corporation cards to Corporation structs
func (r *CardRepositoryImpl) GetCorporations(ctx context.Context) ([]model.Corporation, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	var corporations []model.Corporation

	for _, card := range r.corporationCards {
		corp := r.convertCardToCorporation(card)
		corporations = append(corporations, corp)
	}

	return corporations, nil
}

// GetCorporationByID returns a specific corporation by ID
func (r *CardRepositoryImpl) GetCorporationByID(ctx context.Context, id string) (*model.Corporation, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if !r.loaded {
		return nil, fmt.Errorf("cards not loaded")
	}

	for _, card := range r.corporationCards {
		if card.ID == id {
			corp := r.convertCardToCorporation(card)
			return &corp, nil
		}
	}

	return nil, fmt.Errorf("corporation not found: %s", id)
}

// convertCardToCorporation converts a Corporation Card to a Corporation struct
func (r *CardRepositoryImpl) convertCardToCorporation(card model.Card) model.Corporation {
	// Use the centralized conversion function from model package
	return model.ConvertCardToCorporation(card)
}

// copyCards creates a deep copy of a slice of cards to prevent external mutation
func (r *CardRepositoryImpl) copyCards(cards []model.Card) []model.Card {
	result := make([]model.Card, len(cards))
	copy(result, cards)
	return result
}
