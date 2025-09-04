package card_effects

import (
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/cards"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/repository"

	"go.uber.org/zap"
)

// CardEffectsListener handles card played events and manages card interactions
type CardEffectsListener struct {
	gameRepo     *repository.GameRepository
	cardRegistry *cards.CardHandlerRegistry
}

// NewCardEffectsListener creates a new card effects listener
func NewCardEffectsListener(gameRepo *repository.GameRepository, cardRegistry *cards.CardHandlerRegistry) *CardEffectsListener {
	return &CardEffectsListener{
		gameRepo:     gameRepo,
		cardRegistry: cardRegistry,
	}
}

// HandleCardPlayed processes card played events for potential card interactions
func (l *CardEffectsListener) HandleCardPlayed(event events.Event) error {
	log := logger.Get()
	
	payload, ok := event.GetPayload().(events.CardPlayedPayload)
	if !ok {
		log.Error("Invalid card played event payload")
		return nil // Don't propagate error to prevent event bus issues
	}
	
	log.Info("Processing card played event", 
		zap.String("gameId", payload.GameID),
		zap.String("playerId", payload.PlayerID),
		zap.String("cardId", payload.CardID),
	)
	
	// Get the game state
	game, err := l.gameRepo.GetGame(payload.GameID)
	if err != nil {
		log.Error("Failed to get game for card interaction", zap.Error(err))
		return nil
	}
	
	// Find the player
	var targetPlayer *model.Player
	for i := range game.Players {
		if game.Players[i].ID == payload.PlayerID {
			targetPlayer = &game.Players[i]
			break
		}
	}
	
	if targetPlayer == nil {
		log.Error("Player not found for card interaction", zap.String("playerId", payload.PlayerID))
		return nil
	}
	
	// Process potential card interactions
	if err := l.processCardInteractions(game, targetPlayer, payload.CardID); err != nil {
		log.Error("Failed to process card interactions", zap.Error(err))
		return nil
	}
	
	// Save updated game state
	if err := l.gameRepo.UpdateGame(game); err != nil {
		log.Error("Failed to save game after card interactions", zap.Error(err))
		return nil
	}
	
	return nil
}

// processCardInteractions handles interactions between the played card and other cards/game state
func (l *CardEffectsListener) processCardInteractions(game *model.Game, player *model.Player, playedCardID string) error {
	log := logger.Get()
	
	// This is where complex card interactions would be handled
	// For example:
	// - Cards that trigger when other cards are played
	// - Cards that modify the effects of other cards
	// - Cards that react to global parameter changes
	
	// For now, we'll just log that the interaction system is ready
	log.Debug("Card interaction system processed card",
		zap.String("cardId", playedCardID),
		zap.String("playerId", player.ID),
	)
	
	// Future implementations could include:
	// 1. Check all played cards for reaction abilities
	// 2. Apply any triggered effects
	// 3. Handle resource conversions or bonuses
	// 4. Process milestone achievements
	// 5. Update victory point calculations
	
	return nil
}