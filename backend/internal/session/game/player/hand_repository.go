package player

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
)

var logHand = logger.Get()

// HandRepository handles card hand operations for a specific player
// Auto-saves changes after every operation
type HandRepository struct {
	player *Player // Reference to parent player
}

// NewHandRepository creates a new hand repository for a specific player
func NewHandRepository(player *Player) *HandRepository {
	return &HandRepository{
		player: player,
	}
}

// AddCard adds a card to player's hand
// Auto-saves changes to the player
func (r *HandRepository) AddCard(ctx context.Context, cardID string) error {
	logHand.Debug("🃏 BEFORE AddCard",
		zap.String("game_id", r.player.GameID),
		zap.String("player_id", r.player.ID),
		zap.String("card_to_add", cardID),
		zap.Strings("current_cards", r.player.Cards),
		zap.Int("card_count", len(r.player.Cards)))

	r.player.Cards = append(r.player.Cards, cardID)

	logHand.Debug("🃏 AFTER AddCard",
		zap.String("game_id", r.player.GameID),
		zap.String("player_id", r.player.ID),
		zap.String("card_added", cardID),
		zap.Strings("current_cards", r.player.Cards),
		zap.Int("card_count", len(r.player.Cards)))

	return nil
}

// RemoveCard removes a card from the player's hand and adds it to played cards
// Auto-saves changes to the player
func (r *HandRepository) RemoveCard(ctx context.Context, cardID string) error {
	// Find and remove the card from the player's hand
	for i, id := range r.player.Cards {
		if id == cardID {
			r.player.Cards = append(r.player.Cards[:i], r.player.Cards[i+1:]...)
			// Add to played cards
			r.player.PlayedCards = append(r.player.PlayedCards, cardID)
			return nil
		}
	}

	return fmt.Errorf("card %s not found in player's hand", cardID)
}
