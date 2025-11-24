package action

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/session"
	"terraforming-mars-backend/internal/session/game"
	"terraforming-mars-backend/internal/session/player"

	"go.uber.org/zap"
)

// SellPatentsAction handles the business logic for initiating sell patents standard project
// This is Phase 1: creates pending card selection for player to choose which cards to sell
type SellPatentsAction struct {
	BaseAction
}

// NewSellPatentsAction creates a new sell patents action
func NewSellPatentsAction(
	gameRepo game.Repository,
	playerRepo player.Repository,
	sessionMgrFactory session.SessionManagerFactory,
) *SellPatentsAction {
	return &SellPatentsAction{
		BaseAction: NewBaseAction(gameRepo, playerRepo, sessionMgrFactory),
	}
}

// Execute performs the sell patents action (Phase 1: initiate card selection)
func (a *SellPatentsAction) Execute(ctx context.Context, gameID, playerID string) error {
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

	// 3. Validate player exists
	p, err := ValidatePlayer(ctx, a.playerRepo, gameID, playerID, log)
	if err != nil {
		return err
	}

	// 4. Validate player has cards to sell
	if len(p.Cards) == 0 {
		log.Warn("Player has no cards to sell")
		return fmt.Errorf("no cards available to sell")
	}

	// 5. Create pending card selection
	// Each card costs 0 M€ to sell and rewards 1 M€
	cardCosts := make(map[string]int)
	cardRewards := make(map[string]int)
	for _, cardID := range p.Cards {
		cardCosts[cardID] = 0   // No cost to sell
		cardRewards[cardID] = 1 // 1 M€ reward per card
	}

	pendingSelection := &player.PendingCardSelection{
		Source:         "sell-patents",
		AvailableCards: p.Cards,
		CardCosts:      cardCosts,
		CardRewards:    cardRewards,
		MinCards:       0, // Can cancel by selecting 0 cards
		MaxCards:       len(p.Cards),
	}

	// 6. Store pending card selection
	err = a.playerRepo.UpdatePendingCardSelection(ctx, gameID, playerID, pendingSelection)
	if err != nil {
		log.Error("Failed to create pending card selection", zap.Error(err))
		return fmt.Errorf("failed to create pending card selection: %w", err)
	}

	log.Info("📋 Created pending card selection for sell patents",
		zap.Int("available_cards", len(p.Cards)))

	// 7. Broadcast state (DO NOT consume action - happens in Phase 2)
	a.BroadcastGameState(gameID, log)

	log.Info("✅ Sell patents initiated successfully, awaiting card selection")
	return nil
}
