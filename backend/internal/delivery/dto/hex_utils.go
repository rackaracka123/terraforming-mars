package dto

import (
	"terraforming-mars-backend/internal/shared/types"
)

// ConvertHexPosition converts a DTO hex position to a model hex position
func ConvertHexPosition(hexDto HexPositionDto) types.HexPosition {
	return types.HexPosition{
		Q: hexDto.Q,
		R: hexDto.R,
		S: hexDto.S,
	}
}
