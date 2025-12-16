package shared

// ResourceType represents different types of resources in the game
type ResourceType string

const (
	// Basic resources (singular form for grammatically correct error messages)
	ResourceCredit   ResourceType = "credit"
	ResourceSteel    ResourceType = "steel"
	ResourceTitanium ResourceType = "titanium"
	ResourcePlant    ResourceType = "plant"
	ResourceEnergy   ResourceType = "energy"
	ResourceHeat     ResourceType = "heat"
	ResourceMicrobe  ResourceType = "microbe"
	ResourceAnimal   ResourceType = "animal"
	ResourceFloater  ResourceType = "floater"
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

	// Tile types for board spaces
	ResourceLandTile   ResourceType = "land"
	ResourceOceanSpace ResourceType = "ocean-space"

	// Global parameters
	ResourceTemperature     ResourceType = "temperature"
	ResourceOxygen          ResourceType = "oxygen"
	ResourceOcean           ResourceType = "ocean"
	ResourceVenus           ResourceType = "venus"
	ResourceTR              ResourceType = "tr"
	ResourceGlobalParameter ResourceType = "global-parameter"

	// Production resources (singular form for grammatically correct error messages)
	ResourceCreditProduction   ResourceType = "credit-production"
	ResourceSteelProduction    ResourceType = "steel-production"
	ResourceTitaniumProduction ResourceType = "titanium-production"
	ResourcePlantProduction    ResourceType = "plant-production"
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
