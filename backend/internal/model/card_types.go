package model

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
)

// ResourceSet represents a collection of resources and their amounts
type ResourceSet struct {
	Credits  int `json:"credits" ts:"number"`
	Steel    int `json:"steel" ts:"number"`
	Titanium int `json:"titanium" ts:"number"`
	Plants   int `json:"plants" ts:"number"`
	Energy   int `json:"energy" ts:"number"`
	Heat     int `json:"heat" ts:"number"`
}

// CardRequirements defines what conditions must be met to play a card
type CardRequirements struct {
	// MinTemperature is the minimum global temperature required (-30 to +8)
	MinTemperature *int `json:"minTemperature,omitempty" ts:"number | undefined"`

	// MaxTemperature is the maximum global temperature allowed (-30 to +8)
	MaxTemperature *int `json:"maxTemperature,omitempty" ts:"number | undefined"`

	// MinOxygen is the minimum oxygen percentage required (0-14)
	MinOxygen *int `json:"minOxygen,omitempty" ts:"number | undefined"`

	// MaxOxygen is the maximum oxygen percentage allowed (0-14)
	MaxOxygen *int `json:"maxOxygen,omitempty" ts:"number | undefined"`

	// MinOceans is the minimum number of ocean tiles required (0-9)
	MinOceans *int `json:"minOceans,omitempty" ts:"number | undefined"`

	// MaxOceans is the maximum number of ocean tiles allowed (0-9)
	MaxOceans *int `json:"maxOceans,omitempty" ts:"number | undefined"`

	// RequiredTags are tags that the player must have from played cards
	RequiredTags []CardTag `json:"requiredTags" ts:"CardTag[]"`

	// RequiredProduction specifies minimum production requirements
	RequiredProduction *ResourceSet `json:"requiredProduction,omitempty" ts:"ResourceSet | undefined"`
}
