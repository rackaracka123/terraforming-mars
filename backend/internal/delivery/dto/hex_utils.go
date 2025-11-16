package dto

import (
	"terraforming-mars-backend/internal/features/tiles"
)

// ConvertHexPosition converts a DTO hex position to a model hex position
func ConvertHexPosition(hexDto HexPositionDto) tiles.HexPosition {
	return tiles.HexPosition{
		Q: hexDto.Q,
		R: hexDto.R,
		S: hexDto.S,
	}
}
