package model

// ResourceType represents different types of resources in the game.
// This includes basic resources, production resources, global parameters,
// tile types, and special effect types used throughout the game system.
type ResourceType string

const (
	// Basic resources

	// ResourceCredits represents the credits (M€) resource
	ResourceCredits ResourceType = "credits"
	// ResourceSteel represents the steel resource
	ResourceSteel ResourceType = "steel"
	// ResourceTitanium represents the titanium resource
	ResourceTitanium ResourceType = "titanium"
	// ResourcePlants represents the plants resource
	ResourcePlants ResourceType = "plants"
	// ResourceEnergy represents the energy resource
	ResourceEnergy ResourceType = "energy"
	// ResourceHeat represents the heat resource
	ResourceHeat ResourceType = "heat"
	// ResourceMicrobes represents the microbes resource stored on cards
	ResourceMicrobes ResourceType = "microbes"
	// ResourceAnimals represents the animals resource stored on cards
	ResourceAnimals ResourceType = "animals"
	// ResourceFloaters represents the floaters resource stored on cards
	ResourceFloaters ResourceType = "floaters"
	// ResourceScience represents the science resource stored on cards
	ResourceScience ResourceType = "science"
	// ResourceAsteroid represents asteroid resources
	ResourceAsteroid ResourceType = "asteroid"
	// ResourceDisease represents disease resources that can be removed
	ResourceDisease ResourceType = "disease"

	// Card actions

	// ResourceCardDraw represents drawing cards (take = peek, common case)
	ResourceCardDraw ResourceType = "card-draw"
	// ResourceCardTake represents drawing cards (take from deck)
	ResourceCardTake ResourceType = "card-take"
	// ResourceCardPeek represents looking at cards (peek without taking all)
	ResourceCardPeek ResourceType = "card-peek"
	// ResourceCardBuy represents buying cards from peeked cards (buy for standard price)
	ResourceCardBuy ResourceType = "card-buy"

	// Terraforming actions

	// ResourceCityPlacement represents placing city tiles
	ResourceCityPlacement ResourceType = "city-placement"
	// ResourceOceanPlacement represents placing ocean tiles
	ResourceOceanPlacement ResourceType = "ocean-placement"
	// ResourceGreeneryPlacement represents placing greenery tiles
	ResourceGreeneryPlacement ResourceType = "greenery-placement"

	// Tile counting (for per conditions)

	// ResourceCityTile represents counting existing city tiles
	ResourceCityTile ResourceType = "city-tile"
	// ResourceOceanTile represents counting existing ocean tiles
	ResourceOceanTile ResourceType = "ocean-tile"
	// ResourceGreeneryTile represents counting existing greenery tiles
	ResourceGreeneryTile ResourceType = "greenery-tile"
	// ResourceColonyTile represents counting existing colonies
	ResourceColonyTile ResourceType = "colony-tile"

	// Global parameters

	// ResourceTemperature represents temperature change
	ResourceTemperature ResourceType = "temperature"
	// ResourceOxygen represents oxygen change
	ResourceOxygen ResourceType = "oxygen"
	// ResourceOceans represents ocean tile placement (global parameter)
	ResourceOceans ResourceType = "oceans"
	// ResourceVenus represents venus change
	ResourceVenus ResourceType = "venus"
	// ResourceTR represents terraform Rating change
	ResourceTR ResourceType = "tr"
	// ResourceGlobalParameter represents any global parameter (for leniences/requirements)
	ResourceGlobalParameter ResourceType = "global-parameter"

	// Production resources (for spending production)

	// ResourceCreditsProduction represents credits production
	ResourceCreditsProduction ResourceType = "credits-production"
	// ResourceSteelProduction represents steel production
	ResourceSteelProduction ResourceType = "steel-production"
	// ResourceTitaniumProduction represents titanium production
	ResourceTitaniumProduction ResourceType = "titanium-production"
	// ResourcePlantsProduction represents plants production
	ResourcePlantsProduction ResourceType = "plants-production"
	// ResourceEnergyProduction represents energy production
	ResourceEnergyProduction ResourceType = "energy-production"
	// ResourceHeatProduction represents heat production
	ResourceHeatProduction ResourceType = "heat-production"
	// ResourceAnyProduction represents any production type (for dynamic effects like Manutech)
	ResourceAnyProduction ResourceType = "any-production"

	// Effect type (for triggered effects like Rover Construction)

	// ResourceEffect represents triggered effect
	ResourceEffect ResourceType = "effect"

	// Tag counting (for VP conditions like "1 VP per jovian tag")

	// ResourceTag represents count tags of a specific type
	ResourceTag ResourceType = "tag"

	// Special ongoing effects

	// ResourceGlobalParameterLenience represents global parameter requirement flexibility
	ResourceGlobalParameterLenience ResourceType = "global-parameter-lenience"
	// ResourceVenusLenience represents venus requirement flexibility (+/- steps)
	ResourceVenusLenience ResourceType = "venus-lenience"
	// ResourceDefense represents protection from attacks or resource removal
	ResourceDefense ResourceType = "defense"
	// ResourceDiscount represents cost reduction for playing cards
	ResourceDiscount ResourceType = "discount"
	// ResourceValueModifier represents increases resource values (e.g., steel/titanium worth more)
	ResourceValueModifier ResourceType = "value-modifier"
	// ResourcePaymentSubstitute represents using one resource type as payment for another (e.g., heat as credits)
	ResourcePaymentSubstitute ResourceType = "payment-substitute"
	// ResourceOceanAdjacencyBonus represents additional bonus for placing tiles adjacent to oceans
	ResourceOceanAdjacencyBonus ResourceType = "ocean-adjacency-bonus"
)
