package domain

// ProjectCard represents a project card in Terraforming Mars
type ProjectCard struct {
	ID           string          `json:"id" ts:"string"`
	Name         string          `json:"name" ts:"string"`
	Type         CardType        `json:"type" ts:"CardType"`
	Cost         int             `json:"cost" ts:"number"`
	Tags         []Tag           `json:"tags" ts:"Tag[]"`
	Requirements []Requirement   `json:"requirements" ts:"Requirement[]"`
	Effects      []CardEffect    `json:"effects" ts:"CardEffect[]"`
	VictoryPoints *int           `json:"victoryPoints,omitempty" ts:"number | undefined"`
	Description  string          `json:"description" ts:"string"`
	FlavorText   *string         `json:"flavorText,omitempty" ts:"string | undefined"`
	ImagePath    string          `json:"imagePath" ts:"string"`
}

// CardType represents the three types of project cards
type CardType string

const (
	CardTypeAutomated CardType = "automated" // Green cards - immediate effects, go into tableau
	CardTypeEvent     CardType = "event"     // Red cards - one-time effects, discarded after use
	CardTypeActive    CardType = "active"    // Blue cards - ongoing effects or repeatable actions
)

// CardEffect represents various effects that cards can have
type CardEffect struct {
	Type        EffectType       `json:"type" ts:"EffectType"`
	Target      EffectTarget     `json:"target" ts:"EffectTarget"`
	Amount      *int             `json:"amount,omitempty" ts:"number | undefined"`
	ResourceType *ResourceType   `json:"resourceType,omitempty" ts:"ResourceType | undefined"`
	TileType    *TileType        `json:"tileType,omitempty" ts:"TileType | undefined"`
	Condition   *EffectCondition `json:"condition,omitempty" ts:"EffectCondition | undefined"`
	Trigger     *EffectTrigger   `json:"trigger,omitempty" ts:"EffectTrigger | undefined"`
}

// EffectType defines the type of effect
type EffectType string

const (
	EffectTypeGainResource      EffectType = "gain_resource"
	EffectTypeGainProduction    EffectType = "gain_production"
	EffectTypeLoseResource      EffectType = "lose_resource"
	EffectTypeLoseProduction    EffectType = "lose_production"
	EffectTypeIncreaseParameter EffectType = "increase_parameter"
	EffectTypePlaceTile         EffectType = "place_tile"
	EffectTypeDrawCards         EffectType = "draw_cards"
	EffectTypeDiscountCards     EffectType = "discount_cards"
	EffectTypeGainVictoryPoints EffectType = "gain_victory_points"
	EffectTypeAction            EffectType = "action"              // For blue cards with action abilities
	EffectTypeTrigger           EffectType = "trigger"             // For ongoing triggered effects
)

// EffectTarget defines who or what the effect targets
type EffectTarget string

const (
	EffectTargetSelf        EffectTarget = "self"
	EffectTargetAllPlayers  EffectTarget = "all_players"
	EffectTargetOthers      EffectTarget = "others"
	EffectTargetAnyPlayer   EffectTarget = "any_player"
	EffectTargetGlobal      EffectTarget = "global"
)

// EffectCondition represents conditions that must be met for effects
type EffectCondition struct {
	Type         ConditionType   `json:"type" ts:"ConditionType"`
	Parameter    *GlobalParam    `json:"parameter,omitempty" ts:"GlobalParam | undefined"`
	MinValue     *int            `json:"minValue,omitempty" ts:"number | undefined"`
	MaxValue     *int            `json:"maxValue,omitempty" ts:"number | undefined"`
	ResourceType *ResourceType   `json:"resourceType,omitempty" ts:"ResourceType | undefined"`
	Tag          *Tag            `json:"tag,omitempty" ts:"Tag | undefined"`
	TileType     *TileType       `json:"tileType,omitempty" ts:"TileType | undefined"`
}

// ConditionType defines types of conditions for card effects
type ConditionType string

const (
	ConditionTypeParameter         ConditionType = "parameter"          // Global parameter requirement
	ConditionTypeResourceAmount    ConditionType = "resource_amount"    // Player must have X resources
	ConditionTypeProductionAmount  ConditionType = "production_amount"  // Player must have X production
	ConditionTypeTagCount          ConditionType = "tag_count"          // Player must have X tags
	ConditionTypeTileCount         ConditionType = "tile_count"         // Player must have X tiles
	ConditionTypeOceanTiles        ConditionType = "ocean_tiles"        // Must have X ocean tiles on board
	ConditionTypePlayerCount       ConditionType = "player_count"       // Game must have X players
)

// EffectTrigger defines when triggered effects activate
type EffectTrigger struct {
	Event       TriggerEvent    `json:"event" ts:"TriggerEvent"`
	Tag         *Tag            `json:"tag,omitempty" ts:"Tag | undefined"`
	TileType    *TileType       `json:"tileType,omitempty" ts:"TileType | undefined"`
	Parameter   *GlobalParam    `json:"parameter,omitempty" ts:"GlobalParam | undefined"`
}

// TriggerEvent defines events that can trigger card effects
type TriggerEvent string

const (
	TriggerEventCardPlayed      TriggerEvent = "card_played"
	TriggerEventTilePlaced      TriggerEvent = "tile_placed"
	TriggerEventParameterRaised TriggerEvent = "parameter_raised"
	TriggerEventProductionPhase TriggerEvent = "production_phase"
	TriggerEventGenerationEnd   TriggerEvent = "generation_end"
	TriggerEventGameEnd         TriggerEvent = "game_end"
)

// Requirement represents requirements that must be met to play a card
type Requirement struct {
	Type      RequirementType `json:"type" ts:"RequirementType"`
	Parameter *GlobalParam    `json:"parameter,omitempty" ts:"GlobalParam | undefined"`
	MinValue  *int            `json:"minValue,omitempty" ts:"number | undefined"`
	MaxValue  *int            `json:"maxValue,omitempty" ts:"number | undefined"`
	Tag       *Tag            `json:"tag,omitempty" ts:"Tag | undefined"`
	Count     *int            `json:"count,omitempty" ts:"number | undefined"`
}

// RequirementType defines types of requirements for playing cards
type RequirementType string

const (
	RequirementTypeTemperature RequirementType = "temperature"
	RequirementTypeOxygen      RequirementType = "oxygen"
	RequirementTypeOceans      RequirementType = "oceans"
	RequirementTypeTagCount    RequirementType = "tag_count"
	RequirementTypeProduction  RequirementType = "production"
	RequirementTypeResource    RequirementType = "resource"
)

// GlobalParam represents global parameters for requirements and conditions
type GlobalParam string

const (
	GlobalParamTemperature GlobalParam = "temperature"
	GlobalParamOxygen      GlobalParam = "oxygen"  
	GlobalParamOceans      GlobalParam = "oceans"
)