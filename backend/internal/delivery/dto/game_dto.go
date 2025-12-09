package dto

// GamePhase represents the current phase of the game
type GamePhase string

const (
	GamePhaseWaitingForGameStart   GamePhase = "waiting_for_game_start"
	GamePhaseStartingCardSelection GamePhase = "starting_card_selection"
	GamePhaseStartGameSelection    GamePhase = "start_game_selection"
	GamePhaseAction                GamePhase = "action"
	GamePhaseProductionAndCardDraw GamePhase = "production_and_card_draw"
	GamePhaseComplete              GamePhase = "complete"
)

// GameStatus represents the current status of the game
type GameStatus string

const (
	GameStatusLobby     GameStatus = "lobby"
	GameStatusActive    GameStatus = "active"
	GameStatusCompleted GameStatus = "completed"
)

// CardType represents different types of cards
type CardType string

const (
	CardTypeAutomated   CardType = "automated" // Updated from "effect" to match JSON data
	CardTypeActive      CardType = "active"
	CardTypeEvent       CardType = "event"
	CardTypeCorporation CardType = "corporation"
	CardTypePrelude     CardType = "prelude"
)

// StandardProject represents the different types of standard projects
type StandardProject string

const (
	// Standard Projects (M€-based)
	StandardProjectSellPatents StandardProject = "sell-patents"
	StandardProjectPowerPlant  StandardProject = "power-plant"
	StandardProjectAsteroid    StandardProject = "asteroid"
	StandardProjectAquifer     StandardProject = "aquifer"
	StandardProjectGreenery    StandardProject = "greenery"
	StandardProjectCity        StandardProject = "city"
	// Resource Conversion Actions (resource-based, not M€)
	StandardProjectConvertPlantsToGreenery  StandardProject = "convert-plants-to-greenery"
	StandardProjectConvertHeatToTemperature StandardProject = "convert-heat-to-temperature"
)

// CardTag represents different card categories and attributes
type CardTag string

const (
	TagSpace    CardTag = "space"
	TagEarth    CardTag = "earth"
	TagScience  CardTag = "science"
	TagPower    CardTag = "power"
	TagBuilding CardTag = "building"
	TagMicrobe  CardTag = "microbe"
	TagAnimal   CardTag = "animal"
	TagPlant    CardTag = "plant"
	TagEvent    CardTag = "event"
	TagCity     CardTag = "city"
	TagVenus    CardTag = "venus"
	TagJovian   CardTag = "jovian"
	TagWildlife CardTag = "wildlife"
	TagWild     CardTag = "wild"
)

// ResourceType represents different types of resources for client consumption
// This is a 1:1 mapping from types.ResourceType
type ResourceType string

const (
	// Basic resources
	ResourceTypeCredits  ResourceType = "credits"
	ResourceTypeSteel    ResourceType = "steel"
	ResourceTypeTitanium ResourceType = "titanium"
	ResourceTypePlants   ResourceType = "plants"
	ResourceTypeEnergy   ResourceType = "energy"
	ResourceTypeHeat     ResourceType = "heat"
	ResourceTypeMicrobes ResourceType = "microbes"
	ResourceTypeAnimals  ResourceType = "animals"
	ResourceTypeFloaters ResourceType = "floaters"
	ResourceTypeScience  ResourceType = "science"
	ResourceTypeAsteroid ResourceType = "asteroid"
	ResourceTypeDisease  ResourceType = "disease"

	// Card actions
	ResourceTypeCardDraw ResourceType = "card-draw"
	ResourceTypeCardTake ResourceType = "card-take"
	ResourceTypeCardPeek ResourceType = "card-peek"

	// Terraforming actions
	ResourceTypeCityPlacement     ResourceType = "city-placement"
	ResourceTypeOceanPlacement    ResourceType = "ocean-placement"
	ResourceTypeGreeneryPlacement ResourceType = "greenery-placement"

	// Tile counting
	ResourceTypeCityTile     ResourceType = "city-tile"
	ResourceTypeOceanTile    ResourceType = "ocean-tile"
	ResourceTypeGreeneryTile ResourceType = "greenery-tile"
	ResourceTypeColonyTile   ResourceType = "colony-tile"

	// Global parameters
	ResourceTypeTemperature ResourceType = "temperature"
	ResourceTypeOxygen      ResourceType = "oxygen"
	ResourceTypeVenus       ResourceType = "venus"
	ResourceTypeTR          ResourceType = "tr"

	// Production resources
	ResourceTypeCreditsProduction  ResourceType = "credits-production"
	ResourceTypeSteelProduction    ResourceType = "steel-production"
	ResourceTypeTitaniumProduction ResourceType = "titanium-production"
	ResourceTypePlantsProduction   ResourceType = "plants-production"
	ResourceTypeEnergyProduction   ResourceType = "energy-production"
	ResourceTypeHeatProduction     ResourceType = "heat-production"

	// Special effects
	ResourceTypeEffect ResourceType = "effect"
	ResourceTypeTag    ResourceType = "tag"

	// Ongoing effects
	ResourceTypeGlobalParameterLenience ResourceType = "global-parameter-lenience"
	ResourceTypeVenusLenience           ResourceType = "venus-lenience"
	ResourceTypeDefense                 ResourceType = "defense"
	ResourceTypeDiscount                ResourceType = "discount"
	ResourceTypeValueModifier           ResourceType = "value-modifier"
)

// TargetType represents different targeting scopes for resource conditions for client consumption
type TargetType string

const (
	TargetSelfPlayer TargetType = "self-player"
	TargetSelfCard   TargetType = "self-card"
	TargetAnyCard    TargetType = "any-card"
	TargetAnyPlayer  TargetType = "any-player"
	TargetOpponent   TargetType = "opponent"
	TargetNone       TargetType = "none"
)

// CardApplyLocation represents different locations where card conditions can be evaluated for client consumption
type CardApplyLocation string

const (
	CardApplyLocationAnywhere CardApplyLocation = "anywhere"
	CardApplyLocationMars     CardApplyLocation = "mars"
)

// RequirementType represents different card requirement types for client consumption
type RequirementType string

const (
	RequirementTemperature RequirementType = "temperature"
	RequirementOxygen      RequirementType = "oxygen"
	RequirementOceans      RequirementType = "oceans"
	RequirementVenus       RequirementType = "venus"
	RequirementCities      RequirementType = "cities"
	RequirementGreeneries  RequirementType = "greeneries"
	RequirementTags        RequirementType = "tags"
	RequirementProduction  RequirementType = "production"
	RequirementTR          RequirementType = "tr"
	RequirementResource    RequirementType = "resource"
)

// VPConditionType represents different types of VP conditions for client consumption
type VPConditionType string

const (
	VPConditionFixed       VPConditionType = "fixed"
	VPConditionPer         VPConditionType = "per"
	VPConditionResourcesOn VPConditionType = "resources-on"
)

// TriggerType represents different trigger conditions for client consumption
type TriggerType string

const (
	TriggerOceanPlaced      TriggerType = "ocean-placed"
	TriggerTemperatureRaise TriggerType = "temperature-raise"
	TriggerOxygenRaise      TriggerType = "oxygen-raise"
	TriggerCityPlaced       TriggerType = "city-placed"
	TriggerCardPlayed       TriggerType = "card-played"
	TriggerTagPlayed        TriggerType = "tag-played"
	TriggerTilePlaced       TriggerType = "tile-placed"
)

// ResourceTriggerType represents different trigger types for resource exchanges for client consumption
type ResourceTriggerType string

const (
	ResourceTriggerManual                     ResourceTriggerType = "manual"
	ResourceTriggerAuto                       ResourceTriggerType = "auto"
	ResourceTriggerAutoCorporationFirstAction ResourceTriggerType = "auto-corporation-first-action"
	ResourceTriggerAutoCorporationStart       ResourceTriggerType = "auto-corporation-start"
)

// ResourceSet represents a collection of resources and their amounts
type ResourceSet struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// ResourceConditionDto represents a resource condition for client consumption
type ResourceConditionDto struct {
	Type                     ResourceType      `json:"type" ts:"ResourceType"`
	Amount                   int               `json:"amount" ts:"number"`
	Target                   TargetType        `json:"target" ts:"TargetType"`
	AffectedResources        []string          `json:"affectedResources,omitempty" ts:"string[] | undefined"`
	AffectedTags             []CardTag         `json:"affectedTags,omitempty" ts:"CardTag[] | undefined"`
	AffectedCardTypes        []CardType        `json:"affectedCardTypes,omitempty" ts:"CardType[] | undefined"`
	AffectedStandardProjects []StandardProject `json:"affectedStandardProjects,omitempty" ts:"StandardProject[] | undefined"`
	MaxTrigger               *int              `json:"maxTrigger,omitempty" ts:"number | undefined"`
	Per                      *PerConditionDto  `json:"per,omitempty" ts:"PerConditionDto | undefined"`
}

// PerConditionDto represents a per condition for client consumption
type PerConditionDto struct {
	Type     ResourceType       `json:"type" ts:"ResourceType"`
	Amount   int                `json:"amount" ts:"number"`
	Location *CardApplyLocation `json:"location,omitempty" ts:"CardApplyLocation | undefined"`
	Target   *TargetType        `json:"target,omitempty" ts:"TargetType | undefined"`
	Tag      *CardTag           `json:"tag,omitempty" ts:"CardTag | undefined"`
}

// ChoiceDto represents a choice for client consumption
type ChoiceDto struct {
	Inputs  []ResourceConditionDto `json:"inputs,omitempty" ts:"ResourceConditionDto[] | undefined"`
	Outputs []ResourceConditionDto `json:"outputs,omitempty" ts:"ResourceConditionDto[] | undefined"`
}

// TriggerDto represents a trigger for client consumption
type TriggerDto struct {
	Type      ResourceTriggerType          `json:"type" ts:"ResourceTriggerType"`
	Condition *ResourceTriggerConditionDto `json:"condition,omitempty" ts:"ResourceTriggerConditionDto | undefined"`
}

// MinMaxValueDto represents a minimum and/or maximum value constraint for client consumption
type MinMaxValueDto struct {
	Min *int `json:"min,omitempty" ts:"number | undefined"`
	Max *int `json:"max,omitempty" ts:"number | undefined"`
}

// ResourceTriggerConditionDto represents a resource trigger condition for client consumption
type ResourceTriggerConditionDto struct {
	Type                   TriggerType                     `json:"type" ts:"TriggerType"`
	Location               *CardApplyLocation              `json:"location,omitempty" ts:"CardApplyLocation | undefined"`
	AffectedTags           []CardTag                       `json:"affectedTags,omitempty" ts:"CardTag[] | undefined"`
	AffectedResources      []string                        `json:"affectedResources,omitempty" ts:"string[] | undefined"`   // Resource types that trigger this effect (for placement-bonus-gained)
	AffectedCardTypes      []CardType                      `json:"affectedCardTypes,omitempty" ts:"CardType[] | undefined"` // Card types that trigger this effect (for card-played)
	Target                 *TargetType                     `json:"target,omitempty" ts:"TargetType | undefined"`
	RequiredOriginalCost   *MinMaxValueDto                 `json:"requiredOriginalCost,omitempty" ts:"MinMaxValueDto | undefined"`
	RequiredResourceChange map[ResourceType]MinMaxValueDto `json:"requiredResourceChange,omitempty" ts:"Record<ResourceType, MinMaxValueDto> | undefined"`
}

// CardBehaviorDto represents a card behavior for client consumption
type CardBehaviorDto struct {
	Triggers []TriggerDto           `json:"triggers,omitempty" ts:"TriggerDto[] | undefined"`
	Inputs   []ResourceConditionDto `json:"inputs,omitempty" ts:"ResourceConditionDto[] | undefined"`
	Outputs  []ResourceConditionDto `json:"outputs,omitempty" ts:"ResourceConditionDto[] | undefined"`
	Choices  []ChoiceDto            `json:"choices,omitempty" ts:"ChoiceDto[] | undefined"`
}

// PaymentConstantsDto represents payment conversion rates
type PaymentConstantsDto struct {
	SteelValue    int `json:"steelValue" ts:"number"`
	TitaniumValue int `json:"titaniumValue" ts:"number"`
}

// RequirementDto represents a card requirement for client consumption
type RequirementDto struct {
	Type     RequirementType    `json:"type" ts:"RequirementType"`
	Min      *int               `json:"min,omitempty" ts:"number | undefined"`
	Max      *int               `json:"max,omitempty" ts:"number | undefined"`
	Location *CardApplyLocation `json:"location,omitempty" ts:"CardApplyLocation | undefined"`
	Tag      *CardTag           `json:"tag,omitempty" ts:"CardTag | undefined"`
	Resource *ResourceType      `json:"resource,omitempty" ts:"ResourceType | undefined"`
}

// ResourceStorageDto represents a card's resource storage for client consumption
type ResourceStorageDto struct {
	Type     ResourceType `json:"type" ts:"ResourceType"`
	Capacity *int         `json:"capacity,omitempty" ts:"number | undefined"`
	Starting int          `json:"starting" ts:"number"`
}

// VPConditionDto represents a victory point condition for client consumption
type VPConditionDto struct {
	Amount     int              `json:"amount" ts:"number"`
	Condition  VPConditionType  `json:"condition" ts:"VPConditionType"`
	MaxTrigger *int             `json:"maxTrigger,omitempty" ts:"number | undefined"`
	Per        *PerConditionDto `json:"per,omitempty" ts:"PerConditionDto | undefined"`
}

// ValidationErrorDto represents a validation failure reason
type ValidationErrorDto struct {
	Type          string      `json:"type" ts:"string"`
	Message       string      `json:"message" ts:"string"`
	RequiredValue interface{} `json:"requiredValue,omitempty" ts:"any"`
	CurrentValue  interface{} `json:"currentValue,omitempty" ts:"any"`
}

// ChoicePlayabilityDto represents playability for a single choice in a card action
type ChoicePlayabilityDto struct {
	ChoiceIndex        int                  `json:"choiceIndex" ts:"number"`
	IsAffordable       bool                 `json:"isAffordable" ts:"boolean"`
	UnaffordableErrors []ValidationErrorDto `json:"unaffordableErrors" ts:"ValidationErrorDto[]"`
}

// StandardProjectDto represents a standard project with availability information
type StandardProjectDto struct {
	ID                 string               `json:"id" ts:"string"`
	Name               string               `json:"name" ts:"string"`
	Type               string               `json:"type" ts:"string"`
	Cost               int                  `json:"cost" ts:"number"`
	Description        string               `json:"description" ts:"string"`
	IsAvailable        bool                 `json:"isAvailable" ts:"boolean"`
	UnavailableReasons []ValidationErrorDto `json:"unavailableReasons" ts:"ValidationErrorDto[]"`
}

// CardDto represents a card for client consumption
type CardDto struct {
	ID              string              `json:"id" ts:"string"`
	Name            string              `json:"name" ts:"string"`
	Type            CardType            `json:"type" ts:"CardType"`
	Cost            int                 `json:"cost" ts:"number"`
	Description     string              `json:"description" ts:"string"`
	Pack            string              `json:"pack" ts:"string"`
	Tags            []CardTag           `json:"tags,omitempty" ts:"CardTag[] | undefined"`
	Requirements    []RequirementDto    `json:"requirements,omitempty" ts:"RequirementDto[] | undefined"`
	Behaviors       []CardBehaviorDto   `json:"behaviors,omitempty" ts:"CardBehaviorDto[] | undefined"`
	ResourceStorage *ResourceStorageDto `json:"resourceStorage,omitempty" ts:"ResourceStorageDto | undefined"`
	VPConditions    []VPConditionDto    `json:"vpConditions,omitempty" ts:"VPConditionDto[] | undefined"`

	// Corporation-specific fields (nil for non-corporation cards)
	StartingResources  *ResourceSet `json:"startingResources,omitempty" ts:"ResourceSet | undefined"`
	StartingProduction *ResourceSet `json:"startingProduction,omitempty" ts:"ResourceSet | undefined"`

	// Playability fields (only present for cards in player's hand)
	IsPlayable       *bool                `json:"isPlayable,omitempty" ts:"boolean | undefined"`
	UnplayableErrors []ValidationErrorDto `json:"unplayableErrors,omitempty" ts:"ValidationErrorDto[] | undefined"`
}

type SelectStartingCardsPhaseDto struct {
	AvailableCards        []CardDto `json:"availableCards" ts:"CardDto[]"`        // Cards available for selection
	AvailableCorporations []CardDto `json:"availableCorporations" ts:"CardDto[]"` // Corporation cards available for selection (2 corporations)
}

type SelectStartingCardsOtherPlayerDto struct {
	// Empty - other players don't see any selection details
}

// ProductionPhaseDto represents card selection and production phase state for a player
type ProductionPhaseDto struct {
	AvailableCards    []CardDto    `json:"availableCards" ts:"CardDto[]"`  // Cards available for selection
	SelectionComplete bool         `json:"selectionComplete" ts:"boolean"` // Whether player completed card selection
	BeforeResources   ResourcesDto `json:"beforeResources" ts:"ResourcesDto"`
	AfterResources    ResourcesDto `json:"afterResources" ts:"ResourcesDto"`
	ResourceDelta     ResourcesDto `json:"resourceDelta" ts:"ResourceDelta"`
	EnergyConverted   int          `json:"energyConverted" ts:"number"`
	CreditsIncome     int          `json:"creditsIncome" ts:"number"`
}

type ProductionPhaseOtherPlayerDto struct {
	SelectionComplete bool         `json:"selectionComplete" ts:"boolean"` // Whether player completed card selection
	BeforeResources   ResourcesDto `json:"beforeResources" ts:"ResourcesDto"`
	AfterResources    ResourcesDto `json:"afterResources" ts:"ResourcesDto"`
	ResourceDelta     ResourcesDto `json:"resourceDelta" ts:"ResourceDelta"`
	EnergyConverted   int          `json:"energyConverted" ts:"number"`
	CreditsIncome     int          `json:"creditsIncome" ts:"number"`
}

// GameSettingsDto contains configurable game parameters
type GameSettingsDto struct {
	MaxPlayers      int      `json:"maxPlayers" ts:"number"`
	DevelopmentMode bool     `json:"developmentMode" ts:"boolean"`
	CardPacks       []string `json:"cardPacks,omitempty" ts:"string[] | undefined"`
}

// GlobalParametersDto represents the terraforming progress
type GlobalParametersDto struct {
	Temperature int `json:"temperature" ts:"number"` // Range: -30 to +8°C
	Oxygen      int `json:"oxygen" ts:"number"`      // Range: 0-14%
	Oceans      int `json:"oceans" ts:"number"`      // Range: 0-9
}

// ResourcesDto represents a player's resources
type ResourcesDto struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// ProductionDto represents a player's production values
type ProductionDto struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// PaymentSubstituteDto represents an alternative resource that can be used as payment for credits
type PaymentSubstituteDto struct {
	ResourceType   ResourceType `json:"resourceType" ts:"ResourceType"`
	ConversionRate int          `json:"conversionRate" ts:"number"`
}

// RequirementModifierDto represents a discount or lenience that modifies card/standard project requirements
// These are calculated from player effects and automatically updated when card hand or effects change
type RequirementModifierDto struct {
	Amount                int              `json:"amount" ts:"number"`                                               // Modifier amount (discount/lenience value)
	AffectedResources     []ResourceType   `json:"affectedResources" ts:"ResourceType[]"`                            // Resources affected (e.g., ["credits"] for price discount)
	CardTarget            *string          `json:"cardTarget,omitempty" ts:"string | undefined"`                     // Optional: specific card ID this applies to
	StandardProjectTarget *StandardProject `json:"standardProjectTarget,omitempty" ts:"StandardProject | undefined"` // Optional: specific standard project this applies to
}

// PlayerEffectDto represents ongoing effects that a player has active for client consumption
// Aligned with PlayerActionDto structure for consistent behavior handling
type PlayerEffectDto struct {
	CardID        string          `json:"cardId" ts:"string"`            // ID of the card that provides this effect
	CardName      string          `json:"cardName" ts:"string"`          // Name of the card for display purposes
	BehaviorIndex int             `json:"behaviorIndex" ts:"number"`     // Which behavior on the card this effect represents
	Behavior      CardBehaviorDto `json:"behavior" ts:"CardBehaviorDto"` // The actual behavior definition with inputs/outputs
	// Note: No PlayCount since effects are ongoing, not per-generation like actions
}

// PlayerActionDto represents an action that a player can take for client consumption
type PlayerActionDto struct {
	CardID        string          `json:"cardId" ts:"string"`            // ID of the card that provides this action
	CardName      string          `json:"cardName" ts:"string"`          // Name of the card for display purposes
	BehaviorIndex int             `json:"behaviorIndex" ts:"number"`     // Which behavior on the card this action represents
	Behavior      CardBehaviorDto `json:"behavior" ts:"CardBehaviorDto"` // The actual behavior definition with inputs/outputs
	PlayCount     int             `json:"playCount" ts:"number"`         // Number of times this action has been played this generation

	// Playability fields
	IsAffordable        bool                   `json:"isAffordable" ts:"boolean"`
	PlayableChoices     []int                  `json:"playableChoices" ts:"number[]"`
	ChoicePlayabilities []ChoicePlayabilityDto `json:"choicePlayabilities" ts:"ChoicePlayabilityDto[]"`
	UnaffordableErrors  []ValidationErrorDto   `json:"unaffordableErrors" ts:"ValidationErrorDto[]"`
}

// ForcedFirstActionDto represents an action that must be completed as the player's first turn action
type ForcedFirstActionDto struct {
	ActionType    string `json:"actionType" ts:"string"`    // Type of action: "city_placement", "card_draw", etc.
	CorporationID string `json:"corporationId" ts:"string"` // Corporation that requires this action
	Completed     bool   `json:"completed" ts:"boolean"`    // Whether the forced action has been completed
	Description   string `json:"description" ts:"string"`   // Human-readable description for UI
}

// PendingTileSelectionDto represents a pending tile placement action for client consumption
type PendingTileSelectionDto struct {
	TileType       string   `json:"tileType" ts:"string"`         // "city", "greenery", "ocean"
	AvailableHexes []string `json:"availableHexes" ts:"string[]"` // Backend-calculated valid hex coordinates
	Source         string   `json:"source" ts:"string"`           // What triggered this selection (card ID, standard project, etc.)
}

// PendingCardSelectionDto represents a pending card selection action (e.g., sell patents, card effects)
type PendingCardSelectionDto struct {
	AvailableCards []CardDto      `json:"availableCards" ts:"CardDto[]"`           // Card IDs player can select from
	CardCosts      map[string]int `json:"cardCosts" ts:"Record<string, number>"`   // Card ID -> cost to select (0 for sell patents, 3 for buying cards)
	CardRewards    map[string]int `json:"cardRewards" ts:"Record<string, number>"` // Card ID -> reward for selecting (1 MC for sell patents)
	Source         string         `json:"source" ts:"string"`                      // What triggered this selection ("sell-patents", card ID, etc.)
	MinCards       int            `json:"minCards" ts:"number"`                    // Minimum cards to select (0 for sell patents)
	MaxCards       int            `json:"maxCards" ts:"number"`                    // Maximum cards to select (hand size for sell patents)
}

// PendingCardDrawSelectionDto represents a pending card draw/peek/take/buy action from card effects
type PendingCardDrawSelectionDto struct {
	AvailableCards []CardDto `json:"availableCards" ts:"CardDto[]"` // Cards shown to player (drawn or peeked)
	FreeTakeCount  int       `json:"freeTakeCount" ts:"number"`     // Number of cards to take for free (mandatory for card-draw, 0 = optional)
	MaxBuyCount    int       `json:"maxBuyCount" ts:"number"`       // Maximum cards to buy (optional, 0 = no buying allowed)
	CardBuyCost    int       `json:"cardBuyCost" ts:"number"`       // Cost per card when buying (typically 3 MC, 0 if no buying)
	Source         string    `json:"source" ts:"string"`            // Card ID or action that triggered this
}

// PlayerStatus represents the current status of a player in the game
type PlayerStatus string

const (
	PlayerStatusSelectingStartingCards   PlayerStatus = "selecting-starting-cards"
	PlayerStatusSelectingProductionCards PlayerStatus = "selecting-production-cards"
	PlayerStatusWaiting                  PlayerStatus = "waiting"
	PlayerStatusActive                   PlayerStatus = "active"
)

// PlayerDto represents a player in the game for client consumption
type PlayerDto struct {
	ID               string            `json:"id" ts:"string"`
	Name             string            `json:"name" ts:"string"`
	Status           PlayerStatus      `json:"status" ts:"PlayerStatus"`
	Corporation      *CardDto          `json:"corporation" ts:"CardDto | null"`
	Cards            []CardDto         `json:"cards" ts:"CardDto[]"`
	Resources        ResourcesDto      `json:"resources" ts:"ResourcesDto"`
	Production       ProductionDto     `json:"production" ts:"ProductionDto"`
	TerraformRating  int               `json:"terraformRating" ts:"number"`
	PlayedCards      []CardDto         `json:"playedCards" ts:"CardDto[]"` // Full card details for all played cards
	Passed           bool              `json:"passed" ts:"boolean"`
	AvailableActions int               `json:"availableActions" ts:"number"`
	VictoryPoints    int               `json:"victoryPoints" ts:"number"`
	IsConnected      bool              `json:"isConnected" ts:"boolean"`
	Effects          []PlayerEffectDto `json:"effects" ts:"PlayerEffectDto[]"` // Active ongoing effects (discounts, special abilities, etc.)
	Actions          []PlayerActionDto `json:"actions" ts:"PlayerActionDto[]"` // Available actions from played cards with manual triggers

	SelectStartingCardsPhase *SelectStartingCardsPhaseDto `json:"selectStartingCardsPhase" ts:"SelectStartingCardsPhaseDto | null"`
	ProductionPhase          *ProductionPhaseDto          `json:"productionPhase" ts:"ProductionPhaseDto | null"`
	StartingCards            []CardDto                    `json:"startingCards" ts:"CardDto[]"` // Cards dealt at game start (from selectStartingCardsPhase.availableCards)
	// Tile selection - nullable, exists only when player needs to place tiles
	PendingTileSelection *PendingTileSelectionDto `json:"pendingTileSelection" ts:"PendingTileSelectionDto | null"` // Pending tile placement, null when no tiles to place
	// Card selection - nullable, exists only when player needs to select cards
	PendingCardSelection *PendingCardSelectionDto `json:"pendingCardSelection" ts:"PendingCardSelectionDto | null"` // Pending card selection (sell patents, card effects, etc.)
	// Card draw/peek/take/buy selection - nullable, exists only when player needs to confirm card draw selection
	PendingCardDrawSelection *PendingCardDrawSelectionDto `json:"pendingCardDrawSelection" ts:"PendingCardDrawSelectionDto | null"` // Pending card draw/peek/take/buy selection from card effects
	// Forced first action - nullable, exists only when corporation requires specific first turn action
	ForcedFirstAction *ForcedFirstActionDto `json:"forcedFirstAction" ts:"ForcedFirstActionDto | null"` // Action that must be taken on first turn (Tharsis city placement, etc.)
	// Resource storage - maps card IDs to resource counts stored on those cards
	ResourceStorage map[string]int `json:"resourceStorage" ts:"Record<string, number>"` // Card ID -> resource count
	// Payment substitutes - alternative resources usable as payment for credits
	PaymentSubstitutes []PaymentSubstituteDto `json:"paymentSubstitutes" ts:"PaymentSubstituteDto[]"` // Alternative resources usable as payment
	// Requirement modifiers - discounts and leniences calculated from effects (auto-updated on card hand/effects changes)
	RequirementModifiers []RequirementModifierDto `json:"requirementModifiers" ts:"RequirementModifierDto[]"` // Calculated discounts/leniences for cards and standard projects
}

// OtherPlayerDto represents another player from the viewing player's perspective (limited data)
type OtherPlayerDto struct {
	ID               string            `json:"id" ts:"string"`
	Name             string            `json:"name" ts:"string"`
	Status           PlayerStatus      `json:"status" ts:"PlayerStatus"`
	Corporation      *CardDto          `json:"corporation" ts:"CardDto | null"`
	HandCardCount    int               `json:"handCardCount" ts:"number"` // Number of cards in hand (private)
	Resources        ResourcesDto      `json:"resources" ts:"ResourcesDto"`
	Production       ProductionDto     `json:"production" ts:"ProductionDto"`
	TerraformRating  int               `json:"terraformRating" ts:"number"`
	PlayedCards      []CardDto         `json:"playedCards" ts:"CardDto[]"` // Played cards are public - full card details
	Passed           bool              `json:"passed" ts:"boolean"`
	AvailableActions int               `json:"availableActions" ts:"number"`
	VictoryPoints    int               `json:"victoryPoints" ts:"number"`
	IsConnected      bool              `json:"isConnected" ts:"boolean"`
	Effects          []PlayerEffectDto `json:"effects" ts:"PlayerEffectDto[]"`
	Actions          []PlayerActionDto `json:"actions" ts:"PlayerActionDto[]"`

	SelectStartingCardsPhase *SelectStartingCardsOtherPlayerDto `json:"selectStartingCardsPhase" ts:"SelectStartingCardsOtherPlayerDto | null"`
	ProductionPhase          *ProductionPhaseOtherPlayerDto     `json:"productionPhase" ts:"ProductionPhaseOtherPlayerDto | null"`
	// Resource storage - maps card IDs to resource counts stored on those cards (public information)
	ResourceStorage map[string]int `json:"resourceStorage" ts:"Record<string, number>"` // Card ID -> resource count
	// Payment substitutes - alternative resources usable as payment for credits (public information)
	PaymentSubstitutes []PaymentSubstituteDto `json:"paymentSubstitutes" ts:"PaymentSubstituteDto[]"` // Alternative resources usable as payment
}

// GameDto represents a game for client consumption (clean architecture)
type GameDto struct {
	ID               string               `json:"id" ts:"string"`
	Status           GameStatus           `json:"status" ts:"GameStatus"`
	Settings         GameSettingsDto      `json:"settings" ts:"GameSettingsDto"`
	HostPlayerID     string               `json:"hostPlayerId" ts:"string"`
	CurrentPhase     GamePhase            `json:"currentPhase" ts:"GamePhase"`
	GlobalParameters GlobalParametersDto  `json:"globalParameters" ts:"GlobalParametersDto"`
	CurrentPlayer    PlayerDto            `json:"currentPlayer" ts:"PlayerDto"`       // Viewing player's full data
	OtherPlayers     []OtherPlayerDto     `json:"otherPlayers" ts:"OtherPlayerDto[]"` // Other players' limited data
	ViewingPlayerID  string               `json:"viewingPlayerId" ts:"string"`        // The player viewing this game state
	CurrentTurn      *string              `json:"currentTurn" ts:"string|null"`       // Whose turn it is (nullable)
	Generation       int                  `json:"generation" ts:"number"`
	TurnOrder        []string             `json:"turnOrder" ts:"string[]"`                    // Turn order of all players in game
	Board            BoardDto             `json:"board" ts:"BoardDto"`                        // Game board with tiles and occupancy state
	PaymentConstants PaymentConstantsDto  `json:"paymentConstants" ts:"PaymentConstantsDto"`  // Conversion rates for alternative payments
	StandardProjects []StandardProjectDto `json:"standardProjects" ts:"StandardProjectDto[]"` // Available standard projects with playability
}

// Board-related DTOs for tygo generation

// TileBonusDto represents a resource bonus provided by a tile when occupied
type TileBonusDto struct {
	// Type specifies the resource type granted by this bonus
	Type string `json:"type" ts:"string"`
	// Amount specifies the quantity of the bonus granted
	Amount int `json:"amount" ts:"number"`
}

// TileOccupantDto represents what currently occupies a tile
type TileOccupantDto struct {
	// Type specifies the type of occupant (city-tile, ocean-tile, greenery-tile, etc.)
	Type string `json:"type" ts:"string"`
	// Tags specifies special properties of the occupant (e.g., "capital" for cities)
	Tags []string `json:"tags" ts:"string[]"`
}

// TileDto represents a single hexagonal tile on the game board
type TileDto struct {
	// Coordinates specifies the hex position of this tile
	Coordinates HexPositionDto `json:"coordinates" ts:"HexPositionDto"`
	// Tags specifies special properties for placement restrictions (e.g., "noctis-city")
	Tags []string `json:"tags" ts:"string[]"`
	// Type specifies the base type of tile (ocean-tile for ocean spaces, etc.)
	Type string `json:"type" ts:"string"`
	// Location specifies which celestial body this tile is on
	Location string `json:"location" ts:"string"`
	// DisplayName specifies the optional display name shown on the tile
	DisplayName *string `json:"displayName,omitempty" ts:"string|null"`
	// Bonuses specifies the resource bonuses provided by this tile
	Bonuses []TileBonusDto `json:"bonuses" ts:"TileBonusDto[]"`
	// OccupiedBy specifies what currently occupies this tile, if anything
	OccupiedBy *TileOccupantDto `json:"occupiedBy,omitempty" ts:"TileOccupantDto|null"`
	// OwnerID specifies the player who owns this tile, if any
	OwnerID *string `json:"ownerId,omitempty" ts:"string|null"`
}

// BoardDto represents the game board containing all tiles
type BoardDto struct {
	// Tiles specifies all tiles on the game board
	Tiles []TileDto `json:"tiles" ts:"TileDto[]"`
}
