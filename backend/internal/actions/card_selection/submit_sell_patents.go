package card_selection

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/resources"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// SubmitSellPatentsAction handles completion of the sell patents card selection
// This action orchestrates:
// - Validation of pending card selection
// - Validation of selected cards
// - Awarding payment (1 MC per card)
// - Removing cards from hand
// - Clearing pending selection
type SubmitSellPatentsAction struct {
	playerRepo     player.Repository
	resourcesMech  resources.Service
	sessionManager session.SessionManager
}

// NewSubmitSellPatentsAction creates a new submit sell patents action
func NewSubmitSellPatentsAction(
	playerRepo player.Repository,
	resourcesMech resources.Service,
	sessionManager session.SessionManager,
) *SubmitSellPatentsAction {
	return &SubmitSellPatentsAction{
		playerRepo:     playerRepo,
		resourcesMech:  resourcesMech,
		sessionManager: sessionManager,
	}
}

// Execute performs the submit sell patents action
// Steps:
// 1. Get and validate pending card selection exists
// 2. Validate card count is within bounds
// 3. Validate all selected cards are available
// 4. Calculate total reward (1 MC per card)
// 5. Award credits via resources mechanic
// 6. Remove cards from hand
// 7. Clear pending selection
// 8. Broadcast state
func (a *SubmitSellPatentsAction) Execute(ctx context.Context, gameID string, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🏛️ Executing submit sell patents action", zap.Int("cards_to_sell", len(cardIDs)))

	// Get pending card selection
	selection, err := a.playerRepo.GetPendingCardSelection(ctx, gameID, playerID)
	if err != nil {
		log.Error("Failed to get pending card selection", zap.Error(err))
		return fmt.Errorf("failed to get pending card selection: %w", err)
	}

	if selection == nil {
		log.Warn("No pending card selection found")
		return fmt.Errorf("no pending card selection")
	}

	// Validate source is sell-patents
	if selection.Source != "sell-patents" {
		log.Warn("Pending selection is not for sell patents",
			zap.String("source", selection.Source))
		return fmt.Errorf("pending selection is not for sell patents")
	}

	// Validate card count is within bounds
	if len(cardIDs) < selection.MinCards || len(cardIDs) > selection.MaxCards {
		log.Warn("Invalid card selection count",
			zap.Int("selected", len(cardIDs)),
			zap.Int("min", selection.MinCards),
			zap.Int("max", selection.MaxCards))
		return fmt.Errorf("must select between %d and %d cards, got %d", selection.MinCards, selection.MaxCards, len(cardIDs))
	}

	// Validate all selected cards are in the available list
	availableSet := make(map[string]bool)
	for _, cardID := range selection.AvailableCards {
		availableSet[cardID] = true
	}
	for _, cardID := range cardIDs {
		if !availableSet[cardID] {
			log.Warn("Player attempted to select unavailable card", zap.String("card_id", cardID))
			return fmt.Errorf("card %s is not available for selection", cardID)
		}
	}

	// Calculate total reward (1 MC per card for sell patents)
	totalReward := len(cardIDs)

	// Award credits via resources mechanic
	if totalReward > 0 {
		reward := resources.ResourceSet{
			Credits: totalReward,
		}

		if err := a.resourcesMech.AddResources(ctx, gameID, playerID, reward); err != nil {
			log.Error("Failed to award credits", zap.Error(err))
			return fmt.Errorf("failed to award credits: %w", err)
		}

		log.Info("💰 Credits awarded", zap.Int("amount", totalReward))
	}

	// Remove selected cards from hand
	for _, cardID := range cardIDs {
		if err := a.playerRepo.RemoveCard(ctx, gameID, playerID, cardID); err != nil {
			log.Error("Failed to remove card",
				zap.String("card_id", cardID),
				zap.Error(err))
			return fmt.Errorf("failed to remove card %s: %w", cardID, err)
		}
	}

	log.Info("🃏 Cards removed from hand", zap.Int("count", len(cardIDs)))

	// Clear pending card selection
	if err := a.playerRepo.ClearPendingCardSelection(ctx, gameID, playerID); err != nil {
		log.Error("Failed to clear pending card selection", zap.Error(err))
		return fmt.Errorf("failed to clear pending card selection: %w", err)
	}

	// Broadcast updated game state
	if err := a.sessionManager.Broadcast(gameID); err != nil {
		log.Error("Failed to broadcast game state", zap.Error(err))
		// Don't fail the action, just log
	}

	log.Info("✅ Submit sell patents action completed successfully",
		zap.Int("cards_sold", len(cardIDs)),
		zap.Int("credits_earned", totalReward))

	return nil
}
