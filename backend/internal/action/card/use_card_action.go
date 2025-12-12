package card

import (
	"context"
	"fmt"
	baseaction "terraforming-mars-backend/internal/action"

	"go.uber.org/zap"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/game"
	gamecards "terraforming-mars-backend/internal/game/cards"
	"terraforming-mars-backend/internal/game/player"
)

// UseCardActionAction handles the business logic for using a card's manual action
// Card actions are repeatable blue card abilities with inputs and outputs
type UseCardActionAction struct {
	baseaction.BaseAction
}

// NewUseCardActionAction creates a new use card action action
func NewUseCardActionAction(
	gameRepo game.GameRepository,
	cardRegistry cards.CardRegistry,
	logger *zap.Logger,
) *UseCardActionAction {
	return &UseCardActionAction{
		BaseAction: baseaction.NewBaseAction(gameRepo, cardRegistry),
	}
}

// Execute performs the use card action
func (a *UseCardActionAction) Execute(
	ctx context.Context,
	gameID string,
	playerID string,
	cardID string,
	behaviorIndex int,
) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex),
		zap.String("action", "use_card_action"),
	)
	log.Info("🎯 Player attempting to use card action")

	// 1. Validate game exists and is active
	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate game is in action phase
	if err := baseaction.ValidateGamePhase(g, game.GamePhaseAction, log); err != nil {
		return err
	}

	// 3. Validate it's the player's turn
	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 4. Get player from game
	p, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	// 5. BUSINESS LOGIC: Find the card action in player's available actions
	cardAction, err := a.findCardAction(p, cardID, behaviorIndex, log)
	if err != nil {
		return err
	}

	log.Info("✅ Found card action",
		zap.String("card_name", cardAction.CardName),
		zap.Int("play_count", cardAction.PlayCount))

	// 6. BUSINESS LOGIC: Use BehaviorApplier for inputs and outputs
	applier := gamecards.NewBehaviorApplier(p, g, cardAction.CardName, log)

	// 7. BUSINESS LOGIC: Validate and apply inputs (resource costs)
	if err := applier.ApplyInputs(ctx, cardAction.Behavior.Inputs); err != nil {
		log.Error("Failed to apply inputs", zap.Error(err))
		return err
	}

	// 8. BUSINESS LOGIC: Apply outputs (resource gains, etc.)
	if err := applier.ApplyOutputs(ctx, cardAction.Behavior.Outputs); err != nil {
		log.Error("Failed to apply outputs", zap.Error(err))
		return err
	}

	// 9. BUSINESS LOGIC: Increment play count for the action
	a.incrementPlayCount(p, cardID, behaviorIndex, log)

	// 10. BUSINESS LOGIC: Consume a player action
	consumed := a.ConsumePlayerAction(g, log)
	if !consumed {
		log.Warn("⚠️ Action not consumed (unlimited actions or already at 0)")
	}

	log.Info("🎉 Card action executed successfully")
	return nil
}

// findCardAction finds a card action in the player's available actions
func (a *UseCardActionAction) findCardAction(
	p *player.Player,
	cardID string,
	behaviorIndex int,
	log *zap.Logger,
) (*player.CardAction, error) {
	actions := p.Actions().List()

	for i := range actions {
		if actions[i].CardID == cardID && actions[i].BehaviorIndex == behaviorIndex {
			return &actions[i], nil
		}
	}

	log.Error("Card action not found in player's available actions",
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex))
	return nil, fmt.Errorf("card action not found: %s[%d]", cardID, behaviorIndex)
}

// incrementPlayCount increments the play count for a card action
func (a *UseCardActionAction) incrementPlayCount(
	p *player.Player,
	cardID string,
	behaviorIndex int,
	log *zap.Logger,
) {
	actions := p.Actions().List()

	// Find and increment play count
	for i := range actions {
		if actions[i].CardID == cardID && actions[i].BehaviorIndex == behaviorIndex {
			actions[i].PlayCount++
			log.Debug("📊 Incremented action play count",
				zap.Int("new_count", actions[i].PlayCount))
			break
		}
	}

	// Update player actions
	p.Actions().SetActions(actions)
}
