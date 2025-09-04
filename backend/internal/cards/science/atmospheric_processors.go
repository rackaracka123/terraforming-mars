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
	// Raise oxygen 1 step
	if ctx.Game.GlobalParameters.Oxygen < 14 {
		ctx.Game.GlobalParameters.Oxygen += 1
		if ctx.Game.GlobalParameters.Oxygen > 14 {
			ctx.Game.GlobalParameters.Oxygen = 14
		}
		
		// Player gains TR when raising global parameters
		ctx.Player.TerraformRating += 1
	} else {
		return fmt.Errorf("oxygen already at maximum level")
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