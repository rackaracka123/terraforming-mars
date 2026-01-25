package cards

import (
	"terraforming-mars-backend/internal/game/shared"
)

// ResourceStorage represents a card's ability to hold resources
type ResourceStorage struct {
	Type     shared.ResourceType `json:"type"`
	Capacity *int                `json:"capacity,omitempty"`
	Starting int                 `json:"starting"`
}

// VictoryPointCondition represents a VP condition
type VictoryPointCondition struct {
	Amount     int             `json:"amount"`
	Condition  VPConditionType `json:"condition"`
	MaxTrigger *int            `json:"maxTrigger,omitempty"`
	Per        *PerCondition   `json:"per,omitempty"`
}

// VPConditionType represents different types of VP conditions
type VPConditionType string

const (
	VPConditionPer   VPConditionType = "per"
	VPConditionOnce  VPConditionType = "once"
	VPConditionFixed VPConditionType = "fixed"
)
