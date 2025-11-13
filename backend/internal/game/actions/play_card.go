package actions

import (
	"context"
	"fmt"
	"slices"

	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/game/tiles"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// PlayCardAction handles playing a card from hand
// This action orchestrates:
// - Turn and action validation
// - Card ownership validation
// - Auto-triggered choice validation
// - Card validation and payment via cards.CardManager
// - Card playing via cards.CardManager
// - Tile queue processing if card creates tiles
// - Action consumption
type PlayCardAction struct {
	cardManager    cards.CardManager
	tilesMech      tiles.Service
	gameRepo       repository.GameRepository
	playerRepo     repository.PlayerRepository
	cardRepo       repository.CardRepository
	sessionManager session.SessionManager
}

// NewPlayCardAction creates a new play card action
func NewPlayCardAction(
	cardManager cards.CardManager,
	tilesMech tiles.Service,
	gameRepo repository.GameRepository,
	playerRepo repository.PlayerRepository,
	cardRepo repository.CardRepository,
	sessionManager session.SessionManager,
) *PlayCardAction {
	return &PlayCardAction{
		cardManager:    cardManager,
		tilesMech:      tilesMech,
		gameRepo:       gameRepo,
		playerRepo:     playerRepo,
		cardRepo:       cardRepo,
		sessionManager: sessionManager,
	}
}

// Execute performs the play card action
// Steps:
// 1. Validate turn is player's turn
// 2. Validate player has available actions
// 3. Validate player owns card
// 4. Validate choice selection for auto-triggered choices
// 5. Validate card can be played (via CardManager)
// 6. Play card (via CardManager - handles payment, effects, etc.)
// 7. Process tile queue if card created tiles
// 8. Consume player action (if not infinite)
// 9. Broadcast state
func (a *PlayCardAction) Execute(ctx context.Context, gameID string, playerID string, cardID string, payment *model.CardPayment, choiceIndex *int, cardStorageTarget *string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🃏 Executing play card action", zap.String("card_id", cardID))

	// Validate payment is provided
	if payment == nil {
		log.Warn("Payment is required but not provided")
		return fmt.Errorf("payment is required")
	}

	// Validate turn and actions
	game, err := a.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		log.Error("Failed to get game", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	if game.CurrentTurn == nil {
		log.Warn("No current player turn set")
		return fmt.Errorf("no current player turn set")
	}

	if *game.CurrentTurn != playerID {
		log.Warn("Not current player's turn",
			zap.String("current_turn", *game.CurrentTurn),
			zap.String("requesting_player", playerID))
		return fmt.Errorf("not current player's turn: current turn is %s", *game.CurrentTurn)
	}

	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// -1 Available actions means we have infinite (solo game)
	if player.AvailableActions <= 0 && player.AvailableActions != -1 {
		log.Warn("Player has no actions available", zap.Int("available_actions", player.AvailableActions))
		return fmt.Errorf("no actions available: player has %d actions", player.AvailableActions)
	}

	if !slices.Contains(player.Cards, cardID) {
		log.Warn("Player does not own card", zap.String("card_id", cardID))
		return fmt.Errorf("player does not have card %s", cardID)
	}

	// Validate choice selection for cards with AUTO-triggered choices
	// Manual-triggered behaviors (actions) will have their choices resolved when the action is played
	card, err := a.cardRepo.GetCardByID(ctx, cardID)
	if err != nil {
		log.Error("Failed to get card", zap.Error(err))
		return fmt.Errorf("failed to get card: %w", err)
	}

	// Check if any AUTO-triggered behavior has choices
	hasAutoChoices := false
	for _, behavior := range card.Behaviors {
		// Only check behaviors with auto triggers
		hasAutoTrigger := false
		for _, trigger := range behavior.Triggers {
			if trigger.Type == model.ResourceTriggerAuto {
				hasAutoTrigger = true
				break
			}
		}

		// If this is an auto-triggered behavior with choices, validate choiceIndex
		if hasAutoTrigger && len(behavior.Choices) > 0 {
			hasAutoChoices = true
			// Validate that choiceIndex is provided and within valid range
			if choiceIndex == nil {
				log.Warn("Card has auto-triggered choices but no choiceIndex provided")
				return fmt.Errorf("card has auto-triggered choices but no choiceIndex provided")
			}
			if *choiceIndex < 0 || *choiceIndex >= len(behavior.Choices) {
				log.Warn("Invalid choiceIndex",
					zap.Int("choice_index", *choiceIndex),
					zap.Int("max_choices", len(behavior.Choices)))
				return fmt.Errorf("invalid choiceIndex %d: must be between 0 and %d", *choiceIndex, len(behavior.Choices)-1)
			}
			break
		}
	}

	if hasAutoChoices {
		log.Debug("🎯 Card has auto-triggered choices, using choiceIndex", zap.Int("choice_index", *choiceIndex))
	}

	// Validate card can be played via CardManager
	if err := a.cardManager.CanPlay(ctx, gameID, playerID, cardID, payment, choiceIndex, cardStorageTarget); err != nil {
		log.Warn("Card cannot be played", zap.Error(err))
		return fmt.Errorf("card cannot be played: %w", err)
	}

	log.Debug("✅ Card validation passed")

	// Play card via CardManager (handles payment, effects, card removal, etc.)
	if err := a.cardManager.PlayCard(ctx, gameID, playerID, cardID, payment, choiceIndex, cardStorageTarget); err != nil {
		log.Error("Failed to play card", zap.Error(err))
		return fmt.Errorf("failed to play card: %w", err)
	}

	log.Info("🃏 Card played via CardManager", zap.String("card_id", cardID))

	// Process any tile queue created by the card
	if err := a.tilesMech.ProcessTileQueue(ctx, gameID, playerID); err != nil {
		log.Error("Failed to process tile queue", zap.Error(err))
		return fmt.Errorf("card played but failed to process tile queue: %w", err)
	}
	log.Debug("🎯 Tile queue processed (if any existed)")

	// Consume action (if not infinite)
	if player.AvailableActions != -1 {
		newActions := player.AvailableActions - 1
		if err := a.playerRepo.UpdateAvailableActions(ctx, gameID, playerID, newActions); err != nil {
			log.Error("Failed to consume action", zap.Error(err))
			return fmt.Errorf("card played but failed to consume action: %w", err)
		}
		log.Debug("🎯 Action consumed", zap.Int("remaining_actions", newActions))
	}

	// Broadcast game state update
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state after card play", zap.Error(err))
		// Don't fail the card play operation, just log the error
	}

	log.Info("✅ Play card action completed successfully", zap.String("card_id", cardID))
	return nil
}
