package action

import (
	"context"
	"fmt"
	"slices"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/card"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"
	"terraforming-mars-backend/internal/session/tile"
	"terraforming-mars-backend/internal/session/types"

	"go.uber.org/zap"
)

// PlayCardAction handles the business logic for playing cards from hand
type PlayCardAction struct {
	BaseAction
	gameRepo      game.Repository
	cardManager   card.CardManager
	tileProcessor *tile.Processor
}

// NewPlayCardAction creates a new play card action
func NewPlayCardAction(
	gameRepo game.Repository,
	sessionFactory session.SessionFactory,
	cardManager card.CardManager,
	tileProcessor *tile.Processor,
	sessionMgrFactory session.SessionManagerFactory,
) *PlayCardAction {
	return &PlayCardAction{
		BaseAction:    NewBaseAction(sessionFactory, sessionMgrFactory),
		gameRepo:      gameRepo,
		cardManager:   cardManager,
		tileProcessor: tileProcessor,
	}
}

// Execute performs the play card action
func (a *PlayCardAction) Execute(
	ctx context.Context,
	gameID, playerID, cardID string,
	payment *types.CardPayment,
	choiceIndex *int,
	cardStorageTarget *string,
) error {
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
	sess := a.GetSessionFactory().Get(gameID)
	if sess == nil {
		log.Error("Game session not found")
		return fmt.Errorf("game not found: %s", gameID)
	}

	player, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Validate card is in player's hand
	if !slices.Contains(player.Cards, cardID) {
		log.Error("❌ Card not found in player's hand",
			zap.String("requested_card", cardID),
			zap.Strings("player_cards", player.Cards),
			zap.Int("card_count", len(player.Cards)))
		return fmt.Errorf("card %s not in player's hand", cardID)
	}

	log.Debug("✅ Card validated in hand",
		zap.Strings("player_cards", player.Cards),
		zap.Int("card_count", len(player.Cards)))

	// 5. Validate card can be played (requirements, affordability, choices)
	err = a.cardManager.CanPlay(ctx, gameID, playerID, cardID, payment, choiceIndex, cardStorageTarget)
	if err != nil {
		log.Error("Cannot play card", zap.Error(err))
		return fmt.Errorf("cannot play card: %w", err)
	}

	log.Debug("✅ Card requirements and affordability validated")

	// 6. Play card (deduct payment, move to played cards, apply effects, subscribe passive effects)
	err = a.cardManager.PlayCard(ctx, gameID, playerID, cardID, payment, choiceIndex, cardStorageTarget)
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

	if player.AvailableActions > 0 {
		newActions := player.AvailableActions - 1
		err = player.Action.UpdateAvailableActions(ctx, newActions)
		if err != nil {
			log.Error("Failed to consume action", zap.Error(err))
			return fmt.Errorf("failed to consume action: %w", err)
		}
		log.Debug("✅ Action consumed", zap.Int("remaining_actions", newActions))
	}

	// 9. Broadcast state to all players
	a.BroadcastGameState(gameID, log)

	log.Info("🎉 Card played successfully")
	return nil
}
