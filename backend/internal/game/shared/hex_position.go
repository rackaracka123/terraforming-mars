package shared

import "fmt"

// HexPosition represents a position on the Mars board using cube coordinates
type HexPosition struct {
	Q int `json:"q"` // Column coordinate
	R int `json:"r"` // Row coordinate
	S int `json:"s"` // Third coordinate (Q + R + S = 0)
}

// String returns a string representation of the hex position
func (h HexPosition) String() string {
	return fmt.Sprintf("%d,%d,%d", h.Q, h.R, h.S)
}

// GetNeighbors returns all 6 adjacent hex positions using cube coordinate system
func (h HexPosition) GetNeighbors() []HexPosition {
	directions := []HexPosition{
		{Q: 1, R: -1, S: 0}, // East
		{Q: 1, R: 0, S: -1}, // Northeast
		{Q: 0, R: 1, S: -1}, // Northwest
		{Q: -1, R: 1, S: 0}, // West
		{Q: -1, R: 0, S: 1}, // Southwest
		{Q: 0, R: -1, S: 1}, // Southeast
	}

	neighbors := make([]HexPosition, 0, 6)
	for _, dir := range directions {
		neighbors = append(neighbors, HexPosition{
			Q: h.Q + dir.Q,
			R: h.R + dir.R,
			S: h.S + dir.S,
		})
	}

	return neighbors
}

// Equals checks if two hex positions are equal
func (h HexPosition) Equals(other HexPosition) bool {
	return h.Q == other.Q && h.R == other.R && h.S == other.S
}
