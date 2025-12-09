package playability

// ValidationErrorType represents the type of validation error
type ValidationErrorType string

const (
	// ValidationErrorTypeCost indicates insufficient resources for cost
	ValidationErrorTypeCost ValidationErrorType = "cost"
	// ValidationErrorTypeRequirement indicates unmet card requirement
	ValidationErrorTypeRequirement ValidationErrorType = "requirement"
	// ValidationErrorTypeGlobalParam indicates global parameter at limit
	ValidationErrorTypeGlobalParam ValidationErrorType = "global-param"
	// ValidationErrorTypePhase indicates wrong game phase
	ValidationErrorTypePhase ValidationErrorType = "phase"
	// ValidationErrorTypeResource indicates insufficient resources
	ValidationErrorTypeResource ValidationErrorType = "resource"
	// ValidationErrorTypeTurn indicates not player's turn
	ValidationErrorTypeTurn ValidationErrorType = "turn"
	// ValidationErrorTypeGameState indicates invalid game state
	ValidationErrorTypeGameState ValidationErrorType = "game-state"
)

// ValidationError represents a single validation failure
type ValidationError struct {
	Type          ValidationErrorType `json:"type"`
	Message       string              `json:"message"`
	RequiredValue interface{}         `json:"requiredValue,omitempty"`
	CurrentValue  interface{}         `json:"currentValue,omitempty"`
}

// PlayabilityResult represents the result of a playability check
type PlayabilityResult struct {
	IsPlayable bool              `json:"isPlayable"`
	Errors     []ValidationError `json:"errors"`
}

// NewPlayabilityResult creates a new playability result
func NewPlayabilityResult(isPlayable bool, errors []ValidationError) PlayabilityResult {
	if errors == nil {
		errors = []ValidationError{}
	}
	return PlayabilityResult{
		IsPlayable: isPlayable,
		Errors:     errors,
	}
}

// AddError adds a validation error to the result
func (r *PlayabilityResult) AddError(err ValidationError) {
	r.IsPlayable = false
	r.Errors = append(r.Errors, err)
}

// ChoicePlayability represents playability for a choice-based action
type ChoicePlayability struct {
	ChoiceIndex        int               `json:"choiceIndex"`
	IsAffordable       bool              `json:"isAffordable"`
	UnaffordableErrors []ValidationError `json:"unaffordableErrors"`
}

// ActionPlayabilityResult extends PlayabilityResult for card actions with choices
type ActionPlayabilityResult struct {
	IsAffordable        bool                `json:"isAffordable"`
	Errors              []ValidationError   `json:"errors"`
	PlayableChoices     []int               `json:"playableChoices"`
	ChoicePlayabilities []ChoicePlayability `json:"choicePlayabilities"`
}

// NewActionPlayabilityResult creates a new action playability result
func NewActionPlayabilityResult() ActionPlayabilityResult {
	return ActionPlayabilityResult{
		IsAffordable:        false,
		Errors:              []ValidationError{},
		PlayableChoices:     []int{},
		ChoicePlayabilities: []ChoicePlayability{},
	}
}

// AddPlayableChoice marks a choice as playable
func (r *ActionPlayabilityResult) AddPlayableChoice(choiceIndex int) {
	r.IsAffordable = true
	r.PlayableChoices = append(r.PlayableChoices, choiceIndex)
	r.ChoicePlayabilities = append(r.ChoicePlayabilities, ChoicePlayability{
		ChoiceIndex:        choiceIndex,
		IsAffordable:       true,
		UnaffordableErrors: []ValidationError{},
	})
}

// AddUnaffordableChoice marks a choice as unaffordable with reasons
func (r *ActionPlayabilityResult) AddUnaffordableChoice(choiceIndex int, errors []ValidationError) {
	r.ChoicePlayabilities = append(r.ChoicePlayabilities, ChoicePlayability{
		ChoiceIndex:        choiceIndex,
		IsAffordable:       false,
		UnaffordableErrors: errors,
	})
}

// StandardProjectType represents a standard project
type StandardProjectType string

const (
	StandardProjectSellPatents    StandardProjectType = "sell-patents"
	StandardProjectPowerPlant     StandardProjectType = "power-plant"
	StandardProjectAsteroid       StandardProjectType = "asteroid"
	StandardProjectAquifer        StandardProjectType = "aquifer"
	StandardProjectGreenery       StandardProjectType = "greenery"
	StandardProjectCity           StandardProjectType = "city"
	StandardProjectCapitalToVenus StandardProjectType = "capital-to-venus"
	StandardProjectAirScrapping   StandardProjectType = "air-scrapping"
)

// StandardProject represents a standard project with its properties
type StandardProject struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Type        StandardProjectType `json:"type"`
	Cost        int                 `json:"cost"`
	Description string              `json:"description"`
	IsAvailable bool                `json:"isAvailable"`
	Errors      []ValidationError   `json:"unavailableReasons"`
}
