package player

import (
	"terraforming-mars-backend/internal/game/shared"
)

// CardEffect represents an ongoing effect defined by a card
type CardEffect struct {
	CardID        string
	CardName      string
	BehaviorIndex int
	Behavior      shared.CardBehavior
}

// DeepCopy creates a deep copy of the CardEffect
func (pe *CardEffect) DeepCopy() *CardEffect {
	if pe == nil {
		return nil
	}

	return &CardEffect{
		CardID:        pe.CardID,
		CardName:      pe.CardName,
		BehaviorIndex: pe.BehaviorIndex,
		Behavior:      pe.Behavior.DeepCopy(),
	}
}

// CardAction represents a repeatable manual action defined by a card
type CardAction struct {
	CardID        string
	CardName      string
	BehaviorIndex int
	Behavior      shared.CardBehavior
	PlayCount     int
}

// DeepCopy creates a deep copy of the CardAction
func (pa *CardAction) DeepCopy() *CardAction {
	if pa == nil {
		return nil
	}

	return &CardAction{
		CardID:        pa.CardID,
		CardName:      pa.CardName,
		BehaviorIndex: pa.BehaviorIndex,
		Behavior:      pa.Behavior.DeepCopy(),
		PlayCount:     pa.PlayCount,
	}
}
