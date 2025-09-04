package starting_cards

import (
	"context"
	"fmt"
	"math/rand"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/logger"
	"terraforming-mars-backend/internal/repository"
	"time"

	"go.uber.org/zap"
)

// Listener handles starting card selection events
type Listener struct {
	gameRepo *repository.GameRepository
	eventBus events.EventBus
	rng      *rand.Rand
}

// NewListener creates a new starting cards listener
func NewListener(gameRepo *repository.GameRepository, eventBus events.EventBus) *Listener {
	return &Listener{
		gameRepo: gameRepo,
		eventBus: eventBus,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// OnGameStarted handles the GameStartedEvent by dealing 5 starting cards to each player
func (l *Listener) OnGameStarted(ctx context.Context, event events.Event) error {
	log := logger.WithGameContext(event.GetGameID(), "")
	log.Info("Processing GameStarted event for starting card selection")

	// Get the game
	game, err := l.gameRepo.GetGame(event.GetGameID())
	if err != nil {
		log.Error("Failed to get game for starting card selection", zap.Error(err))
		return fmt.Errorf("failed to get game: %w", err)
	}

	// Validate game state
	if game.CurrentPhase != model.GamePhaseStartingCardSelection {
		log.Warn("Game is not in starting card selection phase", 
			zap.String("current_phase", string(game.CurrentPhase)),
		)
		return fmt.Errorf("game is not in starting card selection phase")
	}

	// Get available starting cards
	availableCards := model.GetStartingCards()
	if len(availableCards) < 5 {
		log.Error("Not enough starting cards available", 
			zap.Int("available", len(availableCards)),
		)
		return fmt.Errorf("insufficient starting cards available: %d", len(availableCards))
	}

	// Deal 5 cards to each player and publish individual events
	for i := range game.Players {
		cardOptions := l.selectRandomCards(availableCards, 5)
		
		// Publish event with card options for this specific player
		cardOptionsEvent := events.NewPlayerStartingCardOptionsEvent(
			game.ID, 
			game.Players[i].ID, 
			cardOptions,
		)
		
		if err := l.eventBus.Publish(ctx, cardOptionsEvent); err != nil {
			log.Error("Failed to publish player starting card options event",
				zap.String("player_id", game.Players[i].ID),
				zap.Error(err),
			)
			return fmt.Errorf("failed to publish card options event: %w", err)
		}
		
		log.Info("Published starting card options for player",
			zap.String("player_id", game.Players[i].ID),
			zap.String("player_name", game.Players[i].Name),
			zap.Int("card_count", len(cardOptions)),
		)
	}

	log.Info("Starting card options published for all players",
		zap.Int("players", len(game.Players)),
	)

	return nil
}

// selectRandomCards selects n random cards from the available cards
func (l *Listener) selectRandomCards(availableCards []model.Card, count int) []string {
	if count > len(availableCards) {
		count = len(availableCards)
	}

	// Create a copy of card IDs to avoid modifying the original
	cardIDs := make([]string, len(availableCards))
	for i, card := range availableCards {
		cardIDs[i] = card.ID
	}

	// Shuffle and select first n cards
	l.rng.Shuffle(len(cardIDs), func(i, j int) {
		cardIDs[i], cardIDs[j] = cardIDs[j], cardIDs[i]
	})

	return cardIDs[:count]
}