package player

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"terraforming-mars-backend/internal/logger"
)

var logHand = logger.Get()

// PlayerHandRepository handles player card hand operations (adding/removing cards)
type PlayerHandRepository struct {
	storage *PlayerStorage
}

// NewPlayerHandRepository creates a new player hand repository
func NewPlayerHandRepository(storage *PlayerStorage) *PlayerHandRepository {
	return &PlayerHandRepository{
		storage: storage,
	}
}

// AddCard adds a card to player's hand
func (r *PlayerHandRepository) AddCard(ctx context.Context, gameID string, playerID string, cardID string) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	logHand.Debug("🃏 BEFORE AddCard",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("card_to_add", cardID),
		zap.Strings("current_cards", p.Cards),
		zap.Int("card_count", len(p.Cards)))

	p.Cards = append(p.Cards, cardID)

	logHand.Debug("🃏 AFTER AddCard",
		zap.String("game_id", gameID),
		zap.String("player_id", playerID),
		zap.String("card_added", cardID),
		zap.Strings("current_cards", p.Cards),
		zap.Int("card_count", len(p.Cards)))

	return r.storage.Set(gameID, playerID, p)
}

// RemoveCardFromHand removes a card from the player's hand
func (r *PlayerHandRepository) RemoveCardFromHand(ctx context.Context, gameID string, playerID string, cardID string) error {
	p, err := r.storage.Get(gameID, playerID)
	if err != nil {
		return err
	}

	// Find and remove the card from the player's hand
	for i, id := range p.Cards {
		if id == cardID {
			p.Cards = append(p.Cards[:i], p.Cards[i+1:]...)
			// Add to played cards
			p.PlayedCards = append(p.PlayedCards, cardID)
			return r.storage.Set(gameID, playerID, p)
		}
	}

	return fmt.Errorf("card %s not found in player's hand", cardID)
}
