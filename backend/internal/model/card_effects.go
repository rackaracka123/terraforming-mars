package model

// ResourceConditionType represents different types of resources in the game
type ResourceConditionType string

const (
	ResourceCredits  ResourceConditionType = "credits"
	ResourceSteel    ResourceConditionType = "steel"
	ResourceTitanium ResourceConditionType = "titanium"
	ResourcePlants   ResourceConditionType = "plants"
	ResourceEnergy   ResourceConditionType = "energy"
	ResourceHeat     ResourceConditionType = "heat"
	ResourceMicrobes ResourceConditionType = "microbes"
	ResourceAnimals  ResourceConditionType = "animals"
	ResourceFloaters ResourceConditionType = "floaters"
	ResourceScience  ResourceConditionType = "science"
	ResourceAsteroid ResourceConditionType = "asteroid"
	ResourceDisease  ResourceConditionType = "disease"
	ResourceCardDraw ResourceConditionType = "card-draw" // Drawing cards (take = peek, common case)
	ResourceCardTake ResourceConditionType = "card-take" // Drawing cards (take from deck)
	ResourceCardPeek ResourceConditionType = "card-peek" // Looking at cards (peek without taking all)

	// Terraforming actions
	ResourceCityPlacement     ResourceConditionType = "city-placement"     // Placing city tiles
	ResourceOceanPlacement    ResourceConditionType = "ocean-placement"    // Placing ocean tiles
	ResourceGreeneryPlacement ResourceConditionType = "greenery-placement" // Placing greenery tiles

	// Tile counting (for per conditions)
	ResourceCityTile     ResourceConditionType = "city-tile"     // Counting existing city tiles
	ResourceOceanTile    ResourceConditionType = "ocean-tile"    // Counting existing ocean tiles
	ResourceGreeneryTile ResourceConditionType = "greenery-tile" // Counting existing greenery tiles
	ResourceColonyTile   ResourceConditionType = "colony-tile"   // Counting existing colonies

	// Global parameters
	ResourceTemperature ResourceConditionType = "temperature" // Temperature change
	ResourceOxygen      ResourceConditionType = "oxygen"      // Oxygen change
	ResourceVenus       ResourceConditionType = "venus"       // Venus change
	ResourceTR          ResourceConditionType = "tr"          // Terraform Rating change

	// Production resources (for spending production)
	ResourceCreditsProduction  ResourceConditionType = "credits-production"  // Credits production
	ResourceSteelProduction    ResourceConditionType = "steel-production"    // Steel production
	ResourceTitaniumProduction ResourceConditionType = "titanium-production" // Titanium production
	ResourcePlantsProduction   ResourceConditionType = "plants-production"   // Plants production
	ResourceEnergyProduction   ResourceConditionType = "energy-production"   // Energy production
	ResourceHeatProduction     ResourceConditionType = "heat-production"     // Heat production

	// Effect type (for triggered effects like Rover Construction)
	ResourceEffect ResourceConditionType = "effect" // Triggered effect

	// Tag counting (for VP conditions like "1 VP per jovian tag")
	ResourceTag ResourceConditionType = "tag" // Count tags of a specific type

	// Special ongoing effects
	ResourceGlobalParameterLenience ResourceConditionType = "global-parameter-lenience" // Global parameter requirement flexibility
	ResourceVenusLenience           ResourceConditionType = "venus-lenience"            // Venus requirement flexibility (+/- steps)
	ResourceDefense                 ResourceConditionType = "defense"                   // Protection from attacks or resource removal
	ResourceDiscount                ResourceConditionType = "discount"                  // Cost reduction for playing cards
	ResourceValueModifier           ResourceConditionType = "value-modifier"            // Increases resource values (e.g., steel/titanium worth more)
)

// TriggerType represents different trigger conditions
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

// TerraformingActions represents tile placement actions
type TerraformingActions struct {
	CityPlacement     int `json:"cityPlacement,omitempty" ts:"number"`     // Number of city tiles to place
	OceanPlacement    int `json:"oceanPlacement,omitempty" ts:"number"`    // Number of ocean tiles to place
	GreeneryPlacement int `json:"greeneryPlacement,omitempty" ts:"number"` // Number of greenery tiles to place
}

// RequirementType represents different types of card requirements
type RequirementType string

const (
	RequirementTemperature RequirementType = "temperature" // Global temperature requirement
	RequirementOxygen      RequirementType = "oxygen"      // Global oxygen requirement
	RequirementOceans      RequirementType = "oceans"      // Ocean tiles requirement
	RequirementVenus       RequirementType = "venus"       // Venus terraforming requirement
	RequirementCities      RequirementType = "cities"      // City tiles requirement
	RequirementGreeneries  RequirementType = "greeneries"  // Greenery tiles requirement
	RequirementTags        RequirementType = "tags"        // Tag requirement (e.g., science tags)
	RequirementProduction  RequirementType = "production"  // Production requirement
	RequirementTR          RequirementType = "tr"          // Terraform Rating requirement
	RequirementResource    RequirementType = "resource"    // Resource requirement (e.g., floaters, microbes)
)

// Requirement represents a single card requirement with flexible min/max values
type Requirement struct {
	Type     RequirementType        `json:"type" ts:"RequirementType"`                                 // Type of requirement
	Min      *int                   `json:"min,omitempty" ts:"number | undefined"`                     // Minimum value required
	Max      *int                   `json:"max,omitempty" ts:"number | undefined"`                     // Maximum value allowed
	Location *Location              `json:"location,omitempty" ts:"Location | undefined"`              // Location constraint (Mars, anywhere, etc.)
	Tag      *CardTag               `json:"tag,omitempty" ts:"CardTag | undefined"`                    // For tag requirements: which tag
	Resource *ResourceConditionType `json:"resource,omitempty" ts:"ResourceConditionType | undefined"` // For production: which resource
}

// CardBehavior represents any card behavior - both immediate (when played) and repeatable (activated by player)
// The trigger type determines when it executes: auto = immediate, manual = repeatable
type CardBehavior struct {
	Triggers []Trigger           `json:"triggers,omitempty" ts:"Trigger[] | undefined"`          // When/how this action is activated
	Inputs   []ResourceCondition `json:"inputs,omitempty" ts:"ResourceCondition[] | undefined"`  // Resources spent (input side of arrow)
	Outputs  []ResourceCondition `json:"outputs,omitempty" ts:"ResourceCondition[] | undefined"` // Resources gained (output side of arrow)
	Choices  []Choice            `json:"choices,omitempty" ts:"Choice[] | undefined"`            // Player choices between different input/output combinations
}

// ResourceStorage represents a card's ability to hold resources
type ResourceStorage struct {
	Type     ResourceConditionType `json:"type" ts:"ResourceConditionType"`            // Type of resource stored
	Capacity *int                  `json:"capacity,omitempty" ts:"number | undefined"` // Max capacity (if limited)
	Starting int                   `json:"starting" ts:"number"`                       // Starting amount
}

// VictoryPointCondition represents a VP condition like "1 VP per jovian tag"
type VictoryPointCondition struct {
	Amount     int             `json:"amount" ts:"number"`                           // VP awarded
	Condition  VPConditionType `json:"condition" ts:"VPConditionType"`               // Type of condition
	MaxTrigger *int            `json:"maxTrigger,omitempty" ts:"number | undefined"` // Max times it can trigger (-1 = unlimited), only for "per" conditions
	Per        *PerCondition   `json:"per,omitempty" ts:"PerCondition | undefined"`  // Per condition details, only for "per" conditions
}

// VPConditionType represents different types of VP conditions
type VPConditionType string

const (
	VPConditionPer   VPConditionType = "per"   // VP per resource/tag
	VPConditionOnce  VPConditionType = "once"  // VP awarded once when condition met
	VPConditionFixed VPConditionType = "fixed" // Fixed VP amount (no condition)
)

// Location represents different locations where conditions can be evaluated
type Location string

const (
	LocationAnywhere Location = "anywhere" // No location restriction
	LocationMars     Location = "mars"     // On Mars only
)

// DiscountEffect represents cost reductions for playing cards
type DiscountEffect struct {
	Amount      int       `json:"amount" ts:"number"`                        // M€ discount per qualifying tag
	Tags        []CardTag `json:"tags,omitempty" ts:"CardTag[] | undefined"` // Tags that qualify for discount (empty = all cards)
	Description string    `json:"description" ts:"string"`                   // Human readable description
}

// PlayerEffectType represents different types of ongoing effects a player can have
type PlayerEffectType string

const (
	PlayerEffectDiscount                PlayerEffectType = "discount"                  // Cost reduction for playing cards
	PlayerEffectGlobalParameterLenience PlayerEffectType = "global-parameter-lenience" // Global parameter requirement flexibility
	PlayerEffectDefense                 PlayerEffectType = "defense"                   // Protection from attacks or resource removal
	PlayerEffectValueModifier           PlayerEffectType = "value-modifier"            // Increases resource values (e.g., steel/titanium worth more)
)

// PlayerEffect represents ongoing effects that a player has active
type PlayerEffect struct {
	Type         PlayerEffectType `json:"type" ts:"PlayerEffectType"`                        // Type of effect
	Amount       int              `json:"amount" ts:"number"`                                // Effect amount (e.g., M€ discount, steps of flexibility)
	AffectedTags []CardTag        `json:"affectedTags,omitempty" ts:"CardTag[] | undefined"` // Tags that qualify for this effect (empty = all cards)
}

// TargetType represents different targeting scopes for resource conditions
type TargetType string

const (
	TargetSelfPlayer TargetType = "self-player" // Targets the player who played/owns the card
	TargetSelfCard   TargetType = "self-card"   // Targets the specific card itself
	TargetAnyPlayer  TargetType = "any-player"  // Can target any player
	TargetOpponent   TargetType = "opponent"    // Targets opponent players
	TargetAny        TargetType = "any"         // No specific target or global scope
	TargetNone       TargetType = "none"        // No target (e.g., global parameter changes)
)

// PerCondition represents what to count for conditional resource gains
type PerCondition struct {
	Type     ResourceConditionType `json:"type" ts:"ResourceConditionType"`              // What to count (city-tile, ocean-tile, etc.)
	Amount   int                   `json:"amount" ts:"number"`                           // How many of the counted thing per gain
	Location *Location             `json:"location,omitempty" ts:"Location | undefined"` // Location constraint (Mars, anywhere, etc.)
	Target   *TargetType           `json:"target,omitempty" ts:"TargetType | undefined"` // Whose tags/resources to count (self-player, any-player, etc.)
	Tag      *CardTag              `json:"tag,omitempty" ts:"CardTag | undefined"`       // For tag-based VP conditions (jovian tag, science tag, etc.)
}

// Choice represents a single choice option with inputs and outputs
type Choice struct {
	Inputs  []ResourceCondition `json:"inputs,omitempty" ts:"ResourceCondition[] | undefined"`  // Resources spent for this choice
	Outputs []ResourceCondition `json:"outputs,omitempty" ts:"ResourceCondition[] | undefined"` // Resources gained from this choice
}

// ResourceCondition represents a resource amount (input or output)
type ResourceCondition struct {
	Type              ResourceConditionType `json:"type" ts:"ResourceConditionType"`                       // Type of resource
	Amount            int                   `json:"amount" ts:"number"`                                    // Amount of resource
	Target            TargetType            `json:"target" ts:"TargetType"`                                // Target for this resource condition
	AffectedResources []string              `json:"affectedResources,omitempty" ts:"string[] | undefined"` // For defense: resources being protected
	AffectedTags      []CardTag             `json:"affectedTags,omitempty" ts:"CardTag[] | undefined"`     // For discount: tags qualifying for discount
	Per               *PerCondition         `json:"per,omitempty" ts:"PerCondition | undefined"`           // For conditional gains: what to count
}

// ResourceTriggerType represents different trigger types for resource exchanges
type ResourceTriggerType string

const (
	ResourceTriggerManual ResourceTriggerType = "manual" // Manual activation (actions)
	ResourceTriggerAuto   ResourceTriggerType = "auto"   // Automatic activation (effects, immediate)
)

// ResourceTriggerCondition represents what triggers an automatic resource exchange
type ResourceTriggerCondition struct {
	Type         TriggerType `json:"type" ts:"TriggerType"`                             // What triggers this (onCityPlaced, etc.)
	Location     *Location   `json:"location,omitempty" ts:"Location | undefined"`      // Where the trigger applies (mars, anywhere)
	AffectedTags []CardTag   `json:"affectedTags,omitempty" ts:"CardTag[] | undefined"` // Tags that trigger this effect
}

// Trigger represents when and how an action or effect is activated
type Trigger struct {
	Type      ResourceTriggerType       `json:"type" ts:"ResourceTriggerType"`                                 // Manual or auto activation
	Condition *ResourceTriggerCondition `json:"condition,omitempty" ts:"ResourceTriggerCondition | undefined"` // What triggers auto actions
}

// ResourceExchange represents a directional resource trade (input → output)
type ResourceExchange struct {
	Triggers []Trigger           `json:"triggers,omitempty" ts:"Trigger[] | undefined"`          // When/how this exchange is activated
	Inputs   []ResourceCondition `json:"inputs,omitempty" ts:"ResourceCondition[] | undefined"`  // Resources spent (input side of arrow)
	Outputs  []ResourceCondition `json:"outputs,omitempty" ts:"ResourceCondition[] | undefined"` // Resources gained (output side of arrow)
}
