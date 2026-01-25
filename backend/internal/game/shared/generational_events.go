package shared

type GenerationalEvent string

const (
	GenerationalEventTRRaise           GenerationalEvent = "tr-raise"
	GenerationalEventOceanPlacement    GenerationalEvent = "ocean-placement"
	GenerationalEventCityPlacement     GenerationalEvent = "city-placement"
	GenerationalEventGreeneryPlacement GenerationalEvent = "greenery-placement"
)

type MinMax struct {
	Min *int `json:"min,omitempty" ts:"number | undefined"`
	Max *int `json:"max,omitempty" ts:"number | undefined"`
}

type GenerationalEventRequirement struct {
	Event  GenerationalEvent `json:"event" ts:"GenerationalEvent"`
	Count  *MinMax           `json:"count,omitempty" ts:"MinMax | undefined"`
	Target *string           `json:"target,omitempty" ts:"string | undefined"`
}

type PlayerGenerationalEventEntry struct {
	Event GenerationalEvent `json:"event" ts:"GenerationalEvent"`
	Count int               `json:"count" ts:"number"`
}
