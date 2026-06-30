package action

import (
	"time"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
	"terraforming-mars-backend/internal/game/shared"
)

// Canonical base costs for a single colony trade, before any action discounts.
// A trade is paid with exactly one of these resource types.
const (
	TradeCreditsCost  = 9
	TradeEnergyCost   = 3
	TradeTitaniumCost = 3
)

// tradePaymentCost pairs a payment resource with its base cost so affordability
// and effective-cost computation share a single iteration.
type tradePaymentCost struct {
	resource shared.ResourceType
	baseCost int
}

var tradePaymentCosts = []tradePaymentCost{
	{shared.ResourceCredit, TradeCreditsCost},
	{shared.ResourceEnergy, TradeEnergyCost},
	{shared.ResourceTitanium, TradeTitaniumCost},
}

// CalculateColonyTradeState is the single source of truth for colony-trade
// affordability and availability. The trade is available when the player's
// trade fleet is free, at least one colony has not yet been traded this
// generation, and the player can afford at least one payment type after
// applying action discounts (e.g. Rim Freighters).
//
// The returned EntityState carries the effective per-resource trade cost in
// Cost and any discounts in Metadata, so both the DTO mapper and the
// available-actions aggregator consume identical numbers.
func CalculateColonyTradeState(
	p *player.Player,
	g *game.Game,
	cardRegistry cards.CardRegistry,
) player.EntityState {
	var errors []player.StateError
	metadata := make(map[string]any)

	errors = append(errors, validatePhase(g)...)
	errors = append(errors, validateActionsRemaining(p, g)...)
	errors = append(errors, validateNoPendingSelection(p, g)...)

	effectiveCosts, discounts := CalculateEffectiveTradeCosts(p, cardRegistry)
	if len(discounts) > 0 {
		metadata["discounts"] = discounts
	}

	if !g.Colonies().GetTradeFleetAvailable(p.ID()) {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeNoActionsRemaining,
			Category: player.ErrorCategoryAvailability,
			Message:  "Trade fleet is not available",
		})
	}

	if len(g.Colonies().GetTradeableIDs()) == 0 {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeNoTilePlacements,
			Category: player.ErrorCategoryAvailability,
			Message:  "No tradeable colonies",
		})
	}

	if !canAffordAnyTrade(p, effectiveCosts) {
		errors = append(errors, player.StateError{
			Code:     player.ErrorCodeInsufficientResources,
			Category: player.ErrorCategoryCost,
			Message:  "Cannot afford any trade payment type",
		})
	}

	return player.EntityState{
		Errors:         errors,
		Cost:           effectiveCosts,
		Metadata:       metadata,
		LastCalculated: time.Now(),
	}
}

// CalculateEffectiveTradeCosts returns the per-resource effective trade cost
// after applying the player's colony-trade action discounts, alongside the
// non-zero discounts that were applied. It is the shared cost computation used
// by trade execution, the DTO mapper, and CalculateColonyTradeState.
func CalculateEffectiveTradeCosts(
	p *player.Player,
	cardRegistry cards.CardRegistry,
) (effectiveCosts map[string]int, discounts map[string]int) {
	calc := gamecards.NewRequirementModifierCalculator(cardRegistry)
	tradeDiscounts := calc.CalculateActionDiscounts(p, shared.ActionColonyTrade)

	effectiveCosts = make(map[string]int, len(tradePaymentCosts))
	discounts = make(map[string]int)
	for _, pc := range tradePaymentCosts {
		discount := tradeDiscounts[pc.resource]
		effective := pc.baseCost - discount
		if effective < 0 {
			effective = 0
		}
		effectiveCosts[string(pc.resource)] = effective
		if discount > 0 {
			discounts[string(pc.resource)] = discount
		}
	}
	return effectiveCosts, discounts
}

// canAffordAnyTrade reports whether the player can pay at least one payment type
// at its effective cost.
func canAffordAnyTrade(p *player.Player, effectiveCosts map[string]int) bool {
	resources := p.Resources().Get()
	available := map[shared.ResourceType]int{
		shared.ResourceCredit:   resources.Credits,
		shared.ResourceEnergy:   resources.Energy,
		shared.ResourceTitanium: resources.Titanium,
	}
	for _, pc := range tradePaymentCosts {
		if available[pc.resource] >= effectiveCosts[string(pc.resource)] {
			return true
		}
	}
	return false
}
