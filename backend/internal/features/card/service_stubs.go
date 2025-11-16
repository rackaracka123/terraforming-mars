package card

import (
	"context"
	"fmt"
)

// STUB IMPLEMENTATIONS - INTENTIONALLY NOT IMPLEMENTED
//
// These service stubs exist for backward compatibility but should NOT be implemented as feature services.
// The functionality they represent is already implemented in the Actions layer, following clean architecture.
//
// WHY THESE ARE STUBS:
// Per ARCHITECTURE_FLOW.md, Features must NOT import Player/Game domains. Card playing, action execution,
// and effect processing all require orchestrating multiple domains (Player, Game, Parameters, Tiles),
// which is the responsibility of the Actions layer, not isolated feature services.
//
// ACTUAL IMPLEMENTATIONS:
// - ActionService.ExecuteCardAction() → See actions/play_card_action.go::executeCardAction()
// - EffectProcessor.ProcessImmediateEffects() → See actions/play_card.go::processCardEffects()
// - EffectProcessor.ProcessManualAction() → See actions/play_card_action.go::executeCardAction()
// - PlayService.PlayCard() → See actions/play_card.go::Execute()
// - PlayService.CanPlayCard() → See actions/play_card.go::validateCardPlay()
// - PlayService.PlayCardAction() → See actions/play_card_action.go::Execute()
// - DrawService.ConfirmCardDraw() → See actions/card_selection/confirm_card_draw.go::confirmCardDrawInline()
//
// ARCHITECTURE PRINCIPLE:
// Actions orchestrate. Features execute pure domain logic. Cross-domain operations belong in Actions.

// ActionService stub
type ActionServiceStub struct{}

func NewActionService(cardRepo CardRepository, playerRepo interface{}, effectProcessor interface{}) interface{} {
	return &ActionServiceStub{}
}

func (s *ActionServiceStub) ExecuteCardAction(ctx context.Context, gameID, playerID, cardID string, behaviorIndex int, choiceIndex *int, cardStorageTarget *string) error {
	return fmt.Errorf("ActionService not yet implemented - needs refactoring to pure architecture")
}

func (s *ActionServiceStub) CanExecuteCardAction(ctx context.Context, gameID, playerID, cardID string, behaviorIndex int) (bool, error) {
	return false, fmt.Errorf("ActionService not yet implemented - needs refactoring to pure architecture")
}

// EffectProcessor stub
type EffectProcessorStub struct{}

func NewEffectProcessor(cardRepo CardRepository, playerRepo interface{}, gameRepo interface{}, cardDeckRepo interface{}) interface{} {
	return &EffectProcessorStub{}
}

func (s *EffectProcessorStub) ProcessImmediateEffects(ctx context.Context, gameID, playerID string, card *Card, choiceIndex *int, cardStorageTarget *string) error {
	return fmt.Errorf("EffectProcessor not yet implemented - needs refactoring to pure architecture")
}

func (s *EffectProcessorStub) ProcessManualAction(ctx context.Context, gameID, playerID string, action interface{}, choiceIndex *int, cardStorageTarget *string) error {
	return fmt.Errorf("EffectProcessor not yet implemented - needs refactoring to pure architecture")
}

// DrawService stub
type DrawServiceStub struct{}

func NewDrawService(cardRepo CardRepository, cardDeckRepo interface{}, playerRepo interface{}) interface{} {
	return &DrawServiceStub{}
}

func (s *DrawServiceStub) ConfirmCardDraw(ctx context.Context, gameID, playerID string, selectedCardIDs []string) error {
	return fmt.Errorf("DrawService not yet implemented - needs refactoring to pure architecture")
}

// PlayService stub
type PlayServiceStub struct{}

func NewPlayService(cardRepo CardRepository, playerRepo interface{}, effectProcessor interface{}, effectSubscriber interface{}) interface{} {
	return &PlayServiceStub{}
}

func (s *PlayServiceStub) PlayCard(ctx context.Context, gameID, playerID, cardID string, payment interface{}, choiceIndex *int, cardStorageTarget *string) error {
	return fmt.Errorf("PlayService not yet implemented - needs refactoring to pure architecture")
}

func (s *PlayServiceStub) CanPlayCard(ctx context.Context, gameID, playerID, cardID string) (bool, error) {
	return false, fmt.Errorf("PlayService not yet implemented - needs refactoring to pure architecture")
}

func (s *PlayServiceStub) PlayCardAction(ctx context.Context, gameID, playerID, cardID string, behaviorIndex int, choiceIndex *int, cardStorageTarget *string) error {
	return fmt.Errorf("PlayService not yet implemented - needs refactoring to pure architecture")
}
