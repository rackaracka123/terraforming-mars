package parameters

// GlobalParameters represents the global terraforming parameters
type GlobalParameters struct {
	Temperature int
	Oxygen      int
	Oceans      int
}

// Constants for parameter limits
const (
	MaxTemperature = 8  // Maximum temperature in degrees Celsius
	MaxOxygen      = 14 // Maximum oxygen percentage
	MaxOceans      = 9  // Maximum number of ocean tiles
)
