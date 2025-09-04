package economy

import (
	"context"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
)

// MiningGuildHandler implements the Mining Guild card
// This card demonstrates listener registration for cards that react to other players' actions
type MiningGuildHandler struct {
	cards.EffectCardHandler
}

// NewMiningGuildHandler creates a new Mining Guild card handler
func NewMiningGuildHandler() *MiningGuildHandler {
	return &MiningGuildHandler{
		EffectCardHandler: cards.EffectCardHandler{
			BaseCardHandler: cards.BaseCardHandler{
				CardID:       "mining-guild",
				Requirements: model.CardRequirements{},
			},
		},
	}
}

// Play executes the Mining Guild card effect
func (h *MiningGuildHandler) Play(ctx *cards.CardHandlerContext) error {
	// Gain 1 Steel production (immediate effect)
	cards.AddProduction(ctx.Player, model.ResourceSet{
		Steel: 1,
	})
	
	// The ongoing effect (gain resources when steel/titanium is produced) 
	// is handled by the listener system
	return nil
}

// RegisterListeners registers event listeners for Mining Guild
// This card gains bonuses when any player produces steel or titanium
func (h *MiningGuildHandler) RegisterListeners(eventBus events.EventBus) error {
	// Listen for production phase events to gain bonus resources
	eventBus.Subscribe("production-phase", func(ctx context.Context, event events.Event) error {
		return h.handleProductionPhase(event)
	})
	
	// Listen for specific resource production events
	eventBus.Subscribe("steel-produced", func(ctx context.Context, event events.Event) error {
		return h.handleSteelProduction(event)
	})
	
	eventBus.Subscribe("titanium-produced", func(ctx context.Context, event events.Event) error {
		return h.handleTitaniumProduction(event)
	})
	
	return nil
}

// UnregisterListeners cleans up event listeners for Mining Guild
func (h *MiningGuildHandler) UnregisterListeners(eventBus events.EventBus) error {
	// In a production system, this would properly unsubscribe from events
	// For now, this demonstrates the interface
	return nil
}

// handleProductionPhase processes production phase events for Mining Guild benefits
func (h *MiningGuildHandler) handleProductionPhase(event events.Event) error {
	// In a real implementation, this would:
	// 1. Check if any player with Mining Guild in play is producing steel/titanium
	// 2. Grant appropriate bonuses to those players
	// 3. Update game state accordingly
	
	// For demonstration purposes, this is a placeholder
	return nil
}

// handleSteelProduction processes steel production events
func (h *MiningGuildHandler) handleSteelProduction(event events.Event) error {
	// This would grant bonuses when steel is produced
	return nil
}

// handleTitaniumProduction processes titanium production events
func (h *MiningGuildHandler) handleTitaniumProduction(event events.Event) error {
	// This would grant bonuses when titanium is produced
	return nil
}