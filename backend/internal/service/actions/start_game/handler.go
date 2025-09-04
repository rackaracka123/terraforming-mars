package start_game

import (
	"context"
	"fmt"
	"math/rand"
	"terraforming-mars-backend/internal/delivery/dto"
	"terraforming-mars-backend/internal/events"
	"terraforming-mars-backend/internal/model"
	"terraforming-mars-backend/internal/model/card_selection"
	"terraforming-mars-backend/internal/repository"
	"terraforming-mars-backend/internal/logger"
	"time"
	
	"go.uber.org/zap"
)

// StartGameHandler handles start game actions
type StartGameHandler struct{
	rng                *rand.Rand
	cardSelectionRepo  *repository.CardSelectionRepository
	eventRepository    *events.EventRepository
}

// NewStartGameHandler creates a new start game handler
func NewStartGameHandler(cardSelectionRepo *repository.CardSelectionRepository, eventRepository *events.EventRepository) *StartGameHandler {
	return &StartGameHandler{
		rng:               rand.New(rand.NewSource(time.Now().UnixNano())),
		cardSelectionRepo: cardSelectionRepo,
		eventRepository:   eventRepository,
	}
}

// Handle applies the start game action
func (h *StartGameHandler) Handle(game *model.Game, player *model.Player, actionRequest dto.ActionStartGameRequest) error {
	action := actionRequest.GetAction()
	return h.applyStartGame(game, player, *action)
}

// applyStartGame applies start game action
func (h *StartGameHandler) applyStartGame(game *model.Game, player *model.Player, action dto.StartGameAction) error {
	log := logger.WithGameContext(game.ID, player.ID)
	
	// Validate that the player is the host
	if !game.IsHost(player.ID) {
		return fmt.Errorf("only the host can start the game")
	}

	// Validate game can be started
	if game.Status != model.GameStatusLobby {
		return fmt.Errorf("game is not in lobby status")
	}

	if len(game.Players) < 1 {
		return fmt.Errorf("need at least 1 player to start the game")
	}

	// Start the game
	game.Status = model.GameStatusActive
	game.CurrentPhase = model.GamePhaseStartingCardSelection

	// Set first player as active
	if len(game.Players) > 0 {
		game.CurrentPlayerID = game.Players[0].ID
	}

	// Deal starting cards directly to each player
	if err := h.dealStartingCards(game); err != nil {
		log.Error("Failed to deal starting cards", zap.Error(err))
		return fmt.Errorf("failed to deal starting cards: %w", err)
	}

	// Publish game started event
	if h.eventRepository != nil {
		gameStartedEvent := events.NewGameStartedEvent(game.ID, len(game.Players))
		if err := h.eventRepository.Publish(context.Background(), gameStartedEvent); err != nil {
			log.Warn("Failed to publish game started event", zap.Error(err))
		}
	}

	log.Info("Game started and starting cards dealt", 
		zap.Int("player_count", len(game.Players)),
	)

	return nil
}

// dealStartingCards deals 5 starting cards to each player
func (h *StartGameHandler) dealStartingCards(game *model.Game) error {
	log := logger.WithGameContext(game.ID, "")
	log.Info("Dealing starting cards to all players")

	// Get available starting cards
	availableCards := model.GetStartingCards()
	if len(availableCards) < 5 {
		log.Error("Not enough starting cards available", 
			zap.Int("available", len(availableCards)),
		)
		return fmt.Errorf("insufficient starting cards available: %d", len(availableCards))
	}

	// Create player card options for each player
	var playerCardOptions []card_selection.PlayerCardOptions
	for i := range game.Players {
		cardOptions := h.selectRandomCards(availableCards, 5)
		
		playerCardOptions = append(playerCardOptions, card_selection.PlayerCardOptions{
			PlayerID:    game.Players[i].ID,
			CardOptions: cardOptions,
		})
		
		log.Info("Generated starting card options for player",
			zap.String("player_id", game.Players[i].ID),
			zap.String("player_name", game.Players[i].Name),
			zap.Int("card_count", len(cardOptions)),
		)

		// Publish player starting card options event
		if h.eventRepository != nil {
			cardOptionsEvent := events.NewPlayerStartingCardOptionsEvent(game.ID, game.Players[i].ID, cardOptions)
			if err := h.eventRepository.Publish(context.Background(), cardOptionsEvent); err != nil {
				log.Warn("Failed to publish player starting card options event", 
					zap.String("player_id", game.Players[i].ID),
					zap.Error(err),
				)
			}
		}
	}

	// Store the card selection data in the repository
	if err := h.cardSelectionRepo.CreateStartingCardSelection(game.ID, playerCardOptions); err != nil {
		log.Error("Failed to store starting card selection", zap.Error(err))
		return fmt.Errorf("failed to store starting card selection: %w", err)
	}

	log.Info("Starting cards dealt to all players",
		zap.Int("players", len(game.Players)),
	)

	return nil
}

// selectRandomCards selects n random cards from the available cards
func (h *StartGameHandler) selectRandomCards(availableCards []model.Card, count int) []string {
	if count > len(availableCards) {
		count = len(availableCards)
	}

	// Create a copy of card IDs to avoid modifying the original
	cardIDs := make([]string, len(availableCards))
	for i, card := range availableCards {
		cardIDs[i] = card.ID
	}

	// Shuffle and select first n cards
	h.rng.Shuffle(len(cardIDs), func(i, j int) {
		cardIDs[i], cardIDs[j] = cardIDs[j], cardIDs[i]
	})

	return cardIDs[:count]
}