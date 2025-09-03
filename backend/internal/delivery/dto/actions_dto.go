package dto

// ActionType represents different types of game actions
type ActionType string

const (
	ActionTypeSelectStartingCard ActionType = "select-starting-card"
	ActionTypeStartGame          ActionType = "start-game"
)

// ActionRequest is the base interface for all action requests
type ActionRequest interface {
	GetActionType() ActionType
}

// SelectStartingCardAction represents selecting starting cards
type SelectStartingCardAction struct {
	Type    ActionType `json:"type" ts:"ActionType"`
	CardIDs []string   `json:"cardIds" ts:"string[]"`
}

func (a SelectStartingCardAction) GetActionType() ActionType {
	return ActionTypeSelectStartingCard
}

// StartGameAction represents starting the game (host only)
type StartGameAction struct {
	Type ActionType `json:"type" ts:"ActionType"`
}

func (a StartGameAction) GetActionType() ActionType {
	return ActionTypeStartGame
}

// ActionPayload contains the action data for WebSocket messages
type ActionPayload struct {
	Type    ActionType `json:"type" ts:"ActionType"`
	CardIDs []string   `json:"cardIds,omitempty" ts:"string[]"`
}

// GetAction returns the specific action based on the type
func (ap *ActionPayload) GetAction() ActionRequest {
	switch ap.Type {
	case ActionTypeSelectStartingCard:
		if ap.CardIDs != nil {
			return &SelectStartingCardAction{Type: ap.Type, CardIDs: ap.CardIDs}
		}
		return nil
	case ActionTypeStartGame:
		return &StartGameAction{Type: ap.Type}
	default:
		return nil
	}
}