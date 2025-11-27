package shared

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
