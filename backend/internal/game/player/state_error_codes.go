package player

// StateErrorCode represents error codes for entity state validation.
// All codes use kebab-case for consistency with JSON serialization.
type StateErrorCode string

const (
	ErrorCodeNotYourTurn StateErrorCode = "not-your-turn"
	ErrorCodeWrongPhase  StateErrorCode = "wrong-phase"

	ErrorCodeInsufficientCredits StateErrorCode = "insufficient-credits"

	ErrorCodeInsufficientResources StateErrorCode = "insufficient-resources"
	ErrorCodeTooManyResources      StateErrorCode = "too-many-resources"

	ErrorCodeTemperatureTooLow  StateErrorCode = "temperature-too-low"
	ErrorCodeTemperatureTooHigh StateErrorCode = "temperature-too-high"
	ErrorCodeOxygenTooLow       StateErrorCode = "oxygen-too-low"
	ErrorCodeOxygenTooHigh      StateErrorCode = "oxygen-too-high"
	ErrorCodeOceansTooLow       StateErrorCode = "oceans-too-low"
	ErrorCodeOceansTooHigh      StateErrorCode = "oceans-too-high"
	ErrorCodeTRTooLow           StateErrorCode = "tr-too-low"
	ErrorCodeTRTooHigh          StateErrorCode = "tr-too-high"

	ErrorCodeInsufficientTags       StateErrorCode = "insufficient-tags"
	ErrorCodeTooManyTags            StateErrorCode = "too-many-tags"
	ErrorCodeInsufficientProduction StateErrorCode = "insufficient-production"

	ErrorCodeNoOceanTiles         StateErrorCode = "no-ocean-tiles"
	ErrorCodeNoCityPlacements     StateErrorCode = "no-city-placements"
	ErrorCodeNoGreeneryPlacements StateErrorCode = "no-greenery-placements"
	ErrorCodeNoCardsInHand        StateErrorCode = "no-cards-in-hand"
	ErrorCodeInvalidProjectType   StateErrorCode = "invalid-project-type"
	ErrorCodeInvalidRequirement   StateErrorCode = "invalid-requirement"

	// Milestone error codes
	ErrorCodeMilestoneAlreadyClaimed    StateErrorCode = "milestone-already-claimed"
	ErrorCodeMaxMilestonesClaimed       StateErrorCode = "max-milestones-claimed"
	ErrorCodeMilestoneRequirementNotMet StateErrorCode = "milestone-requirement-not-met"

	// Award error codes
	ErrorCodeAwardAlreadyFunded StateErrorCode = "award-already-funded"
	ErrorCodeMaxAwardsFunded    StateErrorCode = "max-awards-funded"
)

// StateErrorCategory represents categories for error grouping.
// Categories enable UI filtering and display organization.
type StateErrorCategory string

const (
	ErrorCategoryTurn          StateErrorCategory = "turn"
	ErrorCategoryPhase         StateErrorCategory = "phase"
	ErrorCategoryCost          StateErrorCategory = "cost"
	ErrorCategoryInput         StateErrorCategory = "input"
	ErrorCategoryRequirement   StateErrorCategory = "requirement"
	ErrorCategoryAvailability  StateErrorCategory = "availability"
	ErrorCategoryConfiguration StateErrorCategory = "configuration"
	ErrorCategoryAchievement   StateErrorCategory = "achievement" // Milestones and awards
)
