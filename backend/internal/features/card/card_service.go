package card

import (
	"context"
)

type PendingCardDrawSelection struct {
	AvailableCards []string
	FreeTakeCount  int
	MaxBuyCount    int
	CardBuyCost    int
	Source         string
}

type SelectStartingCardsPhase struct {
	AvailableCards        []string
	AvailableCorporations []string
	SelectionComplete     bool
}

type ProductionPhaseState struct {
	AvailableCards    []string
	SelectionComplete bool
}

type PlayerForcedFirstAction struct {
	ActionType    string
	CorporationID string
	Completed     bool
	Description   string
}

// Service interfaces for card operations
// These are placeholder interfaces until proper implementations are created

// DrawService handles card drawing operations
type DrawService interface {
	ConfirmCardDraw(ctx context.Context, gameID, playerID string, selectedCardIDs []string) error
}

// PlayService handles card playing operations
type PlayService interface {
	PlayCardAction(ctx context.Context, gameID, playerID, cardID string, behaviorIndex int, choiceIndex *int, cardStorageTarget *string) error
}
