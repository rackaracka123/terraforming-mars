package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"terraforming-mars-backend/internal/model"
)

// CardDataService handles loading and managing card data from JSON
type CardDataService interface {
	LoadCards() error
	GetAllCards() []model.Card
	GetProjectCards() []model.Card
	GetCorporationCards() []model.Card
	GetPreludeCards() []model.Card
	GetCardByID(id string) (*model.Card, error)
	GetCardsByType(cardType model.CardType) []model.Card
	GetCardsByTag(tag model.CardTag) []model.Card
	GetStartingCardPool() []model.Card // Returns cards suitable for starting selection
	GetCardsByCostRange(minCost, maxCost int) []model.Card
	GetCardsByTags(tags []model.CardTag) []model.Card                                 // Cards that have ANY of the specified tags
	GetCardsByAllTags(tags []model.CardTag) []model.Card                              // Cards that have ALL of the specified tags
	FilterCardsByRequirements(cards []model.Card, gameState interface{}) []model.Card // Filter cards by requirements
	GetCorporations() []model.Corporation                                             // Returns corporations converted from corporation cards
	GetCorporationByID(id string) (*model.Corporation, error)                         // Returns a specific corporation
}

// CardDataServiceImpl implements CardDataService
type CardDataServiceImpl struct {
	allCards         []model.Card
	projectCards     []model.Card
	corporationCards []model.Card
	preludeCards     []model.Card
	cardLookup       map[string]*model.Card
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
	Name              string                 `json:"name"`
	Cost              int                    `json:"cost"`
	Number            string                 `json:"number"`
	Type              string                 `json:"type"`
	Tags              []string               `json:"tags"`
	Requirements      map[string]interface{} `json:"requirements,omitempty"`
	VictoryPoints     JSONVictoryPoints      `json:"victory_points,omitempty"`
	Description       string                 `json:"description,omitempty"`
	ProductionEffects JSONProductionEffects  `json:"production_effects,omitempty"`
	ImmediateEffects  JSONImmediateEffects   `json:"immediate_effects,omitempty"`
	Action            map[string]interface{} `json:"action,omitempty"`
}

// JSONVictoryPoints represents victory point structure in JSON
type JSONVictoryPoints struct {
	BasePoints        int                      `json:"base_points"`
	ConditionalPoints []map[string]interface{} `json:"conditional_points"`
}

// JSONProductionEffects represents production effects structure in JSON
type JSONProductionEffects struct {
	Increase map[string]int `json:"increase"`
	Decrease map[string]int `json:"decrease"`
	Choices  []interface{}  `json:"choices"`
}

// JSONImmediateEffects represents immediate effects structure in JSON
type JSONImmediateEffects struct {
	ResourceGains    map[string]int `json:"resource_gains"`
	ResourceCosts    map[string]int `json:"resource_costs"`
	TileEffects      []string       `json:"tile_effects"`
	ParameterChanges map[string]int `json:"parameter_changes"`
	OtherEffects     []string       `json:"other_effects"`
}

// NewCardDataService creates a new CardDataService instance
func NewCardDataService() CardDataService {
	return &CardDataServiceImpl{
		cardLookup: make(map[string]*model.Card),
	}
}

// LoadCards loads all cards from the JSON file
func (s *CardDataServiceImpl) LoadCards() error {
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

	// Convert and store cards
	s.allCards = make([]model.Card, 0)
	s.projectCards = make([]model.Card, 0)
	s.corporationCards = make([]model.Card, 0)
	s.preludeCards = make([]model.Card, 0)

	// Process project cards
	for _, jsonCard := range jsonData.Cards.ProjectCards {
		card, err := s.convertJSONCard(jsonCard)
		if err != nil {
			return fmt.Errorf("failed to convert project card %s: %w", jsonCard.Name, err)
		}
		s.projectCards = append(s.projectCards, card)
		s.allCards = append(s.allCards, card)
		s.cardLookup[card.ID] = &card
	}

	// Process corporation cards
	for _, jsonCard := range jsonData.Cards.CorporationCards {
		card, err := s.convertJSONCard(jsonCard)
		if err != nil {
			return fmt.Errorf("failed to convert corporation card %s: %w", jsonCard.Name, err)
		}
		card.Type = model.CardTypeCorporation
		s.corporationCards = append(s.corporationCards, card)
		s.allCards = append(s.allCards, card)
		s.cardLookup[card.ID] = &card
	}

	// Process prelude cards
	for _, jsonCard := range jsonData.Cards.PreludeCards {
		card, err := s.convertJSONCard(jsonCard)
		if err != nil {
			return fmt.Errorf("failed to convert prelude card %s: %w", jsonCard.Name, err)
		}
		card.Type = model.CardTypePrelude
		s.preludeCards = append(s.preludeCards, card)
		s.allCards = append(s.allCards, card)
		s.cardLookup[card.ID] = &card
	}

	return nil
}

// convertJSONCard converts a JSONCard to a model.Card
func (s *CardDataServiceImpl) convertJSONCard(jsonCard JSONCard) (model.Card, error) {
	// Generate ID from card number or name
	id := s.generateCardID(jsonCard.Number, jsonCard.Name)

	// Convert type
	cardType, err := s.convertCardType(jsonCard.Type)
	if err != nil {
		return model.Card{}, err
	}

	// Convert tags
	tags := s.convertTags(jsonCard.Tags)

	// Parse requirements from enhanced JSON structure
	requirements := s.parseEnhancedRequirements(jsonCard.Requirements)

	// Parse production effects from enhanced JSON structure
	var productionEffects *model.ProductionEffects
	if jsonCard.ProductionEffects.Increase != nil || jsonCard.ProductionEffects.Decrease != nil {
		effects := &model.ProductionEffects{}
		// Convert production effects - simplified for now
		for resource, amount := range jsonCard.ProductionEffects.Increase {
			switch resource {
			case "credits":
				effects.Credits += amount
			case "steel":
				effects.Steel += amount
			case "titanium":
				effects.Titanium += amount
			case "plants":
				effects.Plants += amount
			case "energy":
				effects.Energy += amount
			case "heat":
				effects.Heat += amount
			}
		}
		// Handle decreases as negative values
		for resource, amount := range jsonCard.ProductionEffects.Decrease {
			switch resource {
			case "credits":
				effects.Credits -= amount
			case "steel":
				effects.Steel -= amount
			case "titanium":
				effects.Titanium -= amount
			case "plants":
				effects.Plants -= amount
			case "energy":
				effects.Energy -= amount
			case "heat":
				effects.Heat -= amount
			}
		}
		productionEffects = effects
	}

	card := model.Card{
		ID:                id,
		Name:              jsonCard.Name,
		Type:              cardType,
		Cost:              jsonCard.Cost,
		Description:       jsonCard.Description,
		Tags:              tags,
		Requirements:      requirements,
		VictoryPoints:     jsonCard.VictoryPoints.BasePoints,
		Number:            jsonCard.Number,
		ProductionEffects: productionEffects,
	}

	return card, nil
}

// generateCardID creates a unique ID from card number or name
func (s *CardDataServiceImpl) generateCardID(number, name string) string {
	if number != "" {
		// Remove # and convert to lowercase
		id := strings.ToLower(strings.TrimPrefix(number, "#"))
		// Replace non-alphanumeric with underscore
		re := regexp.MustCompile(`[^a-z0-9]+`)
		return re.ReplaceAllString(id, "_")
	}

	// Fall back to name-based ID
	id := strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	return re.ReplaceAllString(id, "_")
}

// convertCardType converts string type to model.CardType
func (s *CardDataServiceImpl) convertCardType(typeStr string) (model.CardType, error) {
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
func (s *CardDataServiceImpl) convertTags(tagStrs []string) []model.CardTag {
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

// parseEnhancedRequirements converts enhanced JSON requirements to CardRequirements struct
func (s *CardDataServiceImpl) parseEnhancedRequirements(reqMap map[string]interface{}) model.CardRequirements {
	requirements := model.CardRequirements{}

	if reqMap == nil {
		return requirements
	}

	// Parse structured requirements
	if val, ok := reqMap["min_oxygen"]; ok {
		if oxygen, ok := val.(float64); ok {
			oxygenInt := int(oxygen)
			requirements.MinOxygen = &oxygenInt
		}
	}

	if val, ok := reqMap["max_oxygen"]; ok {
		if oxygen, ok := val.(float64); ok {
			oxygenInt := int(oxygen)
			requirements.MaxOxygen = &oxygenInt
		}
	}

	if val, ok := reqMap["min_temperature"]; ok {
		if temp, ok := val.(float64); ok {
			tempInt := int(temp)
			requirements.MinTemperature = &tempInt
		}
	}

	if val, ok := reqMap["max_temperature"]; ok {
		if temp, ok := val.(float64); ok {
			tempInt := int(temp)
			requirements.MaxTemperature = &tempInt
		}
	}

	if val, ok := reqMap["min_oceans"]; ok {
		if oceans, ok := val.(float64); ok {
			oceansInt := int(oceans)
			requirements.MinOceans = &oceansInt
		}
	}

	if val, ok := reqMap["max_oceans"]; ok {
		if oceans, ok := val.(float64); ok {
			oceansInt := int(oceans)
			requirements.MaxOceans = &oceansInt
		}
	}

	return requirements
}

// parseRequirements converts requirement string to CardRequirements struct (legacy)
func (s *CardDataServiceImpl) parseRequirements(reqStr string) (model.CardRequirements, error) {
	requirements := model.CardRequirements{}

	if reqStr == "" {
		return requirements, nil
	}

	// Parse oxygen requirements (e.g., "max 5% O2", "6% O2")
	oxygenRegex := regexp.MustCompile(`(?i)(max\s+)?(\d+)%\s*o2`)
	if matches := oxygenRegex.FindStringSubmatch(reqStr); len(matches) > 0 {
		oxygen, _ := strconv.Atoi(matches[2])
		if strings.Contains(strings.ToLower(matches[0]), "max") {
			requirements.MaxOxygen = &oxygen
		} else {
			requirements.MinOxygen = &oxygen
		}
	}

	// Parse temperature requirements (e.g., "max -14°C", "-6°C")
	tempRegex := regexp.MustCompile(`(?i)(max\s+)?(-?\d+)°?c`)
	if matches := tempRegex.FindStringSubmatch(reqStr); len(matches) > 0 {
		temp, _ := strconv.Atoi(matches[2])
		if strings.Contains(strings.ToLower(matches[0]), "max") {
			requirements.MaxTemperature = &temp
		} else {
			requirements.MinTemperature = &temp
		}
	}

	// Parse ocean requirements (e.g., "3 Oceans", "max 3 oceans")
	oceanRegex := regexp.MustCompile(`(?i)(max\s+)?(\d+)\s*oceans?`)
	if matches := oceanRegex.FindStringSubmatch(reqStr); len(matches) > 0 {
		oceans, _ := strconv.Atoi(matches[2])
		if strings.Contains(strings.ToLower(matches[0]), "max") {
			requirements.MaxOceans = &oceans
		} else {
			requirements.MinOceans = &oceans
		}
	}

	// Parse production requirements (e.g., "Titanium production")
	if strings.Contains(strings.ToLower(reqStr), "production") {
		// Simplified - in full implementation you'd parse specific production types
		requirements.RequiredProduction = &model.ResourceSet{}
	}

	return requirements, nil
}

// GetAllCards returns all loaded cards
func (s *CardDataServiceImpl) GetAllCards() []model.Card {
	return s.allCards
}

// GetProjectCards returns only project cards (automated, active, event)
func (s *CardDataServiceImpl) GetProjectCards() []model.Card {
	return s.projectCards
}

// GetCorporationCards returns only corporation cards
func (s *CardDataServiceImpl) GetCorporationCards() []model.Card {
	return s.corporationCards
}

// GetPreludeCards returns only prelude cards
func (s *CardDataServiceImpl) GetPreludeCards() []model.Card {
	return s.preludeCards
}

// GetCardByID finds a card by its ID
func (s *CardDataServiceImpl) GetCardByID(id string) (*model.Card, error) {
	if card, exists := s.cardLookup[id]; exists {
		return card, nil
	}
	return nil, fmt.Errorf("card not found: %s", id)
}

// GetCardsByType returns cards of a specific type
func (s *CardDataServiceImpl) GetCardsByType(cardType model.CardType) []model.Card {
	var cards []model.Card
	for _, card := range s.allCards {
		if card.Type == cardType {
			cards = append(cards, card)
		}
	}
	return cards
}

// GetCardsByTag returns cards with a specific tag
func (s *CardDataServiceImpl) GetCardsByTag(tag model.CardTag) []model.Card {
	var cards []model.Card
	for _, card := range s.allCards {
		for _, cardTag := range card.Tags {
			if cardTag == tag {
				cards = append(cards, card)
				break
			}
		}
	}
	return cards
}

// GetStartingCardPool returns cards suitable for starting selection
// This includes lower-cost cards that are good for game start
func (s *CardDataServiceImpl) GetStartingCardPool() []model.Card {
	var startingCards []model.Card

	// Include automated and active cards with reasonable cost (up to 15 MC)
	for _, card := range s.projectCards {
		if card.ID != "" && card.Cost <= 15 && (card.Type == model.CardTypeAutomated || card.Type == model.CardTypeActive) {
			startingCards = append(startingCards, card)
		}
	}

	return startingCards
}

// GetCardsByCostRange returns cards within a specific cost range
func (s *CardDataServiceImpl) GetCardsByCostRange(minCost, maxCost int) []model.Card {
	var cards []model.Card
	for _, card := range s.allCards {
		if card.Cost >= minCost && card.Cost <= maxCost {
			cards = append(cards, card)
		}
	}
	return cards
}

// GetCardsByTags returns cards that have ANY of the specified tags
func (s *CardDataServiceImpl) GetCardsByTags(tags []model.CardTag) []model.Card {
	if len(tags) == 0 {
		return []model.Card{}
	}

	var cards []model.Card
	tagSet := make(map[model.CardTag]bool)
	for _, tag := range tags {
		tagSet[tag] = true
	}

	for _, card := range s.allCards {
		for _, cardTag := range card.Tags {
			if tagSet[cardTag] {
				cards = append(cards, card)
				break // Found at least one matching tag
			}
		}
	}

	return cards
}

// GetCardsByAllTags returns cards that have ALL of the specified tags
func (s *CardDataServiceImpl) GetCardsByAllTags(tags []model.CardTag) []model.Card {
	if len(tags) == 0 {
		return s.allCards // If no tags specified, return all cards
	}

	var cards []model.Card

	for _, card := range s.allCards {
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

	return cards
}

// FilterCardsByRequirements filters cards based on current game state requirements
// For now, this is a simplified implementation that just returns all cards
// In a full implementation, you would check temperature, oxygen, oceans, etc.
func (s *CardDataServiceImpl) FilterCardsByRequirements(cards []model.Card, gameState interface{}) []model.Card {
	// Simplified implementation - just return all cards for now
	// In a full implementation, you would:
	// 1. Cast gameState to the appropriate type
	// 2. Check each card's requirements against current game parameters
	// 3. Check player's production requirements
	// 4. Filter out cards that cannot be played

	var playableCards []model.Card
	for _, card := range cards {
		// For now, include all cards except those with complex requirements
		if card.Requirements.MinTemperature == nil &&
			card.Requirements.MaxTemperature == nil &&
			card.Requirements.MinOxygen == nil &&
			card.Requirements.MaxOxygen == nil &&
			card.Requirements.MinOceans == nil &&
			card.Requirements.MaxOceans == nil &&
			card.Requirements.RequiredProduction == nil {
			playableCards = append(playableCards, card)
		}
	}

	return playableCards
}

// GetCorporations converts corporation cards to Corporation structs
func (s *CardDataServiceImpl) GetCorporations() []model.Corporation {
	var corporations []model.Corporation

	for _, card := range s.corporationCards {
		corp := s.convertCardToCorporation(card)
		corporations = append(corporations, corp)
	}

	return corporations
}

// GetCorporationByID returns a specific corporation by ID
func (s *CardDataServiceImpl) GetCorporationByID(id string) (*model.Corporation, error) {
	for _, card := range s.corporationCards {
		if card.ID == id {
			corp := s.convertCardToCorporation(card)
			return &corp, nil
		}
	}

	return nil, fmt.Errorf("corporation not found: %s", id)
}

// convertCardToCorporation converts a Corporation Card to a Corporation struct
func (s *CardDataServiceImpl) convertCardToCorporation(card model.Card) model.Corporation {
	corp := model.Corporation{
		ID:                 card.ID,
		Name:               card.Name,
		Description:        card.Description,
		StartingCredits:    42, // Default starting credits
		StartingResources:  model.ResourceSet{Credits: 42},
		StartingProduction: model.ResourceSet{},
		Tags:               card.Tags,
		SpecialEffects:     []string{card.Description},
		Number:             card.Number,
	}

	// Parse corporation-specific starting conditions from JSON production effects
	if card.ProductionEffects != nil {
		corp.StartingProduction = model.ResourceSet{
			Credits:  card.ProductionEffects.Credits,
			Steel:    card.ProductionEffects.Steel,
			Titanium: card.ProductionEffects.Titanium,
			Plants:   card.ProductionEffects.Plants,
			Energy:   card.ProductionEffects.Energy,
			Heat:     card.ProductionEffects.Heat,
		}
	}

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
