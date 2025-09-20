package cards

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"terraforming-mars-backend/internal/model"
)

// CardRegistry holds all card data loaded from JSON
type CardRegistry struct {
	Cards        map[string]*model.Card `json:"cards" ts:"Record<string, Card>"`
	Corporations map[string]*model.Card `json:"corporations" ts:"Record<string, Card>"`
	StartingDeck []string               `json:"startingDeck" ts:"string[]"`
}

// NewCardRegistry creates a new empty card registry
func NewCardRegistry() *CardRegistry {
	return &CardRegistry{
		Cards:        make(map[string]*model.Card),
		Corporations: make(map[string]*model.Card),
		StartingDeck: make([]string, 0),
	}
}

// DeepCopy creates a deep copy of the card registry
func (cr *CardRegistry) DeepCopy() *CardRegistry {
	newRegistry := &CardRegistry{
		Cards:        make(map[string]*model.Card),
		Corporations: make(map[string]*model.Card),
		StartingDeck: make([]string, len(cr.StartingDeck)),
	}

	// Copy all cards
	for id, card := range cr.Cards {
		newRegistry.Cards[id] = card.DeepCopy()
	}

	// Copy corporations
	for id, card := range cr.Corporations {
		newRegistry.Corporations[id] = card.DeepCopy()
	}

	// Copy starting deck
	copy(newRegistry.StartingDeck, cr.StartingDeck)

	return newRegistry
}

// GetCard returns a card by ID
func (cr *CardRegistry) GetCard(id string) (*model.Card, bool) {
	card, exists := cr.Cards[id]
	return card, exists
}

// GetCorporation returns a corporation card by ID
func (cr *CardRegistry) GetCorporation(id string) (*model.Card, bool) {
	card, exists := cr.Corporations[id]
	return card, exists
}

// GetStartingCardPool returns all cards available for starting selection
func (cr *CardRegistry) GetStartingCardPool() []*model.Card {
	cards := make([]*model.Card, 0, len(cr.StartingDeck))
	for _, cardID := range cr.StartingDeck {
		if card, exists := cr.Cards[cardID]; exists {
			cards = append(cards, card)
		}
	}
	return cards
}

// LoadCardsFromFile loads cards from the terraforming_mars_cards.json file
func LoadCardsFromFile(filePath string) (*CardRegistry, error) {
	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cards file %s: %w", filePath, err)
	}

	// Parse JSON into intermediate structures
	var jsonCards []jsonCard
	if err := json.Unmarshal(data, &jsonCards); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %w", filePath, err)
	}

	// Create registry and convert JSON cards to model cards
	registry := NewCardRegistry()
	for _, jsonCard := range jsonCards {
		card, err := convertJSONCard(jsonCard)
		if err != nil {
			return nil, fmt.Errorf("failed to convert card %s: %w", jsonCard.ID, err)
		}

		// Add to main cards collection
		registry.Cards[card.ID] = card

		// Add to corporations collection if it's a corporation
		if card.Type == model.CardTypeCorporation {
			registry.Corporations[card.ID] = card
		} else {
			// Non-corporation cards are available for starting selection
			registry.StartingDeck = append(registry.StartingDeck, card.ID)
		}
	}

	return registry, nil
}

// LoadCardsFromAssets loads cards from the default assets location
func LoadCardsFromAssets() (*CardRegistry, error) {
	// Get the path relative to the backend directory
	assetsPath := filepath.Join("assets", "terraforming_mars_cards.json")
	return LoadCardsFromFile(assetsPath)
}

// jsonCard represents the JSON structure for a card
type jsonCard struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Cost         int               `json:"cost"`
	Description  string            `json:"description"`
	Tags         []string          `json:"tags,omitempty"`
	Requirements []jsonRequirement `json:"requirements,omitempty"`
	Behaviors    []jsonBehavior    `json:"behaviors,omitempty"`
	VPConditions []jsonVPCondition `json:"vpConditions,omitempty"`
}

// jsonRequirement represents the JSON structure for a requirement
type jsonRequirement struct {
	Type     string  `json:"type"`
	Min      *int    `json:"min,omitempty"`
	Max      *int    `json:"max,omitempty"`
	Location *string `json:"location,omitempty"`
	Tag      *string `json:"tag,omitempty"`
	Resource *string `json:"resource,omitempty"`
}

// jsonBehavior represents the JSON structure for a card behavior
type jsonBehavior struct {
	Triggers []jsonTrigger           `json:"triggers,omitempty"`
	Inputs   []jsonResourceCondition `json:"inputs,omitempty"`
	Outputs  []jsonResourceCondition `json:"outputs,omitempty"`
	Choices  []jsonChoice            `json:"choices,omitempty"`
}

// jsonTrigger represents the JSON structure for a trigger
type jsonTrigger struct {
	Type      string                `json:"type"`
	Condition *jsonTriggerCondition `json:"condition,omitempty"`
}

// jsonTriggerCondition represents the JSON structure for trigger conditions
type jsonTriggerCondition struct {
	Type     string  `json:"type"`
	Location *string `json:"location,omitempty"`
}

// jsonResourceCondition represents the JSON structure for resource conditions
type jsonResourceCondition struct {
	Type         string            `json:"type"`
	Amount       int               `json:"amount"`
	Target       string            `json:"target"`
	AffectedTags []string          `json:"affectedTags,omitempty"`
	Per          *jsonPerCondition `json:"per,omitempty"`
}

// jsonChoice represents the JSON structure for a choice
type jsonChoice struct {
	Inputs  []jsonResourceCondition `json:"inputs,omitempty"`
	Outputs []jsonResourceCondition `json:"outputs,omitempty"`
}

// jsonPerCondition represents the JSON structure for per conditions
type jsonPerCondition struct {
	Type     string  `json:"type"`
	Amount   int     `json:"amount"`
	Location *string `json:"location,omitempty"`
	Target   *string `json:"target,omitempty"`
	Tag      *string `json:"tag,omitempty"`
}

// jsonVPCondition represents the JSON structure for VP conditions
type jsonVPCondition struct {
	Amount     int               `json:"amount"`
	Condition  string            `json:"condition"`
	MaxTrigger *int              `json:"maxTrigger,omitempty"`
	Per        *jsonPerCondition `json:"per,omitempty"`
}

// convertJSONCard converts a JSON card to a model card
func convertJSONCard(jCard jsonCard) (*model.Card, error) {
	card := &model.Card{
		ID:          jCard.ID,
		Name:        jCard.Name,
		Cost:        jCard.Cost,
		Description: jCard.Description,
	}

	// Convert card type
	switch jCard.Type {
	case "corporation":
		card.Type = model.CardTypeCorporation
	case "automated":
		card.Type = model.CardTypeAutomated
	case "active":
		card.Type = model.CardTypeActive
	case "event":
		card.Type = model.CardTypeEvent
	case "prelude":
		card.Type = model.CardTypePrelude
	default:
		return nil, fmt.Errorf("unknown card type: %s", jCard.Type)
	}

	// Convert tags
	if len(jCard.Tags) > 0 {
		card.Tags = make([]model.CardTag, len(jCard.Tags))
		for i, tag := range jCard.Tags {
			cardTag, err := convertTag(tag)
			if err != nil {
				return nil, fmt.Errorf("invalid tag %s: %w", tag, err)
			}
			card.Tags[i] = cardTag
		}
	}

	// Convert requirements
	if len(jCard.Requirements) > 0 {
		card.Requirements = make([]model.Requirement, len(jCard.Requirements))
		for i, req := range jCard.Requirements {
			requirement, err := convertRequirement(req)
			if err != nil {
				return nil, fmt.Errorf("invalid requirement: %w", err)
			}
			card.Requirements[i] = requirement
		}
	}

	// Convert behaviors
	if len(jCard.Behaviors) > 0 {
		card.Behaviors = make([]model.CardBehavior, len(jCard.Behaviors))
		for i, behavior := range jCard.Behaviors {
			cardBehavior, err := convertBehavior(behavior)
			if err != nil {
				return nil, fmt.Errorf("invalid behavior: %w", err)
			}
			card.Behaviors[i] = cardBehavior
		}
	}

	// Convert VP conditions
	if len(jCard.VPConditions) > 0 {
		card.VPConditions = make([]model.VictoryPointCondition, len(jCard.VPConditions))
		for i, vpCondition := range jCard.VPConditions {
			condition, err := convertVPCondition(vpCondition)
			if err != nil {
				return nil, fmt.Errorf("invalid VP condition: %w", err)
			}
			card.VPConditions[i] = condition
		}
	}

	return card, nil
}

// convertTag converts a string tag to a CardTag
func convertTag(tag string) (model.CardTag, error) {
	switch tag {
	case "space":
		return model.TagSpace, nil
	case "earth":
		return model.TagEarth, nil
	case "science":
		return model.TagScience, nil
	case "power":
		return model.TagPower, nil
	case "building":
		return model.TagBuilding, nil
	case "microbe":
		return model.TagMicrobe, nil
	case "animal":
		return model.TagAnimal, nil
	case "plant":
		return model.TagPlant, nil
	case "event":
		return model.TagEvent, nil
	case "city":
		return model.TagCity, nil
	case "venus":
		return model.TagVenus, nil
	case "jovian":
		return model.TagJovian, nil
	case "wildlife":
		return model.TagWildlife, nil
	case "wild":
		return model.TagWild, nil
	default:
		return "", fmt.Errorf("unknown tag: %s", tag)
	}
}

// convertRequirement converts a JSON requirement to a model requirement
func convertRequirement(jReq jsonRequirement) (model.Requirement, error) {
	req := model.Requirement{
		Min: jReq.Min,
		Max: jReq.Max,
	}

	// Convert requirement type
	switch jReq.Type {
	case "temperature":
		req.Type = model.RequirementTemperature
	case "oxygen":
		req.Type = model.RequirementOxygen
	case "oceans":
		req.Type = model.RequirementOceans
	case "venus":
		req.Type = model.RequirementVenus
	case "cities":
		req.Type = model.RequirementCities
	case "greeneries":
		req.Type = model.RequirementGreeneries
	case "tags":
		req.Type = model.RequirementTags
	case "production":
		req.Type = model.RequirementProduction
	case "tr":
		req.Type = model.RequirementTR
	case "resource":
		req.Type = model.RequirementResource
	default:
		return req, fmt.Errorf("unknown requirement type: %s", jReq.Type)
	}

	// Convert optional fields
	if jReq.Location != nil {
		location := model.Location(*jReq.Location)
		req.Location = &location
	}

	if jReq.Tag != nil {
		tag, err := convertTag(*jReq.Tag)
		if err != nil {
			return req, err
		}
		req.Tag = &tag
	}

	if jReq.Resource != nil {
		resource, err := convertResourceType(*jReq.Resource)
		if err != nil {
			return req, err
		}
		req.Resource = &resource
	}

	return req, nil
}

// convertBehavior converts a JSON behavior to a model behavior
func convertBehavior(jBehavior jsonBehavior) (model.CardBehavior, error) {
	behavior := model.CardBehavior{}

	// Convert triggers
	if len(jBehavior.Triggers) > 0 {
		behavior.Triggers = make([]model.Trigger, len(jBehavior.Triggers))
		for i, trigger := range jBehavior.Triggers {
			modelTrigger, err := convertTrigger(trigger)
			if err != nil {
				return behavior, err
			}
			behavior.Triggers[i] = modelTrigger
		}
	}

	// Convert inputs
	if len(jBehavior.Inputs) > 0 {
		behavior.Inputs = make([]model.ResourceCondition, len(jBehavior.Inputs))
		for i, input := range jBehavior.Inputs {
			condition, err := convertResourceCondition(input)
			if err != nil {
				return behavior, err
			}
			behavior.Inputs[i] = condition
		}
	}

	// Convert outputs
	if len(jBehavior.Outputs) > 0 {
		behavior.Outputs = make([]model.ResourceCondition, len(jBehavior.Outputs))
		for i, output := range jBehavior.Outputs {
			condition, err := convertResourceCondition(output)
			if err != nil {
				return behavior, err
			}
			behavior.Outputs[i] = condition
		}
	}

	// Convert choices
	if len(jBehavior.Choices) > 0 {
		behavior.Choices = make([]model.Choice, len(jBehavior.Choices))
		for i, choice := range jBehavior.Choices {
			modelChoice, err := convertChoice(choice)
			if err != nil {
				return behavior, err
			}
			behavior.Choices[i] = modelChoice
		}
	}

	return behavior, nil
}

// convertTrigger converts a JSON trigger to a model trigger
func convertTrigger(jTrigger jsonTrigger) (model.Trigger, error) {
	trigger := model.Trigger{}

	// Convert trigger type
	switch jTrigger.Type {
	case "auto":
		trigger.Type = model.ResourceTriggerAuto
	case "manual":
		trigger.Type = model.ResourceTriggerManual
	default:
		return trigger, fmt.Errorf("unknown trigger type: %s", jTrigger.Type)
	}

	// Convert condition if present
	if jTrigger.Condition != nil {
		condition, err := convertTriggerCondition(*jTrigger.Condition)
		if err != nil {
			return trigger, err
		}
		trigger.Condition = &condition
	}

	return trigger, nil
}

// convertTriggerCondition converts a JSON trigger condition to a model trigger condition
func convertTriggerCondition(jCondition jsonTriggerCondition) (model.ResourceTriggerCondition, error) {
	condition := model.ResourceTriggerCondition{}

	// Convert condition type
	switch jCondition.Type {
	case "ocean-placed":
		condition.Type = model.TriggerOceanPlaced
	case "temperature-raise":
		condition.Type = model.TriggerTemperatureRaise
	case "oxygen-raise":
		condition.Type = model.TriggerOxygenRaise
	case "city-placed":
		condition.Type = model.TriggerCityPlaced
	case "card-played":
		condition.Type = model.TriggerCardPlayed
	case "tag-played":
		condition.Type = model.TriggerTagPlayed
	case "tile-placed":
		condition.Type = model.TriggerTilePlaced
	default:
		return condition, fmt.Errorf("unknown trigger condition type: %s", jCondition.Type)
	}

	// Convert location if present
	if jCondition.Location != nil {
		location := model.Location(*jCondition.Location)
		condition.Location = &location
	}

	return condition, nil
}

// convertResourceCondition converts a JSON resource condition to a model resource condition
func convertResourceCondition(jCondition jsonResourceCondition) (model.ResourceCondition, error) {
	condition := model.ResourceCondition{
		Amount: jCondition.Amount,
	}

	// Convert resource type
	resourceType, err := convertResourceType(jCondition.Type)
	if err != nil {
		return condition, err
	}
	condition.Type = resourceType

	// Convert target
	target, err := convertTarget(jCondition.Target)
	if err != nil {
		return condition, err
	}
	condition.Target = target

	// Convert affected tags if present
	if len(jCondition.AffectedTags) > 0 {
		condition.AffectedTags = make([]model.CardTag, len(jCondition.AffectedTags))
		for i, tag := range jCondition.AffectedTags {
			cardTag, err := convertTag(tag)
			if err != nil {
				return condition, err
			}
			condition.AffectedTags[i] = cardTag
		}
	}

	// Convert per condition if present
	if jCondition.Per != nil {
		perCondition, err := convertPerCondition(*jCondition.Per)
		if err != nil {
			return condition, err
		}
		condition.Per = &perCondition
	}

	return condition, nil
}

// convertChoice converts a JSON choice to a model choice
func convertChoice(jChoice jsonChoice) (model.Choice, error) {
	choice := model.Choice{}

	// Convert inputs
	if len(jChoice.Inputs) > 0 {
		choice.Inputs = make([]model.ResourceCondition, len(jChoice.Inputs))
		for i, input := range jChoice.Inputs {
			condition, err := convertResourceCondition(input)
			if err != nil {
				return choice, err
			}
			choice.Inputs[i] = condition
		}
	}

	// Convert outputs
	if len(jChoice.Outputs) > 0 {
		choice.Outputs = make([]model.ResourceCondition, len(jChoice.Outputs))
		for i, output := range jChoice.Outputs {
			condition, err := convertResourceCondition(output)
			if err != nil {
				return choice, err
			}
			choice.Outputs[i] = condition
		}
	}

	return choice, nil
}

// convertPerCondition converts a JSON per condition to a model per condition
func convertPerCondition(jPer jsonPerCondition) (model.PerCondition, error) {
	per := model.PerCondition{
		Amount: jPer.Amount,
	}

	// Convert type
	resourceType, err := convertResourceType(jPer.Type)
	if err != nil {
		return per, err
	}
	per.Type = resourceType

	// Convert optional fields
	if jPer.Location != nil {
		location := model.Location(*jPer.Location)
		per.Location = &location
	}

	if jPer.Target != nil {
		target, err := convertTarget(*jPer.Target)
		if err != nil {
			return per, err
		}
		per.Target = &target
	}

	if jPer.Tag != nil {
		tag, err := convertTag(*jPer.Tag)
		if err != nil {
			return per, err
		}
		per.Tag = &tag
	}

	return per, nil
}

// convertVPCondition converts a JSON VP condition to a model VP condition
func convertVPCondition(jVP jsonVPCondition) (model.VictoryPointCondition, error) {
	vpCondition := model.VictoryPointCondition{
		Amount:     jVP.Amount,
		MaxTrigger: jVP.MaxTrigger,
	}

	// Convert condition type
	switch jVP.Condition {
	case "per":
		vpCondition.Condition = model.VPConditionPer
	case "once":
		vpCondition.Condition = model.VPConditionOnce
	case "fixed":
		vpCondition.Condition = model.VPConditionFixed
	default:
		return vpCondition, fmt.Errorf("unknown VP condition type: %s", jVP.Condition)
	}

	// Convert per condition if present
	if jVP.Per != nil {
		perCondition, err := convertPerCondition(*jVP.Per)
		if err != nil {
			return vpCondition, err
		}
		vpCondition.Per = &perCondition
	}

	return vpCondition, nil
}

// convertResourceType converts a string resource type to a ResourceConditionType
func convertResourceType(resourceType string) (model.ResourceConditionType, error) {
	switch resourceType {
	case "credits":
		return model.ResourceCredits, nil
	case "steel":
		return model.ResourceSteel, nil
	case "titanium":
		return model.ResourceTitanium, nil
	case "plants":
		return model.ResourcePlants, nil
	case "energy":
		return model.ResourceEnergy, nil
	case "heat":
		return model.ResourceHeat, nil
	case "microbes":
		return model.ResourceMicrobes, nil
	case "animals":
		return model.ResourceAnimals, nil
	case "floaters":
		return model.ResourceFloaters, nil
	case "science":
		return model.ResourceScience, nil
	case "asteroid":
		return model.ResourceAsteroid, nil
	case "disease":
		return model.ResourceDisease, nil
	case "card-draw":
		return model.ResourceCardDraw, nil
	case "card-take":
		return model.ResourceCardTake, nil
	case "card-peek":
		return model.ResourceCardPeek, nil
	case "city-placement":
		return model.ResourceCityPlacement, nil
	case "ocean-placement":
		return model.ResourceOceanPlacement, nil
	case "greenery-placement":
		return model.ResourceGreeneryPlacement, nil
	case "city-tile":
		return model.ResourceCityTile, nil
	case "ocean-tile":
		return model.ResourceOceanTile, nil
	case "greenery-tile":
		return model.ResourceGreeneryTile, nil
	case "colony-tile":
		return model.ResourceColonyTile, nil
	case "temperature":
		return model.ResourceTemperature, nil
	case "oxygen":
		return model.ResourceOxygen, nil
	case "venus":
		return model.ResourceVenus, nil
	case "tr":
		return model.ResourceTR, nil
	case "credits-production":
		return model.ResourceCreditsProduction, nil
	case "steel-production":
		return model.ResourceSteelProduction, nil
	case "titanium-production":
		return model.ResourceTitaniumProduction, nil
	case "plants-production":
		return model.ResourcePlantsProduction, nil
	case "energy-production":
		return model.ResourceEnergyProduction, nil
	case "heat-production":
		return model.ResourceHeatProduction, nil
	case "effect":
		return model.ResourceEffect, nil
	case "tag":
		return model.ResourceTag, nil
	case "global-parameter-lenience":
		return model.ResourceGlobalParameterLenience, nil
	case "venus-lenience":
		return model.ResourceVenusLenience, nil
	case "defense":
		return model.ResourceDefense, nil
	case "discount":
		return model.ResourceDiscount, nil
	case "value-modifier":
		return model.ResourceValueModifier, nil
	default:
		return "", fmt.Errorf("unknown resource type: %s", resourceType)
	}
}

// convertTarget converts a string target to a TargetType
func convertTarget(target string) (model.TargetType, error) {
	switch target {
	case "self-player":
		return model.TargetSelfPlayer, nil
	case "self-card":
		return model.TargetSelfCard, nil
	case "any-player":
		return model.TargetAnyPlayer, nil
	case "opponent":
		return model.TargetOpponent, nil
	case "any":
		return model.TargetAny, nil
	case "none":
		return model.TargetNone, nil
	default:
		return "", fmt.Errorf("unknown target type: %s", target)
	}
}
