package dto

import (
	"terraforming-mars-backend/internal/model"
)

// ConvertHexPosition converts a DTO hex position to a model hex position
func ConvertHexPosition(hexDto HexPositionDto) model.HexPosition {
	return model.HexPosition{
		Q: hexDto.Q,
		R: hexDto.R,
		S: hexDto.S,
	}
}
