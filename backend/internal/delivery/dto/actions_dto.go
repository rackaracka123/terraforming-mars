package dto

// ActionType represents different types of game actions
type ActionType string

const (
	ActionTypeSelectStartingCard ActionType = "select-starting-card"
	ActionTypeStartGame          ActionType = "start-game"
	ActionTypePlayCard           ActionType = "play-card"
)

// SelectStartingCardAction represents selecting starting cards
type SelectStartingCardAction struct {
	Type    ActionType `json:"type" ts:"ActionType"`
	CardIDs []string   `json:"cardIds" ts:"string[]"`
}

// StartGameAction represents starting the game (host only)
type StartGameAction struct {
	Type ActionType `json:"type" ts:"ActionType"`
}

// PlayCardAction represents playing a card from hand
type PlayCardAction struct {
	CardID string `json:"cardId" ts:"string"`
}


// ActionSelectStartingCardRequest contains the action data for select starting card actions
type ActionSelectStartingCardRequest struct {
	Type    ActionType `json:"type" ts:"ActionType"`
	CardIDs []string   `json:"cardIds" ts:"string[]"`
}

// GetAction returns the select starting card action
func (ap *ActionSelectStartingCardRequest) GetAction() *SelectStartingCardAction {
	return &SelectStartingCardAction{Type: ap.Type, CardIDs: ap.CardIDs}
}

// ActionStartGameRequest contains the action data for start game actions
type ActionStartGameRequest struct {
	Type ActionType `json:"type" ts:"ActionType"`
}

// GetAction returns the start game action
func (ap *ActionStartGameRequest) GetAction() *StartGameAction {
	return &StartGameAction{Type: ap.Type}
}

// ActionPlayCardRequest contains the action data for play card actions
type ActionPlayCardRequest struct {
	Type   ActionType `json:"type" ts:"ActionType"`
	CardID string     `json:"cardId" ts:"string"`
}

// GetAction returns the play card action
func (ap *ActionPlayCardRequest) GetAction() *PlayCardAction {
	return &PlayCardAction{CardID: ap.CardID}
}

