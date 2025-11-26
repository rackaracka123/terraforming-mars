package types

// StandardProject represents the different types of standard projects available to players
type StandardProject string

const (
	// Standard Projects (M€-based)
	StandardProjectSellPatents StandardProject = "sell-patents"
	StandardProjectPowerPlant  StandardProject = "power-plant"
	StandardProjectAsteroid    StandardProject = "asteroid"
	StandardProjectAquifer     StandardProject = "aquifer"
	StandardProjectGreenery    StandardProject = "greenery"
	StandardProjectCity        StandardProject = "city"
	// Resource Conversion Actions (resource-based, not M€)
	StandardProjectConvertPlantsToGreenery  StandardProject = "convert-plants-to-greenery"
	StandardProjectConvertHeatToTemperature StandardProject = "convert-heat-to-temperature"
)

// StandardProjectCost represents the credit cost for each standard project
var StandardProjectCost = map[StandardProject]int{
	StandardProjectSellPatents: 0,  // No cost - player gains M€ instead
	StandardProjectPowerPlant:  11, // 11 M€
	StandardProjectAsteroid:    14, // 14 M€
	StandardProjectAquifer:     18, // 18 M€
	StandardProjectGreenery:    23, // 23 M€
	StandardProjectCity:        25, // 25 M€
}
