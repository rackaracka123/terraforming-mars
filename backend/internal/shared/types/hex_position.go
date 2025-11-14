package types

import "fmt"

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
