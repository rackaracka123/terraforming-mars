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

// HexPosition represents a position on the Mars board using cube coordinates
type HexPosition struct {
	Q int `json:"q" ts:"number"` // Column coordinate
	R int `json:"r" ts:"number"` // Row coordinate
	S int `json:"s" ts:"number"` // Third coordinate (Q + R + S = 0)
}

// DeepCopy creates a deep copy of the HexPosition
func (h *HexPosition) DeepCopy() *HexPosition {
	if h == nil {
		return nil
	}

	return &HexPosition{
		Q: h.Q,
		R: h.R,
		S: h.S,
	}
}
