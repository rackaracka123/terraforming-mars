package model

// ResourceConditionType represents different types of resources in the game
// Simplified to only include actually used resource types
type ResourceConditionType string

const (
	// Basic resources - actually used in game
	ResourceCredits  ResourceConditionType = "credits"
	ResourceSteel    ResourceConditionType = "steel"
	ResourceTitanium ResourceConditionType = "titanium"
	ResourcePlants   ResourceConditionType = "plants"
	ResourceEnergy   ResourceConditionType = "energy"
	ResourceHeat     ResourceConditionType = "heat"

	// Card resources - for JSON compatibility only
	ResourceMicrobes ResourceConditionType = "microbes"
	ResourceAnimals  ResourceConditionType = "animals"
	ResourceFloaters ResourceConditionType = "floaters"
	ResourceScience  ResourceConditionType = "science"
	ResourceAsteroid ResourceConditionType = "asteroid"
	ResourceDisease  ResourceConditionType = "disease"
	ResourceCardDraw ResourceConditionType = "card-draw"
	ResourceCardTake ResourceConditionType = "card-take"
	ResourceCardPeek ResourceConditionType = "card-peek"

	// Terraforming - for JSON compatibility only
	ResourceCityPlacement     ResourceConditionType = "city-placement"
	ResourceOceanPlacement    ResourceConditionType = "ocean-placement"
	ResourceGreeneryPlacement ResourceConditionType = "greenery-placement"
	ResourceCityTile          ResourceConditionType = "city-tile"
	ResourceOceanTile         ResourceConditionType = "ocean-tile"
	ResourceGreeneryTile      ResourceConditionType = "greenery-tile"
	ResourceColonyTile        ResourceConditionType = "colony-tile"
	ResourceTemperature       ResourceConditionType = "temperature"
	ResourceOxygen            ResourceConditionType = "oxygen"
	ResourceVenus             ResourceConditionType = "venus"
	ResourceTR                ResourceConditionType = "tr"

	// Production - for JSON compatibility only
	ResourceCreditsProduction  ResourceConditionType = "credits-production"
	ResourceSteelProduction    ResourceConditionType = "steel-production"
	ResourceTitaniumProduction ResourceConditionType = "titanium-production"
	ResourcePlantsProduction   ResourceConditionType = "plants-production"
	ResourceEnergyProduction   ResourceConditionType = "energy-production"
	ResourceHeatProduction     ResourceConditionType = "heat-production"

	// Special effects - for JSON compatibility only
	ResourceEffect                  ResourceConditionType = "effect"
	ResourceTag                     ResourceConditionType = "tag"
	ResourceGlobalParameterLenience ResourceConditionType = "global-parameter-lenience"
	ResourceVenusLenience           ResourceConditionType = "venus-lenience"
	ResourceDefense                 ResourceConditionType = "defense"
	ResourceDiscount                ResourceConditionType = "discount"
	ResourceValueModifier           ResourceConditionType = "value-modifier"
)

// TriggerType represents different trigger conditions
// Kept for JSON compatibility only
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

// RequirementType represents different types of card requirements
// Used for card loading compatibility
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

// Requirement represents a single card requirement with flexible min/max values
type Requirement struct {
	Type     RequirementType        `json:"type" ts:"RequirementType"`
	Min      *int                   `json:"min,omitempty" ts:"number | undefined"`
	Max      *int                   `json:"max,omitempty" ts:"number | undefined"`
	Location *Location              `json:"location,omitempty" ts:"Location | undefined"`
	Tag      *CardTag               `json:"tag,omitempty" ts:"CardTag | undefined"`
	Resource *ResourceConditionType `json:"resource,omitempty" ts:"ResourceConditionType | undefined"`
}

// Simplified structures for JSON compatibility only
// These are loaded from JSON but not used in game logic

// CardBehavior represents card behavior (for JSON loading compatibility)
type CardBehavior struct {
	Triggers []Trigger           `json:"triggers,omitempty" ts:"Trigger[] | undefined"`
	Inputs   []ResourceCondition `json:"inputs,omitempty" ts:"ResourceCondition[] | undefined"`
	Outputs  []ResourceCondition `json:"outputs,omitempty" ts:"ResourceCondition[] | undefined"`
	Choices  []Choice            `json:"choices,omitempty" ts:"Choice[] | undefined"`
}

// ResourceStorage represents card resource storage (for JSON loading compatibility)
type ResourceStorage struct {
	Type     ResourceConditionType `json:"type" ts:"ResourceConditionType"`
	Capacity *int                  `json:"capacity,omitempty" ts:"number | undefined"`
	Starting int                   `json:"starting" ts:"number"`
}

// VictoryPointCondition represents VP conditions (for JSON loading compatibility)
type VictoryPointCondition struct {
	Amount     int             `json:"amount" ts:"number"`
	Condition  VPConditionType `json:"condition" ts:"VPConditionType"`
	MaxTrigger *int            `json:"maxTrigger,omitempty" ts:"number | undefined"`
	Per        *PerCondition   `json:"per,omitempty" ts:"PerCondition | undefined"`
}

type VPConditionType string

const (
	VPConditionPer   VPConditionType = "per"
	VPConditionOnce  VPConditionType = "once"
	VPConditionFixed VPConditionType = "fixed"
)

// Location represents different locations (for JSON loading compatibility)
type Location string

const (
	LocationAnywhere Location = "anywhere"
	LocationMars     Location = "mars"
)

// TargetType represents targeting scopes (for JSON loading compatibility)
type TargetType string

const (
	TargetSelfPlayer TargetType = "self-player"
	TargetSelfCard   TargetType = "self-card"
	TargetAnyPlayer  TargetType = "any-player"
	TargetOpponent   TargetType = "opponent"
	TargetAny        TargetType = "any"
	TargetNone       TargetType = "none"
)

// Supporting structures for JSON compatibility

type PerCondition struct {
	Type     ResourceConditionType `json:"type" ts:"ResourceConditionType"`
	Amount   int                   `json:"amount" ts:"number"`
	Location *Location             `json:"location,omitempty" ts:"Location | undefined"`
	Target   *TargetType           `json:"target,omitempty" ts:"TargetType | undefined"`
	Tag      *CardTag              `json:"tag,omitempty" ts:"CardTag | undefined"`
}

type Choice struct {
	Inputs  []ResourceCondition `json:"inputs,omitempty" ts:"ResourceCondition[] | undefined"`
	Outputs []ResourceCondition `json:"outputs,omitempty" ts:"ResourceCondition[] | undefined"`
}

type ResourceCondition struct {
	Type              ResourceConditionType `json:"type" ts:"ResourceConditionType"`
	Amount            int                   `json:"amount" ts:"number"`
	Target            TargetType            `json:"target" ts:"TargetType"`
	AffectedResources []string              `json:"affectedResources,omitempty" ts:"string[] | undefined"`
	AffectedTags      []CardTag             `json:"affectedTags,omitempty" ts:"CardTag[] | undefined"`
	Per               *PerCondition         `json:"per,omitempty" ts:"PerCondition | undefined"`
}

// Trigger-related structures for JSON compatibility

type ResourceTriggerType string

const (
	ResourceTriggerManual ResourceTriggerType = "manual"
	ResourceTriggerAuto   ResourceTriggerType = "auto"
)

type ResourceTriggerCondition struct {
	Type         TriggerType `json:"type" ts:"TriggerType"`
	Location     *Location   `json:"location,omitempty" ts:"Location | undefined"`
	AffectedTags []CardTag   `json:"affectedTags,omitempty" ts:"CardTag[] | undefined"`
}

type Trigger struct {
	Type      ResourceTriggerType       `json:"type" ts:"ResourceTriggerType"`
	Condition *ResourceTriggerCondition `json:"condition,omitempty" ts:"ResourceTriggerCondition | undefined"`
}
