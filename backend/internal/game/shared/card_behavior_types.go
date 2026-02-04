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

// ResourceTriggerCondition represents what triggers an automatic resource exchange
type ResourceTriggerCondition struct {
	Type              string         `json:"type"`
	ResourceTypes     []ResourceType `json:"resourceTypes,omitempty"`
	Location          *string        `json:"location,omitempty"`
	AffectedTags      []CardTag      `json:"affectedTags,omitempty"`
	AffectedResources []string       `json:"affectedResources,omitempty"`
	AffectedCardTypes []string       `json:"affectedCardTypes,omitempty"`
	Target            *string        `json:"target,omitempty"`
}

// TileRestrictions represents restrictions for tile placement
type TileRestrictions struct {
	BoardTags []string `json:"boardTags,omitempty" ts:"string[]"`
	Adjacency string   `json:"adjacency,omitempty" ts:"string"` // "none" = no adjacent occupied tiles
}

// ResourceCondition represents a resource amount (input or output)
type ResourceCondition struct {
	ResourceType             ResourceType      `json:"type"`
	Amount                   int               `json:"amount"`
	Target                   string            `json:"target"`
	AffectedResources        []string          `json:"affectedResources,omitempty"`
	AffectedTags             []CardTag         `json:"affectedTags,omitempty"`
	AffectedCardTypes        []string          `json:"affectedCardTypes,omitempty"`
	AffectedStandardProjects []StandardProject `json:"affectedStandardProjects,omitempty"`
	MaxTrigger               *int              `json:"maxTrigger,omitempty"`
	Per                      *PerCondition     `json:"per,omitempty"`
	TileRestrictions         *TileRestrictions `json:"tileRestrictions,omitempty" ts:"TileRestrictions | undefined"`
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
