package card_selection

import (
	"context"
	"fmt"

	"terraforming-mars-backend/internal/delivery/websocket/session"
	"terraforming-mars-backend/internal/features/card"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/player"

	"go.uber.org/zap"
)

// SelectProductionCardsAction handles card selection during production phase
// This action orchestrates:
// - Card selection and validation
// - State broadcasting to all players
//
// TODO: Implement production phase card selection logic when production feature is implemented.
// This is currently a stub that returns "not yet implemented".
type SelectProductionCardsAction struct {
	playerRepo     player.Repository
	cardRepo       card.CardRepository
	sessionManager session.SessionManager
}

// NewSelectProductionCardsAction creates a new select production cards action
func NewSelectProductionCardsAction(
	playerRepo player.Repository,
	cardRepo card.CardRepository,
	sessionManager session.SessionManager,
) *SelectProductionCardsAction {
	return &SelectProductionCardsAction{
		playerRepo:     playerRepo,
		cardRepo:       cardRepo,
		sessionManager: sessionManager,
	}
}

// Execute performs the select production cards action
// Steps:
// 1. Validate and process card selection
// 2. Broadcast updated game state to all players
//
// TODO: Implement when production phase is defined. Needs:
// - Production phase state management
// - Multi-player coordination
// - Phase advancement logic
func (a *SelectProductionCardsAction) Execute(ctx context.Context, gameID string, playerID string, cardIDs []string) error {
	log := logger.WithGameContext(gameID, playerID)
	log.Info("🃏 Executing select production cards action", zap.Int("card_count", len(cardIDs)))

	// TODO: Implement production phase card selection
	// For now, return not implemented
	return fmt.Errorf("SelectProductionCards not yet implemented")
}
