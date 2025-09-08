package science

import (
	"context"
	"fmt"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
)

// AtmosphericProcessorsHandler implements the Atmospheric Processors card
type AtmosphericProcessorsHandler struct {
	cards.EffectCardHandler
}

// NewAtmosphericProcessorsHandler creates a new Atmospheric Processors card handler
func NewAtmosphericProcessorsHandler() *AtmosphericProcessorsHandler {
	return &AtmosphericProcessorsHandler{
		EffectCardHandler: cards.EffectCardHandler{
			BaseCardHandler: cards.BaseCardHandler{
				CardID:       "atmospheric-processors",
				Requirements: model.CardRequirements{},
			},
		},
	}
}

// Play executes the Atmospheric Processors card effect
func (h *AtmosphericProcessorsHandler) Play(ctx *cards.CardHandlerContext) error {
	// Check current oxygen level before attempting to raise it
	params, err := ctx.GameService.GetGlobalParameters(ctx.Context, ctx.Game.ID)
	if err != nil {
		return fmt.Errorf("failed to get global parameters: %w", err)
	}

	if params.Oxygen >= 14 {
		return fmt.Errorf("oxygen already at maximum level")
	}

	// Raise oxygen 1 step using the service
	if err := ctx.GameService.IncreaseOxygen(ctx.Context, ctx.Game.ID, 1); err != nil {
		return fmt.Errorf("failed to increase oxygen: %w", err)
	}

	// Player gains TR when raising global parameters
	if err := ctx.PlayerService.AddPlayerTR(ctx.Context, ctx.Game.ID, ctx.PlayerID, 1); err != nil {
		return fmt.Errorf("failed to increase player TR: %w", err)
	}

	return nil
}

// RegisterListeners registers event listeners for Atmospheric Processors
// This card reacts to temperature changes by providing bonus oxygen
func (h *AtmosphericProcessorsHandler) RegisterListeners(eventBus events.EventBus) error {
	// Listen for temperature increases to provide synergy effects
	eventBus.Subscribe("temperature-increased", func(ctx context.Context, event events.Event) error {
		// In a real implementation, this would check if the card is in play
		// for a specific player and provide appropriate bonuses

		// For now, this is a demonstration of the pattern
		// The actual implementation would:
		// 1. Check if this card is in play for any player
		// 2. Apply appropriate synergy effects
		// 3. Update game state accordingly

		return nil
	})

	return nil
}

// UnregisterListeners cleans up event listeners for Atmospheric Processors
func (h *AtmosphericProcessorsHandler) UnregisterListeners(eventBus events.EventBus) error {
	// In a real implementation, this would properly unsubscribe
	// For demonstration purposes, we'll leave this as a placeholder
	return nil
}
