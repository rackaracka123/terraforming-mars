package execute_card_action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/action"
	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game/card"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/game/deck"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// ExecuteCardActionAction handles the business logic for executing card actions
// Fully migrated to session-based architecture
type ExecuteCardActionAction struct {
	action.BaseAction
	gameRepo  game.Repository
	validator *Validator
	processor *Processor
}

// NewExecuteCardActionAction creates a new execute card action action
func NewExecuteCardActionAction(
	gameRepo game.Repository,
	sessionFactory session.SessionFactory,
	sessionMgrFactory session.SessionManagerFactory,
	cardProcessor *card.CardProcessor,
	deckRepo deck.Repository,
) *ExecuteCardActionAction {
	return &ExecuteCardActionAction{
		BaseAction: action.NewBaseAction(sessionFactory, sessionMgrFactory),
		gameRepo:   gameRepo,
		validator:  NewValidator(sessionFactory),
		processor:  NewProcessor(sessionFactory, cardProcessor, deckRepo),
	}
}

// Execute performs the execute card action
func (a *ExecuteCardActionAction) Execute(
	ctx context.Context,
	gameID, playerID, cardID string,
	behaviorIndex int,
	choiceIndex *int,
	cardStorageTarget *string,
) error {
	log := a.InitLogger(gameID, playerID)
	log.Debug("🎯 Starting card action play",
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex))

	// 1. Get game and validate current turn
	g, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for card action", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	if g.CurrentTurn == nil {
		log.Error("No current player turn set", zap.String("requesting_player", playerID))
		return fmt.Errorf("no current player turn set, requesting player is %s", playerID)
	}

	if *g.CurrentTurn != playerID {
		log.Error("Not current players turn", zap.String("current_turn", *g.CurrentTurn), zap.String("requesting_player", playerID))
		return fmt.Errorf("not current player's turn: current turn is %s, requesting player is %s", *g.CurrentTurn, playerID)
	}

	// 2. Get session and player to validate they exist and check their actions
	sess := a.GetSessionFactory().Get(gameID)
	if sess == nil {
		log.Error("Game session not found")
		return fmt.Errorf("game session not found: %s", gameID)
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// Check if player has available actions
	if player.AvailableActions <= 0 && player.AvailableActions != -1 {
		return fmt.Errorf("no available actions remaining")
	}

	// 3. Find the specific action in the player's action list
	var targetAction *types.PlayerAction
	for i := range player.Actions {
		action := &player.Actions[i]
		if action.CardID == cardID && action.BehaviorIndex == behaviorIndex {
			targetAction = action
			break
		}
	}

	if targetAction == nil {
		return fmt.Errorf("card action not found in player's action list: card %s, behavior %d", cardID, behaviorIndex)
	}

	// 4. Validate that the action hasn't been played this generation (playCount must be 0)
	if targetAction.PlayCount > 0 {
		return fmt.Errorf("action has already been played this generation: current play count %d", targetAction.PlayCount)
	}

	// 5. Validate choice selection for actions with choices
	if len(targetAction.Behavior.Choices) > 0 {
		if choiceIndex == nil {
			return fmt.Errorf("action has choices but no choiceIndex provided")
		}
		if *choiceIndex < 0 || *choiceIndex >= len(targetAction.Behavior.Choices) {
			return fmt.Errorf("invalid choiceIndex %d: must be between 0 and %d", *choiceIndex, len(targetAction.Behavior.Choices)-1)
		}
		log.Debug("🎯 Action has choices, using choiceIndex", zap.Int("choice_index", *choiceIndex))
	}

	log.Debug("🎯 Found target action",
		zap.String("card_name", targetAction.CardName),
		zap.Int("play_count", targetAction.PlayCount))

	// 6. Validate that the player can afford the action inputs (including choice-specific inputs)
	if err := a.validator.ValidateActionInputs(ctx, gameID, playerID, targetAction, choiceIndex); err != nil {
		return fmt.Errorf("action input validation failed: %w", err)
	}

	// 7. Apply the action inputs (deduct resources, including choice-specific inputs)
	if err := a.processor.ApplyActionInputs(ctx, gameID, playerID, targetAction, choiceIndex); err != nil {
		return fmt.Errorf("failed to apply action inputs: %w", err)
	}

	// 8. Apply the action outputs (give resources/production/etc., including choice-specific outputs)
	if err := a.processor.ApplyActionOutputs(ctx, gameID, playerID, targetAction, choiceIndex, cardStorageTarget); err != nil {
		return fmt.Errorf("failed to apply action outputs: %w", err)
	}

	// 9. Increment the play count for this action
	if err := a.processor.IncrementActionPlayCount(ctx, gameID, playerID, cardID, behaviorIndex); err != nil {
		return fmt.Errorf("failed to increment action play count: %w", err)
	}

	// 10. Consume one action now that all steps have succeeded
	if player.AvailableActions > 0 {
		newActions := player.AvailableActions - 1
		if err := player.Action.UpdateAvailableActions(ctx, newActions); err != nil {
			log.Error("Failed to consume player action", zap.Error(err))
			// Note: Action has already been applied, but we couldn't consume the action
			// This is a critical error but we don't rollback the entire action
			return fmt.Errorf("action applied but failed to consume available action: %w", err)
		}
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", newActions))
	} else {
		log.Debug("✅ Action consumed (unlimited actions)", zap.Int("available_actions", -1))
	}

	// 11. Broadcast game state update
	if err := a.BaseAction.GetSessionManagerFactory().GetOrCreate(gameID).Broadcast(); err != nil {
		log.Error("Failed to broadcast game state after card action play",
			zap.Error(err))
		// Don't fail the action, just log the error
	}

	log.Info("✅ Card action played successfully",
		zap.String("card_id", cardID),
		zap.String("card_name", targetAction.CardName),
		zap.Int("behavior_index", behaviorIndex))
	return nil
}
