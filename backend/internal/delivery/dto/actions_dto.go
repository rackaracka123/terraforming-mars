package dto

// ActionType represents different types of game actions
type ActionType string

const (
	ActionTypeStandardProjectAsteroid ActionType = "standard-project-asteroid"
	ActionTypeRaiseTemperature        ActionType = "raise-temperature"
	ActionTypeSelectCorporation       ActionType = "select-corporation"
	ActionTypeSkipAction              ActionType = "skip-action"
)

// ActionRequest is the base interface for all action requests
type ActionRequest interface {
	GetActionType() ActionType
}

// StandardProjectAsteroidAction represents the asteroid standard project action
type StandardProjectAsteroidAction struct {
	Type ActionType `json:"type" ts:"ActionType"`
}

func (a StandardProjectAsteroidAction) GetActionType() ActionType {
	return ActionTypeStandardProjectAsteroid
}

// RaiseTemperatureAction represents the raise temperature action using heat
type RaiseTemperatureAction struct {
	Type       ActionType `json:"type" ts:"ActionType"`
	HeatAmount int        `json:"heatAmount" ts:"number"`
}

func (a RaiseTemperatureAction) GetActionType() ActionType {
	return ActionTypeRaiseTemperature
}

// SelectCorporationAction represents selecting a corporation
type SelectCorporationAction struct {
	Type            ActionType `json:"type" ts:"ActionType"`
	CorporationName string     `json:"corporationName" ts:"string"`
}

func (a SelectCorporationAction) GetActionType() ActionType {
	return ActionTypeSelectCorporation
}

// SkipActionAction represents skipping the current action
type SkipActionAction struct {
	Type ActionType `json:"type" ts:"ActionType"`
}

func (a SkipActionAction) GetActionType() ActionType {
	return ActionTypeSkipAction
}

// ActionPayload contains the action data for WebSocket messages
type ActionPayload struct {
	Type            ActionType `json:"type" ts:"ActionType"`
	HeatAmount      *int       `json:"heatAmount,omitempty" ts:"number"`
	CorporationName *string    `json:"corporationName,omitempty" ts:"string"`
}

// GetAction returns the specific action based on the type
func (ap *ActionPayload) GetAction() ActionRequest {
	switch ap.Type {
	case ActionTypeStandardProjectAsteroid:
		return &StandardProjectAsteroidAction{Type: ap.Type}
	case ActionTypeRaiseTemperature:
		if ap.HeatAmount != nil {
			return &RaiseTemperatureAction{Type: ap.Type, HeatAmount: *ap.HeatAmount}
		}
		return nil
	case ActionTypeSelectCorporation:
		if ap.CorporationName != nil {
			return &SelectCorporationAction{Type: ap.Type, CorporationName: *ap.CorporationName}
		}
		return nil
	case ActionTypeSkipAction:
		return &SkipActionAction{Type: ap.Type}
	default:
		return nil
	}
}