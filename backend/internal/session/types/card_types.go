package types

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

// ResourceSet represents a collection of resources and their amounts
type ResourceSet struct {
	Credits  int
	Steel    int
	Titanium int
	Plants   int
	Energy   int
	Heat     int
}

// CardRequirements defines what conditions must be met to play a card
type CardRequirements struct {
	// MinTemperature is the minimum global temperature required (-30 to +8)
	MinTemperature *int

	// MaxTemperature is the maximum global temperature allowed (-30 to +8)
	MaxTemperature *int

	// MinOxygen is the minimum oxygen percentage required (0-14)
	MinOxygen *int

	// MaxOxygen is the maximum oxygen percentage allowed (0-14)
	MaxOxygen *int

	// MinOceans is the minimum number of ocean tiles required (0-9)
	MinOceans *int

	// MaxOceans is the maximum number of ocean tiles allowed (0-9)
	MaxOceans *int

	// RequiredTags are tags that the player must have from played cards
	RequiredTags []CardTag

	// RequiredProduction specifies minimum production requirements
	RequiredProduction *ResourceSet
}
