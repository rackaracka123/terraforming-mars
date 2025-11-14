package actions

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/game"
	"terraforming-mars-backend/internal/player"
	"terraforming-mars-backend/internal/service"

	"go.uber.org/zap"
)

// PlayCardActionAction handles playing a card action (blue card ability)
// This action orchestrates:
// - Turn and action validation
// - Finding the specific action in player's action list
// - Play count validation (once per generation)
// - Choice validation for actions with choices
// - Input validation and application (resource costs)
// - Output application (resource/production gains, card effects)
// - Play count increment
// - Action consumption
//
// Note: The complex input/output logic is delegated to CardService helper methods.
// This is temporary until card effect processing is refactored into a dedicated
// mechanic or effects service.
type PlayCardActionAction struct {
	cardService    service.CardService
	gameRepo       game.Repository
	playerRepo     player.Repository
	sessionManager session.SessionManager
}

// NewPlayCardActionAction creates a new play card action action
func NewPlayCardActionAction(
	cardService service.CardService,
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionManager session.SessionManager,
) *PlayCardActionAction {
	return &PlayCardActionAction{
		cardService:    cardService,
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		sessionManager: sessionManager,
	}
}

// Execute performs the play card action action
// Steps:
// 1. Validate turn is player's turn
// 2. Validate player has available actions
// 3. Find specific action in player's action list
// 4. Validate play count (must be 0 - not played this generation)
// 5. Validate choice selection if action has choices
// 6. Validate action inputs (delegate to CardService)
// 7. Apply action inputs (delegate to CardService)
// 8. Apply action outputs (delegate to CardService)
// 9. Increment play count (delegate to CardService)
// 10. Consume player action (if not infinite)
// 11. Broadcast state
//
// Note: Steps 6-9 are delegated to CardService helper methods which contain
// complex choice-based logic (~900 lines). This delegation is temporary.
func (a *PlayCardActionAction) Execute(ctx context.Context, gameID string, playerID string, cardID string, behaviorIndex int, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🎯 Executing play card action action",
		zap.String("card_id", cardID),
		zap.Int("behavior_index", behaviorIndex))

	// Validate turn
	game, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game for card action", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	if game.CurrentTurn == nil {
		log.Error("No current player turn set", zap.String("requesting_player", playerID))
		return fmt.Errorf("no current player turn set, requesting player is %s", playerID)
	}

	if *game.CurrentTurn != playerID {
		log.Error("Not current player's turn",
			zap.String("current_turn", *game.CurrentTurn),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("not current player's turn: current turn is %s, requesting player is %s", *game.CurrentTurn, playerID)
	}

	// Get the player to validate they exist and check their actions
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player for card action play", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Check if player has available actions
	if player.AvailableActions <= 0 && player.AvailableActions != -1 {
		log.Warn("Player has no available actions", zap.Int("available_actions", player.AvailableActions))
		return fmt.Errorf("no available actions remaining")
	}

	// Find the specific action in the player's action list
	var targetAction *model.PlayerAction
	for i := range player.Actions {
		action := &player.Actions[i]
		if action.CardID == cardID && action.BehaviorIndex == behaviorIndex {
			targetAction = action
			break
		}
	}

	if targetAction == nil {
		log.Warn("Card action not found in player's action list",
			zap.String("card_id", cardID),
			zap.Int("behavior_index", behaviorIndex))
		return fmt.Errorf("card action not found in player's action list: card %s, behavior %d", cardID, behaviorIndex)
	}

	// Validate that the action hasn't been played this generation (playCount must be 0)
	if targetAction.PlayCount > 0 {
		log.Warn("Action has already been played this generation",
			zap.Int("play_count", targetAction.PlayCount))
		return fmt.Errorf("action has already been played this generation: current play count %d", targetAction.PlayCount)
	}

	// Validate choice selection for actions with choices
	if len(targetAction.Behavior.Choices) > 0 {
		if choiceIndex == nil {
			log.Warn("Action has choices but no choiceIndex provided")
			return fmt.Errorf("action has choices but no choiceIndex provided")
		}
		if *choiceIndex < 0 || *choiceIndex >= len(targetAction.Behavior.Choices) {
			log.Warn("Invalid choiceIndex",
				zap.Int("choice_index", *choiceIndex),
				zap.Int("max_choices", len(targetAction.Behavior.Choices)))
			return fmt.Errorf("invalid choiceIndex %d: must be between 0 and %d", *choiceIndex, len(targetAction.Behavior.Choices)-1)
		}
		log.Debug("🎯 Action has choices, using choiceIndex", zap.Int("choice_index", *choiceIndex))
	}

	log.Debug("🎯 Found target action",
		zap.String("card_name", targetAction.CardName),
		zap.Int("play_count", targetAction.PlayCount))

	// Delegate to CardService for complex input/output/increment logic
	// TODO: Refactor these into a dedicated card effects service
	cardServiceImpl, ok := a.cardService.(*service.CardServiceImpl)
	if !ok {
		log.Error("CardService is not of expected type")
		return fmt.Errorf("internal error: CardService type mismatch")
	}

	// Validate that the player can afford the action inputs (including choice-specific inputs)
	if err := cardServiceImpl.ValidateActionInputs(ctx, gameID, playerID, targetAction, choiceIndex); err != nil {
		log.Warn("Action input validation failed", zap.Error(err))
		return fmt.Errorf("action input validation failed: %w", err)
	}

	log.Debug("✅ Action inputs validated")

	// Apply the action inputs (deduct resources, including choice-specific inputs)
	if err := cardServiceImpl.ApplyActionInputs(ctx, gameID, playerID, targetAction, choiceIndex); err != nil {
		log.Error("Failed to apply action inputs", zap.Error(err))
		return fmt.Errorf("failed to apply action inputs: %w", err)
	}

	log.Debug("✅ Action inputs applied")

	// Apply the action outputs (give resources/production/etc., including choice-specific outputs)
	if err := cardServiceImpl.ApplyActionOutputs(ctx, gameID, playerID, targetAction, choiceIndex, cardStorageTarget); err != nil {
		log.Error("Failed to apply action outputs", zap.Error(err))
		return fmt.Errorf("failed to apply action outputs: %w", err)
	}

	log.Debug("✅ Action outputs applied")

	// Increment the play count for this action
	if err := cardServiceImpl.IncrementActionPlayCount(ctx, gameID, playerID, cardID, behaviorIndex); err != nil {
		log.Error("Failed to increment action play count", zap.Error(err))
		return fmt.Errorf("failed to increment action play count: %w", err)
	}

	log.Debug("✅ Action play count incremented")

	// Consume one action now that all steps have succeeded
	if player.AvailableActions > 0 {
		newActions := player.AvailableActions - 1
		if err := a.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, newActions); err != nil {
			log.Error("Failed to consume player action", zap.Error(err))
			// Note: Action has already been applied, but we couldn't consume the action
			// This is a critical error but we don't rollback the entire action
			return fmt.Errorf("action applied but failed to consume available action: %w", err)
		}
		log.Debug("🎯 Action consumed", zap.Int("remaining_actions", newActions))
	} else {
		log.Debug("🎯 Action consumed (unlimited actions)", zap.Int("available_actions", -1))
	}

	// Broadcast game state update
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after card action play", zap.Error(err))
		// Don't fail the action, just log the error
	}

	log.Info("✅ Play card action action completed successfully",
		zap.String("card_id", cardID),
		zap.String("card_name", targetAction.CardName),
		zap.Int("behavior_index", behaviorIndex))

	return nil
}
