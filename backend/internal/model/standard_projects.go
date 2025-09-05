package model

// StandardProject represents the different types of standard projects available to players
type StandardProject string

const (
	StandardProjectSellPatents StandardProject = "SELL_PATENTS"
	StandardProjectPowerPlant  StandardProject = "POWER_PLANT"
	StandardProjectAsteroid    StandardProject = "ASTEROID"
	StandardProjectAquifer     StandardProject = "AQUIFER"
	StandardProjectGreenery    StandardProject = "GREENERY"
	StandardProjectCity        StandardProject = "CITY"
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

// StandardProjectRequiresHexPosition returns true if the standard project requires a hex position
func StandardProjectRequiresHexPosition(project StandardProject) bool {
	switch project {
	case StandardProjectAquifer, StandardProjectGreenery, StandardProjectCity:
		return true
	default:
		return false
	}
}

// StandardProjectProvidesTR returns true if the standard project increases terraform rating
func StandardProjectProvidesTR(project StandardProject) bool {
	switch project {
	case StandardProjectAsteroid, StandardProjectAquifer, StandardProjectGreenery:
		return true
	default:
		return false
	}
}

// HexPosition represents a position on the Mars board using cube coordinates
type HexPosition struct {
	Q int `json:"q" ts:"number"` // Column coordinate
	R int `json:"r" ts:"number"` // Row coordinate
	S int `json:"s" ts:"number"` // Third coordinate (Q + R + S = 0)
}

// IsValid validates that the hex position follows cube coordinate rules
func (h HexPosition) IsValid() bool {
	return h.Q+h.R+h.S == 0
}

// Distance calculates the distance between two hex positions
func (h HexPosition) Distance(other HexPosition) int {
	return (abs(h.Q-other.Q) + abs(h.R-other.R) + abs(h.S-other.S)) / 2
}

// GetNeighbors returns all adjacent hex positions
func (h HexPosition) GetNeighbors() []HexPosition {
	directions := []HexPosition{
		{1, -1, 0},  // East
		{1, 0, -1},  // Southeast
		{0, 1, -1},  // Southwest
		{-1, 1, 0},  // West
		{-1, 0, 1},  // Northwest
		{0, -1, 1},  // Northeast
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

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}