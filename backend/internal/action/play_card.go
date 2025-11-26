package action

import (
	"context"
	"fmt"
	"slices"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/game/board"
	"terraforming-mars-backend/internal/session/game/card"
	gameCore "terraforming-mars-backend/internal/session/game/core"

	"go.uber.org/zap"
)

// PlayCardAction handles the business logic for playing cards from hand
type PlayCardAction struct {
	BaseAction
	gameRepo      gameCore.Repository
	cardManager   game.CardManager
	tileProcessor *board.Processor
}

// NewPlayCardAction creates a new play card action
func NewPlayCardAction(
	gameRepo gameCore.Repository,
	cardManager game.CardManager,
	tileProcessor *board.Processor,
	sessionMgrFactory session.SessionManagerFactory,
) *PlayCardAction {
	return &PlayCardAction{
		BaseAction:    NewBaseAction(sessionMgrFactory),
		gameRepo:      gameRepo,
		cardManager:   cardManager,
		tileProcessor: tileProcessor,
	}
}

// Execute performs the play card action
func (a *PlayCardAction) Execute(
	ctx context.Context,
	sess *session.Session,
	playerID, cardID string,
	payment *card.CardPayment,
	choiceIndex *int,
	cardStorageTarget *string,
) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID).With(zap.String("card_id", cardID))
	log.Info("🎴 Playing card from hand")

	// 1. Validate game is active
	g, err := ValidateActiveGame(ctx, a.gameRepo, gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate it's the player's turn
	if err := ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 3. Get session and player
	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Validate card is in player's hand
	playerCards := player.Hand().Cards()
	if !slices.Contains(playerCards, cardID) {
		log.Error("❌ Card not found in player's hand",
			zap.String("requested_card", cardID),
			zap.Strings("player_cards", playerCards),
			zap.Int("card_count", len(playerCards)))
		return fmt.Errorf("card %s not in player's hand", cardID)
	}

	log.Debug("✅ Card validated in hand",
		zap.Strings("player_cards", playerCards),
		zap.Int("card_count", len(playerCards)))

	// 5. Validate card can be played (requirements, affordability, choices)
	// Note: g is already *game.Game, no conversion needed
	err = a.cardManager.CanPlay(ctx, g, player, cardID, payment, choiceIndex, cardStorageTarget)
	if err != nil {
		log.Error("Cannot play card", zap.Error(err))
		return fmt.Errorf("cannot play card: %w", err)
	}

	log.Debug("✅ Card requirements and affordability validated")

	// 6. Play card (deduct payment, move to played cards, apply effects, subscribe passive effects)
	err = a.cardManager.PlayCard(ctx, g, player, cardID, payment, choiceIndex, cardStorageTarget)
	if err != nil {
		log.Error("Failed to play card", zap.Error(err))
		return fmt.Errorf("failed to play card: %w", err)
	}

	log.Debug("✅ Card played and effects applied")

	// 7. Tile queue processing (now automatic via TileQueueCreatedEvent)
	// No manual call needed - TileProcessor subscribes to events and processes automatically

	// 8. Consume action (only if not unlimited actions)
	// Refresh player to get updated state
	player, exists = sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found after card play")
		return fmt.Errorf("player not found: %s", playerID)
	}

	availableActions := player.Turn().AvailableActions()
	if availableActions > 0 {
		player.Turn().SetAvailableActions(availableActions - 1)
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", availableActions-1))
	}

	// 10. Broadcast state to all players
	a.BroadcastGameState(gameID, log)

	log.Info("🎉 Card played successfully")
	return nil
}
