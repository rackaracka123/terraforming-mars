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
	// Check if oceans can be raised using service
	params, err := ctx.GlobalParametersService.GetGlobalParameters(ctx.Context, ctx.Game.ID)
	if err != nil {
		return fmt.Errorf("failed to get global parameters: %w", err)
	}

	if params.Oceans >= 9 {
		return fmt.Errorf("maximum number of ocean tiles already placed")
	}

	// Use service methods to update oceans and player TR
	if err := ctx.GlobalParametersService.PlaceOcean(ctx.Context, ctx.Game.ID, 1); err != nil {
		return fmt.Errorf("failed to place ocean: %w", err)
	}

	// Player gains TR when raising global parameters
	if err := ctx.PlayerService.AddPlayerTR(ctx.Context, ctx.Game.ID, ctx.PlayerID, 1); err != nil {
		return fmt.Errorf("failed to increase player TR: %w", err)
	}

	return nil
}
