package domain

// ResourceType represents the different types of resources in Terraforming Mars
type ResourceType string

const (
	ResourceTypeCredits  ResourceType = "credits"
	ResourceTypeSteel    ResourceType = "steel"
	ResourceTypeTitanium ResourceType = "titanium"
	ResourceTypePlants   ResourceType = "plants"
	ResourceTypeEnergy   ResourceType = "energy"
	ResourceTypeHeat     ResourceType = "heat"
)

// Resource type constants (without "Type" prefix for backward compatibility)
const (
	ResourceCredits  = ResourceTypeCredits
	ResourceSteel    = ResourceTypeSteel
	ResourceTitanium = ResourceTypeTitanium
	ResourcePlants   = ResourceTypePlants
	ResourceEnergy   = ResourceTypeEnergy
	ResourceHeat     = ResourceTypeHeat
)

// Production type constants
const (
	ResourceCreditsProduction  ResourceType = "credits-production"
	ResourceSteelProduction    ResourceType = "steel-production"
	ResourceTitaniumProduction ResourceType = "titanium-production"
	ResourcePlantsProduction   ResourceType = "plants-production"
	ResourceEnergyProduction   ResourceType = "energy-production"
	ResourceHeatProduction     ResourceType = "heat-production"
)

// Tile type constants
const (
	ResourceCityTile     ResourceType = "city-tile"
	ResourceGreeneryTile ResourceType = "greenery-tile"
	ResourceOceanTile    ResourceType = "ocean-tile"
)

// Special bonus types
const (
	ResourceCardDraw ResourceType = "card-draw"
	ResourceTR       ResourceType = "tr"
)

// Global parameter types (used in card outputs)
const (
	ResourceTemperature ResourceType = "temperature"
	ResourceOxygen      ResourceType = "oxygen"
	ResourceOceans      ResourceType = "oceans"
)

// Tile placement types (used in card outputs)
const (
	ResourceCityPlacement     ResourceType = "city-placement"
	ResourceGreeneryPlacement ResourceType = "greenery-placement"
	ResourceOceanPlacement    ResourceType = "ocean-placement"
)

// Card storage resource types
const (
	ResourceAnimals  ResourceType = "animals"
	ResourceMicrobes ResourceType = "microbes"
	ResourceFloaters ResourceType = "floaters"
	ResourceScience  ResourceType = "science"
	ResourceAsteroid ResourceType = "asteroid"
)

// Card draw/selection types
const (
	ResourceCardPeek ResourceType = "card-peek"
	ResourceCardTake ResourceType = "card-take"
	ResourceCardBuy  ResourceType = "card-buy"
)

// ResourceSet represents a collection of resources and their amounts
type ResourceSet struct {
	Credits  int `json:"credits"`
	Steel    int `json:"steel"`
	Titanium int `json:"titanium"`
	Plants   int `json:"plants"`
	Energy   int `json:"energy"`
	Heat     int `json:"heat"`
}

// NewResourceSet creates a new ResourceSet with all values initialized to zero
func NewResourceSet() ResourceSet {
	return ResourceSet{}
}

// Add adds resources from another ResourceSet
func (r *ResourceSet) Add(other ResourceSet) {
	r.Credits += other.Credits
	r.Steel += other.Steel
	r.Titanium += other.Titanium
	r.Plants += other.Plants
	r.Energy += other.Energy
	r.Heat += other.Heat
}

// Subtract subtracts resources from another ResourceSet
func (r *ResourceSet) Subtract(other ResourceSet) {
	r.Credits -= other.Credits
	r.Steel -= other.Steel
	r.Titanium -= other.Titanium
	r.Plants -= other.Plants
	r.Energy -= other.Energy
	r.Heat -= other.Heat
}

// CanAfford checks if this ResourceSet has enough resources to afford the cost
func (r ResourceSet) CanAfford(cost ResourceSet) bool {
	return r.Credits >= cost.Credits &&
		r.Steel >= cost.Steel &&
		r.Titanium >= cost.Titanium &&
		r.Plants >= cost.Plants &&
		r.Energy >= cost.Energy &&
		r.Heat >= cost.Heat
}

// IsEmpty returns true if all resource amounts are zero
func (r ResourceSet) IsEmpty() bool {
	return r.Credits == 0 &&
		r.Steel == 0 &&
		r.Titanium == 0 &&
		r.Plants == 0 &&
		r.Energy == 0 &&
		r.Heat == 0
}

// StandardProjectCosts defines the costs for standard projects
var StandardProjectCosts = struct {
	SellPatents              ResourceSet
	PowerPlant               ResourceSet
	Asteroid                 ResourceSet
	Aquifer                  ResourceSet
	Greenery                 ResourceSet
	City                     ResourceSet
	ConvertHeatToTemperature ResourceSet
	ConvertPlantsToGreenery  ResourceSet
}{
	SellPatents:              ResourceSet{Credits: 0}, // Variable cost based on cards sold
	PowerPlant:               ResourceSet{Credits: 11},
	Asteroid:                 ResourceSet{Credits: 14},
	Aquifer:                  ResourceSet{Credits: 18},
	Greenery:                 ResourceSet{Credits: 23},
	City:                     ResourceSet{Credits: 25},
	ConvertHeatToTemperature: ResourceSet{Heat: 8},
	ConvertPlantsToGreenery:  ResourceSet{Plants: 8},
}
