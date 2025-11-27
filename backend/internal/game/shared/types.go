package shared

import "fmt"

// ==================== Resource Types ====================

// Resources represents a player's resources
type Resources struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// IsZero returns true if all resource values are zero
func (r Resources) IsZero() bool {
	return r.Credits == 0 && r.Steel == 0 && r.Titanium == 0 &&
		r.Plants == 0 && r.Energy == 0 && r.Heat == 0
}

// DeepCopy creates a deep copy of the Resources struct
func (r Resources) DeepCopy() Resources {
	return Resources{
		Credits:  r.Credits,
		Steel:    r.Steel,
		Titanium: r.Titanium,
		Plants:   r.Plants,
		Energy:   r.Energy,
		Heat:     r.Heat,
	}
}

// Production represents a player's production values
type Production struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// DeepCopy creates a deep copy of the Production struct
func (p Production) DeepCopy() Production {
	return Production{
		Credits:  p.Credits,
		Steel:    p.Steel,
		Titanium: p.Titanium,
		Plants:   p.Plants,
		Energy:   p.Energy,
		Heat:     p.Heat,
	}
}

// ResourceSet represents a collection of resources and their amounts
type ResourceSet struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// ResourceType represents different types of resources in the game
type ResourceType string

const (
	// Basic resources
	ResourceCredits  ResourceType = "credits"
	ResourceSteel    ResourceType = "steel"
	ResourceTitanium ResourceType = "titanium"
	ResourcePlants   ResourceType = "plants"
	ResourceEnergy   ResourceType = "energy"
	ResourceHeat     ResourceType = "heat"
	ResourceMicrobes ResourceType = "microbes"
	ResourceAnimals  ResourceType = "animals"
	ResourceFloaters ResourceType = "floaters"
	ResourceScience  ResourceType = "science"
	ResourceAsteroid ResourceType = "asteroid"
	ResourceDisease  ResourceType = "disease"

	// Card actions
	ResourceCardDraw ResourceType = "card-draw"
	ResourceCardTake ResourceType = "card-take"
	ResourceCardPeek ResourceType = "card-peek"
	ResourceCardBuy  ResourceType = "card-buy"

	// Terraforming actions
	ResourceCityPlacement     ResourceType = "city-placement"
	ResourceOceanPlacement    ResourceType = "ocean-placement"
	ResourceGreeneryPlacement ResourceType = "greenery-placement"

	// Tile counting
	ResourceCityTile     ResourceType = "city-tile"
	ResourceOceanTile    ResourceType = "ocean-tile"
	ResourceGreeneryTile ResourceType = "greenery-tile"
	ResourceColonyTile   ResourceType = "colony-tile"

	// Global parameters
	ResourceTemperature     ResourceType = "temperature"
	ResourceOxygen          ResourceType = "oxygen"
	ResourceOceans          ResourceType = "oceans"
	ResourceVenus           ResourceType = "venus"
	ResourceTR              ResourceType = "tr"
	ResourceGlobalParameter ResourceType = "global-parameter"

	// Production resources
	ResourceCreditsProduction  ResourceType = "credits-production"
	ResourceSteelProduction    ResourceType = "steel-production"
	ResourceTitaniumProduction ResourceType = "titanium-production"
	ResourcePlantsProduction   ResourceType = "plants-production"
	ResourceEnergyProduction   ResourceType = "energy-production"
	ResourceHeatProduction     ResourceType = "heat-production"
	ResourceAnyProduction      ResourceType = "any-production"

	// Effect type
	ResourceEffect ResourceType = "effect"

	// Tag counting
	ResourceTag ResourceType = "tag"

	// Special ongoing effects
	ResourceGlobalParameterLenience ResourceType = "global-parameter-lenience"
	ResourceVenusLenience           ResourceType = "venus-lenience"
	ResourceDefense                 ResourceType = "defense"
	ResourceDiscount                ResourceType = "discount"
	ResourceValueModifier           ResourceType = "value-modifier"
	ResourcePaymentSubstitute       ResourceType = "payment-substitute"
	ResourceOceanAdjacencyBonus     ResourceType = "ocean-adjacency-bonus"
)

// ==================== Card Tag Types ====================

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

// ==================== Standard Projects ====================

// StandardProject represents the different types of standard projects
type StandardProject string

const (
	StandardProjectSellPatents              StandardProject = "sell-patents"
	StandardProjectPowerPlant               StandardProject = "power-plant"
	StandardProjectAsteroid                 StandardProject = "asteroid"
	StandardProjectAquifer                  StandardProject = "aquifer"
	StandardProjectGreenery                 StandardProject = "greenery"
	StandardProjectCity                     StandardProject = "city"
	StandardProjectConvertPlantsToGreenery  StandardProject = "convert-plants-to-greenery"
	StandardProjectConvertHeatToTemperature StandardProject = "convert-heat-to-temperature"
)

// StandardProjectCost represents the credit cost for each standard project
var StandardProjectCost = map[StandardProject]int{
	StandardProjectSellPatents: 0,
	StandardProjectPowerPlant:  11,
	StandardProjectAsteroid:    14,
	StandardProjectAquifer:     18,
	StandardProjectGreenery:    23,
	StandardProjectCity:        25,
}

// ==================== Hex Position Types ====================

// HexPosition represents a position on the Mars board using cube coordinates
type HexPosition struct {
	Q int `json:"q"` // Column coordinate
	R int `json:"r"` // Row coordinate
	S int `json:"s"` // Third coordinate (Q + R + S = 0)
}

// String returns a string representation of the hex position
func (h HexPosition) String() string {
	return fmt.Sprintf("%d,%d,%d", h.Q, h.R, h.S)
}

// GetNeighbors returns all 6 adjacent hex positions using cube coordinate system
func (h HexPosition) GetNeighbors() []HexPosition {
	directions := []HexPosition{
		{Q: 1, R: -1, S: 0}, // East
		{Q: 1, R: 0, S: -1}, // Northeast
		{Q: 0, R: 1, S: -1}, // Northwest
		{Q: -1, R: 1, S: 0}, // West
		{Q: -1, R: 0, S: 1}, // Southwest
		{Q: 0, R: -1, S: 1}, // Southeast
	}

	neighbors := make([]HexPosition, 0, 6)
	for _, dir := range directions {
		neighbors = append(neighbors, HexPosition{
			Q: h.Q + dir.Q,
			R: h.R + dir.R,
			S: h.S + dir.S,
		})
	}

	return neighbors
}

// Equals checks if two hex positions are equal
func (h HexPosition) Equals(other HexPosition) bool {
	return h.Q == other.Q && h.R == other.R && h.S == other.S
}

// ==================== Payment and Modifier Types ====================

// PaymentSubstitute represents a resource type substitution for payments
type PaymentSubstitute struct {
	ResourceType   ResourceType
	ConversionRate int
}

// RequirementModifier represents a modification to requirements
type RequirementModifier struct {
	Amount                int
	AffectedResources     []ResourceType
	CardTarget            *string
	StandardProjectTarget *StandardProject
}
