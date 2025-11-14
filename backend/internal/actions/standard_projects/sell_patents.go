package standard_projects

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/logger"
	playerPkg "terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// SellPatentsAction handles the sell patents standard project
// This action orchestrates:
// - Validation that player has cards to sell
// - Creating pending card selection for player to choose which cards to sell
type SellPatentsAction struct {
	playerRepo     playerPkg.Repository
	sessionManager session.SessionManager
}

// NewSellPatentsAction creates a new sell patents action
func NewSellPatentsAction(
	playerRepo playerPkg.Repository,
	sessionManager session.SessionManager,
) *SellPatentsAction {
	return &SellPatentsAction{
		playerRepo:     playerRepo,
		sessionManager: sessionManager,
	}
}

// Execute performs the sell patents action
// Steps:
// 1. Validate player has cards to sell
// 2. Create pending card selection with all player's cards
// 3. Broadcast state (player will see card selection UI)
func (a *SellPatentsAction) Execute(ctx context.Context, gameID string, playerID string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🏛️ Executing sell patents action")

	// Get player
	player, err := a.playerRepo.GetByID(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get player", zap.Error(err))
		return fmt.Errorf("failed to get player: %w", err)
	}

	// Validate player has cards to sell
	if len(player.Cards) == 0 {
		log.Warn("Player attempted to sell patents with no cards in hand")
		return fmt.Errorf("player has no cards to sell")
	}

	// Create pending card selection with all player's cards
	// Each card costs 0 MC to "select" (sell) and rewards 1 MC
	cardCosts := make(map[string]int)
	cardRewards := make(map[string]int)
	for _, cardID := range player.Cards {
		cardCosts[cardID] = 0   // Free to select (selling)
		cardRewards[cardID] = 1 // Gain 1 MC per card sold
	}

	selection := &playerPkg.PendingCardSelection{
		AvailableCards: player.Cards, // All cards in hand available
		CardCosts:      cardCosts,
		CardRewards:    cardRewards,
		Source:         "sell-patents",
		MinCards:       0,                 // Can sell 0 cards (cancel action)
		MaxCards:       len(player.Cards), // Can sell all cards
	}

	// Store the pending card selection
	if err := a.playerRepo.UpdatePendingCardSelection(ctx, gameID, playerID, selection); err != nil {
		log.Error("Failed to create pending card selection", zap.Error(err))
		return fmt.Errorf("failed to create pending card selection: %w", err)
	}

	log.Info("🃏 Pending card selection created", zap.Int("available_cards", len(player.Cards)))

	// Broadcast updated game state (includes pendingCardSelection)
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Don't fail the action, just log
	}

	log.Info("✅ Sell patents action completed successfully - awaiting card selection")

	return nil
}
