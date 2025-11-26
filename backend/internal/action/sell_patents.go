package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	game "terraforming-mars-backend/internal/session/game/core"
	"terraforming-mars-backend/internal/session/game/player/selection"

	"go.uber.org/zap"
)

// SellPatentsAction handles the business logic for initiating sell patents standard project
// This is Phase 1: creates pending card selection for player to choose which cards to sell
type SellPatentsAction struct {
	BaseAction
	gameRepo game.Repository // Still needed for validation
}

// NewSellPatentsAction creates a new sell patents action
func NewSellPatentsAction(
	gameRepo game.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *SellPatentsAction {
	return &SellPatentsAction{
		BaseAction: NewBaseAction(sessionMgrFactory),
		gameRepo:   gameRepo,
	}
}

// Execute performs the sell patents action (Phase 1: initiate card selection)
func (a *SellPatentsAction) Execute(ctx context.Context, sess *session.Session, playerID string) error {
	gameID := sess.GetGameID()
	log := a.InitLogger(gameID, playerID)
	log.Info("🏛️ Initiating sell patents")

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
	p, exists := sess.GetPlayer(playerID)
	if !exists {
		log.Error("Player not found in session")
		return fmt.Errorf("player not found: %s", playerID)
	}

	// 4. Validate player has cards to sell
	playerCards := p.Hand().Cards()
	if len(playerCards) == 0 {
		log.Warn("Player has no cards to sell")
		return fmt.Errorf("no cards available to sell")
	}

	// 5. Create pending card selection
	// Each card costs 0 M€ to sell and rewards 1 M€
	cardCosts := make(map[string]int)
	cardRewards := make(map[string]int)
	for _, cardID := range playerCards {
		cardCosts[cardID] = 0   // No cost to sell
		cardRewards[cardID] = 1 // 1 M€ reward per card
	}

	pendingSelection := &selection.PendingCardSelection{
		Source:         "sell-patents",
		AvailableCards: playerCards,
		CardCosts:      cardCosts,
		CardRewards:    cardRewards,
		MinCards:       0, // Can cancel by selecting 0 cards
		MaxCards:       len(playerCards),
	}

	// 6. Store pending card selection (card selection phase state on Player)
	p.Selection().SetPendingCardSelection(pendingSelection)

	log.Info("📋 Created pending card selection for sell patents",
		zap.Int("available_cards", len(playerCards)))

	// 7. Broadcast state (DO NOT consume action - happens in Phase 2)
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Sell patents initiated successfully, awaiting card selection")
	return nil
}
