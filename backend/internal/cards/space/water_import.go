package space

import (
	"fmt"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/model"
)

// WaterImportHandler implements the Water Import from Europa card
type WaterImportHandler struct {
	cards.EventCardHandler
}

// NewWaterImportHandler creates a new Water Import card handler
func NewWaterImportHandler() *WaterImportHandler {
	return &WaterImportHandler{
		EventCardHandler: cards.EventCardHandler{
			BaseCardHandler: cards.BaseCardHandler{
				CardID:       "water-import",
				Requirements: model.CardRequirements{},
			},
		},
	}
}

// Play executes the Water Import card effect
func (h *WaterImportHandler) Play(ctx *cards.CardHandlerContext) error {
	// Place 1 ocean tile
	if ctx.Game.GlobalParameters.Oceans < 9 {
		ctx.Game.GlobalParameters.Oceans += 1
		if ctx.Game.GlobalParameters.Oceans > 9 {
			ctx.Game.GlobalParameters.Oceans = 9
		}

		// Player gains TR when raising global parameters
		player, found := ctx.Game.GetPlayer(ctx.PlayerID)
		if !found {
			return fmt.Errorf("player not found in game")
		}
		player.TerraformRating += 1
	} else {
		return fmt.Errorf("maximum number of ocean tiles already placed")
	}

	return nil
}
