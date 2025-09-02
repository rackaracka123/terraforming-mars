package domain

// Player represents a player in the Terraforming Mars game
type Player struct {
	ID                  string                 `json:"id" ts:"string"`
	Name                string                 `json:"name" ts:"string"`
	Corporation         *string                `json:"corporation,omitempty" ts:"string | undefined"`
	Resources           ResourcesMap           `json:"resources" ts:"ResourcesMap"`
	Production          ResourcesMap           `json:"production" ts:"ResourcesMap"`
	TerraformRating     int                    `json:"terraformRating" ts:"number"`
	VictoryPoints       int                    `json:"victoryPoints" ts:"number"`
	VictoryPointSources []VictoryPointSource   `json:"victoryPointSources" ts:"VictoryPointSource[]"`
	PlayedCards         []string               `json:"playedCards" ts:"string[]"`
	Hand                []string               `json:"hand" ts:"string[]"`
	AvailableActions    int                    `json:"availableActions" ts:"number"`
	Tags                []Tag                  `json:"tags" ts:"Tag[]"`
	TagCounts           map[Tag]int            `json:"tagCounts" ts:"Record<Tag, number>"`
	ActionsTaken        int                    `json:"actionsTaken" ts:"number"`
	ActionsRemaining    int                    `json:"actionsRemaining" ts:"number"`
	Passed              *bool                  `json:"passed,omitempty" ts:"boolean | undefined"`
	TilePositions       []HexCoordinate        `json:"tilePositions" ts:"HexCoordinate[]"`
	TileCounts          map[TileType]int       `json:"tileCounts" ts:"Record<TileType, number>"`
	Reserved            ResourcesMap           `json:"reserved" ts:"ResourcesMap"`
	ClaimedMilestones   []string               `json:"claimedMilestones" ts:"string[]"`
	FundedAwards        []string               `json:"fundedAwards" ts:"string[]"`
	HandLimit           int                    `json:"handLimit" ts:"number"`
}

// ResourcesMap represents the six resource types in Terraforming Mars
type ResourcesMap struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// ResourceType represents a single resource type
type ResourceType string

const (
	ResourceTypeCredits  ResourceType = "credits"
	ResourceTypeSteel    ResourceType = "steel"
	ResourceTypeTitanium ResourceType = "titanium"
	ResourceTypePlants   ResourceType = "plants"
	ResourceTypeEnergy   ResourceType = "energy"
	ResourceTypeHeat     ResourceType = "heat"
)

// GlobalParameters represents the three global parameters that players terraform
type GlobalParameters struct {
	Temperature int `json:"temperature" ts:"number"` // -30 to +8 (in steps of 2)
	Oxygen      int `json:"oxygen" ts:"number"`      // 0 to 14%
	Oceans      int `json:"oceans" ts:"number"`      // 0 to 9
}

// Tag represents card tags that provide various bonuses and interactions
type Tag string

const (
	TagBuilding     Tag = "building"
	TagSpace        Tag = "space"
	TagPower        Tag = "power"
	TagScience      Tag = "science"
	TagMicrobe      Tag = "microbe"
	TagAnimal       Tag = "animal"
	TagPlant        Tag = "plant"
	TagEarth        Tag = "earth"
	TagJovian       Tag = "jovian"
	TagCity         Tag = "city"
	TagEvent        Tag = "event"
	TagWild         Tag = "wild"
	TagVenus        Tag = "venus"
	TagMoon         Tag = "moon"
	TagMars         Tag = "mars"
	TagClone        Tag = "clone"
)

// HexCoordinate represents a position on the hexagonal Mars board using cube coordinates
type HexCoordinate struct {
	Q int `json:"q" ts:"number"` // Column
	R int `json:"r" ts:"number"` // Row
	S int `json:"s" ts:"number"` // Diagonal (q + r + s = 0)
}

// TileType represents different types of tiles that can be placed on Mars
type TileType string

const (
	TileTypeGreenery         TileType = "greenery"
	TileTypeCity             TileType = "city"
	TileTypeOcean            TileType = "ocean"
	TileTypeSpecialCity      TileType = "special_city"
	TileTypeCapitalCity      TileType = "capital_city"
	TileTypeCommercialCenter TileType = "commercial_center"
	TileTypeEcologicalZone   TileType = "ecological_zone"
	TileTypeIndustrialCenter TileType = "industrial_center"
	TileTypeLavaFlows        TileType = "lava_flows"
	TileTypeMiningArea       TileType = "mining_area"
	TileTypeMiningRights     TileType = "mining_rights"
	TileTypeNaturalPreserve  TileType = "natural_preserve"
	TileTypeNuclearZone      TileType = "nuclear_zone"
	TileTypeRestrictedArea   TileType = "restricted_area"
)

// Tile represents a tile placed on the Mars board
type Tile struct {
	Type        TileType      `json:"type" ts:"TileType"`
	Position    HexCoordinate `json:"position" ts:"HexCoordinate"`
	PlayerID    *string       `json:"playerId,omitempty" ts:"string | undefined"`
	Bonus       []ResourceType `json:"bonus" ts:"ResourceType[]"`
	IsReserved  bool          `json:"isReserved" ts:"boolean"`
}

// ResourceConversion represents resource conversion rules
type ResourceConversion struct {
	From   ResourceType `json:"from" ts:"ResourceType"`
	To     ResourceType `json:"to" ts:"ResourceType"`
	Rate   int          `json:"rate" ts:"number"`   // How many From resources needed
	Gain   int          `json:"gain" ts:"number"`   // How many To resources gained
	Effect *string      `json:"effect,omitempty" ts:"string | undefined"` // Additional effects (e.g., TR increase)
}

// ResourceMultiplier represents discounts for specific card types
type ResourceMultiplier struct {
	ResourceType ResourceType `json:"resourceType" ts:"ResourceType"`
	CardTag      Tag          `json:"cardTag" ts:"Tag"`
	Value        int          `json:"value" ts:"number"` // How much each resource is worth
	Description  string       `json:"description" ts:"string"`
}

// GetResourceConversions returns available resource conversions
func GetResourceConversions() []ResourceConversion {
	return []ResourceConversion{
		{
			From:   ResourceTypeHeat,
			To:     ResourceTypeCredits, // Represents temperature increase effect
			Rate:   8,
			Gain:   1,
			Effect: stringPtr("increase_temperature"),
		},
		{
			From:   ResourceTypePlants,
			To:     ResourceTypeCredits, // Represents greenery placement effect
			Rate:   8,
			Gain:   1,
			Effect: stringPtr("place_greenery"),
		},
		{
			From:   ResourceTypeEnergy,
			To:     ResourceTypeHeat,
			Rate:   1,
			Gain:   1,
			Effect: nil, // Automatic conversion during production phase
		},
	}
}

// GetResourceMultipliers returns standard resource multipliers
func GetResourceMultipliers() []ResourceMultiplier {
	return []ResourceMultiplier{
		{
			ResourceType: ResourceTypeSteel,
			CardTag:      TagBuilding,
			Value:        2,
			Description:  "Steel is worth 2 M€ when paying for building cards",
		},
		{
			ResourceType: ResourceTypeTitanium,
			CardTag:      TagSpace,
			Value:        3,
			Description:  "Titanium is worth 3 M€ when paying for space cards",
		},
	}
}

// VictoryPointSource represents different sources of victory points
type VictoryPointSource struct {
	Type        VPSourceType `json:"type" ts:"VPSourceType"`
	Points      int          `json:"points" ts:"number"`
	Description string       `json:"description" ts:"string"`
	Details     *string      `json:"details,omitempty" ts:"string | undefined"`
}

// VPSourceType defines different sources of victory points
type VPSourceType string

const (
	VPSourceTypeTerraformRating VPSourceType = "terraform_rating"
	VPSourceTypeMilestone       VPSourceType = "milestone"
	VPSourceTypeAward           VPSourceType = "award"
	VPSourceTypeCard            VPSourceType = "card"
	VPSourceTypeGreenery        VPSourceType = "greenery"
	VPSourceTypeCity            VPSourceType = "city"
)

// EndGameCondition represents conditions that must be met to end the game
type EndGameCondition struct {
	Parameter   GlobalParam `json:"parameter" ts:"GlobalParam"`
	TargetValue int         `json:"targetValue" ts:"number"`
	CurrentValue int        `json:"currentValue" ts:"number"`
	IsCompleted bool        `json:"isCompleted" ts:"boolean"`
}

// GetEndGameConditions returns the three terraforming goals
func GetEndGameConditions() []EndGameCondition {
	return []EndGameCondition{
		{
			Parameter:    GlobalParamTemperature,
			TargetValue:  8,
			CurrentValue: -30,
			IsCompleted:  false,
		},
		{
			Parameter:    GlobalParamOxygen,
			TargetValue:  14,
			CurrentValue: 0,
			IsCompleted:  false,
		},
		{
			Parameter:    GlobalParamOceans,
			TargetValue:  9,
			CurrentValue: 0,
			IsCompleted:  false,
		},
	}
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}