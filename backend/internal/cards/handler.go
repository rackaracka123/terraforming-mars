package cards

import (
	"context"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
)

// CardHandlerContext provides the context needed for card handlers to execute
type CardHandlerContext struct {
	Context       context.Context
	Game          *model.Game
	PlayerID      string
	Card          *model.Card
	EventBus      events.EventBus
	PlayerService PlayerService
}

// CardHandler defines the interface that all card implementations must satisfy
type CardHandler interface {
	// GetCardID returns the unique identifier for this card
	GetCardID() string
	
	// CanPlay checks if the card can be played given the current game state
	CanPlay(ctx *CardHandlerContext) error
	
	// Play executes the card's effects
	Play(ctx *CardHandlerContext) error
	
	// GetRequirements returns the requirements needed to play this card
	GetRequirements() model.CardRequirements
}

// ListenerRegistrar defines the interface for cards that need to register event listeners
type ListenerRegistrar interface {
	// RegisterListeners registers event listeners for this card
	RegisterListeners(eventBus events.EventBus) error
	
	// UnregisterListeners cleans up event listeners for this card
	UnregisterListeners(eventBus events.EventBus) error
}

// CardWithListeners combines card handling with listener registration
type CardWithListeners interface {
	CardHandler
	ListenerRegistrar
}


// BaseCardHandler provides common functionality for all card handlers
type BaseCardHandler struct {
	CardID       string
	Requirements model.CardRequirements
}

// GetCardID returns the card's unique identifier
func (b *BaseCardHandler) GetCardID() string {
	return b.CardID
}

// GetRequirements returns the card's play requirements
func (b *BaseCardHandler) GetRequirements() model.CardRequirements {
	return b.Requirements
}

// CanPlay performs basic requirement checking that all cards need
func (b *BaseCardHandler) CanPlay(ctx *CardHandlerContext) error {
	return ValidateCardRequirements(ctx.Context, ctx.Game, ctx.PlayerID, ctx.PlayerService, b.Requirements)
}

// RegisterListeners provides a default implementation that does nothing
// Cards that need listeners should override this method
func (b *BaseCardHandler) RegisterListeners(eventBus events.EventBus) error {
	// Default implementation - no listeners to register
	return nil
}

// UnregisterListeners provides a default implementation that does nothing
// Cards that need listeners should override this method
func (b *BaseCardHandler) UnregisterListeners(eventBus events.EventBus) error {
	// Default implementation - no listeners to unregister
	return nil
}

// EventCardHandler is for red cards that have immediate one-time effects
type EventCardHandler struct {
	BaseCardHandler
}

// EffectCardHandler is for green cards that provide ongoing benefits/production
type EffectCardHandler struct {
	BaseCardHandler
}

// ActiveCardHandler is for blue cards that can be activated during action phase
type ActiveCardHandler struct {
	BaseCardHandler
	// ActivationCost defines what it costs to use this card's action
	ActivationCost *model.ResourceSet
}

// CanActivate checks if the active card can be used this turn
func (a *ActiveCardHandler) CanActivate(ctx *CardHandlerContext) error {
	if a.ActivationCost != nil {
		return ctx.PlayerService.ValidateResourceCost(ctx.Context, ctx.Game.ID, ctx.PlayerID, *a.ActivationCost)
	}
	return nil
}

// Activate performs the card's repeatable action
func (a *ActiveCardHandler) Activate(ctx *CardHandlerContext) error {
	// Base implementation - subclasses should override
	return nil
}