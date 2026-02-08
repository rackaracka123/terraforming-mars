package shared

// Trigger represents when and how an action or effect is activated
type Trigger struct {
	Type      string                    `json:"type"`
	Condition *ResourceTriggerCondition `json:"condition,omitempty"`
}

// TriggerType constants for card behavior triggers
const (
	TriggerTypeAuto   = "auto"   // Automatic trigger (immediate effect when card is played)
	TriggerTypeManual = "manual" // Manual trigger (blue card action, activated by player)
)

// MinMaxValue represents a min/max value constraint
type MinMaxValue struct {
	Min *int `json:"min,omitempty"`
	Max *int `json:"max,omitempty"`
}

// Selector represents matching criteria for cards, resources, or projects.
// Multiple fields within a Selector use AND logic (all must match).
// Multiple Selectors in a slice use OR logic (any match is sufficient).
type Selector struct {
	Tags                 []CardTag         `json:"tags,omitempty"`
	CardTypes            []string          `json:"cardTypes,omitempty"`
	Resources            []string          `json:"resources,omitempty"`
	StandardProjects     []StandardProject `json:"standardProjects,omitempty"`
	RequiredOriginalCost *MinMaxValue      `json:"requiredOriginalCost,omitempty"`
}

// ResourceTriggerCondition represents what triggers an automatic resource exchange
type ResourceTriggerCondition struct {
	Type                 string         `json:"type"`
	ResourceTypes        []ResourceType `json:"resourceTypes,omitempty"`
	Location             *string        `json:"location,omitempty"`
	Selectors            []Selector     `json:"selectors,omitempty"`
	Target               *string        `json:"target,omitempty"`
	RequiredOriginalCost *MinMaxValue   `json:"requiredOriginalCost,omitempty"`
}

// TileRestrictions represents restrictions for tile placement
type TileRestrictions struct {
	BoardTags  []string `json:"boardTags,omitempty" ts:"string[]"`
	Adjacency  string   `json:"adjacency,omitempty" ts:"string"`  // "none" = no adjacent occupied tiles
	OnTileType string   `json:"onTileType,omitempty" ts:"string"` // "ocean" = only on ocean spaces
}

// ResourceCondition represents a resource amount (input or output)
type ResourceCondition struct {
	ResourceType     ResourceType      `json:"type"`
	Amount           int               `json:"amount"`
	Target           string            `json:"target"`
	Selectors        []Selector        `json:"selectors,omitempty"`
	MaxTrigger       *int              `json:"maxTrigger,omitempty"`
	Per              *PerCondition     `json:"per,omitempty"`
	TileRestrictions *TileRestrictions `json:"tileRestrictions,omitempty" ts:"TileRestrictions | undefined"`
}

// PerCondition represents what to count for conditional resource gains
type PerCondition struct {
	ResourceType ResourceType `json:"type"`
	Amount       int          `json:"amount"`
	Location     *string      `json:"location,omitempty"`
	Target       *string      `json:"target,omitempty"`
	Tag          *CardTag     `json:"tag,omitempty"`
}

// Choice represents a player choice option
type Choice struct {
	Inputs  []ResourceCondition `json:"inputs,omitempty"`
	Outputs []ResourceCondition `json:"outputs,omitempty"`
}
