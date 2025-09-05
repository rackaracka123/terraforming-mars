package dto

// ActionType represents different types of game actions
type ActionType string

const (
	ActionTypeSelectStartingCard ActionType = "select-starting-card"
	ActionTypeStartGame          ActionType = "start-game"
	ActionTypePlayCard           ActionType = "play-card"
	// Standard Projects
	ActionTypeSellPatents        ActionType = "sell-patents"
	ActionTypeBuildPowerPlant    ActionType = "build-power-plant"
	ActionTypeLaunchAsteroid     ActionType = "launch-asteroid"
	ActionTypeBuildAquifer       ActionType = "build-aquifer"
	ActionTypePlantGreenery      ActionType = "plant-greenery"
	ActionTypeBuildCity          ActionType = "build-city"
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

// HexPositionDto represents a position on the Mars board
type HexPositionDto struct {
	Q int `json:"q" ts:"number"`
	R int `json:"r" ts:"number"`
	S int `json:"s" ts:"number"`
}

// Standard Project Actions

// SellPatentsAction represents selling patent cards for megacredits
type SellPatentsAction struct {
	Type      ActionType `json:"type" ts:"ActionType"`
	CardCount int        `json:"cardCount" ts:"number"`
}

// BuildPowerPlantAction represents building a power plant
type BuildPowerPlantAction struct {
	Type ActionType `json:"type" ts:"ActionType"`
}

// LaunchAsteroidAction represents launching an asteroid
type LaunchAsteroidAction struct {
	Type ActionType `json:"type" ts:"ActionType"`
}

// BuildAquiferAction represents building an aquifer
type BuildAquiferAction struct {
	Type        ActionType     `json:"type" ts:"ActionType"`
	HexPosition HexPositionDto `json:"hexPosition" ts:"HexPositionDto"`
}

// PlantGreeneryAction represents planting greenery
type PlantGreeneryAction struct {
	Type        ActionType     `json:"type" ts:"ActionType"`
	HexPosition HexPositionDto `json:"hexPosition" ts:"HexPositionDto"`
}

// BuildCityAction represents building a city
type BuildCityAction struct {
	Type        ActionType     `json:"type" ts:"ActionType"`
	HexPosition HexPositionDto `json:"hexPosition" ts:"HexPositionDto"`
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

// Standard Project Action Requests

// ActionSellPatentsRequest contains the action data for sell patents actions
type ActionSellPatentsRequest struct {
	Type      ActionType `json:"type" ts:"ActionType"`
	CardCount int        `json:"cardCount" ts:"number"`
}

// GetAction returns the sell patents action
func (ap *ActionSellPatentsRequest) GetAction() *SellPatentsAction {
	return &SellPatentsAction{Type: ap.Type, CardCount: ap.CardCount}
}

// ActionBuildPowerPlantRequest contains the action data for build power plant actions
type ActionBuildPowerPlantRequest struct {
	Type ActionType `json:"type" ts:"ActionType"`
}

// GetAction returns the build power plant action
func (ap *ActionBuildPowerPlantRequest) GetAction() *BuildPowerPlantAction {
	return &BuildPowerPlantAction{Type: ap.Type}
}

// ActionLaunchAsteroidRequest contains the action data for launch asteroid actions
type ActionLaunchAsteroidRequest struct {
	Type ActionType `json:"type" ts:"ActionType"`
}

// GetAction returns the launch asteroid action
func (ap *ActionLaunchAsteroidRequest) GetAction() *LaunchAsteroidAction {
	return &LaunchAsteroidAction{Type: ap.Type}
}

// ActionBuildAquiferRequest contains the action data for build aquifer actions
type ActionBuildAquiferRequest struct {
	Type        ActionType     `json:"type" ts:"ActionType"`
	HexPosition HexPositionDto `json:"hexPosition" ts:"HexPositionDto"`
}

// GetAction returns the build aquifer action
func (ap *ActionBuildAquiferRequest) GetAction() *BuildAquiferAction {
	return &BuildAquiferAction{Type: ap.Type, HexPosition: ap.HexPosition}
}

// ActionPlantGreeneryRequest contains the action data for plant greenery actions
type ActionPlantGreeneryRequest struct {
	Type        ActionType     `json:"type" ts:"ActionType"`
	HexPosition HexPositionDto `json:"hexPosition" ts:"HexPositionDto"`
}

// GetAction returns the plant greenery action
func (ap *ActionPlantGreeneryRequest) GetAction() *PlantGreeneryAction {
	return &PlantGreeneryAction{Type: ap.Type, HexPosition: ap.HexPosition}
}

// ActionBuildCityRequest contains the action data for build city actions
type ActionBuildCityRequest struct {
	Type        ActionType     `json:"type" ts:"ActionType"`
	HexPosition HexPositionDto `json:"hexPosition" ts:"HexPositionDto"`
}

// GetAction returns the build city action
func (ap *ActionBuildCityRequest) GetAction() *BuildCityAction {
	return &BuildCityAction{Type: ap.Type, HexPosition: ap.HexPosition}
}

