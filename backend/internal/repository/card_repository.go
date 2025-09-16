package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"terraforming-mars-backend/internal/model"
	"time"
)

// CardRepository manages card data as the single source of truth
type CardRepository interface {
	// LoadCards loads all cards from JSON into memory
	LoadCards(ctx context.Context) error

	// GetCardByID finds a card by its ID
	GetCardByID(ctx context.Context, cardID string) (*model.Card, error)

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

// JSONCardData represents the structure of the JSON file
type JSONCardData struct {
	Metadata struct {
		Source                string `json:"source"`
		ExtractionDate        string `json:"extraction_date"`
		TotalProjectCards     int    `json:"total_project_cards"`
		TotalCorporationCards int    `json:"total_corporation_cards"`
		TotalPreludeCards     int    `json:"total_prelude_cards"`
		Description           string `json:"description"`
	} `json:"metadata"`
	Cards struct {
		ProjectCards     []JSONCard `json:"project_cards"`
		CorporationCards []JSONCard `json:"corporation_cards"`
		PreludeCards     []JSONCard `json:"prelude_cards"`
	} `json:"cards"`
}

// JSONCard represents a card in the enhanced JSON format
type JSONCard struct {
	Name         string      `json:"name"`
	Cost         int         `json:"cost"`
	Number       string      `json:"number"`
	Type         string      `json:"type"`
	Tags         []string    `json:"tags"`
	Requirements interface{} `json:"requirements,omitempty"`
	Description  string      `json:"description,omitempty"`
}

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

	// Parse JSON
	var jsonData JSONCardData
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("failed to parse card data: %w", err)
	}

	// Initialize slices
	r.allCards = make([]model.Card, 0)
	r.projectCards = make([]model.Card, 0)
	r.corporationCards = make([]model.Card, 0)
	r.preludeCards = make([]model.Card, 0)

	// Process project cards
	for _, jsonCard := range jsonData.Cards.ProjectCards {
		card, err := r.convertJSONCard(jsonCard)
		if err != nil {
			return fmt.Errorf("failed to convert project card %s: %w", jsonCard.Name, err)
		}
		r.projectCards = append(r.projectCards, card)
		r.allCards = append(r.allCards, card)
		r.cardLookup[card.ID] = &card
	}

	// Process corporation cards
	for _, jsonCard := range jsonData.Cards.CorporationCards {
		card, err := r.convertJSONCard(jsonCard)
		if err != nil {
			return fmt.Errorf("failed to convert corporation card %s: %w", jsonCard.Name, err)
		}
		card.Type = model.CardTypeCorporation
		r.corporationCards = append(r.corporationCards, card)
		r.allCards = append(r.allCards, card)
		r.cardLookup[card.ID] = &card
	}

	// Process prelude cards
	for _, jsonCard := range jsonData.Cards.PreludeCards {
		card, err := r.convertJSONCard(jsonCard)
		if err != nil {
			return fmt.Errorf("failed to convert prelude card %s: %w", jsonCard.Name, err)
		}
		card.Type = model.CardTypePrelude
		r.preludeCards = append(r.preludeCards, card)
		r.allCards = append(r.allCards, card)
		r.cardLookup[card.ID] = &card
	}

	r.loaded = true
	return nil
}

// convertJSONCard converts a JSONCard to a model.Card
func (r *CardRepositoryImpl) convertJSONCard(jsonCard JSONCard) (model.Card, error) {
	// Generate ID from card number or name
	id := r.generateCardID(jsonCard.Number, jsonCard.Name)

	// Convert type
	cardType, err := r.convertCardType(jsonCard.Type)
	if err != nil {
		return model.Card{}, err
	}

	// Convert tags
	tags := r.convertTags(jsonCard.Tags)

	// Parse requirements from enhanced JSON structure
	requirements := r.parseEnhancedRequirements(jsonCard.Requirements)

	card := model.Card{
		ID:           id,
		Name:         jsonCard.Name,
		Type:         cardType,
		Cost:         jsonCard.Cost,
		Description:  jsonCard.Description,
		Tags:         tags,
		Requirements: requirements,
	}

	return card, nil
}

// generateCardID creates a unique ID from card number or name
func (r *CardRepositoryImpl) generateCardID(number, name string) string {
	var generatedID string

	if number != "" {
		// Remove # and convert to lowercase
		id := strings.ToLower(strings.TrimPrefix(number, "#"))
		// Replace non-alphanumeric with underscore
		re := regexp.MustCompile(`[^a-z0-9]+`)
		generatedID = re.ReplaceAllString(id, "_")
	} else {
		// Fall back to name-based ID
		id := strings.ToLower(name)
		re := regexp.MustCompile(`[^a-z0-9]+`)
		generatedID = re.ReplaceAllString(id, "_")
	}

	// Clean up leading/trailing underscores and ensure we have a valid ID
	generatedID = strings.Trim(generatedID, "_")

	// If the generated ID is still empty, create a fallback
	if generatedID == "" {
		generatedID = fmt.Sprintf("card_%s_%s", strings.ReplaceAll(number, "#", ""), strings.ReplaceAll(name, " ", "_"))
		// If still empty, use timestamp-based ID
		if generatedID == "card__" {
			generatedID = fmt.Sprintf("card_%d", time.Now().UnixNano())
		}
	}

	return generatedID
}

// convertCardType converts string type to model.CardType
func (r *CardRepositoryImpl) convertCardType(typeStr string) (model.CardType, error) {
	switch typeStr {
	case "automated":
		return model.CardTypeAutomated, nil
	case "active":
		return model.CardTypeActive, nil
	case "event":
		return model.CardTypeEvent, nil
	case "corporation":
		return model.CardTypeCorporation, nil
	case "prelude":
		return model.CardTypePrelude, nil
	case "unknown", "":
		// Default unknown types to automated for now
		return model.CardTypeAutomated, nil
	default:
		// Log warning and default to automated
		return model.CardTypeAutomated, fmt.Errorf("unrecognized card type '%s', defaulting to automated", typeStr)
	}
}

// convertTags converts string slice to CardTag slice
func (r *CardRepositoryImpl) convertTags(tagStrs []string) []model.CardTag {
	tags := make([]model.CardTag, 0, len(tagStrs))

	tagMapping := map[string]model.CardTag{
		"space":    model.TagSpace,
		"earth":    model.TagEarth,
		"science":  model.TagScience,
		"power":    model.TagPower,
		"building": model.TagBuilding,
		"microbe":  model.TagMicrobe,
		"animal":   model.TagAnimal,
		"plant":    model.TagPlant,
		"event":    model.TagEvent,
		"city":     model.TagCity,
		"venus":    model.TagVenus,
		"jovian":   model.TagJovian,
		"wild":     model.TagWild,
	}

	for _, tagStr := range tagStrs {
		if tag, exists := tagMapping[tagStr]; exists {
			tags = append(tags, tag)
		}
	}

	return tags
}

// parseEnhancedRequirements converts enhanced JSON requirements to []Requirement
func (r *CardRepositoryImpl) parseEnhancedRequirements(reqArray interface{}) []model.Requirement {
	var requirements []model.Requirement

	if reqArray == nil {
		return requirements
	}

	// Handle array of requirement objects
	if reqSlice, ok := reqArray.([]interface{}); ok {
		for _, reqItem := range reqSlice {
			if reqMap, ok := reqItem.(map[string]interface{}); ok {
				requirement := model.Requirement{}

				// Parse type
				if reqType, ok := reqMap["type"].(string); ok {
					requirement.Type = model.RequirementType(reqType)
				}

				// Parse min value
				if minVal, ok := reqMap["min"]; ok {
					if minFloat, ok := minVal.(float64); ok {
						minInt := int(minFloat)
						requirement.Min = &minInt
					}
				}

				// Parse max value
				if maxVal, ok := reqMap["max"]; ok {
					if maxFloat, ok := maxVal.(float64); ok {
						maxInt := int(maxFloat)
						requirement.Max = &maxInt
					}
				}

				// Parse location
				if location, ok := reqMap["location"].(string); ok {
					loc := model.Location(location)
					requirement.Location = &loc
				}

				requirements = append(requirements, requirement)
			}
		}
	}

	return requirements
}

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

	// Include automated and active cards with reasonable cost (up to 15 MC)
	for _, card := range r.projectCards {
		if card.ID != "" && card.Cost <= 15 && (card.Type == model.CardTypeAutomated || card.Type == model.CardTypeActive) {
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
	corp := model.Corporation{
		ID:                 card.ID,
		Name:               card.Name,
		Description:        card.Description,
		StartingCredits:    42, // Default starting credits
		StartingResources:  model.ResourceSet{Credits: 42},
		StartingProduction: model.ResourceSet{},
		Tags:               card.Tags,
		SpecialEffects:     []string{card.Description},
	}

	// Corporation-specific starting conditions would be parsed from behaviors
	// For now, using default values

	// TODO: Parse starting resources and credits from JSON immediate effects
	// For now, we'll use some hardcoded logic based on well-known corporations
	// This should be replaced with proper JSON parsing once the JSON structure
	// includes starting bonuses for corporations

	switch corp.Name {
	case "Credicor":
		corp.StartingCredits = 57
		corp.StartingResources.Credits = 57
	case "Ecoline":
		corp.StartingCredits = 36
		corp.StartingResources.Credits = 36
		corp.StartingResources.Plants = 3
		corp.StartingProduction.Plants = 2
	case "Helion":
		corp.StartingCredits = 42
		corp.StartingResources.Credits = 42
		corp.StartingResources.Heat = 3
		corp.StartingProduction.Heat = 3
	case "Mining Guild":
		corp.StartingCredits = 30
		corp.StartingResources.Credits = 30
		corp.StartingResources.Steel = 5
		corp.StartingProduction.Steel = 1
	}

	return corp
}

// copyCards creates a deep copy of a slice of cards to prevent external mutation
func (r *CardRepositoryImpl) copyCards(cards []model.Card) []model.Card {
	result := make([]model.Card, len(cards))
	copy(result, cards)
	return result
}
