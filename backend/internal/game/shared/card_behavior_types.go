package shared

// ==================== Card Behavior Related Types ====================

// Trigger represents when and how an action or effect is activated
type Trigger struct {
	Type      string                    `json:"type"`
	Condition *ResourceTriggerCondition `json:"condition,omitempty"`
}

// ResourceTriggerCondition represents what triggers an automatic resource exchange
type ResourceTriggerCondition struct {
	Type              string         `json:"type"`
	ResourceTypes     []ResourceType `json:"resourceTypes,omitempty"`
	Location          *string        `json:"location,omitempty"`
	AffectedTags      []CardTag      `json:"affectedTags,omitempty"`
	AffectedResources []string       `json:"affectedResources,omitempty"`
	Target            *string        `json:"target,omitempty"`
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
