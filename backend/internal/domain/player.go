package domain

// Player represents a player in the Terraforming Mars game
type Player struct {
	ID                string            `json:"id" ts:"string"`
	Name              string            `json:"name" ts:"string"`
	Corporation       *string           `json:"corporation,omitempty" ts:"string | undefined"`
	Resources         ResourcesMap      `json:"resources" ts:"ResourcesMap"`
	Production        ResourcesMap      `json:"production" ts:"ResourcesMap"`
	TerraformRating   int               `json:"terraformRating" ts:"number"`
	VictoryPoints     int               `json:"victoryPoints" ts:"number"`
	PlayedCards       []string          `json:"playedCards" ts:"string[]"`
	Hand              []string          `json:"hand" ts:"string[]"`
	AvailableActions  int               `json:"availableActions" ts:"number"`
	Tags              []Tag             `json:"tags" ts:"Tag[]"`
	ActionsTaken      int               `json:"actionsTaken" ts:"number"`
	ActionsRemaining  int               `json:"actionsRemaining" ts:"number"`
	Passed            *bool             `json:"passed,omitempty" ts:"boolean | undefined"`
	TilePositions     []HexCoordinate   `json:"tilePositions" ts:"HexCoordinate[]"`
	Reserved          ResourcesMap      `json:"reserved" ts:"ResourcesMap"`
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