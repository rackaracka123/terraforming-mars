package model

import "fmt"

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

// HexPosition represents a position on the Mars board using cube coordinates
type HexPosition struct {
	Q int `json:"q" ts:"number"` // Column coordinate
	R int `json:"r" ts:"number"` // Row coordinate
	S int `json:"s" ts:"number"` // Third coordinate (Q + R + S = 0)
}

// GetNeighbors returns all 6 adjacent hex positions using cube coordinate system
func (h HexPosition) GetNeighbors() []HexPosition {
	// Six adjacent directions in cube coordinates (hexagonal grid)
	directions := []HexPosition{
		{Q: 1, R: -1, S: 0}, // East
		{Q: 1, R: 0, S: -1}, // Southeast
		{Q: 0, R: 1, S: -1}, // Southwest
		{Q: -1, R: 1, S: 0}, // West
		{Q: -1, R: 0, S: 1}, // Northwest
		{Q: 0, R: -1, S: 1}, // Northeast
	}

	neighbors := make([]HexPosition, 6)
	for i, dir := range directions {
		neighbors[i] = HexPosition{
			Q: h.Q + dir.Q,
			R: h.R + dir.R,
			S: h.S + dir.S,
		}
	}
	return neighbors
}

// Equals checks if two hex positions are equal
func (h HexPosition) Equals(other HexPosition) bool {
	return h.Q == other.Q && h.R == other.R && h.S == other.S
}

// String returns the hex position as a formatted string "q,r,s"
func (h HexPosition) String() string {
	return fmt.Sprintf("%d,%d,%d", h.Q, h.R, h.S)
}
