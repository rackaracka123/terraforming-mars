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
	choiceIndex *int,
	cardStorageTarget *string,
) error {
	log := a.InitLogger(gameID, playerID).With(
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex),
		zap.String("action", "use_card_action"),
	)
	if choiceIndex != nil {
		log = log.With(zap.Int("choice_index", *choiceIndex))
	}
	if cardStorageTarget != nil {
		log = log.With(zap.String("card_storage_target", *cardStorageTarget))
	}
	log.Info("🎯 Player attempting to use card action")

	g, err := baseaction.ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	if err := baseaction.ValidateGamePhase(g, game.GamePhaseAction, log); err != nil {
		return err
	}

	if err := baseaction.ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	p, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	cardAction, err := a.findCardAction(p, cardID, behaviorIndex, log)
	if err != nil {
		return err
	}

	log.Info("✅ Found card action",
		zap.String("card_name", cardAction.CardName),
		zap.Int("times_used_this_generation", cardAction.TimesUsedThisGeneration))

	applier := gamecards.NewBehaviorApplier(p, g, cardAction.CardName, log).
		WithSourceCardID(cardID)
	if cardStorageTarget != nil {
		applier = applier.WithTargetCardID(*cardStorageTarget)
	}

	inputs, outputs := cardAction.Behavior.ExtractInputsOutputs(choiceIndex)

	if choiceIndex != nil {
		log.Info("📋 Using choice-specific behavior",
			zap.Int("choice_index", *choiceIndex),
			zap.Int("input_count", len(inputs)),
			zap.Int("output_count", len(outputs)))
	}

	if err := applier.ApplyInputs(ctx, inputs); err != nil {
		log.Error("Failed to apply inputs", zap.Error(err))
		return err
	}

	if err := applier.ApplyOutputs(ctx, outputs); err != nil {
		log.Error("Failed to apply outputs", zap.Error(err))
		return err
	}

	a.incrementUsageCounts(p, cardID, behaviorIndex, log)

	a.ConsumePlayerAction(g, log)

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

// incrementUsageCounts increments the usage counts for a card action
func (a *UseCardActionAction) incrementUsageCounts(
	p *player.Player,
	cardID string,
	behaviorIndex int,
	log *zap.Logger,
) {
	actions := p.Actions().List()

	// Find and increment both turn and generation counts
	for i := range actions {
		if actions[i].CardID == cardID && actions[i].BehaviorIndex == behaviorIndex {
			actions[i].TimesUsedThisTurn++
			actions[i].TimesUsedThisGeneration++
			log.Debug("📊 Incremented action usage counts",
				zap.Int("times_used_this_turn", actions[i].TimesUsedThisTurn),
				zap.Int("times_used_this_generation", actions[i].TimesUsedThisGeneration))
			break
		}
	}

	// Update player actions
	p.Actions().SetActions(actions)
}
