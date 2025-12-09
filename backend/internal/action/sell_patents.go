package action

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/game"
	playerPkg "terraforming-mars-backend/internal/game/player"
)

// SellPatentsAction handles the business logic for initiating sell patents standard project
// This is Phase 1: creates pending card selection for player to choose which cards to sell
// MIGRATION: Uses new architecture (GameRepository only, event-driven broadcasting)
type SellPatentsAction struct {
	BaseAction
}

// NewSellPatentsAction creates a new sell patents action
func NewSellPatentsAction(
	gameRepo game.GameRepository,
	logger *zap.Logger,
) *SellPatentsAction {
	return &SellPatentsAction{
		BaseAction: BaseAction{
			gameRepo: gameRepo,
			logger:   logger,
		},
	}
}

// Execute performs the sell patents action (Phase 1: initiate card selection)
func (a *SellPatentsAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := a.InitLogger(gameID, playerID).With(zap.String("action", "sell_patents"))
	log.Info("🏛️ Initiating sell patents")

	// 1. Fetch game from repository and validate it's active
	g, err := ValidateActiveGame(ctx, a.GameRepository(), gameID, log)
	if err != nil {
		return err
	}

	// 2. Validate it's the player's turn
	if err := ValidateCurrentTurn(g, playerID, log); err != nil {
		return err
	}

	// 3. Get player from game
	player, err := a.GetPlayerFromGame(g, playerID, log)
	if err != nil {
		return err
	}

	// 5. BUSINESS LOGIC: Validate player has cards to sell
	playerCards := player.Hand().Cards()
	if len(playerCards) == 0 {
		log.Warn("Player has no cards to sell")
		return fmt.Errorf("no cards available to sell")
	}

	// 6. BUSINESS LOGIC: Create pending card selection
	// Each card costs 0 M€ to sell and rewards 1 M€
	cardCosts := make(map[string]int)
	cardRewards := make(map[string]int)
	for _, cardID := range playerCards {
		cardCosts[cardID] = 0   // No cost to sell
		cardRewards[cardID] = 1 // 1 M€ reward per card
	}

	pendingSelection := &playerPkg.PendingCardSelection{
		Source:         "sell-patents",
		AvailableCards: playerCards,
		CardCosts:      cardCosts,
		CardRewards:    cardRewards,
		MinCards:       0, // Can cancel by selecting 0 cards
		MaxCards:       len(playerCards),
	}

	// 7. Store pending card selection (card selection phase state on Player)
	player.Selection().SetPendingCardSelection(pendingSelection)

	log.Info("📋 Created pending card selection for sell patents",
		zap.Int("available_cards", len(playerCards)))

	// Note: DO NOT consume action here - happens in Phase 2 (confirm_sell_patents)

	log.Info("✅ Sell patents initiated successfully, awaiting card selection")
	return nil
}
