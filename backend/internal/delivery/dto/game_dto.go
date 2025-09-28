package dto

import "terraforming-mars-backend/internal/model"

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
type ResourceType string

const (
	ResourceTypeCredits         ResourceType = "credits"
	ResourceTypeSteel           ResourceType = "steel"
	ResourceTypeTitanium        ResourceType = "titanium"
	ResourceTypePlants          ResourceType = "plants"
	ResourceTypeEnergy          ResourceType = "energy"
	ResourceTypeHeat            ResourceType = "heat"
	ResourceTypeFloaters        ResourceType = "floaters"
	ResourceTypeMicrobes        ResourceType = "microbes"
	ResourceTypeAnimals         ResourceType = "animals"
	ResourceTypeScience         ResourceType = "science"
	ResourceTypeFighters        ResourceType = "fighters"
	ResourceTypeCamps           ResourceType = "camps"
	ResourceTypePreservation    ResourceType = "preservation"
	ResourceTypeData            ResourceType = "data"
	ResourceTypeAsteroid        ResourceType = "asteroid"
	ResourceTypeDisease         ResourceType = "disease"
	ResourceTypeSpecialized     ResourceType = "specialized"
	ResourceTypeDelegate        ResourceType = "delegate"
	ResourceTypeInfluence       ResourceType = "influence"
	ResourceTypeGreeneryTile    ResourceType = "greenery-tile"
	ResourceTypeCityTile        ResourceType = "city-tile"
	ResourceTypeOceanTile       ResourceType = "ocean-tile"
	ResourceTypeSpecialTile     ResourceType = "special-tile"
	ResourceTypeTerraformRating ResourceType = "terraform-rating"
	ResourceTypeTemperature     ResourceType = "temperature"
	ResourceTypeOxygen          ResourceType = "oxygen"
	ResourceTypeOceans          ResourceType = "oceans"
)

// TargetType represents different targeting scopes for resource conditions for client consumption
type TargetType string

const (
	TargetSelfPlayer TargetType = "self-player"
	TargetSelfCard   TargetType = "self-card"
	TargetAnyPlayer  TargetType = "any-player"
	TargetOpponent   TargetType = "opponent"
	TargetAny        TargetType = "any"
	TargetNone       TargetType = "none"
)

// CardApplyLocation represents different locations where card conditions can be evaluated for client consumption
type CardApplyLocation string

const (
	CardApplyLocationAnywhere CardApplyLocation = "anywhere"
	CardApplyLocationMars     CardApplyLocation = "mars"
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
	ResourceTriggerManual ResourceTriggerType = "manual"
	ResourceTriggerAuto   ResourceTriggerType = "auto"
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
	Type              ResourceType     `json:"type" ts:"ResourceType"`
	Amount            int              `json:"amount" ts:"number"`
	Target            TargetType       `json:"target" ts:"TargetType"`
	AffectedResources []string         `json:"affectedResources,omitempty" ts:"string[] | undefined"`
	AffectedTags      []CardTag        `json:"affectedTags,omitempty" ts:"CardTag[] | undefined"`
	MaxTrigger        *int             `json:"maxTrigger,omitempty" ts:"number | undefined"`
	Per               *PerConditionDto `json:"per,omitempty" ts:"PerConditionDto | undefined"`
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

// ResourceTriggerConditionDto represents a resource trigger condition for client consumption
type ResourceTriggerConditionDto struct {
	Type         TriggerType        `json:"type" ts:"TriggerType"`
	Location     *CardApplyLocation `json:"location,omitempty" ts:"CardApplyLocation | undefined"`
	AffectedTags []CardTag          `json:"affectedTags,omitempty" ts:"CardTag[] | undefined"`
}

// CardBehaviorDto represents a card behavior for client consumption
type CardBehaviorDto struct {
	Triggers []TriggerDto           `json:"triggers,omitempty" ts:"TriggerDto[] | undefined"`
	Inputs   []ResourceConditionDto `json:"inputs,omitempty" ts:"ResourceConditionDto[] | undefined"`
	Outputs  []ResourceConditionDto `json:"outputs,omitempty" ts:"ResourceConditionDto[] | undefined"`
	Choices  []ChoiceDto            `json:"choices,omitempty" ts:"ChoiceDto[] | undefined"`
}

// ResourceStorageDto represents a card's resource storage capability for client consumption
type ResourceStorageDto struct {
	Type     model.ResourceType `json:"type" ts:"ResourceType"`
	Capacity *int               `json:"capacity,omitempty" ts:"number | undefined"`
	Starting int                `json:"starting" ts:"number"`
}

// CardDto represents a card for client consumption
type CardDto struct {
	ID              string                        `json:"id" ts:"string"`
	Name            string                        `json:"name" ts:"string"`
	Type            CardType                      `json:"type" ts:"CardType"`
	Cost            int                           `json:"cost" ts:"number"`
	Description     string                        `json:"description" ts:"string"`
	Tags            []CardTag                     `json:"tags" ts:"CardTag[]"`
	Requirements    []model.Requirement           `json:"requirements,omitempty" ts:"Requirement[] | undefined"`
	Behaviors       []CardBehaviorDto             `json:"behaviors,omitempty" ts:"CardBehaviorDto[] | undefined"`
	ResourceStorage *ResourceStorageDto           `json:"resourceStorage,omitempty" ts:"ResourceStorageDto | undefined"`
	VPConditions    []model.VictoryPointCondition `json:"vpConditions,omitempty" ts:"VictoryPointCondition[] | undefined"`
}

// CorporationDto represents a corporation for client consumption
type CorporationDto struct {
	ID                 string      `json:"id" ts:"string"`
	Name               string      `json:"name" ts:"string"`
	Description        string      `json:"description" ts:"string"`
	StartingCredits    int         `json:"startingCredits" ts:"number"`
	StartingResources  ResourceSet `json:"startingResources" ts:"ResourceSet"`
	StartingProduction ResourceSet `json:"startingProduction" ts:"ResourceSet"`
	Tags               []CardTag   `json:"tags" ts:"CardTag[]"`
	SpecialEffects     []string    `json:"specialEffects" ts:"string[]"`
	Number             string      `json:"number" ts:"string"`
}

type SelectStartingCardsPhaseDto struct {
	AvailableCards    []CardDto `json:"availableCards" ts:"CardDto[]"`  // Cards available for selection
	SelectionComplete bool      `json:"selectionComplete" ts:"boolean"` // Whether player completed card selection
}

type SelectStartingCardsOtherPlayerDto struct {
	SelectionComplete bool `json:"selectionComplete" ts:"boolean"` // Whether player completed card selection
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
	MaxPlayers      int  `json:"maxPlayers" ts:"number"`
	DevelopmentMode bool `json:"developmentMode" ts:"boolean"`
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

// PlayerEffectType represents different types of ongoing effects a player can have
type PlayerEffectType string

const (
	PlayerEffectTypeDiscount                PlayerEffectType = "discount"                  // Cost reduction for playing cards
	PlayerEffectTypeGlobalParameterLenience PlayerEffectType = "global-parameter-lenience" // Global parameter requirement flexibility
	PlayerEffectTypeDefense                 PlayerEffectType = "defense"                   // Protection from attacks or resource removal
	PlayerEffectTypeValueModifier           PlayerEffectType = "value-modifier"            // Increases resource values (e.g., steel/titanium worth more)
)

// PlayerEffectDto represents ongoing effects that a player has active for client consumption
type PlayerEffectDto struct {
	Type         PlayerEffectType `json:"type" ts:"PlayerEffectType"`                        // Type of effect
	Amount       int              `json:"amount" ts:"number"`                                // Effect amount (e.g., M€ discount, steps of flexibility)
	AffectedTags []CardTag        `json:"affectedTags,omitempty" ts:"CardTag[] | undefined"` // Tags that qualify for this effect (empty = all cards)
}

// PlayerActionDto represents an action that a player can take for client consumption
type PlayerActionDto struct {
	CardID        string          `json:"cardId" ts:"string"`            // ID of the card that provides this action
	CardName      string          `json:"cardName" ts:"string"`          // Name of the card for display purposes
	BehaviorIndex int             `json:"behaviorIndex" ts:"number"`     // Which behavior on the card this action represents
	Behavior      CardBehaviorDto `json:"behavior" ts:"CardBehaviorDto"` // The actual behavior definition with inputs/outputs
	PlayCount     int             `json:"playCount" ts:"number"`         // Number of times this action has been played this generation
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
	Corporation      *string           `json:"corporation" ts:"string | null"`
	Cards            []CardDto         `json:"cards" ts:"CardDto[]"`
	Resources        ResourcesDto      `json:"resources" ts:"ResourcesDto"`
	Production       ProductionDto     `json:"resourceProduction" ts:"ProductionDto"`
	TerraformRating  int               `json:"terraformRating" ts:"number"`
	PlayedCards      []string          `json:"playedCards" ts:"string[]"`
	Passed           bool              `json:"passed" ts:"boolean"`
	AvailableActions int               `json:"availableActions" ts:"number"`
	VictoryPoints    int               `json:"victoryPoints" ts:"number"`
	IsConnected      bool              `json:"isConnected" ts:"boolean"`
	Effects          []PlayerEffectDto `json:"effects" ts:"PlayerEffectDto[]"` // Active ongoing effects (discounts, special abilities, etc.)
	Actions          []PlayerActionDto `json:"actions" ts:"PlayerActionDto[]"` // Available actions from played cards with manual triggers

	SelectStartingCardsPhase *SelectStartingCardsPhaseDto `json:"selectStartingCardsPhase" ts:"SelectStartingCardsPhaseDto | null"`
	ProductionPhase          *ProductionPhaseDto          `json:"productionPhase" ts:"ProductionPhaseDto | null"`
}

// OtherPlayerDto represents another player from the viewing player's perspective (limited data)
type OtherPlayerDto struct {
	ID               string            `json:"id" ts:"string"`
	Name             string            `json:"name" ts:"string"`
	Status           PlayerStatus      `json:"status" ts:"PlayerStatus"`
	Corporation      string            `json:"corporation" ts:"string"`
	HandCardCount    int               `json:"handCardCount" ts:"number"` // Number of cards in hand (private)
	Resources        ResourcesDto      `json:"resources" ts:"ResourcesDto"`
	Production       ProductionDto     `json:"resourceProduction" ts:"ProductionDto"`
	TerraformRating  int               `json:"terraformRating" ts:"number"`
	PlayedCards      []string          `json:"playedCards" ts:"string[]"` // Played cards are public
	Passed           bool              `json:"passed" ts:"boolean"`
	AvailableActions int               `json:"availableActions" ts:"number"`
	VictoryPoints    int               `json:"victoryPoints" ts:"number"`
	IsConnected      bool              `json:"isConnected" ts:"boolean"`
	Effects          []PlayerEffectDto `json:"effects" ts:"PlayerEffectDto[]"`
	Actions          []PlayerActionDto `json:"actions" ts:"PlayerActionDto[]"`

	SelectStartingCardsPhase *SelectStartingCardsOtherPlayerDto `json:"selectStartingCardsPhase" ts:"SelectStartingCardsOtherPlayerDto | null"`
	ProductionPhase          *ProductionPhaseOtherPlayerDto     `json:"productionPhase" ts:"ProductionPhaseOtherPlayerDto | null"`
}

// GameDto represents a game for client consumption (clean architecture)
type GameDto struct {
	ID               string              `json:"id" ts:"string"`
	Status           GameStatus          `json:"status" ts:"GameStatus"`
	Settings         GameSettingsDto     `json:"settings" ts:"GameSettingsDto"`
	HostPlayerID     string              `json:"hostPlayerId" ts:"string"`
	CurrentPhase     GamePhase           `json:"currentPhase" ts:"GamePhase"`
	GlobalParameters GlobalParametersDto `json:"globalParameters" ts:"GlobalParametersDto"`
	CurrentPlayer    PlayerDto           `json:"currentPlayer" ts:"PlayerDto"`       // Viewing player's full data
	OtherPlayers     []OtherPlayerDto    `json:"otherPlayers" ts:"OtherPlayerDto[]"` // Other players' limited data
	ViewingPlayerID  string              `json:"viewingPlayerId" ts:"string"`        // The player viewing this game state
	CurrentTurn      *string             `json:"currentTurn" ts:"string|null"`       // Whose turn it is (nullable)
	Generation       int                 `json:"generation" ts:"number"`
	TurnOrder        []string            `json:"turnOrder" ts:"string[]"` // Turn order of all players in game
	Board            BoardDto            `json:"board" ts:"BoardDto"`     // Game board with tiles and occupancy state
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
